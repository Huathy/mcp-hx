package postgres

import (
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/yourname/mcp-x/internal/driver"
	"github.com/yourname/mcp-x/internal/driver/pglike"
)

type PostgresDriver struct {
	*pglike.Driver
}

func (d *PostgresDriver) Name() string { return "postgres" }

func init() {
	driver.Register("postgres", func() driver.AnyDriver {
		return &PostgresDriver{Driver: pglike.New("pgx", "")}
	})
}
