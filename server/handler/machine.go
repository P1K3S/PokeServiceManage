package handler

import (
	"encoding/json"
	"fmt"
	"net"
	"strconv"
	"strings"

	"service-manage/model"
	sshutil "service-manage/utils/ssh"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type MachineHandler struct {
	DB *gorm.DB
}

func NewMachineHandler(db *gorm.DB) *MachineHandler {
	return &MachineHandler{DB: db}
}

func clearMachinePassword(m *model.Machine) {
	m.SSHPassword = ""
}

func (h *MachineHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	keyword := c.Query("keyword")
	machineType := c.Query("machineType")
	statusStr := c.Query("status")

	query := h.DB.Model(&model.Machine{})

	if keyword != "" {
		query = query.Where("name LIKE ?", "%"+keyword+"%")
	}
	if machineType != "" {
		query = query.Where("machine_type = ?", machineType)
	}
	if statusStr != "" {
		status, _ := strconv.Atoi(statusStr)
		query = query.Where("status = ?", status)
	}

	var total int64
	query.Count(&total)

	var machines []model.Machine
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("id DESC").Find(&machines).Error; err != nil {
		jsonError(c, "查询主机列表失败")
		return
	}

	type MachineVO struct {
		model.Machine
		ServiceCount int64 `json:"serviceCount"`
	}

	var result []MachineVO
	for _, m := range machines {
		clearMachinePassword(&m)
		var count int64
		var dockerCount, otherCount int64
		h.DB.Model(&model.DockerService{}).Where("machine_id = ?", m.ID).Count(&dockerCount)
		h.DB.Model(&model.OtherService{}).Where("machine_id = ?", m.ID).Count(&otherCount)
		count = dockerCount + otherCount
		result = append(result, MachineVO{
			Machine:      m,
			ServiceCount: count,
		})
	}

	jsonPage(c, result, total, page, pageSize)
}

func (h *MachineHandler) Get(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var machine model.Machine
	if err := h.DB.First(&machine, id).Error; err != nil {
		jsonError(c, "主机不存在")
		return
	}
	clearMachinePassword(&machine)
	jsonSuccess(c, machine)
}

func (h *MachineHandler) Create(c *gin.Context) {
	var machine model.Machine
	if err := c.ShouldBindJSON(&machine); err != nil {
		jsonError(c, "请求参数错误")
		return
	}

	if machine.SSHPort == 0 {
		machine.SSHPort = 22
	}
	if machine.SSHUser == "" {
		machine.SSHUser = "root"
	}

	var count int64
	h.DB.Model(&model.Machine{}).Where("name = ?", machine.Name).Count(&count)
	if count > 0 {
		jsonErrorCode(c, 1004, "主机名称已存在")
		return
	}

	if err := h.DB.Create(&machine).Error; err != nil {
		jsonError(c, "创建主机失败")
		return
	}

	if machine.SSHUser != "" && machine.SSHPassword != "" {
		go h.discoverDockerServices(machine.ID)
	}

	jsonSuccess(c, gin.H{"id": machine.ID})
}

func (h *MachineHandler) Update(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var machine model.Machine
	if err := h.DB.First(&machine, id).Error; err != nil {
		jsonError(c, "主机不存在")
		return
	}

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		jsonError(c, "请求参数错误")
		return
	}

	if name, ok := updates["name"].(string); ok {
		var count int64
		h.DB.Model(&model.Machine{}).Where("name = ? AND id != ?", name, id).Count(&count)
		if count > 0 {
			jsonErrorCode(c, 1004, "主机名称已存在")
			return
		}
	}

	if pwd, ok := updates["sshPassword"].(string); ok && pwd == "" {
		delete(updates, "sshPassword")
	}

	updates = convertKeys(updates)

	oldIP := machine.IP

	if err := h.DB.Model(&machine).Updates(updates).Error; err != nil {
		jsonError(c, "更新主机失败")
		return
	}

	newIP, _ := updates["ip"].(string)
	if newIP != "" && newIP != oldIP {
		h.syncEgressMethodIP(uint(id), oldIP, newIP)
	}

	jsonSuccess(c, nil)
}

