package controller

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"net/http"
	"runtime"
)

// ErrSys 内部错误
func ErrSys(c *gin.Context, err error) {
	// 打印日志文件
	zap.L().Error("系统内部错误",
		zap.String("errTyp", "Inn"),
		zap.Error(err),
		zap.String("caller", caller()))
	c.Writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
	c.AbortWithStatus(http.StatusInternalServerError)
	_, _ = c.Writer.WriteString("系统内部错误")
}

// ErrIllegal 参数错误
func ErrIllegal(c *gin.Context, hit string) {
	// 打印日志
	zap.L().Info("参数错误",
		zap.String("errTyp", "Illegal"),
		zap.String("desp", hit),
		zap.String("caller", caller()))
	// 返回状态码
	c.Writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
	c.AbortWithStatus(http.StatusBadRequest)
	_, _ = c.Writer.WriteString(hit)
}

// ErrIllegalE 参数错误
func ErrIllegalE(c *gin.Context, err error) {
	// 打印日志
	zap.L().Info("参数错误",
		zap.String("errTyp", "Illegal"),
		zap.Error(err),
		zap.String("caller", caller()))
	// 返回状态码
	c.Writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
	c.AbortWithStatus(http.StatusBadRequest)
	_, _ = c.Writer.WriteString(err.Error())
}

// ErrAuth 认证错误,不提供任何参数和响应体，不记录日志
func ErrAuth(c *gin.Context) {
	c.AbortWithStatus(http.StatusUnauthorized)
}

// ErrForbidden 禁止访问
func ErrForbidden(c *gin.Context, msg string) {
	c.AbortWithStatus(http.StatusForbidden)
	_, _ = c.Writer.WriteString(msg)
}

// ErrNormal 普通错误，非系统错误由于运行时内部错误错误导致的错误，或不知道该如何处理的错误
func ErrNormal(c *gin.Context, hit string, err error) {
	// 打印日志文件
	zap.L().Warn("异常",
		zap.String("errTyp", "Normal"),
		zap.String("desp", hit),
		zap.Error(err),
		zap.String("caller", caller()))

	c.Writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
	c.AbortWithStatus(http.StatusNotAcceptable)
	_, _ = c.Writer.WriteString(hit)
}

// 打印调用信息
func caller() string {
	pc, file, line, _ := runtime.Caller(2)
	return fmt.Sprintf("%s %s %d", runtime.FuncForPC(pc).Name(), file, line)
}
