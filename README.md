# echo-admin

`echo-admin` 是一个 module-first 的 Go 中后台基础框架。它提供登录认证、管理员管理、角色权限、菜单管理、系统配置、数据字典、文件上传、操作日志和登录日志，为后续业务模块提供统一后台管理能力。

项目后端使用 Go 1.26、Echo v5、Casbin 和 `samber/do`。前端位于 `web/`，使用 Umi Max、React、Ant Design 和 ProComponents。

## 架构边界

- `business module`：业务模块，位于 `internal/modules/<module>`。
- `platform`：平台运行时代码，位于 `internal/platform`。
- `capability`：可复用基础设施能力，例如 MySQL、Redis、MongoDB、tracing、logging。
- `composition root`：启动装配层，位于 `internal/boot`。

核心目录：

```text
.
├── cmd/                         # CLI 入口，默认读取 configs/config.toml
├── configs/                     # 静态配置示例和配置说明
├── internal/boot/               # composition root，负责装配资源、模块和路由
├── internal/modules/auth/       # 登录认证、JWT 签发、当前用户和授权判断
├── internal/modules/identity/   # 管理员管理
├── internal/modules/access/     # 角色权限和菜单管理
├── internal/modules/settings/   # 系统配置和数据字典
├── internal/modules/fileasset/  # 文件上传元数据
├── internal/modules/audit/      # 操作日志和登录日志
├── internal/platform/           # HTTP runtime、配置、错误、请求响应、基础设施能力
├── web/                         # 前端中后台应用
├── k8s/                         # Kubernetes 示例
├── docker-compose.yaml          # 本地开发 Compose
├── Dockerfile                   # runtime image
└── Makefile                     # 构建、测试、验证命令
```

不要把本项目改回 `service/biz/data` 分层心智。业务路由只在 `internal/boot` 通过 `NewModule`、`Provide`、`Route` 显式装配，`internal/platform/server` 不 import 业务模块。

## 快速开始

后端依赖 MySQL 存储。先启动本地 MySQL：

```bash
docker compose up mysql
```

再启动 API：

```bash
go run main.go --config configs/config.toml
```

默认监听 `:9322`。基础检查：

```bash
curl http://127.0.0.1:9322/api/health
curl http://127.0.0.1:9322/api/info
curl http://127.0.0.1:9322/api/ready
curl http://127.0.0.1:9322/api/capabilities
```

第一次启动会在 MySQL 中创建基础菜单、`super_admin` 角色、系统配置、状态字典和一个本地管理员：

```text
username: admin
password: admin123
```

`super_admin` 是根角色，默认拥有全部权限、全部基础菜单和默认入口 `/dashboard`。管理员可以拥有多个角色，JWT 会携带当前生效的 `role_id`；切换角色后后端会重新签发只包含该角色权限的新 token，前端菜单和按钮权限也会随之刷新。

前端：

```bash
cd web
npm install
npm run dev
```

前端 dev server 默认把 `/api` 代理到 `http://127.0.0.1:9322`；如需连接其他后端，设置 `ECHO_ADMIN_WEB_API_TARGET`。生产部署按同源 `/api` 访问，登录后通过 `Authorization: Bearer <token>` 调用后端。

## 配置

默认配置文件是 `configs/config.toml`。配置支持 TOML、YAML、JSON，未知配置字段会导致启动失败。环境变量前缀是 `ECHO_ADMIN_`。

本地默认启用 JWT，并跳过这些公开路径：

```toml
[jwt]
enabled = true
skip_paths = ["/api/health", "/api/info", "/api/ready", "/api/capabilities", "/api/auth/login"]
```

真实环境必须通过 `ECHO_ADMIN_JWT_SECRET` 或 Secret 管理系统提供 JWT secret，不要使用示例配置里的开发 secret。

可选 capability 默认关闭：

- `redis.enabled=false`
- `mongodb.enabled=false`
- `tracing.enabled=false`
- `http.cors.enabled=false`

MySQL 是后台基础能力的必需存储，示例配置默认 `mysql.enabled=true`。Redis、MongoDB、tracing 等可选资源启用后会进入 `/api/ready` 和 `/api/capabilities`。上传文件默认保存到 `admin.upload_dir`，本地示例值是 `uploads`。

完整配置说明见 `configs/README.md`。

## 本地 Compose

默认 `docker compose up` 启动 API 和 MySQL，公开端口默认绑定 `127.0.0.1`：

```bash
docker compose up
```

Redis、MongoDB 和 Jaeger 在 `resources` profile 下：

```bash
cp env.example .env
docker compose --profile resources up
```

`.env` 只给 Docker Compose 做本地开发端口和凭据替换；应用本身读取真实进程环境变量，不会自动加载 `.env` 文件。

资源服务使用 Compose 命名卷保存本地开发数据。需要重置本地资源状态时，显式运行 `docker compose down -v`。

## Foundation Modules

后台基础能力按边界拆成多个 business module：

