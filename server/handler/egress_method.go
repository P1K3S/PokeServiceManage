package handler

import (
	"strconv"

	"service-manage/model"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type EgressMethodHandler struct {
	DB *gorm.DB
}

func NewEgressMethodHandler(db *gorm.DB) *EgressMethodHandler {
	return &EgressMethodHandler{DB: db}
}

func (h *EgressMethodHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	serviceIDStr := c.Query("serviceId")
	methodType := c.Query("methodType")
	statusStr := c.Query("status")

	query := h.DB.Model(&model.EgressMethod{}).
		Preload("DockerService", func(db *gorm.DB) *gorm.DB {
			return db.Unscoped()
		})

	if serviceIDStr != "" {
		serviceID, _ := strconv.Atoi(serviceIDStr)
		query = query.Where("service_id = ?", serviceID)
	}
	if methodType != "" {
		query = query.Where("method_type = ?", methodType)
	}
	if statusStr != "" {
		status, _ := strconv.Atoi(statusStr)
		query = query.Where("status = ?", status)
	}

	var total int64
	query.Count(&total)

	var methods []model.EgressMethod
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("id DESC").Find(&methods).Error; err != nil {
		jsonError(c, "查询出站方式列表失败")
		return
	}

	type EgressMethodVO struct {
		model.EgressMethod
		ServiceName string `json:"serviceName"`
		MachineName string `json:"machineName"`
	}

	var result []EgressMethodVO
	for _, m := range methods {
		serviceName := ""
		machineName := ""
		if m.DockerService.ID != 0 {
			serviceName = m.DockerService.Name
			var machine model.Machine
			if err := h.DB.Unscoped().First(&machine, m.DockerService.MachineID).Error; err == nil {
				machineName = machine.Name
			}
		}
		result = append(result, EgressMethodVO{
			EgressMethod: m,
			ServiceName:  serviceName,
			MachineName:  machineName,
		})
	}

	jsonPage(c, result, total, page, pageSize)
}

func (h *EgressMethodHandler) Create(c *gin.Context) {
	var method model.EgressMethod
	if err := c.ShouldBindJSON(&method); err != nil {
		jsonError(c, "请求参数错误")
		return
	}

	var dockerService model.DockerService
	if err := h.DB.First(&dockerService, method.ServiceID).Error; err != nil {
		jsonError(c, "所属Docker服务不存在")
		return
	}

	if err := h.DB.Create(&method).Error; err != nil {
		jsonError(c, "创建出站方式失败")
		return
	}
	jsonSuccess(c, gin.H{"id": method.ID})
}

func (h *EgressMethodHandler) Update(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var method model.EgressMethod
	if err := h.DB.First(&method, id).Error; err != nil {
		jsonError(c, "出站方式不存在")
		return
	}

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		jsonError(c, "请求参数错误")
		return
	}

	updates = convertKeys(updates)
	if err := h.DB.Model(&method).Updates(updates).Error; err != nil {
		jsonError(c, "更新出站方式失败")
		return
	}
	jsonSuccess(c, nil)
}

func (h *EgressMethodHandler) Delete(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var method model.EgressMethod
	if err := h.DB.First(&method, id).Error; err != nil {
		jsonError(c, "出站方式不存在")
		return
	}

	if err := h.DB.Delete(&method).Error; err != nil {
		jsonError(c, "删除出站方式失败")
		return
	}
	jsonSuccess(c, nil)
}
