package controller

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"note/controller/dto"
	"note/controller/middle"
	"note/logg/applog"
	"note/repo"
	"note/repo/entity"
	"note/reuint/jwt"
	"strconv"
)

// NewFolderController 创建文件夹控制器
func NewFolderController(router gin.IRouter) *FolderController {
	res := &FolderController{}
	r := router.Group("/folder")
	// 创建文件夹
	r.POST("/create", User, res.create)
	// 获取文件夹列表
	r.GET("/list", User, res.list)
	// 删除文件夹
	r.DELETE("/delete", User, res.delete)
	// 重命名
	r.POST("/rename", User, res.rename)
	// 移动文件夹
	r.POST("/remove", User, res.remove)
	return res
}

// FolderController 文件夹控制器
type FolderController struct {
}

/**
@api {POST} /api/folder/create 创建文件夹
@apiDescription 创建文件夹
@apiName FolderCreate
@apiGroup Folder

@apiPermission 用户

@apiParam {String} name 文件夹名称。
@apiParam {Integer} parentId 父文件夹ID

@apiParamExample {json} 请求示例
{
	"name": "日报",
	"parentId":5,
}


@apiSuccess {Integer} id 文件夹ID。

@apiSuccessExample 成功响应
HTTP/1.1 200 OK

1

@apiErrorExample 失败响应
HTTP/1.1 400 Bad Request

参数非法，无法解析
*/

// create 创建文件夹
func (c *FolderController) create(ctx *gin.Context) {
	var folderDto dto.FolderCreateDto

	err := ctx.BindJSON(&folderDto)
	// 记录日志
	applog.L(ctx, "创建文件夹", map[string]interface{}{
		"name": folderDto.Name,
	})
	if err != nil || folderDto.Name == "" {
		ErrIllegal(ctx, "参数非法，无法解析")
		return
	}
	// 获取用户信息
	claimsValue, _ := ctx.Get(middle.FlagClaims)
	// 获取创建文件夹的用户id
	claims := claimsValue.(*jwt.Claims)

	// 查询文件夹是否已存在
	exist, err := repo.NewFolderRepository().Exist(claims.Sub, folderDto.Name, folderDto.ParentId)
	if err != nil {
		ErrSys(ctx, err)
		return
	}
	if exist {
		ErrIllegal(ctx, "文件夹已存在")
		return
	}

	// 创建文件夹
	folder := entity.Folder{
		UserId:   claims.Sub,
		Name:     folderDto.Name,
		ParentId: folderDto.ParentId,
	}
	err = repo.DBDao.Create(&folder).Error
	if err != nil {
		ErrSys(ctx, err)
		return
	}

	ctx.JSON(200, folder.ID)
}

/**
@api {GET} /api/folder/list 获取文件夹列表
@apiDescription 获取文件夹列表
@apiName FolderList
@apiGroup Folder

@apiPermission 用户

@apiParam {Integer} id 文件夹id。

@apiParamExample {http} 请求示例
GET /api/folder/list?id=0

@apiSuccess {Folder[]} list 文件夹列表。
@apiSuccess (Folder) {Integer} id 文件夹ID。
@apiSuccess (Folder) {Integer} userId 用户ID。
@apiSuccess (Folder) {String} name 文件夹名称。
@apiSuccess (Folder) {Integer} parentId 父文件夹ID。
@apiSuccess (Folder) {String} createdAt 创建时间。

@apiSuccessExample 成功响应
HTTP/1.1 200 OK

[
    {
        "id": 118,
        "userId": 30,
        "name": "最新科技",
        "parentId": 0,
        "createdAt": "2024-01-03 14:09:25"
    }
]

@apiErrorExample 失败响应
HTTP/1.1 400 Bad Request

系统内部错误
*/

// list 获取文件夹列表
func (c *FolderController) list(ctx *gin.Context) {

	id := ctx.Query("id")

	// 获取用户信息
	claimsValue, _ := ctx.Get(middle.FlagClaims)
	// 获取创建文件夹的用户id
	claims := claimsValue.(*jwt.Claims)

	folders := []entity.Folder{}

	if id != "" {
		folderId, err := strconv.Atoi(id)
		if err != nil {
			ErrIllegal(ctx, "参数解析错误")
			return
		}
		err = repo.DBDao.Where("user_id = ? AND parent_id = ?", claims.Sub, folderId).Find(&folders).Error
		if err != nil {
			ErrSys(ctx, err)
			return
		}
	} else {
		err := repo.DBDao.Where("user_id = ? ", claims.Sub).Find(&folders).Error
		if err != nil {
			ErrSys(ctx, err)
			return
		}
	}

	ctx.JSON(200, folders)
}

