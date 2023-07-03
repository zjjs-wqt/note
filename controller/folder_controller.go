package controller

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"note/controller/dto"
	"note/controller/middle"
	"note/logg/applog"
	"note/repo"
	"note/repo/entity"
	"note/reuint/jwt"
	"strings"
)

// NewFolderController 创建文件夹控制器
func NewFolderController(router gin.IRouter) *FolderController {
	res := &FolderController{}
	r := router.Group("/folder")
	// 创建文件夹
	r.POST("/create", User, res.create)
	return res
}

// FolderController 文件夹控制器
type FolderController struct {
}

// 创建文件夹 create
func (c FolderController) create(ctx *gin.Context) {
	var folder dto.FolderCreateDto

	err := ctx.BindJSON(&folder)
	// 记录日志
	applog.L(ctx, "创建文件夹", map[string]interface{}{
		"name": folder.Name,
	})
	if err != nil {
		ErrIllegal(ctx, "参数非法，无法解析")
		return
	}
	// 获取用户信息
	claimsValue, _ := ctx.Get(middle.FlagClaims)
	claims := claimsValue.(*jwt.Claims)

	var usr entity.User
	err = repo.DBDao.First(&usr, "id = ?", claims.Sub).Error
	if err != nil {
		ErrIllegalE(ctx, err)
		return
	}
	tags := strings.Split(usr.NoteTags, ",")
	noteTags := strings.Split(folder.Name, ",")
	tags = append(tags, noteTags...)
	result := ""
	temp := map[string]struct{}{}
	for _, item := range tags {
		if _, ok := temp[item]; !ok {
			temp[item] = struct{}{}
			if result == "" {
				result = fmt.Sprintf("%s", item)
			} else {
				result = fmt.Sprintf("%s,%s", result, item)
			}
		}
	}
	err = repo.DBDao.Model(&usr).Where("id", claims.Sub).Update("note_tags", result).Error
	if err != nil {
		ErrIllegalE(ctx, err)
		return
	}

	var info dto.FolderIDDto
	info.ID = folder.ID
	info.Tags = strings.Split(result, ",")
	ctx.JSON(200, info)
}
