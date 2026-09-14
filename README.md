# mcp-x

閫氱敤鏁版嵁搴?MCP Server锛岃 AI 缂栫▼鍔╂墜锛坘ilocode / cursor / claude code 绛夛級閫氳繃 MCP 鍗忚瀹夊叏鎿嶄綔鏁版嵁搴撱€?
## 鐗规€?
- **MySQL + Redis** 寮€绠卞嵆鐢紝鏋舵瀯鍙墿灞曞浗浜ф暟鎹簱锛圤ceanBase / 杈炬ⅵ / 閲戜粨锛?- **瀹夊叏灞?*锛氶粯璁ゅ彧璇伙紝鍐欐搷浣滈渶鏄惧紡閰嶇疆锛涘嵄闄╁叧閿瘝鎷︽埅锛汥ELETE/UPDATE 鏃?WHERE 鎷︽埅锛涜鏁伴檺鍒?+ 鏌ヨ瓒呮椂
- **stdio 浼犺緭**锛宬ilocode / cursor 鏈湴鐩磋繛
- **YAML 閰嶇疆**锛屽鏁版嵁婧愬懡鍚嶇鐞嗭紝鎸夊悕绉拌矾鐢?- **鍗曚簩杩涘埗**锛岄浂渚濊禆閮ㄧ讲

## 蹇€熷紑濮?
### 1. 缂栬瘧

```bash
# 闇€ Go 1.25+
make build
# 鎴栫洿鎺?go build -o bin/mcp-x ./cmd/mcp-x
```

璺ㄥ钩鍙扮紪璇戯細

```bash
# Windows
$env:GOOS="windows"; $env:GOARCH="amd64"; go build -o bin/mcp-x.exe ./cmd/mcp-x
# Linux
$env:GOOS="linux"; $env:GOARCH="amd64"; go build -o bin/mcp-x ./cmd/mcp-x
# macOS
$env:GOOS="darwin"; $env:GOARCH="arm64"; go build -o bin/mcp-x ./cmd/mcp-x
```

### 2. 缂栧啓閰嶇疆

澶嶅埗 `examples/mcp-x.yaml.example` 骞朵慨鏀癸細

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
      mode: "read-only"     # 姝ゆ暟鎹簮寮哄埗鍙

  - name: "cache-redis"
    driver: "redis"
    addr: "127.0.0.1:6379"
    password: ""
    db: 0
    pool_size: 10
    safety:
      mode: "read-write"
