package datasource

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/yourname/mcp-dbx/internal/config"
	"github.com/yourname/mcp-dbx/internal/driver"
)

type Manager struct {
	sources map[string]driver.AnyDriver
	logger  *slog.Logger
}

func NewManager(cfgs []config.DataSourceConfig, logger *slog.Logger) (*Manager, error) {
	m := &Manager{
		sources: make(map[string]driver.AnyDriver),
		logger:  logger,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	for _, dc := range cfgs {
		d, err := driver.New(dc.Driver)
		if err != nil {
			return nil, fmt.Errorf("datasource %s: %w", dc.Name, err)
		}
		connCfg := driver.ConnConfig{
			DSN:            dc.DSN,
			MaxOpenConns:   dc.MaxOpenConns,
			MaxIdleConns:   dc.MaxIdleConns,
			ConnMaxLifetime: int64(dc.ConnMaxLifetime.Std()),
			Addr:           dc.Addr,
			Addrs:          dc.Addrs,
			Password:       dc.Password,
			DB:             dc.DB,
			Mode:           dc.Mode,
			PoolSize:       dc.PoolSize,
		}
		if err := d.Connect(ctx, connCfg); err != nil {
			return nil, fmt.Errorf("connect datasource %s: %w", dc.Name, err)
		}
		m.sources[dc.Name] = d
		m.logger.Info("datasource connected", "name", dc.Name, "driver", dc.Driver)
	}
	return m, nil
}

func (m *Manager) Get(name string) (driver.AnyDriver, error) {
	d, ok := m.sources[name]
	if !ok {
		return nil, fmt.Errorf("data source not found: %s", name)
	}
	return d, nil
}

func (m *Manager) Names() []string {
	names := make([]string, 0, len(m.sources))
	for name := range m.sources {
		names = append(names, name)
	}
	return names
}

func (m *Manager) Close() error {
	var firstErr error
	for name, d := range m.sources {
		if err := d.Close(); err != nil {
			m.logger.Error("close datasource failed", "name", name, "err", err)
			if firstErr == nil {
				firstErr = err
			}
		}
	}
	return firstErr
}
