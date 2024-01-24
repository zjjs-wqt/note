package dto

// FolderCreateDto 文件夹创建
type FolderCreateDto struct {
	Name     string `json:"name"`     // 文件夹名称
	ParentId int    `json:"parentId"` // 父文件夹ID
}

// FolderRenameDto 文件夹重命名
type FolderRenameDto struct {
	ID       int    `json:"id"`       // 文件夹ID
	Name     string `json:"name"`     // 文件夹名称
	ParentId int    `json:"parentId"` // 父文件夹ID
}

// FolderRemoveDto 文件夹移动
type FolderRemoveDto struct {
	ID       int `json:"id"`       // 文件夹ID
	ParentId int `json:"parentId"` // 移动到的文件夹ID
}
