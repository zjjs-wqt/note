package controller

import (
	"bufio"
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"io"
	"math/rand"
	"note/appconf/dir"
	"note/controller/dto"
	"note/controller/middle"
	"note/logg/applog"
	repo "note/repo"
	"note/repo/entity"
	"note/reuint"
	"note/reuint/jwt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// NewNoteController 创建笔记控制器
func NewNoteController(router gin.IRouter) *NoteController {
	res := &NoteController{}
	r := router.Group("/note")
	// 创建项目笔记
	r.POST("/create", User, res.create)
	// 获取笔记信息
	r.GET("/info", User, res.info)
	// 获取笔记列表
	r.GET("/noteList", Authed, res.noteList)
	// 笔记重命名
	r.POST("/rename", User, res.rename)
	// 更新笔记内容
	r.POST("/content", User, res.contentPost)
	// 获取笔记内容
	r.GET("/content", User, res.contentGet)
	// 上传笔记资源
	r.POST("/assert", User, res.assertPost)
	// 下载笔记资源
	r.GET("/assert", User, res.assertGet)
	// 保存笔记至文件夹
	r.GET("/group", User, res.group)
	// 获取笔记编辑锁
	r.POST("/lock", User, res.lock)
	// 取消编辑
	r.POST("/cancel", User, res.cancel)
	// 导出笔记
	r.GET("/export", User, res.export)
	// 删除笔记
	r.DELETE("/delete", User, res.delete)
	// 恢复笔记
	r.GET("/restore", Authed, res.restore)
	return res
}

// NoteController 笔记控制器
type NoteController struct {
}

/**
@api {POST} /api/note/create 创建
@apiDescription 创建笔记
@apiName NoteCreate
@apiGroup Note

@apiPermission 用户



@apiParam {String} title 笔记名。
@apiParam {Integer} priority 优先级，默认为0，数值越大优先级越高显示位置越靠前。
@apiParam {Integer} folderId 文件夹ID。

@apiParamExample {json} 请求示例
{
    "title":"日报",
	"priority":0,
	"folderId":0
}

@apiSuccess {Integer} id 文档记录ID。

@apiSuccessExample 成功响应
HTTP/1.1 200 OK

1

@apiErrorExample 失败响应
HTTP/1.1 400 Bad Request

参数非法，无法解析
*/

// create 创建笔记
func (c *NoteController) create(ctx *gin.Context) {
	var note entity.Note
	var noteCreateDto dto.NoteCreateDto

	err := ctx.BindJSON(&noteCreateDto)
	// 记录日志
	applog.L(ctx, "创建笔记", map[string]interface{}{
		"name": noteCreateDto.Title,
	})
	if err != nil {
		ErrIllegal(ctx, "参数非法，无法解析")
		return
	}

	if noteCreateDto.Title == "undefined" || noteCreateDto.Title == "null" || noteCreateDto.Title == "" {
		ErrIllegal(ctx, "笔记名称为空")
		return
	}

	// 获取用户信息
	claimsValue, _ := ctx.Get(middle.FlagClaims)
	claims := claimsValue.(*jwt.Claims)
	// 若该用户笔记名已存在，则生成随机数在笔记名后
	err = repo.DBDao.First(&entity.Note{}, "title = ? AND user_id = ? ", noteCreateDto.Title, claims.Sub).Error
	if err == gorm.ErrRecordNotFound {
		note.Title = noteCreateDto.Title
	} else if err != nil {
		ErrSys(ctx, err)
		return
	} else if err == nil {
		// 生成随机数
		r := fmt.Sprintf("%04v", rand.New(rand.NewSource(time.Now().UnixNano())).Int31n(10000))
		note.Title = fmt.Sprintf("%s%s", noteCreateDto.Title, r)
	}

	//项目名称拼音缩写生成
	str, err := reuint.PinyinConversion(note.Title)
	if err != nil {
		ErrIllegalE(ctx, err)
		return
	}

	// 笔记记录 赋值
	note.UserId = claims.Sub
	note.TitlePy = str
	note.Priority = noteCreateDto.Priority
	note.IsDelete = 0

	err = repo.DBDao.Transaction(func(tx *gorm.DB) error {
		// 创建笔记记录
		dberr := tx.Create(&note).Error
		if dberr != nil {
			return dberr
		}

		// 创建文件目录
		notePath := filepath.Join(dir.NoteDir, strconv.Itoa(note.ID))
		dberr = os.MkdirAll(notePath, os.ModePerm)
		if dberr != nil {
			return dberr
		}
		// 生成文件名称
		now := time.Now().Format("20060102150405")
		r := fmt.Sprintf("%04v", rand.New(rand.NewSource(time.Now().UnixNano())).Int31n(10000))
		filename := fmt.Sprintf("%s%s.md", now, r)
		note.Filename = filename
		docFile := filepath.Join(notePath, note.Filename)
		// 生成文件
		_, dberr = os.Create(docFile)
		if dberr != nil {
			return dberr
		}
		// 更新文件信息
		find := tx.Where("id", note.ID).Find(&entity.Note{})
		dberr = find.Update("filename", note.Filename).Error
		if dberr != nil {
			return dberr
		}

		// 生成笔记成员记录
		noteMember := &entity.NoteMember{}
		noteMember.Role = 0
		noteMember.UserId = claims.Sub
		noteMember.NoteId = note.ID
		noteMember.FolderId = noteCreateDto.FolderId
		dberr = tx.Create(&noteMember).Error
		if dberr != nil {
			return dberr
		}

		return nil
	})
	if err != nil {
		ErrIllegalE(ctx, err)
		return
	}

	ctx.JSON(200, note.ID)
}

/**
@api {GET} /api/note/info 获取笔记信息
@apiDescription 获取笔记信息
@apiName NoteInfo
@apiGroup Note

@apiPermission 用户

@apiParam {Integer} id 笔记ID。

@apiParamExample {json} 请求示例
GET /api/note/info?id=13

@apiSuccess {Integer} id 笔记ID。
@apiSuccess {String} updatedAt 更新时间
@apiSuccess {String} title 标题。
@apiSuccess {String} remark 备注。
@apiSuccess {String} userName 笔记拥有者名称。
@apiSuccess {String} tags 标签。
@apiSuccess {Integer} role 用户权限。
@apiSuccess {Integer} isDelete 用户权限。
@apiSuccess {Integer} folderId 所属文件夹ID


@apiSuccessExample 成功响应
HTTP/1.1 200 OK
{
	"id": 3,
	"title": "测试文档",
	"remark": "",
	"userName": "王沁涛",
	"tags": "运维",
	"role": 0,
	"isDelete": 0,
	"updatedAt": "2023-03-22 16:06:05"
}

@apiErrorExample 失败响应
HTTP/1.1 400 Bad Request

参数非法，无法解析
*/

// info 获取笔记信息
func (c *NoteController) info(ctx *gin.Context) {
	var noteMember entity.NoteMember
	var note entity.Note
	var user entity.User
	var info dto.NoteInfoDto
	id, _ := strconv.Atoi(ctx.Query("id"))
	if id <= 0 {
		ErrIllegal(ctx, "参数非法，无法解析")
		return
	}

	// 获取用户信息
	claimsValue, _ := ctx.Get(middle.FlagClaims)
	claims := claimsValue.(*jwt.Claims)

	// 获取笔记信息
	err := repo.DBDao.First(&note, "id = ? ", id).Error
	if err == gorm.ErrRecordNotFound {
		ErrIllegal(ctx, "笔记不存在")
		return
	}
	if err != nil {
		ErrSys(ctx, err)
		return
	}
	info.UpdatedAt = note.UpdatedAt
	info.ID = note.ID

	// 判断是否拥有笔记权限
	err = repo.DBDao.First(&noteMember, "user_id = ? AND note_id = ?", claims.Sub, id).Error
	if err == gorm.ErrRecordNotFound {
		ErrIllegal(ctx, "未拥有该笔记权限")
		return
	}
	if err != nil {
		ErrSys(ctx, err)
		return
	}
	info.NoteGroup = noteMember.NoteGroup
	info.Role = noteMember.Role
	info.Remark = noteMember.Remark
	info.Title = note.Title
	info.IsDelete = note.IsDelete
	info.FolderId = noteMember.FolderId

	// 获取笔记拥有者信息
	err = repo.DBDao.First(&user, "id = ? AND is_delete = 0 ", note.UserId).Error
	if err == gorm.ErrRecordNotFound {
		ErrIllegal(ctx, "用户不存在")
		return
	}
	if err != nil {
		ErrSys(ctx, err)
		return
	}
	info.UserName = user.Name

	// 获取文件夹信息
	info.FolderName, err = repo.FolderRepo.GetFolderFullPath(noteMember.FolderId, claims.Sub, "")
	if err != nil {
		ErrSys(ctx, err)
		return
	}

	ctx.JSON(200, &info)
}

/**
@api {GET} /api/note/noteList 笔记列表
@apiDescription 根据条件获取用户笔记列表

@apiName NoteNoteList
@apiGroup Note

@apiPermission 用户、管理员

@apiParam {Integer} role 权限
@apiParam {String} keyword 查询关键字
@apiParam {Integer} page 页数
@apiParam {Integer} isDelete 是否删除
@apiParam {Integer} group 用户组ID
@apiParam {Integer} folder 文件夹ID

@apiParamExample {http} 请求示例

GET /api/note/noteList?role=255&keyword=1&page=1&isDelete=0&group=1

@apiSuccess {NoteInfo[]} records 查询结果列表。
@apiSuccess {Integer} total 记录总数。
@apiSuccess {Integer} size 每页显示条数，默认 20。
@apiSuccess {Integer} current 当前页。
@apiSuccess {Integer} pages 总页数。

@apiSuccess {Object} NoteInfo 笔记信息
@apiSuccess {Integer} NoteInfo.id 笔记ID。
@apiSuccess {String} NoteInfo.title 标题。
@apiSuccess {String} NoteInfo.remark 备注。
@apiSuccess {String} NoteInfo.username 笔记拥有者名称。
@apiSuccess {Integer} NoteInfo.role 权限
@apiSuccess {String} NoteInfo.updatedAt 更新时间
@apiSuccess {Integer} NoteInfo.folderId 文件夹ID

@apiSuccessExample 成功响应
HTTP/1.1 200 OK

[
    {
	"records": [{
		"id": 3,
		"title": "测试文档",
		"remark": "",
		"username": "王沁涛",
		"role": 0,
		"updatedAt": "2023-03-22 16:06:05",
		"folderId":0
	}, {
		"id": 7,
		"title": "标签测试",
		"remark": "",
		"username": "王沁涛",
		"role": 0,
		"updatedAt": "2023-03-21 14:24:12",
		"folderId":0
	}],
	"total": 2,
	"size": 13,
	"current": 1,
	"pages": 1
}
]

@apiErrorExample 失败响应
HTTP/1.1 400 Bad Request

参数非法，无法解析

*/

// noteList 获取笔记列表
func (c *NoteController) noteList(ctx *gin.Context) {
	keyword := ctx.Query("keyword")
	// 笔记权限
	role, _ := strconv.Atoi(ctx.DefaultQuery("role", "255"))
	// 文件夹ID
	folder, _ := strconv.Atoi(ctx.DefaultQuery("folder", "0"))
	// 用户组ID
	group, _ := strconv.Atoi(ctx.DefaultQuery("group", "0"))
	// 是否删除
	isDelete, _ := strconv.Atoi(ctx.DefaultQuery("isDelete", "0"))
	// 页面
	page, err := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	if err != nil {
		ErrIllegal(ctx, "参数非法，无法解析")
		return
	}
	// 每页页数
	limit, err := strconv.Atoi(ctx.DefaultQuery("limit", "13"))
	if err != nil {
		ErrIllegal(ctx, "参数非法，无法解析")
		return
	}
	// 获取用户信息
	claimsValue, _ := ctx.Get(middle.FlagClaims)
	claims := claimsValue.(*jwt.Claims)

	query, tx := repo.NewPageQueryFnc(repo.DBDao, &entity.NoteMember{}, page, limit, func(db *gorm.DB) *gorm.DB {
		// SELECT note_members.id AS id , notes.updated_at , notes.title , note_members.remark, users.`name` AS username,
		//	notes.tags ,note_members.role   FROM note_members
		//	LEFT JOIN notes ON notes.id = note_members.note_id
		//	RIGHT JOIN users on users.id = note_members.user_id
		db = db.Table("note_members").
			Select("notes.id AS id ,notes.updated_at,notes.title,note_members.remark,users.`name` AS username, note_members.role , note_members.folder_id").
			Joins("LEFT JOIN notes ON notes.id = note_members.note_id").Joins("RIGHT JOIN users on users.id = note_members.user_id")

		// 前端数据展示排序
		db = db.Order("updated_at desc")

		// 用户类型
		if claims.Type == "user" {
			db = db.Where("note_members.user_id = ? ", claims.Sub)
		} else if claims.Type == "admin" {
			db = db.Where("note_members.role = 0 AND notes.is_delete = 1")
			return db
		}

		// 关键字查询 - 标题、标题拼音
		if keyword != "" {
			db = db.Where("notes.title like ? OR notes.title_py like ?", fmt.Sprintf("%%%s%%", keyword), fmt.Sprintf("%%%s%%", keyword))
		}

		// 是否删除
		if isDelete != 0 {
			db = db.Where("note_members.user_id = ? AND notes.user_id = ? AND notes.is_delete = ?", claims.Sub, claims.Sub, isDelete)
			return db
		} else {
			db = db.Where("notes.is_delete = 0")
		}

		// 用户组
		if group != 0 {
			db = db.Where("note_members.group_id = ?", group)
		}

		// 用户权限
		if role == 254 {
			db = db.Where("note_members.role = 1 OR note_members.role = 2")
		} else if role != 255 {
			db = db.Where("note_members.role = ? ", role)
		}

		// 文件夹
		if folder != 0 {
			db = db.Where("note_members.folder_id = ?", folder)
		}

		return db
	})
	noteList := []dto.NoteListDto{}
	if err := tx.Find(&noteList).Error; err != nil {
		ErrSys(ctx, err)
		return
	}

	query.Records = noteList
	ctx.JSON(200, query)
}

/**
@api {POST} /api/note/content 更新文档内容
@apiDescription 更新文档内容

注意文档内容仅由项目成员可以更新，非项目成员返回无权限访问错误。

@apiName NoteContentPOST
@apiGroup Note

@apiPermission 用户


@apiParam {Integer} docId 文档ID。
@apiParam {Integer} projectId 项目ID。
@apiParam {String} title 文档名
@apiParam {String} [query] 更新类型 - file 上传文档 - title 仅修改文档名称
<ul>
    <li>file</li>
    <li>title</li>
</ul>
@apiParam {String} content 文档内容，当文档类型为markdown时使用该字段更新文档。
@apiParam {Boolean} autoSave 是否自动保存

@apiParamExample {json} 请求示例
{
    "docId":1,
	"projectId":5,
	"title":"测试",
	"query":"file",
	"content":"te...ng",
	"autoSave":false
}


@apiSuccess {Integer} body 文档记录ID。


@apiSuccessExample 成功响应
HTTP/1.1 200 OK
13

@apiErrorExample 失败响应
HTTP/1.1 400 Bad Request

参数非法，无法解析
*/

// contentPost 更新笔记内容
func (c *NoteController) contentPost(ctx *gin.Context) {
	var note entity.Note

	id := ctx.PostForm("id")
	noteId, _ := strconv.Atoi(id)
	// 记录日志
	applog.L(ctx, "更新笔记内容", map[string]interface{}{
		"id": id,
	})
	if noteId <= 0 {
		ErrIllegal(ctx, "参数非法，无法解析")
		return
	}

	title := ctx.PostForm("title")
	if title == "" {
		ErrIllegal(ctx, "笔记名称不可为空")
		return
	}
	remark := ctx.PostForm("remark")

	// 判断是否拥有权限
	claimsValue, _ := ctx.Get(middle.FlagClaims)
	claims := claimsValue.(*jwt.Claims)
	role, err := repo.NoteMemberRepo.Check(claims.Sub, noteId)
	if err != nil {
		ErrSys(ctx, err)
		return
	}
	if role == -1 {
		ErrIllegal(ctx, "无权限")
		return
	}

	// 获取笔记信息
	err = repo.DBDao.First(&note, "id = ?", id).Error
	if err == gorm.ErrRecordNotFound {
		ErrIllegal(ctx, "该笔记不存在或被删除")
		return
	}
	if err != nil {
		ErrSys(ctx, err)
		return
	}

	// 查看用户仅可修改备注
	if role == 1 {
		err = repo.DBDao.Where("note_id", note.ID).Where("user_id", claims.Sub).Find(&entity.NoteMember{}).Update("remark", remark).Error
		if err != nil {
			ErrSys(ctx, err)
			return
		}
		return
	}

	// 获取修改后的内容信息
	content := ctx.PostForm("content")
	// 判断是否为自动保存
	autoSave, err := strconv.ParseBool(ctx.PostForm("autoSave"))
	if err != nil {
		ErrIllegal(ctx, "参数非法，无法解析")
		return
	}
	var user entity.User
	// 判断用户是否拥有读写锁
	// 获取拥有该锁的用户ID
	v := editLock.Query(id)
	// 若非自动保存
	if autoSave == false {
		// 处理 同一笔记 同一用户 在多处打开 问题
		if v.UserId == claims.Sub && v.Exp != claims.Exp {
			ErrIllegal(ctx, fmt.Sprintf("该笔记已在别处打开"))
			return
		}
		// 若无人拥有该锁或拥有该锁的用户为申请者本身
		if v.UserId == claims.Sub {
			editLock.Lock(ctx, id, claims.Sub)
			defer editLock.Unlock(id)
		} else if v.UserId == middle.NoLock {
			ErrIllegal(ctx, fmt.Sprintf("无该笔记编辑锁"))
			return
		} else {
			repo.DBDao.Where("id", v.UserId).Find(&user)
			ErrIllegal(ctx, fmt.Sprintf("%s正在编辑该笔记", user.Name))
			return
		}
	} else {
		// 处理 同一笔记 同一用户 在多处打开 问题
		if v.UserId == claims.Sub && v.Exp != claims.Exp {
			ErrIllegal(ctx, fmt.Sprintf("该笔记已在别处打开"))
			return
		} else if v.UserId != claims.Sub && v.UserId != middle.NoLock {
			repo.DBDao.Where("id", v.UserId).Find(&user)
			ErrIllegal(ctx, fmt.Sprintf("%s正在编辑该笔记", user.Name))
			return
		} else if v.UserId == middle.NoLock {
			ErrIllegal(ctx, fmt.Sprintf("无该笔记编辑锁"))
			return
		}

	}

	// 打开一个存在的文件，将原来的内容覆盖掉
	filename := filepath.Join(dir.NoteDir, id, note.Filename)
	// O_WRONLY: 只写, O_TRUNC: 清空文件
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		ErrIllegal(ctx, "文件打开错误")
		return
	}

	defer file.Close()

	// 防止删除文件本身
	fileContent := fmt.Sprintf("(%s &file=%s)", content, note.Filename)

	// 手动保存情况下，删除文件中未被引用的笔记资源
	if autoSave == false {
		err = reuint.DeleteUnreferencedFiles(fileContent, filepath.Join(dir.NoteDir, id))
		if err != nil {
			ErrSys(ctx, err)
			return
		}
	}

	// 带缓冲区的*Writer
	writer := bufio.NewWriter(file)
	_, err = writer.WriteString(content)
	if err != nil {
		ErrSys(ctx, err)
		return
	}

	// 将缓冲区中的内容写入到文件里
	err = writer.Flush()
	if err != nil {
		ErrSys(ctx, err)
		return
	}

	// 若笔记名称发生变化
	if note.Title != title {
		// 若该用户笔记名已存在，则生成随机数在笔记名后
		err = repo.DBDao.First(&entity.Note{}, "title = ? AND user_id = ? ", title, claims.Sub).Error
		if err == gorm.ErrRecordNotFound {
			note.Title = title
		} else if err != nil {
			ErrSys(ctx, err)
			return
		} else if err == nil {
			// 生成随机数
			r := fmt.Sprintf("%04v", rand.New(rand.NewSource(time.Now().UnixNano())).Int31n(10000))
			title = fmt.Sprintf("%s%s", title, r)
		}

		err = repo.DBDao.Where("id", note.ID).Find(&entity.Note{}).Update("title", title).Error
		if err != nil {
			ErrSys(ctx, err)
			return
		}
	}

	err = repo.DBDao.Where("note_id", note.ID).Where("user_id", claims.Sub).Find(&entity.NoteMember{}).Update("remark", remark).Error
	if err != nil {
		ErrSys(ctx, err)
		return
	}

}

