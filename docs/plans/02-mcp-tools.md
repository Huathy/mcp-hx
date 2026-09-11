# 02 MCP 工具定义

## 1. 工具设计原则

- 工具粒度适中：不过细（AI 拼装太累），不过粗（一个工具干所有事）
- 参数用 JSON Schema 严格定义，让 AI 明确需要传什么
- 返回统一 Markdown 文本格式，便于 AI 阅读和转述
- 每个工具有明确的数据源路由参数 `datasource`

## 2. 通用工具（所有数据源）

### 2.1 `db_list` — 列出数据源

列出所有已配置的数据源及其类型。

```json
{
  "name": "db_list",
  "description": "List all configured data sources and their types (mysql/redis/etc). Use this first to know available databases.",
  "inputSchema": {
    "type": "object",
    "properties": {},
    "additionalProperties": false
  }
}
```

**返回示例**：
```text
## Data Sources

| Name       | Type  | Driver | Status |
|------------|-------|--------|--------|
| main-mysql | sql   | mysql  | ok     |
| cache-redis | nosql | redis  | ok     |
```

### 2.2 `db_ping` — 健康检查

```json
{
  "name": "db_ping",
  "description": "Ping a data source to check connectivity. Returns latency in ms.",
  "inputSchema": {
    "type": "object",
    "properties": {
      "datasource": {
        "type": "string",
        "description": "Data source name from db_list"
      }
    },
    "required": ["datasource"]
  }
}
```

## 3. SQL 类工具（MySQL / 达梦 / OceanBase / TiDB / Kingbase）

### 3.1 `db_query` — 只读查询

执行 SELECT 语句，受只读模式保护。

```json
{
  "name": "db_query",
  "description": "Execute a read-only SQL query (SELECT) and return results as a table. Only SELECT statements allowed. Max 1000 rows by default.",
  "inputSchema": {
    "type": "object",
    "properties": {
      "datasource": {
        "type": "string",
        "description": "Data source name"
      },
      "sql": {
        "type": "string",
        "description": "SQL SELECT query. Supports parameterized placeholders (?)"
      },
      "args": {
        "type": "array",
        "items": {},
        "description": "Query parameters for placeholders"
      },
      "limit": {
        "type": "integer",
        "description": "Max rows to return (default 1000, hard cap from config)",
        "default": 1000
      }
    },
    "required": ["datasource", "sql"]
  }
}
```

**返回示例**：
```text
## Query Result (3 rows, 12ms)

| id | name    | email          |
|----|---------|----------------|
| 1  | Alice   | alice@mail.com |
| 2  | Bob     | bob@mail.com   |
| 3  | Carol   | carol@mail.com |
```

### 3.2 `db_execute` — 写操作

执行 INSERT/UPDATE/DELETE，受写模式和危险词拦截保护。

```json
{
  "name": "db_execute",
  "description": "Execute a write SQL statement (INSERT/UPDATE/DELETE). Requires write mode enabled in config. DDL (CREATE/DROP/TRUNCATE) blocked unless explicitly allowed.",
  "inputSchema": {
    "type": "object",
    "properties": {
      "datasource": {
        "type": "string",
        "description": "Data source name"
      },
      "sql": {
        "type": "string",
        "description": "SQL write statement. Supports parameterized placeholders (?)."
      },
      "args": {
        "type": "array",
        "items": {},
        "description": "Query parameters for placeholders"
      }
    },
    "required": ["datasource", "sql"]
  }
}
```

**返回示例**：
```text
## Execute Result
- Affected rows: 5
- Duration: 3ms
- Last insert ID: 0
```

### 3.3 `db_tables` — 列出表

```json
{
  "name": "db_tables",
  "description": "List all tables in the database. Use to explore schema before writing queries.",
  "inputSchema": {
    "type": "object",
    "properties": {
      "datasource": {
        "type": "string",
        "description": "Data source name"
      }
    },
    "required": ["datasource"]
  }
}
```

### 3.4 `db_schema` — 查看表结构

```json
{
  "name": "db_schema",
  "description": "Describe table schema: columns, types, nullable, keys, defaults.",
  "inputSchema": {
    "type": "object",
    "properties": {
      "datasource": {
        "type": "string",
        "description": "Data source name"
      },
      "table": {
        "type": "string",
        "description": "Table name to inspect"
      }
    },
    "required": ["datasource", "table"]
  }
}
```

