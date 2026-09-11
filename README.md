# mcp-dbx

通用数据库 MCP Server，让 AI 编程助手（kilocode / cursor / claude code 等）通过 MCP 协议安全操作数据库。

## 特性

- **MySQL + Redis** 开箱即用，架构可扩展国产数据库（OceanBase / 达梦 / 金仓）
- **安全层**：默认只读，写操作需显式配置；危险关键词拦截；DELETE/UPDATE 无 WHERE 拦截；行数限制 + 查询超时
- **stdio 传输**，kilocode / cursor 本地直连
- **YAML 配置**，多数据源命名管理，按名称路由
- **单二进制**，零依赖部署

## 快速开始

### 1. 编译

```bash
# 需 Go 1.25+
make build
# 或直接
go build -o bin/mcp-dbx ./cmd/mcp-dbx
```

跨平台编译：

```bash
# Windows
$env:GOOS="windows"; $env:GOARCH="amd64"; go build -o bin/mcp-dbx.exe ./cmd/mcp-dbx
# Linux
$env:GOOS="linux"; $env:GOARCH="amd64"; go build -o bin/mcp-dbx ./cmd/mcp-dbx
# macOS
$env:GOOS="darwin"; $env:GOARCH="arm64"; go build -o bin/mcp-dbx ./cmd/mcp-dbx
```

### 2. 编写配置

复制 `examples/mcp-dbx.yaml.example` 并修改：

```yaml
server:
  name: "mcp-dbx"
  version: "0.1.0"

safety:
  mode: "read-write"        # read-only | read-write
  max_rows: 1000
  query_timeout: 30s
  blocked_keywords: ["DROP", "TRUNCATE", "GRANT", "REVOKE", "ALTER"]
  blocked_commands: ["FLUSHALL", "FLUSHDB", "CONFIG", "SHUTDOWN", "KEYS"]

datasources:
  - name: "main-mysql"
    driver: "mysql"
    dsn: "user:password@tcp(127.0.0.1:3306)/mydb?charset=utf8mb4&parseTime=true"
    max_open_conns: 10
    max_idle_conns: 5
    conn_max_lifetime: 5m
    safety:
      mode: "read-only"     # 此数据源强制只读

  - name: "cache-redis"
    driver: "redis"
    addr: "127.0.0.1:6379"
    password: ""
    db: 0
    pool_size: 10
    safety:
      mode: "read-write"
```

### 3. 手动验证

```bash
# 启动并发送 MCP JSON-RPC 测试
echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-11-25","capabilities":{},"clientInfo":{"name":"test","version":"1.0"}}}' | ./bin/mcp-dbx --config ./examples/mcp-dbx.yaml.example
```

正常会返回 `initialize` 响应，包含 server capabilities。

### 4. 接入 AI 编程助手

#### kilocode

`.kilocode/mcp.json`：

```json
{
  "mcpServers": {
    "mcp-dbx": {
      "command": "/path/to/mcp-dbx",
      "args": ["--config", "/path/to/mcp-dbx.yaml"],
      "env": {}
    }
  }
}
```

#### cursor

`.cursor/mcp.json`：

```json
{
  "mcpServers": {
    "mcp-dbx": {
      "command": "/path/to/mcp-dbx",
      "args": ["--config", "/path/to/mcp-dbx.yaml"]
    }
  }
}
```

重启 IDE 后，AI 可直接调用数据库工具。

## 工具列表

共 12 个工具：

### 通用工具

| 工具 | 说明 |
|------|------|
| `db_list` | 列出所有已配置数据源及状态 |
| `db_ping` | 健康检查，返回延迟 |

### SQL 工具（MySQL / OceanBase / 达梦 / 金仓）

| 工具 | 说明 |
|------|------|
| `db_query` | 执行 SELECT，返回 Markdown 表格，最大 1000 行 |
| `db_execute` | 执行 INSERT/UPDATE/DELETE，返回影响行数 |
| `db_tables` | 列出所有表 |
| `db_schema` | 查看表结构（列/类型/索引） |

### Redis 工具