/**
@api {Get} /api/note/content 获取文档内容
@apiDescription 获取文档内容

@apiName NoteContentGET
@apiGroup Note

@apiPermission 用户


@apiParam {Integer} id 文档ID。


@apiParamExample {json} 请求示例

/api/note/content?id=5

@apiSuccess {String} content 文档内容。


@apiSuccessExample 成功响应
HTTP/1.1 200 OK

功能....请参考上述内容。

@apiErrorExample 失败响应
HTTP/1.1 400 Bad Request

参数非法，无法解析
*/

// contentGet 获取笔记内容
func (c *NoteController) contentGet(ctx *gin.Context) {

	id, _ := strconv.Atoi(ctx.Query("id"))
	if id <= 0 {
		ErrIllegal(ctx, "参数非法，无法解析")
		return
	}

	// 获取用户信息
	claimsValue, _ := ctx.Get(middle.FlagClaims)
	claims := claimsValue.(*jwt.Claims)

	// 获取用户笔记权限
	role, err := repo.NoteMemberRepo.Check(claims.Sub, id)
	if err != nil {
		ErrSys(ctx, err)
		return
	}
	if role == -1 {
		ErrIllegal(ctx, "无权限")
		return
	}

	// 查询笔记信息
	var note entity.Note
	err = repo.DBDao.First(&note, "id = ? ", id).Error
	if err == gorm.ErrRecordNotFound {
		ErrIllegal(ctx, "笔记不存在")
		return
	} else if err != nil {
		ErrSys(ctx, err)
		return
	}

	filePath := filepath.Join(dir.NoteDir, strconv.Itoa(id), note.Filename)

	// 读取文件内容
	file, err := os.Open(filePath)
	if err != nil {
		ErrSys(ctx, err)
		return
	}
	defer file.Close()
	content, err := io.ReadAll(file)
	if err != nil {
		ErrSys(ctx, err)
		return
	}
	ctx.JSON(200, string(content))

}

