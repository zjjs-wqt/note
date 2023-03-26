package repo

import (
	"fmt"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"note/appconf"
	"strings"
	"time"
)

// DBDao 数据库连接实例
var DBDao *gorm.DB

var (
	NoteMemberRepo *NoteMemberRepository
	UserRepo       *UserRepository
	UserGroupRepo  *UserGroupRepository
)

// Init 初始化数据库信息
func Init(config *appconf.Application) error {
	var err error
	zap.L().Info("连接数据库",
		zap.String("dsn", config.Database.DSN),
		zap.String("type", config.Database.Type))
	config.Database.Type = strings.ToLower(config.Database.Type)

	level := logger.Warn
	if config.Debug {
		level = logger.Info
	}
	output := logger.New(log.Default(), logger.Config{
		SlowThreshold:             200 * time.Millisecond,
		LogLevel:                  level,
		IgnoreRecordNotFoundError: false,
		Colorful:                  false,
	})

	var dialector gorm.Dialector
	switch config.Database.Type {
	case "mysql", "mariadb":
		dialector = mysql.Open(config.Database.DSN)
	default:
		err = fmt.Errorf("未知的数据库类型: %s", config.Database.Type)
	}
	if err != nil {
		return fmt.Errorf("数据库配置信息错误，%s", err.Error())
	}

	DBDao, err = gorm.Open(dialector, &gorm.Config{Logger: output})
	if err != nil {
		return err
	}

	// 服务注册
	NoteMemberRepo = NewNoteMemberRepository()
	UserRepo = NewUserRepository()
	UserGroupRepo = NewUserGroupRepository()
	return nil
}
