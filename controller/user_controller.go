package controller

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"io"
	"net/http"
	"note/appconf/dir"
	"note/controller/dto"
	"note/controller/middle"
	"note/logg/applog"
	"note/repo"
	"note/repo/entity"
	"note/reuint"
	"note/reuint/jwt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// NewUserController 创建用户控制器
func NewUserController(router gin.IRouter) *UserController {
	res := &UserController{}
	r := router.Group("/user")
	// 创建用户
	r.POST("/create", Admin, res.create)
	// 查找用户
	r.GET("/search", Admin, res.search)
	// 修改口令
	r.POST("/modifyPwd", Authed, res.modifyPwd)
	// 展示下拉框名称列表
	r.GET("/nameList", Authed, res.nameList)
	// 查询用户个人信息
	r.GET("/info", Authed, res.info)
	// 编辑用户信息
	r.POST("/edit", Authed, res.edit)
	// 更换头像
	r.POST("/updateAvatar", User, res.updateAvatar)
	// 获取头像
	r.GET("/avatar", User, res.avatar)
	// 删除用户
	r.DELETE("/delete", Admin, res.delete)
	return res
}

// UserController 用户控制器
type UserController struct {
}

/**
@api {POST} /api/user/create 创建用户
@apiDescription 创建用户
@apiName UserCreate
@apiGroup User

@apiPermission 管理员

@apiParam {String} openid 工号。
@apiParam {String} name 姓名。
@apiParam {String} [phone] 手机号。
@apiParam {String} [email] 邮箱。
@apiParam {String} password 口令

@apiParamExample {json} 请求示例
{
    "email":"419384048@qq.com",
	"name":"测试用户01",
	"openid":"5001",
	"password":"w7980361",
	"phone":"19212555512"
}

@apiSuccess {Integer} id 用户ID。

@apiSuccessExample 成功响应
HTTP/1.1 200 OK

1

@apiErrorExample 失败响应
HTTP/1.1 400 Bad Request

参数非法，无法解析
*/

// create 创建用户
func (c *UserController) create(ctx *gin.Context) {
	var reqInfo entity.User
	var userCreateDto dto.UserCreateDto

	err := ctx.BindJSON(&userCreateDto)
	// 记录日志
	applog.L(ctx, "创建用户", map[string]interface{}{
		"name": userCreateDto.Name,
	})
	if err != nil {
		ErrIllegal(ctx, "参数非法，无法解析")
		return
	}

	// 工号不能为空
	if len(strings.Trim(userCreateDto.Openid, " ")) == 0 {
		ErrIllegal(ctx, "工号不能为空")
		return
	}

	// 姓名不能为空
	if len(strings.Trim(userCreateDto.Name, " ")) == 0 {
		ErrIllegal(ctx, "用户姓名不能为空")
		return
	}

	// 手机号格式校验
	if len(userCreateDto.Phone) != 0 && !reuint.PhoneValidate(userCreateDto.Phone) {
		ErrIllegal(ctx, "手机号格式错误")
		return
	}

	// 邮箱格式校验
	if len(userCreateDto.Email) != 0 && !reuint.EmailValidate(userCreateDto.Email) {
		ErrIllegal(ctx, "邮箱格式错误")
		return
	}

	// 工号唯一
	exist, err := repo.UserRepo.ExistOpenId(userCreateDto.Openid)
	if err != nil {
		ErrSys(ctx, err)
		return
	}
	if exist {
		ErrIllegal(ctx, "工号已经存在")
		return
	}
	// 用户名唯一
	exist, err = repo.UserRepo.ExistUsername(userCreateDto.Openid)
	if err != nil {
		ErrSys(ctx, err)
		return
	}
	if exist {
		ErrIllegal(ctx, "用户已经存在")
		return
	}
	//姓名转化为拼音首字母
	str, err := reuint.PinyinConversion(userCreateDto.Name)
	if err != nil {
		ErrIllegalE(ctx, err)
		return
	}
	// 口令长度 大于等8位
	if len(strings.Trim(userCreateDto.Password.String(), " ")) < 8 {
		ErrIllegal(ctx, "口令长度不少于8位")
		return
	}
	pwd, salt, err := reuint.GenPasswordSalt(userCreateDto.Password.String())
	if err != nil {
		ErrSys(ctx, err)
		return
	}

	// 赋值并创建
	reqInfo = entity.User{
		Username: userCreateDto.Openid,
		Name:     userCreateDto.Name,
		NamePy:   str,
		Password: entity.Pwd(pwd),
		Salt:     salt,
		Openid:   userCreateDto.Openid,
		Phone:    userCreateDto.Phone,
		Email:    userCreateDto.Email,
	}
	err = repo.DBDao.Create(&reqInfo).Error

	// 返回前端信息
	ctx.JSON(200, reqInfo.ID)
}

