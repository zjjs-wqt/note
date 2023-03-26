package middle

import (
	"crypto/rand"
	"note/reuint/jwt"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"net/http"
	"time"
)

// TokenManager Token管理器
type TokenManager struct {
	key    []byte // 当前HMAC密钥
	oldKey []byte // 过去HMAC密钥
	ticker *time.Ticker
}

// NewTokenFilter 新建token过滤器
func NewTokenFilter() *TokenManager {
	res := &TokenManager{
		key:    make([]byte, 32),
		oldKey: make([]byte, 32),
		ticker: time.NewTicker(time.Hour * 12),
	}
	_, _ = rand.Reader.Read(res.key)
	// 12小时更新一次密钥
	go func() {
		for _ = range res.ticker.C {
			zap.L().Info("JWT密钥更新")
			copy(res.oldKey, res.key)
			_, _ = rand.Reader.Read(res.key)

		}
	}()
	return res
}

// Filter token校验拦截器
func (t *TokenManager) Filter(ctx *gin.Context) {
	// 忽略匿名访问接口
	if _, exists := ctx.Get(FlagAnonymous); exists {
		return
	}

	// 从cookies中获取token
	token, _ := ctx.Cookie("token")
	if token == "" {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	// 验证Token有效性
	claims, err := jwt.Verify(t.key, token)
	if err != nil {
		var err2 error
		// token 密钥发生更新，尝试使用过去的密钥验证
		claims, err2 = jwt.Verify(t.oldKey, token)
		if err2 != nil {
			// 清除头里失效的token
			ctx.SetCookie("token", "", -1, "", "", false, true)
			ctx.AbortWithStatus(http.StatusUnauthorized)
			_, _ = ctx.Writer.WriteString(err.Error())
			return
		}
	}
	ctx.Set(FlagClaims, claims)
	return
}

// GenToken 生成新的token
func (t *TokenManager) GenToken(claims *jwt.Claims) string {
	return jwt.New(t.key, claims)
}
