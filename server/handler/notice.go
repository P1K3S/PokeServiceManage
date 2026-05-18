package handler

import (
	"net/http"
	"strconv"
	"service-manage/model"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type NoticeHandler struct {
	DB *gorm.DB
}

func NewNoticeHandler(db *gorm.DB) *NoticeHandler {
	return &NoticeHandler{DB: db}
}

func (h *NoticeHandler) ListNotices(c *gin.Context) {
	var notices []model.Notice
	if err := h.DB.Where("status = 1").Order("id DESC").Find(&notices).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 200, "data": []model.Notice{}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": notices})
}

func (h *NoticeHandler) CreateNotice(c *gin.Context) {
	role, _ := c.Get("role")
	if role != model.RoleSuperAdmin {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "无权限操作"})
		return
	}

	var req struct {
		Title   string `json:"title"`
		Content string `json:"content"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	notice := model.Notice{
		Title:   req.Title,
		Content: req.Content,
		Status:  1,
	}
	if err := h.DB.Create(&notice).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "创建失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "创建成功", "data": notice})
}

func (h *NoticeHandler) UpdateNotice(c *gin.Context) {
	role, _ := c.Get("role")
	if role != model.RoleSuperAdmin {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "无权限操作"})
		return
	}

	id, _ := strconv.Atoi(c.Param("id"))
	var notice model.Notice
	if err := h.DB.First(&notice, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "通知不存在"})
		return
	}

	var req struct {
		Title   string `json:"title"`
		Content string `json:"content"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	notice.Title = req.Title
	notice.Content = req.Content
	if err := h.DB.Save(&notice).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "保存失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "保存成功", "data": notice})
}

func (h *NoticeHandler) DeleteNotice(c *gin.Context) {
	role, _ := c.Get("role")
	if role != model.RoleSuperAdmin {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "无权限操作"})
		return
	}

	id, _ := strconv.Atoi(c.Param("id"))
	var notice model.Notice
	if err := h.DB.First(&notice, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "通知不存在"})
		return
	}

	if err := h.DB.Delete(&notice).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "删除失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "删除成功"})
}
