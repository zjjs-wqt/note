-- 创建文件夹表
CREATE TABLE folders
(
    id         INTEGER PRIMARY KEY AUTO_INCREMENT,-- 自增主键
    created_at  DATETIME,-- 创建时间
    user_id 	 INTEGER,-- 所属用户ID
    name   VARCHAR(256),-- 文件夹名称
    parent_id   INTEGER-- 父文件夹ID，若为0，则为根文件夹。
);

-- 笔记成员表增加文件夹ID字段
ALTER TABLE note_members
    ADD folder_id INTEGER DEFAULT 0;

-- 更新版本号记录
UPDATE configs SET content = 20240124 WHERE item_name = "db_version";