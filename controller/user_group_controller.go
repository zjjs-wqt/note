package controller

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"note/controller/dto"
	"note/controller/middle"
	"note/logg/applog"
	"note/repo"
	"note/repo/entity"
	"note/reuint"
	"note/reuint/jwt"
	"strconv"
)

// NewUserGroupController 创建用户组控制器
func NewUserGroupController(router gin.IRouter) *UserGroupController {
	res := &UserGroupController{}
	r := router.Group("/userGroup")
	// 创建用户组
	r.POST("/create", User, res.create)
	// 添加成员
	r.POST("/add", User, res.add)
	// 修改角色权限
	r.POST("/change", User, res.change)
	// 删除成员
	r.DELETE("/delete", User, res.delete)
	// 查询所有成员
	r.GET("/all", User, res.all)
	// 查询用户的用户组列表
	r.GET("/list", User, res.list)
	// 查询用户组用户权限
	r.GET("/role", User, res.role)
	return res
}

// UserGroupController 用户组控制器
type UserGroupController struct {
}

/**
@api {POST} /api/userGroup/create 创建用户组
@apiDescription 创建用户组
@apiName UserGroupCreate
@apiGroup UserGroup

@apiPermission 用户

@apiParam {String} name 用户组名称。
@apiParam {String} description 描述。


@apiParamExample {json} 请求示例
{
	"name":"测试用户组",
	"description":"",
}

@apiSuccess {Integer} id 用户组ID。

@apiSuccessExample 成功响应
HTTP/1.1 200 OK

1

@apiErrorExample 失败响应
HTTP/1.1 400 Bad Request

参数非法，无法解析
*/