/**
@api {GET} /api/user/search 搜索用户
@apiDescription 搜索用户
@apiName UserSearch
@apiGroup User

@apiPermission 管理员

@apiParam {String} keyword 关键字。
@apiParam {Integer} page 页码。
@apiParam {Integer} limit 单页显示数量。

@apiParamExample {json} 请求示例
GET /api/user/search?keyword=w&page=1&limit=20

@apiSuccess {User[]} records 查询结果列表。
@apiSuccess {Integer} total 记录总数。
@apiSuccess {Integer} size 每页显示条数，默认 20。
@apiSuccess {Integer} current 当前页。
@apiSuccess {Integer} pages 总页数。

@apiSuccess {Object} User 用户信息
@apiSuccess {Integer} User.id 用户ID。
@apiSuccess {String} User.username 用户名
@apiSuccess {String} User.name 姓名
@apiSuccess {String} User.namePy 姓名拼音
@apiSuccess {String} User.password 口令加盐摘要Hex
@apiSuccess {String} User.avatar 头像文件名
@apiSuccess {String} User.openid 开放ID
@apiSuccess {String} User.phone 手机号
@apiSuccess {String} User.email 邮箱
@apiSuccess {String} User.sn 身份证
@apiSuccess {String} User.noteTags 笔记标签 - 已弃用
@apiSuccess {String} User.groupTags 用户组标签 - 已弃用
@apiSuccess {Integer} User.isDelete 是否删除 0 - 未删除（默认值） 1 - 删除
@apiSuccess {String} User.createdAt 创建时间

@apiSuccessExample 成功响应
HTTP/1.1 200 OK
{
    "records": [
        {
            "id": 31,
            "username": "5001",
            "name": "测试用户01",
            "namePy": "csyh",
            "password": "",
            "avatar": "",
            "openid": "5001",
            "phone": "19858124520",
            "email": "419384048@qq.com",
            "sn": "",
            "noteTags": "",
            "groupTags": "",
            "isDelete": 0,
            "createdAt": "2024-01-16 14:06:21"
        }
    ],
    "total": 21,
    "size": 20,
    "current": 2,
    "pages": 2
}

@apiErrorExample 失败响应
HTTP/1.1 400 Bad Request

参数非法，无法解析
*/

// search 查找用户
func (c *UserController) search(ctx *gin.Context) {
	keyword := ctx.Query("keyword")
	page, err := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	if err != nil {
		ErrIllegal(ctx, "参数非法，无法解析")
		return
	}
	limit, err := strconv.Atoi(ctx.DefaultQuery("limit", "20"))
	if err != nil {
		ErrIllegal(ctx, "参数非法，无法解析")
		return
	}

	reqInfo, tx := repo.NewPageQueryFnc(repo.DBDao, &entity.User{}, page, limit, func(db *gorm.DB) *gorm.DB {
		//拼音模糊条件
		queryPinyin := db.Where("name_py like ?", fmt.Sprintf("%%%s%%", keyword))
		//姓名模糊条件
		queryName := db.Where("name like ?", fmt.Sprintf("%%%s%%", keyword))
		//用户名模糊条件
		queryUserName := db.Where("username like ?", fmt.Sprintf("%%%s%%", keyword))
		//模糊查询
		db = db.Where("is_delete", 0).Where(queryPinyin.Or(queryName).Or(queryUserName))
		return db
	})
	res := []entity.User{}
	err = tx.Find(&res).Error
	if err != nil {
		ErrSys(ctx, err)
		return
	}
	reqInfo.Records = res
	ctx.JSON(200, &reqInfo)
}

/**
@api {POST} /api/user/modifyPwd 修改用户密码
@apiDescription 修改用户密码

@apiName UserModifyPwd
@apiGroup User

@apiPermission 管理员,用户


@apiParam {String} username 用户名。
@apiParam {String} oldPwd 旧口令。
@apiParam {String} newPwd 新口令。


@apiParamExample {json} 请求示例
{
    "username":"22004",
	"oldPwd":"12345678",
	"newPwd":"Gm123qwe"
}

@apiSuccessExample 成功响应
HTTP/1.1 200 OK

@apiErrorExample 失败响应
HTTP/1.1 400 Bad Request

参数非法，无法解析
*/

