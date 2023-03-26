// 可重用单元 reuse uint

package reuint

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"github.com/emmansun/gmsm/sm3"
)

// GenPasswordSalt 通过明文和随机源生产口令+加盐
// return: 口令加盐摘要Hex, 盐值Hex, 错误
func GenPasswordSalt(password string) (string, string, error) {
	if password == "" {
		return "", "", fmt.Errorf("password 为空")
	}
	salt := make([]byte, 16)
	// 使用随机源生成盐值
	_, err := rand.Reader.Read(salt)
	if err != nil {
		return "", "", err
	}
	// 拼接 口令 + 盐值
	plaintext := append([]byte(password), salt...)
	hash := sm3.New()
	hash.Write(plaintext)
	pwdHash := hash.Sum(nil)
	return hex.EncodeToString(pwdHash), hex.EncodeToString(salt), nil
}

// VerifyPasswordSalt 验证口令是否有效
// password: 待验证的口令
// pwdSaltHex: 口令加盐Hash值
// saltHex: 盐值
func VerifyPasswordSalt(password, pwdSaltHex, saltHex string) bool {
	if password == "" || pwdSaltHex == "" || saltHex == "" {
		return false
	}
	// 期望的处理结果
	exp, err := hex.DecodeString(pwdSaltHex)
	if err != nil {
		// 无法解码
		return false
	}
	salt, err := hex.DecodeString(saltHex)
	if err != nil {
		return false
	}

	plaintext := append([]byte(password), salt...)
	hash := sm3.New()
	hash.Write(plaintext)
	actual := hash.Sum(nil)
	return bytes.Equal(exp, actual)
}
