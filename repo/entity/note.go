package entity

import "time"

// Note 笔记
type Note struct {
	ID        int       `gorm:"autoIncrement" json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	UserId    int       `json:"userId"`   // 所属用户ID
	Title     string    `json:"title"`    // 文档名
	TitlePy   string    `json:"titlePy"`  // 文档名称缩写
	Priority  int       `json:"priority"` // 优先级 默认为0，越大优先级越高，用于文档排序，非特殊情况保持0即可。
	Filename  string    `json:"filename"` // 文件名称
	Tags      string    `json:"tags"`     // 笔记标签列表 多个标签使用“,”分隔。  例如： “运维,常见问题”
	IsDelete  int       `json:"isDelete"` // 是否删除 0 - 未删除（默认值） 1 - 删除
}