**返回示例**：
```text
## Table: users

| Column    | Type         | Nullable | Key     | Default |
|-----------|-------------|----------|---------|---------|
| id        | BIGINT       | NO       | PRI     | auto    |
| name      | VARCHAR(255) | NO       |         |         |
| email     | VARCHAR(255) | YES      | UNI     |         |
| created_at| TIMESTAMP    | NO       |         | now()   |

**Indexes**: PRIMARY(id), UNIQUE(email), INDEX(name)
```

## 4. Redis 工具

### 4.1 `redis_get` — 读取键

```json
{
  "name": "redis_get",
  "description": "Get value of a Redis key. Returns value and type.",
  "inputSchema": {
    "type": "object",
    "properties": {
      "datasource": { "type": "string", "description": "Redis data source name" },
      "key": { "type": "string", "description": "Redis key" }
    },
    "required": ["datasource", "key"]
  }
}
```

### 4.2 `redis_set` — 写入键

```json
{
  "name": "redis_set",
  "description": "Set a Redis key-value pair. Requires write mode.",
  "inputSchema": {
    "type": "object",
    "properties": {
      "datasource": { "type": "string" },
      "key": { "type": "string" },
      "value": { "type": "string" },
      "ttl": { "type": "integer", "description": "TTL in seconds. 0 = no expiry." }
    },
    "required": ["datasource", "key", "value"]
  }
}
```

### 4.3 `redis_del` — 删除键

```json
{
  "name": "redis_del",
  "description": "Delete a Redis key. Requires write mode.",
  "inputSchema": {
    "type": "object",
    "properties": {
      "datasource": { "type": "string" },
      "key": { "type": "string" }
    },
    "required": ["datasource", "key"]
  }
}
```

### 4.4 `redis_keys` — 扫描键

```json
{
  "name": "redis_keys",
  "description": "Scan Redis keys matching a pattern. Uses SCAN (non-blocking). Limited to 100 keys by default.",
  "inputSchema": {
    "type": "object",
    "properties": {
      "datasource": { "type": "string" },
      "pattern": { "type": "string", "description": "Glob pattern, e.g. user:*", "default": "*" },
      "limit": { "type": "integer", "default": 100 }
    },
    "required": ["datasource"]
  }
}
```

### 4.5 `redis_type` — 查看键类型

```json
{
  "name": "redis_type",
  "description": "Get the data type of a Redis key (string/hash/list/set/zset/stream).",
  "inputSchema": {
    "type": "object",
    "properties": {
      "datasource": { "type": "string" },
      "key": { "type": "string" }
    },
    "required": ["datasource", "key"]
  }
}
```

### 4.6 `redis_ttl` — 查看过期时间

```json
{
  "name": "redis_ttl",
  "description": "Get remaining TTL of a Redis key in seconds. Returns -1 if no expiry, -2 if key doesn't exist.",
  "inputSchema": {
    "type": "object",
    "properties": {
      "datasource": { "type": "string" },
      "key": { "type": "string" }
    },
    "required": ["datasource", "key"]
  }
}
```

## 5. 工具返回格式规范

所有工具返回统一 Markdown 文本格式（MCP `TextContent`）：

| 场景 | 格式 |
|------|------|
| 查询结果 | Markdown 表格 |
| 写操作结果 | 键值对文本 |
| 列表 | Markdown 表格或列表 |
| 错误 | `## Error\n\n**Reason**: ...\n\n**Detail**: ...` |

不用 JSON：AI 阅读表格比 JSON 更高效，节省 token。

## 6. 工具数量控制

初期 12 个工具：
- 通用 2 个：`db_list`, `db_ping`
- SQL 4 个：`db_query`, `db_execute`, `db_tables`, `db_schema`
- Redis 6 个：`redis_get`, `redis_set`, `redis_del`, `redis_keys`, `redis_type`, `redis_ttl`

> ponytail: 未来如果 Redis 需要操作 Hash/List/Set 等复杂数据结构，可加 `redis_hgetall`、`redis_lrange` 等，但初期保持精简。加当 AI 需要操作复杂数据结构时再扩展。
