package handler

import (
	"strings"
	"unicode"

	"github.com/gin-gonic/gin"
)

func camelToSnake(s string) string {
	var result strings.Builder
	for i, r := range s {
		if unicode.IsUpper(r) {
			if i > 0 {
				result.WriteByte('_')
			}
			result.WriteRune(unicode.ToLower(r))
		} else {
			result.WriteRune(r)
		}
	}
	return result.String()
}

func convertKeys(m map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{}, len(m))
	for k, v := range m {
		result[camelToSnake(k)] = v
	}
	return result
}

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type PageResult struct {
	List     interface{} `json:"list"`
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"pageSize"`
}

const (
	Success       = 0
	ErrBadRequest = 1001
	ErrNotFound   = 1002
	ErrDatabase   = 1003
	ErrDuplicate  = 1004
)

func jsonSuccess(c *gin.Context, data interface{}) {
	c.JSON(200, Response{
		Code:    Success,
		Message: "success",
		Data:    data,
	})
}

func jsonError(c *gin.Context, message string) {
	c.JSON(200, Response{
		Code:    ErrBadRequest,
		Message: message,
	})
}

func jsonErrorCode(c *gin.Context, code int, message string) {
	c.JSON(200, Response{
		Code:    code,
		Message: message,
	})
}

func jsonPage(c *gin.Context, list interface{}, total int64, page, pageSize int) {
	jsonSuccess(c, PageResult{
		List:     list,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	})
}

func toInt(v interface{}) int {
	switch val := v.(type) {
	case float64:
		return int(val)
	case int:
		return val
	case int64:
		return int(val)
	default:
		return 0
	}
}