/**
@api {POST} /api/note/assert 上传文档资源
@apiDescription 以表单的方式上传文档相关的图片或附件，文件上传后台需要判断文档类型是图片还是附件类型。

注意文档资源仅由项目成员可以使用，非项目成员返回无权限访问错误。

文件上传后返回资源的类型以及，资源访问的路径。

资源访问路径为文档资源的下载接口，格式为: "/api/note/assert?id=xxxx&file=202210131620160001.png"

资源文件名格式为：`YYYYMMDDHHmmss` + `4位随机数`

@apiName NoteAssertPOST
@apiGroup Note

@apiPermission 用户

@apiHeader {String} Content-type multipart/form-data 多类型表单固定值。

@apiParam {Integer} id 文档ID。
@apiParam {File} file 资源文件，文档相关的图片或文件附件。

@apiSuccess {String} type 资源类型
<ul>
	<li>image</li>
	<li>file</li>
</ul>
@apiSuccess {String} uri 资源访问地址，资源访问路径为文档资源的下载接口，格式为: "/api/note/assert?id=xxxx&file=202210131620160001.png"

@apiSuccessExample {json} 成功响应
HTTP/1.1 200 OK

	{
	    "type": "image",
	    "uri": "/api/note/assert?id=xxxx&file=202210131620160001.png"
	}

@apiErrorExample 失败响应
HTTP/1.1 400 Bad Request

参数非法，无法解析
*/

