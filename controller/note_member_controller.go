package controller

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"note/controller/dto"
	"note/controller/middle"
	"note/logg/applog"
	"note/repo"
	"note/repo/entity"
	"note/reuint/jwt"
	"strconv"
)

// NewNoteMemberController 创建笔记成员控制器
func NewNoteMemberController(router gin.IRouter) *NoteMemberController {
	res := &NoteMemberController{}
	r := router.Group("/noteMember")
	// 分享笔记
	r.POST("/share", User, res.share)
	// 取消分享
	r.POST("/cancel", User, res.cancel)
	// 获取已分享用户
	r.GET("/allUser", User, res.allUser)
	// 获取已分享用户组
	r.GET("/allGroup", User, res.allGroup)
	return res
}

// NoteMemberController 笔记成员控制器
type NoteMemberController struct {
}

// share 分享笔记
func (c *NoteMemberController) share(ctx *gin.Context) {
	var info dto.NoteShareDto
	var reqInfo entity.NoteMember

	err := ctx.BindJSON(&info)

	// 记录日志
	applog.L(ctx, "分享笔记", map[string]interface{}{
		"id":     info.Id,
		"noteId": info.NoteId,
		"type":   info.ShareType,
		"role":   info.Role,
	})
	if err != nil || info.Id <= 0 || info.NoteId <= 0 || info.Role <= 0 || info.ShareType == "" {
		ErrIllegal(ctx, "参数方法，无法解析")
		return
	}

	// 获取用户信息
	claimsValue, _ := ctx.Get(middle.FlagClaims)
	claims := claimsValue.(*jwt.Claims)
	// 判断用户是否拥有分享权限
	userRole, err := repo.NoteMemberRepo.Check(claims.Sub, info.NoteId)
	if err != nil {
		ErrSys(ctx, err)
		return
	}
	if userRole != 0 {
		ErrIllegal(ctx, "无权限")
		return
	}

	exist, err := repo.NoteMemberRepo.Exist(info.Id, info.NoteId, info.ShareType)
	if err != nil {
		ErrSys(ctx, err)
		return
	}
	//if exist && info.ShareType == "group" {
	//	ErrIllegal(ctx, "已被分享")
	//	return
	//}

	// 若分享对象为用户
	if info.ShareType == "user" {
		// 若用户已存在 则修改其权限
		if exist {
			err = repo.DBDao.First(&entity.NoteMember{}, "user_id = ? AND note_id = ?", info.Id, info.NoteId).Update("role", info.Role).Error
			if err != nil {
				ErrSys(ctx, err)
				return
			}
		} else if !exist { // 不存在则创建记录
			reqInfo.UserId = info.Id
			reqInfo.Role = info.Role
			reqInfo.NoteId = info.NoteId
			err = repo.DBDao.Create(&reqInfo).Error
			if err != nil {
				ErrSys(ctx, err)
				return
			}
		}

	} else if info.ShareType == "group" {
		err = repo.DBDao.Transaction(func(tx *gorm.DB) error {
			var users []dto.GroupRoleDto
			err = tx.Model(&entity.GroupMember{}).Select("user_id,role").Where("belong", info.Id).Find(&users).Error
			if err != nil {
				return err
			}
			for _, user := range users {
				exist, err = repo.NoteMemberRepo.Exist(user.UserId, info.NoteId, "user")
				if err != nil {
					return err
				}

				if exist { // 若用户已存在 则修改其用户组
					err = tx.First(&entity.NoteMember{}, "user_id = ? AND note_id = ?", user.UserId, info.NoteId).Update("group_id", info.Id).Error
					if err != nil {
						return err
					}
				} else if !exist { // 不存在则创建记录
					var tmp entity.NoteMember
					tmp.GroupId = info.Id
					tmp.UserId = user.UserId
					tmp.NoteId = info.NoteId
					tmp.Role = user.Role
					err = repo.DBDao.Create(&tmp).Error
					if err != nil {
						return err
					}
				}
			}
			return nil
		})
		if err != nil {
			ErrSys(ctx, err)
			return
		}
	}

}

