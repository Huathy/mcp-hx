# 01 鏋舵瀯璁捐

## 1. 鍒嗗眰鏋舵瀯

```
鈹屸攢鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹?鈹? Layer 4: MCP Server (鍗忚灞?              鈹?鈹? - stdio transport                          鈹?鈹? - JSON-RPC 2.0 handler                     鈹?鈹? - tool registry / dispatch                 鈹?鈹溾攢鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹?鈹? Layer 3: Safety Layer (瀹夊叏灞?            鈹?鈹? - read/write mode check                    鈹?鈹? - dangerous keyword interceptor            鈹?鈹? - row limit / query timeout                鈹?鈹溾攢鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹?鈹? Layer 2: Driver Layer (椹卞姩灞?            鈹?鈹? - Driver interface                         鈹?鈹? - MySQL driver impl                        鈹?鈹? - Redis driver impl                        鈹?鈹? - driver registry                          鈹?鈹溾攢鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹?鈹? Layer 1: Config Layer (閰嶇疆灞?            鈹?鈹? - YAML config parser                       鈹?鈹? - data source manager (澶氭暟鎹簮)           鈹?鈹斺攢鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹?```

## 2. Driver 鎺ュ彛璁捐

### 2.1 SQL 绫?Driver 鎺ュ彛锛圡ySQL/杈炬ⅵ/OceanBase/TiDB/Kingbase锛?
```go
// internal/driver/driver.go

package driver

// Driver 鏁版嵁搴撻┍鍔ㄦ娊璞℃帴鍙?type Driver interface {
    // Name 杩斿洖椹卞姩鏍囪瘑鍚嶏紝濡?"mysql"銆?dameng"銆?oceanbase"
    Name() string

    // Type 杩斿洖椹卞姩绫诲瀷锛孲QL 鎴?NoSQL
    Type() DriverType

    // Connect 寤虹珛杩炴帴锛宑onnConfig 浠?YAML 瑙ｆ瀽鐨勮繛鎺ラ厤缃?    Connect(ctx context.Context, cfg ConnConfig) error

    // Query 鎵ц鍙鏌ヨ锛岃繑鍥炵粨鏋勫寲缁撴灉
    Query(ctx context.Context, sql string, args []any) (*QueryResult, error)

    // Execute 鎵ц鍐欐搷浣滐紝杩斿洖褰卞搷琛屾暟
    Execute(ctx context.Context, sql string, args []any) (*ExecResult, error)

    // ListTables 鍒楀嚭鎵€鏈夎〃锛圓I 鎺㈢储鐢級
    ListTables(ctx context.Context) ([]TableInfo, error)

    // DescribeTable 鏌ョ湅琛ㄧ粨鏋勶紙AI 鐞嗚В schema 鐢級
    DescribeTable(ctx context.Context, table string) (*TableSchema, error)

    // Ping 鍋ュ悍妫€鏌?    Ping(ctx context.Context) error

    // Close 鍏抽棴杩炴帴
    Close() error
}

type DriverType string

const (
    DriverTypeSQL    DriverType = "sql"
    DriverTypeNoSQL  DriverType = "nosql"
)
```

### 2.2 NoSQL 绫?Driver 鎺ュ彛锛圧edis锛屽悗缁彲鎵╁睍 MongoDB锛?
Redis 鏃犳硶濂楃敤 SQL 鎺ュ彛锛屽崟鐙畾涔夛細

```go
// internal/driver/nosql_driver.go

type NoSQLDriver interface {
    Name() string
    Type() DriverType

    Connect(ctx context.Context, cfg ConnConfig) error

    // Get 璇诲彇閿€?    Get(ctx context.Context, key string) (string, error)

    // Set 鍐欏叆閿€?    Set(ctx context.Context, key string, value string, ttlSeconds int) error

    // Del 鍒犻櫎閿?    Del(ctx context.Context, key string) (int64, error)

    // Keys 鎸夋ā寮忓尮閰嶉敭锛堝彧璇伙紝鍙厑璁?SCAN锛?    Keys(ctx context.Context, pattern string, limit int) ([]string, error)

    // Type 鏌ョ湅閿被鍨?    Type(ctx context.Context, key string) (string, error)

    // TTL 鏌ョ湅杩囨湡鏃堕棿
    TTL(ctx context.Context, key string) (int64, error)

    // Execute 鎵ц鍘熷鍛戒护锛堝彈瀹夊叏灞傞檺鍒讹級
    Execute(ctx context.Context, command string, args ...string) (any, error)

    Ping(ctx context.Context) error
    Close() error
}
```

### 2.3 缁熶竴 union 鎺ュ彛

Server 灞傞渶瑕佸悓鏃跺鐞?SQL 鍜?NoSQL锛岀敤 union 鍖呰锛?
```go
type AnyDriver interface {
    Name() string
    Type() DriverType
    Ping(ctx context.Context) error
    Close() error
}
```

瀹夊叏灞傛牴鎹?`Type()` 璺敱鍒颁笉鍚岀殑妫€鏌ラ€昏緫銆?
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

// 鍒濆鍖栨椂娉ㄥ唽
func init() {
    Register("mysql", func() AnyDriver { return &MySQLDriver{} })
    Register("redis", func() AnyDriver { return &RedisDriver{} })
    // 鏈潵锛歊egister("dameng", ...) Register("oceanbase", ...)
}
```

**鏂板鏁版嵁搴撳彧闇€**锛?1. 鍐欎竴涓?`xxx_driver.go` 瀹炵幇 `Driver` 鎴?`NoSQLDriver` 鎺ュ彛
2. `init()` 娉ㄥ唽
3. 閰嶇疆鏂囦欢鍔犱竴涓?data source锛宍driver: xxx`

