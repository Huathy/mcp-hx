# 03 配置规范

## 1. 配置文件格式

YAML 格式，单文件管理所有数据源和安全策略。

### 文件位置

默认路径（按优先级查找）：

1. 命令行 `--config <path>` 指定
2. 当前目录 `./mcp-dbx.yaml`
3. 用户目录 `~/.mcp-dbx/config.yaml`

### 完整配置示例

```yaml
# mcp-dbx.yaml

# MCP Server 信息
server:
  name: "mcp-dbx"
  version: "0.1.0"

# 安全策略（全局默认，可被数据源覆盖）
safety:
  mode: "read-write"          # read-only | read-write
  max_rows: 1000              # 单次查询最大返回行数
  query_timeout: 30s          # 单次查询超时
  blocked_keywords:           # 拦截的 SQL 关键词
    - "DROP"
    - "TRUNCATE"
    - "GRANT"
    - "REVOKE"
    - "ALTER"
  allow_blocked_keywords: []  # 显式放行的关键词（覆盖 blocked）
  blocked_commands:           # Redis 拦截的命令
    - "FLUSHALL"
    - "FLUSHDB"
    - "CONFIG"
    - "SHUTDOWN"
    - "KEYS"                  # 强制用 SCAN 替代
  allow_blocked_commands: []

# 数据源列表
datasources:
  # MySQL 示例
  - name: "main-mysql"
    driver: "mysql"
    dsn: "user:password@tcp(127.0.0.1:3306)/mydb?charset=utf8mb4&parseTime=true"
    max_open_conns: 10
    max_idle_conns: 5
    conn_max_lifetime: 5m
    safety:                   # 覆盖全局安全策略（可选）
      mode: "read-only"       # 这个数据源强制只读

  # Redis 示例
  - name: "cache-redis"
    driver: "redis"
    addr: "127.0.0.1:6379"
    password: ""
    db: 0
    pool_size: 10
    safety:
      mode: "read-write"

  # Redis 集群示例
  - name: "cluster-redis"
    driver: "redis"
    mode: "cluster"           # standalone | cluster | sentinel
    addrs:                    # 集群模式用 addrs（复数）
      - "10.0.0.1:6379"
      - "10.0.0.2:6379"
      - "10.0.0.3:6379"
    password: ""
    pool_size: 20

  # OceanBase 示例（复用 mysql 驱动，改 driver 名）
  - name: "ob-prod"
    driver: "oceanbase"       # 内部映射到 mysql 驱动，仅标识区分
    dsn: "user:pass@tcp(ob-host:2883)/db?charset=utf8mb4"
    safety:
      mode: "read-only"

  # 达梦 DM8 示例（未来扩展）
  - name: "dm-finance"
    driver: "dameng"
    dsn: "dm://user:pass@127.0.0.1:5236/SCHEMA"
    safety:
      mode: "read-only"
```

## 2. 配置字段说明

### 2.1 `server`

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| name | string | 是 | MCP server 名称 |
| version | string | 是 | 版本号 |

### 2.2 `safety`

| 字段 | 类型 | 默认 | 说明 |
|------|------|------|------|
| mode | string | `read-write` | `read-only` 禁止所有写操作；`read-write` 允许写（受关键词拦截限制） |
| max_rows | int | 1000 | `db_query` 返回最大行数 |
| query_timeout | duration | 30s | 单次操作超时 |
| blocked_keywords | []string | 见示例 | 拦截的 SQL 关键词，大小写不敏感 |
| allow_blocked_keywords | []string | `[]` | 放行的关键词，覆盖 blocked |
| blocked_commands | []string | 见示例 | 拦截的 Redis 命令 |
| allow_blocked_commands | []string | `[]` | 放行的命令 |

### 2.3 `datasources[]`

通用字段：

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| name | string | 是 | 数据源名称，工具调用时用此名称路由 |
| driver | string | 是 | `mysql` / `redis` / `oceanbase` / `dameng` / `kingbase` |
| safety | object | 否 | 数据源级安全策略，覆盖全局 |

MySQL/OceanBase/达梦/金仓（SQL 类）特有：

| 字段 | 类型 | 说明 |
|------|------|------|
| dsn | string | 标准 DSN 连接串 |
| max_open_conns | int | 最大连接数 |
| max_idle_conns | int | 最大空闲连接 |
| conn_max_lifetime | duration | 连接最大存活时间 |

Redis 特有：

| 字段 | 类型 | 说明 |
|------|------|------|
| mode | string | `standalone`(默认) / `cluster` / `sentinel` |
| addr | string | 单机地址 |
| addrs | []string | 集群地址列表（cluster/sentinel 模式） |
| password | string | 密码 |
| db | int | 数据库编号 |
| pool_size | int | 连接池大小 |

## 3. 敏感信息处理

### DSN 密码脱敏

配置文件中密码明文存储。日志输出和 MCP 返回中自动脱敏：

```
原始：user:s3cr3t@tcp(host)/db
日志：user:****@tcp(host)/db
```

### 环境变量引用（预留）

> ponytail: 当前版本用纯配置文件。如未来需要环境变量引用，支持 `${ENV_VAR}` 语法替换。加当需要从 CI/CD 注入密码时。

## 4. kilocode 接入配置

`.kilocode/mcp.json`（或 Settings → MCP）中添加：

```json
{
  "mcpServers": {
    "mcp-dbx": {
      "command": "mcp-dbx",
      "args": ["--config", "/path/to/mcp-dbx.yaml"],
      "env": {}
    }
  }
}
```

**前提**：`mcp-dbx` 可执行文件在 PATH 中，或用绝对路径。

## 5. cursor 接入配置

`.cursor/mcp.json`：

```json
{
  "mcpServers": {
    "mcp-dbx": {
      "command": "mcp-dbx",
      "args": ["--config", "/path/to/mcp-dbx.yaml"]
    }
  }
}
```

格式与 kilocode 一致，MCP 协议标准。
