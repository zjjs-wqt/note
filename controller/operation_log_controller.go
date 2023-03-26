package controller

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"note/controller/dto"
	"note/repo"
	"note/repo/entity"
	"strings"
	"time"
)

// NewOperationLogController 创建操作日志控制器
func NewOperationLogController(router gin.IRouter) *OperationLogController {
	res := &OperationLogController{}
	r := router.Group("/oplog")
	// 搜索操作日志
	r.GET("/search", Audit, res.search)
	// 导出日志
	r.GET("/export", Audit, res.export)
	return res
}

// OperationLogController 日志控制器
type OperationLogController struct {
}

/**
@api {GET} /api/oplog/search 搜索
@apiDescription 搜索操作日志，支持分页查询，
查询条件支持时间段搜索，用户类型搜索，用户名搜索，操作名称模糊搜索。
@apiName OplogSearch
@apiGroup Oplog

@apiPermission 审计员

@apiParam {String} [start] 时间段搜索：开始时间
@apiParam {String} [end] 时间段搜索：截止时间
@apiParam {Integer=0,1,2,255} [opType=255] 角色类型
<ul>

	    <li>0 - 匿名</li>
	    <li>1 - 管理员</li>
	    <li>2 - 用户</li>
		<li>255 - 所有</li>

</ul>
@apiParam {Integer} [opID] 用户ID
@apiParam {String} [opName] 操作名称,支持模糊搜索

@apiParam {Integer} [page=1] 分页查询页码，表示第几页，默认 1。
@apiParam {Integer} [limit=20] 单页多少数据，默认 20。

@apiParamExample {get} 请求示例
GET /api/project/search?opType=2&opName=操作&page=1&limit=20

@apiSuccess {Log[]} records 查询结果列表。
@apiSuccess {Integer} total 记录总数。
@apiSuccess {Integer} size 每页显示条数，默认 20。
@apiSuccess {Integer} current 当前页。
@apiSuccess {Integer} pages 总页数。

@apiSuccess {Object} Log 日志数据结构。
@apiSuccess {Integer} Log.id ID。
@apiSuccess {String} Log.CreatedAt 操作时间。
@apiSuccess {String} Log.OpName 操作名称。
@apiSuccess {String} Log.OpParam 操作参数。
@apiSuccess {Integer} Log.userId 用户ID。
@apiSuccess {String} Log.name 姓名。

@apiSuccessExample 成功响应
HTTP/1.1 200 OK

	{
		"records": [
			{
	            "id": 0,
	            "createdAt": "2022-12-09 16:54:17",
	            "opType": 2,
	            "userId": 2,
	            "name": "test",
	            "opName": "退出项目",
	            "opParam": "{}"
	        },
	    ],
		"total": 19,
	    "size": 2,
	    "current": 1,
	    "pages": 10
	}

@apiErrorExample 失败响应
HTTP/1.1 400 Bad Request

权限错误
*/

// search 搜索
func (c *OperationLogController) search(ctx *gin.Context) {
	var param dto.OplogSearchDto

	// 设置默认值
	param.OpType = 255
	param.Page = 1
	param.Limit = 20

	if ctx.ShouldBindQuery(&param) != nil {
		ErrIllegal(ctx, "参数非法，无法解析")
		return
	}

	query, tx := repo.NewPageQueryFnc(repo.DBDao, &entity.Log{}, param.Page, param.Limit, func(db *gorm.DB) *gorm.DB {
		// SELECT logs.id AS log_id,logs.created_at,logs.op_type,logs.op_id,logs.op_name,logs.op_param,users.id AS user_id, users.name
		// FROM logs LEFT JOIN users
		// ON logs.op_id = users.id AND logs.op_type = 2
		// WHERE
		db = db.Table("logs").
			Select("logs.id AS log_id,logs.created_at,logs.op_type,logs.op_id,logs.op_name,logs.op_param,users.id AS user_id, users.name").
			Joins("left join users ON logs.op_id = users.id AND logs.op_type = 2 ")
		if param.Start != 0 && param.End != 0 {
			db = db.Where("logs.created_at BETWEEN ? AND ? ", time.UnixMilli(param.Start), time.UnixMilli(param.End))
		}
		if param.OpType != 255 {
			db = db.Where("logs.op_type = ?", param.OpType)
		}
		if param.OpId != 0 {
			db = db.Where("logs.op_id = ?", param.OpId)
		}
		if param.OpName != "" {
			db = db.Where("logs.op_name like ?",
				fmt.Sprintf("%%%s%%", param.OpName))
		}
		// 前端数据展示排序
		db = db.Order("logs.created_at desc")
		return db
	})
	log := []dto.OplogDto{}
	if err := tx.Find(&log).Error; err != nil {
		ErrSys(ctx, err)
		return
	}

	query.Records = log

	ctx.JSON(200, query)
}