// assertPost 上传笔记资源
func (c *NoteController) assertPost(ctx *gin.Context) {
	var noteUri dto.NoteUriDto

	// 获取表单的id
	id := ctx.PostForm("id")
	// 记录日志
	applog.L(ctx, "上传笔记资源", map[string]interface{}{
		"id": id,
	})

	// 获取用户信息
	claimsValue, _ := ctx.Get(middle.FlagClaims)
	claims := claimsValue.(*jwt.Claims)

	noteId, _ := strconv.Atoi(id)
	role, err := repo.NoteMemberRepo.Check(claims.Sub, noteId)
	if err != nil {
		ErrSys(ctx, err)
		return
	}
	// 若用户无权限
	if role != 0 && role != 2 {
		ErrIllegal(ctx, "无权限")
		return
	}

	// 获取表单的文件
	file, err := ctx.FormFile("file")
	if err != nil {
		ErrIllegal(ctx, "参数非法，无法解析")
		return
	}

	// 获取文件类型
	contentType := file.Header.Get("Content-Type")
	if strings.Contains(contentType, "image") {
		noteUri.DocType = "image"
	} else {
		noteUri.DocType = "file"
	}

	// 生成随机文件名称
	filename := reuint.GenTimeFileName(file.Filename)

	filePath := filepath.Join(dir.NoteDir, id, filename)
	err = ctx.SaveUploadedFile(file, filePath)
	if err != nil {
		ErrSys(ctx, err)
		return
	}

	uri := fmt.Sprintf("/api/note/assert?id=%s&file=%s", id, filename)
	noteUri.Uri = uri
	ctx.JSON(200, noteUri)
}

