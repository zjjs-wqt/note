package logg

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"log"
	"note/appconf/dir"
	"os"
	"path/filepath"
	"time"
)

var (
	LogOutput zapcore.WriteSyncer
	ZapLog    *zap.Logger
)

// InitConsole 初始化控制台日志，同时向文件和控制写入日志
// 文件日志每天自动切分，保存180天，文件日志保存于工作目录下的 ./logs/ 目录
func InitConsole(debug bool) *zap.Logger {
	filename := filepath.Join(dir.LogDir, "note.log")
	// 创建文件目录
	spliceFile := &lumberjack.Logger{
		Filename:  filename,
		MaxSize:   10,   // MB
		MaxAge:    180,  // Day
		LocalTime: true, // 本地时间
		Compress:  true, // 启用压缩
	}

	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "name",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.TimeEncoderOfLayout(time.RFC3339),
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
		EncodeName:     zapcore.FullNameEncoder,
	}
	fileEncoder := zapcore.NewConsoleEncoder(encoderConfig)
	// 同时向控制台和文件写入日志
	zapLevel := zapcore.InfoLevel
	ginLevel := gin.ReleaseMode

	if debug {
		zapLevel = zapcore.DebugLevel
		ginLevel = gin.DebugMode
	}

	syncer := zapcore.NewMultiWriteSyncer(zapcore.AddSync(spliceFile), zapcore.AddSync(os.Stdout))
	// 同时向控制台和文件写入日志
	ZapLog = zap.New(zapcore.NewCore(fileEncoder, syncer, zapLevel), zap.AddCaller())

	zap.ReplaceGlobals(ZapLog)
	gin.SetMode(ginLevel)

	LogOutput = syncer
	// gin 日志
	gin.DefaultWriter = syncer
	// 日志
	log.SetOutput(syncer)
	return ZapLog
}
