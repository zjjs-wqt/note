package repo

import (
	"errors"
	"gorm.io/gorm"
	"note/repo/entity"
)

// NoteMemberRepository 笔记成员支持层
type NoteMemberRepository struct {
}

// Check 检查用户权限 -1 - 无权限 1 - 可查看 2 - 可编辑
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

// GetFolderNotes 获取当前文件夹下的笔记
func (r *NoteMemberRepository) GetFolderNotes(userId int, folderId int) ([]entity.NoteMember, error) {
	if userId <= 0 || folderId <= 0 {
		return nil, errors.New("参数错误")
	}
	noteMembers := []entity.NoteMember{}
	err := DBDao.Where("user_id = ? AND folder_id = ?", userId, folderId).Find(&noteMembers).Error
	if err != nil {
		return nil, err
	}
	return noteMembers, nil
}

// GetFolderOwnerNotes 获取当前文件夹下的自己的笔记
func (r *NoteMemberRepository) GetFolderOwnerNotes(userId int, folderId int) ([]entity.NoteMember, error) {
	if userId <= 0 || folderId <= 0 {
		return nil, errors.New("参数错误")
	}
	noteMembers := []entity.NoteMember{}
	err := DBDao.Where("user_id = ? AND folder_id = ? AND role = 0", userId, folderId).Find(&noteMembers).Error
	if err != nil {
		return nil, err
	}
	return noteMembers, nil
}

func NewNoteMemberRepository() *NoteMemberRepository {
	return &NoteMemberRepository{}
}
