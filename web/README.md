# Echo Admin Web

Echo Admin 的中后台前端，基于 Umi Max、React、Ant Design 和 ProComponents。

## 开发命令

```sh
npm install
npm run dev
npm run lint
npm run test
npm run build
```

开发服务会把 `/api` 代理到 `ECHO_ADMIN_WEB_API_TARGET`，默认是 `http://127.0.0.1:9322`。后端接口路径以 `/api` 开头，登录后前端会把 JWT 写入 `localStorage`，并通过 `Authorization: Bearer <token>` 发送。生产部署仍按同源 `/api` 访问。

Utoo 的 persistent cache 使用项目级 `.turbopack/lock`。如果本地 dev server 正在运行，构建验证可使用 `ECHO_ADMIN_WEB_DISABLE_UTOOPACK_CACHE=1 npm run build` 禁用 persistent cache，避免停止当前 dev server。若 CI 或沙箱环境不允许 utoopack 创建子进程，可使用 `ECHO_ADMIN_WEB_DISABLE_UTOOPACK=1 npm run build` 走默认 webpack 构建。

登录态中的当前角色由后端 `/api/auth/me` 返回。切换角色时调用 `/api/auth/role`，前端会保存后端重新签发的 token，并按新角色的菜单、按钮权限、数据权限和默认入口刷新界面。工作台展示当前用户、已授权能力、后台菜单、应用信息和 capability 状态。头像菜单可更新当前用户展示名和邮箱，也可修改当前密码；改密成功后服务端会拉黑当前 JWT，前端清理本地 token 并回到登录页。退出登录会调用 `/api/auth/logout` 拉黑当前 JWT，然后清理本地 token。API Token 页面创建 token 后只展示一次明文，后续列表只显示 token 前缀、状态、归属和使用时间。系统参数页面支持名称/键筛选、分页、详情、创建、更新、单条删除和批量删除。数据字典页面支持字典 JSON 导出和导入，字典项支持树形父子关系、扩展值和新增子项。版本管理页面维护发布记录，支持批量删除、选择菜单/API/字典导出版本 JSON，并导入版本 JSON。文件页支持分类树、本地上传、外部 URL 导入、重命名和删除，上传和导入 URL 可选择分类；URL 导入只登记元数据，不由服务端下载远端文件。日志页包含操作日志、登录日志和系统错误三个视图，支持单条和批量删除；系统错误可标记已处理、填写处理备注和取消处理状态。

## 目录

- `src/pages`：后台页面，包括工作台、管理员、角色、API、API Token、菜单、配置、参数、版本、字典、文件和日志。
- `src/services/admin.ts`：后台 API client 和 DTO 类型。
- `src/requestErrorConfig.ts`：统一请求错误处理和 token 注入。
- `config/routes.ts`：前端路由。
- `config/config.ts`：Umi Max 配置。

## 约定

- 不使用 OpenAPI 生成器作为默认开发路径。
- 新 API 先在对应后端 business module 定义清楚，再在 `src/services/admin.ts` 增加显式方法。
- 页面只做表单、列表、状态和 DTO 转换，不承载核心业务规则。
- 路由可见性使用后端菜单控制，`hidden` 菜单不会进入侧边栏；页面写操作按钮使用 `resource:action` 权限 token 控制，并随菜单按钮种子一起管理；角色编辑页会同时提交菜单 `menu_ids`、API `api_ids`、菜单按钮 `button_ids` 和管理员列表数据范围 `data_role_ids`。
- 构建产物 `dist`、Umi 生成目录和 Utoo cache 不提交。
