# 03 閰嶇疆瑙勮寖

## 1. 閰嶇疆鏂囦欢鏍煎紡

YAML 鏍煎紡锛屽崟鏂囦欢绠＄悊鎵€鏈夋暟鎹簮鍜屽畨鍏ㄧ瓥鐣ャ€?
### 鏂囦欢浣嶇疆

榛樿璺緞锛堟寜浼樺厛绾ф煡鎵撅級锛?
1. 鍛戒护琛?`--config <path>` 鎸囧畾
2. 褰撳墠鐩綍 `./mcp-x.yaml`
3. 鐢ㄦ埛鐩綍 `~/.mcp-x/config.yaml`

### 瀹屾暣閰嶇疆绀轰緥

```yaml
# mcp-x.yaml

# MCP Server 淇℃伅
server:
  name: "mcp-x"
  version: "0.1.0"

# 瀹夊叏绛栫暐锛堝叏灞€榛樿锛屽彲琚暟鎹簮瑕嗙洊锛?safety:
  mode: "read-write"          # read-only | read-write
  max_rows: 1000              # 鍗曟鏌ヨ鏈€澶ц繑鍥炶鏁?  query_timeout: 30s          # 鍗曟鏌ヨ瓒呮椂
  blocked_keywords:           # 鎷︽埅鐨?SQL 鍏抽敭璇?    - "DROP"
    - "TRUNCATE"
    - "GRANT"
    - "REVOKE"
    - "ALTER"
  allow_blocked_keywords: []  # 鏄惧紡鏀捐鐨勫叧閿瘝锛堣鐩?blocked锛?  blocked_commands:           # Redis 鎷︽埅鐨勫懡浠?    - "FLUSHALL"
    - "FLUSHDB"
    - "CONFIG"
    - "SHUTDOWN"
    - "KEYS"                  # 寮哄埗鐢?SCAN 鏇夸唬
  allow_blocked_commands: []

# 鏁版嵁婧愬垪琛?datasources:
  # MySQL 绀轰緥
  - name: "main-mysql"
    driver: "mysql"
    dsn: "user:password@tcp(127.0.0.1:3306)/mydb?charset=utf8mb4&parseTime=true"
    max_open_conns: 10
    max_idle_conns: 5
    conn_max_lifetime: 5m
    safety:                   # 瑕嗙洊鍏ㄥ眬瀹夊叏绛栫暐锛堝彲閫夛級
      mode: "read-only"       # 杩欎釜鏁版嵁婧愬己鍒跺彧璇?
  # Redis 绀轰緥
  - name: "cache-redis"
    driver: "redis"
    addr: "127.0.0.1:6379"
    password: ""
    db: 0
    pool_size: 10
    safety:
      mode: "read-write"

  # Redis 闆嗙兢绀轰緥
  - name: "cluster-redis"
    driver: "redis"
    mode: "cluster"           # standalone | cluster | sentinel
    addrs:                    # 闆嗙兢妯″紡鐢?addrs锛堝鏁帮級
      - "10.0.0.1:6379"
      - "10.0.0.2:6379"
      - "10.0.0.3:6379"
    password: ""
    pool_size: 20

  # OceanBase 绀轰緥锛堝鐢?mysql 椹卞姩锛屾敼 driver 鍚嶏級
  - name: "ob-prod"
    driver: "oceanbase"       # 鍐呴儴鏄犲皠鍒?mysql 椹卞姩锛屼粎鏍囪瘑鍖哄垎
    dsn: "user:pass@tcp(ob-host:2883)/db?charset=utf8mb4"
    safety:
      mode: "read-only"

  # 杈炬ⅵ DM8 绀轰緥锛堟湭鏉ユ墿灞曪級
  - name: "dm-finance"
    driver: "dameng"
    dsn: "dm://user:pass@127.0.0.1:5236/SCHEMA"
    safety:
      mode: "read-only"
```

## 2. 閰嶇疆瀛楁璇存槑

### 2.1 `server`

| 瀛楁 | 绫诲瀷 | 蹇呭～ | 璇存槑 |
|------|------|------|------|
| name | string | 鏄?| MCP server 鍚嶇О |
| version | string | 鏄?| 鐗堟湰鍙?|

### 2.2 `safety`

