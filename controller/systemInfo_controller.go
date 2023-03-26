package controller

import (
	"github.com/gin-gonic/gin"
	"note/appconf"
	"note/repo/entity"
)

// SystemInfoController 系统版本控制器
type SystemInfoController struct {
}

// NewSystemInfoController 创建用户控制器
func NewSystemInfoController(router gin.IRouter) *SystemInfoController {
	res := &SystemInfoController{}
	r := router.Group("/system")
	r.GET("/version", res.version)
	return res
}

/**
@api {GET} /api/system/version 版本号
@apiDescription 查询系统相关配置版本号
@apiName SystemInfoVersion
@apiGroup System

@apiPermission 匿名


@apiParamExample 请求示例
GET /api/system/version

@apiSuccess {String} systemVersion 系统版本号。

@apiSuccessExample 成功响应
HTTP/1.1 200 OK

[
    {
      "systemVersion":"1.0.0",
    }
]

@apiErrorExample 失败响应
HTTP/1.1 500

系统内部错误
*/

// version 读取版本号
func (c *SystemInfoController) version(ctx *gin.Context) {
	var reqInfo entity.SystemInfo
	reqInfo.SystemVersion = appconf.Version
	ctx.JSON(200, reqInfo)
}