/**
@api {GET} /api/note/assert 下载文档资源
@apiDescription 下载文档相关的资源。

@apiName NoteAssertGET
@apiGroup Note

@apiPermission 用户

@apiParam {Integer} id 文档ID。
@apiParam {String} file 文件名称。

@apiParamExample {http} 请求示例

GET /api/note/assert?id=xxxx&file=202210131620160001.png

@apiErrorExample 失败响应
HTTP/1.1 400 Bad Request

权限错误
*/

// assert 下载笔记资源
func (c *NoteController) assertGet(ctx *gin.Context) {

	id := ctx.Query("id")

	// 去除转义字符
	if len(id) > 0 && id[len(id)-1] == '\\' {
		id = id[:len(id)-1]
	}

	filename := ctx.Query("file")

	// 文件路径
	filePath := filepath.Join(dir.NoteDir, id, filename)

	// 防止用户通过 ../../ 的方式下载到操作系统内的重要文件
	if !strings.HasPrefix(filePath, filepath.Join(dir.NoteDir, id)) {
		ErrIllegal(ctx, "文件路径错误")
		return
	}

	file, err := os.Open(filePath)
	if err != nil {
		ErrIllegal(ctx, "文件解析失败")
		return
	}
	defer file.Close()

	// 下载文件名称
	ctx.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))

	fileNameWithSuffix := path.Base(filename)
	//获取文件的后缀(文件类型)
	fileType := path.Ext(fileNameWithSuffix)

	ctx.Header("Content-Type", reuint.GetMIME(fileType))

	_, err = io.Copy(ctx.Writer, file)
	if err != nil {
		ErrSys(ctx, err)
		return
	}
}

