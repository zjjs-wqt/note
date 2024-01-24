package entity

import (
	"encoding/json"
	"time"
)

// Folder 文件夹
type Folder struct {
	ID        int       `gorm:"autoIncrement" json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	UserId    int       `json:"userId"`   // 用户ID
	Name      string    `json:"name"`     // 文件夹名称
	ParentId  int       `json:"parentId"` // 父文件夹ID
}

func (c *Folder) MarshalJSON() ([]byte, error) {
	type Alias Folder
	return json.Marshal(&struct {
		*Alias
		CreatedAt DateTime `json:"createdAt"`
	}{
		(*Alias)(c),
		DateTime(c.CreatedAt),
	})
}
