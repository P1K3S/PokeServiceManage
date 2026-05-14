package handler

import (
	"fmt"
	"strconv"
	"strings"

	"service-manage/model"
	sshutil "service-manage/utils/ssh"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type DockerServiceHandler struct {
	DB *gorm.DB
}

func NewDockerServiceHandler(db *gorm.DB) *DockerServiceHandler {
	return &DockerServiceHandler{DB: db}
}

func (h *DockerServiceHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	keyword := c.Query("keyword")
	machineIDStr := c.Query("machineId")
	statusStr := c.Query("status")

	query := serviceScope(c, h.DB).Model(&model.DockerService{})

	if keyword != "" {
		query = query.Where("docker_services.name LIKE ?", "%"+keyword+"%")
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

	var services []model.DockerService
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("port ASC").Find(&services).Error; err != nil {
		jsonError(c, "查询服务列表失败")
		return
	}

	type DockerServiceVO struct {
		model.DockerService
		MachineName string `json:"machineName"`
		MachineIP   string `json:"machineIp"`
		EgressCount int64  `json:"egressCount"`
	}

	var result []DockerServiceVO
	for _, s := range services {
		var egressCount int64
		h.DB.Model(&model.EgressMethod{}).Where("service_id = ? AND service_type = ?", s.ID, "docker").Count(&egressCount)
		var machine model.Machine
		machineName := ""
		machineIP := ""
		if err := h.DB.Unscoped().First(&machine, s.MachineID).Error; err == nil {
			machineName = machine.Name
			machineIP = machine.IP
		}
		result = append(result, DockerServiceVO{
			DockerService: s,
			MachineName:   machineName,
			MachineIP:     machineIP,
			EgressCount:   egressCount,
		})
	}

	jsonPage(c, result, total, page, pageSize)
}

func (h *DockerServiceHandler) Create(c *gin.Context) {
	var service model.DockerService
	if err := c.ShouldBindJSON(&service); err != nil {
		jsonError(c, "请求参数错误")
		return
	}

	service.UserID = getUserId(c)

	var machine model.Machine
	if err := h.DB.First(&machine, service.MachineID).Error; err != nil {
		jsonError(c, "所属主机不存在")
		return
	}

	if err := h.DB.Create(&service).Error; err != nil {
		jsonError(c, "创建服务失败")
		return
	}
	jsonSuccess(c, gin.H{"id": service.ID})
}

func (h *DockerServiceHandler) Update(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var service model.DockerService
	if err := userScope(c, h.DB).First(&service, id).Error; err != nil {
		jsonError(c, "服务不存在")
		return
	}

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		jsonError(c, "请求参数错误")
		return
	}

	updates = convertKeys(updates)
	if err := h.DB.Model(&service).Updates(updates).Error; err != nil {
		jsonError(c, "更新服务失败")
		return
	}
	jsonSuccess(c, nil)
}

func (h *DockerServiceHandler) Delete(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var service model.DockerService
	if err := userScope(c, h.DB).First(&service, id).Error; err != nil {
		jsonError(c, "服务不存在")
		return
	}

	tx := h.DB.Begin()

	if err := tx.Where("service_id = ? AND service_type = ?", id, "docker").Delete(&model.EgressMethod{}).Error; err != nil {
		tx.Rollback()
		jsonError(c, "删除关联出站方式失败")
		return
	}

	if err := tx.Delete(&service).Error; err != nil {
		tx.Rollback()
		jsonError(c, "删除服务失败")
		return
	}

	tx.Commit()
	jsonSuccess(c, nil)
}

func (h *DockerServiceHandler) Check(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var service model.DockerService
	if err := userScope(c, h.DB).First(&service, id).Error; err != nil {
		jsonError(c, "服务不存在")
		return
	}
	var machine model.Machine
	if err := h.DB.First(&machine, service.MachineID).Error; err != nil {
		jsonError(c, "所属主机不存在")
		return
	}
	if machine.SSHUser == "" || machine.SSHPassword == "" {
		jsonError(c, "主机未配置 SSH 连接信息，请先在主机管理中设置")
		return
	}
	sshPort := machine.SSHPort
	if sshPort == 0 {
		sshPort = 22
	}
	cmd := fmt.Sprintf("docker ps --format '{{.Names}}' | grep -i %s", service.Name)
	output, err := sshutil.RunCommand(&sshutil.Config{
		Host:     machine.IP,
		Port:     sshPort,
		User:     machine.SSHUser,
		Password: machine.SSHPassword,
	}, cmd)
	newStatus := int8(0)
	msg := "Docker 容器未运行"
	if err == nil && strings.TrimSpace(output) != "" {
		newStatus = 1
		msg = "Docker 容器运行中"
	}
	h.DB.Model(&service).Update("status", newStatus)
	jsonSuccess(c, gin.H{
		"status":  newStatus,
		"message": msg,
	})
}
