package dto

// CertItemDto 搜索文件返回值
type CertItemDto struct {
	Name      string `json:"name"`      // 根证书名称
	CreatedAt string `json:"createdAt"` // 根证书上传时间格式 YYYY-MM-DD HH:mm:ss
}