// create 创建用户组
func (c *UserGroupController) create(ctx *gin.Context) {
	var userGroup entity.UserGroup
	var groupMember entity.GroupMember

	var info dto.UserGroupCreateDto
	err := ctx.BindJSON(&info)
	// 记录日志
	applog.L(ctx, "创建用户组", map[string]interface{}{
		"name": info.Name,
	})
	if err != nil {
		ErrIllegal(ctx, "参数非法，无法解析")
		return
	}
	if info.Name == "undefined" || info.Name == "null" || info.Name == "" {
		ErrIllegal(ctx, "笔记名称为空")
		return
	}

	// 用户组名唯一
	exist, err := repo.UserGroupRepo.Exist(info.Name)
	if err != nil {
		ErrSys(ctx, err)
		return
	}
	if exist {
		ErrIllegal(ctx, "用户组已经存在")
		return
	}

	//用户组名称转化为拼音首字母
	str, err := reuint.PinyinConversion(info.Name)
	if err != nil {
		ErrIllegalE(ctx, err)
		return
	}
	userGroup.Name = info.Name
	userGroup.Description = info.Description
	userGroup.NamePy = str

	// 获取用户信息
	claimsValue, _ := ctx.Get(middle.FlagClaims)
	claims := claimsValue.(*jwt.Claims)

	err = repo.DBDao.Transaction(func(tx *gorm.DB) error {

		// 创建用户组
		err = tx.Create(&userGroup).Error
		if err != nil {
			return err
		}

		// 创建用户组成员
		groupMember.UserId = claims.Sub
		groupMember.Belong = userGroup.ID
		groupMember.Role = 0
		err = tx.Create(&groupMember).Error
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		ErrSys(ctx, err)
		return
	}
	ctx.JSON(200, userGroup.ID)
}

/**
@api {POST} /api/userGroup/add 添加用户
@apiDescription 添加用户
@apiName UserGroupAdd
@apiGroup UserGroup

@apiPermission 用户

@apiParam {Integer} groupId 用户组ID。
@apiParam {Integer} userId 用户ID。
@apiParam {Integer=1,2} role 用户权限。
<ul>
	<li> 1 - 普通用户</li>
	<li> 2 - 维护</li>
</ul>


@apiParamExample {json} 请求示例
{
	"groupId":1,
	"userId":1,
	"role":2
}


@apiSuccessExample 成功响应
HTTP/1.1 200 OK


@apiErrorExample 失败响应
HTTP/1.1 400 Bad Request

参数非法，无法解析
*/

// add 添加成员
func (c *UserGroupController) add(ctx *gin.Context) {
	var reqInfo entity.GroupMember
	var info dto.UserGroupMemberAddDto
	err := ctx.BindJSON(&info)

	// 记录日志
	applog.L(ctx, "添加成员", map[string]interface{}{
		"groupId": info.GroupId,
		"userId":  info.UserId,
	})
	if err != nil {
		ErrIllegal(ctx, "参数非法，无法解析")
		return
	}

	if info.UserId == 0 {
		ErrIllegal(ctx, "请添加成员名称")
		return
	}

	// 获取用户信息
	claimsValue, _ := ctx.Get(middle.FlagClaims)
	claims := claimsValue.(*jwt.Claims)

	// 判断用户是否拥有添加权限
	role, err := repo.UserGroupRepo.Role(claims.Sub, info.GroupId)
	if err == gorm.ErrRecordNotFound {
		ErrIllegal(ctx, "用户组不存在或用户未在用户组")
		return
	} else if err != nil {
		ErrSys(ctx, err)
		return
	}
	if role != 0 {
		ErrIllegal(ctx, "用户权限不足")
		return
	}

	reqInfo.Belong = info.GroupId
	reqInfo.Role = info.Role
	reqInfo.UserId = info.UserId

	// 检测用户是否存在
	exist, err := repo.UserRepo.Exist(info.UserId)
	if err != nil {
		ErrSys(ctx, err)
		return
	}
	if !exist {
		ErrIllegal(ctx, "用户不存在或被删除")
		return
	}

	// 判断用户在该用户组中是否存在角色
	exist, err = repo.UserGroupRepo.ExistUser(reqInfo.UserId, reqInfo.Belong)
	if err != nil {
		ErrSys(ctx, err)
		return
	}
	if exist {
		ErrIllegal(ctx, "用户组中已存在该用户")
		return
	}

	// 创建用户角色
	err = repo.DBDao.Transaction(func(tx *gorm.DB) error {
		// 创建用户组成员记录
		err = tx.Create(&reqInfo).Error
		if err != nil {
			return err
		}

		var notes []int
		err = repo.DBDao.Model(&entity.NoteMember{}).Select("note_id").Where("group_id", info.GroupId).Find(&notes).Error
		if err != nil {
			return err
		}
		// 添加用户至用户组时将分享的笔记分享给该用户
		for _, note := range notes {
			exist, err = repo.NoteMemberRepo.Exist(reqInfo.UserId, note, "user")
			if err != nil {
				return err
			}
			// 若该用户未曾被分享过
			if !exist {
				var tmp entity.NoteMember

				tmp.Role = reqInfo.Role
				tmp.UserId = reqInfo.UserId
				tmp.GroupId = reqInfo.Belong
				tmp.NoteId = note
				err = repo.DBDao.Create(&tmp).Error
				if err != nil {
					return err
				}
			}
			// 若该用户已经被分享过则保留原有权限
		}

		return nil
	})
	if err != nil {
		ErrSys(ctx, err)
		return
	}

}

/**
@api {POST} /api/userGroup/change 修改用户权限
@apiDescription 修改用户权限
@apiName UserGroupChange
@apiGroup UserGroup

@apiPermission 用户

@apiParam {Integer} groupId 用户组ID。
@apiParam {Integer} userId 用户ID。
@apiParam {Integer=1,2} role 用户权限。
<ul>
	<li> 1 - 普通用户</li>
	<li> 2 - 维护</li>
</ul>


@apiParamExample {json} 请求示例
{
	"groupId":1,
	"userId":1,
	"role":2
}


@apiSuccessExample 成功响应
HTTP/1.1 200 OK


@apiErrorExample 失败响应
HTTP/1.1 400 Bad Request

参数非法，无法解析
*/

// change 修改角色权限
func (c *UserGroupController) change(ctx *gin.Context) {
	var reqInfo entity.GroupMember
	var info dto.UserGroupMemberAddDto
	err := ctx.BindJSON(&info)
	// 记录日志
	applog.L(ctx, "修改成员的角色", map[string]interface{}{
		"id":      info.UserId,
		"groupId": info.GroupId,
		"role":    info.Role,
	})
	if err != nil {
		ErrIllegal(ctx, "参数非法，无法解析")
		return
	}
	if info.UserId == 0 {
		ErrIllegal(ctx, "请添加成员名称")
		return
	}

	// 获取用户信息
	claimsValue, _ := ctx.Get(middle.FlagClaims)
	claims := claimsValue.(*jwt.Claims)

	// 判断用户是否拥有修改权限
	role, err := repo.UserGroupRepo.Role(claims.Sub, info.GroupId)
	if err == gorm.ErrRecordNotFound {
		ErrIllegal(ctx, "用户组不存在或用户未在用户组")
		return
	} else if err != nil {
		ErrSys(ctx, err)
		return
	}
	if role != 0 {
		ErrIllegal(ctx, "用户权限不足")
		return
	}
	if info.Role == 0 {
		ErrIllegal(ctx, "无法修改用户权限为管理者")
		return
	}

	// 判断用户在该用户组中是否存在角色
	err = repo.DBDao.First(&reqInfo, "user_id = ? AND belong = ? ", info.UserId, info.GroupId).Error
	if err == gorm.ErrRecordNotFound {
		ErrIllegal(ctx, "用户组中不存在该用户")
		return
	}
	if err != nil {
		ErrSys(ctx, err)
		return
	}

	reqInfo.Role = info.Role

	err = repo.DBDao.Transaction(func(tx *gorm.DB) error {
		// 修改用户
		err = tx.Save(&reqInfo).Error
		if err != nil {
			return err
		}
		// 修改笔记成员中的权限
		err = tx.Model(&entity.NoteMember{}).Where("group_id", info.GroupId).Where("user_id", info.UserId).Update("role", reqInfo.Role).Error
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		ErrSys(ctx, err)
		return
	}
}

/**
@api {DELETE} /api/userGroup/delete 删除用户
@apiDescription 删除用户
@apiName UserGroupDelete
@apiGroup UserGroup

@apiPermission 用户

@apiParam {Integer} id 用户ID。
@apiParam {Integer} groupId 用户组ID

@apiParamExample 请求示例
DELETE /api/userGroup/delete?id=12&groupId=5

@apiSuccessExample 成功响应
HTTP/1.1 200 OK

@apiErrorExample 失败响应
HTTP/1.1 500

系统内部错误
*/

// delete 删除用户
func (c *UserGroupController) delete(ctx *gin.Context) {
	id, _ := strconv.Atoi(ctx.Query("id"))
	if id <= 0 {
		ErrIllegal(ctx, "参数非法，无法解析")
		return
	}
	groupId, _ := strconv.Atoi(ctx.Query("groupId"))
	if groupId <= 0 {
		ErrIllegal(ctx, "参数非法，无法解析")
		return
	}
	// 记录日志
	applog.L(ctx, "删除成员", map[string]interface{}{
		"id":      id,
		"groupId": groupId,
	})

	// 查询是否有该用户组成员
	res := &entity.GroupMember{}
	err := repo.DBDao.First(res, "user_id = ? AND belong = ?", id, groupId).Error
	// 没有该记录
	if err == gorm.ErrRecordNotFound {
		return
	}
	if err != nil {
		ErrSys(ctx, err)
		return
	}
	// 用户为管理员不进行删除操作
	if res.Role == 0 {
		return
	}
	// 取消删除用户的分享
	err = repo.DBDao.Transaction(func(tx *gorm.DB) error {
		err = tx.Delete(&entity.GroupMember{}, res.ID).Error
		if err != nil {
			return err
		}
		err = tx.Where("group_id", groupId).Where("user_id", id).Delete(&entity.NoteMember{}).Error
		if err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		ErrSys(ctx, err)
		return
	}
}

/**
@api {GET} /api/userGroup/all 查询用户组成员
@apiDescription 查询用户组成员
@apiName UserGroupAll
@apiGroup UserGroup

@apiPermission 用户

@apiParam {String} keyword 关键字。
@apiParam {Integer} groupId 用户组ID

@apiParamExample 请求示例
DELETE /api/userGroup/all?keyword=w&groupId=5

@apiSuccess {MemberAllDTO[]} list 用户信息。
@apiSuccess (MemberAllDTO) {Integer} id 用户组成员ID。
@apiSuccess (MemberAllDTO) {Integer} userId 用户ID。
@apiSuccess (MemberAllDTO) {Integer} role 权限。
@apiSuccess (MemberAllDTO) {String} name 用户组名称。

@apiSuccessExample 成功响应
HTTP/1.1 200 OK

[
	{
		"id":4,
		"userId":30,
		"role":0,
		"name":"王沁涛"
	}
]

@apiErrorExample 失败响应
HTTP/1.1 500

系统内部错误
*/

// all 查询所有成员
func (c *UserGroupController) all(ctx *gin.Context) {
	var reqInfo []dto.MemberAllDTO
	groupId, _ := strconv.Atoi(ctx.Query("groupId"))
	if groupId <= 0 {
		ErrIllegal(ctx, "参数非法，无法解析")
		return
	}
	keyword := ctx.Query("keyword")

	// 判断项目是否存在
	exist, err := repo.UserGroupRepo.ExistByID(groupId)
	if err != nil {
		ErrSys(ctx, err)
		return
	}
	if !exist {
		ErrIllegal(ctx, "该用户组不存在或被删除")
		return
	}

	// 获取用户信息
	claimsValue, _ := ctx.Get(middle.FlagClaims)
	claims := claimsValue.(*jwt.Claims)
	// 判断用户在该项目中是否存在角色
	exist, err = repo.UserGroupRepo.ExistUser(claims.Sub, groupId)
	if err != nil {
		ErrSys(ctx, err)
		return
	}
	if !exist {
		ErrIllegal(ctx, "用户组中不存在该用户")
		return
	}

	//拼音模糊条件
	queryPinyin := repo.DBDao.Where("users.name_py like ?", fmt.Sprintf("%%%s%%", keyword))
	//姓名模糊条件
	queryName := repo.DBDao.Where("users.name like ?", fmt.Sprintf("%%%s%%", keyword))
	//用户名模糊条件
	queryOpenid := repo.DBDao.Where("users.openid like ?", fmt.Sprintf("%%%s%%", keyword))

	// 联表后条件查询
	// SELECT group_members.id , role , user_id ,users.`name` FROM group_members LEFT JOIN users on group_members.user_id =  users.id
	err = repo.DBDao.Table("group_members").
		Select("group_members.id , role , user_id ,users.`name`").
		Joins("LEFT JOIN users on group_members.user_id =  users.id").Where("is_delete = 0").
		Where("group_members.belong = ?", groupId).
		Where(queryPinyin.Or(queryName).Or(queryOpenid)).Find(&reqInfo).Error

	ctx.JSON(200, reqInfo)
}

/**
@api {GET} /api/userGroup/list 查询用户的用户组列表
@apiDescription 查询用户的用户组列表
@apiName UserGroupList
@apiGroup UserGroup

@apiPermission 用户

@apiParam {String} keyword 关键字。
@apiParam {Integer} [role=255] 用户权限

@apiParamExample 请求示例
DELETE /api/userGroup/list?keyword=w

@apiSuccess {MemberListDTO[]} list 用户信息。
@apiSuccess (MemberListDTO) {Integer} id 用户组成员ID。
@apiSuccess (MemberListDTO) {Integer} userId 用户ID。
@apiSuccess (MemberListDTO) {Integer} belong 用户组ID。
@apiSuccess (MemberListDTO) {Integer} role 权限。
@apiSuccess (MemberListDTO) {String} name 用户组名称。

@apiSuccessExample 成功响应
HTTP/1.1 200 OK

[
	{
		"id":4,
		"userId":30,
		"belong":2,
		"role":0,
		"name":"测试用户组"
	}
]

@apiErrorExample 失败响应
HTTP/1.1 500

系统内部错误
*/

// list 查询用户的用户组列表
func (c *UserGroupController) list(ctx *gin.Context) {

	var reqInfo []dto.MemberListDTO
	keyword := ctx.Query("keyword")
	role, _ := strconv.Atoi(ctx.DefaultQuery("role", "255"))

	// 获取用户信息
	claimsValue, _ := ctx.Get(middle.FlagClaims)
	claims := claimsValue.(*jwt.Claims)

	queryName := repo.DBDao.Where("user_groups.name like ? ", fmt.Sprintf("%%%s%%", keyword))
	queryNamePy := repo.DBDao.Where("user_groups.name_py like ? ", fmt.Sprintf("%%%s%%", keyword))
	queryDescription := repo.DBDao.Where("user_groups.description like ? ", fmt.Sprintf("%%%s%%", keyword))

	// 联表后条件查询
	// SELECT  group_members.id , role , user_id , user_groups.`name` FROM group_members LEFT JOIN user_groups on group_members.belong = user_groups.id
	tx := repo.DBDao.Table("group_members").
		Select("group_members.id , role , user_id ,belong ,user_groups.`name`").
		Joins("LEFT JOIN user_groups on group_members.belong = user_groups.id").
		Where("group_members.user_id = ?", claims.Sub).
		Where(queryName.Or(queryNamePy).Or(queryDescription))
	if role != 255 {
		tx.Where("group_members.role = ?", role)
	}
	err := tx.Find(&reqInfo).Error
	if err != nil {
		ErrSys(ctx, err)
		return
	}

	ctx.JSON(200, reqInfo)
}

/**
@api {GET} /api/userGroup/role 查询用户组用户权限
@apiDescription 查询用户组用户权限
@apiName UserGroupRole
@apiGroup UserGroup

@apiPermission 用户

@apiParam {Integer} groupId 用户组ID

@apiParamExample 请求示例
DELETE /api/userGroup/role?groupId=5

@apiSuccess  {Integer} role 用户权限。

@apiSuccessExample 成功响应
HTTP/1.1 200 OK

0

@apiErrorExample 失败响应
HTTP/1.1 500

系统内部错误
*/

// role 查询用户组用户权限
func (c *UserGroupController) role(ctx *gin.Context) {

	groupId, _ := strconv.Atoi(ctx.Query("groupId"))

	if groupId <= 0 {
		ErrIllegal(ctx, "参数非法，无法解析")
		return
	}
	// 获取用户信息
	claimsValue, _ := ctx.Get(middle.FlagClaims)
	claims := claimsValue.(*jwt.Claims)

	var tmp entity.GroupMember

	err := repo.DBDao.First(&tmp, "user_id = ? AND belong = ?", claims.Sub, groupId).Error
	if err == gorm.ErrRecordNotFound {
		ErrIllegal(ctx, "用户无权限")
		return
	}
	if err != nil {
		ErrSys(ctx, err)
		return
	}

	ctx.JSON(200, tmp.Role)
}
