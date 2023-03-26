-- 创建管理员表
CREATE TABLE admins
(
    id         INTEGER PRIMARY KEY AUTO_INCREMENT, -- 自增主键
    created_at DATETIME,  -- 创建时间
    username   VARCHAR(128), -- 用户名
    password   VARCHAR(512), -- 口令加盐Hash结果 16进制字符串
    salt       VARCHAR(512), -- 盐值 16进制字符串
    role       TINYINT, -- 角色类型
    cert       TEXT-- 证书
);
-- 创建admin
INSERT INTO `admins`
VALUES (1, '2022-11-07 09:19:44', 'admin',
        'ba182cee746bc776a9bec5c73293dc730d517acf4a5f9c88213184739ef54693', 'a79e9fc93a41399c0e2a87971434655f', 0 ,  NULL);

INSERT INTO `admins`
VALUES (2, '2022-11-07 09:19:44', 'audit',
        '9f1a7062905d2a4e208f92ecb56f967569762bdbd09f7e020414736e79067893', '9946f8047b368c6219ec246e3f4638cb', 1 , NULL);


-- 创建用户表
CREATE TABLE users
(
    id          INTEGER PRIMARY KEY AUTO_INCREMENT, -- 自增主键
    created_at  DATETIME,                           -- 创建时间
    openid     VARCHAR(200),                        -- 开放ID 用于关联三方系统，可以是工号
    username    VARCHAR(128),                       -- 用户登录时输入的账户名称
    name        VARCHAR(256),                       -- 用户真实姓名
    name_py     VARCHAR(32),                        -- 姓名拼音缩写
    password    VARCHAR(512),                       -- 口令加盐Hash结果 16进制字符串
    salt        VARCHAR(512),                       -- 盐值 16进制字符串
    avatar      VARCHAR(512),                       -- 头像
    phone       VARCHAR(256),                       -- 手机号
    email       VARCHAR(256),                       -- 邮箱
    sn          VARCHAR(512),                       -- 身份证
    note_tags   VARCHAR(1024),-- 用户组标签列表 "多个标签使用“,”分隔。例如： “运维,常见问题”"
    group_tags  VARCHAR(1024),-- 用户组标签列表 "多个标签使用“,”分隔。例如： “运维,常见问题”"
    is_delete   TINYINT															-- 是否删除 0 - 未删除（默认值） 1 - 删除
);

-- 创建用户组表
CREATE TABLE user_groups
(
    id            INTEGER PRIMARY KEY AUTO_INCREMENT,-- 自增主键
    created_at    DATETIME,-- 创建时间
    name          VARCHAR(512) NOT NULL,-- 用户组名称
    name_py  			VARCHAR(32),-- 用户组名称拼音缩写
    description   VARCHAR(256),-- 描述
    tags        	VARCHAR(1024)-- 用户组标签列表 "多个标签使用“,”分隔。例如： “运维,常见问题”"
);

-- 创建用户组成员表
CREATE TABLE group_members
(
    id         INTEGER PRIMARY KEY AUTO_INCREMENT,-- 自增主键
    created_at DATETIME,-- 创建时间
    user_id    INTEGER,-- 用户ID
    belong     INTEGER,-- 所属项目
    role       TINYINT-- 用户类型 枚举值：0 - 用户组拥有者/管理者 ， 1 - 维护 ， 2 - 普通用户
);

-- 创建文档表
CREATE TABLE notes
(
    id         INTEGER PRIMARY KEY AUTO_INCREMENT,-- 自增主键
    created_at DATETIME,-- 创建时间
    updated_at DATETIME,-- 更新时间
    user_id 	 INTEGER,-- 所属用户ID
    title      VARCHAR(512) NOT NULL,-- 文档名
    title_py   VARCHAR(255), -- 文档名拼音缩写
    priority   INTEGER,-- 优先级 默认为0，越大优先级越高，用于文档排序，非特殊情况保持0即可。
    filename   VARCHAR(512),-- 文件名称
    tags       VARCHAR(1024),-- 用户组标签列表 "多个标签使用“,”分隔。例如： “运维,常见问题”"
    is_delete   TINYINT-- 是否删除 0 - 未删除（默认值） 1 - 删除
);

-- 创建笔记成员表
CREATE TABLE note_members
(
    id          INTEGER PRIMARY KEY AUTO_INCREMENT, -- 自增主键
    created_at  DATETIME,                           -- 创建时间
    user_id  		INTEGER,                            -- 用户ID
    note_id  		INTEGER,                            -- 笔记ID
    role       TINYINT,-- 用户类型 枚举值：0 - 笔记拥有者/管理者 ， 1 - 可查看 ， 2 - 可编辑
    remark 			VARCHAR(512), -- 备注
    group_id INTEGER                            -- 用户组ID
);


-- 创建日志表
CREATE TABLE logs
(
    id         INTEGER PRIMARY KEY AUTO_INCREMENT, -- 自增主键
    created_at DATETIME,                           -- 创建时间
    op_type    TINYINT,                            -- 操作者类型 类型如下包括：0 - 匿名，1 - 管理员，2 - 用户 若不知道用户或没有用户信息，则使用匿名。
    op_id      INTEGER,                            -- 操作者记录ID 0 表示匿名
    op_name    VARCHAR(512) NOT NULL,              -- 操作名称
    op_param   TEXT NULL                           -- 操作的关键参数 可选参数，例如删除用户时，删除的用户ID，复杂参数请使用JSON对象字符串，如{id: 1}
);


-- 创建版本号表
CREATE TABLE configs
(
    id        INTEGER PRIMARY KEY AUTO_INCREMENT, -- 自增主键
    item_name VARCHAR(256),
    content   VARCHAR(256)                        -- 版本号时间
);
-- 创建版本号记录
INSERT INTO configs(item_name, content)
VALUES ("db_version", "2023031401");