/**
@api {POST} /api/note/lock 获取文档编辑锁
@apiDescription 获取文档编辑锁
@apiName NoteLock
@apiGroup Note

@apiPermission 用户

@apiParam {Integer} userId 用户ID
@apiParam {Integer} id        文档ID

@apiParamExample {json} 请求示例

	{
	    "userId": 2,
		"id":3,
	}

@apiSuccessExample 成功响应
HTTP/1.1 200 OK

@apiErrorExample 失败响应
HTTP/1.1 400 Bad Request

权限错误
*/

// lock 获取笔记编辑锁
func (c *NoteController) lock(ctx *gin.Context) {
	var lockDto dto.LockDto
	var user entity.User

	err := ctx.BindJSON(&lockDto)
	// 记录日志
	applog.L(ctx, "获取笔记编辑锁", map[string]interface{}{
		"userId": lockDto.UserId,
		"id":     lockDto.Id,
	})
	if err != nil {
		ErrIllegal(ctx, "参数解析错误")
		return
	}

	// 判断用户是否 拥有笔记权限
	id, _ := strconv.Atoi(lockDto.Id)
	role, err := repo.NoteMemberRepo.Check(lockDto.UserId, id)
	if err != nil {
		ErrSys(ctx, err)
		return
	}
	// 若用户无权限
	if role != 0 && role != 2 {
		ErrIllegal(ctx, "无权限")
		return
	}

	// 获取锁信息
	v := editLock.Query(lockDto.Id)
	// 若无人拥有该锁或拥有该锁的用户为申请者本身
	if v.UserId == middle.NoLock || v.UserId == lockDto.UserId {
		editLock.Lock(ctx, lockDto.Id, lockDto.UserId)
		ctx.JSON(200, "获取到锁")
	} else {
		repo.DBDao.Where("id", v.UserId).Find(&user)
		ErrIllegal(ctx, fmt.Sprintf("%s正在编辑该笔记", user.Name))
		return
	}
}

/**
@api {POST} /api/note/cancel 取消编辑
@apiDescription 取消编辑
@apiName NoteCancel
@apiGroup Note

@apiPermission 用户

@apiParam {Integer} userId 用户ID
@apiParam {Integer} id     文档ID

@apiParamExample {json} 请求示例

	{
	    "userId": 2,
		"id":3,
	}

@apiSuccessExample 成功响应
HTTP/1.1 200 OK

@apiErrorExample 失败响应
HTTP/1.1 400 Bad Request

参数解析错误
*/

// cancel 取消编辑
func (c *NoteController) cancel(ctx *gin.Context) {
	var lockDto dto.LockDto
	var note entity.Note

	err := ctx.BindJSON(&lockDto)
	// 记录日志
	applog.L(ctx, "笔记取消编辑", map[string]interface{}{
		"userId": lockDto.UserId,
		"id":     lockDto.Id,
	})

	if err != nil {
		ErrIllegal(ctx, "参数解析错误")
		return
	}

	// 获取用户信息
	claimsValue, _ := ctx.Get(middle.FlagClaims)
	claims := claimsValue.(*jwt.Claims)

	// 获取锁信息
	v := editLock.Query(lockDto.Id)
	if v.UserId == lockDto.UserId && v.Exp == claims.Exp {
		editLock.Unlock(lockDto.Id)
	} else if (v.UserId == lockDto.UserId && v.Exp != claims.Exp) || v.UserId == middle.NoLock {
	} else {
		ErrIllegal(ctx, "操作异常")
		return
	}

	err = repo.DBDao.Where("id", lockDto.Id).Find(&note).Error
	if err != nil {
		ErrSys(ctx, err)
		return
	}

	filePath := filepath.Join(dir.NoteDir, lockDto.Id, note.Filename)
	file, err := os.Open(filePath)
	if err != nil {
		ErrSys(ctx, err)
		return
	}
	defer file.Close()
	content, err := io.ReadAll(file)
	if err != nil {
		ErrSys(ctx, err)
		return
	}

	// 防止删除文件本身
	fileContent := fmt.Sprintf("(%s &file=%s)", string(content), note.Filename)

	// 删除文件中未被引用的笔记资源
	err = reuint.DeleteUnreferencedFiles(fileContent, filepath.Join(dir.NoteDir, strconv.Itoa(note.ID)))
	if err != nil {
		ErrSys(ctx, err)
		return
	}

	ctx.JSON(200, string(content))
}

