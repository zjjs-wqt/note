package reuint

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Zip 压缩并输出到流
// out: 输出流，应由调用者负责关闭该流。
// tbz: 带压缩文件列表，to be zipped。
func Zip(out io.Writer, tbz ...string) error {
	// 打开：zip文件
	archive := zip.NewWriter(out)
	defer archive.Close()

	for _, src := range tbz {
		src, _ = filepath.Abs(src)
		// 遍历路径信息
		parent := ""
		err := filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			if path == src {
				off := strings.LastIndex(path, info.Name())
				parent = path[:off]
			}

			// 获取：文件头信息
			header, _ := zip.FileInfoHeader(info)
			header.Name = strings.TrimPrefix(path, parent)
			if info.IsDir() {
				// 如果是目录 压缩文件路径需要包含后缀 "/"
				header.Name = fmt.Sprintf("%s%c", header.Name, filepath.Separator)
			} else {
				// 设置：zip的文件压缩算法
				header.Method = zip.Deflate
			}
			fmt.Println(">> ", header.Name)
			// 压缩文件头
			writer, _ := archive.CreateHeader(header)
			if !info.IsDir() {
				// 写入文件
				file, _ := os.Open(path)
				_, err := io.Copy(writer, file)
				_ = file.Close()
				if err != nil {
					return err
				}
			}
			return nil
		})
		if err != nil {
			return err
		}
	}
	return nil
}
