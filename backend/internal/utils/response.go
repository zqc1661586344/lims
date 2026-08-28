package utils

import (
	"net/http"
	"reflect"

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