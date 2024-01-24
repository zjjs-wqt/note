package middle

import (
	"github.com/gin-gonic/gin"
	"strings"
)

const (
	FlagAnonymous = "Anonymous" // 匿名标志
	FlagClaims    = "Claims"    // 用户信息
)

// Anonymous 匿名访问接口
func Anonymous(ctx *gin.Context) {
	dest := ctx.Request.URL.Path
	// 静态资源
	if strings.HasPrefix(dest, "/ui") || dest == "" || dest == "/" {
		ctx.Set(FlagAnonymous, true)
		return
	}
	switch dest {
	case "/api/login", "/api/system/version", "/api/random", "/api/entityAuth", "/api/certBinding", "/api/redirect", "/api/sync":
		ctx.Set(FlagAnonymous, true)
		return
	}
}
