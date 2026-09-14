# 05 寮€鍙戣鍒?
## 1. 閲岀▼纰戞瑙?
| 闃舵 | 鐩爣 | 棰勪及宸ユ椂 |
|------|------|---------|
| M1: 楠ㄦ灦 | 椤圭洰鍒濆鍖?+ 閰嶇疆灞?+ Driver 鎺ュ彛 | 1-2 澶?|
| M2: MySQL | MySQL Driver + SQL 宸ュ叿 | 2-3 澶?|
| M3: Redis | Redis Driver + Redis 宸ュ叿 | 2 澶?|
| M4: 瀹夊叏灞?| 璇诲啓鍒嗙 + 鍗遍櫓璇嶆嫤鎴?| 1-2 澶?|
| M5: 闆嗘垚娴嬭瘯 | kilocode/cursor 鑱旇皟 | 1 澶?|
| M6: 鍥戒骇鏁版嵁搴?| 杈炬ⅵ/OceanBase 鎵╁睍 | 2-3 澶?|

鎬昏绾?9-13 澶╋紙涓氫綑寮€鍙戣妭濂忥級銆?
## 2. M1: 椤圭洰楠ㄦ灦

### 2.1 鍒濆鍖?
```bash
cd D:\works\mystudy\my_open\mcp_dbx
go mod init github.com/yourname/mcp-x
go get github.com/modelcontextprotocol/go-sdk/mcp
go get gopkg.in/yaml.v3
```

### 2.2 浜や粯鐗?
- [ ] `go.mod` 鍒濆鍖?- [ ] `internal/config/config.go` 鈥?YAML 閰嶇疆瑙ｆ瀽
- [ ] `internal/driver/driver.go` 鈥?Driver 鎺ュ彛瀹氫箟
- [ ] `internal/driver/registry.go` 鈥?椹卞姩娉ㄥ唽琛?- [ ] `internal/datasource/manager.go` 鈥?鏁版嵁婧愮鐞嗗櫒
- [ ] `internal/mcp/server.go` 鈥?MCP server 楠ㄦ灦锛堢┖宸ュ叿锛?- [ ] `cmd/mcp-x/main.go` 鈥?鍏ュ彛锛屽姞杞介厤缃惎鍔?server
- [ ] `examples/mcp-x.yaml.example` 鈥?閰嶇疆绀轰緥
- [ ] 鑳藉惎鍔?stdio server锛屽搷搴?`tools/list`锛堣繑鍥炵┖鍒楄〃锛?
### 2.3 楠岃瘉

```powershell
# 鏋勫缓
go build -o bin\mcp-x.exe .\cmd\mcp-x

# 鍚姩娴嬭瘯锛堟墜鍔ㄥ彂 JSON-RPC锛?echo '{"jsonrpc":"2.0","id":1,"method":"tools/list"}' | .\bin\mcp-x.exe --config .\examples\mcp-x.yaml.example
```

## 3. M2: MySQL Driver + SQL 宸ュ叿

### 3.1 渚濊禆

```bash
go get github.com/go-sql-driver/mysql
```

### 3.2 浜や粯鐗?
- [ ] `internal/driver/mysql/mysql.go` 鈥?瀹炵幇 Driver 鎺ュ彛
- [ ] `internal/mcp/tools.go` 鈥?娉ㄥ唽 `db_list`, `db_ping`, `db_query`, `db_execute`, `db_tables`, `db_schema`
- [ ] `internal/mcp/tools_sql.go` 鈥?SQL 宸ュ叿 handler
- [ ] 鍗曞厓娴嬭瘯锛歚internal/driver/mysql/mysql_test.go`锛堢敤 SQLite 鎴栧唴瀛樺簱 mock锛屾垨璺宠繃闆嗘垚娴嬭瘯锛?
### 3.3 MySQL Driver 瀹炵幇瑕佺偣

```go
type MySQLDriver struct {
    db *sql.DB
}

