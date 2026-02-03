# 部署说明

## 1. 依赖
- Go 1.20+
- 数据库：MySQL / SQLite / PostgreSQL（按需）
- SQLite 使用 go-sqlite3 时需启用 CGO

## 2. 配置方式
优先级：`.env` > `configs/config.yaml`

### 2.1 `.env` 示例
参考 `.env.example`，常用变量：
- `DATABASE_DRIVER`
- `DATABASE_DSN`
- `LOGIN_ADDRESS`
- `GATEWAY_ADDRESS`
- `HTTP_ADDRESS`
- `GM_ADDRESS`

### 2.2 配置文件
编辑 `configs/config.yaml`，字段含义：
- `database.driver`: `mysql` / `sqlite` / `postgres`
- `database.dsn`: 连接串
- `gateway.address`: TCP 网关端口
- `http.address`: 资源服务端口
- `http.static_root`: 本地资源根目录
- `http.proxy_root`: 资源覆盖目录
- `gm.address`: GM 服务端口

## 3. Ent 代码生成（必需）
```bash
go generate ./ent
```

## 4. 启动服务
```bash
# 登录服务器 (1863)
go run ./cmd/loginserver

# TCP 网关 (5000)
go run ./cmd/gateway

# 资源服务 (32401)
go run ./cmd/ressrv

# GM 服务 (3001)
go run ./cmd/gmserver
```

## 5. 多数据库 DSN 示例
- MySQL: `user:pass@tcp(127.0.0.1:3306)/jseer?parseTime=true`
- Postgres: `postgres://user:pass@127.0.0.1:5432/jseer?sslmode=disable`
- SQLite: `file:jseer.db?_fk=1`

## 6. 运行注意事项
- Ent 代码未生成会导致服务无法启动。
- SQLite 必须带 `_fk=1`，否则外键约束无法生效。
