# 鍛藉悕瑙勮寖

> 鏁版嵁鐗堟湰锛?026-09-14  
> 鐘舵€侊細鐢熸晥涓?
## 1. cmd 鍖呭悕瑙勫垯

`cmd/` 涓嬫瘡涓瓙鐩綍鏄竴涓嫭绔?MCP Server 浜岃繘鍒讹紝鍛藉悕鏍煎紡锛?
```
mcp-<domain>x
```

- 鍓嶇紑 `mcp-`锛氱粺涓€鏍囪瘑 MCP Server
- 涓 `<domain>`锛氶鍩熷叧閿瘝锛屽叏灏忓啓锛屽崟鏁版蹇?- 鍚庣紑 `x`锛氬惈涔変负 **extension**锛堟墿灞曪級锛岃〃绀鸿繖鏄?MCP 鐢熸€佺殑涓€涓彲鎻掓嫈鎵╁睍

### 鐜版湁

| 鍖呰矾寰?| 棰嗗煙 | 浜岃繘鍒?| 璇存槑 |
|--------|------|--------|------|
| `cmd/mcp-x` | 鏁版嵁搴擄紙database锛?| `mcp-x` | 鏁版嵁搴?MCP Server锛孧ySQL / Redis 绛?|
| `cmd/mcp-filex` | 鏂囦欢绯荤粺锛坒ile锛?| `mcp-filex` | 鏂囦欢鎿嶄綔 MCP Server锛堣鍒掍腑锛?|

### 鍛藉悕瑙勫垯

1. **鍏ㄥ皬鍐?*锛歚mcp-x`锛屼笉鍐?`mcp-x`
2. **杩炲瓧绗﹀垎闅?*锛歚mcp-` 涓?domain 涔嬮棿鐢?`-`锛宒omain 涓?`x` 涔嬮棿鏃犲垎闅?3. **鍗曟暟姒傚康**锛歚dbx`锛堟暟鎹簱鍗曟暟锛夎€岄潪 `dbsx`锛沗filex` 鑰岄潪 `filesx`
4. **`x` 鍥哄畾鍚庣紑**锛氭墍鏈?cmd 鍖呭繀椤讳互 `x` 缁撳熬锛屼笉鍙渷鐣?
## 2. 棰勭暀鍛藉悕绌洪棿

瑙勫垝涓殑 MCP Server 鎵╁睍锛岄伒寰悓涓€瑙勫垯锛?
| 鍖呰矾寰?| 棰嗗煙 | 澶囨敞 |
|--------|------|------|
| `mcp-httpx` | HTTP / API 璋冪敤 | 娉ㄦ剰锛歚httpx` 涓?Python 鐭ュ悕搴撳悓鍚嶏紝鍙戝竷鍓嶆煡 PyPI/npm 閲嶅悕 |
| `mcp-queuex` | 娑堟伅闃熷垪锛圞afka / RabbitMQ锛?| 鈥?|
| `mcp-authx` | 璁よ瘉 / 鎺堟潈 | 鈥?|
| `mcp-shellx` | Shell 鍛戒护鎵ц | 鈥?|
| `mcp-gitx` | Git 鎿嶄綔 | 鈥?|

鏂板棰嗗煙鍓嶏紝鍏堟鏌ワ細
1. 涓庡凡鏈夌煡鍚嶅紑婧愰」鐩噸鍚嶏紙PyPI / npm / Go module锛?2. domain 鍏抽敭璇嶆槸鍚﹁兘娓呮櫚琛ㄨ揪棰嗗煙璇箟
3. 鏄惁涓庣幇鏈夊寘璇箟閲嶅彔

## 3. internal 鍖呭懡鍚?
`internal/` 涓嬫寜鑱岃矗鍒掑垎锛屼笌 cmd 鍖呭悕瑙ｈ€︼細

```
internal/
鈹溾攢鈹€ config/       鈥?閰嶇疆瑙ｆ瀽
鈹溾攢鈹€ driver/       鈥?椹卞姩鎺ュ彛 + 瀹炵幇
鈹溾攢鈹€ datasource/   鈥?鏁版嵁婧愮鐞?鈹溾攢鈹€ safety/       鈥?瀹夊叏灞?鈹斺攢鈹€ mcp/          鈥?MCP server + 宸ュ叿 handler
```

瑙勫垯锛?- 鍏ㄥ皬鍐欙紝鍗曟暟
- 鑱岃矗鍗曚竴锛屼竴涓寘涓€涓洰褰?- 璺?cmd 澶嶇敤鐨勯€昏緫鏀?`internal/`锛宑md 涓撳睘閫昏緫鏀?`cmd/<pkg>/`

## 4. 浜岃繘鍒朵骇鐗╁懡鍚?
缂栬瘧浜х墿涓?cmd 鐩綍鍚嶄竴鑷达細

```
bin/mcp-x       # Linux/macOS
bin/mcp-x.exe   # Windows
```

Makefile 鎸?cmd 鐩綍鐢熸垚锛?
```makefile
build:
	go build -o bin/mcp-x ./cmd/mcp-x
```

## 5. 鍙嶄緥

| 閿欒鍐欐硶 | 姝ｇ‘ | 鍘熷洜 |
|---------|------|------|
| `mcp-db` | `mcp-x` | 缂?`x` 鍚庣紑 |
| `mcp-x` | `mcp-x` | 澶у啓 |
| `mcp-dbs-x` | `mcp-x` | 澶氫綑鍒嗛殧绗︼紝澶嶆暟 |
| `mcp-database-x` | `mcp-x` | domain 杩囬暱锛屽簲缂╁啓 |
| `mcpdbx` | `mcp-x` | 缂鸿繛瀛楃 |

## 6. 鐩稿叧鏂囨。

- [00-overview](./plans/00-overview.md) 鈥?椤圭洰璁捐姒傝
- [01-architecture](./plans/01-architecture.md) 鈥?鏋舵瀯璁捐
- [05-roadmap](./plans/05-roadmap.md) 鈥?寮€鍙戣鍒?