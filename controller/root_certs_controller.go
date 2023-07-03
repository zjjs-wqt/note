package controller

import (
	"fmt"
	"github.com/emmansun/gmsm/smx509"
	"github.com/gin-gonic/gin"
	"io"
	"io/fs"
	"net/url"
	"note/appconf/dir"
	"note/controller/dto"
	"note/logg/applog"
	"note/reuint"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// NewRootCertsController 创建根证书管理
func NewRootCertsController(router gin.IRouter) *RootCertsController {
	res := &RootCertsController{}
	r := router.Group("/rootCerts")
	// 列表
	r.GET("/list", Admin, res.list)
	// 上传
	r.POST("/upload", Admin, res.upload)
	// 下载
	r.GET("/download", Admin, res.download)
	// 删除
	r.DELETE("/remove", Admin, res.remove)
	return res
}

// RootCertsController 根证书管理控制器
type RootCertsController struct {
}

/**
@api {GET} /api/rootCerts/list 列表
@apiDescription 返回指定目录下所有的文件或目录。
@apiName RootCertsList
@apiGroup RootCerts

@apiPermission 管理员


@apiParamExample {get} 请求示例
GET /api/rootCerts/list

@apiSuccess {CertItemDto[]} Body 查询结果列表。

@apiSuccess (CertItemDto) {String} name 文件名。
@apiSuccess (CertItemDto) {String} createdAt 根证书上传时间，格式"YYYY-MM-DD HH:mm:ss"。

@apiSuccessExample 成功响应
HTTP/1.1 200 OK

[
    {
        "name": "RSA根证书.crt",
        "createdAt": "2022-12-09 16:58:57"
    },
    {
        "name": "SM2根证书.crt",
        "createdAt": "2022-12-09 17:41:58"
    }
]

@apiErrorExample 失败响应
HTTP/1.1 500

系统内部错误
*/

// list 查看根证书列表
func (c *RootCertsController) list(ctx *gin.Context) {
	res := []dto.CertItemDto{}
	zone := time.FixedZone("CST", 8*3600)
	_ = filepath.Walk(dir.RootCertDir, func(path string, info fs.FileInfo, err error) error {

		if path == dir.RootCertDir {
			// 忽略根路径
			return nil
		}

		item := &dto.CertItemDto{}
		item.Name = info.Name()
		item.CreatedAt = info.ModTime().In(zone).Format("2006-01-02 15:04:05")
		res = append(res, *item)
		return nil
	})
	// 根据上传时间排序
	sort.Slice(res, func(i, j int) bool {
		return res[i].CreatedAt > res[j].CreatedAt
	})

	ctx.JSON(200, res)
}

/**
@api {POST} /api/rootCerts/upload 上传
@apiDescription 以表单的方式上传根证书。

@apiName RootCertsUpload
@apiGroup RootCerts

@apiPermission 管理员

@apiHeader {String} Content-type multipart/form-data 多类型表单固定值。

@apiParam {[]File} files 上传文件列表。

@apiSuccessExample 成功响应
HTTP/1.1 200 OK

@apiErrorExample 失败响应
HTTP/1.1 500

系统内部错误
*/

// upload 上传根证书
func (c *RootCertsController) upload(ctx *gin.Context) {
	// 获取表单文件
	form, _ := ctx.MultipartForm()
	files := form.File["files"]
	// 记录日志
	applog.L(ctx, "上传根证书", map[string]interface{}{
		"files": "上传根证书",
	})

	for _, file := range files {
		filePath := filepath.Join(dir.RootCertDir, file.Filename)
		_, err := os.Stat(filePath)
		if err == nil {
			ErrIllegal(ctx, "根证书已存在")
			return
		}
		open, err := file.Open()
		if err != nil {
			ErrSys(ctx, err)
			return
		}
		temp, err := io.ReadAll(open)
		if err != nil {
			ErrSys(ctx, err)
			return
		}
		_ = open.Close()
		cert, err := smx509.ParseCertificate(reuint.Decode2DER(temp))
		if err != nil {
			ErrIllegal(ctx, "解析证书失败")
			return
		}
		// 证书链验证
		verify, err := cert.Verify(smx509.VerifyOptions{Roots: reuint.CertPool, KeyUsages: []smx509.ExtKeyUsage{smx509.ExtKeyUsageAny}})
		// 若存在证书链，且验证失败，则不加入证书池
		if err != nil && len(verify) != 0 {
			continue
		}
		reuint.CertPool.AddCert(cert)
		err = ctx.SaveUploadedFile(file, filePath)
		if err != nil {
			ErrSys(ctx, err)
			return
		}
	}
}

/**
@api {GET} /api/rootCerts/download 下载
@apiDescription 下载根证书。
@apiName RootCertsDownload
@apiGroup RootCerts

@apiPermission 管理员

@apiParam {String} name 根证书名称，URI编码。

@apiParamExample {get} 请求示例
GET /api/rootCerts/download?name=RSA%E6%A0%B9%E8%AF%81%E4%B9%A6.crt

@apiSuccessExample 成功响应
HTTP/1.1 200 OK
@apiHeader Content-Disposition
@apiHeader Content-Type

@apiErrorExample 失败响应
HTTP/1.1 500

系统内部错误
*/
// download 下载根证书
func (c *RootCertsController) download(ctx *gin.Context) {
	name, _ := url.QueryUnescape(ctx.Query("name"))

	// 记录日志
	applog.L(ctx, "下载根证书", map[string]interface{}{
		"name": name,
	})

	filePath := filepath.Join(dir.RootCertDir, name)
	if !strings.HasPrefix(filePath, dir.RootCertDir) || filePath == dir.RootCertDir {
		ErrIllegal(ctx, "路径错误")
		return
	}
	info, err := os.Stat(filePath)
	if err != nil {
		ErrIllegal(ctx, "路径错误")
		return
	}

	// 单文件下载
	if !info.IsDir() {
		file, err := os.Open(filePath)
		if err != nil {
			ErrSys(ctx, err)
			return
		}
		defer file.Close()
		// 下载文件名称
		ctx.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", url.QueryEscape(info.Name())))
		//获取文件的后缀(文件类型)
		ctx.Header("Content-Type", reuint.GetMIME(path.Ext(info.Name())))
		_, err = io.Copy(ctx.Writer, file)
		if err != nil {
			ErrSys(ctx, err)
		}
	}
}

/**
@api {DELETE} /api/rootCerts/remove 删除
@apiDescription 删除根证书。
@apiName RootCertsRemove
@apiGroup RootCerts

@apiPermission 管理员

@apiParam {String} name 根证书名称，URI编码。

@apiParamExample {delete} 请求示例
DELETE /api/rootCerts/remove?name=RSA%E6%A0%B9%E8%AF%81%E4%B9%A6.crt

@apiSuccessExample 成功响应
HTTP/1.1 200 OK

@apiErrorExample 失败响应
HTTP/1.1 500

系统内部错误
*/

// remove 删除
func (c *RootCertsController) remove(ctx *gin.Context) {
	name, _ := url.QueryUnescape(ctx.Query("name"))

	// 记录日志
	applog.L(ctx, "删除根证书", map[string]interface{}{
		"name": name,
	})

	p := filepath.Join(dir.RootCertDir, name)
	if !strings.HasPrefix(p, dir.RootCertDir) || p == dir.RootCertDir {
		ErrIllegal(ctx, "路径错误")
		return
	}

	err := os.RemoveAll(p)
	if err != nil {
		ErrSys(ctx, err)
		return
	}
	reuint.LoadCertsPool()
}
