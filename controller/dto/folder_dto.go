package dto

type FolderCreateDto struct {
	ID   int    `json:"id"`   // 用户ID
	Name string `json:"name"` //文件夹名称
}

type FolderIDDto struct {
	ID   int      `gorm:"autoIncrement" json:"id"`
	Tags []string `json:"tags"`
}
