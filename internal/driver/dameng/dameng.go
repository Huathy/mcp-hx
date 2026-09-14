package dameng

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	_ "github.com/ganl/go-dm"
	"github.com/yourname/mcp-x/internal/driver"
)

type DamengDriver struct {
	db *sql.DB
}

func (d *DamengDriver) Name() string            { return "dameng" }
func (d *DamengDriver) Type() driver.DriverType { return driver.DriverTypeSQL }

func (d *DamengDriver) Connect(ctx context.Context, cfg driver.ConnConfig) error {
	db, err := sql.Open("dm", cfg.DSN)
	if err != nil {
		return fmt.Errorf("dameng open: %w", err)
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
		return fmt.Errorf("dameng ping: %w", err)
	}
	d.db = db
	return nil
}

func (d *DamengDriver) Query(ctx context.Context, sqlStr string, args []any) (*driver.QueryResult, error) {
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

func (d *DamengDriver) Execute(ctx context.Context, sqlStr string, args []any) (*driver.ExecResult, error) {
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

func (d *DamengDriver) ListTables(ctx context.Context) ([]driver.TableInfo, error) {
	rows, err := d.db.QueryContext(ctx, "SELECT TABLE_NAME FROM USER_TABLES ORDER BY TABLE_NAME")
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

func (d *DamengDriver) DescribeTable(ctx context.Context, table string) (*driver.TableSchema, error) {
	rows, err := d.db.QueryContext(ctx,
		"SELECT COLUMN_NAME, DATA_TYPE, NULLABLE, DATA_DEFAULT FROM USER_TAB_COLUMNS WHERE TABLE_NAME = ? ORDER BY COLUMN_ID",
		strings.ToUpper(table))
	if err != nil {
		return nil, fmt.Errorf("describe: %w", err)
	}
	defer rows.Close()

	schema := &driver.TableSchema{Table: table}
	for rows.Next() {
		var field, typ, nullable sql.NullString
		var defVal sql.NullString
		if err := rows.Scan(&field, &typ, &nullable, &defVal); err != nil {
			return nil, fmt.Errorf("describe scan: %w", err)
		}
		nullStr := "NO"
		if nullable.Valid && nullable.String == "Y" {
			nullStr = "YES"
		}
		schema.Columns = append(schema.Columns, driver.ColumnInfo{
			Name:     field.String,
			Type:     typ.String,
			Nullable: nullStr,
			Default:  defVal.String,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	idxRows, err := d.db.QueryContext(ctx,
		"SELECT INDEX_NAME, COLUMN_NAME, UNIQUENESS FROM USER_IND_COLUMNS JOIN USER_INDEXES USING(INDEX_NAME) WHERE TABLE_NAME = ? ORDER BY INDEX_NAME, COLUMN_POSITION",
		strings.ToUpper(table))
	if err == nil {
		defer idxRows.Close()
		idxMap := make(map[string]*driver.IndexInfo)
		var idxOrder []string
		for idxRows.Next() {
			var idxName, colName, uniqueness sql.NullString
			if err := idxRows.Scan(&idxName, &colName, &uniqueness); err != nil {
				break
			}
			if !idxName.Valid {
				continue
			}
			idx, ok := idxMap[idxName.String]
			if !ok {
				idx = &driver.IndexInfo{
					Name:   idxName.String,
					Unique: uniqueness.Valid && uniqueness.String == "UNIQUE",
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

func (d *DamengDriver) Ping(ctx context.Context) error {
	return d.db.PingContext(ctx)
}

func (d *DamengDriver) Close() error {
	if d.db == nil {
		return nil
	}
	return d.db.Close()
}

func init() {
	driver.Register("dameng", func() driver.AnyDriver { return &DamengDriver{} })
}
