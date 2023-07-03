package main

import (
	"go.uber.org/zap"
	"note/appconf"
	"note/appconf/dir"
	"note/logg"
	"note/logg/applog"
	"note/noteDaemon"
	"note/repo"
)

func main() {
	// 初始化各级目录
	dir.Init()
	// 加载配置文件配置
	appcfg := appconf.Load()
	// 初始化日志
	logg.InitConsole(appcfg.Debug)
	// 数据库初始化
	err := repo.Init(appcfg)
	if err != nil {
		zap.L().Fatal("持久层初始化失败", zap.Error(err))
	}
	// 初始化操作日志模块
	applog.InitLogger(appcfg)
	// 初始化笔记定时清除模块
	noteDaemon.InitNote(appcfg)
	// 启动Web服务器
	server := NewHttpServer(appcfg)
	if err = server.ListenAndServe(); err != nil {
		zap.L().Fatal("服务启动失败", zap.Error(err))
	}
}
