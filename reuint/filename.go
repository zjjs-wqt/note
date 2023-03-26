package reuint

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

var rdSrc = rand.New(rand.NewSource(time.Now().UnixNano()))

// GenTimeFileName 生成随机的文件名称，格式为： YYYYMMDDHHmmssRRRR  R 表示随机数
func GenTimeFileName(name string) string {
	//生成文件名 - now = `YYYYMMDDHHmmss`   rand = `4位随机数`
	now := time.Now().Format("20060102150405")
	r := fmt.Sprintf("%04v", rdSrc.Int31n(10000))

	// 获取文件格式
	fileFormat := strings.Split(name, ".")
	filename := ""
	if len(fileFormat) == 1 {
		filename = fmt.Sprintf("%s%s", now, r)
	} else {
		filename = fmt.Sprintf("%s%s.%s", now, r, fileFormat[len(fileFormat)-1])
	}
	return filename
}
