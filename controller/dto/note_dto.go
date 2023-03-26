package dto

import (
	"encoding/json"
	"note/repo/entity"
	"time"
)

type NoteCreateDto struct {
	Title    string `json:"title"`
	Priority int    `json:"priority"`
	Tags     string `json:"tags"`
}

type NoteIDDto struct {
	ID   int      `gorm:"autoIncrement" json:"id"`
	Tags []string `json:"tags"`
}

type NoteInfoDto struct {
	ID        int       `gorm:"autoIncrement" json:"id"`
	UpdatedAt time.Time `json:"updatedAt"`
	Title     string    `json:"title"`
	Remark    string    `json:"remark"`
	UserName  string    `json:"userName"`
	Tags      string    `json:"tags"`     // 笔记标签列表 多个标签使用“,”分隔。  例如： “运维,常见问题”
	Role      int       `json:"role"`     // 用户权限
	IsDelete  int       `json:"isDelete"` // 是否删除
}

func (c *NoteInfoDto) MarshalJSON() ([]byte, error) {
	type Alias NoteInfoDto
	return json.Marshal(&struct {
		*Alias
		UpdatedAt entity.DateTime `json:"updatedAt"`
	}{
		(*Alias)(c),
		entity.DateTime(c.UpdatedAt),
	})
}

type NoteListDto struct {
	ID        int       `gorm:"autoIncrement" json:"id"`
	UpdatedAt time.Time `json:"updatedAt"`
	Title     string    `json:"title"`
	Remark    string    `json:"remark"`
	Username  string    `json:"username"`
	Tags      string    `json:"tags"` // 笔记标签列表 多个标签使用“,”分隔。  例如： “运维,常见问题”
	Role      int       `json:"role"` // 用户权限
}

func (c *NoteListDto) MarshalJSON() ([]byte, error) {
	type Alias NoteListDto
	return json.Marshal(&struct {
		*Alias
		UpdatedAt entity.DateTime `json:"updatedAt"`
	}{
		(*Alias)(c),
		entity.DateTime(c.UpdatedAt),
	})
}

// NoteUriDto 笔记资源dto
type NoteUriDto struct {
	DocType string `json:"docType"` // 资源类型:image,file
	Uri     string `json:"uri"`     // 资源访问地址，资源访问路径为文档资源的下载接口，格式为: "/api/doc /assert?stageId=xxxx&file=202210131620160001.png"
}