func (h *MachineHandler) syncEgressMethodIP(machineID uint, oldIP, newIP string) {
	var dockerServiceIDs []uint
	h.DB.Model(&model.DockerService{}).Where("machine_id = ?", machineID).Pluck("id", &dockerServiceIDs)

	var otherServiceIDs []uint
	h.DB.Model(&model.OtherService{}).Where("machine_id = ?", machineID).Pluck("id", &otherServiceIDs)

	if len(dockerServiceIDs) > 0 {
		h.DB.Model(&model.EgressMethod{}).
			Where("service_id IN ? AND service_type = ?", dockerServiceIDs, "docker").
			Where("internal_ip = ?", oldIP).
			Update("internal_ip", newIP)

		h.DB.Model(&model.EgressMethod{}).
			Where("service_id IN ? AND service_type = ? AND is_direct = ?", dockerServiceIDs, "docker", true).
			Where("public_ip = ?", oldIP).
			Update("public_ip", newIP)

		h.DB.Model(&model.EgressMethod{}).
			Where("egress_service_id IN ? AND is_direct = ?", dockerServiceIDs, false).
			Where("public_ip = ?", oldIP).
			Update("public_ip", newIP)
	}

	if len(otherServiceIDs) > 0 {
		h.DB.Model(&model.EgressMethod{}).
			Where("service_id IN ? AND service_type = ?", otherServiceIDs, "other").
			Where("internal_ip = ?", oldIP).
			Update("internal_ip", newIP)

		h.DB.Model(&model.EgressMethod{}).
			Where("service_id IN ? AND service_type = ? AND is_direct = ?", otherServiceIDs, "other", true).
			Where("public_ip = ?", oldIP).
			Update("public_ip", newIP)
	}
}

func (h *MachineHandler) Delete(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var machine model.Machine
	if err := h.DB.First(&machine, id).Error; err != nil {
		jsonError(c, "主机不存在")
		return
	}

	tx := h.DB.Begin()

	if err := tx.Where("machine_id = ?", id).Delete(&model.DockerService{}).Error; err != nil {
		tx.Rollback()
		jsonError(c, "删除关联Docker服务失败")
		return
	}
	if err := tx.Where("machine_id = ?", id).Delete(&model.OtherService{}).Error; err != nil {
		tx.Rollback()
		jsonError(c, "删除关联其他服务失败")
		return
	}

	var dockerIDs, otherIDs []uint
	tx.Model(&model.DockerService{}).Unscoped().Where("machine_id = ?", id).Pluck("id", &dockerIDs)
	tx.Model(&model.OtherService{}).Unscoped().Where("machine_id = ?", id).Pluck("id", &otherIDs)
	if len(dockerIDs) > 0 {
		if err := tx.Where("service_id IN ? AND service_type = ?", dockerIDs, "docker").Delete(&model.EgressMethod{}).Error; err != nil {
			tx.Rollback()
			jsonError(c, "删除关联出站方式失败")
			return
		}
	}
	if len(otherIDs) > 0 {
		if err := tx.Where("service_id IN ? AND service_type = ?", otherIDs, "other").Delete(&model.EgressMethod{}).Error; err != nil {
			tx.Rollback()
			jsonError(c, "删除关联出站方式失败")
			return
		}
	}

	if err := tx.Delete(&machine).Error; err != nil {
		tx.Rollback()
		jsonError(c, "删除主机失败")
		return
	}

	tx.Commit()
	jsonSuccess(c, nil)
}

func (h *MachineHandler) Overview(c *gin.Context) {
	var machineTotal, serviceTotal int64
	var machineOnline, serviceRunning int64

	h.DB.Model(&model.Machine{}).Count(&machineTotal)
	h.DB.Model(&model.Machine{}).Where("status = 1").Count(&machineOnline)
	var dockerTotal, otherTotal int64
	h.DB.Model(&model.DockerService{}).Count(&dockerTotal)
	h.DB.Model(&model.OtherService{}).Count(&otherTotal)
	serviceTotal = dockerTotal + otherTotal
	var dockerRunning, otherRunning int64
	h.DB.Model(&model.DockerService{}).Where("status = 1").Count(&dockerRunning)
	h.DB.Model(&model.OtherService{}).Where("status = 1").Count(&otherRunning)
	serviceRunning = dockerRunning + otherRunning

	var recentMachines []model.Machine
	h.DB.Model(&model.Machine{}).Order("id DESC").Limit(5).Find(&recentMachines)

	type MachineVO struct {
		model.Machine
		ServiceCount int64 `json:"serviceCount"`
	}
	var recentMachineVOs []MachineVO
	for _, m := range recentMachines {
		clearMachinePassword(&m)
		var count int64
		var dockerCount, otherCount int64
		h.DB.Model(&model.DockerService{}).Where("machine_id = ?", m.ID).Count(&dockerCount)
		h.DB.Model(&model.OtherService{}).Where("machine_id = ?", m.ID).Count(&otherCount)
		count = dockerCount + otherCount
		recentMachineVOs = append(recentMachineVOs, MachineVO{
			Machine:      m,
			ServiceCount: count,
		})
	}

	jsonSuccess(c, gin.H{
		"machineTotal":   machineTotal,
		"serviceTotal":   serviceTotal,
		"machineOnline":  machineOnline,
		"serviceRunning": serviceRunning,
		"recentMachines": recentMachineVOs,
	})
}

