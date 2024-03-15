package controller

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
	"note/controller/dto"
	"note/repo"
	"note/repo/entity"
	"note/reuint"
	"strconv"
)

// AyncController 同步控制器
type AyncController struct {
}

// NewAyncController 创建用户控制器
func NewAyncController(router gin.IRouter) *AyncController {
	res := &AyncController{}
	r := router.Group("/user")
	// 用户同步
	r.POST("/aync", res.aync)
	return res
}

/**
@api {post} /api/user/aync 用户同步

@apiGroup User
@apiName UserAync

@apiDescription 该接口由员工管理系统调用，用于实现用户数据同步。
当员工管理系统中的用户发生变更时（如：创建、数据更新），员工管理系统
将会调用该接口对外推送发生变更的用户信息，所有推送消息工位为唯一的用户ID。

@apiHeader {String} Content-Type application/json

@apiParam {Integer} jobNumber 工号，必选无论什么情况都不能为空
@apiParam {Integer=1,2,3,4,5,6,7} state 用户状态。
<ul>
 <li>1 - 正式</li>
 <li>2 - 试用</li>
 <li>3 - 实习</li>
 <li>4 - 外聘</li>
 <li>5 - 劳务派遣</li>
 <li>6 - 离职</li>
 <li>7 - 返聘</li>
</ul>
@apiParam {String} [name] 姓名。
@apiParam {String} [phone] 手机号。
@apiParam {String} [email] 邮箱。

@apiParamExample {json} 用户创建或更新
{
     "jobNumber": 21011,
     "state": 1,
     "name": "张三",
     "phone": "13875648756",
     "email": "123456@mail.com"
}
@apiParamExample {json} 用户禁用（和上面一个样例没有区别，是状态6的特殊情况）
{
     "jobNumber": 21011,
     "state": 6,
     "name": "张三",
     "phone": "13875648756",
     "email": "123456@mail.com"
}
@apiSuccessExample {http} 成功响应
HTTP/1.1 200 OK

@apiErrorExample 失败响应
HTTP/1.1 500
HTTP/1.1 500

系统内部错误
*/

// aync 用户同步
func (c *AyncController) aync(ctx *gin.Context) {
	ayncUser := new(dto.AyncUserDto)

	if err := ctx.BindJSON(ayncUser); err != nil {
		ErrIllegal(ctx, "参数非法，无法解析")
	}

	var user entity.User
	err := repo.DBDao.First(&user, "openid = ? ", ayncUser.JobNumber).Error

	// 若 查找不到 则创建用户
	if err == gorm.ErrRecordNotFound {
		defaultPwd := "Gm123qwe"
		pwd, salt, err := reuint.GenPasswordSalt(defaultPwd)
		if err != nil {
			ErrSys(ctx, err)
			return
		}
		// 密码和盐值
		user.Password = entity.Pwd(pwd)
		user.Salt = salt
	} else if err != nil {
		ErrSys(ctx, err)
		return
	}

	// 状态设置
	if ayncUser.State == 6 {
		user.IsDelete = 1
	} else {
		user.IsDelete = 0
	}

	openId := strconv.Itoa(ayncUser.JobNumber)
	user.Openid = openId
	user.Username = openId
	user.Name = ayncUser.Name
	user.Phone = ayncUser.Phone
	user.Email = ayncUser.Email

	// 姓名转化为拼音首字母
	if len(ayncUser.Name) > 0 {
		str, err := reuint.PinyinConversion(ayncUser.Name)
		if err != nil {
			ErrSys(ctx, err)
			return
		}
		user.NamePy = str
	}

	if err := repo.DBDao.Save(&user).Error; err != nil {
		ErrSys(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, user.ID)
}
