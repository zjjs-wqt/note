package controller

import (
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"github.com/emmansun/gmsm/sm2"
	"github.com/emmansun/gmsm/smx509"
	"github.com/gin-gonic/gin"
	"github.com/patrickmn/go-cache"
	"gorm.io/gorm"
	"log"
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
	// 获取随机数
	r.GET("/random", res.random)
	// 验证
	r.POST("/entityAuth", res.entityAuth)
	// 证书绑定
	r.POST("/certBinding", res.certBinding)

	// 验证登录token
	r.GET("/check", res.check)
	res.userCache = cache.New(cache.NoExpiration, 10*time.Minute)

	// 初始化证书池
	reuint.LoadCertsPool()
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
	//}

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
@api {GET} /api/random 获取随机数
@apiDescription 获取32字节随机数。
@apiName AuthRandom
@apiGroup Auth

@apiPermission 匿名


@apiParamExample {HTTP} 请求示例
GET /api/random


@apiSuccessExample 成功响应
HTTP/1.1 200 OK
DJOaejDriTWQzIpPFBypQ+dLclQB+naBWRaJWAqIXEs=

@apiErrorExample 失败响应
HTTP/1.1 500

系统内部错误
*/

// random 获取随机数
func (c *LoginController) random(ctx *gin.Context) {
	buf := make([]byte, 32)
	n, err := rand.Reader.Read(buf)
	if len(buf) != n || err != nil {
		ErrSys(ctx, err)
		return
	}
	ctx.String(200, base64.StdEncoding.EncodeToString(buf))
}

/**
@api {POST} /api/entityAuth 实体鉴别
@apiDescription 管理员验证登录，登录后在cookies加入token字段，并携带管理员信息和类型。
注意：除了系统内部错误外，其他都返还固定错误“身份认证失败”。
@apiName AuthEntityAuth
@apiGroup Auth

@apiPermission 匿名

@apiParam {String} RA 随机数RA base64编码
@apiParam {String} RB 随机数RB base64编码
@apiParam {String} B 可区分标识符B
@apiParam {String} text3 tokenAB所携带自定义文本，这里指用户名
@apiParam {String} signature 签名值 base64编码


@apiSuccess {String="admin","audit"} type 用户类型
<ul>
    <li>admin - 管理员</li>
    <li>audit - 审计员</li>
</ul>
@apiSuccess {String} username 用户名
@apiSuccess {Integer} id 管理员记录ID
@apiSuccess {String} username 用户名
@apiSuccess {[]byte} avatar 头像
@apiSuccess {Integer} exp 会话过期时间，单位Unix时间戳毫秒（ms）

@apiParamExample {json} 请求示例
{
    "Ra": "c+20947+I0eDR8Ce6uj7ciPH+9WimuPlSZBC5YgozA0=",
    "Rb": "cGn7JKJzxoFPDsV0N/5n2/OhAHo7rMDkF+2EKxF2/+4=",
    "B": "dpm",
    "text3": "admin",
    "signature": "MEYCIQD5VxRH+jEpPOjfBD2DqCJKirNOCLlNNkZkpPfHa+EwtQIhAMRlQI1c4NYuWJCyphtwI6uM2W8JAiB+L2NffhSaSHtH"
}

@apiSuccessExample 成功响应
HTTP/1.1 200 OK

{
	"type": "admin",
    "id": 1,
    "username": "admin",
    "avatar": null,
    "exp": 1668523424095
}

@apiErrorExample 失败响应1
HTTP/1.1 500

系统内部错误

@apiErrorExample 失败响应2
HTTP/1.1 400

身份认证失败
*/

// entityAuth 验证
func (c *LoginController) entityAuth(ctx *gin.Context) {
	// 1. 接收TokenAB的json数据
	tokenAB := dto.EntityAuthDto{}
	if err := ctx.BindJSON(&tokenAB); err != nil {
		ErrIllegal(ctx, "参数异常，无法解析")
		return
	}
	// 2. 从数据库里获得证书
	var info entity.Admin
	err := repo.DBDao.First(&info, "username = ?", tokenAB.Text3).Error
	if err == gorm.ErrRecordNotFound {
		ErrIllegal(ctx, "用户不存在")
		return
	}
	if err != nil {
		ErrSys(ctx, err)
		return
	}
	if len(info.Cert) <= 0 {
		ErrIllegal(ctx, "未绑定证书")
		return
	}
	// 3. 解析证书，验证证书可用
	cert, err := reuint.ParseCert(info.Cert)
	if err != nil {
		ErrSys(ctx, err)
		return
	}
	// 证书链验证可用性
	_, err = cert.Verify(smx509.VerifyOptions{Roots: reuint.CertPool, KeyUsages: []smx509.ExtKeyUsage{smx509.ExtKeyUsageAny}})
	if err != nil {
		ErrIllegal(ctx, "证书不可用")
		return
	}

	// 4. 验签，对比随机数是否相同，检验Token中的标识符B是否等于B的可区分标识符
	rA, err := base64.StdEncoding.DecodeString(tokenAB.Ra)
	if err != nil {
		ErrSys(ctx, err)
		return
	}
	rB, err := base64.StdEncoding.DecodeString(tokenAB.Rb)
	if err != nil {
		ErrSys(ctx, err)
		return
	}
	// 签名原文
	msg := append(rA, rB...)
	// 解析签名值
	sig, err := base64.StdEncoding.DecodeString(tokenAB.Signature)
	if err != nil {
		ErrSys(ctx, err)
		return
	}
	// 验签
	verify := sm2.VerifyASN1WithSM2(cert.PublicKey.(*ecdsa.PublicKey), nil, msg, sig)
	log.Println("验签结果 = ", verify)

	// 验签不通过，或tokenAB中的标识符B不等于B的可区分标识符
	if !verify || tokenAB.B != entity.B {
		ErrIllegal(ctx, "身份认证失败")
		return
	}
	// 5. 检验成功，允许登录
	reqInfo := dto.AdminLoginDto{}
	role := ""
	if info.Role == 0 {
		role = "admin"
	} else if info.Role == 1 {
		role = "audit"
	}
	claims := jwt.Claims{Type: role, Sub: info.ID, Exp: time.Now().Add(8 * time.Hour).UnixMilli()}
	token := tokenManager.GenToken(&claims)
	reqInfo.Transform(&claims, &info)
	// 设置头部 Cookies 有效时间为8小时
	ctx.SetCookie("token", token, 8*3600, "", "", false, true)
	ctx.JSON(200, reqInfo)
}

/**
@api {POST} /api/certBinding 证书绑定
@apiDescription 将证书与用户绑定，若用户以绑定证书，则重新绑定
@apiName AuthCertBinding
@apiGroup Auth

@apiPermission 匿名

@apiParam {String} cert 证书
@apiParam {String} RA 随机数RA base64编码
@apiParam {String} RB 随机数RB base64编码
@apiParam {String} B 可区分标识符B
@apiParam {String} text3 tokenAB所携带自定义文本，这里指用户名
@apiParam {String} signature 签名值 base64编码



@apiParamExample {json} 请求示例
{
    "cert": "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUNBakNDQWFlZ0F3SUJBZ0lJQXM2M3g2N1BaNzR3Q2dZSUtvRWN6MVVCZzNVd1FqRUxNQWtHQTFVRUJoTUMKUTA0eER6QU5CZ05WQkFnTUJ1YTFtZWF4bnpFUE1BMEdBMVVFQnd3RzVwMnQ1YmVlTVJFd0R3WURWUVFLREFqbQp0WXZvcjVWRFFUQWVGdzB5TXpBeE1UQXdOek0zTkRWYUZ3MHlOREF4TVRBd056TTNORFZhTUdreER6QU5CZ05WCkJBWU1CdVM0cmVXYnZURVBNQTBHQTFVRUNBd0c1cldaNXJHZk1ROHdEUVlEVlFRSERBYm1uYTNsdDU0eER6QU4KQmdOVkJBb01CdWlFaWVpdXJ6RVBNQTBHQTFVRUN3d0c1NkNVNVkrUk1SSXdFQVlEVlFRRERBbDNiSG5tdFl2bwpyNVV3V1RBVEJnY3Foa2pPUFFJQkJnZ3FnUnpQVlFHQ0xRTkNBQVJyeXQvbk9ZeXNVRmdRRWZ4WVpGUVRUcDY5Cjg2YnIrWTVYRDhrb3U2MnllcVFJM1ZidFMxcXluKzgyWkE4dFVJOFBBWlkyTEp2SWJKMmROZzVwT0F0Z28yQXcKWGpBT0JnTlZIUThCQWY4RUJBTUNCc0F3REFZRFZSMFRBUUgvQkFJd0FEQWRCZ05WSFE0RUZnUVU4SXFoYVl0cwp6bWFyWFVueXhoSFA2QXE3aGRZd0h3WURWUjBqQkJnd0ZvQVVOcFBqRk9kRkNmclY3K292RWkzVG9aWTh3cVF3CkNnWUlLb0VjejFVQmczVURTUUF3UmdJaEFNY0tQa09pTTg4YjhoZWY4ZHlPOHdiMGtpeDFMYXVxc1owOUE4WmMKVUFVMUFpRUEyeFdYMURwUE55cDVtVkdqY25LaDZDT2JpOXF5Q0tNbFRlYUgzdWhpTHJvPQotLS0tLUVORCBDRVJUSUZJQ0FURS0tLS0tCgo=",
    "Ra": "c+20947+I0eDR8Ce6uj7ciPH+9WimuPlSZBC5YgozA0=",
    "Rb": "cGn7JKJzxoFPDsV0N/5n2/OhAHo7rMDkF+2EKxF2/+4=",
    "B": "dpm",
    "text3": "admin",
    "signature": "MEYCIQD5VxRH+jEpPOjfBD2DqCJKirNOCLlNNkZkpPfHa+EwtQIhAMRlQI1c4NYuWJCyphtwI6uM2W8JAiB+L2NffhSaSHtH"
}


@apiSuccessExample 成功响应
HTTP/1.1 200 OK


@apiErrorExample 失败响应1
HTTP/1.1 500

系统内部错误

@apiErrorExample 失败响应2
HTTP/1.1 400

证书绑定失败
*/

// certBinding 证书绑定
func (c *LoginController) certBinding(ctx *gin.Context) {
	params := dto.CertBindingDto{}
	if err := ctx.BindJSON(&params); err != nil {
		ErrIllegal(ctx, "参数非法，无法解析")
		return
	}

	var info entity.Admin
	err := repo.DBDao.First(&info, "username = ?", params.Text3).Error
	if err == gorm.ErrRecordNotFound {
		ErrIllegal(ctx, "用户不存在")
		return
	}
	if err != nil {
		ErrSys(ctx, err)
		return
	}
	// 用户绑定证书后，无法再次绑定
	if len(info.Cert) > 0 {
		ErrIllegal(ctx, "用户已绑定证书")
		return
	}

	// 解析证书，验证可用性

	cert, err := reuint.ParseCert(params.Cert)
	if err != nil {
		ErrIllegal(ctx, "证书无法解析")
		return
	}
	_, err = cert.Verify(smx509.VerifyOptions{Roots: reuint.CertPool, KeyUsages: []smx509.ExtKeyUsage{smx509.ExtKeyUsageAny}})
	if err != nil {
		ErrIllegal(ctx, "证书不可用")
		return
	}

	rA, err := base64.StdEncoding.DecodeString(params.Ra)
	if err != nil {
		ErrSys(ctx, err)
		return
	}
	rB, err := base64.StdEncoding.DecodeString(params.Rb)
	if err != nil {
		ErrSys(ctx, err)
		return
	}
	// 签名原文
	msg := append(rA, rB...)
	// 解析签名值
	sig, err := base64.StdEncoding.DecodeString(params.Signature)
	if err != nil {
		ErrSys(ctx, err)
		return
	}
	// 验签
	verify := sm2.VerifyASN1WithSM2(cert.PublicKey.(*ecdsa.PublicKey), nil, msg, sig)
	log.Println("验签结果 = ", verify)

	// 验签不通过
	if !verify {
		ErrIllegal(ctx, "证书绑定失败")
		return
	}
	// 验签成功，绑定证书
	info.Cert = params.Cert
	if err = repo.DBDao.Save(info).Error; err != nil {
		ErrSys(ctx, err)
		return
	}

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
