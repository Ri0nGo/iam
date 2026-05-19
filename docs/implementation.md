# IAM 具体实现方案

## 1. 目标

基于 `Gin + Redis + MySQL + Viper + GORM` 实现一个独立的 `IAM` 服务，提供统一用户管理、认证、令牌签发、用户识别和角色管理能力。

当前阶段目标：

- 支持用户名密码登录。
- 支持统一用户管理。
- 支持基础角色管理。
- 支持 `JWT` 访问令牌签发与校验。
- 支持使用 `Redis` 管理登录态、黑名单和验证码等缓存能力。
- 支持使用配置文件驱动运行配置。

后续预留目标：

- 支持 `QQ`、`GitHub` 等第三方登录。
- 支持统一登录页。
- 支持刷新令牌、单点登录、设备管理。
- 支持更细粒度权限模型。

## 2. 技术选型

### 2.1 框架与组件

- Web 框架：`Gin`
- 配置管理：`Viper`
- 数据库：`MySQL 8.x`
- 缓存：`Redis`
- ORM：`GORM`
- 日志：`log/slog`
- 密码加密：`bcrypt`
- 令牌：`JWT`

### 2.2 选型原因

- `Gin` 轻量、成熟，适合中小型认证中心快速落地。
- `Viper` 适合读取 `yaml` 配置，并支持环境变量覆盖。
- `MySQL` 负责持久化用户、角色、第三方账号绑定信息。
- `Redis` 适合保存令牌黑名单、验证码、限流计数和少量会话态缓存。
- `GORM` 适合快速完成模型映射、基础查询和迁移控制。
- `slog` 是 Go 官方标准日志库，生态稳定，便于统一结构化日志输出。

## 3. 总体架构

```text
Client
  |
  v
Gin Router
  |
  v
Handler -> Service -> Repository -> MySQL
             |
             +-> Redis
             |
             +-> JWT
             |
             +-> slog Logger
```

职责分层：

- `Handler`：处理 HTTP 请求、参数校验、响应格式。
- `Service`：处理认证、用户管理、角色分配等业务逻辑。
- `Repository`：基于 `GORM` 访问 `MySQL`，按需访问 `Redis`。
- `Middleware`：处理鉴权、日志、追踪、错误恢复。

## 3.1 依赖注入设计

当前文档原先没有单独展开依赖注入设计，建议补上，并且这个项目适合使用依赖注入。

推荐原则：

- 统一使用构造函数注入，不使用全局变量承载核心依赖。
- 在 `bootstrap` 层集中组装 `Config`、`*gorm.DB`、`*redis.Client`、`*slog.Logger`、`JWT Manager`、`Repository`、`Service`、`Handler`。
- 业务层只依赖接口或明确的结构体，不直接感知初始化细节。

建议启动装配顺序：

```text
Load Config
  -> Init slog
  -> Init MySQL(GORM)
  -> Init Redis
  -> Init JWT Manager
  -> Build Repository
  -> Build Service
  -> Build Handler
  -> Register Router
```

### 是否需要依赖注入工具

当前阶段不建议引入额外依赖注入工具，直接使用手动构造函数注入即可。

原因：

- 当前模块规模可控，手动装配最清晰。
- Go 项目中手动注入是最常见、最稳定的方式。
- 不引入额外容器，启动链路更直接，排错成本更低。

如果后续模块和生命周期明显复杂化，再评估下面两个工具：

- `uber/fx`：适合大型服务，生命周期管理能力强，但偏重。
- `uber/dig`：比 `fx` 更轻，只提供容器能力，但仍会引入运行时反射。

当前结论：

- 第一阶段采用手动依赖注入。
- 在 `bootstrap/app.go` 中集中完成依赖装配。

## 4. 核心功能范围

### 4.1 第一阶段必须实现

- 用户创建、查询、禁用、重置密码。
- 用户名密码登录。
- 获取当前登录用户信息。
- 角色分配与角色查询。
- 基于 `JWT` 的鉴权中间件。
- `Redis` 支持登录会话和退出登录。

