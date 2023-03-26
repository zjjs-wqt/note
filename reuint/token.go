package reuint

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"note/reuint/jwt"
	"strings"
	"time"
)

// VerifyToken 验证token是否有效
// 若有效则返还解析后的有效荷载
// 若无效则返还错误，并说明错误原因
func VerifyToken(token string) (*jwt.Claims, error) {
	start := strings.IndexByte(token, '.')
	end := strings.LastIndexByte(token, '.')
	if start == -1 || end == -1 || start == end {
		return nil, errors.New("非法token")
	}

	// 获取 payload部分，并解析
	payload := token[start+1 : end]
	tokenInfo, err := base64.URLEncoding.DecodeString(payload)
	if err != nil {
		return nil, errors.New("非法token")

	}
	var access jwt.Claims
	err = json.Unmarshal(tokenInfo, &access)

	// 验证access_token 是否过期
	now := time.Now().UnixMilli()
	if now >= access.Exp {
		return nil, errors.New("token过期")
	}
	return &access, nil
}
