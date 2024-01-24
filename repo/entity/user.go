package entity

import (
	"encoding/json"
	"time"
)

type User struct {
	ID        int       `gorm:"autoIncrement" json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	Username  string    `json:"username"` // 用户名【唯一】
	Name      string    `json:"name"`
	NamePy    string    `json:"namePy"`
	Password  Pwd       `json:"password"`  //口令加盐摘要Hex
	Salt      string    `json:"-"`         // 盐值Hex
	Avatar    string    `json:"avatar"`    // 头像文件名
	Openid    string    `json:"openid"`    // 开放ID 用于关联三方系统，可以是工号
	Phone     string    `json:"phone"`     // 手机号
	Email     string    `json:"email"`     // 邮箱
	Sn        string    `json:"sn"`        // 身份证
	NoteTags  string    `json:"noteTags"`  // 笔记标签 - 已弃用
	GroupTags string    `json:"groupTags"` // 用户组标签 - 已弃用
	IsDelete  int       `json:"isDelete"`  // 是否删除 0 - 未删除（默认值） 1 - 删除
}

func (c *User) MarshalJSON() ([]byte, error) {
	type Alias User
	return json.Marshal(&struct {
		*Alias
		CreatedAt DateTime `json:"createdAt"`
	}{
		(*Alias)(c),
		DateTime(c.CreatedAt),
	})
}
