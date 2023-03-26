package controller

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/patrickmn/go-cache"
	"gorm.io/gorm"
	"note/controller/dto"
	"note/repo"
	"note/repo/entity"
	"note/reuint"
	"note/reuint/jwt"
	"strings"
	"time"
)

// NewLoginController 创建登录控制器
func NewLoginController(r gin.IRouter) *LoginController {
	res := &LoginController{}
	// 登录
	r.POST("/login", res.login)
	// 登出
	r.DELETE("/logout", res.logout)
	// 验证登录token
	r.GET("/check", res.check)
	res.userCache = cache.New(cache.NoExpiration, 10*time.Minute)

	return res
}

// LoginController 登录控制器
type LoginController struct {
	userCache *cache.Cache // cache 判断用户口令错误次数
}

// userName string 传入用户名
// pwdAttempts 判断错误口令尝试次数，如果达到5次则锁定10分钟
func (c *LoginController) pwdAttempts(userName string) {
	var usrTry = 1
	userTry, found := c.userCache.Get(userName)
	if found {
		usrTry = userTry.(int)
		usrTry++
	}
	if usrTry >= 5 {
		// 错误尝试达到5次，设置10分钟后清除缓存
		c.userCache.Set(userName, usrTry, 10*time.Minute)
	} else {
		// 将缓存中的userName对应值更新
		c.userCache.Set(userName, usrTry, cache.NoExpiration)
	}
}

/**
@api {POST} /api/login 登录
@apiDescription 用户登录，登录后在cookies加入token字段，并用户信息和类型。
对于单一用户口令错误次数不能超过5次，超过则锁定不允许登录10分钟。
注意：除了系统内部错误，以及超过尝试次数外，其他用户名或口令错误都返还固定错误“用户名或口令错误”。
@apiName AuthLogin
@apiGroup Auth

@apiPermission 匿名

@apiParam {String} username 用户名
@apiParam {String} password 登录口令


@apiSuccess {String} type 用户类型
@apiSuccess {Integer} id 用户记录ID
@apiSuccess {String} username 用户名(工号、手机号、邮箱）
@apiSuccess {String} name 姓名
@apiSuccess {Integer} exp 会话过期时间，单位Unix时间戳毫秒（ms）

@apiParamExample {json} 请求示例
{
    "username": "22001",
    "password": "Gm123qwe"

}

@apiSuccessExample 成功响应
HTTP/1.1 200 OK

{
	"type": "user",
    "id": 1,
    "username": "zhangsan",
    "name": "张三",
    "exp": 1668523424095
}

@apiErrorExample 失败响应1
HTTP/1.1 500

系统内部错误

@apiErrorExample 失败响应2
HTTP/1.1 400

用户名或口令错误
*/

