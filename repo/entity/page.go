package entity

// Page 页面对象
type Page struct {
	Records interface{} `json:"records"` // 分页记录
	Total   int64       `json:"total"`   // 记录总数
	Size    int         `json:"size"`    // 每页显示条数
	Current int         `json:"current"` // 当前页
	Pages   int         `json:"pages"`   // 总页数
}
