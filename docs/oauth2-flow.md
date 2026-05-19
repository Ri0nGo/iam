# OAuth2 认证流程

当前 IAM 的 OAuth2 登录流程参照微信移动应用登录模型实现，采用 `authorization_code` 模式。

## 核心约定

- Base URL: `/api/v1`
- 授权码 `code` 一次性使用，默认有效期 `300` 秒。
- OAuth2 `access_token` 默认有效期 `7200` 秒。
- OAuth2 `refresh_token` 默认有效期 `180` 天。
- 控制台登录 token 与 OAuth2 token 已隔离。
- 控制台登录 token 只能访问 IAM 管理接口。
- OAuth2 token 只能访问 OAuth2 资源接口。

## Token 类型隔离

JWT 中通过 `token_use` 区分用途：

| token 来源 | token_use | 用途 |
|---|---|---|
| `/auth/login` | `console` | IAM 控制台和管理接口 |
| `/oauth/token` | `oauth2` | 第三方应用调用 OAuth2 资源接口 |

隔离规则：

- `/oauth/userinfo` 拒绝 `console` token。
- IAM 管理接口拒绝 `oauth2` token。
- OAuth2 token 绑定 `client_id`、`scope` 和用户。

## 1. 请求授权码

第三方应用引导用户访问 IAM 授权地址：

```text
GET /api/v1/oauth/authorize?response_type=code&client_id=APPID&redirect_uri=REDIRECT_URI&scope=snsapi_userinfo&state=STATE
```

参数：

| 参数 | 必填 | 说明 |
|---|---|---|
| `response_type` | 是 | 固定为 `code` |
| `client_id` | 是 | 认证应用的客户端 ID |
| `redirect_uri` | 是 | 回调地址，必须与认证应用配置严格一致 |
| `scope` | 否 | 授权范围，示例：`snsapi_userinfo` |
| `state` | 否 | 客户端状态参数，回调时原样返回 |

如果用户未登录 IAM，会跳转到登录页：

```http
HTTP/1.1 302 Found
Location: http://localhost:5173/login?redirect=...
```

如果用户已登录，IAM 会生成授权码并跳转回客户端。该阶段只返回 `code` 和 `state`，不会返回 `openid`：

```http
HTTP/1.1 302 Found
Location: REDIRECT_URI?code=CODE&state=STATE
```

当前实现没有单独的授权确认页，属于“登录即授权”的简化模式。

## 2. 通过 code 换 access_token

对照微信 `/sns/oauth2/access_token`：

```text
GET /api/v1/oauth/token?client_id=CLIENT_ID&secret=SECRET&code=CODE&grant_type=authorization_code
```

curl 示例：

```bash
curl "http://localhost:8080/api/v1/oauth/token?client_id=system-a&secret=system-a-secret&code=<code>&grant_type=authorization_code"
```

参数：

| 参数 | 必填 | 说明 |
|---|---|---|
| `client_id` | 是 | 认证应用的 `client_id` |
| `secret` | 是 | 认证应用的 `secret_key` |
| `code` | 是 | 上一步回调返回的授权码 |
| `grant_type` | 是 | 固定为 `authorization_code` |

成功响应：

```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "access_token": "oauth2-jwt-token",
    "expires_in": 7200,
    "refresh_token": "refresh-token",
    "openid": "ou_admin",
    "scope": "snsapi_userinfo"
  }
}
```

说明：

- `code` 只能成功兑换一次。
- `access_token` 是 OAuth2 token，不能访问 IAM 管理接口。
- `openid` 在换 token 阶段返回，不在授权回调阶段返回；后续获取用户信息时需要携带。
- `secret` 对应认证应用的 `secret_key`；不使用 `client_secret` 入参。

## 3. 校验 access_token

对照微信 `/sns/auth`：

```text
GET /api/v1/oauth/auth?access_token=ACCESS_TOKEN&openid=OPENID
```

curl 示例：

```bash
curl "http://localhost:8080/api/v1/oauth/auth?access_token=<access_token>&openid=<openid>"
```

成功响应：