| 瀛楁 | 绫诲瀷 | 榛樿 | 璇存槑 |
|------|------|------|------|
| mode | string | `read-write` | `read-only` 绂佹鎵€鏈夊啓鎿嶄綔锛沗read-write` 鍏佽鍐欙紙鍙楀叧閿瘝鎷︽埅闄愬埗锛?|
| max_rows | int | 1000 | `db_query` 杩斿洖鏈€澶ц鏁?|
| query_timeout | duration | 30s | 鍗曟鎿嶄綔瓒呮椂 |
| blocked_keywords | []string | 瑙佺ず渚?| 鎷︽埅鐨?SQL 鍏抽敭璇嶏紝澶у皬鍐欎笉鏁忔劅 |
| allow_blocked_keywords | []string | `[]` | 鏀捐鐨勫叧閿瘝锛岃鐩?blocked |
| blocked_commands | []string | 瑙佺ず渚?| 鎷︽埅鐨?Redis 鍛戒护 |
| allow_blocked_commands | []string | `[]` | 鏀捐鐨勫懡浠?|

### 2.3 `datasources[]`

閫氱敤瀛楁锛?
| 瀛楁 | 绫诲瀷 | 蹇呭～ | 璇存槑 |
|------|------|------|------|
| name | string | 鏄?| 鏁版嵁婧愬悕绉帮紝宸ュ叿璋冪敤鏃剁敤姝ゅ悕绉拌矾鐢?|
| driver | string | 鏄?| `mysql` / `redis` / `oceanbase` / `dameng` / `kingbase` |
| safety | object | 鍚?| 鏁版嵁婧愮骇瀹夊叏绛栫暐锛岃鐩栧叏灞€ |

MySQL/OceanBase/杈炬ⅵ/閲戜粨锛圫QL 绫伙級鐗规湁锛?
| 瀛楁 | 绫诲瀷 | 璇存槑 |
|------|------|------|
| dsn | string | 鏍囧噯 DSN 杩炴帴涓?|
| max_open_conns | int | 鏈€澶ц繛鎺ユ暟 |
| max_idle_conns | int | 鏈€澶х┖闂茶繛鎺?|
| conn_max_lifetime | duration | 杩炴帴鏈€澶у瓨娲绘椂闂?|

Redis 鐗规湁锛?
| 瀛楁 | 绫诲瀷 | 璇存槑 |
|------|------|------|
| mode | string | `standalone`(榛樿) / `cluster` / `sentinel` |
| addr | string | 鍗曟満鍦板潃 |
| addrs | []string | 闆嗙兢鍦板潃鍒楄〃锛坈luster/sentinel 妯″紡锛?|
| password | string | 瀵嗙爜 |
| db | int | 鏁版嵁搴撶紪鍙?|
| pool_size | int | 杩炴帴姹犲ぇ灏?|

## 3. 鏁忔劅淇℃伅澶勭悊

### DSN 瀵嗙爜鑴辨晱

閰嶇疆鏂囦欢涓瘑鐮佹槑鏂囧瓨鍌ㄣ€傛棩蹇楄緭鍑哄拰 MCP 杩斿洖涓嚜鍔ㄨ劚鏁忥細

```
鍘熷锛歶ser:s3cr3t@tcp(host)/db
鏃ュ織锛歶ser:****@tcp(host)/db
```

### 鐜鍙橀噺寮曠敤锛堥鐣欙級

> ponytail: 褰撳墠鐗堟湰鐢ㄧ函閰嶇疆鏂囦欢銆傚鏈潵闇€瑕佺幆澧冨彉閲忓紩鐢紝鏀寔 `${ENV_VAR}` 璇硶鏇挎崲銆傚姞褰撻渶瑕佷粠 CI/CD 娉ㄥ叆瀵嗙爜鏃躲€?
## 4. kilocode 鎺ュ叆閰嶇疆

`.kilocode/mcp.json`锛堟垨 Settings 鈫?MCP锛変腑娣诲姞锛?
```json
{
  "mcpServers": {
    "mcp-x": {
      "command": "mcp-x",
      "args": ["--config", "/path/to/mcp-x.yaml"],
      "env": {}
    }
  }
}
```

**鍓嶆彁**锛歚mcp-x` 鍙墽琛屾枃浠跺湪 PATH 涓紝鎴栫敤缁濆璺緞銆?
## 5. cursor 鎺ュ叆閰嶇疆

`.cursor/mcp.json`锛?
```json
{
  "mcpServers": {
    "mcp-x": {
      "command": "mcp-x",
      "args": ["--config", "/path/to/mcp-x.yaml"]
    }
  }
}
```

鏍煎紡涓?kilocode 涓€鑷达紝MCP 鍗忚鏍囧噯銆?