package jwt

// Claims JWT 荷载信息
type Claims struct {
	Type string `json:"type"` // 用户类型: user - 用户、 admin - 管理员
	Sub  int    `json:"sub"`  // 用户ID或管理员ID
	Exp  int64  `json:"exp"`  // 过期时间，Unix 毫秒数
	Role int    `json:"role"` // 角色
}