```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "success": true
  }
}
```

校验内容：

- `access_token` 有效且未过期。
- `access_token` 是 OAuth2 token。
- `openid` 与 token 对应用户一致。

## 4. 刷新 access_token

对照微信 `/sns/oauth2/refresh_token`：

```text
GET /api/v1/oauth/refresh_token?client_id=CLIENT_ID&grant_type=refresh_token&refresh_token=REFRESH_TOKEN
```

curl 示例：

```bash
curl "http://localhost:8080/api/v1/oauth/refresh_token?client_id=system-a&grant_type=refresh_token&refresh_token=<refresh_token>"
```

参数：

| 参数 | 必填 | 说明 |
|---|---|---|
| `client_id` | 是 | 认证应用的 `client_id` |
| `grant_type` | 是 | 固定为 `refresh_token` |
| `refresh_token` | 是 | 换 token 时返回的长效刷新凭证 |

成功响应：

```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "access_token": "new-oauth2-jwt-token",
    "expires_in": 7200,
    "refresh_token": "refresh-token",
    "openid": "ou_admin",
    "scope": "snsapi_userinfo"
  }
}
```

说明：

- `refresh_token` 绑定 `client_id` 和用户。
- `refresh_token` 泄露后风险较高，应保存在客户端服务端。
- 当前实现刷新时复用原 `refresh_token`，返回新的 `access_token`。

## 5. 获取用户信息

对照微信 `/sns/userinfo`：

```text
GET /api/v1/oauth/userinfo?access_token=ACCESS_TOKEN&openid=OPENID
```

curl 示例：

```bash
curl "http://localhost:8080/api/v1/oauth/userinfo?access_token=<access_token>&openid=<openid>"
```

成功响应：

```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "openid": "ou_admin",
    "username": "admin",
    "display_name": "系统管理员",
    "status": 1,
    "roles": ["admin"]
  }
}
```

注意：

- `access_token` 和 `openid` 都通过 query 参数传递。
- 不使用 `Authorization: Bearer` Header。
- 控制台登录 token 调用该接口会被拒绝。
- 当前返回中不包含 IAM 内部用户 `id`，避免把内部主键暴露给第三方应用。

## 完整 curl 示例

```bash
# 1. 浏览器访问授权地址，获取回调中的 code
http://localhost:8080/api/v1/oauth/authorize?response_type=code&client_id=system-a&redirect_uri=http://system-a.local/callback&scope=snsapi_userinfo&state=demo_state

# 2. code 换 token
curl "http://localhost:8080/api/v1/oauth/token?client_id=system-a&secret=system-a-secret&code=<code>&grant_type=authorization_code"

# 3. 校验 token
curl "http://localhost:8080/api/v1/oauth/auth?access_token=<access_token>&openid=<openid>"

# 4. 刷新 token
curl "http://localhost:8080/api/v1/oauth/refresh_token?client_id=system-a&grant_type=refresh_token&refresh_token=<refresh_token>"

# 5. 获取用户信息
curl "http://localhost:8080/api/v1/oauth/userinfo?access_token=<access_token>&openid=<openid>"
```

## 与微信模型的对应关系

| 微信参数/接口 | IAM 当前实现 |
|---|---|
| `client_id` | 认证应用 `client_id` |
| `secret` | 认证应用 `secret_key` |
| `/sns/oauth2/access_token` | `/api/v1/oauth/token` |
| `/sns/oauth2/refresh_token` | `/api/v1/oauth/refresh_token` |
| `/sns/auth` | `/api/v1/oauth/auth` |
| `/sns/userinfo` | `/api/v1/oauth/userinfo` |
| `openid` | IAM 用户的 `openid` |

## 后续可增强项

- 增加授权确认页，避免当前“登录即授权”。
- 增加用户对应用授权记录，例如 `oauth_user_grants`。
- 增加撤销授权和撤销 `refresh_token` 能力。
- 对 `scope` 做更严格的注册、校验和资源授权控制。
- 对 query 中的 `access_token` 做日志脱敏，避免 token 进入访问日志。
