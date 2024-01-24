package repo

import (
	"errors"
	"gorm.io/gorm"
	"note/repo/entity"
)

// NoteRepository 笔记支持层
type NoteRepository struct {
}

// GetNoteById 获取笔记信息(通过笔记ID 和 用户ID)
func (n *NoteRepository) GetNoteById(id int, userId int) (*entity.Note, error) {
	if id <= 0 || userId <= 0 {
		return nil, errors.New("参数错误")
	}
	note := &entity.Note{}

	err := DBDao.First(note, "id = ? AND user_id = ?", id, userId).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return note, nil
}

func NewNoteRepository() *NoteRepository {
	return &NoteRepository{}
}
