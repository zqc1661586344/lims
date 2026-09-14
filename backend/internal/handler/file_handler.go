package handler

import (
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"lims-backend/internal/middleware"
	"lims-backend/internal/model"
	"lims-backend/internal/service"
	"lims-backend/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

const maxUploadSize = 50 * 1024 * 1024

type FileHandler struct {
	DB      *gorm.DB
	Logger  *zap.Logger
	Storage service.StorageService
}

func NewFileHandler(db *gorm.DB, logger *zap.Logger, storage service.StorageService) *FileHandler {
	return &FileHandler{DB: db, Logger: logger, Storage: storage}
}

var allowedMimePrefixes = []string{
	"image/",
	"application/pdf",
	"application/msword",
	"application/vnd.openxmlformats-officedocument",
	"application/vnd.ms-excel",
	"text/",
	"application/zip",
}

var allowedExts = map[string]bool{
	".pdf": true, ".doc": true, ".docx": true,
	".xls": true, ".xlsx": true, ".csv": true,
	".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".bmp": true,
	".txt": true, ".zip": true,
}

func isAllowedFile(filename, contentType string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	if allowedExts[ext] {
		return true
	}
	for _, p := range allowedMimePrefixes {
		if strings.HasPrefix(contentType, p) {
			return true
		}
	}
	return false
}

func (h *FileHandler) Upload(c *gin.Context) {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		utils.BadRequest(c, "缺少文件字段 file")
		return
	}
	if fileHeader.Size > maxUploadSize {
		utils.BadRequest(c, "文件大小超过 50MB 限制")
		return
	}
	if !isAllowedFile(fileHeader.Filename, fileHeader.Header.Get("Content-Type")) {
		utils.BadRequest(c, "不支持的文件类型")
		return
	}

	category := c.PostForm("category")
	businessType := c.PostForm("business_type")
	businessIDStr := c.PostForm("business_id")

	var businessIDPtr *uint
	if businessIDStr != "" {
		if v, err := strconv.ParseUint(businessIDStr, 10, 64); err == nil {
			id := uint(v)
			businessIDPtr = &id
		}
	}

	objectName := fmt.Sprintf("%s/%s_%s", businessType, uuid.New().String(), filepath.Base(fileHeader.Filename))

	opened, err := fileHeader.Open()
	if err != nil {
		utils.InternalError(c, "无法读取上传文件")
		return
	}
	defer opened.Close()

	ctx := c.Request.Context()
	if err := h.Storage.Upload(ctx, objectName, opened, fileHeader.Size, fileHeader.Header.Get("Content-Type")); err != nil {
		h.Logger.Error("storage upload failed", zap.Error(err))
		utils.InternalError(c, "文件存储失败")
		return
	}

	userID := middleware.GetUserID(c)
	fm := model.File{
		ObjectName:   objectName,
		OriginalName: fileHeader.Filename,
		ContentType:  fileHeader.Header.Get("Content-Type"),
		Size:         fileHeader.Size,
		Category:     category,
		BusinessType: businessType,
		BusinessID:   businessIDPtr,
		CreatedBy:    &userID,
	}
	if err := h.DB.Create(&fm).Error; err != nil {
		h.Logger.Error("file record insert failed", zap.Error(err))
		utils.InternalError(c, "保存文件记录失败")
		return
	}

	utils.Created(c, fm)
}

func (h *FileHandler) List(c *gin.Context) {
	businessType := c.Query("business_type")
	businessIDStr := c.Query("business_id")
	category := c.Query("category")
	keyword := c.Query("keyword")
	page, pageSize, _ := utils.GetPagination(c)

	q := h.DB.Model(&model.File{})
	if businessType != "" {
		q = q.Where("business_type = ?", businessType)
	}
	if businessIDStr != "" {
		if v, err := strconv.ParseUint(businessIDStr, 10, 64); err == nil {
			q = q.Where("business_id = ?", uint(v))
		}
	}
	if category != "" {
		q = q.Where("category = ?", category)
	}
	if keyword != "" {
		q = q.Where("original_name ILIKE ?", "%"+keyword+"%")
	}

	userID := middleware.GetUserID(c)
	if !middleware.IsAdmin(c) {
		q = q.Where("created_by = ? OR (business_type = 'task_order' AND business_id IN (SELECT id FROM task_orders WHERE created_by = ?))", userID, userID)
	}

	var total int64
	q.Count(&total)

	var items []model.File
	if err := q.Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&items).Error; err != nil {
		utils.InternalError(c, err.Error())
		return
	}
	utils.SuccessPage(c, items, total, page, pageSize)
}

