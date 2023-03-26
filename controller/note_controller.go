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
	// 更新笔记内容
	r.POST("/content", User, res.contentPost)
	// 获取笔记内容
	r.GET("/content", User, res.contentGet)
	// 上传笔记资源
	r.POST("/assert", User, res.assertPost)
	// 下载笔记资源
	r.GET("/assert", User, res.assertGet)
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

@apiHeader {String} Content-type multipart/form-data 多类型表单固定值。


@apiParam {String} title 笔记名。
@apiParam {Integer} priority 优先级，默认为0，数值越大优先级越高显示位置越靠前。
@apiParam {String} tags 标签。


@apiSuccess {Integer} id 文档记录ID。
@apiSuccess {[]String} tags 标签。

@apiSuccessExample 成功响应
HTTP/1.1 200 OK

{
	"id":1,
	"tags":"运维,常见问题,测试,分享"
}

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

	note.UserId = claims.Sub
	note.TitlePy = str
	note.Priority = noteCreateDto.Priority
	note.Tags = noteCreateDto.Tags
	note.IsDelete = 0

	result := ""
	err = repo.DBDao.Transaction(func(tx *gorm.DB) error {
		// 创建笔记记录
		err = tx.Create(&note).Error
		if err != nil {
			return err
		}
		// 创建文件目录
		notePath := filepath.Join(dir.NoteDir, strconv.Itoa(note.ID))
		err = os.MkdirAll(notePath, os.ModePerm)
		if err != nil {
			return err
		}
		// 生成文件名称
		now := time.Now().Format("20060102150405")
		r := fmt.Sprintf("%04v", rand.New(rand.NewSource(time.Now().UnixNano())).Int31n(10000))
		filename := fmt.Sprintf("%s%s.md", now, r)
		note.Filename = filename
		docFile := filepath.Join(notePath, note.Filename)
		_, err = os.Create(docFile)
		if err != nil {
			return err
		}
		find := tx.Where("id", note.ID).Find(&entity.Note{})
		err := find.Update("filename", note.Filename).Error
		if err != nil {
			return err
		}

		// 生成笔记成员记录
		noteMember := &entity.NoteMember{}
		noteMember.Role = 0
		noteMember.UserId = claims.Sub
		noteMember.NoteId = note.ID
		err = tx.Create(&noteMember).Error
		if err != nil {
			return err
		}

		// 更新标签
		if note.Tags != "" {
			var usr entity.User
			err = repo.DBDao.First(&usr, "id = ?", claims.Sub).Error
			if err != nil {
				return err
			}
			tags := strings.Split(usr.NoteTags, ",")
			noteTags := strings.Split(note.Tags, ",")
			tags = append(tags, noteTags...)

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
				return err
			}
		}

		return nil
	})
	if err != nil {
		ErrIllegalE(ctx, err)
		return
	}

	var info dto.NoteIDDto
	info.ID = note.ID
	info.Tags = strings.Split(result, ",")

	ctx.JSON(200, info)
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
	info.Tags = note.Tags
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
	info.Role = noteMember.Role
	info.Remark = noteMember.Remark
	info.Title = note.Title
	info.IsDelete = note.IsDelete

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
@apiParam {String} tag 标签

@apiParamExample {http} 请求示例

GET /api/note/noteList?role=255&keyword=1&page=1&isDelete=0&group=1&tag=%E8%BF%90%E7%BB%B4

@apiSuccess {NoteInfo[]} records 查询结果列表。
@apiSuccess {Integer} total 记录总数。
@apiSuccess {Integer} size 每页显示条数，默认 20。
@apiSuccess {Integer} current 当前页。
@apiSuccess {Integer} pages 总页数。

@apiSuccess {Object} NoteInfo 笔记信息
@apiSuccess {Integer} NoteInfo.id 笔记ID。
@apiSuccess {String} NoteInfo.title 标题。
@apiSuccess {String} NoteInfo.remark 备注。
@apiSuccess {String} NocInfo.username 笔记拥有者名称。
@apiSuccess {String} NocInfo.tags 标签
@apiSuccess {Integer} NoteInfo.role 权限
@apiSuccess {String} NocInfo.updatedAt 更新时间

@apiSuccessExample 成功响应
HTTP/1.1 200 OK