/**
@api {GET} /api/note/export 导出文档
@apiDescription 导出文档。

@apiName NoteExport
@apiGroup Note

@apiPermission 用户

@apiParam {String} id 文档ID。

@apiParamExample {http} 请求示例

GET /api/note/export?id=46

@apiErrorExample 失败响应
HTTP/1.1 400 Bad Request

权限错误
*/

// export 导出笔记
func (c *NoteController) export(ctx *gin.Context) {
	var note entity.Note
	id := ctx.Query("id")

	applog.L(ctx, "导出笔记", map[string]interface{}{
		"id": id,
	})

	// 获取用户信息
	claimsValue, _ := ctx.Get(middle.FlagClaims)
	claims := claimsValue.(*jwt.Claims)

	noteId, _ := strconv.Atoi(id)
	role, err := repo.NoteMemberRepo.Check(claims.Sub, noteId)
	if err != nil {
		ErrSys(ctx, err)
		return
	}
	// 若用户未拥有笔记编辑权限
	if role == -1 || role == 1 {
		ErrIllegal(ctx, "无权限")
		return
	}

	err = repo.DBDao.First(&note, "id = ?", id).Error
	if err != nil {
		ErrSys(ctx, err)
		return
	}

	docFilePath := filepath.Join(dir.NoteDir, id)

	// 生成临时文件夹 文件夹名称格式 文件名_更新时间
	temporaryFolderName := fmt.Sprintf("%s_%s", note.Title, note.UpdatedAt.Format("20060102"))

	temp, err := os.MkdirTemp("", temporaryFolderName)
	if err != nil {
		ErrSys(ctx, err)
		return
	}

	// 无论导出文件是否成功，都需要删除临时文件
	defer os.RemoveAll(temp)

	err = reuint.CopyTempDir(docFilePath, temp, note.Filename)
	if err != nil {
		ErrSys(ctx, err)
		return
	}
	// 读取文件
	temporaryFilename := filepath.Join(temp, "README.md")

	// 防止用户通过 ../../ 的方式下载到操作系统内的重要文件
	if !strings.HasPrefix(temporaryFilename, temp) {
		ErrIllegal(ctx, "文件路径错误")
		return
	}

	file, err := os.Open(temporaryFilename)
	if err != nil {
		ErrSys(ctx, err)
		return
	}
	content, err := io.ReadAll(file)
	if err != nil {
		ErrSys(ctx, err)
		return
	}
	file.Close()

	// 修改文件中的资源访问路径
	file, err = os.OpenFile(temporaryFilename, os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		ErrSys(ctx, err)
		return
	}
	reg := regexp.MustCompile("\\/api\\/note\\/assert\\?id=[0-9]+(\\\\)?&file=")
	results := reg.ReplaceAllString(string(content), "./assert/")

	// 将修改后的内容写入文件

	// 带缓冲区的*Writer
	writer := bufio.NewWriter(file)
	_, err = writer.WriteString(results)
	if err != nil {
		ErrSys(ctx, err)
		return
	}

	// 将缓冲区中的内容写入到文件里
	err = writer.Flush()
	if err != nil {
		ErrSys(ctx, err)
		return
	}
	file.Close()

	//打包压缩下载
	ctx.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s.zip", temporaryFolderName))
	ctx.Header("Content-Type", "application/zip")

	err = reuint.Zip(ctx.Writer, temp)
	if err != nil {
		ErrSys(ctx, err)
		return
	}
}

/**
@api {DELETE} /api/note/delete 删除文档
@apiDescription doc，删除文档。
该接口仅在数据库操作异常时返回500系统错误的状态码，其他情况均返回200。
@apiName NoteDelete
@apiGroup Note

@apiPermission 用户

@apiParam {String} id 文档ID。

@apiParamExample 请求示例
DELETE /api/note/delete?id=12

@apiSuccessExample 成功响应
HTTP/1.1 200 OK

@apiErrorExample 失败响应
HTTP/1.1 500

系统内部错误
*/

// delete 删除笔记
func (c *NoteController) delete(ctx *gin.Context) {

	id := ctx.Query("id")
	// 记录日志
	applog.L(ctx, "删除笔记", map[string]interface{}{
		"id": id,
	})
	// 获取用户信息
	claimsValue, _ := ctx.Get(middle.FlagClaims)
	claims := claimsValue.(*jwt.Claims)

	noteId, _ := strconv.Atoi(id)
	if noteId <= 0 {
		ErrIllegal(ctx, "参数解析错误")
		return
	}
	role, err := repo.NoteMemberRepo.Check(claims.Sub, noteId)
	if err != nil {
		ErrSys(ctx, err)
		return
	}
	// 若用户不是笔记拥有者
	if role != 0 {
		ErrIllegal(ctx, "无权限")
		return
	}

	// 将 笔记 记录 中 is_delete 字段 修改为 1
	err = repo.DBDao.Where("id", noteId).Find(&entity.Note{}).Update("is_delete", 1).Error
	if err != nil {
		ErrSys(ctx, err)
		return
	}

}

/**
@api {GET} /api/note/restore 恢复文档
@apiDescription 恢复文档。
该接口仅在数据库操作异常时返回500系统错误的状态码，其他情况均返回200。
@apiName NoteRestore
@apiGroup Note

@apiPermission 用户

@apiParam {String} id 文档ID。

@apiParamExample 请求示例
DELETE /api/note/restore?id=12

@apiSuccessExample 成功响应
HTTP/1.1 200 OK

@apiErrorExample 失败响应
HTTP/1.1 500

系统内部错误
*/