// cancel 取消分享
func (c *NoteMemberController) cancel(ctx *gin.Context) {
	var info dto.NoteUnShareDto

	err := ctx.BindJSON(&info)

	// 记录日志
	applog.L(ctx, "取消分享", map[string]interface{}{
		"id":     info.Id,
		"noteId": info.NoteId,
		"type":   info.ShareType,
	})
	if err != nil || info.Id <= 0 || info.NoteId <= 0 || info.ShareType == "" {
		ErrIllegal(ctx, "参数方法，无法解析")
		return
	}

	// 获取用户信息
	claimsValue, _ := ctx.Get(middle.FlagClaims)
	claims := claimsValue.(*jwt.Claims)

	// 判断用户是否拥有取消权限
	userRole, err := repo.NoteMemberRepo.Check(claims.Sub, info.NoteId)
	if err != nil {
		ErrSys(ctx, err)
		return
	}
	if userRole != 0 {
		ErrIllegal(ctx, "无权限")
		return
	}

	exist, err := repo.NoteMemberRepo.Exist(info.Id, info.NoteId, info.ShareType)
	if err != nil {
		ErrSys(ctx, err)
		return
	}
	if !exist {
		ErrIllegal(ctx, "未被分享")
		return
	}

	// 若分享对象为用户
	if info.ShareType == "user" {
		err = repo.DBDao.Where("user_id", info.Id).Where("note_id", info.NoteId).Delete(&entity.NoteMember{}).Error
		if err != nil {
			ErrSys(ctx, err)
			return
		}
	} else if info.ShareType == "group" {
		err = repo.DBDao.First(&entity.NoteMember{}, "group_id = ? AND note_id = ? AND role = 0", info.Id, info.NoteId).Update("group_id", 0).Error

		err = repo.DBDao.Where("group_id", info.Id).Where("note_id", info.NoteId).Where("role != 0 ").Delete(&entity.NoteMember{}).Error
		if err != nil {
			ErrSys(ctx, err)
			return
		}
		if err != nil {
			ErrSys(ctx, err)
			return
		}
	}
}

// 获取已分享用户
func (c *NoteMemberController) allUser(ctx *gin.Context) {

	id, _ := strconv.Atoi(ctx.Query("id"))
	if id <= 0 {
		ErrIllegal(ctx, "参数非法,无法解析")
		return
	}

	// 获取用户信息
	claimsValue, _ := ctx.Get(middle.FlagClaims)
	claims := claimsValue.(*jwt.Claims)
	// 判断用户是否拥有分享权限
	userRole, err := repo.NoteMemberRepo.Check(claims.Sub, id)
	if err != nil {
		ErrSys(ctx, err)
		return
	}
	if userRole != 0 {
		ErrIllegal(ctx, "无权限")
		return
	}

	userList := []dto.UserListDto{}

	err = repo.DBDao.Table("note_members").
		Select("note_members.user_id AS id , role ,users.`name` ").
		Joins("LEFT JOIN users on users.id = note_members.user_id").
		//Where("note_members.group_id = 0").
		Where("note_members.role != 0 ").Where("note_members.note_id", id).Find(&userList).Error
	if err != nil {
		ErrSys(ctx, err)
		return
	}

	ctx.JSON(200, userList)

}

// 获取已分享用户组
func (c *NoteMemberController) allGroup(ctx *gin.Context) {

	id, _ := strconv.Atoi(ctx.Query("id"))
	if id <= 0 {
		ErrIllegal(ctx, "参数非法,无法解析")
		return
	}

	// 获取用户信息
	claimsValue, _ := ctx.Get(middle.FlagClaims)
	claims := claimsValue.(*jwt.Claims)
	// 判断用户是否拥有分享权限
	userRole, err := repo.NoteMemberRepo.Check(claims.Sub, id)
	if err != nil {
		ErrSys(ctx, err)
		return
	}
	if userRole != 0 {
		ErrIllegal(ctx, "无权限")
		return
	}

	groupList := []dto.GroupListDto{}

	err = repo.DBDao.Table("note_members").
		Select("DISTINCT user_groups.id AS id , user_groups.`name` ").
		Joins("LEFT JOIN user_groups on user_groups.id = note_members.group_id").Where("note_members.note_id", id).Where("note_members.group_id != 0 ").Find(&groupList).Error
	if err != nil {
		ErrSys(ctx, err)
		return
	}

	ctx.JSON(200, groupList)

}
