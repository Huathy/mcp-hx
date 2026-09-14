# 06 閫氱敤宸ュ叿鎵╁睍寮€鍙戣鍒?
> 鐗堟湰锛歷0.5.0+ 璺嚎锛屾壙鎺?05-roadmap 涔嬪悗銆?> 鐩爣锛氬湪 MCP 妗嗘灦涓紩鍏ラ潪鏁版嵁婧愮被閫氱敤宸ュ叿锛堣繍缁?/ 鏂囨。 / 缃戠粶 / 缂栫爜锛夈€?
---

## 0. 鑳屾櫙涓庤竟鐣?
鐜版湁鏋舵瀯锛?- `internal/driver/`锛欴river 鎺ュ彛 + 鏁版嵁婧愭敞鍐岋紙mysql/redis/dameng/...锛?- `internal/datasource/manager.go`锛氭寜 name 绠＄悊宸茶繛鎺ュ疄渚?- `internal/mcp/server.go` + `tools_*.go`锛氭敞鍐?MCP Tool锛宧andler 璋?driver

鏂板宸ュ叿**涓嶆槸鏁版嵁婧?*锛屾棤闇€ Connect/Close 鐢熷懡鍛ㄦ湡锛屼絾澶嶇敤 MCP 娉ㄥ唽鏈哄埗銆?鍏抽敭鍐崇瓥锛氬紩鍏?`Tool` 鎶借薄锛屼笌 Driver 骞跺垪锛岀粺涓€娉ㄥ唽鍒?MCP server銆?
涓嶅湪鏈鍒掕寖鍥达細
- HTTP/SSE 浼犺緭锛堝彟鍒?v0.4.0锛?- 杩炴帴姹犵洃鎺э紙鍙﹀垪 v0.5.0锛?
---

## 1. 宸ュ叿娓呭崟锛堝叡 11 绫伙紝绾?30+ MCP Tool锛?
| # | 妯″潡 | MCP Tool 鏁?| 渚濊禆搴?| 缂栬瘧褰卞搷 |
|---|------|------------|--------|---------|
| 1 | Docker 瀹瑰櫒绠＄悊 | 5 | Docker SDK Go | 涓?|
| 2 | Docker-Compose | 4 | Docker SDK + compose-go | 涓?|
| 3 | docx鈫攎d 杞崲 | 2 | unioffice / pandoc 浜岄€変竴 | 澶?|
| 4 | Excel 璇诲啓 | 2 | xuri/excelize | 涓?|
| 5 | CSV 璇诲啓 | 2 | stdlib encoding/csv | 闆?|
| 6 | JSON/YAML 浜掕浆+鏍￠獙 | 4 | stdlib + yaml.v3锛堝凡鏈夛級 | 闆?|
| 7 | HTTP 璇锋眰 | 1 | stdlib net/http | 闆?|
| 8 | PDF 鏂囨湰鎻愬彇 | 1 | ledongthuc/pdf | 灏?|
| 9 | 浜岀淮鐮佺敓鎴?瑙ｆ瀽 | 2 | go-qrcode + gozxing | 灏?|
| 10 | SQLite锛堟寜闇€锛?| 4 | modernc.org/sqlite 绾?Go | 涓?|
| 11 | DuckDB锛堟寜闇€锛?| 4 | duckdb-go锛坈go 鎴栫函 Go 绔彛锛?| 澶?|

### 鎸夐渶鍔犺浇绛栫暐锛圫QLite / DuckDB锛?
鐢ㄦ埛瑕佹眰"榛樿涓嶄笅杞斤紝鏈夐渶瑕佹墠涓嬭浇"銆傛柟妗堬細

**鏂规 A锛氭瀯寤烘爣绛撅紙鎺ㄨ崘锛?*

```go
// internal/driver/sqlite/sqlite.go
//go:build sqlite

package sqlite
```

