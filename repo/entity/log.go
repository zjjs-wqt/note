package entity

import (
	"time"
)

// Log 操作日志
type Log struct {
	ID        int       `gorm:"autoIncrement" json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	OpType    int       `json:"opType"`  // 操作者类型 类型如下包括：0 - 匿名 1 - 管理员 2 - 用户 若不知道用户或没有用户信息，则使用匿名。
	OpId      int       `json:"opId"`    // 操作者记录ID
	OpName    string    `json:"opName"`  // 操作名称
	OpParam   string    `json:"opParam"` // 操作的关键参数 可选参数，例如删除用户时，删除的用户ID，复杂参数请使用JSON对象字符串，如{id: 1}
}
