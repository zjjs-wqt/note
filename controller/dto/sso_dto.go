package dto

type AccessTokenDto struct {
	AccessToken string `json:"access_token"` // access_token 授权凭证
	ExpiresIn   int64  `json:"expires_in"`   // 过期时间 单位秒
}

type InfoDto struct {
	Code int  `json:"code"` // 状态码
	Data data `json:"data"` // 用户信息数据
}
type data struct {
	Openid   string `json:"openid"`
	Username string `json:"username"`
	Name     string `json:"name"`
	Phone    string `json:"phone"`
}