```

### 3. 鎵嬪姩楠岃瘉

```bash
# 鍚姩骞跺彂閫?MCP JSON-RPC 娴嬭瘯
echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-11-25","capabilities":{},"clientInfo":{"name":"test","version":"1.0"}}}' | ./bin/mcp-x --config ./examples/mcp-x.yaml.example
```

姝ｅ父浼氳繑鍥?`initialize` 鍝嶅簲锛屽寘鍚?server capabilities銆?
### 4. 鎺ュ叆 AI 缂栫▼鍔╂墜

#### kilocode

`.kilocode/mcp.json`锛?
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

`.cursor/mcp.json`锛?
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

閲嶅惎 IDE 鍚庯紝AI 鍙洿鎺ヨ皟鐢ㄦ暟鎹簱宸ュ叿銆?
## 宸ュ叿鍒楄〃

鍏?12 涓伐鍏凤細

### 閫氱敤宸ュ叿

| 宸ュ叿 | 璇存槑 |
|------|------|
| `db_list` | 鍒楀嚭鎵€鏈夊凡閰嶇疆鏁版嵁婧愬強鐘舵€?|
| `db_ping` | 鍋ュ悍妫€鏌ワ紝杩斿洖寤惰繜 |

### SQL 宸ュ叿锛圡ySQL / OceanBase / 杈炬ⅵ / 閲戜粨锛?
| 宸ュ叿 | 璇存槑 |
|------|------|
| `db_query` | 鎵ц SELECT锛岃繑鍥?Markdown 琛ㄦ牸锛屾渶澶?1000 琛?|
| `db_execute` | 鎵ц INSERT/UPDATE/DELETE锛岃繑鍥炲奖鍝嶈鏁?|
| `db_tables` | 鍒楀嚭鎵€鏈夎〃 |
| `db_schema` | 鏌ョ湅琛ㄧ粨鏋勶紙鍒?绫诲瀷/绱㈠紩锛?|

### Redis 宸ュ叿

| 宸ュ叿 | 璇存槑 |
|------|------|
| `redis_get` | 璇诲彇閿€?|
| `redis_set` | 鍐欏叆閿€硷紙鏀寔 TTL锛?|
| `redis_del` | 鍒犻櫎閿?|
| `redis_keys` | SCAN 鎵弿閿紙闈為樆濉烇紝闄?100锛?|
| `redis_type` | 鏌ョ湅閿被鍨?|
| `redis_ttl` | 鏌ョ湅杩囨湡鏃堕棿 |

## 閰嶇疆璇存槑

### 閰嶇疆鏂囦欢鏌ユ壘椤哄簭

1. `--config <path>` 鍛戒护琛屾寚瀹?2. `./.kilo/mcp_x.yaml` 椤圭洰 .kilo 鐩綍锛堟帹鑽愶級
3. `~/.config/mcp_x/config.yaml` 鐢ㄦ埛鍏ㄥ眬鐩綍

### 瀹夊叏绛栫暐浼樺厛绾?
```
鏁版嵁婧?safety > 鍏ㄥ眬 safety
```

鏁版嵁婧愰厤缃?`safety` 瀛楁瑕嗙洊鍏ㄥ眬銆傛湭閰嶇疆鍒欑户鎵垮叏灞€銆?
### Redis 闆嗙兢閰嶇疆

```yaml
datasources:
  - name: "cluster-redis"
    driver: "redis"
    mode: "cluster"       # standalone(榛樿) | cluster | sentinel
    addrs:                # 闆嗙兢鐢?addrs锛堝鏁帮級
      - "10.0.0.1:6379"
      - "10.0.0.2:6379"
      - "10.0.0.3:6379"
    password: ""
    pool_size: 20
```

## 瀹夊叏绛栫暐

| 椋庨櫓 | 闃叉姢 |
|------|------|
| 璇垹鏁版嵁 | DELETE/UPDATE 鏃?WHERE 鑷姩鎷︽埅 |
| 鍒犺〃/娓呭簱 | DROP/TRUNCATE 榛樿鎷︽埅锛屽彲閰嶇疆鏀捐 |
| 鍏ㄨ〃鎵弿 | max_rows 闄愬埗 + 鏌ヨ瓒呮椂 |
| Redis KEYS 闃诲 | 寮哄埗鐢?SCAN 鏇夸唬锛孠EYS 鍛戒护鎷︽埅 |
| 鏉冮檺鎻愬崌 | GRANT/REVOKE 榛樿鎷︽埅 |

瀹夊叏灞傝皟鐢ㄩ摼锛?
```
宸ュ叿 handler
  鈫?safety.CheckWrite()           # 璇诲啓妯″紡妫€鏌?  鈫?safety.CheckSQL() / CheckRedisCommand()  # 鍗遍櫓璇嶆嫤鎴?  鈫?DELETE/UPDATE WHERE 妫€娴?  鈫?driver.Query/Execute(ctxWithTimeout)     # 瓒呮椂鎺у埗