// modifyPwd 修改口令
func (c *UserController) modifyPwd(ctx *gin.Context) {

	var info dto.PasswordDto
	err := ctx.BindJSON(&info)
	// 记录日志
	applog.L(ctx, "用户修改口令", map[string]interface{}{
		"username": info.Username,
	})
	if err != nil {
		ErrIllegal(ctx, "参数非法，无法解析")
		return
	}

	// 获取用户信息
	claimsValue, _ := ctx.Get(middle.FlagClaims)
	claims := claimsValue.(*jwt.Claims)

	if len(info.OldPwd) > 0 && len(info.NewPwd) == 0 {
		ErrIllegal(ctx, "请输入口令")
		return
	}

	//数据库搜索用户
	reqInfo := &entity.User{}
	err = repo.DBDao.First(reqInfo, "username = ? AND is_delete = ?", info.Username, 0).Error
	if err == gorm.ErrRecordNotFound {
		ErrIllegal(ctx, "不存在该用户")
		return
	}
	if err != nil {
		ErrSys(ctx, err)
		return
	}

	//if info.NewPwd == info.OldPwd {
	//	ErrIllegal(ctx, "新旧口令不可相同")
	//	return
	//}

	if claims.Type == "user" {
		//旧口令正确性校验
		if reuint.VerifyPasswordSalt(info.OldPwd.String(), reqInfo.Password.String(), reqInfo.Salt) == false {
			ErrIllegal(ctx, "旧口令错误")
			return
		}
	}

	//新口令长度校验
	if len(info.NewPwd.String()) < 8 {
		ErrIllegal(ctx, "口令长度不少于8位")
		return
	}

	pwd, salt, err := reuint.GenPasswordSalt(info.NewPwd.String())
	if err != nil {
		ErrSys(ctx, err)
		return
	}
	reqInfo.Password = entity.Pwd(pwd)
	reqInfo.Salt = salt

	err = repo.DBDao.Save(reqInfo).Error
	if err != nil {
		ErrSys(ctx, err)
		return
	}
}

/**
@api {GET} /api/user/nameList 展示下拉框名称列表
@apiDescription 展示下拉框名称列表
@apiName UserNameList
@apiGroup User

@apiPermission 管理员，用户

@apiParam {String} keyword 关键字。

@apiParamExample 请求示例
GET /api/user/nameList?keyword=zs

@apiSuccess {Integer} id 用户ID。
@apiSuccess {String} name 用户姓名。

@apiSuccessExample 成功响应
HTTP/1.1 200 OK

[
    { "id": 1, "name": "张三"},
]

@apiErrorExample 失败响应
HTTP/1.1 500

系统内部错误
*/

// nameList 展示下拉框名称列表
func (c *UserController) nameList(ctx *gin.Context) {
	keyword := ctx.Query("keyword")
	reqInfo := &[]dto.NameListDto{}

	//拼音模糊条件
	queryPinyin := repo.DBDao.Where("name_py like ?", fmt.Sprintf("%%%s%%", keyword))
	//姓名模糊条件
	queryName := repo.DBDao.Where("name like ?", fmt.Sprintf("%%%s%%", keyword))
	//用户名模糊条件
	queryUserName := repo.DBDao.Where("username like ?", fmt.Sprintf("%%%s%%", keyword))
	//模糊查询
	err := repo.DBDao.Model(&entity.User{}).Select("id,name").Where("is_delete", 0).Where(queryPinyin.Or(queryName).Or(queryUserName)).Find(reqInfo).Error
	if err != nil {
		ErrSys(ctx, err)
		return
	}
	ctx.JSON(200, *reqInfo)
}

/**
@api {GET} /api/user/info 用户个人信息
@apiDescription 查询用户个人信息
@apiName UserInfo
@apiGroup User

@apiPermission 管理员，用户

@apiParam {Integer} userId 用户ID

@apiParamExample 请求示例
GET /api/user/info?userId=1

@apiSuccess {Integer} id 用户ID。
@apiSuccess {String} username 用户名，不可重复。
@apiSuccess {String} name 用户姓名。
@apiSuccess {Integer} openId 用户工号。
@apiSuccess {String} phone 手机号。
@apiSuccess {String} email 邮箱。
@apiSuccess {String} sn 邮箱。

@apiSuccessExample 成功响应
HTTP/1.1 200 OK

{
	"id":1,
	"username":"zs",
	"name": "张三",
	"sn": "354512459555454512",
	"phone":"13855555555",
	"email":"123@qq.com",
	"openId": 1001
}

@apiErrorExample 失败响应
HTTP/1.1 500

系统内部错误
*/

