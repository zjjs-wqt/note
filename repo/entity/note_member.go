package entity

import (
	"encoding/json"
	"time"
)

type NoteMember struct {
	ID        int       `gorm:"autoIncrement" json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	UserId    int       `json:"userId"`    // 用户ID
	NoteId    int       `json:"noteId"`    // 笔记ID
	Role      int       `json:"role"`      // 用户权限 0 - 笔记拥有者/管理者 1 - 可查看 2 - 可编辑
	Remark    string    `json:"remark"`    // 备注
	NoteGroup string    `json:"noteGroup"` // 笔记分组（兼容，最新版已采用文件夹Id）
	GroupId   int       `json:"groupId"`   // 用户组ID
	FolderId  int       `json:"folderId"`  // 文件夹Id
}

func (c *NoteMember) MarshalJSON() ([]byte, error) {
	type Alias NoteMember
	return json.Marshal(&struct {
		*Alias
		CreatedAt DateTime `json:"createdAt"`
	}{
		(*Alias)(c),
		DateTime(c.CreatedAt),
	})
}
