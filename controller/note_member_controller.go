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

/**
@api {POST} /api/noteMember/share 分享笔记
@apiDescription 分享笔记
@apiName NoteMemberShare
@apiGroup NoteMember

@apiPermission 用户

@apiParam {Integer} id 		被分享者ID
@apiParam {Integer} noteId  笔记ID
@apiParam {String{"user","group"}} shareType 被分享者类型
<ul>

	<li>"user" - 用户 </li>
	<li>"group" - 用户组 </li>

</ul>
@apiParam {Integer{1,2}} [role] 用户权限
<ul>

	<li>1 - 可查看 </li>
	<li>2 - 可编辑 </li>

</ul>

@apiParamExample {json} 请求示例

	{
	    "id": 2,
		"noteId":3,
		"shareType":"user",
		"role":1
	}

@apiSuccessExample 成功响应
HTTP/1.1 200 OK

@apiErrorExample 失败响应
HTTP/1.1 400 Bad Request

参数解析错误
*/

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
	// 若用户非该笔记的拥有者，无法分享
	if userRole != 0 {
		ErrIllegal(ctx, "无权限")
		return
	}

	// 若用户对自己进行分享，则提示错误
	if info.Id == claims.Sub {
		ErrIllegal(ctx, "无法对自己进行分享")
		return
	}

	// 判断是 修改分享权限 还是 新增分享
	exist, err := repo.NoteMemberRepo.Exist(info.Id, info.NoteId, info.ShareType)
	if err != nil {
		ErrSys(ctx, err)
		return
	}

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
		// 事务处理
		err = repo.DBDao.Transaction(func(tx *gorm.DB) error {
			var users []dto.GroupRoleDto
			// 查询用户组用户
			err = tx.Model(&entity.GroupMember{}).Select("user_id,role").Where("belong", info.Id).Find(&users).Error
			if err != nil {
				return err
			}
			for _, user := range users {
				// 判断用户是否已被分享
				exist, err = repo.NoteMemberRepo.Exist(user.UserId, info.NoteId, "user")
				if err != nil {
					return err
				}
				if exist {
					// 若用户已存在 则修改其用户组
					err = tx.First(&entity.NoteMember{}, "user_id = ? AND note_id = ?", user.UserId, info.NoteId).Update("group_id", info.Id).Error
					if err != nil {
						return err
					}
				} else if !exist {
					// 不存在则创建记录
					var tmp entity.NoteMember
					tmp.GroupId = info.Id
					tmp.UserId = user.UserId
					tmp.NoteId = info.NoteId
					// 若被分享的用户是该用户组的管理员，则赋予其编辑权限
					if user.Role == 0 {
						tmp.Role = 2
					} else {
						tmp.Role = user.Role
					}
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

/**
@api {POST} /api/noteMember/cancel 取消分享
@apiDescription 取消分享
@apiName NoteMemberCancel
@apiGroup NoteMember

@apiPermission 用户

@apiParam {Integer} id 		被分享者ID
@apiParam {Integer} noteId  笔记ID
@apiParam {String{"user","group"}} shareType 被分享者类型
<ul>

	<li>"user" - 用户 </li>
	<li>"group" - 用户组 </li>

</ul>


@apiParamExample {json} 请求示例

	{
	    "id": 2,
		"noteId":3,
		"shareType":"user",
	}

@apiSuccessExample 成功响应
HTTP/1.1 200 OK

@apiErrorExample 失败响应
HTTP/1.1 400 Bad Request

参数解析错误
*/

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
	// 若用户非该笔记的拥有者
	if userRole != 0 {
		ErrIllegal(ctx, "无权限")
		return
	}

	// 判断取消分享的用户是否存在
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
		// 将用户本身 用户组 设置为空
		err = repo.DBDao.First(&entity.NoteMember{}, "group_id = ? AND note_id = ? AND role = 0", info.Id, info.NoteId).Update("group_id", 0).Error
		if err != nil {
			ErrSys(ctx, err)
			return
		}

		// 删除其他成员
		err = repo.DBDao.Where("group_id", info.Id).Where("note_id", info.NoteId).Where("role != 0 ").Delete(&entity.NoteMember{}).Error
		if err != nil {
			ErrSys(ctx, err)
			return
		}

	}
}

/**
@api {GET} /api/noteMember/allUser 获取已分享用户
@apiDescription 获取已分享用户
@apiName noteMemberAllUser
@apiGroup noteMember

@apiPermission 用户


@apiParam {Integer} id 笔记ID。

@apiParamExample 请求示例
DELETE /api/noteMember/allUser?id=12

@apiSuccess {User[]} user 用户列表。
@apiSuccess (User) {Integer} id 用户ID。
@apiSuccess (User) {String} name 用户姓名。
@apiSuccess (User) {Integer} role 用户权限。

@apiSuccessExample 成功响应
HTTP/1.1 200 OK

[
    {
        "id": 118,
		"name":"墨小菊",
		"role":2
    }
]

@apiErrorExample 失败响应
HTTP/1.1 400 Bad Request

系统内部错误
*/

// allUser 获取已分享用户
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

/**
@api {GET} /api/noteMember/allGroup 获取已分享用户组
@apiDescription 获取已分享用户组
@apiName noteMemberAllGroup
@apiGroup noteMember

@apiPermission 用户


@apiParam {Integer} id 组ID。

@apiParamExample 请求示例
DELETE /api/noteMember/allGroup?id=12

@apiSuccess {Group[]} group 用户列表。
@apiSuccess (Group) {Integer} id 组ID。
@apiSuccess (Group) {String} name 组名。

@apiSuccessExample 成功响应
HTTP/1.1 200 OK

[
    {
        "id": 1,
		"name":"研发部",
    }
]

@apiErrorExample 失败响应
HTTP/1.1 400 Bad Request

系统内部错误
*/

// allGroup 获取已分享用户组
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
