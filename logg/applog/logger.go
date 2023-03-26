package applog

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"note/appconf"
	"note/controller/middle"
	"note/repo"
	"note/repo/entity"
	"note/reuint/jwt"
	"time"
)

var _globalL *Logger

// Logger 日志模块
type Logger struct {
	buff        chan *entity.Log // 日志缓冲区
	maxKeepDays int              // 日志最大存储时间（单位：天），注意若该值小于等于0则表示不删除。
}

// Log 写入日志
func (l *Logger) Log(record *entity.Log) {
	if record == nil {
		return
	}
	l.buff <- record
}

// daemon 日志精灵用于将缓存中的日志持久化
func (l *Logger) daemon() {
	var err error
	zap.L().Info("日志持久化存储精灵 [启动]")
	for item := range l.buff {
		record := item
		err = repo.DBDao.Create(item).Error

		if err != nil {
			zap.L().Warn("日志写入失败", zap.Any("record", record), zap.Error(err))
		}
	}
}

// 超时日志清理精灵
// 注意该函数不应抛出任何错误，若有错误请手动恢复并打印，继续下一个循环。
func (l *Logger) timeoutDeleteDaemon() {
	if _globalL.maxKeepDays > 0 {
		for {
			now := time.Now().AddDate(0, 0, -_globalL.maxKeepDays).Format("2006-01-02 15:04:05")
			repo.DBDao.Where("created_at < ?", now).Delete(&entity.Log{})
			time.Sleep(24 * time.Hour)
		}
	}
}

// InitLogger 初始化日志记录器
func InitLogger(cfg *appconf.Application) {
	if _globalL != nil {
		return
	}
	_globalL = &Logger{
		buff:        make(chan *entity.Log, 32),
		maxKeepDays: cfg.LogKeepMaxDays,
	}
	// 日志写入精灵
	go _globalL.daemon()
	// 日志超时删除精灵
	go _globalL.timeoutDeleteDaemon()
}

// L 记录日志
func L(ctx *gin.Context, name string, param interface{}) {
	var record entity.Log

	// 获取用户信息
	claimsValue, _ := ctx.Get(middle.FlagClaims)
	claims := claimsValue.(*jwt.Claims)
	if claims.Type == "user" {
		record.OpType = 2
	} else if claims.Type == "admin" {
		record.OpType = 1
	} else {
		record.OpType = 0
	}
	record.OpId = claims.Sub
	record.OpName = name
	if param != nil {
		marshal, _ := json.Marshal(param)
		record.OpParam = string(marshal)
	}

	if _globalL == nil {
		zap.L().Info("日志", zap.Any("record", record))
		return
	}
	_globalL.buff <- &record
}

func Init(c entity.Log, claimsType string, id int, name string, param interface{}) *entity.Log {

	if claimsType == "user" {
		c.OpType = 2
	} else if claimsType == "admin" {
		c.OpType = 1
	} else {
		c.OpType = 0
	}
	c.OpId = id
	if param != nil {
		marshal, _ := json.Marshal(param)
		c.OpParam = string(marshal)
	}
	c.OpName = name

	return &c
}
