package dto

type LockDto struct {
	UserId int    `json:"userId"` // 用户ID
	Id     string `json:"id"`     // 笔记ID
}
