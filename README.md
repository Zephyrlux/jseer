# jseer

面向赛尔号经典客户端的 Go 服务端实现，提供完整登录/网关/资源/GM 后台能力，并支持 MySQL、SQLite、PostgreSQL 三种数据库引擎。项目目标是高可维护、高并发、高一致性的开源服务端实现。

## 特性
| 模块 | 说明 |
| --- | --- |
| 账号与角色 | 登录、角色创建与基础属性管理 |
| 游戏网关 | TCP 协议收发、指令路由、并发连接管理 |
| 资源服务 | 静态资源与动态配置响应 |
| GM 后台 | 独立 API 与前端管理系统 |
| 数据存储 | Ent ORM 统一访问，支持多引擎 |
| 配置热更新 | 支持配置版本与审计 |

## 技术栈
| 层级 | 技术 |
| --- | --- |
| 后端 | Go + Iris |
| ORM | Ent |
| 数据库 | MySQL / SQLite / PostgreSQL |
| GM 前端 | Vue3 + Vite |

## 服务与端口
| 服务 | 默认端口 | 说明 |
| --- | --- | --- |
| loginserver | 1863 | 登录服（TCP） |
| gateway | 5000 | 游戏网关（TCP） |
| ressrv | 32401 | 资源服务（HTTP） |
| gmserver | 3001 | GM API（HTTP） |

## 目录结构
| 路径 | 说明 |
| --- | --- |
| `cmd/` | 可执行入口 |
| `internal/` | 协议、逻辑、存储、GM API |
| `ent/schema/` | Ent 数据模型 |
| `gm-web/` | GM 管理后台前端 |
| `docs/` | 架构 / API / 部署 / 测试文档 |

## 快速开始
```bash
# 一键启动（开发）
./scripts/dev-up.sh

# 或分别启动
go run ./cmd/loginserver
go run ./cmd/gateway
go run ./cmd/ressrv
go run ./cmd/gmserver
```

## 配置
| 配置项 | 说明 | 默认值 |
| --- | --- | --- |
| `DATABASE_DRIVER` | 数据库驱动 | `sqlite` |
| `DATABASE_DSN` | 数据库连接串 | `file:jseer.db?_fk=1` |
| `LOGIN_ADDRESS` | 登录服地址 | `:1863` |
| `GATEWAY_ADDRESS` | 网关地址 | `:5000` |
| `HTTP_ADDRESS` | 资源服地址 | `:32401` |
| `GM_ADDRESS` | GM API 地址 | `:3001` |

配置支持 `.env` 覆盖（参考 `.env.example`），也可使用 `JSEER_CONFIG` 指定配置文件路径。

## 开发提示
| 操作 | 命令 |
| --- | --- |
| 生成 Ent 代码 | `go generate ./ent` |
| 一键启动 | `./scripts/dev-up.sh` |

## 文档
| 文档 | 路径 |
| --- | --- |
| 架构说明 | `docs/architecture.md` |
| API 文档 | `docs/api.md` |
| 部署说明 | `docs/deploy.md` |
| 测试用例 | `docs/tests.md` |

## 开源说明
本项目以开源形式发布，用于协议研究、学习与技术验证。

## 免责声明
本项目不包含任何官方客户端资源或授权内容。使用者需自行评估并遵守相关法律法规与资源授权要求，作者不对使用行为产生的任何后果承担责任。
