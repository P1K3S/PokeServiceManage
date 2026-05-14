package handler

import (
	"fmt"
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
	isDirectStr := c.Query("isDirect")
	statusStr := c.Query("status")

	query := h.DB.Model(&model.EgressMethod{}).
		Preload("DockerService", func(db *gorm.DB) *gorm.DB {
			return db.Unscoped()
		})

	if serviceIDStr != "" {
		serviceID, _ := strconv.Atoi(serviceIDStr)
		query = query.Where("service_id = ?", serviceID)
	}
	if isDirectStr == "true" {
		query = query.Where("is_direct = ?", true)
	} else if isDirectStr == "false" {
		query = query.Where("is_direct = ?", false)
	}
	if statusStr != "" {
		status, _ := strconv.Atoi(statusStr)
		query = query.Where("status = ?", status)
	}

	var total int64
	query.Count(&total)

	var methods []model.EgressMethod
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("public_port ASC").Find(&methods).Error; err != nil {
		jsonError(c, "查询出站方式列表失败")
		return
	}

	type EgressMethodVO struct {
		model.EgressMethod
		ServiceName       string `json:"serviceName"`
		MachineName       string `json:"machineName"`
		EgressServiceName string `json:"egressServiceName"`
	}

	var result []EgressMethodVO
	for _, m := range methods {
		serviceName := ""
		machineName := ""
		egressServiceName := ""
		if m.ServiceType == "other" {
			var otherService model.OtherService
			if err := h.DB.Unscoped().First(&otherService, m.ServiceID).Error; err == nil {
				serviceName = otherService.Name
				var machine model.Machine
				if err := h.DB.Unscoped().First(&machine, otherService.MachineID).Error; err == nil {
					machineName = machine.Name
				}
			}
		} else {
			if m.DockerService.ID != 0 {
				serviceName = m.DockerService.Name
				var machine model.Machine
				if err := h.DB.Unscoped().First(&machine, m.DockerService.MachineID).Error; err == nil {
					machineName = machine.Name
				}
			}
		}
		if m.IsDirect {
			egressServiceName = "本机直连"
		} else if m.EgressServiceID > 0 {
			var egressService model.DockerService
			if err := h.DB.Unscoped().First(&egressService, m.EgressServiceID).Error; err == nil {
				egressServiceName = egressService.Name
				var egressMachine model.Machine
				if err := h.DB.Unscoped().First(&egressMachine, egressService.MachineID).Error; err == nil {
					egressServiceName = egressService.Name + "-" + egressMachine.Name
				}
			}
		}
		result = append(result, EgressMethodVO{
			EgressMethod:      m,
			ServiceName:       serviceName,
			MachineName:       machineName,
			EgressServiceName: egressServiceName,
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

	var serviceMachineID uint
	if method.ServiceType == "other" {
		var otherService model.OtherService
		if err := h.DB.First(&otherService, method.ServiceID).Error; err != nil {
			jsonError(c, "所属服务不存在")
			return
		}
		serviceMachineID = otherService.MachineID
	} else {
		var dockerService model.DockerService
		if err := h.DB.First(&dockerService, method.ServiceID).Error; err != nil {
			jsonError(c, "所属服务不存在")
			return
		}
		serviceMachineID = dockerService.MachineID
	}

	if method.IsDirect {
		var machine model.Machine
		if err := h.DB.First(&machine, serviceMachineID).Error; err != nil {
			jsonError(c, "所属主机不存在")
			return
		}
	} else if method.EgressServiceID > 0 {
		var egressService model.DockerService
		if err := h.DB.First(&egressService, method.EgressServiceID).Error; err != nil {
			jsonError(c, "出站服务不存在")
			return
		}
		if !egressService.IsEgress {
			jsonError(c, "所选服务不是出站服务")
			return
		}
	} else {
		jsonError(c, "请选择出站服务")
		return
	}

	if err := h.DB.Create(&method).Error; err != nil {
		fmt.Printf("创建出站方式失败: %v, method: %+v\n", err, method)
		jsonError(c, "创建出站方式失败: "+err.Error())
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