### 4.2 第一阶段暂不实现

- `QQ` 登录。
- `GitHub` 登录。
- 多因子认证。
- 单点登出。
- 细粒度数据权限。

## 5. 项目目录设计

建议目录如下：

```text
iam/
  cmd/
    server/
      main.go
  config/
    config.yaml
  internal/
    bootstrap/
      app.go
      config.go
      db.go
      redis.go
      logger.go
      providers.go
    router/
      router.go
      middleware.go
    handler/
      auth_handler.go
      user_handler.go
      role_handler.go
    service/
      auth_service.go
      user_service.go
      role_service.go
      oauth_service.go
    repository/
      user_repo.go
      role_repo.go
      auth_identity_repo.go
      token_repo.go
    model/
      user.go
      role.go
      auth_identity.go
    dto/
      auth.go
      user.go
      role.go
    pkg/
      jwt/
        jwt.go
      password/
        bcrypt.go
      resp/
        response.go
      errs/
        error.go
      xid/
        id.go
  migrations/
    001_init.sql
  docs/
    project.md
    implementation.md
```

说明：

- `oauth_service.go` 当前用于承载 OAuth2 授权码模式，后续也继续承接第三方登录扩展。
- `auth_identity` 单独建模，用于兼容本地密码登录和未来第三方账号登录。
- `providers.go` 用于组织构造函数和依赖装配逻辑，也可以直接在 `app.go` 完成初始化。

## 6. 配置设计

### 6.1 配置文件路径

建议使用：`config/config.yaml`

### 6.2 配置示例

```yaml
app:
  name: iam
  env: dev
  host: 0.0.0.0
  port: 8080

mysql:
  host: 127.0.0.1
  port: 3306
  user: root
  password: root
  dbname: iam
  max_idle_conns: 10
  max_open_conns: 50
  conn_max_lifetime: 3600

redis:
  addr: 127.0.0.1:6379
  password: ""
  db: 0

jwt:
  issuer: iam
  secret: change-me
  expire_seconds: 7200
  refresh_expire_seconds: 604800

security:
  password_cost: 12
  login_fail_limit: 5
  login_fail_window_seconds: 900

oauth:
  authorize_code_expire_seconds: 300
```

### 6.3 配置加载方式

启动时由 `Viper` 读取配置文件，并支持环境变量覆盖。

建议规则：

- 配置文件默认路径：`config/config.yaml`
- 环境变量前缀：`IAM`
- 环境变量映射：`.` 转 `_`

例如：

- `IAM_MYSQL_HOST`
- `IAM_JWT_SECRET`
- `IAM_REDIS_ADDR`

### 6.4 配置结构体示例

```go
type Config struct {
    App      AppConfig      `mapstructure:"app"`
    MySQL    MySQLConfig    `mapstructure:"mysql"`
    Redis    RedisConfig    `mapstructure:"redis"`
    JWT      JWTConfig      `mapstructure:"jwt"`
    Security SecurityConfig `mapstructure:"security"`
    OAuth    OAuthConfig    `mapstructure:"oauth"`
}
```

## 7. 数据库表设计

## 7.1 设计原则

- 主身份与认证方式分离。
- 本地账号密码登录和第三方登录统一归到“认证标识”模型下。
- 当前只启用本地密码登录，但表结构提前兼容未来 OAuth 扩展。

## 7.2 表清单

- `users`：用户主表
- `roles`：角色表
- `user_roles`：用户角色关系表
- `auth_identities`：认证标识表，兼容本地登录和第三方登录
- `oauth_clients`：OAuth2 客户端表

## 7.3 users

