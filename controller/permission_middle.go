package controller

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"net/http"
	"note/controller/middle"
	"note/reuint/jwt"
)

const (
	// 用户类型
	UserTypeAdmin = "admin" // 系统管理员 具有项目管理、用户管理权限
	UserTypeUser  = "user"  // 普通用户	查看任务清单、用户个人信息
	UserTypeAudit = "audit" // 日志审计员 查看操作日志、程序日志

)

var (
	Admin  = Authenticate([]string{UserTypeAdmin})                              // 管理员 具有项目管理、用户管理权限
	User   = Authenticate([]string{UserTypeUser})                               // 普通用户 查看任务清单、用户个人信息
	Audit  = Authenticate([]string{UserTypeAudit})                              // 日志审计员 查看操作日志、程序日志
	Authed = Authenticate([]string{UserTypeAdmin, UserTypeUser, UserTypeAudit}) // 所有已经认证的用户（不限角色），包括用户、管理员、审计员

)

// Authenticate 接口调用权限鉴别
// userType 可访问用户类型
// role 可访问用户角色
func Authenticate(userType []string, role ...int) func(ctx *gin.Context) {

	return func(ctx *gin.Context) {
		// 获取当前用户信息
		claimsValue, _ := ctx.Get(middle.FlagClaims)
		claims := claimsValue.(*jwt.Claims)

		zap.L().Info("接口鉴权", zap.Int("role", claims.Role))
		// 判断用户类型是否在接口访问类型中
		if !isTypeContain(claims.Type, userType) {
			// 用户类型不在可访问类型中，禁止访问
			ErrIllegal(ctx, "权限错误")
			ctx.AbortWithStatus(http.StatusForbidden)
			return
		}
		if len(role) == 0 {
			// 不需要校验角色
			return
		}
		// 登录用户角色不在可访问角色列表中，禁止访问
		if !isRoleContain(claims.Role, role) {
			ErrIllegal(ctx, "权限错误")
			ctx.AbortWithStatus(http.StatusForbidden)
			return
		}
		return
	}
}

// isTypeContain 判断用户是否在接口访问用户类型列表中
func isTypeContain(tye string, typList []string) bool {
	for _, val := range typList {
		if val == tye {
			return true
		}
	}
	return false
}

// isRoleContain 判断用户是否在接口访问角色列表中
func isRoleContain(role int, roleList []int) bool {
	for _, val := range roleList {
		if val == role {
			return true
		}
	}
	return false
}