```bash
# 榛樿缂栬瘧涓嶅惈 SQLite/DuckDB
go build -o bin/mcp-x ./cmd/mcp-x

# 鎸夐渶缂栬瘧
go build -tags sqlite -o bin/mcp-x ./cmd/mcp-x
go build -tags duckdb -o bin/mcp-x ./cmd/mcp-x
go build -tags sqlite,duckdb -o bin/mcp-x ./cmd/mcp-x
```

main.go 涓寜 build tag 鏉′欢 import锛?
```go
// cmd/mcp-x/main.go
import (
    _ "github.com/yourname/mcp-x/internal/driver/mysql"
    _ "github.com/yourname/mcp-x/internal/driver/redis"
    // 鎸夐渶
    // _ "github.com/yourname/mcp-x/internal/driver/sqlite"
    // _ "github.com/yourname/mcp-x/internal/driver/duckdb"
)
```

鐢?`//go:build` 闅旂锛屾湭鍚敤 tag 鏃?import 琛屾敞閲婃帀锛堟垨鐢ㄧ┖ stub 鏂囦欢锛夈€?
**鏂规 B锛氭彃浠朵簩杩涘埗**锛堣繃搴﹁璁★紝涓嶇敤锛?
閫夋柟妗?A銆侻akefile 鎻愪緵鐩爣锛?
```makefile
build:        # 榛樿锛屼笉鍚?sqlite/duckdb
build-full:   # 鍚?sqlite
build-db:     # 鍚?sqlite + duckdb
```

---

## 2. 鏋舵瀯璁捐

### 2.1 鏂板 `Tool` 鎺ュ彛

```go
// internal/tool/tool.go
package tool

import "context"

// Tool 鏄棤鐘舵€佸伐鍏锋帴鍙ｏ紝涓?driver.Driver 骞跺垪銆?// 涓嶉渶瑕?Connect/Close锛屾寜闇€璋冪敤澶栭儴璧勬簮銆?type Tool interface {
    Name() string
    Description() string
    Handle(ctx context.Context, params map[string]any) (string, error)
}
```

**涓嶉€夋鏂规**銆傜悊鐢憋細鐜版湁 `mcp.AddTool` 宸茬敤娉涘瀷 input struct 鐢熸垚 JSON Schema锛屽己琛屽寘 Tool 鎺ュ彛浼氫涪绫诲瀷淇℃伅锛孉I 鎷夸笉鍒板弬鏁?schema銆?
### 2.2 鐩存帴娉ㄥ唽 MCP Tool锛堥噰鐢級

娌跨敤鐜版湁妯″紡锛氭瘡涓伐鍏锋ā鍧楀鍑?`Register(s *mcp.Server)` 鍑芥暟锛屽唴閮ㄧ洿鎺?`mcp.AddTool`銆?
```go
// internal/mcp/tools_docker.go
package mcp

func (s *Server) registerDockerTools() {
    mcp.AddTool(s.server, &mcp.Tool{
        Name:        "docker_ps",
        Description: "...",
    }, s.handleDockerPS)
    // ...
}
```

`server.go` 鐨?`registerTools()` 涓拷鍔犺皟鐢ㄣ€?
### 2.3 閰嶇疆寮€鍏?
鏂板伐鍏烽粯璁ゅ惎鐢紝鍙湪閰嶇疆涓叧闂細

```yaml
tools:
  docker:
    enabled: true
    socket: "unix:///var/run/docker.sock"  # Windows: npipe:////./pipe/docker_engine
  compose:
    enabled: true
  docs:
    enabled: true          # docx/excel/pdf
    workdir: "./data"      # 鏂囦欢鎿嶄綔鏍圭洰褰曪紝瀹夊叏杈圭晫
  http:
    enabled: true
    timeout: 30s
    allowed_hosts: []      # 绌?涓嶉檺锛涢潪绌?鐧藉悕鍗?  qr:
    enabled: true
  json_yaml:
    enabled: true
  csv:
    enabled: true
```

