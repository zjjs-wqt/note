package noteDaemon

import (
	"note/appconf"
	"note/appconf/dir"
	"note/repo"
	"note/repo/entity"
	"os"
	"path/filepath"
	"time"
)

var _globalL *Note

// Note 笔记模块
type Note struct {
	maxKeepDays int // 日志最大存储时间（单位：天），注意若该值小于等于0则表示不删除。
}

func InitNote(cfg *appconf.Application) {
	if _globalL != nil {
		return
	}
	_globalL = &Note{
		maxKeepDays: cfg.NoteKeepMaxDays,
	}
	// 日志超时删除精灵
	go _globalL.timeoutDeleteDaemon()
}

// 超时删除笔记清理精灵
// 注意该函数不应抛出任何错误，若有错误请手动恢复并打印，继续下一个循环。
func (l *Note) timeoutDeleteDaemon() {
	if _globalL.maxKeepDays > 0 {
		var notes []string
		for {
			now := time.Now().AddDate(0, 0, -_globalL.maxKeepDays).Format("2006-01-02 15:04:05")
			repo.DBDao.Model(&entity.Note{}).Select("id").Where("updated_at < ? AND is_delete = 1", now).Find(&notes)
			for _, note := range notes {
				noteDir := filepath.Join(dir.NoteDir, note)
				os.RemoveAll(noteDir)
			}
			repo.DBDao.Where("updated_at < ? AND is_delete = 1", now).Delete(&entity.Note{})
			time.Sleep(24 * time.Hour)
		}
	}
}
