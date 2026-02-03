# API 文档（概要）

本项目 GM API 使用 JSON，统一前缀为 `/api`。

## 1. 认证
| 方法 | 路径 | 说明 |
| --- | --- | --- |
| POST | `/auth/login` | 登录并获取 Token |

请求示例：
```json
{"username": "admin", "password": "admin"}
```

响应示例：
```json
{"token": "<jwt>", "expires_at": 0}
```

## 2. 配置管理
| 方法 | 路径 | 说明 |
| --- | --- | --- |
| GET | `/config/keys` | 获取配置键列表 |
| GET | `/config/{key}` | 获取配置值 |
| POST | `/config/{key}` | 写入配置值 |
| GET | `/config/{key}/versions` | 查看配置版本 |

## 3. 审计
| 方法 | 路径 | 说明 |
| --- | --- | --- |
| GET | `/audit` | 审计日志列表 |

## 4. 资源服务
| 方法 | 路径 | 说明 |
| --- | --- | --- |
| GET | `/ip.txt` | 返回登录服地址（示例：`127.0.0.1:1863`） |

> 注：GM API 与配置模型会随模块补齐逐步扩展。