```sql
CREATE TABLE users (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  username VARCHAR(64) NOT NULL,
  openid VARCHAR(128) DEFAULT NULL,
  display_name VARCHAR(128) NOT NULL DEFAULT '',
  avatar_url VARCHAR(255) NOT NULL DEFAULT '',
  mobile VARCHAR(32) NOT NULL DEFAULT '',
  email VARCHAR(128) NOT NULL DEFAULT '',
  status TINYINT NOT NULL DEFAULT 1,
  remark VARCHAR(255) NOT NULL DEFAULT '',
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  UNIQUE KEY uk_users_username (username),
  UNIQUE KEY uk_users_openid (openid),
  UNIQUE KEY uk_users_mobile (mobile),
  UNIQUE KEY uk_users_email (email)
);
```

字段说明：

- `status`: `1` 启用，`2` 禁用。
- `username` 当前阶段作为本地登录唯一用户名。
- `openid` 先预留为外部身份唯一标识字段，允许为 `NULL`，并添加唯一索引。

## 7.4 roles

```sql
CREATE TABLE roles (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  code VARCHAR(64) NOT NULL,
  name VARCHAR(128) NOT NULL,
  status TINYINT NOT NULL DEFAULT 1,
  remark VARCHAR(255) NOT NULL DEFAULT '',
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  UNIQUE KEY uk_roles_code (code)
);
```

## 7.5 user_roles

```sql
CREATE TABLE user_roles (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  user_id BIGINT NOT NULL,
  role_id BIGINT NOT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE KEY uk_user_role (user_id, role_id),
  KEY idx_user_roles_user_id (user_id),
  KEY idx_user_roles_role_id (role_id)
);
```

## 7.6 auth_identities

```sql
CREATE TABLE auth_identities (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  user_id BIGINT NOT NULL,
  identity_type VARCHAR(32) NOT NULL,
  identifier VARCHAR(128) NOT NULL,
  credential VARCHAR(255) NOT NULL DEFAULT '',
  extra JSON NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  UNIQUE KEY uk_identity_type_identifier (identity_type, identifier),
  KEY idx_auth_identities_user_id (user_id)
);
```

字段说明：

- `identity_type` 可取：`password`、`qq`、`github`。
- `identifier`：本地密码登录时可存 `username`，第三方登录时可存第三方平台 `openid` / `unionid` / `github user id`。
- `credential`：本地密码登录时存密码哈希；第三方登录通常为空。
- `extra`：存第三方授权附加信息，比如 `access_token`、`refresh_token`、`scope`、头像地址等。

这个表是后续扩展的关键。即使当前只支持用户名密码，也建议第一版直接建出来。

## 7.7 oauth_clients

```sql
CREATE TABLE oauth_clients (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  client_id VARCHAR(64) NOT NULL,
  client_secret VARCHAR(128) NOT NULL,
  name VARCHAR(128) NOT NULL,
  redirect_uri VARCHAR(255) NOT NULL,
  status TINYINT NOT NULL DEFAULT 1,
  remark VARCHAR(255) NOT NULL DEFAULT '',
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  UNIQUE KEY uk_oauth_clients_client_id (client_id)
);
```

字段说明：

- `client_id`：客户端唯一标识。
- `client_secret`：客户端密钥。
- `redirect_uri`：授权码回调地址，必须严格匹配。

## 8. Redis 设计

建议键设计如下：

- `iam:login:fail:{username}`：登录失败次数
- `iam:token:blacklist:{jti}`：已注销令牌黑名单
- `iam:oauth:code:{code}`：OAuth2 授权码缓存
- `iam:captcha:{scene}:{target}`：验证码

当前阶段最小使用方式：

- 记录登录失败次数，防爆破。
- 退出登录时将 `JWT jti` 放入黑名单。
- 存储 OAuth2 授权码，并控制短时有效期和单次使用。
- 可选缓存用户基础信息，降低数据库压力。
- 如果后续需要设备会话管理，再补 `login_sessions` 表，不在第一阶段落库。

## 9. 认证与鉴权设计

## 9.1 登录流程

