package pglike

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/yourname/mcp-x/internal/driver"
)

type Driver struct {
	db           *sql.DB
	driverName   string
	schemaFilter string
}

func New(driverName, schemaFilter string) *Driver {
	return &Driver{
		driverName:   driverName,
		schemaFilter: schemaFilter,
	}
}

func (d *Driver) Name() string            { return d.driverName }
func (d *Driver) Type() driver.DriverType { return driver.DriverTypeSQL }

func (d *Driver) Connect(ctx context.Context, cfg driver.ConnConfig) error {
	db, err := sql.Open(d.driverName, cfg.DSN)
	if err != nil {
		return fmt.Errorf("%s open: %w", d.driverName, err)
	}
	if cfg.MaxOpenConns > 0 {
		db.SetMaxOpenConns(cfg.MaxOpenConns)
	}
	if cfg.MaxIdleConns > 0 {
		db.SetMaxIdleConns(cfg.MaxIdleConns)
	}
	if cfg.ConnMaxLifetime > 0 {
		db.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime))
	}
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return fmt.Errorf("%s ping: %w", d.driverName, err)
	}
	d.db = db
	return nil
}

func (d *Driver) Query(ctx context.Context, sqlStr string, args []any) (*driver.QueryResult, error) {
	rows, err := d.db.QueryContext(ctx, sqlStr, args...)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("columns: %w", err)
	}

	var resultRows [][]any
	for rows.Next() {
		values := make([]any, len(cols))
		ptrs := make([]any, len(cols))
		for i := range values {
			ptrs[i] = &values[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}
		for i, v := range values {
			if b, ok := v.([]byte); ok {
				values[i] = string(b)
			}
			if v == nil {
				values[i] = "NULL"
			}
		}
		resultRows = append(resultRows, values)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iter: %w", err)
	}

	return &driver.QueryResult{
		Columns: cols,
		Rows:    resultRows,
	}, nil
}

func (d *Driver) Execute(ctx context.Context, sqlStr string, args []any) (*driver.ExecResult, error) {
	start := time.Now()
	res, err := d.db.ExecContext(ctx, sqlStr, args...)
	if err != nil {
		return nil, fmt.Errorf("exec: %w", err)
	}
	affected, _ := res.RowsAffected()
	return &driver.ExecResult{
		AffectedRows: affected,
		Duration:     time.Since(start).Milliseconds(),
	}, nil
}

func (d *Driver) ListTables(ctx context.Context) ([]driver.TableInfo, error) {
	q := "SELECT TABLE_NAME FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_SCHEMA NOT IN ('pg_catalog','information_schema') ORDER BY TABLE_NAME"
	if d.schemaFilter != "" {
		q = strings.Replace(q, "NOT IN ('pg_catalog','information_schema')",
			"= '"+d.schemaFilter+"'", 1)
	}
	rows, err := d.db.QueryContext(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("list tables: %w", err)
	}
	defer rows.Close()

	var tables []driver.TableInfo
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, fmt.Errorf("scan table: %w", err)
		}
		tables = append(tables, driver.TableInfo{Name: name})
	}
	return tables, rows.Err()
}

func (d *Driver) DescribeTable(ctx context.Context, table string) (*driver.TableSchema, error) {
	rows, err := d.db.QueryContext(ctx, `
SELECT COLUMN_NAME, DATA_TYPE, IS_NULLABLE, COLUMN_DEFAULT
FROM INFORMATION_SCHEMA.COLUMNS
WHERE TABLE_NAME = $1
ORDER BY ORDINAL_POSITION`, table)
	if err != nil {
		return nil, fmt.Errorf("describe: %w", err)
	}
	defer rows.Close()

	schema := &driver.TableSchema{Table: table}
	for rows.Next() {
		var field, typ, nullable, defVal sql.NullString
		if err := rows.Scan(&field, &typ, &nullable, &defVal); err != nil {
			return nil, fmt.Errorf("describe scan: %w", err)
		}
		schema.Columns = append(schema.Columns, driver.ColumnInfo{
			Name:     field.String,
			Type:     typ.String,
			Nullable: nullable.String,
			Default:  defVal.String,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	idxRows, err := d.db.QueryContext(ctx, `
SELECT i.relname AS index_name, a.attname AS column_name, ix.indisunique AS is_unique
FROM pg_index ix
JOIN pg_class i ON i.oid = ix.indexrelid
JOIN pg_class t ON t.oid = ix.indrelid
JOIN pg_attribute a ON a.attrelid = ix.indrelid AND a.attnum = ANY(ix.indkey)
WHERE t.relname = $1
ORDER BY i.relname, a.attnum`, table)
	if err == nil {
		defer idxRows.Close()
		idxMap := make(map[string]*driver.IndexInfo)
		var idxOrder []string
		for idxRows.Next() {
			var idxName, colName sql.NullString
			var isUnique sql.NullBool
			if err := idxRows.Scan(&idxName, &colName, &isUnique); err != nil {
				break
			}
			if !idxName.Valid {
				continue
			}
			idx, ok := idxMap[idxName.String]
			if !ok {
				idx = &driver.IndexInfo{
					Name:   idxName.String,
					Unique: isUnique.Valid && isUnique.Bool,
				}
				idxMap[idxName.String] = idx
				idxOrder = append(idxOrder, idxName.String)
			}
			if colName.Valid {
				idx.Columns = append(idx.Columns, colName.String)
			}
		}
		for _, name := range idxOrder {
			schema.Indexes = append(schema.Indexes, *idxMap[name])
		}
	}

	return schema, nil
}

func (d *Driver) Ping(ctx context.Context) error {
	return d.db.PingContext(ctx)
}

func (d *Driver) Close() error {
	if d.db == nil {
		return nil
	}
	return d.db.Close()
}
