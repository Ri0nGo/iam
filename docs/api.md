# IAM API 文档

## 基础说明

- Base URL: `/api/v1`
- 后台管理接口需要登录后获取的 `access_token` 鉴权。
- 通用响应结构：所有 JSON 接口均使用统一包裹结构。

```json
{
  "code": 0,
  "msg": "ok",
  "data": {}
}
```

说明：

- `code=0` 表示业务处理成功。
- HTTP 状态码仍用于表达请求是否成功。
- OAuth2 token 相关接口的业务数据放在 `data` 字段中。

## 认证接口

### 登录

- Method: `POST`
- Path: `/auth/login`
- Auth: 否

请求：

```json
{
  "username": "admin",
  "password": "123456"
}
```

响应：

```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "access_token": "jwt-token",
    "expires_in": 7200,
    "user": {
      "id": 1,
      "username": "admin",
      "display_name": "系统管理员",
      "status": 1,
      "roles": ["admin"]
    }
  }
}
```

### 退出登录

- Method: `POST`
- Path: `/auth/logout`
- Auth: 是

响应：

```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "success": true
  }
}
```

### 获取当前用户

- Method: `GET`
- Path: `/auth/me`
- Auth: 是

响应：

```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "openid": "ou_xxx",
    "username": "admin",
    "display_name": "系统管理员",
    "status": 1,
    "roles": ["admin"]
  }
}
```

## OAuth2 接口

### 授权码申请

- Method: `GET`
- Path: `/oauth/authorize`
- Auth: 浏览器入口不需要手动传 token；未登录会跳转到 IAM 登录页，已登录会使用 `iam_access_token` Cookie 生成授权码

Query 参数：

- `response_type`: 固定为 `code`
- `client_id`: 客户端 ID，例如 `system-a`
- `redirect_uri`: 回调地址，必须和客户端注册地址严格一致
- `scope`: 授权范围，可选
- `state`: 客户端状态参数，可选

示例：

```text
GET /api/v1/oauth/authorize?response_type=code&client_id=system-a&redirect_uri=http://system-a.local/callback&state=xyz&scope=basic
```

成功响应：

```http
HTTP/1.1 302 Found
Location: http://system-a.local/callback?code=authorization-code&state=xyz
```

未登录响应：

```http
HTTP/1.1 302 Found
Location: http://localhost:5173/login?redirect=...
```

### 授权码换令牌

- Method: `POST`
- Path: `/oauth/token`
- Auth: 否
- Response: 使用统一响应结构，token 数据放在 `data` 中

请求：

```json
{
  "grant_type": "authorization_code",
  "client_id": "system-a",
  "secret": "system-a-secret",
  "code": "authorization-code",
  "redirect_uri": "http://system-a.local/callback"
}
```

响应：

```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "access_token": "jwt-token",
    "expires_in": 7200,
    "scope": "basic"
  }
}
```

### OAuth2 用户信息

- Method: `GET`
- Path: `/oauth/userinfo?access_token=ACCESS_TOKEN`
- Auth: 是，通过 `access_token` query 参数传递

响应：

```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "id": 1,
    "username": "admin",
    "display_name": "系统管理员",
    "status": 1,
    "roles": ["admin"]
  }
}
```

## 用户接口

## 认证应用接口

认证应用用于管理 OAuth2 客户端，对应后端 `oauth_clients` 表。所有接口都需要管理员登录后的 access token。

### 创建认证应用

- Method: `POST`
- Path: `/auth-applications`
- Auth: 是

请求：

```json
{
  "name": "系统 A OAuth2 认证",
  "code": "system-a-oauth2",
  "client_id": "system-a",
  "secret_key": "system-a-secret",
  "response_type": "code",
  "redirect_uri": "http://system-a.local/callback",
  "status": 1,
  "remark": "系统 A 认证应用"
}
```

响应：

