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

## 目录

- `src/pages`：后台页面，包括工作台、管理员、角色、菜单、配置、字典、文件和日志。
- `src/services/admin.ts`：后台 API client 和 DTO 类型。
- `src/requestErrorConfig.ts`：统一请求错误处理和 token 注入。
- `config/routes.ts`：前端路由。
- `config/config.ts`：Umi Max 配置。

## 约定

- 不使用 OpenAPI 生成器作为默认开发路径。
- 新 API 先在对应后端 business module 定义清楚，再在 `src/services/admin.ts` 增加显式方法。
- 页面只做表单、列表、状态和 DTO 转换，不承载核心业务规则。
- 构建产物 `dist`、Umi 生成目录和 Utoo cache 不提交。
