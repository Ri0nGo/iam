create table auth_identities
(
    id            bigint unsigned auto_increment
        primary key,
    user_id       bigint unsigned         not null,
    identity_type varchar(32)             not null,
    identifier    varchar(128)            not null,
    credential    varchar(255) default '' not null,
    extra         text                    null,
    created_at    datetime(3)             null,
    updated_at    datetime(3)             null,
    constraint uk_identity
        unique (identity_type, identifier),
    constraint uk_identity_type_identifier
        unique (identity_type, identifier)
);

create index idx_auth_identities_user_id
    on auth_identities (user_id);

create table oauth_clients
(
    id            bigint unsigned auto_increment
        primary key,
    client_id     varchar(64)                 not null,
    client_secret varchar(128)                not null,
    name          varchar(128)                not null,
    redirect_uri  varchar(255)                not null,
    status        tinyint      default 1      not null,
    remark        varchar(255) default ''     not null,
    created_at    datetime(3)                 null,
    updated_at    datetime(3)                 null,
    code          varchar(64)  default ''     not null,
    response_type varchar(32)  default 'code' not null,
    constraint idx_oauth_clients_client_id
        unique (client_id)
);

create table roles
(
    id         bigint unsigned auto_increment
        primary key,
    code       varchar(64)             not null,
    name       varchar(128)            not null,
    status     tinyint      default 1  not null,
    remark     varchar(255) default '' not null,
    created_at datetime(3)             null,
    updated_at datetime(3)             null,
    constraint idx_roles_code
        unique (code)
);

create table user_roles
(
    user_id bigint unsigned not null,
    role_id bigint unsigned not null,
    primary key (user_id, role_id)
);

create table users
(
    id            bigint unsigned auto_increment
        primary key,
    username      varchar(64)             not null,
    openid        varchar(128)            null,
    display_name  varchar(128)            not null,
    avatar_url    varchar(255) default '' not null,
    mobile        varchar(32)             null,
    email         varchar(128)            null,
    status        tinyint      default 1  not null,
    remark        varchar(255) default '' not null,
    created_at    datetime(3)             null,
    updated_at    datetime(3)             null,
    open_id       varchar(128)            null,
    last_login_at datetime(3)             null,
    constraint idx_users_email
        unique (email),
    constraint idx_users_mobile
        unique (mobile),
    constraint idx_users_open_id
        unique (open_id),
    constraint idx_users_username
        unique (username),
    constraint uk_users_openid
        unique (openid)
);

