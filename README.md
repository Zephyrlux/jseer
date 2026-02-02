# jseer

Golang + Ent + Iris 的赛尔号服务端重构项目，目标完整复刻 Lua 客户端功能，并提供独立 GM 后台（Vue3 + Vite）。

## 目录结构
- `cmd/`：可执行入口（TCP 网关 / 资源服务 / GM 服务）
- `internal/`：协议、游戏逻辑、存储、GM API
- `ent/schema/`：Ent 数据模型
- `gm-web/`：GM 管理后台前端
- `docs/`：架构 / API / 部署 / 测试文档

## 快速开始
```
# 一键启动（开发）
./scripts/dev-up.sh

# 或分别启动
go run ./cmd/loginserver
go run ./cmd/gateway
go run ./cmd/ressrv
go run ./cmd/gmserver
```

## 说明
- 当前版本以可运行骨架 + 协议占位为主，后续逐条补齐 Lua 客户端协议体与业务逻辑。
- 数据库层统一使用 Ent ORM（mysql/sqlite/postgres）。
- 首次使用需生成 Ent 代码：`go generate ./ent`。
- 支持环境变量覆盖（参考 `.env.example`），也可用 `JSEER_CONFIG` 指定配置文件路径。