func (h *MachineHandler) CheckSSH(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var machine model.Machine
	if err := h.DB.First(&machine, id).Error; err != nil {
		jsonError(c, "主机不存在")
		return
	}

	sshPort := machine.SSHPort
	if sshPort == 0 {
		sshPort = 22
	}

	err := sshutil.CheckConnection(&sshutil.Config{
		Host:     machine.IP,
		Port:     sshPort,
		User:     machine.SSHUser,
		Password: machine.SSHPassword,
	})

	newStatus := int8(1)
	msg := "SSH连接成功"
	if err != nil {
		newStatus = 0
		msg = "SSH连接失败"
	}

	h.DB.Model(&machine).Update("status", newStatus)
	jsonSuccess(c, gin.H{
		"status":  newStatus,
		"message": msg,
	})
}

type dockerContainer struct {
	Names string `json:"Names"`
	State string `json:"State"`
	Ports string `json:"Ports"`
	Image string `json:"Image"`
}

type portMapping struct {
	HostPort      string `json:"hostPort"`
	ContainerPort string `json:"containerPort"`
	Protocol      string `json:"protocol"`
}

func parseRange(s string) (start, end int) {
	if !strings.Contains(s, "-") {
		v, _ := strconv.Atoi(s)
		return v, v
	}
	parts := strings.SplitN(s, "-", 2)
	start, _ = strconv.Atoi(parts[0])
	end, _ = strconv.Atoi(parts[1])
	return
}

func expandPortRange(hostP, containerP, proto string) []portMapping {
	hs, he := parseRange(hostP)
	cs, ce := parseRange(containerP)
	hl := he - hs + 1
	cl := ce - cs + 1
	max := hl
	if cl > max {
		max = cl
	}
	var result []portMapping
	for i := 0; i < max; i++ {
		h := hs
		if hl > 1 {
			h = hs + i
		}
		c := cs
		if cl > 1 {
			c = cs + i
		}
		result = append(result, portMapping{
			HostPort:      strconv.Itoa(h),
			ContainerPort: strconv.Itoa(c),
			Protocol:      proto,
		})
	}
	return result
}

func parseDockerPorts(ports string) (hostPort, containerPort int, protocol string, mappingsJSON string) {
	protocol = "TCP"
	if ports == "" {
		mappingsJSON = "[]"
		return
	}

	var mappings []portMapping
	seen := make(map[string]bool)

	for _, part := range strings.Split(ports, ", ") {
		part = strings.TrimSpace(part)
		if part == "" || strings.HasPrefix(part, "[") || !strings.Contains(part, "->") {
			continue
		}

		var hostP, containerP, proto string

		if strings.Contains(part, "->") {
			mapping := strings.SplitN(part, "->", 2)
			hostPart := mapping[0]
			containerPart := mapping[1]

			if idx := strings.Index(containerPart, "/"); idx >= 0 {
				containerP = containerPart[:idx]
				proto = strings.ToUpper(containerPart[idx+1:])
			} else {
				containerP = containerPart
			}
			if idx := strings.LastIndex(hostPart, ":"); idx >= 0 {
				hostP = hostPart[idx+1:]
			} else {
				hostP = hostPart
			}
		} else {
			if idx := strings.Index(part, "/"); idx >= 0 {
				containerP = part[:idx]
				proto = strings.ToUpper(part[idx+1:])
			} else {
				containerP = part
			}
		}

		if proto == "" {
			proto = "TCP"
		}

		expanded := expandPortRange(hostP, containerP, proto)
		for _, m := range expanded {
			dk := m.HostPort + "|" + m.ContainerPort + "|" + m.Protocol
			if seen[dk] {
				continue
			}
			seen[dk] = true
			mappings = append(mappings, m)
		}
	}

	for _, m := range mappings {
		if hostPort == 0 {
			hostPort, _ = strconv.Atoi(m.HostPort)
		}
		if containerPort == 0 {
			containerPort, _ = strconv.Atoi(m.ContainerPort)
		}
		if protocol == "TCP" && m.Protocol != "" {
			protocol = m.Protocol
		}
	}
	if protocol == "" {
		protocol = "TCP"
	}

	jsonBytes, _ := json.Marshal(mappings)
	mappingsJSON = string(jsonBytes)
	return
}

