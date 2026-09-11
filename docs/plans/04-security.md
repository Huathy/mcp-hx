# 04 安全策略

## 1. 威胁模型

AI 编程助手操作数据库的风险：

| 风险 | 场景 | 严重度 |
|------|------|--------|
| 误删数据 | AI 执行 DELETE 无 WHERE | 高 |
| 删表 | AI 执行 DROP TABLE | 致命 |
| 清库 | AI 执行 TRUNCATE / FLUSHALL | 致命 |
| 全表扫描 | AI 执行 SELECT * 无 LIMIT，打爆性能 | 中 |
| 权限提升 | AI 执行 GRANT 给自己加权限 | 高 |
| 配置篡改 | AI 执行 Redis CONFIG 改密码 | 致命 |
| KEYS 命令 | AI 执行 KEYS * 阻塞 Redis | 高 |

## 2. 安全层架构

```
工具调用入口
  │
  ▼
┌─────────────────┐
│ 1. 读写模式检查  │  ← mode: read-only 时拒绝所有写操作
└────────┬────────┘
         │ 通过
         ▼
┌─────────────────┐
│ 2. 危险词拦截    │  ← 检测 DROP/TRUNCATE/FLUSHALL 等
└────────┬────────┘
         │ 通过
         ▼
┌─────────────────┐
│ 3. 行数限制      │  ← SELECT 结果截断到 max_rows
└────────┬────────┘
         │ 通过
         ▼
┌─────────────────┐
│ 4. 超时控制      │  ← context.WithTimeout(query_timeout)
└────────┬────────┘
         │ 通过
         ▼
   执行操作，返回结果
```

## 3. 读写模式

### 3.1 `read-only` 模式

- SQL：只允许 `SELECT` 开头的语句
- Redis：只允许 `GET`、`TYPE`、`TTL`、`SCAN`（等价于只读工具）
- 拦截所有 `db_execute`、`redis_set`、`redis_del` 工具调用

### 3.2 `read-write` 模式

- 允许写操作，但受危险词拦截限制
- 仍拦截 DDL（DROP/TRUNCATE/ALTER）除非 `allow_blocked_keywords` 显式放行

### 3.3 SQL 语句类型判断

```go
// internal/safety/checker.go

func isReadOnlySQL(sql string) bool {
    // 去除前导空格和注释
    trimmed := strings.TrimSpace(sql)
    upper := strings.ToUpper(trimmed)
    // 只允许 SELECT（含 WITH ... SELECT）
    return strings.HasPrefix(upper, "SELECT") ||
           strings.HasPrefix(upper, "WITH")
}
```

### 3.4 模式优先级

```
数据源 safety.mode > 全局 safety.mode
```

数据源配置覆盖全局。未配数据源 safety 则继承全局。

## 4. 危险词拦截

### 4.1 SQL 关键词拦截

```go
// internal/safety/interceptor.go

type Interceptor struct {
    blockedKeywords  map[string]bool  // 大写
    allowKeywords    map[string]bool  // 大写，覆盖 blocked
}

func (i *Interceptor) CheckSQL(sql string) error {
    upper := strings.ToUpper(sql)
    for kw := range i.blockedKeywords {
        if containsWord(upper, kw) && !i.allowKeywords[kw] {
            return fmt.Errorf("blocked keyword '%s' in SQL. Add to allow_blocked_keywords to override", kw)
        }
    }
    return nil
}

// containsWord 检测完整单词匹配，避免 SUBSTRING 误判 SUB
func containsWord(text, word string) bool {
    // 用正则 \b 边界匹配
    re := regexp.MustCompile(`\b` + word + `\b`)
    return re.MatchString(text)
}
```

默认拦截词：

```
DROP, TRUNCATE, GRANT, REVOKE, ALTER, SHUTDOWN
```

### 4.2 Redis 命令拦截

