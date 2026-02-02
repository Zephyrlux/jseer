# 部署说明

## 1. 配置
编辑 `configs/config.yaml`：
- `database.driver`: `mysql` / `sqlite` / `postgres`（Ent ORM）
- `database.dsn`: 对应数据库 DSN
- `gateway.address`: TCP 网关端口（默认 5000）
- `http.address`: 资源服务端口（默认 32400）
- `gm.address`: GM 服务端口（默认 3001）

## 2. Ent 代码生成（必需）
```
go generate ./ent
```

## 3. 启动
```
# 登录服务器 (1863)
go run ./cmd/loginserver

# TCP 网关 (5000)
go run ./cmd/gateway

# 资源服务 (32401)
go run ./cmd/ressrv

# GM 服务 (3001)
go run ./cmd/gmserver
```

## 4. 多数据库示例
- MySQL: `user:pass@tcp(127.0.0.1:3306)/jseer?parseTime=true`
- Postgres: `postgres://user:pass@127.0.0.1:5432/jseer?sslmode=disable`
- SQLite: `file:jseer.db?_fk=1`

## 5. 环境变量覆盖
可使用环境变量覆盖配置（Viper 自动读取）：
- `JSEER_CONFIG` 覆盖配置文件路径（可为空，纯 env 模式）
- 例如 `LOGIN_ADDRESS`、`GAME_PUBLIC_IP`、`GAME_PORT`

## 6. 注意事项
- Ent 代码必须生成，否则服务无法启动。
- SQLite 使用 go-sqlite3 驱动，需要启用 CGO。
