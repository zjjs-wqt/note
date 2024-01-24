package repo

import (
	"errors"
	"gorm.io/gorm"
	"note/repo/entity"
)

// FolderRepository 文件夹支持层
type FolderRepository struct {
}

// Exist 判断文件夹是否已存在
func (f *FolderRepository) Exist(id int, name string, parentId int) (bool, error) {
	if id <= 0 || name == "" {
		return false, errors.New("参数错误")
	}
	res := &entity.Folder{}

	err := DBDao.First(res, "user_id = ? AND name = ? AND parent_id = ?", id, name, parentId).Error
	if err == gorm.ErrRecordNotFound {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return true, nil
}

// GetFolder 获取当前文件夹
func (f *FolderRepository) GetFolder(id int, userId int) (*entity.Folder, error) {
	if id <= 0 || userId <= 0 {
		return nil, errors.New("参数错误")
	}

	res := &entity.Folder{}

	err := DBDao.First(res, "id = ? AND user_id = ?  ", id, userId).Error
	if err != nil {
		return nil, err
	}

	return res, nil
}

// GetSubFolders 获取当前文件的子文件夹
func (f *FolderRepository) GetSubFolders(id int) []entity.Folder {
	// 参数非法
	if id <= 0 {
		return nil
	}

	folders := []entity.Folder{}

	DBDao.Where("parent_id = ? ", id).Find(&folders)

	for _, folder := range folders {
		folders = append(folders, f.GetSubFolders(folder.ID)...)
	}

	return folders
}

// GetFolderFullPath 获取当前文件夹的完整路径
func (f *FolderRepository) GetFolderFullPath(id int, userId int, name string) (string, error) {
	if userId <= 0 {
		return "", errors.New("参数错误")
	}
	if id <= 0 {
		return name, nil
	}

	res := &entity.Folder{}

	err := DBDao.First(res, "id = ? AND user_id = ?  ", id, userId).Error
	if err != nil {
		return "", err
	}

	name = res.Name + "/" + name

	name, err = f.GetFolderFullPath(res.ParentId, userId, name)
	if err != nil {
		return "", err
	}

	return name, nil

}

func NewFolderRepository() *FolderRepository {
	return &FolderRepository{}
}