`Config` 缁撴瀯鏂板 `Tools ToolsConfig`銆?
### 2.4 瀹夊叏杈圭晫

| 宸ュ叿 | 椋庨櫓 | 闃叉姢 |
|------|------|------|
| Docker | 璇垹瀹瑰櫒/闀滃儚 | 榛樿鍙锛坧s/logs/stats锛夛紝鍐欐搷浣滐紙stop/rm锛夐渶 `safety.mode: read-write` |
| Compose | 鍚屼笂 | 鍚屼笂 |
| docx/excel/pdf | 璺緞绌胯秺 | 闄愬埗 workdir锛屾嫆缁?`../` 缁濆璺緞 |
| HTTP | SSRF / 鍐呯綉鎺㈡祴 | 榛樿绂佹鍐呯綉 IP锛?0/172.16/192.168/127/::1锛夛紝鍙厤缃斁琛?|
| CSV/JSON | 鏃?| 鈥?|
| QR | 鏃?| 鈥?|

---

## 3. 妯″潡璁捐

### 3.1 Docker锛坄internal/mcp/tools_docker.go`锛?
渚濊禆锛歚github.com/docker/docker/client`锛圫DK锛?
宸ュ叿锛?
| Tool | 璇存槑 | 璇诲啓 |
|------|------|------|
| `docker_ps` | 鍒楀鍣紙鍚姸鎬?闀滃儚/绔彛锛?| 鍙 |
| `docker_logs` | 瀹瑰櫒鏃ュ織锛坱ail/ since/ until锛?| 鍙 |
| `docker_stats` | 瀹瑰櫒璧勬簮鍗犵敤 | 鍙 |
| `docker_stop` | 鍋滄瀹瑰櫒 | 鍐?|
| `docker_start` | 鍚姩瀹瑰櫒 | 鍐?|

瀹炵幇瑕佺偣锛?- 瀹㈡埛绔垵濮嬪寲澶嶇敤 `client.NewClientWithOpts(client.FromEnv)`
- Windows 涓?npipe 鑷姩閫傞厤
- `docker_logs` 榛樿 tail 100 琛岋紝涓婇檺 1000

### 3.2 Docker-Compose锛坄internal/mcp/tools_compose.go`锛?
渚濊禆锛歚github.com/docker/compose/v2`锛坈ompose-go锛?
宸ュ叿锛?
| Tool | 璇存槑 | 璇诲啓 |
|------|------|------|
| `compose_up` | up -d 鎸囧畾 compose 鏂囦欢 | 鍐?|
| `compose_down` | down 鎸囧畾 compose 鏂囦欢 | 鍐?|
| `compose_ps` | 鍒楁湇鍔＄姸鎬?| 鍙 |
| `compose_logs` | 鏈嶅姟鏃ュ織 | 鍙 |

瀹炵幇瑕佺偣锛?- 杈撳叆锛歚compose_file` 璺緞锛堥檺 workdir锛?- 澶嶇敤 docker client
- compose-go API锛歚compose.NewProject` + `Up`/`Down`

### 3.3 鏂囨。杞崲锛坄internal/mcp/tools_docs.go`锛?
#### docx 鈫?md

渚濊禆閫夊瀷锛?
| 搴?| 鏂瑰悜 | 璁稿彲 | 浣撶Н | cgo |
|----|------|------|------|-----|
| unioffice | 鍙屽悜 | MIT | 涓?| 鏃?|
| pandoc锛堣皟鐢ㄤ簩杩涘埗锛?| 鍙屽悜 | GPL | 闆讹紙澶栭儴锛?| 鏃?|

