package handler

import (
	"strconv"

	"service-manage/model"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type OperationLogHandler struct {
	DB *gorm.DB
}

func NewOperationLogHandler(db *gorm.DB) *OperationLogHandler {
	return &OperationLogHandler{DB: db}
}

func logOperation(db *gorm.DB, userID uint, username, action, target string, targetID uint, detail string) {
	go func() {
		db.Create(&model.OperationLog{
			UserID:   userID,
			Username: username,
			Action:   action,
			Target:   target,
			TargetID: targetID,
			Detail:   detail,
		})
	}()
}

func (h *OperationLogHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	action := c.Query("action")
	target := c.Query("target")
	username := c.Query("username")

	query := h.DB.Model(&model.OperationLog{})

	if action != "" {
		query = query.Where("action = ?", action)
	}
	if target != "" {
		query = query.Where("target = ?", target)
	}
	if username != "" {
		query = query.Where("username LIKE ?", "%"+username+"%")
	}

	var total int64
	query.Count(&total)

	var logs []model.OperationLog
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("id DESC").Find(&logs).Error; err != nil {
		jsonError(c, "查询操作日志失败")
		return
	}

	jsonPage(c, logs, total, page, pageSize)
}