```go
func (i *Interceptor) CheckRedisCommand(cmd string) error {
    upper := strings.ToUpper(strings.TrimSpace(cmd))
    if i.blockedCommands[upper] && !i.allowCommands[upper] {
        return fmt.Errorf("blocked Redis command: %s", upper)
    }
    return nil
}
```

默认拦截命令：

```
FLUSHALL, FLUSHDB, CONFIG, SHUTDOWN, KEYS, BGREWRITEAOF, BGSAVE
```

`KEYS` 强制用 `SCAN` 替代，避免阻塞 Redis。

## 5. 行数限制

```go
// internal/safety/limiter.go

type Limiter struct {
    maxRows    int
    timeout    time.Duration
}

// ApplyQueryLimit 在 SQL 后追加 LIMIT（如果原 SQL 没有）
func (l *Limiter) ApplyQueryLimit(sql string) string {
    upper := strings.ToUpper(sql)
    if strings.Contains(upper, "LIMIT") {
        return sql // 已有 LIMIT，不重复加
    }
    return sql + fmt.Sprintf(" LIMIT %d", l.maxRows)
}

// WrapContext 包装超时 context
func (l *Limiter) WrapContext(ctx context.Context) (context.Context, context.CancelFunc) {
    return context.WithTimeout(ctx, l.timeout)
}
```

**注意**：`ApplyQueryLimit` 是简单实现。复杂 SQL（含子查询、UNION）的 LIMIT 注入可能不正确。

> ponytail: 复杂 SQL 的 LIMIT 注入需要 SQL parser（如 `github.com/xwb1989/sqlparser`）。初期用简单追加，AI 可以自己在 SQL 中写 LIMIT。加当 AI 生成的复杂 SQL 需要自动行数保护时。

## 6. DELETE 无 WHERE 检测

```go
// 检测 DELETE FROM table 无 WHERE 子句
func checkDeleteWithoutWhere(sql string) error {
    upper := strings.ToUpper(strings.TrimSpace(sql))
    if !strings.HasPrefix(upper, "DELETE") {
        return nil
    }
    if !strings.Contains(upper, "WHERE") {
        return fmt.Errorf("DELETE without WHERE clause is blocked. Add WHERE to prevent accidental full-table delete")
    }
    return nil
}
```

同理 UPDATE 无 WHERE 也拦截。

## 7. 安全策略配置示例

### 生产环境（最严格）

```yaml
safety:
  mode: "read-only"
  max_rows: 500
  query_timeout: 10s
  blocked_keywords:
    - "DROP"
    - "TRUNCATE"
    - "GRANT"
    - "REVOKE"
    - "ALTER"
    - "DELETE"    # 连 DELETE 都拦截
    - "UPDATE"
```

### 开发环境（较宽松）

```yaml
safety:
  mode: "read-write"
  max_rows: 1000
  query_timeout: 30s
  blocked_keywords:
    - "DROP"
    - "TRUNCATE"
    - "SHUTDOWN"
  allow_blocked_keywords: []  # 需要时手动放行
```

### 混合模式（推荐）

生产库只读，开发库可写，在数据源级别配置：

```yaml
safety:                         # 全局
  mode: "read-write"
  max_rows: 1000
  blocked_keywords: ["DROP", "TRUNCATE"]

datasources:
  - name: "prod-db"
    driver: "mysql"
    dsn: "..."
    safety:                     # 覆盖为只读
      mode: "read-only"
      max_rows: 500

  - name: "dev-db"
    driver: "mysql"
    dsn: "..."
    # 不配 safety，继承全局 read-write
```

## 8. 审计日志（预留）

> ponytail: 当前版本只做拦截，不记录审计。审计日志（谁/何时/什么 SQL/是否放行）在 v0.2+ 加。加当需要合规审计或事后追溯 AI 操作时。

预留接口位置：

```go
// internal/safety/audit.go (未来)

type AuditEntry struct {
    Timestamp   time.Time
    DataSource  string
    Tool        string
    SQL         string  // 或 Redis command
    Allowed     bool
    BlockedBy   string  // 拦截原因
    Duration    time.Duration
    AffectedRows int64
}
```
