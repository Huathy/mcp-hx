# 01 架构设计

## 1. 分层架构

```
┌─────────────────────────────────────────────┐
│  Layer 4: MCP Server (协议层)              │
│  - stdio transport                          │
│  - JSON-RPC 2.0 handler                     │
│  - tool registry / dispatch                 │
├─────────────────────────────────────────────┤
│  Layer 3: Safety Layer (安全层)            │
│  - read/write mode check                    │
│  - dangerous keyword interceptor            │
│  - row limit / query timeout                │
├─────────────────────────────────────────────┤
│  Layer 2: Driver Layer (驱动层)            │
│  - Driver interface                         │
│  - MySQL driver impl                        │
│  - Redis driver impl                        │
│  - driver registry                          │
├─────────────────────────────────────────────┤
│  Layer 1: Config Layer (配置层)            │
│  - YAML config parser                       │
│  - data source manager (多数据源)           │
└─────────────────────────────────────────────┘
```

## 2. Driver 接口设计

### 2.1 SQL 类 Driver 接口（MySQL/达梦/OceanBase/TiDB/Kingbase）

```go
// internal/driver/driver.go

package driver

// Driver 数据库驱动抽象接口
type Driver interface {
    // Name 返回驱动标识名，如 "mysql"、"dameng"、"oceanbase"
    Name() string

    // Type 返回驱动类型，SQL 或 NoSQL
    Type() DriverType

    // Connect 建立连接，connConfig 从 YAML 解析的连接配置
    Connect(ctx context.Context, cfg ConnConfig) error

    // Query 执行只读查询，返回结构化结果
    Query(ctx context.Context, sql string, args []any) (*QueryResult, error)

    // Execute 执行写操作，返回影响行数
    Execute(ctx context.Context, sql string, args []any) (*ExecResult, error)

    // ListTables 列出所有表（AI 探索用）
    ListTables(ctx context.Context) ([]TableInfo, error)

    // DescribeTable 查看表结构（AI 理解 schema 用）
    DescribeTable(ctx context.Context, table string) (*TableSchema, error)

    // Ping 健康检查
    Ping(ctx context.Context) error

    // Close 关闭连接
    Close() error
}

type DriverType string

const (
    DriverTypeSQL    DriverType = "sql"
    DriverTypeNoSQL  DriverType = "nosql"
)
```

### 2.2 NoSQL 类 Driver 接口（Redis，后续可扩展 MongoDB）

Redis 无法套用 SQL 接口，单独定义：

```go
// internal/driver/nosql_driver.go

type NoSQLDriver interface {
    Name() string
    Type() DriverType

    Connect(ctx context.Context, cfg ConnConfig) error

    // Get 读取键值
    Get(ctx context.Context, key string) (string, error)

    // Set 写入键值
    Set(ctx context.Context, key string, value string, ttlSeconds int) error

    // Del 删除键
    Del(ctx context.Context, key string) (int64, error)

    // Keys 按模式匹配键（只读，只允许 SCAN）
    Keys(ctx context.Context, pattern string, limit int) ([]string, error)

    // Type 查看键类型
    Type(ctx context.Context, key string) (string, error)

    // TTL 查看过期时间
    TTL(ctx context.Context, key string) (int64, error)

    // Execute 执行原始命令（受安全层限制）
    Execute(ctx context.Context, command string, args ...string) (any, error)

    Ping(ctx context.Context) error
    Close() error
}
```

### 2.3 统一 union 接口

Server 层需要同时处理 SQL 和 NoSQL，用 union 包装：

```go
type AnyDriver interface {
    Name() string
    Type() DriverType
    Ping(ctx context.Context) error
    Close() error
}
```

安全层根据 `Type()` 路由到不同的检查逻辑。

## 3. Driver Registry

