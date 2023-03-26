package reuint

import (
	"errors"
	"fmt"
	"github.com/mozillazg/go-pinyin"
)

// PinyinConversion 将中文姓名转化为拼音首字母
// return：拼音首字母串，错误
func PinyinConversion(name string) (string, error) {
	//将中文转化为拼音的字符串切片
	strList := pinyin.LazyPinyin(name, pinyin.NewArgs())
	if strList == nil {
		return "", errors.New("拼音转化失败")
	}
	//生成姓名的拼音缩写。
	var res string
	for _, val := range strList {
		res = fmt.Sprintf("%s%s", res, val[0:1])
	}
	return res, nil
}