```

## 娴嬭瘯鐜

鏈湴 Docker 璧?MySQL + Redis锛?
```bash
docker run -d --name mysql-test -e MYSQL_ROOT_PASSWORD=test -p 3306:3306 mysql:8
docker run -d --name redis-test -p 6379:6379 redis:7
```

## UTF8 澶氳瑷€璇诲啓楠岃瘉

MySQL DSN 閰嶇疆 `charset=utf8mb4`锛屽畬鏁存敮鎸佸璇█ UTF8 璇诲啓銆傛祴璇曡〃 `zy_alarm`锛屽瓧娈?`err longtext`銆?
### 娴嬭瘯鏁版嵁锛坕d 830-847锛?
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

### 鏀寔鐨勮瑷€

- 鑻辨枃锛歚connection timed out after 30s`
- 涓枃锛歚娓╁害瓒呰繃闃堝€?0搴
- 鏃ユ枃锛歚娓╁害銇屻仐銇嶃亜鍊?0搴︺倰瓒呫亪銇俱仐銇焋
- 闊╂枃锛歚鞓弰臧€ 鞛勱硠臧?80霃勲ゼ 齑堦臣頄堨姷雼堧嫟`
- 娣峰悎鏍囩偣锛歚鞛レ箻 鞓ろ攧霛检澑锛氻晿韸鸽箘韸?鞁滉皠 齑堦臣`锛堝惈鍏ㄨ鍐掑彿锛?
### 鍏抽敭鐐?
- DSN 蹇呴』 `charset=utf8mb4`锛堥潪 `utf8`锛屽悗鑰呭彧鏀寔 BMP 3 瀛楄妭锛屼笉鏀寔閮ㄥ垎 emoji 鍜岀敓鍍诲瓧锛?- MySQL 琛?鍒?`COLLATE` 寤鸿鐢?`utf8mb4_general_ci` 鎴?`utf8mb4_unicode_ci`
- MCP JSON-RPC over stdio 澶╃劧浼犺緭 UTF8锛屾棤闇€棰濆缂栫爜

## 椤圭洰缁撴瀯

```
mcp_dbx/
鈹溾攢鈹€ cmd/mcp-x/main.go           鈥?鍏ュ彛
鈹溾攢鈹€ internal/
鈹?  鈹溾攢鈹€ config/                   鈥?YAML 閰嶇疆瑙ｆ瀽
鈹?  鈹溾攢鈹€ driver/                   鈥?Driver 鎺ュ彛 + MySQL/Redis 瀹炵幇
鈹?  鈹溾攢鈹€ datasource/               鈥?鏁版嵁婧愮鐞嗗櫒
鈹?  鈹溾攢鈹€ safety/                   鈥?瀹夊叏灞?鈹?  鈹斺攢鈹€ mcp/                      鈥?MCP server + 宸ュ叿 handler
鈹溾攢鈹€ examples/                     鈥?閰嶇疆绀轰緥
鈹溾攢鈹€ docs/plans/                   鈥?璁捐鏂囨。
鈹斺攢鈹€ Makefile
```

## 鎵╁睍鏂版暟鎹簱

1. 鍐?`internal/driver/xxx/xxx.go` 瀹炵幇 `Driver` 鎴?`NoSQLDriver` 鎺ュ彛
2. `init()` 涓皟鐢?`driver.Register("xxx", ...)`
3. `cmd/mcp-x/main.go` import `_ "github.com/yourname/mcp-x/internal/driver/xxx"`
4. 閰嶇疆鏂囦欢鍔?data source锛宍driver: xxx`

**MySQL 鍗忚鍏煎鐨勬暟鎹簱**锛圤ceanBase / TiDB锛夊彲鐩存帴澶嶇敤 MySQL driver锛屼粎鏀?driver 鏍囪瘑鍚嶅嵆鍙€?
## 寮€鍙?
```bash
make build     # 缂栬瘧
make test      # 娴嬭瘯
make run       # 缂栬瘧+杩愯
make clean     # 娓呯悊
```

## 璺嚎鍥?
| 鐗堟湰 | 鍐呭 |
|------|------|
| v0.1.0 | MySQL + Redis + 瀹夊叏灞傦紝stdio 浼犺緭 鉁?|
| v0.2.0 | OceanBase + 閲戜粨鏀寔锛屽璁℃棩蹇?|
| v0.3.0 | 杈炬ⅵ DM8 鏀寔 |
| v0.4.0 | HTTP/SSE 浼犺緭 |
| v0.5.0 | 杩炴帴姹犵洃鎺с€佹參鏌ヨ鏃ュ織 |

## 鎶€鏈爤

- Go 1.25+
- [MCP Go SDK](https://github.com/modelcontextprotocol/go-sdk) v1.7.0锛堝畼鏂?SDK锛?- [go-sql-driver/mysql](https://github.com/go-sql-driver/mysql) v1.10.1
- [go-redis/v9](https://github.com/redis/go-redis) v9.22.0
- `log/slog` 鏍囧噯搴撶粨鏋勫寲鏃ュ織
