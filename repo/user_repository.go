package repo

import (
	"gorm.io/gorm"
	"note/repo/entity"
)

// UserRepository 用户支持层
type UserRepository struct {
}

// Exist 检查用户是否已经存在
func (r *UserRepository) Exist(userId int) (bool, error) {
	if userId == 0 {
		return false, nil
	}
	res := &entity.User{}
	err := DBDao.First(res, "id = ? AND is_delete = ? ", userId, 0).Error
	if err == gorm.ErrRecordNotFound {
		return false, nil
	}
	if err != nil {
		return true, err
	}
	return true, nil
}

// ExistUsername 检查用户名是否已经存在
func (r *UserRepository) ExistUsername(username string) (bool, error) {
	if username == "" {
		return false, nil
	}
	res := &entity.User{}
	err := DBDao.First(res, "username = ? AND is_delete = ? ", username, 0).Error
	if err == gorm.ErrRecordNotFound {
		return false, nil
	}
	if err != nil {
		return true, err
	}
	return true, nil
}

// ExistOpenId 检查工号是否已经存在
func (r *UserRepository) ExistOpenId(openId string) (bool, error) {
	if openId == "" {
		return false, nil
	}
	res := &entity.User{}
	err := DBDao.First(res, "openid = ? AND is_delete = ? ", openId, 0).Error
	if err == gorm.ErrRecordNotFound {
		return false, nil
	}
	if err != nil {
		return true, err
	}
	return true, nil
}

// NameIcon 查询名字和头像
// id: 记录ID
func (r *UserRepository) NameIcon(id int) (*entity.User, error) {
	if id <= 0 {
		return nil, nil
	}

	res := &entity.User{}
	err := DBDao.Select("id, name, avatar").First(res, "id = ? AND is_delete = 0", id).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return res, nil
}

func NewUserRepository() *UserRepository {
	return &UserRepository{}
}
