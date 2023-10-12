package middle

import (
	"github.com/gin-gonic/gin"
	"github.com/patrickmn/go-cache"
	"note/reuint/jwt"
	"time"
)

type EditLock struct {
	editCache *cache.Cache
}

const (
	NoLock = 0 // 无人持有该锁
)

type LockInfo struct {
	UserId int   // 持有该锁的用户ID
	Exp    int64 // 用户过期时间 - 不同处登录过期时间不同
}

func NewEditLock() *EditLock {
	res := &EditLock{}
	res.editCache = cache.New(5*time.Hour, 7*time.Hour)
	return res
}

// Lock 加锁
func (c *EditLock) Lock(ctx *gin.Context, noteId string, userId int) {
	// 获取用户信息
	claimsValue, _ := ctx.Get(FlagClaims)
	claims := claimsValue.(*jwt.Claims)
	info := LockInfo{userId, claims.Exp}
	c.editCache.Set(noteId, info, cache.DefaultExpiration)
}

// Query 查询锁 0 - 无人持有该锁 其他 - 持有该锁的用户ID
func (c *EditLock) Query(noteId string) LockInfo {

	lockInfo := LockInfo{}

	v, found := c.editCache.Get(noteId)
	if found != false {
		// 类型断言，将x转换为Person类型
		lockInfo, ok := v.(LockInfo)
		if ok {
			return lockInfo
		} else {
			lockInfo.UserId = NoLock
			return lockInfo
		}
	} else {
		lockInfo.UserId = NoLock
		return lockInfo
	}

}

// Unlock 解锁
func (c *EditLock) Unlock(noteId string) {
	c.editCache.Delete(noteId)
}