**閫?unioffice**锛坄github.com/unidoc/unioffice`锛夛紝绾?Go锛屾棤澶栭儴渚濊禆銆?
宸ュ叿锛?
| Tool | 璇存槑 |
|------|------|
| `docx_to_md` | docx 鈫?markdown |
| `md_to_docx` | markdown 鈫?docx |

瀹炵幇瑕佺偣锛?- docx鈫抦d锛氶亶鍘嗘钀?琛ㄦ牸锛屾槧灏勬牱寮忥紙鏍囬/鍒楄〃/绮椾綋锛?- md鈫抎ocx锛氳В鏋?md锛堢敤 `gomarkdown/markdown` AST锛夛紝鐢熸垚娈佃惤

#### Excel 璇诲啓

渚濊禆锛歚github.com/xuri/excelize/v2`锛堢函 Go锛屾渶娴佽锛?
宸ュ叿锛?
| Tool | 璇存槑 |
|------|------|
| `excel_read` | 璇?.xlsx锛屾寚瀹?sheet/range锛岃繑鍥?markdown 琛ㄦ牸 |
| `excel_write` | 鍐?.xlsx锛岃緭鍏?sheet 鍚?+ 琛屾暟鎹?|

#### CSV 璇诲啓

stdlib `encoding/csv`銆?
宸ュ叿锛?
| Tool | 璇存槑 |
|------|------|
| `csv_read` | 璇?CSV锛岃繑鍥?markdown 琛ㄦ牸 |
| `csv_write` | 鍐?CSV锛岃緭鍏ヨ鏁版嵁 |

#### PDF 鏂囨湰鎻愬彇

渚濊禆锛歚github.com/ledongthuc/pdf`锛堢函 Go锛屾棤 cgo锛?
宸ュ叿锛?
| Tool | 璇存槑 |
|------|------|
| `pdf_text` | 鎻愬彇 PDF 鍏ㄦ枃鏂囨湰 |

### 3.4 JSON/YAML锛坄internal/mcp/tools_jsonyaml.go`锛?
stdlib `encoding/json` + `gopkg.in/yaml.v3`锛堝凡鏈変緷璧栵級銆?
宸ュ叿锛?
| Tool | 璇存槑 |
|------|------|
| `json_format` | JSON 鏍煎紡鍖?鍘嬬缉 |
| `json_validate` | JSON 鏍￠獙锛屾姤閿欎綅缃?|
| `yaml_to_json` | YAML 鈫?JSON |
| `json_to_yaml` | JSON 鈫?YAML |

### 3.5 HTTP锛坄internal/mcp/tools_http.go`锛?
stdlib `net/http`銆?
宸ュ叿锛?
| Tool | 璇存槑 |
|------|------|
| `http_request` | 鍙?HTTP 璇锋眰锛岃繑鍥炵姸鎬佺爜/澶?浣?|

鍙傛暟锛歮ethod/url/headers/body/timeout

SSRF 闃叉姢锛氳В鏋?URL host锛屾嫆缁濆唴缃?IP锛堝彲閰嶇疆鐧藉悕鍗曪級銆?
### 3.6 浜岀淮鐮侊紙`internal/mcp/tools_qr.go`锛?
渚濊禆锛?- 鐢熸垚锛歚github.com/skip2/go-qrcode`
- 瑙ｆ瀽锛歚github.com/makiuchi-d/gozxing`

宸ュ叿锛?
| Tool | 璇存槑 |
|------|------|
| `qr_generate` | 鐢熸垚浜岀淮鐮?PNG锛岃繑鍥炴枃浠惰矾寰勬垨 base64 |
| `qr_decode` | 瑙ｆ瀽鍥剧墖涓殑浜岀淮鐮侊紝杩斿洖鍐呭 |

### 3.7 SQLite锛堟寜闇€锛宍internal/driver/sqlite/sqlite.go`锛?
渚濊禆锛歚modernc.org/sqlite`锛堢函 Go锛屾棤 cgo锛?
build tag锛歚//go:build sqlite`

瀹炵幇 `driver.Driver` 鎺ュ彛锛屽鐢ㄧ幇鏈?SQL 宸ュ叿锛坉b_query/execute/tables/schema锛夈€?
DSN 绀轰緥锛歚file:test.db?cache=shared&mode=rwc`

### 3.8 DuckDB锛堟寜闇€锛宍internal/driver/duckdb/duckdb.go`锛?
渚濊禆锛歚github.com/marcboeker/go-duckdb`锛堢函 Go 缁戝畾锛屾棤 cgo锛屼綋绉ぇ锛?
build tag锛歚//go:build duckdb`

