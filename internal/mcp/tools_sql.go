package mcp

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/yourname/mcp-dbx/internal/driver"
	"github.com/yourname/mcp-dbx/internal/safety"
)

func (s *Server) registerSQLTools() {
	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "db_query",
		Description: "Execute a read-only SQL query (SELECT) and return results as a table. Only SELECT statements allowed. Max 1000 rows by default.",
	}, s.handleDBQuery)

	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "db_execute",
		Description: "Execute a write SQL statement (INSERT/UPDATE/DELETE). Requires write mode enabled in config. DDL (CREATE/DROP/TRUNCATE) blocked unless explicitly allowed.",
	}, s.handleDBExecute)

	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "db_tables",
		Description: "List all tables in the database. Use to explore schema before writing queries.",
	}, s.handleDBTables)

	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "db_schema",
		Description: "Describe table schema: columns, types, nullable, keys, defaults.",
	}, s.handleDBSchema)
}

type dbQueryInput struct {
	DataSource string   `json:"datasource" jsonschema:"the name of the data source"`
	SQL       string   `json:"sql" jsonschema:"SQL SELECT query. Supports parameterized placeholders (?)"`
	Args      []any    `json:"args,omitempty" jsonschema:"query parameters for placeholders"`
	Limit     int      `json:"limit,omitempty" jsonschema:"max rows to return (default 1000)"`
}

type queryOutput struct {
	Text string `json:"text" jsonschema:"query result as markdown table"`
}

func (s *Server) handleDBQuery(ctx context.Context, req *mcp.CallToolRequest, in dbQueryInput) (*mcp.CallToolResult, queryOutput, error) {
	d, err := s.getSQLDriver(in.DataSource)
	if err != nil {
		return nil, queryOutput{}, err
	}

	upper := strings.ToUpper(strings.TrimSpace(in.SQL))
	if !strings.HasPrefix(upper, "SELECT") && !strings.HasPrefix(upper, "WITH") {
		return nil, queryOutput{}, fmt.Errorf("db_query only allows SELECT or WITH statements")
	}

	chk := safety.New(s.cfg.EffectiveSafety(in.DataSource))
	limit := in.Limit
	if limit <= 0 || limit > chk.MaxRows() {
		limit = chk.MaxRows()
	}

	queryCtx, cancel := chk.WrapContext(ctx)
	defer cancel()

	result, err := d.Query(queryCtx, in.SQL, in.Args)
	if err != nil {
		return nil, queryOutput{}, fmt.Errorf("query failed: %w", err)
	}

	if len(result.Rows) > limit {
		result.Rows = result.Rows[:limit]
	}

	text := formatQueryResult(result)
	return nil, queryOutput{Text: text}, nil
}

type dbExecuteInput struct {
	DataSource string `json:"datasource" jsonschema:"the name of the data source"`
	SQL        string `json:"sql" jsonschema:"SQL write statement. Supports parameterized placeholders (?)"`
	Args       []any  `json:"args,omitempty" jsonschema:"query parameters for placeholders"`
}

type execOutput struct {
	Text string `json:"text" jsonschema:"execution result as text"`
}

func (s *Server) handleDBExecute(ctx context.Context, req *mcp.CallToolRequest, in dbExecuteInput) (*mcp.CallToolResult, execOutput, error) {
	d, err := s.getSQLDriver(in.DataSource)
	if err != nil {
		return nil, execOutput{}, err
	}

	chk := safety.New(s.cfg.EffectiveSafety(in.DataSource))
	if err := chk.CheckWrite(); err != nil {
		return nil, execOutput{}, err
	}

	upper := strings.ToUpper(strings.TrimSpace(in.SQL))
	if strings.HasPrefix(upper, "SELECT") {
		return nil, execOutput{}, fmt.Errorf("db_execute does not allow SELECT, use db_query instead")
	}

	if err := chk.CheckSQL(in.SQL); err != nil {
		return nil, execOutput{}, err
	}

	execCtx, cancel := chk.WrapContext(ctx)
	defer cancel()

	result, err := d.Execute(execCtx, in.SQL, in.Args)
	if err != nil {
		return nil, execOutput{}, fmt.Errorf("execute failed: %w", err)
	}

	text := fmt.Sprintf("## Execute Result\n\n- Affected rows: %d\n- Duration: %dms\n- Last insert ID: %d",
		result.AffectedRows, result.Duration, result.LastInsertID)
	return nil, execOutput{Text: text}, nil
}