## 4. 鏁版嵁婧愮鐞嗗櫒

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

// Get 鎸夊悕绉拌幏鍙栨暟鎹簮
func (m *Manager) Get(name string) (driver.AnyDriver, error) {
    d, ok := m.sources[name]
    if !ok {
        return nil, fmt.Errorf("data source not found: %s", name)
    }
    return d, nil
}
```

## 5. 浼犺緭灞傞鐣?
```go
// internal/transport/transport.go

type Transport interface {
    Serve(ctx context.Context) error
}

// 褰撳墠瀹炵幇
type StdioTransport struct { ... }

// 鏈潵棰勭暀锛屾殏涓嶅疄鐜?// type HTTPTransport struct { ... }
```

## 6. 鐩綍缁撴瀯

```
mcp_dbx/
鈹溾攢鈹€ cmd/
鈹?  鈹斺攢鈹€ mcp-x/
鈹?      鈹斺攢鈹€ main.go              # 鍏ュ彛锛氬姞杞介厤缃€佸惎鍔?MCP server
鈹溾攢鈹€ internal/
鈹?  鈹溾攢鈹€ config/                  # YAML 閰嶇疆瑙ｆ瀽
鈹?  鈹?  鈹斺攢鈹€ config.go
鈹?  鈹溾攢鈹€ datasource/             # 鏁版嵁婧愮鐞嗗櫒
鈹?  鈹?  鈹斺攢鈹€ manager.go
鈹?  鈹溾攢鈹€ driver/                  # Driver 鎺ュ彛 + 瀹炵幇
鈹?  鈹?  鈹溾攢鈹€ driver.go            # SQL Driver 鎺ュ彛
鈹?  鈹?  鈹溾攢鈹€ nosql_driver.go      # NoSQL Driver 鎺ュ彛
鈹?  鈹?  鈹溾攢鈹€ registry.go          # 椹卞姩娉ㄥ唽琛?鈹?  鈹?  鈹溾攢鈹€ mysql/
鈹?  鈹?  鈹?  鈹斺攢鈹€ mysql.go         # MySQL 瀹炵幇
鈹?  鈹?  鈹斺攢鈹€ redis/
鈹?  鈹?      鈹斺攢鈹€ redis.go         # Redis 瀹炵幇
鈹?  鈹溾攢鈹€ safety/                  # 瀹夊叏灞?鈹?  鈹?  鈹溾攢鈹€ checker.go           # 璇诲啓妯″紡妫€鏌?鈹?  鈹?  鈹溾攢鈹€ interceptor.go       # 鍗遍櫓璇嶆嫤鎴?鈹?  鈹?  鈹斺攢鈹€ limiter.go           # 琛屾暟/瓒呮椂闄愬埗
鈹?  鈹溾攢鈹€ mcp/                     # MCP 鍗忚灞?鈹?  鈹?  鈹溾攢鈹€ server.go            # MCP server
鈹?  鈹?  鈹溾攢鈹€ tools.go             # Tool 瀹氫箟 + handler
鈹?  鈹?  鈹斺攢鈹€ transport.go         # 浼犺緭灞傛帴鍙?鈹?  鈹斺攢鈹€ transport/
鈹?      鈹斺攢鈹€ stdio.go             # stdio 瀹炵幇
鈹溾攢鈹€ docs/
鈹?  鈹斺攢鈹€ plans/                   # 鏈璁℃枃妗?鈹溾攢鈹€ examples/                    # 閰嶇疆绀轰緥
鈹?  鈹斺攢鈹€ mcp-x.yaml.example
鈹溾攢鈹€ go.mod
鈹溾攢鈹€ go.sum
鈹溾攢鈹€ Makefile                     # 鏋勫缓/娴嬭瘯蹇嵎鍛戒护
鈹斺攢鈹€ README.md
```

## 7. 鍚姩娴佺▼

```
main.go
  鈹?  鈹溾攢 1. 瑙ｆ瀽 flag锛?-config <path>锛堥粯璁?mcp-x.yaml锛?  鈹溾攢 2. 鍔犺浇 YAML 閰嶇疆
  鈹溾攢 3. datasource.NewManager(configs) 鈫?鍒濆鍖栨墍鏈夋暟鎹簮杩炴帴
  鈹溾攢 4. safety.NewChecker(config) 鈫?鍒濆鍖栧畨鍏ㄦ鏌ュ櫒
  鈹溾攢 5. mcp.NewServer(tools, manager, safety) 鈫?鍒涘缓 MCP server
  鈹斺攢 6. server.Serve(ctx, StdioTransport{}) 鈫?stdio 闃诲杩愯
```

## 8. 閿欒澶勭悊绛栫暐

- 閰嶇疆閿欒锛氬惎鍔ㄥ嵆澶辫触锛宻tderr 杈撳嚭鏄庣‘閿欒
- 杩炴帴澶辫触锛氬惎鍔ㄥ嵆澶辫触锛屽垪鍑哄け璐ョ殑鏁版嵁婧?- 杩愯鏃舵煡璇㈤敊璇細杩斿洖 MCP error response锛屼笉 panic
- 瀹夊叏鎷︽埅锛氳繑鍥?MCP error response锛屽寘鍚嫤鎴師鍥?