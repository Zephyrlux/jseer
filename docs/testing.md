# 测试说明

## 1. 单元测试
```bash
go test ./...
```

## 2. 协议包体测试
- 协议编解码位于 `internal/protocol`。
- 可通过模拟客户端包体进行回归验证。

## 3. GM API 测试
- 启动 GM 服务后：
```bash
POST /api/auth/login
GET /api/config/keys
POST /api/config/{key}
```

## 4. 集成流程测试
- 资源服 `ip.txt` -> 登录服 -> 网关 -> 游戏业务流程。
- 建议在开发阶段使用 `./scripts/dev-up.sh` 统一启动。

> 注：战斗与活动相关协议仍在补齐，测试覆盖会持续完善。
