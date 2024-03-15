package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"net/http"
	"note/appconf"
	"note/appconf/dir"
	"note/controller"
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

	// 启动用户同步服务
	go bootUserSyncSever(appcfg)

	// 启动Web服务器
	server := NewHttpServer(appcfg)
	if err = server.ListenAndServe(); err != nil {
		zap.L().Fatal("服务启动失败", zap.Error(err))
	}
}

// bootUserSyncSever 启动用户同步服务
func bootUserSyncSever(config *appconf.Application) {
	var r *gin.Engine
	if config.Debug {
		r = gin.Default()
	} else {
		r = gin.New()
	}

	route := r.Group("/api")
	controller.NewAyncController(route)
	zap.L().Info("系统启动", zap.Int("syncPort", config.SyncPort))
	server := http.Server{
		Addr:    fmt.Sprintf(":%d", config.SyncPort),
		Handler: r,
	}
	err := server.ListenAndServe()
	if err != nil {
		zap.L().Fatal("同步服务启动失败", zap.Error(err))
	}
}