瀹炵幇 `driver.Driver` + 鎵╁睍宸ュ叿锛?
| Tool | 璇存槑 |
|------|------|
| `db_query` | 澶嶇敤 SQL 宸ュ叿 |
| `duck_query_csv` | 鐩存帴 `SELECT * FROM 'data.csv'` |
| `duck_query_parquet` | 鐩存帴 `SELECT * FROM 'data.parquet'` |
| `duck_export` | 瀵煎嚭缁撴灉鍒?CSV/Parquet |

### 3.9 鏂囦欢鎿嶄綔瀹夊叏灞傦紙`internal/safety/path.go`锛?
鏂板璺緞妫€鏌ュ伐鍏凤紝缁?docs/csv/qr 绛夋ā鍧楃敤锛?
```go
package safety

func CheckPath(path, workdir string) (string, error) {
    // 娓呯悊 ../锛屾鏌ユ槸鍚﹀湪 workdir 鍐?    abs, _ := filepath.Abs(filepath.Join(workdir, path))
    if !strings.HasPrefix(abs, workdir) {
        return "", fmt.Errorf("path escapes workdir")
    }
    return abs, nil
}
```

---

## 4. 閲岀▼纰戜笌宸ユ椂

| 闃舵 | 鍐呭 | 棰勪及 | 渚濊禆 |
|------|------|------|------|
| M7 | 鏋舵瀯锛歍oolsConfig + 妯″潡娉ㄥ唽楠ㄦ灦 + build tag 鏈哄埗 | 1 澶?| 鏃?|
| M8 | CSV + JSON/YAML锛堥浂渚濊禆鍏堣锛?| 0.5 澶?| M7 |
| M9 | HTTP 璇锋眰 + SSRF 闃叉姢 | 1 澶?| M7 |
| M10 | Excel 璇诲啓锛坋xcelize锛?| 1 澶?| M7 |
| M11 | docx鈫攎d锛坲nioffice锛?| 2 澶?| M7 |
| M12 | PDF 鎻愬彇 | 0.5 澶?| M7 |
| M13 | 浜岀淮鐮佺敓鎴?瑙ｆ瀽 | 0.5 澶?| M7 |
| M14 | Docker 瀹瑰櫒绠＄悊 | 2 澶?| M7 + Docker 鐜 |
| M15 | Docker-Compose | 1.5 澶?| M14 |
| M16 | SQLite 鎸夐渶锛坆uild tag锛?| 1 澶?| M7 |
| M17 | DuckDB 鎸夐渶锛坆uild tag锛?| 2 澶?| M7 |
| M18 | 闆嗘垚娴嬭瘯 + 鏂囨。鏇存柊 | 1 澶?| 鍏ㄩ儴 |

鎬昏绾?14 澶╋紙涓氫綑鑺傚锛夈€侻8-M13 鍙苟琛屻€?
### 鎺ㄨ崘椤哄簭

```
M7锛堥鏋讹級
  鈹溾攢 M8锛圕SV/JSON锛岄浂渚濊禆锛岄獙璇佹祦绋嬶級
  鈹溾攢 M9锛圚TTP锛?  鈹溾攢 M10锛圗xcel锛?  鈹溾攢 M11锛坉ocx锛?  鈹溾攢 M12锛圥DF锛?  鈹溾攢 M13锛圦R锛?  鈹溾攢 M16锛圫QLite锛?  鈹斺攢 M17锛圖uckDB锛?M14-M15锛圖ocker锛岄渶鐜锛屾斁鍚庯級
M18锛堟敹灏撅級
```

---

