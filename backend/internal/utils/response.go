package utils

import (
	"net/http"
	"reflect"
	"strconv"

	"github.com/gin-gonic/gin"
)

// Response standard JSON response structure.
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Success returns a successful response, converting nil slices/maps to [] so
// the JSON payload is [] instead of null (avoids frontend .length crashes).
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: "success",
		Data:    normalizeData(data),
	})
}

// SuccessWithMessage returns a successful response with a custom message.
func SuccessWithMessage(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: message,
		Data:    normalizeData(data),
	})
}

// Created returns a 201 response.
func Created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, Response{
		Code:    0,
		Message: "created",
		Data:    normalizeData(data),
	})
}

// normalizeData converts a nil slice/map into its empty (non-nil) form so
// that JSON encoding yields []/{} instead of null.
func normalizeData(data interface{}) interface{} {
	if data == nil {
		return []interface{}{}
	}
	v := reflect.ValueOf(data)
	switch v.Kind() {
	case reflect.Slice, reflect.Map:
		if v.IsNil() {
			return []interface{}{}
		}
	}
	return data
}

// Error returns an error response.
func Error(c *gin.Context, httpStatus int, message string) {
	c.JSON(httpStatus, Response{
		Code:    httpStatus,
		Message: message,
	})
}

// BadRequest returns a 400 error.
func BadRequest(c *gin.Context, message string) {
	Error(c, http.StatusBadRequest, message)
}

// NotFound returns a 404 error.
func NotFound(c *gin.Context, message string) {
	Error(c, http.StatusNotFound, message)
}

// InternalError returns a 500 error.
func InternalError(c *gin.Context, message string) {
	Error(c, http.StatusInternalServerError, message)
}

// Unauthorized returns a 401 error.
func Unauthorized(c *gin.Context, message string) {
	Error(c, http.StatusUnauthorized, message)
}

// Forbidden returns a 403 error.
func Forbidden(c *gin.Context, message string) {
	Error(c, http.StatusForbidden, message)
}

func GetPagination(c *gin.Context) (page, pageSize, offset int) {
	page = 1
	pageSize = 20
	if p := c.Query("page"); p != "" {
		if v, err := strconv.Atoi(p); err == nil && v > 0 {
			page = v
		}
	}
	if ps := c.Query("page_size"); ps != "" {
		if v, err := strconv.Atoi(ps); err == nil && v > 0 && v <= 500 {
			pageSize = v
		}
	}
	offset = (page - 1) * pageSize
	return
}

type PageResult struct {
	Items    interface{} `json:"items"`
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"page_size"`
}

func SuccessPage(c *gin.Context, items interface{}, total int64, page, pageSize int) {
	Success(c, PageResult{
		Items:    normalizeData(items),
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	})
}
