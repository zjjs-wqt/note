package repo

import (
	"errors"
	"gorm.io/gorm"
	"note/repo/entity"
)

// NoteMemberRepository 笔记成员支持层
type NoteMemberRepository struct {
}

// Check 检查用户权限
func (r *NoteMemberRepository) Check(userId int, noteId int) (int, error) {

	var res entity.NoteMember
	err := DBDao.First(&res, "user_id = ? AND note_id = ?", userId, noteId).Error
	if err == gorm.ErrRecordNotFound {
		return -1, nil
	}
	if err != nil {
		return -1, err
	}
	return res.Role, nil
}

// Exist 判断用户是否已被分享
func (r *NoteMemberRepository) Exist(id int, noteId int, shareType string) (bool, error) {
	if id <= 0 || noteId <= 0 || shareType == "" {
		return false, nil
	}
	res := &entity.NoteMember{}
	if shareType == "user" {
		err := DBDao.First(res, "user_id = ? AND note_id = ?", id, noteId).Error
		if err == gorm.ErrRecordNotFound {
			return false, nil
		}
		if err != nil {
			return true, err
		}
	} else if shareType == "group" {
		err := DBDao.First(res, "group_id = ? AND note_id = ?", id, noteId).Error
		if err == gorm.ErrRecordNotFound {
			return false, nil
		}
		if err != nil {
			return true, err
		}
	} else {
		return true, errors.New("未知类型")
	}

	return true, nil
}

func NewNoteMemberRepository() *NoteMemberRepository {
	return &NoteMemberRepository{}
}