[
    {
	"records": [{
		"id": 3,
		"title": "测试文档",
		"remark": "",
		"username": "王沁涛",
		"tags": "运维",
		"role": 0,
		"updatedAt": "2023-03-22 16:06:05"
	}, {
		"id": 7,
		"title": "标签测试",
		"remark": "",
		"username": "王沁涛",
		"tags": "运维,常见问题,测试",
		"role": 0,
		"updatedAt": "2023-03-21 14:24:12"
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
	role, _ := strconv.Atoi(ctx.DefaultQuery("role", "255"))
	tag := ctx.Query("tag")
	group, _ := strconv.Atoi(ctx.DefaultQuery("group", "255"))

	isDelete, _ := strconv.Atoi(ctx.Query("isDelete"))

	page, err := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	if err != nil {
		ErrIllegal(ctx, "参数非法，无法解析")
		return
	}
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
			Select("notes.id AS id ,notes.updated_at,notes.title,note_members.remark,users.`name` AS username, notes.tags ,note_members.role ").
			Joins("LEFT JOIN notes ON notes.id = note_members.note_id").Joins("RIGHT JOIN users on users.id = note_members.user_id")

		// 前端数据展示排序
		db = db.Order("updated_at desc")

		if claims.Type == "user" {
			db = db.Where("note_members.user_id = ? ", claims.Sub)
		} else if claims.Type == "admin" {
			db = db.Where("note_members.role = 0 AND notes.is_delete = 1")
			return db
		}

		if keyword != "" {
			db = db.Where("notes.title like ? OR notes.title_py like ?", fmt.Sprintf("%%%s%%", keyword), fmt.Sprintf("%%%s%%", keyword))
		}

		if isDelete != 0 {
			db = db.Where("note_members.user_id = ? AND notes.user_id = ? AND notes.is_delete = ?", claims.Sub, claims.Sub, isDelete)
			return db
		} else {
			db = db.Where("notes.is_delete = 0")
		}

		if group != 255 {
			db = db.Where("note_members.group_id = ?", group)
		}

		if role == 254 {
			db = db.Where("note_members.role = 1 OR note_members.role = 2")
		} else if role != 255 {
			db = db.Where("note_members.role = ? ", role)
		}

		if tag != "" {
			db = db.Where("tags like ?", fmt.Sprintf("%%%s%%", tag))
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
@apiDescription 以表单的方式上传新的文档内容替换原有的文档内容。

注意文档内容仅由项目成员可以更新，非项目成员返回无权限访问错误。

@apiName NoteContentPOST
@apiGroup Note

@apiPermission 用户

@apiHeader {String} Content-type multipart/form-data 多类型表单固定值。

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

	content := ctx.PostForm("content")
	autoSave, err := strconv.ParseBool(ctx.PostForm("autoSave"))
	if err != nil {
		ErrIllegal(ctx, "参数非法，无法解析")
		return
	}
	var user entity.User
	// 判断用户是否拥有读写锁
	if autoSave == false {
		// 获取拥有该锁的用户ID
		v := editLock.Query(id)
		// 若无人拥有该锁或拥有该锁的用户为申请者本身
		if v == middle.NoLock || v == claims.Sub {
			editLock.Lock(id, claims.Sub)
			defer editLock.Unlock(id)
		} else {
			repo.DBDao.Where("id", v).Find(&user)
			ErrIllegal(ctx, fmt.Sprintf("%s正在编辑该笔记", user.Name))
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

	// 删除文件中未被引用的笔记资源
	err = reuint.DeleteUnreferencedFiles(fileContent, filepath.Join(dir.NoteDir, id))
	if err != nil {
		ErrSys(ctx, err)
		return
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

	err = repo.DBDao.Where("id", note.ID).Find(&entity.Note{}).Update("title", title).Error
	if err != nil {
		ErrSys(ctx, err)
		return
	}

	err = repo.DBDao.Where("note_id", note.ID).Where("user_id", claims.Sub).Find(&entity.NoteMember{}).Update("remark", remark).Error
	if err != nil {
		ErrSys(ctx, err)
		return
	}

}

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

	role, err := repo.NoteMemberRepo.Check(claims.Sub, id)
	if err != nil {
		ErrSys(ctx, err)
		return
	}
	if role == -1 {
		ErrIllegal(ctx, "无权限")
		return
	}

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
	if role == -1 || role == 1 {
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

// assert 下载笔记资源
func (c *NoteController) assertGet(ctx *gin.Context) {

	id := ctx.Query("id")
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

	id, _ := strconv.Atoi(lockDto.Id)
	role, err := repo.NoteMemberRepo.Check(lockDto.UserId, id)
	if err != nil {
		ErrSys(ctx, err)
		return
	}
	if role == -1 || role == 1 {
		ErrIllegal(ctx, "无权限")
		return
	}

	// 获取锁信息
	v := editLock.Query(lockDto.Id)
	// 若无人拥有该锁或拥有该锁的用户为申请者本身
	if v == middle.NoLock || v == lockDto.UserId {
		editLock.Lock(lockDto.Id, lockDto.UserId)
		ctx.JSON(200, "获取到锁")
	} else {
		repo.DBDao.Where("id", v).Find(&user)
		ErrIllegal(ctx, fmt.Sprintf("%s正在编辑该笔记", user.Name))
		return
	}
}

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

	// 获取锁信息
	v := editLock.Query(lockDto.Id)
	if v == lockDto.UserId {
		editLock.Unlock(lockDto.Id)
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

	err = reuint.CopyTempDir(docFilePath, temp)
	if err != nil {
		ErrSys(ctx, err)
		return
	}
	// 读取文件
	temporaryFilename := filepath.Join(temp, note.Filename)

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
	results := reg.ReplaceAllString(string(content), "./")

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
	if role == -1 || role == 1 || role == 2 {
		ErrIllegal(ctx, "无权限")
		return
	}

	err = repo.DBDao.Where("id", noteId).Find(&entity.Note{}).Update("is_delete", 1).Error
	if err != nil {
		ErrSys(ctx, err)
		return
	}

}

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
		if role == -1 || role == 1 || role == 2 {
			ErrIllegal(ctx, "无权限")
			return
		}
	}

	err := repo.DBDao.Where("id", noteId).Find(&entity.Note{}).Update("is_delete", 0).Error
	if err != nil {
		ErrSys(ctx, err)
		return
	}

}