| 工具 | 说明 |
|------|------|
| `redis_get` | 读取键值 |
| `redis_set` | 写入键值（支持 TTL） |
| `redis_del` | 删除键 |
| `redis_keys` | SCAN 扫描键（非阻塞，限 100） |
| `redis_type` | 查看键类型 |
| `redis_ttl` | 查看过期时间 |

## 配置说明

### 配置文件查找顺序

1. `--config <path>` 命令行指定
2. `./mcp-dbx.yaml` 当前目录
3. `~/.mcp-dbx/config.yaml` 用户目录

### 安全策略优先级

```
数据源 safety > 全局 safety
```

数据源配置 `safety` 字段覆盖全局。未配置则继承全局。

### Redis 集群配置

```yaml
datasources:
  - name: "cluster-redis"
    driver: "redis"
    mode: "cluster"       # standalone(默认) | cluster | sentinel
    addrs:                # 集群用 addrs（复数）
      - "10.0.0.1:6379"
      - "10.0.0.2:6379"
      - "10.0.0.3:6379"
    password: ""
    pool_size: 20
```

## 安全策略

| 风险 | 防护 |
|------|------|
| 误删数据 | DELETE/UPDATE 无 WHERE 自动拦截 |
| 删表/清库 | DROP/TRUNCATE 默认拦截，可配置放行 |
| 全表扫描 | max_rows 限制 + 查询超时 |
| Redis KEYS 阻塞 | 强制用 SCAN 替代，KEYS 命令拦截 |
| 权限提升 | GRANT/REVOKE 默认拦截 |

安全层调用链：

```
工具 handler
  → safety.CheckWrite()           # 读写模式检查
  → safety.CheckSQL() / CheckRedisCommand()  # 危险词拦截
  → DELETE/UPDATE WHERE 检测
  → driver.Query/Execute(ctxWithTimeout)     # 超时控制
```

## 测试环境

本地 Docker 起 MySQL + Redis：

```bash
docker run -d --name mysql-test -e MYSQL_ROOT_PASSWORD=test -p 3306:3306 mysql:8
docker run -d --name redis-test -p 6379:6379 redis:7
```

## 项目结构

```
mcp_dbx/
├── cmd/mcp-dbx/main.go           — 入口
├── internal/
│   ├── config/                   — YAML 配置解析
│   ├── driver/                   — Driver 接口 + MySQL/Redis 实现
│   ├── datasource/               — 数据源管理器
│   ├── safety/                   — 安全层
│   └── mcp/                      — MCP server + 工具 handler
├── examples/                     — 配置示例
├── docs/plans/                   — 设计文档
└── Makefile
```

## 扩展新数据库

1. 写 `internal/driver/xxx/xxx.go` 实现 `Driver` 或 `NoSQLDriver` 接口
2. `init()` 中调用 `driver.Register("xxx", ...)`
3. `cmd/mcp-dbx/main.go` import `_ "github.com/yourname/mcp-dbx/internal/driver/xxx"`
4. 配置文件加 data source，`driver: xxx`

**MySQL 协议兼容的数据库**（OceanBase / TiDB）可直接复用 MySQL driver，仅改 driver 标识名即可。

## 开发

```bash
make build     # 编译
make test      # 测试
make run       # 编译+运行
make clean     # 清理
```

## 路线图

| 版本 | 内容 |
|------|------|
| v0.1.0 | MySQL + Redis + 安全层，stdio 传输 ✅ |
| v0.2.0 | OceanBase + 金仓支持，审计日志 |
| v0.3.0 | 达梦 DM8 支持 |
| v0.4.0 | HTTP/SSE 传输 |
| v0.5.0 | 连接池监控、慢查询日志 |

## 技术栈

- Go 1.25+
- [MCP Go SDK](https://github.com/modelcontextprotocol/go-sdk) v1.7.0（官方 SDK）
- [go-sql-driver/mysql](https://github.com/go-sql-driver/mysql) v1.10.1
- [go-redis/v9](https://github.com/redis/go-redis) v9.22.0
- `log/slog` 标准库结构化日志
