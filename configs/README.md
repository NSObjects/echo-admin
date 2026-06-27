# 配置说明

项目默认只加载一个静态配置文件：

```go
cfg, err := configs.Load("configs/config.toml")
```

配置文件支持 TOML、YAML、JSON，格式由文件后缀识别。无后缀时按 TOML 解析，未知后缀会在启动时失败。未知配置字段也会在启动时失败。环境变量可以覆盖同名配置项。

## 当前配置项

```toml
[app]
name = "echo-admin"
version = "dev"

[system]
port = ":9322"
level = 1 # 1=debug, 2=online

[log]
format = "console" # console, json
output = "stdout" # stdout, stderr
caller = false

[http]
recovery_disabled = false
request_context_disabled = false
request_log_disabled = false
gzip_disabled = false

[http.cors]
enabled = false
allow_origins = []
allow_methods = []
allow_headers = []
allow_credentials = false
expose_headers = []
max_age_seconds = 0

[jwt]
enabled = true
secret = ""
skip_paths = ["/api/health", "/api/info", "/api/ready", "/api/capabilities", "/api/auth/login"]

[admin]
upload_dir = "uploads"
bootstrap_password = ""

[mysql]
enabled = true
dsn = "echo_admin:echo_admin_dev_password@tcp(127.0.0.1:3306)/echo_admin?charset=utf8mb4&parseTime=true&loc=Local"

[redis]
enabled = false
address = ""

[mongodb]
enabled = false
uri = ""
database = ""

[tracing]
enabled = false
endpoint = ""
protocol = "grpc" # grpc, http
insecure = true
```

HTTP middleware、MySQL、Redis、MongoDB、JWT、Jaeger tracing、admin upload directory 和首次启动管理员密码都是 configured resources。后台基础能力运行期依赖 MySQL；启用 Redis、MongoDB、tracing 等可选外部依赖后也必须提供连接配置，并会进入 `/api/ready` 和 `/api/capabilities`。

业务代码放在 `internal/modules/<module>`，平台运行时代码放在 `internal/platform`。业务 adapter 复用 boot 已经加载的资源：usecase 定义 usecase-owned outbound interface，adapter 使用配置好的 MySQL/Redis/MongoDB client 实现它，然后业务 module 用 `boot.NewModule`、`boot.Provide` 和 `boot.Route` 声明 adapter、usecase、handler 和 route。

## 环境变量覆盖

```bash
export ECHO_ADMIN_APP_NAME=echo-admin
export ECHO_ADMIN_APP_VERSION=dev
export ECHO_ADMIN_SYSTEM_PORT=:9322
export ECHO_ADMIN_SYSTEM_LEVEL=1
export ECHO_ADMIN_LOG_FORMAT=console
export ECHO_ADMIN_LOG_OUTPUT=stdout
export ECHO_ADMIN_LOG_CALLER=false
export ECHO_ADMIN_HTTP_RECOVERY_DISABLED=false
export ECHO_ADMIN_HTTP_REQUEST_CONTEXT_DISABLED=false
export ECHO_ADMIN_HTTP_REQUEST_LOG_DISABLED=false
export ECHO_ADMIN_HTTP_GZIP_DISABLED=false
export ECHO_ADMIN_HTTP_CORS_ENABLED=false
export ECHO_ADMIN_HTTP_CORS_ALLOW_ORIGINS=
export ECHO_ADMIN_HTTP_CORS_ALLOW_METHODS=
export ECHO_ADMIN_HTTP_CORS_ALLOW_HEADERS=
export ECHO_ADMIN_HTTP_CORS_ALLOW_CREDENTIALS=false
export ECHO_ADMIN_HTTP_CORS_EXPOSE_HEADERS=
export ECHO_ADMIN_HTTP_CORS_MAX_AGE_SECONDS=0
export ECHO_ADMIN_JWT_ENABLED=true
export ECHO_ADMIN_JWT_SECRET="$(openssl rand -base64 32)"
export ECHO_ADMIN_ADMIN_UPLOAD_DIR=uploads
export ECHO_ADMIN_ADMIN_BOOTSTRAP_PASSWORD=replace-with-a-private-password
export ECHO_ADMIN_MYSQL_ENABLED=true
export ECHO_ADMIN_MYSQL_DSN='echo_admin:echo_admin_dev_password@tcp(127.0.0.1:3306)/echo_admin?charset=utf8mb4&parseTime=true&loc=Local'
export ECHO_ADMIN_REDIS_ENABLED=false
export ECHO_ADMIN_REDIS_ADDRESS=
export ECHO_ADMIN_MONGODB_ENABLED=false
export ECHO_ADMIN_MONGODB_URI=
export ECHO_ADMIN_TRACING_ENABLED=false
export ECHO_ADMIN_TRACING_ENDPOINT=
export ECHO_ADMIN_TRACING_PROTOCOL=grpc
export ECHO_ADMIN_TRACING_INSECURE=true
```

JWT 默认启用，但仓库配置不会提供可启动的默认签名密钥。必须通过 `ECHO_ADMIN_JWT_SECRET` 或 Secret 管理系统提供至少 32 个字符的真实 secret，不要把真实 secret 写入 ConfigMap。空库首次启动还必须设置 `ECHO_ADMIN_ADMIN_BOOTSTRAP_PASSWORD`；它只在 `admin` 用户不存在时使用，不会覆盖后续修改过的密码。

`app.name` 和 `app.version` 会出现在 `/api/info` 响应中。`system.level` 只接受 `1` 或 `2`。`log.format` 只接受 `console` 或 `json`；`log.output` 只接受 `stdout` 或 `stderr`。`http.*_disabled` 可关闭默认安装的 HTTP middleware。

`http.cors.enabled=true` 时必须配置 `http.cors.allow_origins`。`http.cors.allow_credentials=true` 时 `allow_origins` 不能包含 `*`。环境变量中的列表项使用逗号分隔，例如 `ECHO_ADMIN_HTTP_CORS_ALLOW_ORIGINS=https://app.example.com,https://admin.example.com`。

`/api/health` 是进程存活，`/api/ready` 只返回整体 readiness，`/api/capabilities` 返回 logging、MySQL、Redis、MongoDB、Jaeger tracing 的 disabled/available/unavailable 状态。业务模块装配会直接使用 MySQL，`mysql.enabled=false` 只适合平台层单元测试，不适合默认业务运行。

`docker-compose.yaml` 只用于本地开发，公开端口默认绑定 `127.0.0.1`。默认启动 API 和 MySQL；Redis、MongoDB 和 Jaeger 在 `resources` profile 下，使用 `docker compose --profile resources up` 启动。资源服务使用 Compose 命名卷保存本地开发数据；需要重置 MySQL、Redis 和 MongoDB 时，显式运行 `docker compose down -v`。
