package dto

// UserGroupCreateDto 用户组创建DTO
type UserGroupCreateDto struct {
	Name        string `json:"name"` // 用户组名称
	Description string `json:"description"`
}

type UserGroupMemberAddDto struct {
	GroupId int `json:"groupId"` // 用户组ID
	UserId  int `json:"userId"`  // 用户ID
	Role    int `json:"role"`    // 用户权限
}

// MemberAllDTO all接口将数据返回前端
type MemberAllDTO struct {
	ID     int    `json:"id"`     // 记录ID
	UserId int    `json:"userId"` // 用户ID
	Role   int    `json:"role"`   // 角色 角色类型包括：0 - 访客，1 - 测试，2 - 开发，3 - 维护，4-负责人
	Name   string `json:"name"`
}

// MemberListDTO list接口将数据返回前端
type MemberListDTO struct {
	ID     int    `json:"id"`     // 记录ID
	UserId int    `json:"userId"` // 用户ID
	Belong int    `json:"belong"` // 所属用户组
	Role   int    `json:"role"`   // 角色 角色类型包括：0 - 访客，1 - 测试，2 - 开发，3 - 维护，4-负责人
	Name   string `json:"name"`   // 用户组名称
}

// GroupListDto 接口将以下数据返回给前端
type GroupListDto struct {
	ID   int    `json:"id"`   // 组ID
	Name string `json:"name"` // 组名
}
