package reuint

import (
	"fmt"
	"go.uber.org/zap"
	"os"
	"path/filepath"
	"regexp"
	"testing"
)

func TestDeleteUnreferencedFiles(t *testing.T) {
	str := "![cc.png](/api/project/stage/assert?stageId=3\\\\[assert.crt](/api/project/stage/assert?stageId=3\\\\&file=202210311427003271.crt)[Schedule.vue](/api/project/stage/assert?stageId=3\\\\&file=202210311427073020.vue)\n"
	reg := regexp.MustCompile("\\&file=.*?\\)")
	results := reg.FindAllString(str, -1)
	fmt.Println(results)

	base, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	filePath := filepath.Join(base, "stage", "3")
	err := DeleteUnreferencedFiles(str, filePath)
	if err != nil {
		zap.L().Debug("err!")
	}
}
