package repo

import (
	"gorm.io/gorm"
	"note/repo/entity"
)

// NewPageQuery 创建一个分页查询
// tmpDbConn: 数据库连接对象
// typ: 分页查询的实体对象指针，例如："&entity.Cert{}"
// page: 页码
// limit: 每页显示数量
// ms: 条件查询map
//
// return:
// 页面返还值对象
// 预设好的查询语句
func NewPageQuery(tmpDbConn *gorm.DB, typ interface{}, page, limit int, ms map[interface{}]interface{}) (*entity.Page, *gorm.DB) {
	// 计算分页参数
	offset := (page - 1) * limit
	var total int64

	for k, v := range ms {
		if v == "" {
			continue
		}
		tmpDbConn = tmpDbConn.Where(k, v)
	}
	tmpDbConn.Model(typ).Count(&total)

	pages := total / int64(limit)
	if total%int64(limit) > 0 {
		pages++
	}
	res := entity.Page{
		Records: []interface{}{},
		Total:   total,
		Size:    limit,
		Current: page,
		Pages:   int(pages),
	}

	// 构造查询调整
	tx := tmpDbConn.Offset(offset).Limit(limit)
	return &res, tx
}

// QueryConditionFnc 查询条件构造函数
type QueryConditionFnc func(*gorm.DB) *gorm.DB

// NewPageQueryFnc 自定义条件的分页查询
//
// tmpDbConn: 数据库连接对象
// typ: 分页查询的实体对象指针，例如："&entity.Cert{}"
// page: 页码，从1起
// limit: 每页数量
// conditionFnc: 条件构造器
//
// return: 页面返还值对象, 预设好的查询语句
func NewPageQueryFnc(tmpDbConn *gorm.DB, typ interface{}, page, limit int, conditionFnc QueryConditionFnc) (*entity.Page, *gorm.DB) {
	// 计算分页参数
	offset := (page - 1) * limit
	var total int64

	if conditionFnc != nil {
		tmpDbConn = conditionFnc(tmpDbConn)
	}
	tmpDbConn.Model(typ).Count(&total)

	pages := total / int64(limit)
	if total%int64(limit) > 0 {
		pages++
	}
	res := entity.Page{
		Records: []interface{}{},
		Total:   total,
		Size:    limit,
		Current: page,
		Pages:   int(pages),
	}

	// 构造查询调整
	tx := tmpDbConn.Offset(offset).Limit(limit)
	return &res, tx
}
