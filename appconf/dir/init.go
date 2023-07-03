package dir

import (
	"log"
	"os"
	"path/filepath"
)

var (
	base        string // 程序运行目录
	LogDir      string // 日志存储目录
	NoteDir     string // 笔记文件存储目录
	AvatarDir   string // 头像存储目录
	UiDir       string // 前端文件存储目录
	RootCertDir string // 根证书管理
)

func Init() {
	base, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	LogDir = filepath.Join(base, "logs")
	UiDir = filepath.Join(base, "ui")
	AvatarDir = filepath.Join(base, "avatar")
	NoteDir = filepath.Join(base, "notes")
	RootCertDir = filepath.Join(base, "rootCerts")

	_ = os.MkdirAll(LogDir, os.ModePerm)
	_ = os.MkdirAll(UiDir, os.ModePerm)
	_ = os.MkdirAll(AvatarDir, os.ModePerm)
	_ = os.MkdirAll(NoteDir, os.ModePerm)
	_ = os.MkdirAll(RootCertDir, os.ModePerm)

	log.Println("程序运行目录:", base)
	log.Println("日志存储目录:", LogDir)
	log.Println("前端文件存储目录:", UiDir)
	log.Println("头像存储目录:", AvatarDir)
	log.Println("笔记文件存储目录:", NoteDir)
	log.Println("根证书目录:", RootCertDir)
}