// restore 恢复笔记
func (c *NoteController) restore(ctx *gin.Context) {

	id := ctx.Query("id")
	// 记录日志
	applog.L(ctx, "恢复笔记", map[string]interface{}{
		"id": id,
	})
	// 获取用户信息
	claimsValue, _ := ctx.Get(middle.FlagClaims)
	claims := claimsValue.(*jwt.Claims)

	noteId, _ := strconv.Atoi(id)
	if noteId <= 0 {
		ErrIllegal(ctx, "参数解析错误")
		return
	}

	if claims.Type == "user" {
		role, err := repo.NoteMemberRepo.Check(claims.Sub, noteId)
		if err != nil {
			ErrSys(ctx, err)
			return
		}
		// 若用户非该笔记拥有者
		if role != 0 {
			ErrIllegal(ctx, "无权限")
			return
		}
	}

	// 将 笔记 记录 中 is_delete 字段 修改为 1
	err := repo.DBDao.Where("id", noteId).Find(&entity.Note{}).Update("is_delete", 0).Error
	if err != nil {
		ErrSys(ctx, err)
		return
	}

}

/**
@api {GET} /api/note/group 保存笔记至文件夹
@apiDescription 保存笔记至文件夹。

@apiName NoteGroup
@apiGroup Note

@apiPermission 用户

@apiParam {Integer} id 文档ID。
@apiParam {Integer} group 组ID。

@apiParamExample 请求示例
DELETE /api/note/group?id=12&group=5

@apiSuccessExample 成功响应
HTTP/1.1 200 OK

@apiErrorExample 失败响应
HTTP/1.1 500

系统内部错误
*/

// group 保存笔记至文件夹
func (c *NoteController) group(ctx *gin.Context) {
	id := ctx.Query("id")
	noteGroup := ctx.Query("group")
	// 获取用户信息
	claimsValue, _ := ctx.Get(middle.FlagClaims)
	claims := claimsValue.(*jwt.Claims)

	noteId, _ := strconv.Atoi(id)
	groupId, _ := strconv.Atoi(noteGroup)

	if noteId <= 0 || groupId < 0 {
		ErrIllegal(ctx, "参数解析错误")
		return
	}
	// 修改笔记成员组 文件ID
	err := repo.DBDao.Where("user_id = ? AND note_id = ?", claims.Sub, noteId).Find(&entity.NoteMember{}).Update("folder_id", groupId).Error
	if err != nil {
		ErrSys(ctx, err)
		return
	}
}

/**
@api {POST} /api/note/rename 笔记重命名
@apiDescription 笔记重命名
@apiName NoteRename
@apiGroup Note

@apiPermission 用户

@apiParam {Integer} id        文档ID
@apiParam {String} title 文档名称

@apiParamExample {json} 请求示例

	{
	    "id": 2,
		"title":“重命名”,
	}

@apiSuccessExample 成功响应
HTTP/1.1 200 OK

@apiErrorExample 失败响应
HTTP/1.1 400 Bad Request

权限错误
*/

// rename 文件重命名
func (c *NoteController) rename(ctx *gin.Context) {

	var noteInfo dto.NoteRenameDto

	err := ctx.BindJSON(&noteInfo)
	if err != nil {
		ErrIllegal(ctx, "参数解析错误")
		return
	}

	if noteInfo.Title == "" {
		ErrIllegal(ctx, "文档名称不可为空")
		return
	}

	// 判断是否拥有权限
	claimsValue, _ := ctx.Get(middle.FlagClaims)
	claims := claimsValue.(*jwt.Claims)
	role, err := repo.NoteMemberRepo.Check(claims.Sub, noteInfo.ID)
	if err != nil {
		ErrSys(ctx, err)
		return
	}
	if role == -1 {
		ErrIllegal(ctx, "无权限")
		return
	}

	var note entity.Note

	// 获取笔记信息
	err = repo.DBDao.First(&note, "id = ?", noteInfo.ID).Error
	if err == gorm.ErrRecordNotFound {
		ErrIllegal(ctx, "该笔记不存在或被删除")
		return
	}
	if err != nil {
		ErrSys(ctx, err)
		return
	}

	// 若用户仅可修改备注
	if role != 0 {
		err = repo.DBDao.Where("note_id", note.ID).Where("user_id", claims.Sub).Find(&entity.NoteMember{}).Update("remark", noteInfo.Title).Error
		if err != nil {
			ErrSys(ctx, err)
			return
		}
	} else {
		// 若该用户笔记名已存在，则生成随机数在笔记名后
		err = repo.DBDao.First(&entity.Note{}, "title = ? AND user_id = ? ", noteInfo.Title, claims.Sub).Error
		if err == gorm.ErrRecordNotFound {
			note.Title = noteInfo.Title
		} else if err != nil {
			ErrSys(ctx, err)
			return
		} else if err == nil {
			// 生成随机数
			r := fmt.Sprintf("%04v", rand.New(rand.NewSource(time.Now().UnixNano())).Int31n(10000))
			note.Title = fmt.Sprintf("%s%s", noteInfo.Title, r)
		}

		err = repo.DBDao.Save(note).Error
		if err != nil {
			ErrSys(ctx, err)
			return
		}
	}
}
