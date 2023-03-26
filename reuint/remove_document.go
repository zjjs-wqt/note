package reuint

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
)

// DeleteUnreferencedFiles 删除文件夹中未被引用的文件
// - content 文档内容 - docType 文档类型 - id 文档ID - exp 额外参数
func DeleteUnreferencedFiles(content string, filePath string) error {

	reg := regexp.MustCompile("\\&file=.*?\\)")
	results := reg.FindAllString(content, -1)

	files, _ := os.ReadDir(filePath)
	for _, f := range files {
		name := fmt.Sprintf("&file=%s)", f.Name())
		removeFlag := true
		for _, result := range results {
			if result == name {
				removeFlag = false
			}
		}
		if removeFlag {
			removePath := filepath.Join(filePath, f.Name())
			err := os.Remove(removePath)
			if err != nil {
				return err
			}
		}

	}
	return nil
}