```go
// internal/driver/registry.go

var registry = map[string]func() AnyDriver{}

func Register(name string, factory func() AnyDriver) {
    registry[name] = factory
}

func New(name string) (AnyDriver, error) {
    f, ok := registry[name]
    if !ok {
        return nil, fmt.Errorf("unknown driver: %s", name)
    }
    return f(), nil
}

// 初始化时注册
func init() {
    Register("mysql", func() AnyDriver { return &MySQLDriver{} })
    Register("redis", func() AnyDriver { return &RedisDriver{} })
    // 未来：Register("dameng", ...) Register("oceanbase", ...)
}
```

**新增数据库只需**：
1. 写一个 `xxx_driver.go` 实现 `Driver` 或 `NoSQLDriver` 接口
2. `init()` 注册
3. 配置文件加一个 data source，`driver: xxx`

## 4. 数据源管理器

```go
// internal/datasource/manager.go

type Manager struct {
    sources map[string]driver.AnyDriver // name -> driver instance
}

func NewManager(configs []DataSourceConfig) (*Manager, error) {
    m := &Manager{sources: make(map[string]driver.AnyDriver)}
    for _, cfg := range configs {
        d, err := driver.New(cfg.Driver)
        if err != nil {
            return nil, err
        }
        if err := d.Connect(ctx, cfg.Conn); err != nil {
            return nil, err
        }
        m.sources[cfg.Name] = d
    }
    return m, nil
}

// Get 按名称获取数据源
func (m *Manager) Get(name string) (driver.AnyDriver, error) {
    d, ok := m.sources[name]
    if !ok {
        return nil, fmt.Errorf("data source not found: %s", name)
    }
    return d, nil
}
```

## 5. 传输层预留

```go
// internal/transport/transport.go

type Transport interface {
    Serve(ctx context.Context) error
}

// 当前实现
type StdioTransport struct { ... }

// 未来预留，暂不实现
// type HTTPTransport struct { ... }
```

## 6. 目录结构

```
mcp_dbx/
├── cmd/
│   └── mcp-dbx/
│       └── main.go              # 入口：加载配置、启动 MCP server
├── internal/
│   ├── config/                  # YAML 配置解析
│   │   └── config.go
│   ├── datasource/             # 数据源管理器
│   │   └── manager.go
│   ├── driver/                  # Driver 接口 + 实现
│   │   ├── driver.go            # SQL Driver 接口
│   │   ├── nosql_driver.go      # NoSQL Driver 接口
│   │   ├── registry.go          # 驱动注册表
│   │   ├── mysql/
│   │   │   └── mysql.go         # MySQL 实现
│   │   └── redis/
│   │       └── redis.go         # Redis 实现
│   ├── safety/                  # 安全层
│   │   ├── checker.go           # 读写模式检查
│   │   ├── interceptor.go       # 危险词拦截
│   │   └── limiter.go           # 行数/超时限制
│   ├── mcp/                     # MCP 协议层
│   │   ├── server.go            # MCP server
│   │   ├── tools.go             # Tool 定义 + handler
│   │   └── transport.go         # 传输层接口
│   └── transport/
│       └── stdio.go             # stdio 实现
├── docs/
│   └── plans/                   # 本设计文档
├── examples/                    # 配置示例
│   └── mcp-dbx.yaml.example
├── go.mod
├── go.sum
├── Makefile                     # 构建/测试快捷命令
└── README.md
```

## 7. 启动流程

```
main.go
  │
  ├─ 1. 解析 flag：--config <path>（默认 mcp-dbx.yaml）
  ├─ 2. 加载 YAML 配置
  ├─ 3. datasource.NewManager(configs) → 初始化所有数据源连接
  ├─ 4. safety.NewChecker(config) → 初始化安全检查器
  ├─ 5. mcp.NewServer(tools, manager, safety) → 创建 MCP server
  └─ 6. server.Serve(ctx, StdioTransport{}) → stdio 阻塞运行
```

## 8. 错误处理策略

- 配置错误：启动即失败，stderr 输出明确错误
- 连接失败：启动即失败，列出失败的数据源
- 运行时查询错误：返回 MCP error response，不 panic
- 安全拦截：返回 MCP error response，包含拦截原因