/**
@api {GET} /api/oplog/export 导出日志
@apiDescription 导出日志
查询条件支持时间段搜索，用户类型搜索，用户名搜索，操作名称模糊搜索。
@apiName OplogExport
@apiGroup Oplog

@apiPermission 审计员

@apiParam {String} [start] 时间段搜索：开始时间
@apiParam {String} [end] 时间段搜索：截止时间
@apiParam {Integer=0,1,2,255} [opType=255] 角色类型
<ul>

	    <li>0 - 匿名</li>
	    <li>1 - 管理员</li>
	    <li>2 - 用户</li>
		<li>255 - 所有</li>

</ul>
@apiParam {Integer} [opID] 用户ID
@apiParam {String} [opName] 操作名称,支持模糊搜索

@apiParamExample {get} 请求示例
GET /api/project/search?opType=2&opName=操作

@apiSuccessExample 成功响应
HTTP/1.1 200 OK

@apiErrorExample 失败响应
HTTP/1.1 400 Bad Request

权限错误
*/

// export 导出日志
func (c *OperationLogController) export(ctx *gin.Context) {
	var param dto.OplogExportDto

	// 设置默认值
	param.OpType = 255

	if ctx.ShouldBindQuery(&param) != nil {
		ErrIllegal(ctx, "参数非法，无法解析")
		return
	}

	// 查询条件
	// SELECT logs.id AS log_id,logs.created_at,logs.op_type,logs.op_id,logs.op_name,logs.op_param,users.id AS user_id, users.name
	// FROM logs LEFT JOIN users
	// ON logs.op_id = users.id AND logs.op_type = 2
	// WHERE
	db := repo.DBDao.Table("logs").
		Select("logs.id AS log_id,logs.created_at,logs.op_type,logs.op_id,logs.op_name,logs.op_param,users.id AS user_id, users.name").
		Joins("left join users ON logs.op_id = users.id AND logs.op_type = 2 ")
	if param.Start != 0 && param.End != 0 {
		db = db.Where("logs.created_at BETWEEN ? AND ? ", time.UnixMilli(param.Start), time.UnixMilli(param.End))
	}
	if param.OpType != 255 {
		db = db.Where("logs.op_type = ?", param.OpType)
	}
	if param.OpId != 0 {
		db = db.Where("logs.op_id = ?", param.OpId)
	}
	if param.OpName != "" {
		db = db.Where("logs.op_name like ?",
			fmt.Sprintf("%%%s%%", param.OpName))
	}

	// 下载文件名称
	ctx.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", "操作日志.csv"))
	ctx.Header("Content-Type", "text/csv")
	var cstZone = time.FixedZone("CST", 8*3600)

	_, err := ctx.Writer.WriteString("操作时间, 操作名称, 用户名, 操作参数\n")
	if err != nil {
		ErrSys(ctx, err)
		return
	}
	offset := 0
	limit := 500
	for {
		// 分多次查询 直到查询结果为空
		db = db.Limit(limit).Offset(offset).Order("logs.created_at desc")
		offset += limit
		logList := []dto.OplogDto{}
		if err := db.Find(&logList).Error; err != nil {
			ErrSys(ctx, err)
			return
		}

		if len(logList) == 0 {
			break
		}

		for _, log := range logList {
			// 转义CSV文件中的逗号
			log.OpParam = fmt.Sprintf("\"%s\"", strings.Replace(log.OpParam, "\"", "\"\"", -1))
			// 转义时间
			createdAt := fmt.Sprintf("\"%s\"", time.Time(log.CreatedAt).In(cstZone).Format("2006-01-02 15:04:05"))
			name := ""
			if log.OpType == 1 {
				name = "管理员"
			} else if log.OpType == 2 {
				name = log.Name
			}
			_, err = ctx.Writer.WriteString(fmt.Sprintf("%s,%s,%s,%s\n", createdAt, log.OpName, name, log.OpParam))
			if err != nil {
				ErrSys(ctx, err)
				return
			}
		}
	}
}