func (h *MachineHandler) discoverDockerServices(machineID uint) (int, error) {
	var machine model.Machine
	if err := h.DB.First(&machine, machineID).Error; err != nil {
		return 0, err
	}
	if machine.SSHUser == "" || machine.SSHPassword == "" {
		return 0, fmt.Errorf("SSH not configured")
	}
	sshPort := machine.SSHPort
	if sshPort == 0 {
		sshPort = 22
	}
	output, err := sshutil.RunCommand(&sshutil.Config{
		Host:     machine.IP,
		Port:     sshPort,
		User:     machine.SSHUser,
		Password: machine.SSHPassword,
	}, "docker ps -a --format '{{json .}}'")
	if err != nil {
		return 0, err
	}

	ipOutput, _ := sshutil.RunCommand(&sshutil.Config{
		Host:     machine.IP,
		Port:     sshPort,
		User:     machine.SSHUser,
		Password: machine.SSHPassword,
	}, "docker inspect $(docker ps -a -q) -f '{{.Name}}\t{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' 2>/dev/null; true")

	ipMap := make(map[string]string)
	for _, line := range strings.Split(strings.TrimSpace(ipOutput), "\n") {
		parts := strings.SplitN(line, "\t", 2)
		if len(parts) == 2 && parts[1] != "" {
			ipMap[strings.TrimPrefix(parts[0], "/")] = parts[1]
		}
	}

	lines := strings.Split(strings.TrimSpace(output), "\n")
	count := 0
	for _, line := range lines {
		if line == "" {
			continue
		}
		var container dockerContainer
		if err := json.Unmarshal([]byte(line), &container); err != nil {
			continue
		}

		containerIP := ipMap[container.Names]

		var existing model.DockerService
		result := h.DB.Where("machine_id = ? AND name = ?",
			machineID, container.Names).First(&existing)
		hostPort, containerPort, protocol, mappingsJSON := parseDockerPorts(container.Ports)
		status := int8(0)
		if container.State == "running" {
			status = 1
		}
		if containerIP == "" || net.ParseIP(containerIP) == nil {
			containerIP = machine.IP
		}
		if result.RowsAffected > 0 {
			if existing.Locked {
				continue
			}
			h.DB.Model(&existing).Updates(map[string]interface{}{
				"status":             status,
				"port":               hostPort,
				"docker_source_ip":   containerIP,
				"docker_source_port": containerPort,
				"protocol":           protocol,
				"port_mappings":      mappingsJSON,
			})
		} else {
			service := model.DockerService{
				MachineID:        machineID,
				Name:             container.Names,
				Port:             hostPort,
				Protocol:         protocol,
				DockerSourceIP:   containerIP,
				DockerSourcePort: containerPort,
				PortMappings:     mappingsJSON,
				Status:           status,
			}
			if err := h.DB.Create(&service).Error; err != nil {
				continue
			}
		}
		count++
	}
	return count, nil
}

func (h *MachineHandler) DiscoverServices(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	count, err := h.discoverDockerServices(uint(id))
	if err != nil {
		msg := err.Error()
		if strings.Contains(msg, "SSH not configured") {
			jsonError(c, "主机未配置 SSH 连接信息，请先在主机管理中设置")
		} else if strings.Contains(msg, "connection refused") || strings.Contains(msg, "no route to host") || strings.Contains(msg, "i/o timeout") {
			jsonError(c, "SSH 连接失败，请检查主机状态和网络连接")
		} else {
			jsonError(c, "检测失败："+msg)
		}
		return
	}
	jsonSuccess(c, gin.H{
		"count":   count,
		"message": fmt.Sprintf("检测完成，更新 %d 个 Docker 服务", count),
	})
}
