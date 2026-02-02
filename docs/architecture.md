# jseer 架构说明

## 1. 目标
- 完整复刻 Lua 客户端协议与功能。  
- Go + Ent + Iris 技术栈，支持 MySQL / SQLite / PostgreSQL。  
- GM 后台全覆盖配置，实时生效、版本管理、审计与权限分级。  
- 高并发、数据一致性、可扩展、可维护。

## 2. 服务拆分
- **Gateway (TCP)**：负责 AS3 客户端协议收发与命令分发。  
- **Login Server (TCP)**：处理登录认证与服务器列表（CMD 104/105/106 等）。  
- **Resource HTTP**：提供 `ip.txt` 等资源入口。  
- **GM HTTP (Iris)**：GM 登录/权限/配置/审计 API。  
- **Game Logic**：地图、战斗、精灵、道具、活动等模块化服务层。

## 3. 目录结构（核心）
```
cmd/
  gateway/        # TCP 游戏网关
  gmserver/       # GM API 服务
  ressrv/         # ip.txt 资源服务
internal/
  config/         # 配置加载
  logging/        # 日志封装
  protocol/       # 协议编解码
  gateway/        # TCP 网关实现
  game/           # 业务逻辑模块（逐步补齐）
  gm/             # GM API 与权限
  storage/        # 数据访问抽象（Ent/Memory）
ent/schema/       # Ent 数据模型
configs/           # 配置样例

gm-web/            # Vue3+Vite GM 管理后台
```

## 4. 配置与数据一致性
- 运行时配置由 GM 平台写入数据库，并提供版本历史。  
- 服务端通过内存缓存与变更通知机制（后续可接入 Redis/NATS）。  
- 战斗与经济系统使用事务/锁保证一致性。

## 5. 协议落地策略
- 基于 Lua 协议与 Seer-golang 已知 CMD 列表，先补齐 **空包/占位包体**。  
- 以模块为单位逐步还原完整协议体与业务。  
- 每条协议都有对应单元测试与模拟客户端测试脚本。
