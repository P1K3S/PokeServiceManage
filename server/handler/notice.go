package handler

import (
	"net/http"
	"service-manage/model"
	"strconv"

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
	if err := h.DB.Where("status = 1").Order("pinned DESC, sort_order DESC, id DESC").Find(&notices).Error; err != nil {
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
		Pinned  bool   `json:"pinned"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	var maxOrder int
	h.DB.Model(&model.Notice{}).Select("COALESCE(MAX(sort_order), 0)").Scan(&maxOrder)

	notice := model.Notice{
		Title:     req.Title,
		Content:   req.Content,
		Status:    1,
		SortOrder: maxOrder + 1,
		Pinned:    req.Pinned,
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
		Pinned  *bool  `json:"pinned"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	notice.Title = req.Title
	notice.Content = req.Content
	if req.Pinned != nil {
		notice.Pinned = *req.Pinned
	}
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

func (h *NoticeHandler) TogglePin(c *gin.Context) {
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

	notice.Pinned = !notice.Pinned
	if err := h.DB.Save(&notice).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "操作失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "操作成功", "data": notice})
}

func (h *NoticeHandler) MoveNotice(c *gin.Context) {
	role, _ := c.Get("role")
	if role != model.RoleSuperAdmin {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "无权限操作"})
		return
	}

	id, _ := strconv.Atoi(c.Param("id"))
	direction := c.Param("direction")
	if direction != "up" && direction != "down" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "方向参数错误"})
		return
	}

	var notice model.Notice
	if err := h.DB.First(&notice, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "通知不存在"})
		return
	}

	var allNotices []model.Notice
	h.DB.Where("status = 1 AND pinned = ?", notice.Pinned).Order("sort_order DESC, id DESC").Find(&allNotices)

	idx := -1
	for i, n := range allNotices {
		if n.ID == uint(id) {
			idx = i
			break
		}
	}
	if idx < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "通知未找到"})
		return
	}

	if direction == "up" && idx == 0 {
		c.JSON(http.StatusOK, gin.H{"code": 200, "message": "已在最前"})
		return
	}
	if direction == "down" && idx == len(allNotices)-1 {
		c.JSON(http.StatusOK, gin.H{"code": 200, "message": "已在最后"})
		return
	}

	swapIdx := idx - 1
	if direction == "down" {
		swapIdx = idx + 1
	}

	h.DB.Model(&model.Notice{}).Where("id = ?", allNotices[idx].ID).Update("sort_order", allNotices[swapIdx].SortOrder)
	h.DB.Model(&model.Notice{}).Where("id = ?", allNotices[swapIdx].ID).Update("sort_order", allNotices[idx].SortOrder)

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "移动成功"})
}
