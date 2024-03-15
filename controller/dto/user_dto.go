package dto

import "note/repo/entity"

// UserCreateDto create接口接收前端数据
type UserCreateDto struct {
	Name     string     `json:"name"`     // 姓名
	Password entity.Pwd `json:"password"` // 口令加盐摘要Hex
	Openid   string     `json:"openid"`   // 开放ID 用于关联三方系统，可以是工号
	Phone    string     `json:"phone"`    // 手机号
	Email    string     `json:"email"`    // 邮箱
}

// PasswordDto 修改口令接口接收前端的数据
type PasswordDto struct {
	Username string     `json:"username"` // 用户名
	OldPwd   entity.Pwd `json:"oldPwd"`   // 旧口令
	NewPwd   entity.Pwd `json:"newPwd"`   // 新口令
}

// NameListDto 接口将以下数据返回给前端
type NameListDto struct {
	ID   int    `json:"id"`   // 用户ID
	Name string `json:"name"` // 用户姓名
}

// Transform 将实体数据赋值给dto，返回前端
func (nameListDto *NameListDto) Transform(usr *entity.User) *NameListDto {
	nameListDto.ID = usr.ID
	nameListDto.Name = usr.Name
	return nameListDto
}

// UserListDto 接口将以下数据返回给前端
type UserListDto struct {
	ID   int    `json:"id"`   // 用户ID
	Name string `json:"name"` // 用户姓名
	Role int    `json:"role"`
}

// UserInfoDto 接口将以下数据返回给前端
type UserInfoDto struct {
	ID       int    `json:"id"`       // 用户ID
	Username string `json:"username"` // 用户名
	Name     string `json:"name"`     // 用户姓名
	Openid   string `json:"openid"`   // 工号
	Phone    string `json:"phone"`    // 手机号
	Email    string `json:"email"`    // 邮箱
	Sn       string `json:"sn"`       // 身份证
}

// AyncUserDto 用户同步DTO
type AyncUserDto struct {
	JobNumber int    `json:"jobNumber"` // 工号
	State     int    `json:"state"`     // 用户状态 6 - 离职
	Name      string `json:"name"`      // 姓名
	Phone     string `json:"phone"`     // 手机号
	Email     string `json:"email"`     // 邮箱
}
