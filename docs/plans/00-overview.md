# mcp-dbx 设计概览

> 数据版本：2026-09-11  
> 状态：设计草案

## 1. 项目定位

mcp-dbx 是一个用 Go 实现的通用数据库 MCP Server，让 AI 编程助手（kilocode / cursor / claude code 等）通过 MCP 协议安全操作数据库。

**核心目标**：
- 优先适配 kilocode（`.kilocode/mcp.json` command 模式）
- 兼顾 cursor、claude code 等 MCP host
- 初期支持 MySQL + Redis
- 架构可扩展国产数据库（达梦 DM8、OceanBase、金仓 Kingbase 等）

**不做的事**：
- 不做 ORM 映射层，不封装业务逻辑
- 不做图形界面，纯 MCP Server
- 不内置 AI 模型调用

## 2. 技术选型

| 层面 | 选型 | 理由 |
|------|------|------|
| 语言 | Go 1.25+ | 用户指定；单二进制部署、交叉编译方便 |
| MCP SDK | `github.com/modelcontextprotocol/go-sdk` | 官方 SDK，Google 协作维护，协议支持到 2025-11-25 |
| 传输层 | stdio 优先，预留 HTTP | stdio 最简单，kilocode/cursor 本地直连首选 |
| 配置 | YAML 配置文件 | 支持多数据源命名，比 env 变量更清晰 |
| MySQL 驱动 | `github.com/go-sql-driver/mysql` | 最成熟的 MySQL Go 驱动 |
| Redis 驱动 | `github.com/redis/go-redis/v9` | 官方维护，类型安全 |
| 日志 | `log/slog` | Go 标准库结构化日志 |

### SDK 选择说明

官方 `go-sdk` vs 社区 `mark3labs/mcp-go` 对比：

| 维度 | 官方 go-sdk | mark3labs/mcp-go |
|------|------------|-------------------|
| 协议版本 | 2025-11-25 | 2025-11-25 |
| 维护方 | MCP 官方 + Google | 社区（Ed Zynda） |
| 被引用数 | 新项目 | 1880+ 项目 |
| API 成熟度 | 稳定，API 偏底层 | 高层封装，更易用 |
| 版本 | v1.4.0+ | v0.58.0 |

**选官方 go-sdk**：长期维护有保障，kilocode 本身生态也以官方为准。API 略底层但本项目工具集固定，写法简单。

## 3. 架构总览

```
┌─────────────────────────────────────────────────┐
│              MCP Host (kilocode/cursor)          │
│         通过 stdio JSON-RPC 2.0 通信              │
└──────────────────────┬──────────────────────────┘
                       │
┌──────────────────────▼──────────────────────────┐
│                 mcp-dbx Server                   │
│  ┌───────────────────────────────────────────┐   │
│  │         MCP Tool Registry                 │   │
│  │  db_list | db_query | db_execute |        │   │
│  │  redis_get | redis_set | redis_del ...   │   │
│  └──────────────────┬───────────────────────┘   │
│  ┌──────────────────▼───────────────────────┐   │
│  │         Safety Layer (安全层)             │   │
│  │  读写模式检查 → 危险词拦截 → 行数限制      │   │
│  └──────────────────┬───────────────────────┘   │
│  ┌──────────────────▼───────────────────────┐   │
│  │       Driver Interface (驱动接口)         │   │
│  │  ┌──────────┐  ┌─────────┐  ┌────────┐  │   │
│  │  │  MySQL   │  │  Redis  │  │ 未来:  │  │   │
│  │  │  Driver  │  │ Driver  │  │ DM/OB  │  │   │
│  │  └──────────┘  └─────────┘  └────────┘  │   │
│  └───────────────────────────────────────────┘   │
└─────────────────────────────────────────────────┘
```

## 4. 核心设计原则

1. **Driver 接口抽象**：所有数据库实现统一 `Driver` 接口，新增数据库只写一个 driver 文件
2. **安全优先**：默认只读，写操作需显式配置开启；危险关键词可配置拦截或放行
3. **配置驱动**：多数据源通过 YAML 配置文件管理，支持按名称路由
4. **单二进制**：编译出一个可执行文件，kilocode 通过 command+args 启动

## 5. 国产数据库扩展路径

| 数据库 | 协议兼容 | Go 驱动 | 扩展难度 |
|--------|---------|---------|---------|
| OceanBase | MySQL 协议 | 复用 `go-sql-driver/mysql` | 极低（改驱动名即可） |
| 达梦 DM8 | 自有协议 | 官方驱动（`drivers/go` 目录）+ `gorm-dameng` | 中（需引入 CGO 或预编译驱动） |
| 金仓 Kingbase | PostgreSQL 协议 | `github.com/lib/pq` 或 `pgx` | 低 |
| TiDB | MySQL 协议 | 复用 `go-sql-driver/mysql` | 极低 |

扩展只需实现 `Driver` 接口的 `Connect / Query / Execute` 方法，注册到 driver registry。

## 6. 相关文档

- [01-架构设计](./01-architecture.md) — 分层架构、Driver 接口、安全层
- [02-MCP工具定义](./02-mcp-tools.md) — 工具列表、参数 schema、返回格式
- [03-配置规范](./03-config-spec.md) — YAML 配置文件格式、多数据源
- [04-安全策略](./04-security.md) — 读写分离、危险词拦截、行数限制
- [05-开发计划](./05-roadmap.md) — 分阶段实现计划、里程碑