/**
@api {DELETE} /api/folder/delete 删除文件夹
@apiDescription 删除文件夹。
该接口仅在数据库操作异常时返回500系统错误的状态码，其他情况均返回200。
@apiName FolderDelete
@apiGroup Folder

@apiPermission 用户

@apiParam {Integer} id 文件夹ID。

@apiParamExample 请求示例
DELETE /api/folder/delete?id=12

@apiSuccessExample 成功响应
HTTP/1.1 200 OK

@apiErrorExample 失败响应
HTTP/1.1 500

系统内部错误
*/

// delete 删除文件夹
func (c *FolderController) delete(ctx *gin.Context) {
	// 删除文件夹
	// 会删除当前文件夹以及子文件夹且会删除这些文件夹中的该用户笔记，而其他用户分享给该用户的则会取消分享

	id := ctx.Query("id")
	// 记录日志
	applog.L(ctx, "删除文件夹", map[string]interface{}{
		"id": id,
	})
	// 获取用户信息
	claimsValue, _ := ctx.Get(middle.FlagClaims)
	// 获取创建文件夹的用户id
	claims := claimsValue.(*jwt.Claims)

	folderId, _ := strconv.Atoi(id)
	if folderId <= 0 {
		ErrIllegal(ctx, "参数解析错误")
		return
	}
	folders := []entity.Folder{}

	// 获取 当前文件夹信息 以及 当前文件夹的子文件夹信息
	folder, err := repo.FolderRepo.GetFolder(folderId, claims.Sub)
	if err != nil {
		ErrSys(ctx, err)
		return
	}
	folders = append(folders, *folder)
	folders = append(folders, repo.FolderRepo.GetSubFolders(folderId)...)

	// 获取这些文件夹的笔记成员信息
	noteMembers := []entity.NoteMember{}
	for _, f := range folders {
		noteMemberList, err := repo.NoteMemberRepo.GetFolderNotes(claims.Sub, f.ID)
		if err != nil {
			ErrSys(ctx, err)
			return
		}
		noteMembers = append(noteMembers, noteMemberList...)
	}

	// 获取这些笔记成员表的笔记信息
	notes := []entity.Note{}
	for _, n := range noteMembers {
		// 若该用户是该笔记的拥有者，则查询该笔记信息
		if n.Role == 0 {
			note, err := repo.NoteRepo.GetNoteById(n.NoteId, claims.Sub)
			if err != nil {
				ErrSys(ctx, err)
				return
			}
			// 若能查询到笔记信息
			if note != nil {
				note.IsDelete = 1
				notes = append(notes, *note)
			}
		}
	}

	// 删除相关记录
	err = repo.DBDao.Transaction(func(tx *gorm.DB) error {
		// 删除文件夹以及其子文件
		if len(folders) != 0 {
			dbErr := tx.Delete(&folders).Error
			if dbErr != nil {
				return dbErr
			}
		}

		if len(noteMembers) != 0 {
			// 删除其他用户分享的笔记成员记录
			dbErr := tx.Delete(&noteMembers, "role != 0").Error
			if dbErr != nil {
				return dbErr
			}
			// 将自己用户的笔记成员的组ID设置为空
			dbErr = tx.Model(&noteMembers).Where("role = ?", 0).Update("folder_id", 0).Error
			if dbErr != nil {
				return dbErr
			}
		}

		// 删除笔记
		if len(notes) != 0 {
			dbErr := tx.Save(&notes).Error
			if dbErr != nil {
				return dbErr
			}
		}

		return nil
	})
	if err != nil {
		ErrIllegalE(ctx, err)
		return
	}

	ctx.Status(200)
}

