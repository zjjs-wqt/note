package jwt

import (
	"bytes"
	"crypto/hmac"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/emmansun/gmsm/sm3"
	"strings"
	"time"
)

// JWT头部 Base64-URL-Encoded: {"alg":"HMAC-SM3","typ":"JWT"}
const header = "eyJhbGciOiJITUFDLVNNMyIsInR5cCI6IkpXVCJ9"

// New 创建Token
func New(key []byte, claims *Claims) string {
	if claims == nil || len(key) == 0 {
		return ""
	}

	bodyBin, _ := json.Marshal(claims)
	payload := base64.URLEncoding.EncodeToString(bodyBin)

	builder := strings.Builder{}
	builder.WriteString(header)
	builder.WriteByte('.')
	builder.WriteString(payload)

	// 计算JWT的签名值部分
	// HMAC-SM3(base64UrlEncode(header) + "." + base64UrlEncode(payload))
	h := hmac.New(sm3.New, key)
	h.Write([]byte(builder.String()))
	sig := h.Sum(nil)

	builder.WriteByte('.')
	builder.WriteString(base64.URLEncoding.EncodeToString(sig))
	return builder.String()
}

// Verify 验证token是否有效
// 若有效则返还解析后的有效荷载
// 若无效则返还错误，并说明错误原因
func Verify(key []byte, token string) (*Claims, error) {
	start := strings.IndexByte(token, '.')
	end := strings.LastIndexByte(token, '.')
	if start == -1 || end == -1 || start == end {
		return nil, fmt.Errorf("非法token")
	}

	// 获取 payload部分，并解析
	payload := token[start+1 : end]
	claims, err := base64.URLEncoding.DecodeString(payload)
	if err != nil {
		return nil, fmt.Errorf("非法token")
	}
	var c Claims
	err = json.Unmarshal(claims, &c)
	if err != nil {
		return nil, fmt.Errorf("非法token")
	}
	now := time.Now().UnixMilli()
	if now >= c.Exp {
		return &c, fmt.Errorf("token过期")
	}

	// 获取 signature 部分
	plaintext := token[0:end]
	actual, err := base64.URLEncoding.DecodeString(token[end+1:])
	if err != nil {
		return nil, fmt.Errorf("非法token")
	}

	// 计算比较 HMAC是否正确
	h := hmac.New(sm3.New, key)
	h.Write([]byte(plaintext))
	exp := h.Sum(nil)
	if !bytes.Equal(exp, actual) {
		return &c, fmt.Errorf("无效token")
	}
	return &c, nil
}
