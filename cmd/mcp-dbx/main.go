package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/yourname/mcp-dbx/internal/config"
	"github.com/yourname/mcp-dbx/internal/datasource"
	mcpserver "github.com/yourname/mcp-dbx/internal/mcp"

	_ "github.com/yourname/mcp-dbx/internal/driver/mysql"
	_ "github.com/yourname/mcp-dbx/internal/driver/redis"
)

func main() {
	var configPath string
	flag.StringVar(&configPath, "config", "", "path to config file (default: ./mcp-dbx.yaml)")
	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	if configPath == "" {
		configPath = findConfig()
	}
	if configPath == "" {
		logger.Error("未找到配置文件，请先配置 mcp-dbx",
			"searched",
			"./mcp-dbx.yaml, ./mcp-dbx.yml, ./.kilo/mcp-dbx.yaml, ./.kilo/mcp-dbx.yml, ~/.mcp-dbx/config.yaml",
		)
		fmt.Fprintln(os.Stderr, "\n[提示] 请先创建配置文件，任选其一：")
		fmt.Fprintln(os.Stderr, "  1. 项目级: ./.kilo/mcp-dbx.yaml  (推荐，仅当前项目)")
		fmt.Fprintln(os.Stderr, "  2. 项目级: ./mcp-dbx.yaml")
		fmt.Fprintln(os.Stderr, "  3. 全局级: ~/.mcp-dbx/config.yaml (所有项目共享)")
		fmt.Fprintln(os.Stderr, "\n示例配置见: examples/mcp-dbx.yaml.example")
		os.Exit(1)
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		logger.Error("load config failed", "err", err, "path", configPath)
		os.Exit(1)
	}
	logger.Info("config loaded", "path", configPath, "datasources", len(cfg.DataSources))

	mgr, err := datasource.NewManager(cfg.DataSources, logger)
	if err != nil {
		logger.Error("init datasource manager failed", "err", err)
		os.Exit(1)
	}
	defer mgr.Close()

	srv, err := mcpserver.NewServer(cfg.Server.Name, cfg.Server.Version, mgr, cfg, logger)
	if err != nil {
		logger.Error("create mcp server failed", "err", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigCh
		logger.Info("signal received, shutting down", "signal", sig)
		cancel()
	}()

	if err := srv.Run(ctx); err != nil {
		logger.Error("mcp server stopped", "err", err)
		os.Exit(1)
	}
}

func findConfig() string {
	candidates := []string{
		"./mcp-dbx.yaml",
		"./mcp-dbx.yml",
		"./.kilo/mcp-dbx.yaml",
		"./.kilo/mcp-dbx.yml",
	}
	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	home, err := os.UserHomeDir()
	if err == nil {
		for _, p := range []string{
			home + "/.mcp-dbx/config.yaml",
			home + "/.mcp-dbx/config.yml",
		} {
			if _, err := os.Stat(p); err == nil {
				return p
			}
		}
	}
	return ""
}
