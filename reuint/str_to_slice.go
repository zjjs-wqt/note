package reuint

import (
	"strconv"
	"strings"
)

// StrToIntSlice 将以 , 为分隔符号的str转化为空格分隔int切片
// 如 str： 26,27 return:[26 27]
// return int切片
func StrToIntSlice(str string) []int {
	strArray := strings.Split(str, ",")
	var intArray []int
	for _, val := range strArray {
		temp, _ := strconv.Atoi(val)
		intArray = append(intArray, temp)
	}
	return intArray
}
