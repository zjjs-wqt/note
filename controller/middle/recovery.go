package middle

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"net/http"
	"runtime"
)

// Recovery 意料外panic或故障的恢复
func Recovery() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, e interface{}) {
		// 函数调用者
		pc, file, line, _ := runtime.Caller(2)
		caller := fmt.Sprintf("%s %s %d", runtime.FuncForPC(pc).Name(), file, line)

		var err error
		if ee, ok := err.(error); ok {
			err = ee
		} else {
			err = fmt.Errorf("未知类型错误发生: %+v", e)
		}

		// 打印日志文件
		zap.L().Error("系统内部错误",
			zap.String("errTyp", "Inn"),
			zap.Error(err),
			zap.String("caller", caller))
		c.Writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
		c.AbortWithStatus(http.StatusInternalServerError)
		_, _ = c.Writer.WriteString("系统内部错误")
	})
}
