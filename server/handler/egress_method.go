package handler

import (
	"fmt"
	"net"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"service-manage/config"
	"service-manage/model"
	frputil "service-manage/utils/frp"
	sshutil "service-manage/utils/ssh"

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
	machineIDStr := c.Query("machineId")
	isDirectStr := c.Query("isDirect")
	egressServiceIDStr := c.Query("egressServiceId")
	statusStr := c.Query("status")

	query := egressScope(c, h.DB).Model(&model.EgressMethod{})

	if serviceIDStr != "" {
		serviceID, _ := strconv.Atoi(serviceIDStr)
		query = query.Where("service_id = ?", serviceID)
	}
	if machineIDStr != "" {
		machineID, _ := strconv.Atoi(machineIDStr)
		var dockerServiceIDs []uint
		h.DB.Model(&model.DockerService{}).Where("machine_id = ?", machineID).Pluck("id", &dockerServiceIDs)
		var otherServiceIDs []uint
		h.DB.Model(&model.OtherService{}).Where("machine_id = ?", machineID).Pluck("id", &otherServiceIDs)

		if len(dockerServiceIDs) == 0 && len(otherServiceIDs) == 0 {
			query = query.Where("1 = 0")
		} else {
			conditions := h.DB.Where("(service_type = 'docker' AND service_id IN ?)", dockerServiceIDs)
			if len(otherServiceIDs) > 0 {
				conditions = conditions.Or("(service_type = 'other' AND service_id IN ?)", otherServiceIDs)
			}
			query = query.Where(conditions)
		}
	}
	if isDirectStr == "true" {
		query = query.Where("is_direct = ?", true)
	} else if isDirectStr == "false" {
		query = query.Where("is_direct = ?", false)
	}
	if egressServiceIDStr != "" {
		egressServiceID, _ := strconv.Atoi(egressServiceIDStr)
		query = query.Where("egress_service_id = ?", egressServiceID)
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

	dockerServiceIDs := make(map[uint]bool)
	otherServiceIDs := make(map[uint]bool)
	egressServiceIDs := make(map[uint]bool)
	for _, m := range methods {
		if m.ServiceType == "docker" {
			dockerServiceIDs[m.ServiceID] = true
		} else {
			otherServiceIDs[m.ServiceID] = true
		}
		if !m.IsDirect && m.EgressServiceID > 0 {
			egressServiceIDs[m.EgressServiceID] = true
		}
	}

	dockerServiceMap := make(map[uint]model.DockerService)
	if len(dockerServiceIDs) > 0 || len(egressServiceIDs) > 0 {
		allIDs := make([]uint, 0, len(dockerServiceIDs)+len(egressServiceIDs))
		for id := range dockerServiceIDs {
			allIDs = append(allIDs, id)
		}
		for id := range egressServiceIDs {
			if !dockerServiceIDs[id] {
				allIDs = append(allIDs, id)
			}
		}
		var dockerServices []model.DockerService
		h.DB.Unscoped().Where("id IN ?", allIDs).Find(&dockerServices)
		for _, ds := range dockerServices {
			dockerServiceMap[ds.ID] = ds
		}
	}

	otherServiceMap := make(map[uint]model.OtherService)
	if len(otherServiceIDs) > 0 {
		ids := make([]uint, 0, len(otherServiceIDs))
		for id := range otherServiceIDs {
			ids = append(ids, id)
		}
		var otherServices []model.OtherService
		h.DB.Unscoped().Where("id IN ?", ids).Find(&otherServices)
		for _, os := range otherServices {
			otherServiceMap[os.ID] = os
		}
	}

	machineIDs := make(map[uint]bool)
	for _, ds := range dockerServiceMap {
		machineIDs[ds.MachineID] = true
	}
	for _, os := range otherServiceMap {
		machineIDs[os.MachineID] = true
	}
	machineMap := make(map[uint]model.Machine)
	if len(machineIDs) > 0 {
		ids := make([]uint, 0, len(machineIDs))
		for id := range machineIDs {
			ids = append(ids, id)
		}
		var machines []model.Machine
		h.DB.Unscoped().Where("id IN ?", ids).Find(&machines)
		for _, m := range machines {
			machineMap[m.ID] = m
		}
	}

	var result []EgressMethodVO
	for _, m := range methods {
		serviceName := ""
		machineName := ""
		egressServiceName := ""

		if m.ServiceType == "docker" {
			if ds, ok := dockerServiceMap[m.ServiceID]; ok {
				serviceName = ds.Name
				if machine, ok := machineMap[ds.MachineID]; ok {
					machineName = machine.Name
				}
			}
		} else {
			if os, ok := otherServiceMap[m.ServiceID]; ok {
				serviceName = os.Name
				if machine, ok := machineMap[os.MachineID]; ok {
					machineName = machine.Name
				}
			}
		}

		if m.IsDirect {
			egressServiceName = "本机直连"
		} else if m.EgressServiceID > 0 {
			if es, ok := dockerServiceMap[m.EgressServiceID]; ok {
				egressServiceName = es.Name
				if machine, ok := machineMap[es.MachineID]; ok {
					egressServiceName = es.Name + "-" + machine.Name
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

	method.UserID = getUserId(c)

	var dup model.EgressMethod
	if err := h.DB.Where("public_ip = ? AND public_port = ?", method.PublicIP, method.PublicPort).First(&dup).Error; err == nil {
		jsonError(c, fmt.Sprintf("公网地址 %s:%d 已存在，端口不可重复", method.PublicIP, method.PublicPort))
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
		if dockerService.DockerSourceIP != "" && method.InternalIP == "" {
			method.InternalIP = dockerService.DockerSourceIP
		}
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
		jsonError(c, "创建出站方式失败")
		return
	}

	uid, uname := getLogUserInfo(c)
	logOperation(h.DB, uid, uname, "create", "egress_method", method.ID, method.ProxyName)

	jsonSuccess(c, gin.H{"id": method.ID})
}

func (h *EgressMethodHandler) Update(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var method model.EgressMethod
	if err := egressScope(c, h.DB).First(&method, id).Error; err != nil {
		jsonError(c, "出站方式不存在")
		return
	}

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		jsonError(c, "请求参数错误")
		return
	}

	updates = convertKeys(updates)

	newIP, ipOk := updates["public_ip"].(string)
	newPort, portOk := updates["public_port"]
	if !ipOk {
		newIP = method.PublicIP
	}
	if !portOk {
		switch v := newPort.(type) {
		case float64:
			newPort = int(v)
		default:
			newPort = method.PublicPort
		}
	} else {
		switch v := newPort.(type) {
		case float64:
			newPort = int(v)
		}
	}
	finalIP := newIP
	finalPort, _ := newPort.(int)
	if finalPort == 0 {
		finalPort = method.PublicPort
	}

	var dup model.EgressMethod
	if err := h.DB.Where("public_ip = ? AND public_port = ? AND id != ?", finalIP, finalPort, method.ID).First(&dup).Error; err == nil {
		jsonError(c, fmt.Sprintf("公网地址 %s:%d 已存在，端口不可重复", finalIP, finalPort))
		return
	}

	if err := h.DB.Model(&method).Updates(updates).Error; err != nil {
		jsonError(c, "更新出站方式失败")
		return
	}

	uid, uname := getLogUserInfo(c)
	logOperation(h.DB, uid, uname, "update", "egress_method", uint(id), method.ProxyName)

	jsonSuccess(c, nil)
}

func (h *EgressMethodHandler) Delete(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var method model.EgressMethod
	if err := egressScope(c, h.DB).First(&method, id).Error; err != nil {
		jsonError(c, "出站方式不存在")
		return
	}

	if err := h.DB.Delete(&method).Error; err != nil {
		jsonError(c, "删除出站方式失败")
		return
	}

	uid, uname := getLogUserInfo(c)
	logOperation(h.DB, uid, uname, "delete", "egress_method", uint(id), method.ProxyName)

	jsonSuccess(c, nil)
}

func (h *EgressMethodHandler) BatchUpdateStatus(c *gin.Context) {
	var req struct {
		IDs    []uint `json:"ids"`
		Status int8   `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || len(req.IDs) == 0 {
		jsonError(c, "请选择至少一个出站方式")
		return
	}
	if req.Status != 0 && req.Status != 1 {
		jsonError(c, "状态值无效")
		return
	}
	if err := egressScope(c, h.DB).Model(&model.EgressMethod{}).Where("id IN ?", req.IDs).Update("status", req.Status).Error; err != nil {
		jsonError(c, "批量更新状态失败")
		return
	}
	jsonSuccess(c, nil)
}

func (h *EgressMethodHandler) BatchDelete(c *gin.Context) {
	var req struct {
		IDs []uint `json:"ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || len(req.IDs) == 0 {
		jsonError(c, "请选择至少一个出站方式")
		return
	}
	if err := egressScope(c, h.DB).Where("id IN ?", req.IDs).Delete(&model.EgressMethod{}).Error; err != nil {
		jsonError(c, "批量删除失败")
		return
	}
	jsonSuccess(c, nil)
}

type ufwRule struct {
	num      int
	portSpec string
	action   string
}

type firewallResult struct {
	MachineID    uint   `json:"machineId"`
	MachineName  string `json:"machineName"`
	MachineIP    string `json:"machineIp"`
	Success      bool   `json:"success"`
	Message      string `json:"message"`
	AllowPorts   []int  `json:"allowPorts"`
	DenyPorts    []int  `json:"denyPorts"`
	SkippedPorts []int  `json:"skippedPorts"`
}

var ufwNumberedRuleRegex = regexp.MustCompile(`\[\s*(\d+)\]\s+(\S+)\s+(ALLOW|DENY)`)

func parsePortsFromSpec(spec string) []int {
	spec = strings.TrimSuffix(spec, "/tcp")
	spec = strings.TrimSuffix(spec, "/udp")
	var ports []int
	for _, part := range strings.Split(spec, ",") {
		if strings.Contains(part, ":") {
			rng := strings.SplitN(part, ":", 2)
			s, _ := strconv.Atoi(rng[0])
			e, _ := strconv.Atoi(rng[1])
			for i := s; i <= e; i++ {
				ports = append(ports, i)
			}
		} else {
			p, _ := strconv.Atoi(part)
			if p > 0 {
				ports = append(ports, p)
			}
		}
	}
	return ports
}

func isProtectedPort(port int) bool {
	return config.AppConfig.PortRange.IsProtectedPort(port)
}

func (h *EgressMethodHandler) resolveMachineID(m model.EgressMethod) uint {
	if m.IsDirect {
		if m.ServiceType == "other" {
			var os model.OtherService
			if err := h.DB.First(&os, m.ServiceID).Error; err != nil {
				return 0
			}
			return os.MachineID
		}
		var ds model.DockerService
		if err := h.DB.First(&ds, m.ServiceID).Error; err != nil {
			return 0
		}
		return ds.MachineID
	}
	if m.EgressServiceID == 0 {
		return 0
	}
	var ds model.DockerService
	if err := h.DB.First(&ds, m.EgressServiceID).Error; err != nil {
		return 0
	}
	return ds.MachineID
}

func (h *EgressMethodHandler) SyncFirewall(c *gin.Context) {
	var methods []model.EgressMethod
	if err := egressScope(c, h.DB).Find(&methods).Error; err != nil {
		jsonError(c, "查询出站方式失败")
		return
	}

	if !isAdmin(c) {
		for _, m := range methods {
			if m.PublicPort < config.AppConfig.PortRange.UserPortMin || m.PublicPort > config.AppConfig.PortRange.UserPortMax {
				jsonError(c, fmt.Sprintf("普通用户仅可使用 %s 端口范围，当前存在超出范围的端口，无法同步", config.AppConfig.PortRange.UserPortRangeDesc()))
				return
			}
		}
	}

	var allMethods []model.EgressMethod
	h.DB.Find(&allMethods)

	type machineAction struct {
		machine    model.Machine
		allowPorts map[int]bool
		denyPorts  map[int]bool
	}
	machineMap := make(map[uint]*machineAction)
	globalAllowMap := make(map[uint]map[int]bool)

	addPortsToMap := func(targetMap map[uint]map[int]bool, machineID uint, port int) {
		if targetMap[machineID] == nil {
			targetMap[machineID] = make(map[int]bool)
		}
		targetMap[machineID][port] = true
	}

	for _, m := range allMethods {
		if m.PublicPort <= 0 {
			continue
		}
		if m.Status != 1 {
			continue
		}
		var machineID uint
		machineID = h.resolveMachineID(m)
		if machineID == 0 {
			continue
		}
		addPortsToMap(globalAllowMap, machineID, m.PublicPort)
	}

	for _, m := range methods {
		if m.PublicPort <= 0 {
			continue
		}
		var machineID uint
		machineID = h.resolveMachineID(m)
		if machineID == 0 {
			continue
		}

		if _, exists := machineMap[machineID]; !exists {
			var machine model.Machine
			if err := h.DB.First(&machine, machineID).Error; err != nil {
				continue
			}
			machineMap[machineID] = &machineAction{
				machine:    machine,
				allowPorts: make(map[int]bool),
				denyPorts:  make(map[int]bool),
			}
		}

		if isProtectedPort(m.PublicPort) {
			continue
		}
		if m.Status == 1 {
			machineMap[machineID].allowPorts[m.PublicPort] = true
		} else {
			if !machineMap[machineID].allowPorts[m.PublicPort] {
				machineMap[machineID].denyPorts[m.PublicPort] = true
			}
		}
	}

	var results []firewallResult

	for _, ma := range machineMap {
		machine := ma.machine

		if machine.SSHUser == "" || machine.SSHPassword == "" {
			results = append(results, firewallResult{
				MachineID: machine.ID, MachineName: machine.Name, MachineIP: machine.IP,
				Success: false, Message: "未配置SSH凭据",
			})
			continue
		}

		sshPort := machine.SSHPort
		if sshPort == 0 {
			sshPort = config.AppConfig.SSH.DefaultPort
		}
		cfg := &sshutil.Config{
			Host: machine.IP, Port: sshPort,
			User: machine.SSHUser, Password: machine.SSHPassword,
		}

		if err := sshutil.CheckConnection(cfg); err != nil {
			results = append(results, firewallResult{
				MachineID: machine.ID, MachineName: machine.Name, MachineIP: machine.IP,
				Success: false, Message: "SSH连接失败: " + err.Error(),
			})
			continue
		}

		statusOutput, err := sshutil.RunCommand(cfg, "ufw status numbered")
		if err != nil {
			results = append(results, firewallResult{
				MachineID: machine.ID, MachineName: machine.Name, MachineIP: machine.IP,
				Success: false, Message: "获取ufw状态失败: " + err.Error(),
			})
			continue
		}

		var rules []ufwRule
		for _, line := range strings.Split(statusOutput, "\n") {
			matches := ufwNumberedRuleRegex.FindStringSubmatch(line)
			if len(matches) == 4 {
				n, _ := strconv.Atoi(matches[1])
				rules = append(rules, ufwRule{num: n, portSpec: matches[2], action: matches[3]})
			}
		}

		var cmds []string
		var deleteCmds []string
		deleteCount := 0
		skipCount := 0
		var skippedPorts []int
		existingAllow := make(map[int]bool)
		existingDeny := make(map[int]bool)

		sort.Slice(rules, func(i, j int) bool { return rules[i].num > rules[j].num })

		globalPorts := globalAllowMap[machine.ID]
		if globalPorts == nil {
			globalPorts = make(map[int]bool)
		}

		for _, rule := range rules {
			ports := parsePortsFromSpec(rule.portSpec)
			if len(ports) == 0 {
				continue
			}

			allProtected := true
			for _, p := range ports {
				if !isProtectedPort(p) {
					allProtected = false
					break
				}
			}
			if allProtected {
				continue
			}

			allGlobal := true
			for _, p := range ports {
				if !globalPorts[p] {
					allGlobal = false
					break
				}
			}
			if allGlobal && rule.action == "ALLOW" {
				for _, p := range ports {
					existingAllow[p] = true
					skippedPorts = append(skippedPorts, p)
				}
				skipCount++
				continue
			}

			allMyDeny := true
			for _, p := range ports {
				if !ma.denyPorts[p] {
					allMyDeny = false
					break
				}
			}
			if allMyDeny && rule.action == "DENY" {
				for _, p := range ports {
					existingDeny[p] = true
					skippedPorts = append(skippedPorts, p)
				}
				skipCount++
				continue
			}

			deleteCmds = append(deleteCmds, fmt.Sprintf("ufw --force delete %d", rule.num))
			deleteCount++
		}

		cmds = append(cmds, deleteCmds...)

		for p := range ma.allowPorts {
			if !existingAllow[p] {
				cmds = append(cmds, fmt.Sprintf("ufw allow %d/tcp", p))
			}
		}
		for p := range ma.denyPorts {
			if !existingDeny[p] {
				cmds = append(cmds, fmt.Sprintf("ufw deny %d/tcp", p))
			}
		}

		if len(cmds) > 0 {
			batchCmd := strings.Join(cmds, "; ")
			if _, err := sshutil.RunCommand(cfg, batchCmd); err != nil {
				results = append(results, firewallResult{
					MachineID: machine.ID, MachineName: machine.Name, MachineIP: machine.IP,
					Success: false, Message: "执行命令失败: " + err.Error(),
				})
				continue
			}
		}

		var allowList, denyList []int
		for p := range ma.allowPorts {
			allowList = append(allowList, p)
		}
		for p := range ma.denyPorts {
			denyList = append(denyList, p)
		}
		sort.Ints(allowList)
		sort.Ints(denyList)

		msg := "同步完成"
		if skipCount > 0 {
			msg += fmt.Sprintf("，%d 条规则无需变动", skipCount)
		}
		if deleteCount > 0 {
			msg += fmt.Sprintf("，删除 %d 条规则", deleteCount)
		}
		newCount := 0
		for p := range ma.allowPorts {
			if !existingAllow[p] {
				newCount++
			}
		}
		for p := range ma.denyPorts {
			if !existingDeny[p] {
				newCount++
			}
		}
		if newCount > 0 {
			msg += fmt.Sprintf("，新增 %d 条规则", newCount)
		}
		if len(cmds) == 0 {
			msg = "防火墙已是最新状态，无需同步"
		}

		results = append(results, firewallResult{
			MachineID: machine.ID, MachineName: machine.Name, MachineIP: machine.IP,
			Success: true, Message: msg,
			AllowPorts: allowList, DenyPorts: denyList, SkippedPorts: skippedPorts,
		})
	}

	jsonSuccess(c, gin.H{
		"results": results,
		"total":   len(results),
	})
}

func (h *EgressMethodHandler) GenerateFrpc(c *gin.Context) {
	var req struct {
		IDs []uint `json:"ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || len(req.IDs) == 0 {
		jsonError(c, "请选择至少一个出站方式")
		return
	}

	var methods []model.EgressMethod
	if err := egressScope(c, h.DB).Where("id IN ?", req.IDs).Find(&methods).Error; err != nil || len(methods) == 0 {
		jsonError(c, "出站方式不存在")
		return
	}

	var egressServiceID uint
	for _, m := range methods {
		if m.IsDirect {
			jsonError(c, "本机直连的出站方式不支持生成 FRP 配置")
			return
		}
		if egressServiceID == 0 {
			egressServiceID = m.EgressServiceID
		} else if m.EgressServiceID != egressServiceID {
			jsonError(c, "请选择同一出站服务的出站方式")
			return
		}
	}

	var egressService model.DockerService
	if err := serviceScope(c, h.DB).First(&egressService, egressServiceID).Error; err != nil {
		jsonError(c, "出站服务不存在")
		return
	}

	var egressMachine model.Machine
	if err := h.DB.First(&egressMachine, egressService.MachineID).Error; err != nil {
		jsonError(c, "出站服务主机不存在")
		return
	}

	frpCfg := frputil.GetServerConfig(&frputil.DiscoverParams{
		MachineIP:     egressMachine.IP,
		SSHPort:       egressMachine.SSHPort,
		SSHUser:       egressMachine.SSHUser,
		SSHPassword:   egressMachine.SSHPassword,
		ContainerName: egressService.Name,
	})
	if frpCfg.ServerPort <= 0 || frpCfg.AuthToken == "" {
		jsonError(c, "FRP 服务端配置获取失败，请检查出站服务所在主机的 SSH 连通性及容器配置文件挂载")
		return
	}

	var lines []string
	lines = append(lines, fmt.Sprintf(`serverAddr = "%s"`, egressMachine.IP))
	lines = append(lines, fmt.Sprintf("serverPort = %d", frpCfg.ServerPort))
	lines = append(lines, fmt.Sprintf(`auth.token = "%s"`, frpCfg.AuthToken))
	lines = append(lines, "")

	for _, m := range methods {
		var serviceName string
		if m.ServiceType == "other" {
			var os model.OtherService
			if err := serviceScope(c, h.DB).First(&os, m.ServiceID).Error; err == nil {
				serviceName = os.Name
			}
		} else {
			var ds model.DockerService
			if err := serviceScope(c, h.DB).First(&ds, m.ServiceID).Error; err == nil {
				serviceName = ds.Name
			}
		}
		if serviceName == "" {
			serviceName = fmt.Sprintf("svc-%d", m.ServiceID)
		}

		sectionName := m.ProxyName
		if sectionName == "" {
			sectionName = fmt.Sprintf("%s.%d", serviceName, m.PublicPort)
		}

		protocol := strings.ToLower(m.Protocol)
		if protocol == "" {
			protocol = "tcp"
		}

		lines = append(lines, "[[proxies]]")
		lines = append(lines, fmt.Sprintf(`name = "%s"`, sectionName))
		lines = append(lines, fmt.Sprintf(`type = "%s"`, protocol))
		lines = append(lines, fmt.Sprintf(`localIP = "%s"`, m.InternalIP))
		lines = append(lines, fmt.Sprintf("localPort = %d", m.InternalPort))
		lines = append(lines, fmt.Sprintf("remotePort = %d", m.PublicPort))
		lines = append(lines, "")
	}

	result := strings.Join(lines, "\n")

	jsonSuccess(c, gin.H{
		"config": result,
	})
}

func (h *EgressMethodHandler) HealthCheck(c *gin.Context) {
	var methods []model.EgressMethod
	egressScope(c, h.DB).Where("status = ?", 1).Find(&methods)

	type CheckResult struct {
		ID          uint   `json:"id"`
		ProxyName   string `json:"proxyName"`
		ServiceName string `json:"serviceName"`
		Reachable   bool   `json:"reachable"`
		Latency     int64  `json:"latency"`
	}

	dockerServiceIDs := make(map[uint]bool)
	otherServiceIDs := make(map[uint]bool)
	for _, m := range methods {
		if m.ServiceID == 0 {
			continue
		}
		if m.ServiceType == "other" {
			otherServiceIDs[m.ServiceID] = true
		} else {
			dockerServiceIDs[m.ServiceID] = true
		}
	}

	serviceNameMap := make(map[string]string)
	if len(dockerServiceIDs) > 0 {
		ids := make([]uint, 0, len(dockerServiceIDs))
		for id := range dockerServiceIDs {
			ids = append(ids, id)
		}
		var list []model.DockerService
		h.DB.Unscoped().Where("id IN ?", ids).Find(&list)
		for _, ds := range list {
			serviceNameMap[fmt.Sprintf("docker-%d", ds.ID)] = ds.Name
		}
	}
	if len(otherServiceIDs) > 0 {
		ids := make([]uint, 0, len(otherServiceIDs))
		for id := range otherServiceIDs {
			ids = append(ids, id)
		}
		var list []model.OtherService
		h.DB.Unscoped().Where("id IN ?", ids).Find(&list)
		for _, os := range list {
			serviceNameMap[fmt.Sprintf("other-%d", os.ID)] = os.Name
		}
	}

	timeout := time.Duration(config.AppConfig.HealthCheck.Timeout) * time.Second
	if timeout <= 0 {
		timeout = time.Second
	}
	usePublicIP := config.AppConfig.HealthCheck.UsePublicIP

	results := make([]CheckResult, len(methods))
	var wg sync.WaitGroup
	for i, m := range methods {
		wg.Add(1)
		go func(idx int, method model.EgressMethod) {
			defer wg.Done()
			results[idx] = CheckResult{
				ID:          method.ID,
				ProxyName:   method.ProxyName,
				ServiceName: serviceNameMap[fmt.Sprintf("%s-%d", method.ServiceType, method.ServiceID)],
			}

			var addr string
			if usePublicIP && method.PublicIP != "" && method.PublicPort > 0 {
				addr = fmt.Sprintf("%s:%d", method.PublicIP, method.PublicPort)
			} else if method.InternalIP != "" && method.InternalPort > 0 {
				addr = fmt.Sprintf("%s:%d", method.InternalIP, method.InternalPort)
			} else {
				return
			}

			start := time.Now()
			conn, err := net.DialTimeout("tcp", addr, timeout)
			latency := time.Since(start).Milliseconds()
			if err == nil {
				conn.Close()
				results[idx].Reachable = true
				results[idx].Latency = latency
			}
		}(i, m)
	}
	wg.Wait()

	jsonSuccess(c, results)
}