```text
1. 用户提交 username/password
2. 服务根据 auth_identities 查询 identity_type=password and identifier=username
3. 校验 credential 与输入密码是否匹配
4. 检查 users.status 是否启用
5. 查询用户角色
6. 生成 JWT access token
7. 返回 token 和用户基础信息
```

## 9.2 JWT 载荷建议

```json
{
  "sub": "10001",
  "username": "alice",
  "roles": ["admin"],
  "status": 1,
  "jti": "session_xxx",
  "iss": "iam",
  "iat": 1710000000,
  "exp": 1710007200
}
```

说明：

- `sub` 存 `userId`
- `jti` 用于退出登录和黑名单控制
- `roles` 只放稳定角色，不建议放过多动态权限明细

## 9.3 鉴权中间件

鉴权中间件职责：

- 校验 `Authorization: <access_token>` 或登录 Cookie
- 校验签名、过期时间、发行方
- 校验 `jti` 是否在 `Redis` 黑名单中
- 将用户上下文写入 `Gin Context`

上下文字段建议：

- `userId`
- `username`
- `roles`
- `tokenId`

## 9.4 退出登录

退出时：

1. 解析当前令牌的 `jti`
2. 将其写入 `Redis` 黑名单，过期时间与令牌剩余有效期一致
3. 后续请求若命中黑名单，则判定为未登录

## 9.5 OAuth2 授权码流程

当前已实现 OAuth2 授权码模式，核心流程如下：

```text
1. 客户端携带 client_id / redirect_uri / state 请求 /oauth/authorize
2. IAM 校验客户端和回调地址
3. IAM 读取当前登录态，生成短时有效 authorization code，并写入 Redis
4. IAM 302 跳转到 redirect_uri，并携带 code / state
5. 客户端调用 /oauth/token，用 code 换 access token
6. 客户端携带 access token 调用 /oauth/userinfo
```

说明：

- `authorization code` 为一次性凭证，换取成功后立即删除。
- `redirect_uri` 采用严格匹配，防止回调地址被篡改。
- IAM 自身后台保留 `/auth/login` 登录接口，登录成功后写入授权端点可读取的 Cookie。

## 10. API 设计

### 10.1 认证接口

- `POST /api/v1/auth/login`
- `POST /api/v1/auth/logout`
- `GET /api/v1/auth/me`

### 10.2 OAuth2 接口

- `GET /api/v1/oauth/authorize`
- `POST /api/v1/oauth/token`
- `GET /api/v1/oauth/userinfo`

### 10.3 用户接口

- `POST /api/v1/users`
- `GET /api/v1/users/:id`
- `GET /api/v1/users`
- `PUT /api/v1/users/:id/status`
- `PUT /api/v1/users/:id/password`

### 10.4 角色接口

- `POST /api/v1/roles`
- `GET /api/v1/roles`
- `PUT /api/v1/users/:id/roles`
- `GET /api/v1/users/:id/roles`

### 10.5 第三方登录预留接口

当前不开放，但建议提前保留路由设计：

- `GET /api/v1/oauth/:provider/url`
- `GET /api/v1/oauth/:provider/callback`
- `POST /api/v1/oauth/:provider/bind`

其中 `provider` 可取：

- `qq`
- `github`

## 11. 模块设计

## 11.1 AuthService

负责：

- 登录认证
- 签发令牌
- 退出登录
- 获取当前用户信息
- 编排本地密码登录与未来第三方登录入口

建议接口：

```go
type AuthService interface {
    Login(ctx context.Context, req LoginRequest) (*LoginResponse, error)
    Logout(ctx context.Context, token string) error
    Me(ctx context.Context, userID int64) (*CurrentUser, error)
}
```

## 11.2 UserService

负责：

- 创建用户
- 查询用户
- 修改状态
- 重置密码

## 11.3 RoleService

负责：

- 创建角色
- 角色列表
- 用户角色绑定

## 11.4 OAuthService