```text
internal/modules/auth/           # 登录认证、当前用户和权限判断
internal/modules/identity/       # 管理员生命周期
internal/modules/access/         # 角色、权限和菜单
internal/modules/settings/       # 系统配置和数据字典
internal/modules/fileasset/      # 文件上传元数据
internal/modules/audit/          # 操作日志和登录日志
```

运行期只使用 MySQL adapter。各模块的 usecase 定义自己的 store interface，`internal/boot` 从已配置的 `*gorm.DB` 创建 concrete store，并负责跨模块装配，例如 auth 通过自己的小接口读取 identity/access，并通过 boot bridge 写入 audit。

授权判断基于 Casbin RBAC：管理员映射为 `user:<id>`，角色映射为 `role:<code>`，权限 token 必须是 `resource:action`，并在授权时映射为 Casbin 的 `{subject, object, action}`。当前生效角色决定本次请求的权限集合，已分配但未激活的其他角色不会参与授权。

`access` 模块提供权限目录、角色树和菜单管理。角色通过 `parent_id` 形成委派树：`super_admin` 可以管理全部角色；普通角色只能查看自己和下级角色，只能把自己的下级角色分配给管理员，并且不能授予自己没有的权限或菜单。菜单项通过 `permission` token 控制可见性，前端静态路由只注册页面，最终菜单显示以后端 `/api/auth/me` 返回的菜单为准。

HTTP adapter 只做请求解析、validator 校验、DTO 转换、调用 usecase、统一响应。核心业务规则放在 `domain` 和 `usecase`。

## HTTP API

系统路由：

| Method | Path | 用途 |
| --- | --- | --- |
| `GET` | `/api/health` | 进程存活检查 |
| `GET` | `/api/info` | 应用名称、版本和当前时间 |
| `GET` | `/api/ready` | 整体 readiness |
| `GET` | `/api/capabilities` | capability 状态详情 |

后台基础路由：

| Method | Path | 用途 |
| --- | --- | --- |
| `POST` | `/api/auth/login` | 管理员登录 |
| `POST` | `/api/auth/role` | 切换当前生效角色并重新签发 token |
| `POST` | `/api/auth/logout` | 客户端退出登录 |
| `GET` | `/api/auth/me` | 当前管理员 |
| `GET` | `/api/admins` | 管理员列表 |
| `POST` | `/api/admins` | 创建管理员 |
| `PATCH` | `/api/admins/:id` | 更新管理员 |
| `GET` | `/api/roles` | 角色列表 |
| `POST` | `/api/roles` | 创建角色 |
| `PATCH` | `/api/roles/:id` | 更新角色 |
| `GET` | `/api/permissions` | 权限目录元数据 |
| `GET` | `/api/menus` | 菜单列表 |
| `POST` | `/api/menus` | 创建菜单 |
| `PATCH` | `/api/menus/:id` | 更新菜单 |
| `GET` | `/api/system/configs` | 系统配置列表 |
| `PUT` | `/api/system/configs/:key` | 创建或更新系统配置 |
| `GET` | `/api/dictionaries` | 字典列表 |
| `POST` | `/api/dictionaries` | 创建字典 |
| `POST` | `/api/dictionaries/:code/items` | 新增字典项 |
| `PATCH` | `/api/dictionaries/:code/items/:item_id` | 更新字典项 |
| `GET` | `/api/files` | 文件列表 |
| `POST` | `/api/files` | 上传文件 |
| `GET` | `/api/logs/operations` | 操作日志 |
| `GET` | `/api/logs/logins` | 登录日志 |

## 新业务模块

新增业务模块时，按这个顺序写：

1. `internal/modules/<module>/domain`
2. `internal/modules/<module>/usecase`
3. `internal/modules/<module>/adapters/mysql`
4. `internal/modules/<module>/http`
5. `internal/boot/business.go`
6. 行为测试

职责边界：

- `domain` 放领域对象、构造函数、领域错误和不可变业务规则。
- `usecase` 放输入输出、业务流程和 usecase-owned outbound interface。
- `adapters/<adapter>` 实现 usecase 定义的 interface。
- `http` 复用 `internal/platform/server/httpreq` 和 `internal/platform/server/httpresp`。
- `internal/boot` 负责选择 adapter、装配依赖、挂载路由。

跨模块依赖不要直接 import 对方 store 或 adapter。由消费方 usecase 定义小接口，再在 `internal/boot` 写 bridge。

## 验证

后端：

```bash
make test
make lint
make build
make verify
```

前端：

```bash
cd web
npm run lint
npm run test
npm run build
```

修改 Go 依赖后必须运行 `go mod tidy`。前端构建产物 `web/dist`、Umi 生成目录和 Utoo cache 不提交。

## 进一步阅读

- `AGENTS.md`：仓库级开发约束。
- `configs/README.md`：配置项和环境变量覆盖规则。
- `internal/boot/README.md`：composition root 和模块装配说明。
- `internal/platform/server/README.md`：HTTP runtime 边界说明。
- `web/README.md`：前端开发说明。