// info 查询用户个人信息
func (c *UserController) info(ctx *gin.Context) {
	var reqInfo dto.UserInfoDto
	userId, _ := strconv.Atoi(ctx.Query("userId"))
	if userId <= 0 {
		ErrIllegal(ctx, "参数异常，无法解析")
		return
	}

	// 获取用户信息
	claimsValue, _ := ctx.Get(middle.FlagClaims)
	claims := claimsValue.(*jwt.Claims)

	if claims.Type == "user" && claims.Sub != userId {
		ErrIllegal(ctx, "权限错误")
		return
	}
	err := repo.DBDao.Model(&entity.User{}).Select("id,username,name,openid,phone,email,sn").First(&reqInfo, "id = ? AND is_delete = ?", userId, 0).Error
	if err == gorm.ErrRecordNotFound {
		ErrIllegal(ctx, "用户不存在或被删除")
		return
	}
	if err != nil {
		ErrSys(ctx, err)
		return
	}

	ctx.JSON(200, reqInfo)
}

/**
@api {POST} /api/user/edit 编辑用户信息
@apiDescription 修改用户信息，管理员具有修改用户全部信息的权限，用户仅可以修改自己的部分信息，如：手机号，邮箱。
@apiName UserEdit
@apiGroup User

@apiPermission 管理员，用户

@apiParam {Integer} id 用户ID。
@apiParam {String} open_id 工号。
@apiParam {String} username 用户名。
@apiParam {String} name 姓名。
@apiParam {String} phone 手机号。
@apiParam {String} email 邮箱。


@apiParamExample {json} 请求示例
{
    "id": 1,
	"openId":"1001",
	"username":"zs",
	"name":"张三",
	"phone":"13855555555",
	"email":"123@qq.com"
}

@apiSuccessExample 成功响应
HTTP/1.1 200 OK

@apiErrorExample 失败响应
HTTP/1.1 500

系统内部错误
*/

// edit 修改用户信息
func (c *UserController) edit(ctx *gin.Context) {
	var info entity.User
	err := ctx.BindJSON(&info)
	// 记录日志
	applog.L(ctx, "编辑用户信息", map[string]interface{}{
		"id":      info.ID,
		"open_id": info.Openid,
	})
	if err != nil {
		ErrIllegal(ctx, "参数非法，无法解析")
		return
	}
	// 获取用户信息
	claimsValue, _ := ctx.Get(middle.FlagClaims)
	claims := claimsValue.(*jwt.Claims)

	// 登录者为普通用户 仅可修改自己的信息且不可修改工号、用户名、姓名
	if claims.Type == "user" && (claims.Sub != info.ID || len(info.Openid) != 0 || len(info.Username) != 0 || len(info.Name) != 0) {
		ErrIllegal(ctx, "权限错误")
		return
	}

	// 查询用户信息
	var reqInfo entity.User
	err = repo.DBDao.First(&reqInfo, "id = ? AND is_delete = ?", info.ID, 0).Error
	if err == gorm.ErrRecordNotFound {
		ErrIllegal(ctx, "用户不存在或被删除")
		return
	}
	if err != nil {
		ErrSys(ctx, err)
		return
	}
	// 若发生工号修改，工号唯一检查
	if reqInfo.Openid != info.Openid {
		existOpenId, err := repo.UserRepo.ExistOpenId(info.Openid)
		if err != nil {
			ErrSys(ctx, err)
			return
		}
		if existOpenId {
			ErrIllegal(ctx, "工号已经存在")
			return
		}
	}
	// 若发生用户名修改，用户名唯一检查
	if reqInfo.Username != info.Username {
		existUsername, err := repo.UserRepo.ExistUsername(info.Username)
		if err != nil {
			ErrSys(ctx, err)
			return
		}
		if existUsername {
			ErrIllegal(ctx, "用户名已经存在")
			return
		}
	}

	// 手机号格式校验
	if info.Phone != "" && !reuint.PhoneValidate(info.Phone) {
		ErrIllegal(ctx, "手机号格式错误")
		return
	}
	// 邮箱格式校验
	if info.Email != "" && !reuint.EmailValidate(info.Email) {
		ErrIllegal(ctx, "邮箱格式错误")
		return
	}
	err = repo.DBDao.Updates(&info).Error
	if err != nil {
		ErrSys(ctx, err)
		return
	}
}

