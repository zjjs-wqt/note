package repo

import (
	"gorm.io/gorm"
	"note/repo/entity"
)

// UserGroupRepository 用户组支持层
type UserGroupRepository struct {
}

// Exist 检查用户组是否已经存在
func (r *UserGroupRepository) Exist(name string) (bool, error) {
	if name == "" {
		return false, nil
	}
	res := &entity.UserGroup{}
	err := DBDao.First(res, "name = ? ", name).Error
	if err == gorm.ErrRecordNotFound {
		return false, nil
	}
	if err != nil {
		return true, err
	}
	return true, nil
}

// ExistByID 检查用户组是否已经存在
func (r *UserGroupRepository) ExistByID(id int) (bool, error) {
	if id == 0 {
		return false, nil
	}
	res := &entity.UserGroup{}
	err := DBDao.First(res, "id = ? ", id).Error
	if err == gorm.ErrRecordNotFound {
		return false, nil
	}
	if err != nil {
		return true, err
	}
	return true, nil
}

// ExistUser 检查用户是否已在用户组内
func (r *UserGroupRepository) ExistUser(userId int, groupId int) (bool, error) {
	if userId == 0 || groupId == 0 {
		return false, nil
	}
	res := &entity.GroupMember{}
	err := DBDao.First(res, "user_id = ? AND belong = ? ", userId, groupId).Error
	if err == gorm.ErrRecordNotFound {
		return false, nil
	}
	if err != nil {
		return true, err
	}
	return true, nil
}

// Role 查看用户权限
func (r *UserGroupRepository) Role(userId int, groupId int) (int, error) {
	if userId == 0 || groupId == 0 {
		return -1, nil
	}
	res := &entity.GroupMember{}
	err := DBDao.First(res, "user_id = ? AND belong = ? ", userId, groupId).Error
	if err != nil {
		return -1, err
	}
	return res.Role, nil
}

func NewUserGroupRepository() *UserGroupRepository {
	return &UserGroupRepository{}
}