当前阶段已经用于承载 OAuth2 授权码逻辑，同时继续为第三方登录预留扩展位：

```go
type OAuthService interface {
    Authorize(ctx context.Context, query AuthorizeQuery, userID uint64) (*AuthorizeResponse, error)
    Token(ctx context.Context, req TokenRequest) (*TokenResponse, error)
    UserInfo(ctx context.Context, token string) (*CurrentUser, error)
}
```

这样后续接入 `QQ`、`GitHub` 时，可以继续在该服务下增加 provider 适配，而不需要推翻现有 OAuth2 客户端模型。

## 12. 第三方登录扩展设计

虽然当前不支持 `QQ`、`GitHub` 登录，但现在就要在设计上预留扩展位。

## 12.1 扩展原则

- 不改变 `users` 主表模型。
- 所有登录方式统一归到 `auth_identities`。
- 本地密码登录只是 `identity_type=password` 的一种实现。
- 第三方登录成功后，仍然映射回同一个内部 `userId`。

## 12.2 接入方式

未来接入 `QQ` / `GitHub` 时：

1. 前端跳转到第三方授权页。
2. 第三方回调 `IAM`。
3. `IAM` 通过 `code` 换取第三方用户标识。
4. 优先根据 `auth_identities(identity_type, identifier)` 查找是否已绑定本地用户。
5. 若已绑定，则直接登录。
6. 若未绑定，则引导绑定已有账号或创建新账号。

## 12.3 为什么不直接把第三方字段写进 users

- 一个用户可能绑定多个第三方平台。
- 一个平台未来也可能有多种标识，例如 `openid` 和 `unionid`。
- `users.openid` 只能作为预留的单值唯一字段，不能替代完整的多身份绑定模型。
- 因此主扩展模型仍然应该是 `auth_identities`。

## 13. 安全建议

- 密码只存 `bcrypt hash`，不存明文。
- 登录接口做限流与失败次数控制。
- `JWT secret` 必须通过环境变量覆盖生产配置。
- 重要操作记录审计日志。
- 用户被禁用后，后续请求建议拒绝并要求重新登录。
- 第三方登录接入时，必须校验 `state` 防止 CSRF。
- 日志中不要打印明文密码、完整令牌、第三方 `client_secret`。

## 14. 落地步骤建议

### 第一步

- 完成项目基础目录。
- 接入 `Viper`、`Gin`、`GORM`、`MySQL`、`Redis`、`slog`。
- 建立数据库初始化脚本。
- 确定依赖注入方式，优先使用构造函数注入。

### 第二步

- 完成 `users`、`roles`、`user_roles`、`auth_identities` 四张核心表。
- 先实现用户名密码登录。

### 第三步

- 实现 `JWT` 中间件、`/login`、`/logout`、`/me`。
- 实现用户管理和角色绑定接口。

### 第四步

- 增加 `Redis` 登录失败次数控制、令牌黑名单。
- 补充基础审计日志。

### 第五步

- 实现 `oauth_clients`、授权码缓存和 `/oauth/token`。
- 后续按需接入 `QQ`、`GitHub`。

## 15. 推荐结论

当前最合适的具体实现方案是：

- 用 `Gin` 构建独立 IAM 服务。
- 用 `Viper` 读取 `yaml` 配置，并支持环境变量覆盖。
- 用 `GORM` 管理模型与数据库访问。
- 用 `MySQL` 保存用户、角色、认证标识等核心数据。
- 用 `Redis` 管理登录失败次数、令牌黑名单和会话缓存。
- 用 `slog` 统一输出结构化日志。
- 用 `auth_identities` 统一抽象登录方式，为未来 `QQ`、`GitHub` 登录预留扩展能力。
- 用构造函数注入组织依赖，在 `bootstrap` 层集中完成依赖装配。

这样当前阶段既能支持 IAM 自身后台登录，也能以 OAuth2 方式给系统A、系统B等客户端提供标准化接入能力。