// login 认证登录
func (c *LoginController) login(ctx *gin.Context) {
	var info dto.LoginDto
	var reqInfo dto.LoginInfoDto

	// 用户类型: usr - 用户、 adm - 管理员
	// 用户类型: user - 用户
	var userType string

	// 用户ID
	var userSub int
	err := ctx.BindJSON(&info)
	if err != nil {
		ErrIllegal(ctx, "参数非法，无法解析")
		return
	}
	// 用户名不可为空
	if len(strings.Trim(info.Username, " ")) == 0 {
		ErrIllegal(ctx, "请输入用户名")
		return
	}
	// 判断用户是否被锁定
	// userTry：尝试次数， date：锁定时间， found：是否找到值
	userTry, date, found := c.userCache.GetWithExpiration(info.Username)
	if found && userTry.(int) >= 5 {
		if date.Sub(time.Now()).Minutes() > 1 {
			ErrIllegal(ctx, fmt.Sprintf("用户锁定，请%.0f分钟后再尝试", date.Sub(time.Now()).Minutes()))
		} else {
			ErrIllegal(ctx, fmt.Sprintf("用户锁定，请%.0f秒后再尝试", date.Sub(time.Now()).Seconds()))
		}
		return
	}

	// 判断是否为admin
	admin := &entity.Admin{}
	err = repo.DBDao.First(admin, "username", info.Username).Error
	// 管理员表找到记录，判断为管理员
	if err == nil {
		if reuint.VerifyPasswordSalt(info.Password.String(), admin.Password.String(), admin.Salt) == false {
			ErrIllegal(ctx, "用户名或口令错误")
			// 判断错误口令尝试次数，如果达到5次则锁定10分钟
			c.pwdAttempts(info.Username)
			return
		}
		if admin.Role == 0 {
			userType = "admin"
		} else if admin.Role == 1 {
			userType = "audit"
		}
		userSub = admin.ID
		reqInfo.Username = admin.Username
	}

	// 管理员表未找到记录，判断是否为用户
	if err == gorm.ErrRecordNotFound {
		// 判断是否为用户
		usr := &entity.User{}
		err = repo.DBDao.First(usr, "(username = ? OR openid = ? OR phone = ? OR email = ?)AND is_delete = ?", info.Username, info.Username, info.Username, info.Username, 0).Error
		// 用户表找到记录，判断为用户
		if err == nil {
			if reuint.VerifyPasswordSalt(info.Password.String(), usr.Password.String(), usr.Salt) == false {
				ErrIllegal(ctx, "用户名或口令错误")
				// 判断错误口令尝试次数，如果达到5次则锁定10分钟
				c.pwdAttempts(info.Username)
				return
			}
			userType = "user"
			userSub = usr.ID
			reqInfo.Username = usr.Username
			reqInfo.Name = usr.Name
			reqInfo.GroupTags = strings.Split(usr.GroupTags, ",")
			reqInfo.NoteTags = strings.Split(usr.NoteTags, ",")
		}
		// 用户表未找到记录，抛出异常
		if err == gorm.ErrRecordNotFound {
			ErrIllegal(ctx, "用户名或口令错误")
			return
		}
		// 数据库查询时发生错误
		if err != nil {
			ErrSys(ctx, err)
			return
		}

		if userSub == 0 {
			ErrIllegal(ctx, "用户名或口令错误")
			return
		}
	}

	c.userCache.Delete(info.Username)
	claims := jwt.Claims{Type: userType, Sub: userSub, Exp: time.Now().Add(8 * time.Hour).UnixMilli()}
	token := tokenManager.GenToken(&claims)
	reqInfo.Transform(&claims)
	// 设置头部 Cookies 有效时间为8小时
	ctx.SetCookie("token", token, 8*3600, "", "", false, true)
	ctx.JSON(200, reqInfo)
}

/**
@api {DELETE} /api/logout 登出
@apiDescription 退出登录，无论登出操作是否成功均返回200状态码无任何信息。
@apiName AuthLogout
@apiGroup Auth

@apiPermission 管理员,用户

@apiHeader Cookies token

@apiParamExample {HTTP} 请求示例
DELETE /api/logout
Cookies: token=...

@apiSuccessExample 成功响应
HTTP/1.1 200 OK
*/

// logout 登出
func (c *LoginController) logout(ctx *gin.Context) {
	ctx.SetCookie("token", "", -1, "", "", false, true)
}

/**
@api {GET} /api/check 验证token
@apiDescription 验证token。
@apiName AuthCheck
@apiGroup Auth

@apiPermission 匿名

@apiSuccess {String} type 用户类型
@apiSuccess {Integer} id 用户记录ID
@apiSuccess {String} username 用户名(工号、手机号、邮箱）
@apiSuccess {String} name 姓名
@apiSuccess {[]byte} avatar 头像
@apiSuccess {Integer} exp 会话过期时间，单位Unix时间戳毫秒（ms）

@apiParamExample {HTTP} 请求示例
GET /api/check

{
	"type": "user",
    "id": 1,
    "username": "zhangsan",
    "name": "张三",
    "avatar": null,
    "exp": 1668523424095
}


@apiSuccessExample 成功响应
HTTP/1.1 200 OK


@apiErrorExample 失败响应
HTTP/1.1 500

系统内部错误
*/

// check 验证登录token
func (c *LoginController) check(ctx *gin.Context) {
	// 从cookies中获取token
	token, _ := ctx.Cookie("token")
	if token == "" {
		return
	}
	claims, err := reuint.VerifyToken(token)
	if err != nil {
		ErrSys(ctx, err)
		return
	}

	var name string
	res := dto.LoginInfoDto{}
	if claims.Type == "user" {
		var usr entity.User
		if err = repo.DBDao.First(&usr, "id = ? AND is_delete = 0 ", claims.Sub).Error; err != nil {
			ErrSys(ctx, err)
			return
		}
		res.GroupTags = strings.Split(usr.GroupTags, ",")
		res.NoteTags = strings.Split(usr.NoteTags, ",")
		res.Username = usr.Username
		res.Name = usr.Name
	} else if claims.Type == "admin" || claims.Type == "audit" {
		if err = repo.DBDao.Model(&entity.Admin{}).Select("username").First(&name, "id = ? ", claims.Sub).Error; err != nil {
			ErrSys(ctx, err)
			return
		}
		res.Name = name
	}
	res.Transform(claims)

	ctx.JSON(200, res)
}
