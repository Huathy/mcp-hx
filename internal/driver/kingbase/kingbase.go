package kingbase

import (
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/yourname/mcp-x/internal/driver"
	"github.com/yourname/mcp-x/internal/driver/pglike"
)

type KingbaseDriver struct {
	*pglike.Driver
}

func (d *KingbaseDriver) Name() string { return "kingbase" }

func init() {
	driver.Register("kingbase", func() driver.AnyDriver {
		return &KingbaseDriver{Driver: pglike.New("pgx", "")}
	})
}
