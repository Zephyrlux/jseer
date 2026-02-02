# 测试说明

## 1. 协议包体测试
- `internal/protocol` 包含基础的封包测试（后续补齐）。

## 2. GM API 测试
- 启动 GM 服务后，可用 Postman/HTTPie：
```
POST /api/auth/login
GET /api/config/keys
POST /api/config/{key}
```

## 3. 集成流程测试（占位）
- 资源服 `ip.txt` -> 登录服 -> 游戏服 的完整流程需要补齐游戏服具体 CMD 业务。
- 当前已补齐登录服基础 CMD（104/105/106/108），可先验证登录服与服务器列表响应。
