package controller

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"io/fs"
	"net/url"
	"note/appconf/dir"
	"note/controller/dto"
	"note/reuint"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// NewProgramLogController 创建系统日志控制器
func NewProgramLogController(router gin.IRouter) *ProgramLogController {
	res := &ProgramLogController{}
	r := router.Group("/log")
	// 系统日志列表
	r.GET("/list", Audit, res.list)
	// 下载系统日志
	r.GET("/download", Audit, res.download)
	return res
}

// ProgramLogController 系统日志控制器
type ProgramLogController struct {
}

/**
@api {GET} /api/log/list 系统日志列表
@apiDescription 返回所有文件
@apiName LogList
@apiGroup Log

@apiPermission 审计员

@apiParamExample {get} 请求示例
GET /api/log/list

@apiSuccess {FileItem[]} Body 查询结果列表。

@apiSuccess (FileItem) {String} name 文件名。
@apiSuccess (FileItem) {String} updatedAt 最后一次更新时间，格式"YYYY-MM-DD HH:mm:ss"。
@apiSuccess (FileItem) {Integer} size 文件大小，默认单位为（B）。

@apiSuccessExample 成功响应
HTTP/1.1 200 OK

	[{
		"name": "电子签章.pdf",
		"updatedAt": "2022-12-05 13:26:14",
		"size":200
	}]

@apiErrorExample 失败响应
HTTP/1.1 400 Bad Request

权限错误
*/

// list 系统日志列表
func (c *ProgramLogController) list(ctx *gin.Context) {
	p := dir.LogDir
	res := []dto.FileItemDto{}
	zone := time.FixedZone("CST", 8*3600)
	_ = filepath.Walk(p, func(path string, info fs.FileInfo, err error) error {

		if path == p {
			// 忽略根路径
			return nil
		}

		item := &dto.FileItemDto{}
		item.Name = info.Name()
		item.UpdatedAt = info.ModTime().In(zone).Format("2006-01-02 15:04:05")
		//item.Path = strings.Replace(path, dir.PublicAreaDir, "", 1)
		//if runtime.GOOS == "windows" {
		//	item.Path = strings.Replace(item.Path, "\\", "/", -1)
		//}
		item.Size = info.Size()
		item.Type = "file"
		if info.IsDir() {
			item.Type = "dir"
		}
		res = append(res, *item)

		if info.IsDir() {
			// 阻止递归
			return filepath.SkipDir
		}
		return nil
	})
	// 排序首先按照文件类型降序，再按照更新时间升序，最后再按照文件名称降序
	sort.Slice(res, func(i, j int) bool {
		if res[i].Type != res[j].Type {
			return len(res[i].Type) < len(res[j].Type)
		}
		if res[i].UpdatedAt != res[j].UpdatedAt {
			return res[i].UpdatedAt > res[j].UpdatedAt
		}
		return res[i].Name < res[j].Name
	})

	ctx.JSON(200, res)
}

/**
@api {GET} /api/log/download 下载
@apiDescription 下载文件或文件夹（文件夹以打包成压缩包的形式下载）。
@apiName LogDownload
@apiGroup Log

@apiPermission 审计员

@apiParam {String} path 下载资源所在目录，URI编码（相对于日志存储区的根目录的绝对路径），多个资源以","隔开。

path= area/1,2,3,4
@apiParamExample {get} 请求示例
GET /api/log/download?path=/%E6%A0%87%E5%87%86/,/%E6%8A%A5%E5%91%8A/

@apiSuccessExample 成功响应
HTTP/1.1 200 OK

@apiErrorExample 失败响应
HTTP/1.1 500

系统内部错误
*/

// download 下载
func (c *ProgramLogController) download(ctx *gin.Context) {
	pathStr, _ := url.QueryUnescape(ctx.Query("path"))
	fileList := strings.Split(pathStr, ",")
	if len(fileList) == 0 {
		ErrIllegal(ctx, "路径错误")
		return
	}

	for i, filename := range fileList {
		filePath := filepath.Join(dir.LogDir, filename)
		if !strings.HasPrefix(filePath, dir.LogDir) || filePath == dir.LogDir {
			ErrIllegal(ctx, "路径错误")
			return
		}
		fileList[i] = filePath
	}

	info, err := os.Stat(fileList[0])
	if err != nil {
		ErrIllegal(ctx, "路径错误")
		return
	}

	// 单文件下载
	if len(fileList) == 1 && !info.IsDir() {
		file, err := os.Open(fileList[0])
		defer file.Close()
		// 下载文件名称
		ctx.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", info.Name()))
		//获取文件的后缀(文件类型)
		ctx.Header("Content-Type", reuint.GetMIME(path.Ext(info.Name())))
		_, err = io.Copy(ctx.Writer, file)
		if err != nil {
			ErrSys(ctx, err)
		}
		// 单文件情况提前返回
		return
	}

	// 多文件或目录 打包压缩下载
	ctx.Header("Content-Disposition", "attachment; filename=archive.zip")
	ctx.Header("Content-Type", "application/zip")
	err = reuint.Zip(ctx.Writer, fileList...)
	if err != nil {
		ErrSys(ctx, err)
		return
	}
}
