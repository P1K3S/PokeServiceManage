package handler

import (
	"strconv"

	"service-manage/model"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type OtherServiceHandler struct {
	DB *gorm.DB
}

func NewOtherServiceHandler(db *gorm.DB) *OtherServiceHandler {
	return &OtherServiceHandler{DB: db}
}

func (h *OtherServiceHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	keyword := c.Query("keyword")
	machineIDStr := c.Query("machineId")
	statusStr := c.Query("status")

	query := h.DB.Model(&model.OtherService{})

	if keyword != "" {
		query = query.Where("other_services.name LIKE ?", "%"+keyword+"%")
	}
	if machineIDStr != "" {
		machineID, _ := strconv.Atoi(machineIDStr)
		query = query.Where("machine_id = ?", machineID)
	}
	if statusStr != "" {
		status, _ := strconv.Atoi(statusStr)
		query = query.Where("status = ?", status)
	}

	var total int64
	query.Count(&total)

	var services []model.OtherService
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("id DESC").Find(&services).Error; err != nil {
		jsonError(c, "查询其他服务列表失败")
		return
	}

	type OtherServiceVO struct {
		model.OtherService
		MachineName string `json:"machineName"`
		MachineIP   string `json:"machineIp"`
	}

	var result []OtherServiceVO
	for _, s := range services {
		var machine model.Machine
		machineName := ""
		machineIP := ""
		if err := h.DB.Unscoped().First(&machine, s.MachineID).Error; err == nil {
			machineName = machine.Name
			machineIP = machine.IP
		}
		result = append(result, OtherServiceVO{
			OtherService: s,
			MachineName:  machineName,
			MachineIP:    machineIP,
		})
	}

	jsonPage(c, result, total, page, pageSize)
}

func (h *OtherServiceHandler) Create(c *gin.Context) {
	var service model.OtherService
	if err := c.ShouldBindJSON(&service); err != nil {
		jsonError(c, "请求参数错误")
		return
	}

	var machine model.Machine
	if err := h.DB.First(&machine, service.MachineID).Error; err != nil {
		jsonError(c, "所属主机不存在")
		return
	}

	if err := h.DB.Create(&service).Error; err != nil {
		jsonError(c, "创建其他服务失败")
		return
	}
	jsonSuccess(c, gin.H{"id": service.ID})
}

func (h *OtherServiceHandler) Update(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var service model.OtherService
	if err := h.DB.First(&service, id).Error; err != nil {
		jsonError(c, "其他服务不存在")
		return
	}

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		jsonError(c, "请求参数错误")
		return
	}

	updates = convertKeys(updates)
	if err := h.DB.Model(&service).Updates(updates).Error; err != nil {
		jsonError(c, "更新其他服务失败")
		return
	}
	jsonSuccess(c, nil)
}

func (h *OtherServiceHandler) Delete(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var service model.OtherService
	if err := h.DB.First(&service, id).Error; err != nil {
		jsonError(c, "其他服务不存在")
		return
	}

	tx := h.DB.Begin()

	if err := tx.Where("service_id = ? AND service_type = ?", id, "other").Delete(&model.EgressMethod{}).Error; err != nil {
		tx.Rollback()
		jsonError(c, "删除关联出站方式失败")
		return
	}

	if err := tx.Delete(&service).Error; err != nil {
		tx.Rollback()
		jsonError(c, "删除其他服务失败")
		return
	}

	tx.Commit()
	jsonSuccess(c, nil)
}