func (d *MySQLDriver) Connect(ctx context.Context, cfg ConnConfig) error {
    db, err := sql.Open("mysql", cfg.DSN)
    // 璁剧疆杩炴帴姹犲弬鏁?    db.SetMaxOpenConns(cfg.MaxOpenConns)
    db.SetMaxIdleConns(cfg.MaxIdleConns)
    db.SetConnMaxLifetime(cfg.ConnMaxLifetime)
    d.db = db
    return db.PingContext(ctx)
}

func (d *MySQLDriver) Query(ctx context.Context, sqlStr string, args []any) (*QueryResult, error) {
    rows, err := d.db.QueryContext(ctx, sqlStr, args...)
    // ...
    // 璇诲彇鍒楀悕
    cols, _ := rows.Columns()
    // 閫愯鎵弿锛宨nterface{} + []byte 澶勭悊
    // 杩斿洖 QueryResult{Columns, Rows}
}

func (d *MySQLDriver) ListTables(ctx context.Context) ([]TableInfo, error) {
    return d.Query(ctx, "SHOW TABLES", nil)
    // 鎴栫敤 information_schema 鏇撮€氱敤
}
```

### 3.4 楠岃瘉

閰嶇疆鐪熷疄 MySQL 杩炴帴锛岄€氳繃 MCP inspector 鎴栨墜鍔?JSON-RPC 璋冪敤 `db_query` 鎵ц `SELECT 1`銆?
## 4. M3: Redis Driver

### 4.1 渚濊禆

```bash
go get github.com/redis/go-redis/v9
```

### 4.2 浜や粯鐗?
- [ ] `internal/driver/redis/redis.go` 鈥?瀹炵幇 NoSQLDriver 鎺ュ彛
- [ ] `internal/mcp/tools_redis.go` 鈥?娉ㄥ唽 `redis_get/set/del/keys/type/ttl`
- [ ] 鍗曞厓娴嬭瘯

### 4.3 Redis Driver 瑕佺偣

```go
type RedisDriver struct {
    client redis.UniversalClient  // 鍏煎鍗曟満/闆嗙兢
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

## 5. M4: 瀹夊叏灞?
### 5.1 浜や粯鐗?
- [ ] `internal/safety/checker.go` 鈥?璇诲啓妯″紡妫€鏌?- [ ] `internal/safety/interceptor.go` 鈥?鍗遍櫓璇嶆嫤鎴?- [ ] `internal/safety/limiter.go` 鈥?琛屾暟/瓒呮椂闄愬埗
- [ ] `internal/safety/delete_check.go` 鈥?DELETE/UPDATE 鏃?WHERE 妫€娴?- [ ] 闆嗘垚鍒板伐鍏?handler 璋冪敤閾?
### 5.2 璋冪敤閾?
```
tool handler
  鈫?safety.CheckMode(datasource, isWrite)
  鈫?safety.CheckSQL(sql) 鎴?safety.CheckRedisCommand(cmd)
  鈫?safety.CheckDeleteWhere(sql)
  鈫?driver.Query/Execute(ctxWithTimeout, sql, args)
```

## 6. M5: 闆嗘垚娴嬭瘯

### 6.1 浜や粯鐗?
- [ ] 閰嶇疆 kilocode `.kilocode/mcp.json`
- [ ] 閰嶇疆 cursor `.cursor/mcp.json`
- [ ] 绔埌绔祴璇曪細AI 璋冪敤 `db_query` 鏌ヨ 鈫?杩斿洖琛ㄦ牸
- [ ] 绔埌绔祴璇曪細AI 璋冪敤 `redis_set` 鈫?`redis_get` 楠岃瘉
- [ ] 瀹夊叏鎷︽埅娴嬭瘯锛歳ead-only 妯″紡涓?`db_execute` 琚嫆
- [ ] 鍗遍櫓璇嶆祴璇曪細`DROP TABLE` 琚嫤鎴?
### 6.2 娴嬭瘯鐜

鏈湴 Docker 璧?MySQL + Redis锛?
```powershell
docker run -d --name mysql-test -e MYSQL_ROOT_PASSWORD=test -p 3306:3306 mysql:8
docker run -d --name redis-test -p 6379:6379 redis:7
```

## 7. M6: 鍥戒骇鏁版嵁搴撴墿灞?
### 7.1 OceanBase锛堟渶绠€鍗曪級

OceanBase 鍏煎 MySQL 鍗忚锛岀洿鎺ュ鐢?MySQL 椹卞姩锛?
```go
// internal/driver/oceanbase/oceanbase.go
// 瀹炶川鏄?MySQLDriver 鐨勫埆鍚嶅寘瑁?type OceanBaseDriver struct {
    MySQLDriver
}

func (d *OceanBaseDriver) Name() string { return "oceanbase" }
```

鎴栨洿绠€鍗曪細鍦?driver registry 涓妸 `oceanbase` 鏄犲皠鍒?MySQLDriver 瀹炰緥銆?
### 7.2 杈炬ⅵ DM8

闇€瑕佽揪姊﹀畼鏂?Go 椹卞姩锛?
```go
import "dameng.com/dm"  // 浠庡畨瑁呭寘 drivers/go 鑾峰彇

// 瀹炵幇鏍囧噯 database/sql 鎺ュ彛
```

**鎸戞垬**锛氳揪姊﹂┍鍔ㄤ笉鍦ㄥ叕鍏?Go module proxy锛岄渶瑕侊細
- 鏂规 A锛氭妸椹卞姩婧愮爜 vendor 鍒?`vendor/dameng/`
- 鏂规 B锛氱敤 `gorm-dameng`锛堢ぞ鍖哄皝瑁咃紝鍦?GitHub锛?
### 7.3 閲戜粨 Kingbase

Kingbase 鍏煎 PostgreSQL 鍗忚锛?
```go
import _ "github.com/lib/pq"  // 鎴?github.com/jackc/pgx/v5/stdlib
// DSN 鐢?postgres:// 鏍煎紡
```

## 8. Makefile

```makefile
.PHONY: build test run clean

build:
	go build -o bin/mcp-x ./cmd/mcp-x

run: build
	./bin/mcp-x --config ./examples/mcp-x.yaml.example

test:
	go test ./... -v

clean:
	rm -rf bin/
```

## 9. 鍙戝竷

### 9.1 璺ㄥ钩鍙版瀯寤?
```powershell
# Windows
$env:GOOS="windows"; $env:GOARCH="amd64"; go build -o bin/mcp-x-windows-amd64.exe ./cmd/mcp-x
# Linux
$env:GOOS="linux"; $env:GOARCH="amd64"; go build -o bin/mcp-x-linux-amd64 ./cmd/mcp-x
# macOS
$env:GOOS="darwin"; $env:GOARCH="arm64"; go build -o bin/mcp-x-darwin-arm64 ./cmd/mcp-x
```

### 9.2 GitHub Release

- 鎵?tag `v0.1.0`
- GitHub Actions 鑷姩鏋勫缓涓夊钩鍙颁簩杩涘埗
- Release 闄勫甫閰嶇疆绀轰緥鍜?kilocode 鎺ュ叆璇存槑

## 10. 鐗堟湰瑙勫垝

| 鐗堟湰 | 鍐呭 |
|------|------|
| v0.1.0 | MySQL + Redis + 鍩虹瀹夊叏灞傦紝stdio 浼犺緭 |
| v0.2.0 | OceanBase + 閲戜粨鏀寔锛屽璁℃棩蹇?|
| v0.3.0 | 杈炬ⅵ DM8 鏀寔 |
| v0.4.0 | HTTP/SSE 浼犺緭 |
| v0.5.0 | 杩炴帴姹犵洃鎺с€佹參鏌ヨ鏃ュ織 |