/**
@api {POST} /api/user/updateAvatar 更换头像
@apiDescription 更新用户头像，json传输base64，仅支持支持小于64k jpeg、png。
@apiName UserUpdateAvatar
@apiGroup User

@apiPermission 用户

@apiHeader {String} Content-type multipart/form-data 多类型表单固定值。

@apiParam {String} username 用户名。
@apiParam {File} avatar 头像文件。



@apiSuccessExample 成功响应
HTTP/1.1 200 OK

@apiErrorExample 失败响应
HTTP/1.1 400 Bad Request

用户记录ID非法

@apiErrorExample 失败响应2
HTTP/1.1 400 Bad Request

图片格式错误，无法解析
*/

// updateAvatar 更新头像
func (c *UserController) updateAvatar(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.PostForm("id"))
	if err != nil {
		ErrIllegal(ctx, "参数非法，无法解析")
		return
	}
	// 接取头像
	file, err := ctx.FormFile("avatar")
	if err != nil {
		ErrIllegal(ctx, "参数非法，无法解析")
		return
	}

	applog.L(ctx, "用户修改头像", map[string]interface{}{
		"userId": id,
	})

	// 获取用户信息
	claimsValue, _ := ctx.Get(middle.FlagClaims)
	claims := claimsValue.(*jwt.Claims)

	if (claims.Type == "user" && claims.Sub != id) || (claims.Type != "user") {
		ErrForbidden(ctx, "权限错误")
		return
	}

	// 查询用户信息
	var user entity.User
	err = repo.DBDao.First(&user, "id = ? AND is_delete = 0 ", id).Error
	if err == gorm.ErrRecordNotFound {
		ErrIllegal(ctx, "用户不存在或已被删除")
		return
	}
	if err != nil {
		ErrSys(ctx, err)
		return
	}

	avatarName := fmt.Sprintf("user-%d", id)
	avatarPath := filepath.Join(dir.AvatarDir, avatarName)

	if user.Avatar == "" {
		user.Avatar = avatarName
		// 更新头像字段
		if err = repo.DBDao.Save(&user).Error; err != nil {
			ErrSys(ctx, err)
			return
		}
	}
	if err = ctx.SaveUploadedFile(file, avatarPath); err != nil {
		ErrSys(ctx, err)
		return
	}
}

/**
@api {GET} /api/user/avatar 获取用户头像
@apiDescription 获取用户头像。

@apiName UserAvatar
@apiGroup User

@apiPermission 用户

@apiParam {Integer} id 用户ID。

@apiParamExample {http} 请求示例

GET /api/user/avatar?id=1

@apiErrorExample 失败响应
HTTP/1.1 400 Bad Request

权限错误
*/

// avatar 获取用户头像
func (c *UserController) avatar(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Query("id"))

	if err != nil {
		ctx.JSON(http.StatusNotFound, nil)
		return
	}

	avatarName := fmt.Sprintf("user-%d", id)

	avatarPath := filepath.Join(dir.AvatarDir, avatarName)
	// 防止用户通过 ../../ 的方式读取到操作系统内的重要文件
	if !strings.HasPrefix(avatarPath, dir.AvatarDir) {
		ctx.JSON(http.StatusNotFound, nil)
		return
	}

	file, err := os.Open(avatarPath)
	if err != nil {
		ErrIllegal(ctx, "文件解析失败")
		return
	}
	defer file.Close()

	// 下载文件名称
	ctx.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", avatarName))

	ctx.Header("Content-Type", "image/png")

	_, err = io.Copy(ctx.Writer, file)
	if err != nil {
		ErrSys(ctx, err)
		return
	}
}

/**
@api {DELETE} /api/user/delete 删除用户
@apiDescription 删除用户，如果存在多个用户，其中某个用户删除失败，依然返回200状态码。
该接口仅在数据库操作异常时返回500系统错误的状态码，其他情况均返回200。
@apiName UserDelete
@apiGroup User

@apiPermission 管理员

@apiParam {String} ids 待删除的ID序列，多个ID用","隔开，如：ids=1,99。

@apiParamExample 请求示例
DELETE /api/user/delete?ids=12,24

@apiSuccessExample 成功响应
HTTP/1.1 200 OK

@apiErrorExample 失败响应
HTTP/1.1 500

系统内部错误
*/

// delete 删除用户
func (c *UserController) delete(ctx *gin.Context) {
	ids := ctx.Query("ids")
	//将string转化为[]int
	idArray := reuint.StrToIntSlice(ids)

	// 记录日志
	applog.L(ctx, "删除用户", map[string]interface{}{
		"ids": ids,
	})

	// 将is_delete字段赋值为1
	err := repo.DBDao.Model(&entity.User{}).Where("id in ?", idArray).Update("is_delete", 1).Error
	if err != nil {
		ErrSys(ctx, err)
		return
	}

}
