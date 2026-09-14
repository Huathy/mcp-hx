# mcp-x

A universal database MCP Server that lets AI coding assistants (kilocode / cursor / claude code, etc.) operate databases safely via the MCP protocol.

## Features

- **MySQL + Redis** out of the box; architecture extensible to domestic databases (OceanBase / DM / Kingbase)
- **Safety layer**: read-only by default; explicit config required for writes; dangerous keyword interception; DELETE/UPDATE without WHERE blocked; row limit + query timeout
- **stdio transport**, direct local connection for kilocode / cursor
- **YAML config**, multi-datasource named management, routing by name
- **Single binary**, zero-dependency deployment

## Quick Start

### 1. Build

```bash
# Requires Go 1.25+
make build
# or directly
go build -o bin/mcp-x ./cmd/mcp-x
```

Cross-platform build:

```bash
# Windows
$env:GOOS="windows"; $env:GOARCH="amd64"; go build -o bin/mcp-x.exe ./cmd/mcp-x
# Linux
$env:GOOS="linux"; $env:GOARCH="amd64"; go build -o bin/mcp-x ./cmd/mcp-x
# macOS
$env:GOOS="darwin"; $env:GOARCH="arm64"; go build -o bin/mcp-x ./cmd/mcp-x
```

### 2. Write Config

Copy `examples/mcp-x.yaml.example` and modify:

```yaml
server:
  name: "mcp-x"
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
      mode: "read-only"     # force read-only for this datasource

  - name: "cache-redis"
    driver: "redis"
    addr: "127.0.0.1:6379"
    password: ""
    db: 0
    pool_size: 10
    safety:
      mode: "read-write"
```

### 3. Manual Verification

```bash
# Start and send MCP JSON-RPC test
echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-11-25","capabilities":{},"clientInfo":{"name":"test","version":"1.0"}}}' | ./bin/mcp-x --config ./examples/mcp-x.yaml.example
```

A normal response returns `initialize` result with server capabilities.

### 4. Connect to AI Coding Assistant

#### kilocode

`.kilocode/mcp.json`:

```json
{
  "mcpServers": {
    "mcp-x": {
      "command": "/path/to/mcp-x",
      "args": ["--config", "/path/to/mcp-x.yaml"],
      "env": {}
    }
  }
}
```

#### cursor

`.cursor/mcp.json`:

```json
{
  "mcpServers": {
    "mcp-x": {
      "command": "/path/to/mcp-x",
      "args": ["--config", "/path/to/mcp-x.yaml"]
    }
  }
}
```

Restart the IDE, then AI can directly call database tools.

## Tools

12 tools total:

### Common Tools

| Tool | Description |
|------|-------------|
| `db_list` | List all configured datasources and status |
| `db_ping` | Health check, returns latency |

### SQL Tools (MySQL / OceanBase / DM / Kingbase)

| Tool | Description |
|------|-------------|
| `db_query` | Execute SELECT, returns Markdown table, max 1000 rows |
| `db_execute` | Execute INSERT/UPDATE/DELETE, returns affected rows |
| `db_tables` | List all tables |
| `db_schema` | View table structure (columns/types/indexes) |

### Redis Tools

| Tool | Description |
|------|-------------|
| `redis_get` | Read key value |
| `redis_set` | Write key value (supports TTL) |
| `redis_del` | Delete key |
| `redis_keys` | SCAN keys (non-blocking, max 100) |
| `redis_type` | Check key type |
| `redis_ttl` | Check TTL |

## Configuration

### Config File Lookup Order

1. `--config <path>` via command line
2. `./mcp-x.yaml` current directory
3. `~/.mcp-x/config.yaml` user directory

### Safety Priority

```
datasource safety > global safety
```

Datasource `safety` field overrides global. Unconfigured inherits global.

### Redis Cluster Config

```yaml
datasources:
  - name: "cluster-redis"
    driver: "redis"
    mode: "cluster"       # standalone(default) | cluster | sentinel
    addrs:                # cluster uses addrs (plural)
      - "10.0.0.1:6379"
      - "10.0.0.2:6379"
      - "10.0.0.3:6379"
    password: ""
    pool_size: 20
```

## Safety Policy

| Risk | Protection |
|------|------------|
| Accidental delete | DELETE/UPDATE without WHERE auto-blocked |
| Drop table / flush db | DROP/TRUNCATE blocked by default, configurable to allow |
| Full table scan | max_rows limit + query timeout |
| Redis KEYS blocking | Forced SCAN replacement, KEYS command blocked |
| Privilege escalation | GRANT/REVOKE blocked by default |

Safety layer call chain:

```
tool handler
  -> safety.CheckWrite()                     # read-write mode check
  -> safety.CheckSQL() / CheckRedisCommand() # dangerous keyword interception
  -> DELETE/UPDATE WHERE detection
  -> driver.Query/Execute(ctxWithTimeout)     # timeout control
```

## Test Environment

Start MySQL + Redis via local Docker:

```bash
docker run -d --name mysql-test -e MYSQL_ROOT_PASSWORD=test -p 3306:3306 mysql:8
docker run -d --name redis-test -p 6379:6379 redis:7
```

## UTF8 Multi-language Read/Write Verification