## 5. 渚濊禆搴撻€夊瀷姹囨€?
| 妯″潡 | 搴?| 鐗堟湰 | 璁稿彲 | cgo |
|------|-----|------|------|-----|
| Docker | github.com/docker/docker | latest | Apache-2.0 | 鏃?|
| Compose | github.com/docker/compose/v2 | latest | Apache-2.0 | 鏃?|
| docx | github.com/unidoc/unioffice | latest | MIT | 鏃?|
| Excel | github.com/xuri/excelize/v2 | v2.8+ | MIT | 鏃?|
| PDF | github.com/ledongthuc/pdf | latest | MIT | 鏃?|
| QR 鐢熸垚 | github.com/skip2/go-qrcode | latest | BSD-2 | 鏃?|
| QR 瑙ｆ瀽 | github.com/makiuchi-d/gozxing | latest | MIT | 鏃?|
| SQLite | modernc.org/sqlite | latest | BSD-3 | 鏃?|
| DuckDB | github.com/marcboeker/go-duckdb | latest | MIT | 鏃?|
| md 瑙ｆ瀽 | github.com/gomarkdown/markdown | latest | BSD-2 | 鏃?|

CSV / JSON / YAML / HTTP锛歴tdlib + 宸叉湁 yaml.v3

---

## 6. 閰嶇疆绀轰緥锛堟墿灞曞悗锛?
```yaml
server:
  name: "mcp-x"
  version: "0.5.0"

safety:
  mode: "read-write"

datasources:
  - name: "main-mysql"
    driver: "mysql"
    dsn: "user:pass@tcp(127.0.0.1:3306)/db"

tools:
  docker:
    enabled: true
    socket: ""              # 绌?鐢?DOCKER_HOST 鐜鍙橀噺
  compose:
    enabled: true
  docs:
    enabled: true
    workdir: "./workspace"
  http:
    enabled: true
    timeout: 30s
    allow_internal: false   # 鏄惁鍏佽鍐呯綉 IP
    allowed_hosts: []
  qr:
    enabled: true
  json_yaml:
    enabled: true
  csv:
    workdir: "./workspace"
```

---

## 7. 楠屾敹鏍囧噯

姣忎釜妯″潡闇€婊¤冻锛?
- [ ] `go build` 閫氳繃锛堥粯璁?tag锛?- [ ] `go build -tags <tag>` 閫氳繃锛堟寜闇€妯″潡锛?- [ ] 鑷冲皯 1 涓崟鍏冩祴璇曟垨鍙繍琛?self-check
- [ ] README 宸ュ叿琛ㄦ洿鏂?- [ ] 閰嶇疆绀轰緥鏇存柊

绔埌绔細

- [ ] MCP JSON-RPC 璋冪敤姣忎釜鏂?tool 杩斿洖姝ｇ‘
- [ ] kilocode/cursor 涓?AI 鑳藉彂鐜板苟璋冪敤鏂?tool
- [ ] 瀹夊叏闃叉姢鐢熸晥锛堣矾寰勭┛瓒?SSRF/鍗遍櫓鎿嶄綔鎷︽埅锛?
---

## 8. 椋庨櫓涓庡喅绛栬褰?
| 椋庨櫓 | 鍐崇瓥 |
|------|------|
| unioffice 浣撶Н澶?| 鍙帴鍙楋紝绾?Go 鏃?cgo 浼樺厛 |
| Docker SDK 渚濊禆閾鹃暱 | 鎺ュ彈锛屽畼鏂?SDK 鏈€绋?|
| DuckDB 浣撶Н ~50MB | 鐢?build tag 闅旂锛岄粯璁や笉鍚?|
| docx/md 杞崲淇濈湡搴?| 鎸夐渶锛氭爣棰?娈佃惤/鍒楄〃/琛ㄦ牸/绮椾綋鏂滀綋锛屽鏉傛牱寮忔爣娉?杩戜技杞崲" |
| SSRF 闃叉姢璇激 | allow_internal + allowed_hosts 鍙厤缃斁琛?|