/**
@api {POST} /api/folder/rename 文件夹重命名
@apiDescription 文件夹重命名
@apiName FolderRename
@apiGroup Folder

@apiPermission 用户

@apiParam {Integer} id  文件夹ID
@apiParam {String} name 文件夹名称。
@apiParam {Integer} parentId 父文件夹ID

@apiParamExample {json} 请求示例
{
	"id":5
	"name": "日报",
	"parentId":5,
}

@apiSuccessExample 成功响应
HTTP/1.1 200 OK


@apiErrorExample 失败响应
HTTP/1.1 400 Bad Request

参数非法，无法解析
*/

// rename 重命名
func (c *FolderController) rename(ctx *gin.Context) {
	var info dto.FolderRenameDto

	err := ctx.BindJSON(&info)
	if err != nil || info.Name == "" || info.ID == 0 {
		ErrIllegal(ctx, "参数非法，无法解析")
		return
	}

	// 获取用户信息
	claiumsValue, _ := ctx.Get(middle.FlagClaims)
	// 获取创建文件夹的用户ID
	claims := claiumsValue.(*jwt.Claims)

	// 获取文件夹信息
	folder, err := repo.FolderRepo.GetFolder(info.ID, claims.Sub)
	if err != nil {
		ErrSys(ctx, err)
		return
	}

	// 若文件夹名称未发生修改
	if folder.Name == info.Name {
		ctx.Status(200)
		return
	}

	// 判断文件夹名称是否已存在
	exist, err := repo.FolderRepo.Exist(claims.Sub, info.Name, info.ParentId)
	if err != nil {
		ErrSys(ctx, err)
		return
	}

	// 若文件夹名称已存在
	if exist {
		ErrIllegal(ctx, "文件夹已存在")
		return
	}

	// 文件夹重命名
	folder.Name = info.Name

	err = repo.DBDao.Save(&folder).Error
	if err != nil {
		ErrSys(ctx, err)
		return
	}

	ctx.Status(200)

}

/**
@api {POST} /api/folder/remove 文件夹移动
@apiDescription 文件夹移动
@apiName FolderRemove
@apiGroup Folder

@apiPermission 用户

@apiParam {Integer} id  文件夹ID
@apiParam {Integer} parentId 父文件夹ID

@apiParamExample {json} 请求示例
{
	"id":5
	"parentId":5,
}

@apiSuccessExample 成功响应
HTTP/1.1 200 OK


@apiErrorExample 失败响应
HTTP/1.1 400 Bad Request

参数非法，无法解析
*/

// remove 移动
func (c *FolderController) remove(ctx *gin.Context) {
	var info dto.FolderRemoveDto

	err := ctx.BindJSON(&info)
	if err != nil || info.ID == 0 {
		ErrIllegal(ctx, "参数非法，无法解析")
		return
	}

	// 获取用户信息
	claiumsValue, _ := ctx.Get(middle.FlagClaims)
	// 获取创建文件夹的用户ID
	claims := claiumsValue.(*jwt.Claims)

	// 获取文件夹信息
	folder, err := repo.FolderRepo.GetFolder(info.ID, claims.Sub)
	if err != nil {
		ErrSys(ctx, err)
		return
	}

	// 获取目标文件夹信息
	if info.ParentId != 0 {
		_, err = repo.FolderRepo.GetFolder(info.ParentId, claims.Sub)
		if err != nil {
			ErrIllegal(ctx, "目标文件夹不存在")
			return
		}
		// 若移动的目的地址为该文件夹或该文件夹的子文件夹
		folders := repo.FolderRepo.GetSubFolders(info.ID)
		folders = append(folders, *folder)
		for _, f := range folders {
			if f.ID == info.ParentId {
				ErrIllegal(ctx, "无法移动到自身文件夹及其子文件夹")
				return
			}
		}
	}

	// 若未发生移动
	if info.ParentId == folder.ParentId {
		ctx.Status(200)
		return
	}

	//判断目标文件夹下是否有同名文件夹
	exist, err := repo.FolderRepo.Exist(claims.Sub, folder.Name, info.ParentId)
	if err != nil {
		ErrSys(ctx, err)
		return
	}

	if exist {
		ErrIllegal(ctx, "目标文件夹下存在同名文件夹")
		return
	}

	// 文件夹重命名
	folder.ParentId = info.ParentId

	err = repo.DBDao.Save(&folder).Error
	if err != nil {
		ErrSys(ctx, err)
		return
	}

	ctx.Status(200)
}
