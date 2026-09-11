package handler

import (
	"lims-backend/internal/middleware"
	"lims-backend/internal/utils"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type GenericHandler[T any] struct {
	Logger *zap.Logger
	DB     *gorm.DB
}

func NewGenericHandler[T any](logger *zap.Logger, db *gorm.DB) *GenericHandler[T] {
	return &GenericHandler[T]{Logger: logger, DB: db}
}

func (h *GenericHandler[T]) GetDB(c *gin.Context) *gorm.DB {
	if db := middleware.GetDB(c); db != nil {
		return db
	}
	return h.DB
}

type ListScope func(db *gorm.DB, c *gin.Context) *gorm.DB

func (h *GenericHandler[T]) List(c *gin.Context, keywordFields []string, scope ListScope) {
	page, pageSize, offset := utils.GetPagination(c)
	var items []T
	q := h.GetDB(c).Order("id DESC")

	if len(keywordFields) > 0 {
		if kw := c.Query("keyword"); kw != "" {
			for i, f := range keywordFields {
				if i == 0 {
					q = q.Where(f+" LIKE ?", "%"+kw+"%")
				} else {
					q = q.Or(f+" LIKE ?", "%"+kw+"%")
				}
			}
		}
	}

	if scope != nil {
		q = scope(q, c)
	}

	var total int64
	if err := q.Model(new(T)).Count(&total).Error; err != nil {
		utils.InternalError(c, "查询失败")
		return
	}
	if err := q.Limit(pageSize).Offset(offset).Find(&items).Error; err != nil {
		utils.InternalError(c, "查询失败: "+err.Error())
		return
	}
	utils.SuccessPage(c, items, total, page, pageSize)
}

func (h *GenericHandler[T]) Get(c *gin.Context, preloads ...string) {
	id, err := parseUint(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的ID")
		return
	}
	var item T
	q := h.GetDB(c)
	for _, p := range preloads {
		q = q.Preload(p)
	}
	if err := q.First(&item, id).Error; err != nil {
		utils.NotFound(c, "记录不存在")
		return
	}
	utils.Success(c, item)
}

func (h *GenericHandler[T]) Create(c *gin.Context, bindFn func(c *gin.Context, item *T) error) {
	var item T
	if bindFn != nil {
		if err := bindFn(c, &item); err != nil {
			utils.BadRequest(c, err.Error())
			return
		}
	} else {
		if err := c.ShouldBindJSON(&item); err != nil {
			utils.BadRequest(c, "参数错误: "+err.Error())
			return
		}
	}
	if err := h.GetDB(c).Create(&item).Error; err != nil {
		utils.InternalError(c, "创建失败: "+err.Error())
		return
	}
	utils.Success(c, item)
}

func (h *GenericHandler[T]) Update(c *gin.Context, bindFn func(c *gin.Context, item *T) error) {
	id, err := parseUint(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的ID")
		return
	}
	var item T
	if err := h.GetDB(c).First(&item, id).Error; err != nil {
		utils.NotFound(c, "记录不存在")
		return
	}
	if bindFn != nil {
		if err := bindFn(c, &item); err != nil {
			utils.BadRequest(c, err.Error())
			return
		}
	} else {
		if err := c.ShouldBindJSON(&item); err != nil {
			utils.BadRequest(c, "参数错误: "+err.Error())
			return
		}
	}
	if err := h.GetDB(c).Save(&item).Error; err != nil {
		utils.InternalError(c, "更新失败: "+err.Error())
		return
	}
	utils.Success(c, item)
}

func (h *GenericHandler[T]) Delete(c *gin.Context) {
	id, err := parseUint(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的ID")
		return
	}
	var item T
	if err := h.GetDB(c).First(&item, id).Error; err != nil {
		utils.NotFound(c, "记录不存在")
		return
	}
	if err := h.GetDB(c).Delete(&item).Error; err != nil {
		utils.InternalError(c, "删除失败: "+err.Error())
		return
	}
	utils.Success(c, nil)
}
