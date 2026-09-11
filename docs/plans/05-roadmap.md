# 05 开发计划

## 1. 里程碑概览

| 阶段 | 目标 | 预估工时 |
|------|------|---------|
| M1: 骨架 | 项目初始化 + 配置层 + Driver 接口 | 1-2 天 |
| M2: MySQL | MySQL Driver + SQL 工具 | 2-3 天 |
| M3: Redis | Redis Driver + Redis 工具 | 2 天 |
| M4: 安全层 | 读写分离 + 危险词拦截 | 1-2 天 |
| M5: 集成测试 | kilocode/cursor 联调 | 1 天 |
| M6: 国产数据库 | 达梦/OceanBase 扩展 | 2-3 天 |

总计约 9-13 天（业余开发节奏）。

## 2. M1: 项目骨架

### 2.1 初始化

```bash
cd D:\works\mystudy\my_open\mcp_dbx
go mod init github.com/yourname/mcp-dbx
go get github.com/modelcontextprotocol/go-sdk/mcp
go get gopkg.in/yaml.v3
```

### 2.2 交付物

- [ ] `go.mod` 初始化
- [ ] `internal/config/config.go` — YAML 配置解析
- [ ] `internal/driver/driver.go` — Driver 接口定义
- [ ] `internal/driver/registry.go` — 驱动注册表
- [ ] `internal/datasource/manager.go` — 数据源管理器
- [ ] `internal/mcp/server.go` — MCP server 骨架（空工具）
- [ ] `cmd/mcp-dbx/main.go` — 入口，加载配置启动 server
- [ ] `examples/mcp-dbx.yaml.example` — 配置示例
- [ ] 能启动 stdio server，响应 `tools/list`（返回空列表）

### 2.3 验证

```powershell
# 构建
go build -o bin\mcp-dbx.exe .\cmd\mcp-dbx

# 启动测试（手动发 JSON-RPC）
echo '{"jsonrpc":"2.0","id":1,"method":"tools/list"}' | .\bin\mcp-dbx.exe --config .\examples\mcp-dbx.yaml.example
```

## 3. M2: MySQL Driver + SQL 工具

### 3.1 依赖

```bash
go get github.com/go-sql-driver/mysql
```

### 3.2 交付物

- [ ] `internal/driver/mysql/mysql.go` — 实现 Driver 接口
- [ ] `internal/mcp/tools.go` — 注册 `db_list`, `db_ping`, `db_query`, `db_execute`, `db_tables`, `db_schema`
- [ ] `internal/mcp/tools_sql.go` — SQL 工具 handler
- [ ] 单元测试：`internal/driver/mysql/mysql_test.go`（用 SQLite 或内存库 mock，或跳过集成测试）

### 3.3 MySQL Driver 实现要点

```go
type MySQLDriver struct {
    db *sql.DB
}

func (d *MySQLDriver) Connect(ctx context.Context, cfg ConnConfig) error {
    db, err := sql.Open("mysql", cfg.DSN)
    // 设置连接池参数
    db.SetMaxOpenConns(cfg.MaxOpenConns)
    db.SetMaxIdleConns(cfg.MaxIdleConns)
    db.SetConnMaxLifetime(cfg.ConnMaxLifetime)
    d.db = db
    return db.PingContext(ctx)
}

func (d *MySQLDriver) Query(ctx context.Context, sqlStr string, args []any) (*QueryResult, error) {
    rows, err := d.db.QueryContext(ctx, sqlStr, args...)
    // ...
    // 读取列名
    cols, _ := rows.Columns()
    // 逐行扫描，interface{} + []byte 处理
    // 返回 QueryResult{Columns, Rows}
}

func (d *MySQLDriver) ListTables(ctx context.Context) ([]TableInfo, error) {
    return d.Query(ctx, "SHOW TABLES", nil)
    // 或用 information_schema 更通用
}
```

### 3.4 验证

配置真实 MySQL 连接，通过 MCP inspector 或手动 JSON-RPC 调用 `db_query` 执行 `SELECT 1`。

## 4. M3: Redis Driver

### 4.1 依赖

```bash
go get github.com/redis/go-redis/v9
```

### 4.2 交付物

- [ ] `internal/driver/redis/redis.go` — 实现 NoSQLDriver 接口
- [ ] `internal/mcp/tools_redis.go` — 注册 `redis_get/set/del/keys/type/ttl`
- [ ] 单元测试

### 4.3 Redis Driver 要点

