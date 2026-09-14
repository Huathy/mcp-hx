# 命名规范

> 数据版本：2026-09-14  
> 状态：生效中

## 1. cmd 包名规则

`cmd/` 下每个子目录是一个独立 MCP Server 二进制，命名格式：

```
mcp-<domain>x
```

- 前缀 `mcp-`：统一标识 MCP Server
- 中段 `<domain>`：领域关键词，全小写，单数概念
- 后缀 `x`：含义为 **extension**（扩展），表示这是 MCP 生态的一个可插拔扩展

### 现有

| 包路径 | 领域 | 二进制 | 说明 |
|--------|------|--------|------|
| `cmd/mcp-dbx` | 数据库（database） | `mcp-dbx` | 数据库 MCP Server，MySQL / Redis 等 |
| `cmd/mcp-filex` | 文件系统（file） | `mcp-filex` | 文件操作 MCP Server（规划中） |

### 命名规则

1. **全小写**：`mcp-dbx`，不写 `mcp-DBx`
2. **连字符分隔**：`mcp-` 与 domain 之间用 `-`，domain 与 `x` 之间无分隔
3. **单数概念**：`dbx`（数据库单数）而非 `dbsx`；`filex` 而非 `filesx`
4. **`x` 固定后缀**：所有 cmd 包必须以 `x` 结尾，不可省略

## 2. 预留命名空间

规划中的 MCP Server 扩展，遵循同一规则：

| 包路径 | 领域 | 备注 |
|--------|------|------|
| `mcp-httpx` | HTTP / API 调用 | 注意：`httpx` 与 Python 知名库同名，发布前查 PyPI/npm 重名 |
| `mcp-queuex` | 消息队列（Kafka / RabbitMQ） | — |
| `mcp-authx` | 认证 / 授权 | — |
| `mcp-shellx` | Shell 命令执行 | — |
| `mcp-gitx` | Git 操作 | — |

新增领域前，先检查：
1. 与已有知名开源项目重名（PyPI / npm / Go module）
2. domain 关键词是否能清晰表达领域语义
3. 是否与现有包语义重叠

## 3. internal 包命名

`internal/` 下按职责划分，与 cmd 包名解耦：

```
internal/
├── config/       — 配置解析
├── driver/       — 驱动接口 + 实现
├── datasource/   — 数据源管理
├── safety/       — 安全层
└── mcp/          — MCP server + 工具 handler
```

规则：
- 全小写，单数
- 职责单一，一个包一个目录
- 跨 cmd 复用的逻辑放 `internal/`，cmd 专属逻辑放 `cmd/<pkg>/`

## 4. 二进制产物命名

编译产物与 cmd 目录名一致：

```
bin/mcp-dbx       # Linux/macOS
bin/mcp-dbx.exe   # Windows
```

Makefile 按 cmd 目录生成：

```makefile
build:
	go build -o bin/mcp-dbx ./cmd/mcp-dbx
```

## 5. 反例

| 错误写法 | 正确 | 原因 |
|---------|------|------|
| `mcp-db` | `mcp-dbx` | 缺 `x` 后缀 |
| `mcp-DBx` | `mcp-dbx` | 大写 |
| `mcp-dbs-x` | `mcp-dbx` | 多余分隔符，复数 |
| `mcp-database-x` | `mcp-dbx` | domain 过长，应缩写 |
| `mcpdbx` | `mcp-dbx` | 缺连字符 |

## 6. 相关文档

- [00-overview](./plans/00-overview.md) — 项目设计概览
- [01-architecture](./plans/01-architecture.md) — 架构设计
- [05-roadmap](./plans/05-roadmap.md) — 开发计划
