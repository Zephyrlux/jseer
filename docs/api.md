# API 文档（概要）

## GM API（Iris）
Base URL: `http://<gm-host>:3001/api`

### 认证
- `POST /auth/login`
  - body: `{ "username": "admin", "password": "admin" }`
  - response: `{ "token": "...", "expires_at": 0 }`

### 配置管理
- `GET /config/keys`
- `GET /config/{key}`
- `POST /config/{key}`
  - body: `{ "value": { ... } }`
- `GET /config/{key}/versions`

### 审计
- `GET /audit`

## Resource API
- `GET /ip.txt` → 返回登录服地址（示例：`127.0.0.1:1863`）

> 注：后续会补齐玩家、道具、战斗、活动等完整 GM API。