```go
type RedisDriver struct {
    client redis.UniversalClient  // 兼容单机/集群
}

func (d *RedisDriver) Connect(ctx context.Context, cfg ConnConfig) error {
    if cfg.Mode == "cluster" {
        d.client = redis.NewClusterClient(&redis.ClusterOptions{
            Addrs: cfg.Addrs,
            Password: cfg.Password,
            PoolSize: cfg.PoolSize,
        })
    } else {
        d.client = redis.NewClient(&redis.Options{
            Addr: cfg.Addr,
            Password: cfg.Password,
            DB: cfg.DB,
            PoolSize: cfg.PoolSize,
        })
    }
    return d.client.Ping(ctx).Err()
}
```

## 5. M4: 安全层

### 5.1 交付物

- [ ] `internal/safety/checker.go` — 读写模式检查
- [ ] `internal/safety/interceptor.go` — 危险词拦截
- [ ] `internal/safety/limiter.go` — 行数/超时限制
- [ ] `internal/safety/delete_check.go` — DELETE/UPDATE 无 WHERE 检测
- [ ] 集成到工具 handler 调用链

### 5.2 调用链

```
tool handler
  → safety.CheckMode(datasource, isWrite)
  → safety.CheckSQL(sql) 或 safety.CheckRedisCommand(cmd)
  → safety.CheckDeleteWhere(sql)
  → driver.Query/Execute(ctxWithTimeout, sql, args)
```

## 6. M5: 集成测试

### 6.1 交付物

- [ ] 配置 kilocode `.kilocode/mcp.json`
- [ ] 配置 cursor `.cursor/mcp.json`
- [ ] 端到端测试：AI 调用 `db_query` 查询 → 返回表格
- [ ] 端到端测试：AI 调用 `redis_set` → `redis_get` 验证
- [ ] 安全拦截测试：read-only 模式下 `db_execute` 被拒
- [ ] 危险词测试：`DROP TABLE` 被拦截

### 6.2 测试环境

本地 Docker 起 MySQL + Redis：

```powershell
docker run -d --name mysql-test -e MYSQL_ROOT_PASSWORD=test -p 3306:3306 mysql:8
docker run -d --name redis-test -p 6379:6379 redis:7
```

## 7. M6: 国产数据库扩展

### 7.1 OceanBase（最简单）

OceanBase 兼容 MySQL 协议，直接复用 MySQL 驱动：

```go
// internal/driver/oceanbase/oceanbase.go
// 实质是 MySQLDriver 的别名包装
type OceanBaseDriver struct {
    MySQLDriver
}

func (d *OceanBaseDriver) Name() string { return "oceanbase" }
```

或更简单：在 driver registry 中把 `oceanbase` 映射到 MySQLDriver 实例。

### 7.2 达梦 DM8

需要达梦官方 Go 驱动：

```go
import "dameng.com/dm"  // 从安装包 drivers/go 获取

// 实现标准 database/sql 接口
```

**挑战**：达梦驱动不在公共 Go module proxy，需要：
- 方案 A：把驱动源码 vendor 到 `vendor/dameng/`
- 方案 B：用 `gorm-dameng`（社区封装，在 GitHub）

### 7.3 金仓 Kingbase

Kingbase 兼容 PostgreSQL 协议：

```go
import _ "github.com/lib/pq"  // 或 github.com/jackc/pgx/v5/stdlib
// DSN 用 postgres:// 格式
```

## 8. Makefile

```makefile
.PHONY: build test run clean

build:
	go build -o bin/mcp-dbx ./cmd/mcp-dbx

run: build
	./bin/mcp-dbx --config ./examples/mcp-dbx.yaml.example

test:
	go test ./... -v

clean:
	rm -rf bin/
```

## 9. 发布

### 9.1 跨平台构建

```powershell
# Windows
$env:GOOS="windows"; $env:GOARCH="amd64"; go build -o bin/mcp-dbx-windows-amd64.exe ./cmd/mcp-dbx
# Linux
$env:GOOS="linux"; $env:GOARCH="amd64"; go build -o bin/mcp-dbx-linux-amd64 ./cmd/mcp-dbx
# macOS
$env:GOOS="darwin"; $env:GOARCH="arm64"; go build -o bin/mcp-dbx-darwin-arm64 ./cmd/mcp-dbx
```

### 9.2 GitHub Release

- 打 tag `v0.1.0`
- GitHub Actions 自动构建三平台二进制
- Release 附带配置示例和 kilocode 接入说明

## 10. 版本规划

| 版本 | 内容 |
|------|------|
| v0.1.0 | MySQL + Redis + 基础安全层，stdio 传输 |
| v0.2.0 | OceanBase + 金仓支持，审计日志 |
| v0.3.0 | 达梦 DM8 支持 |
| v0.4.0 | HTTP/SSE 传输 |
| v0.5.0 | 连接池监控、慢查询日志 |
