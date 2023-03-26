package dto

import "note/repo/entity"

// UserDto 用户接口将数据返回给前端
type UserDto struct {
	ID        int             `json:"id"`
	CreatedAt entity.DateTime `json:"createdAt"`
	Username  string          `json:"username"` // 用户名
	Name      string          `json:"name"`     // 用户姓名
}

// Transform 将实体数据赋值给dto返回给前端
func (userDto *UserDto) Transform(usr *entity.User) *UserDto {
	userDto.ID = usr.ID
	userDto.CreatedAt = entity.DateTime(usr.CreatedAt)
	userDto.Username = usr.Username
	userDto.Name = usr.Name
	return userDto
}

// UserCreateDto create接口接收前端数据
type UserCreateDto struct {
	ID       int        `json:"id"`
	Username string     `json:"username"`
	Name     string     `json:"name"`
	NamePy   string     `json:"namePy"`
	Password entity.Pwd `json:"password"` //口令加盐摘要Hex
	Salt     string     `json:"-"`        // 盐值Hex
	Openid   string     `json:"openid"`   // 开放ID 用于关联三方系统，可以是工号
	Phone    string     `json:"phone"`    // 手机号
	Email    string     `json:"email"`    // 邮箱
	IsDelete int        `json:"isDelete"` // 是否删除 0 - 未删除（默认值） 1 - 删除
}

// NewEntity 将接收到的数据赋值给实体
func (u *UserCreateDto) NewEntity(usr *entity.User) *entity.User {
	usr.ID = u.ID
	usr.Username = u.Username
	usr.Name = u.Name
	usr.NamePy = u.NamePy
	usr.Password = u.Password
	usr.Salt = u.Salt
	usr.Openid = u.Openid
	usr.Phone = u.Phone
	usr.Email = u.Email
	usr.IsDelete = u.IsDelete
	return usr
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