func (h *FileHandler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.BadRequest(c, "无效的文件ID")
		return
	}
	var fm model.File
	if err := h.DB.First(&fm, id).Error; err != nil {
		utils.NotFound(c, "文件不存在")
		return
	}
	if !middleware.IsAdmin(c) {
		userID := middleware.GetUserID(c)
		if fm.CreatedBy == nil || *fm.CreatedBy != userID {
			if fm.BusinessType == "task_order" && fm.BusinessID != nil {
				var cnt int64
				h.DB.Raw(`SELECT COUNT(*) FROM task_orders WHERE id = ? AND (created_by = ? OR id IN (
					SELECT pi.business_id FROM process_instances pi
					JOIN process_tasks pt ON pt.process_instance_id = pi.id
					WHERE pi.business_type = 'task_order' AND pt.assignee_dept_id = (SELECT dept_id FROM users WHERE id = ?)
				))`, *fm.BusinessID, userID, userID).Scan(&cnt)
				if cnt == 0 {
					utils.Forbidden(c, "无权访问该文件")
					return
				}
			} else {
				utils.Forbidden(c, "无权访问该文件")
				return
			}
		}
	}

	url, _ := h.Storage.GetPresignedURL(c.Request.Context(), fm.ObjectName, 15*time.Minute)
	type fileResp struct {
		model.File
		PresignedURL string `json:"presigned_url"`
	}
	utils.Success(c, fileResp{File: fm, PresignedURL: url})
}

func (h *FileHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.BadRequest(c, "无效的文件ID")
		return
	}
	var fm model.File
	if err := h.DB.First(&fm, id).Error; err != nil {
		utils.NotFound(c, "文件不存在")
		return
	}
	if !middleware.IsAdmin(c) {
		userID := middleware.GetUserID(c)
		if fm.CreatedBy == nil || *fm.CreatedBy != userID {
			utils.Forbidden(c, "只能删除自己上传的文件")
			return
		}
	}

	ctx := c.Request.Context()
	if err := h.Storage.Delete(ctx, fm.ObjectName); err != nil {
		h.Logger.Error("storage delete failed", zap.Error(err))
		utils.InternalError(c, "文件存储删除失败")
		return
	}
	if err := h.DB.Delete(&fm).Error; err != nil {
		utils.InternalError(c, "文件记录删除失败")
		return
	}
	utils.Success(c, nil)
}

func (h *FileHandler) Download(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.BadRequest(c, "无效的文件ID")
		return
	}
	var fm model.File
	if err := h.DB.First(&fm, id).Error; err != nil {
		utils.NotFound(c, "文件不存在")
		return
	}
	if !middleware.IsAdmin(c) {
		userID := middleware.GetUserID(c)
		if fm.CreatedBy == nil || *fm.CreatedBy != userID {
			utils.Forbidden(c, "无权访问该文件")
			return
		}
	}

	reader, info, err := h.Storage.Download(c.Request.Context(), fm.ObjectName)
	if err != nil {
		h.Logger.Error("storage download failed", zap.Error(err))
		utils.InternalError(c, "文件下载失败")
		return
	}
	defer reader.Close()

	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, fm.OriginalName))
	c.Header("Content-Type", fm.ContentType)
	if info != nil {
		c.Header("Content-Length", strconv.FormatInt(info.Size, 10))
	}
	c.Status(http.StatusOK)
	io.Copy(c.Writer, reader)
}