MySQL DSN uses `charset=utf8mb4`, fully supporting multi-language UTF8 read/write. Test table `zy_alarm`, column `err longtext`.

### Test Data (id 830-847)

| id | type | ok | err |
|----|------|----|-----|
| 830 | threshold | Y | NULL |
| 831 | timeout | N | connection timed out after 30s |
| 832 | offline | N | device heartbeat lost |
| 835 | threshold | N | 娓╁害瓒呰繃闃堝€?0搴?|
| 836 | offline | N | 璁惧鎺夌嚎锛氬績璺宠秴鏃舵湭鏀跺埌 |
| 837 | timeout | N | 缃戝叧杩炴帴瓒呮椂锛氱瓑寰?0绉掓棤鍝嶅簲 |
| 840 | threshold | N | 娓╁害銇屻仐銇嶃亜鍊?0搴︺倰瓒呫亪銇俱仐銇?|
| 841 | offline | N | 銉囥儛銈ゃ偣銈儠銉┿偆銉筹細銉忋兗銉堛儞銉笺儓銇屻偪銈ゃ儬銈偊銉堛仐銇俱仐銇?|
| 842 | timeout | N | 銈层兗銉堛偊銈с偆鎺ョ稓銈裤偆銉犮偄銈︺儓锛?0绉掑繙绛斻仾銇?|
| 843 | threshold | N | 鞓弰臧€ 鞛勱硠臧?80霃勲ゼ 齑堦臣頄堨姷雼堧嫟 |
| 844 | offline | N | 鞛レ箻 鞓ろ攧霛检澑锛氻晿韸鸽箘韸?鞁滉皠 齑堦臣 |
| 845 | timeout | N | 瓴岇澊韸胳洦鞚?鞐瓣舶 鞁滉皠 齑堦臣锛?0齑?鞚戨嫷 鞐嗢潓 |
| 846 | config | N | 妲嬫垚銈ㄣ儵銉硷細銈汇兂銈点兗銈层偆銉炽儜銉┿儭銉笺偪銇屼笉瓒?|
| 847 | config | N | 甑劚 鞓る锛氺劶靹?鞚措摑 毵り皽氤€靾?雸勲澖 |

### Supported Languages

- English: `connection timed out after 30s`
- Chinese: `娓╁害瓒呰繃闃堝€?0搴
- Japanese: `娓╁害銇屻仐銇嶃亜鍊?0搴︺倰瓒呫亪銇俱仐銇焋
- Korean: `鞓弰臧€ 鞛勱硠臧?80霃勲ゼ 齑堦臣頄堨姷雼堧嫟`
- Mixed punctuation: `鞛レ箻 鞓ろ攧霛检澑锛氻晿韸鸽箘韸?鞁滉皠 齑堦臣` (fullwidth colon)

### Key Points

- DSN must use `charset=utf8mb4` (not `utf8`; the latter only supports BMP 3-byte, missing some emoji and rare chars)
- MySQL table/column `COLLATE` recommended: `utf8mb4_general_ci` or `utf8mb4_unicode_ci`
- MCP JSON-RPC over stdio transmits UTF8 natively, no extra encoding needed

## Project Structure

```
mcp_dbx/
鈹溾攢鈹€ cmd/mcp-x/main.go           鈥?entry point
鈹溾攢鈹€ internal/
鈹?  鈹溾攢鈹€ config/                   鈥?YAML config parsing
鈹?  鈹溾攢鈹€ driver/                   鈥?Driver interface + MySQL/Redis impl
鈹?  鈹溾攢鈹€ datasource/               鈥?datasource manager
鈹?  鈹溾攢鈹€ safety/                   鈥?safety layer
鈹?  鈹斺攢鈹€ mcp/                      鈥?MCP server + tool handlers
鈹溾攢鈹€ examples/                     鈥?config examples
鈹溾攢鈹€ docs/plans/                   鈥?design docs
鈹斺攢鈹€ Makefile
```

## Extending New Databases

1. Write `internal/driver/xxx/xxx.go` implementing `Driver` or `NoSQLDriver` interface
2. In `init()`, call `driver.Register("xxx", ...)`
3. Import in `cmd/mcp-x/main.go`: `_ "github.com/yourname/mcp-x/internal/driver/xxx"`
4. Add datasource in config, `driver: xxx`

**MySQL-protocol-compatible databases** (OceanBase / TiDB) can reuse the MySQL driver directly; just change the driver label.

## Development

```bash
make build     # build
make test      # test
make run       # build + run
make clean     # clean
```

## Roadmap

| Version | Content |
|---------|---------|
| v0.1.0 | MySQL + Redis + safety layer, stdio transport 鉁?|
| v0.2.0 | OceanBase + Kingbase support, audit log |
| v0.3.0 | DM8 support |
| v0.4.0 | HTTP/SSE transport |
| v0.5.0 | Connection pool monitoring, slow query log |

## Tech Stack

- Go 1.25+
- [MCP Go SDK](https://github.com/modelcontextprotocol/go-sdk) v1.7.0 (official SDK)
- [go-sql-driver/mysql](https://github.com/go-sql-driver/mysql) v1.10.1
- [go-redis/v9](https://github.com/redis/go-redis) v9.22.0
- `log/slog` standard library structured logging
