package middle

import (
	"github.com/patrickmn/go-cache"
	"time"
)

type EditLock struct {
	editCache *cache.Cache
}

const (
	NoLock = 0 // 无人持有该锁
)

func NewEditLock() *EditLock {
	res := &EditLock{}
	res.editCache = cache.New(5*time.Hour, 7*time.Hour)
	return res
}

// Lock 加锁
func (c *EditLock) Lock(noteId string, userId int) {
	c.editCache.Set(noteId, userId, cache.DefaultExpiration)
}

// Query 查询锁 0 - 无人持有该锁 其他 - 持有该锁的用户ID
func (c *EditLock) Query(noteId string) interface{} {

	v, found := c.editCache.Get(noteId)
	if found != false {
		return v
	} else {
		return 0
	}
}

// Unlock 解锁
func (c *EditLock) Unlock(noteId string) {
	c.editCache.Delete(noteId)
}
