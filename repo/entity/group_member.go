package entity

import (
	"encoding/json"
	"time"
)

type GroupMember struct {
	ID        int       `gorm:"autoIncrement" json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	UserId    int       `json:"userId"` // 用户ID
	Belong    int       `json:"belong"` // 所属用户组ID
	Role      int       `json:"role"`   //  用户权限： 0 - 用户组拥有者/管理者 1 - 普通用户 2 - 维护
}

func (c *GroupMember) MarshalJSON() ([]byte, error) {
	type Alias GroupMember
	return json.Marshal(&struct {
		*Alias
		CreatedAt DateTime `json:"createdAt"`
	}{
		(*Alias)(c),
		DateTime(c.CreatedAt),
	})
}