type dbTablesInput struct {
	DataSource string `json:"datasource" jsonschema:"the name of the data source"`
}

type tablesOutput struct {
	Text string `json:"text" jsonschema:"table list as markdown"`
}

func (s *Server) handleDBTables(ctx context.Context, req *mcp.CallToolRequest, in dbTablesInput) (*mcp.CallToolResult, tablesOutput, error) {
	d, err := s.getSQLDriver(in.DataSource)
	if err != nil {
		return nil, tablesOutput{}, err
	}

	tables, err := d.ListTables(ctx)
	if err != nil {
		return nil, tablesOutput{}, fmt.Errorf("list tables: %w", err)
	}

	if len(tables) == 0 {
		return nil, tablesOutput{Text: "## Tables\n\nNo tables found."}, nil
	}

	var sb strings.Builder
	sb.WriteString("## Tables\n\n| # | Name |\n|---|------|\n")
	for i, t := range tables {
		sb.WriteString(fmt.Sprintf("| %d | %s |\n", i+1, t.Name))
	}
	return nil, tablesOutput{Text: sb.String()}, nil
}

type dbSchemaInput struct {
	DataSource string `json:"datasource" jsonschema:"the name of the data source"`
	Table      string `json:"table" jsonschema:"table name to inspect"`
}

type schemaOutput struct {
	Text string `json:"text" jsonschema:"table schema as markdown"`
}

func (s *Server) handleDBSchema(ctx context.Context, req *mcp.CallToolRequest, in dbSchemaInput) (*mcp.CallToolResult, schemaOutput, error) {
	d, err := s.getSQLDriver(in.DataSource)
	if err != nil {
		return nil, schemaOutput{}, err
	}

	schema, err := d.DescribeTable(ctx, in.Table)
	if err != nil {
		return nil, schemaOutput{}, fmt.Errorf("describe table: %w", err)
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("## Table: %s\n\n", schema.Table))
	sb.WriteString("| Column | Type | Nullable | Key | Default |\n")
	sb.WriteString("|--------|------|----------|-----|---------|\n")
	for _, c := range schema.Columns {
		sb.WriteString(fmt.Sprintf("| %s | %s | %s | %s | %s |\n", c.Name, c.Type, c.Nullable, c.Key, c.Default))
	}

	if len(schema.Indexes) > 0 {
		sb.WriteString("\n**Indexes**: ")
		for i, idx := range schema.Indexes {
			if i > 0 {
				sb.WriteString(", ")
			}
			typ := "INDEX"
			if idx.Unique {
				typ = "UNIQUE"
			}
			sb.WriteString(fmt.Sprintf("%s(%s) %s", idx.Name, strings.Join(idx.Columns, ","), typ))
		}
		sb.WriteString("\n")
	}

	return nil, schemaOutput{Text: sb.String()}, nil
}

func (s *Server) getSQLDriver(dsName string) (driver.Driver, error) {
	d, err := s.mgr.Get(dsName)
	if err != nil {
		return nil, err
	}
	sqlDrv, ok := d.(driver.Driver)
	if !ok {
		return nil, fmt.Errorf("data source %s is not a SQL driver (type: %s)", dsName, d.Type())
	}
	return sqlDrv, nil
}

func formatQueryResult(r *driver.QueryResult) string {
	if len(r.Columns) == 0 {
		return "## Query Result\n\n(empty - no columns returned)"
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("## Query Result (%d rows)\n\n", len(r.Rows)))
	sb.WriteString("|")
	for _, c := range r.Columns {
		sb.WriteString(" " + c + " |")
	}
	sb.WriteString("\n|")
	for range r.Columns {
		sb.WriteString("---|")
	}
	sb.WriteString("\n")
	for _, row := range r.Rows {
		sb.WriteString("|")
		for _, v := range row {
			sb.WriteString(fmt.Sprintf(" %v |", v))
		}
		sb.WriteString("\n")
	}
	return sb.String()
}