```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "id": 1,
    "name": "系统 A OAuth2 认证",
    "code": "system-a-oauth2",
    "client_id": "system-a",
    "secret_key": "system-a-secret",
    "response_type": "code",
    "redirect_uri": "http://system-a.local/callback",
    "status": 1,
    "remark": "系统 A 认证应用"
  }
}
```

### 认证应用列表

- Method: `GET`
- Path: `/auth-applications`
- Auth: 是

Query 参数：

- `keyword`: 可选，按 `name`、`code`、`client_id`、`redirect_uri` 模糊搜索
- `status`: 可选，`1` 启用，`2` 禁用

### 查询认证应用详情

- Method: `GET`
- Path: `/auth-applications/:id`
- Auth: 是

### 更新认证应用

- Method: `PUT`
- Path: `/auth-applications/:id`
- Auth: 是

请求：

```json
{
  "name": "系统 A OAuth2 认证",
  "code": "system-a-oauth2",
  "client_id": "system-a",
  "secret_key": "system-a-secret",
  "response_type": "code",
  "redirect_uri": "http://system-a.local/callback",
  "status": 1,
  "remark": "系统 A 认证应用"
}
```

### 删除认证应用

- Method: `DELETE`
- Path: `/auth-applications/:id`
- Auth: 是

响应：

```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "success": true
  }
}
```

### 创建用户

- Method: `POST`
- Path: `/users`
- Auth: 是

请求：

```json
{
  "username": "alice",
  "password": "123456",
  "display_name": "Alice",
  "email": "alice@example.com",
  "mobile": "13800000000",
  "remark": "demo user",
  "role_codes": ["operator"]
}
```

说明：

- `openid` 由后端自动生成，创建用户时前端不需要传入。
- 当前格式为 `ou_` 前缀加随机十六进制字符串，例如 `ou_6f2d...`。
- `openid` 不是全数字字符串；业界常见做法是使用不透明、不可推断的字符串标识。

### 用户列表

- Method: `GET`
- Path: `/users`
- Auth: 是

Query 参数：

- `keyword`: 可选，按 `username`、`display_name`、`email`、`mobile` 模糊搜索
- `status`: 可选，`1` 启用，`2` 禁用

### 查询用户详情

- Method: `GET`
- Path: `/users/:id`
- Auth: 是

### 修改用户状态

- Method: `PUT`
- Path: `/users/:id/status`
- Auth: 是

请求：

```json
{
  "status": 2
}
```

### 重置用户密码

- Method: `PUT`
- Path: `/users/:id/password`
- Auth: 是

请求：

```json
{
  "password": "new-password"
}
```

### 查询用户角色

- Method: `GET`
- Path: `/users/:id/roles`
- Auth: 是

### 绑定用户角色

- Method: `PUT`
- Path: `/users/:id/roles`
- Auth: 是

请求：

```json
{
  "role_codes": ["admin", "operator"]
}
```

## 角色接口

### 创建角色

- Method: `POST`
- Path: `/roles`
- Auth: 是

请求：

```json
{
  "code": "operator",
  "name": "运营角色",
  "remark": "业务运营"
}
```

### 角色列表

- Method: `GET`
- Path: `/roles`
- Auth: 是

## 默认数据

服务首次启动会自动初始化：

- 默认角色：`admin`
- 默认账号：`admin`
- 默认密码：`123456`

OAuth2 客户端不再默认初始化，请在认证管理中按需创建。

## 配置说明

当前默认配置已经按开发环境写入 `configs/config.yaml`：

- MySQL: `ubuntu:3306`, `root/123456`, db 默认为 `iam`
- Redis: `ubuntu:6379`, 密码为空, `db=3`

## 测试覆盖

接口级测试位于 `internal` 目录：

- `internal/auth_api_test.go`
- `internal/auth_application_api_test.go`
- `internal/oauth_api_test.go`
- `internal/user_api_test.go`
- `internal/role_api_test.go`
- `internal/test_helper_test.go`

执行：

```bash
go test ./...
```
