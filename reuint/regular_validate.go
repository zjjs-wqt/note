package reuint

import "regexp"

var (
	// 手机号匹配规则
	phoneRegexp = regexp.MustCompile("^(13[0-9]|14[01456879]|15[0-35-9]|16[2567]|17[0-8]|18[0-9]|19[0-35-9])\\d{8}$")
	// 邮箱匹配规则
	emailRegexp = regexp.MustCompile("^\\w+([-+.]\\w+)*@\\w+([-.]\\w+)*\\.\\w+([-.]\\w+)*$")
)

// PhoneValidate 手机号格式校验
// 如 str： 13855555555 return: true
// return true
func PhoneValidate(str string) bool {
	return phoneRegexp.MatchString(str)
}

// EmailValidate 手机号格式校验
// 如 str： 123@qq.com return: true
// return true
func EmailValidate(str string) bool {
	return emailRegexp.MatchString(str)
}
