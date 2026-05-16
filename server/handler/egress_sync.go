package handler

import (
	"fmt"

	"service-manage/model"

	"gorm.io/gorm"
)

func syncEgressInternalIP(db *gorm.DB, serviceID uint, serviceType string, oldIP, newIP string) {
	q := db.Model(&model.EgressMethod{}).
		Where("service_id = ? AND service_type = ?", serviceID, serviceType)
	if oldIP != "" {
		q = q.Where("internal_ip = ?", oldIP)
	}
	q.Update("internal_ip", newIP)
}

func syncEgressPort(db *gorm.DB, serviceID uint, serviceType string, oldPort, newPort int) {
	db.Model(&model.EgressMethod{}).
		Where("service_id = ? AND service_type = ? AND internal_port = ?", serviceID, serviceType, oldPort).
		Update("internal_port", newPort)

	db.Model(&model.EgressMethod{}).
		Where("service_id = ? AND service_type = ? AND is_direct = ? AND public_port = ?", serviceID, serviceType, true, oldPort).
		Update("public_port", newPort)

	if serviceType == "docker" {
		db.Model(&model.EgressMethod{}).
			Where("egress_service_id = ? AND is_direct = ? AND public_port = ?", serviceID, false, oldPort).
			Update("public_port", newPort)
	}

	rebuildProxyNamesByService(db, serviceID, serviceType)
	if serviceType == "docker" {
		rebuildProxyNamesByEgressService(db, serviceID)
	}
}

func syncEgressMachineChange(db *gorm.DB, serviceID uint, serviceType string, newMachineID uint) {
	var machine model.Machine
	if err := db.First(&machine, newMachineID).Error; err != nil {
		return
	}

	db.Model(&model.EgressMethod{}).
		Where("service_id = ? AND service_type = ? AND is_direct = ?", serviceID, serviceType, true).
		Update("public_ip", machine.IP)

	internalIP := machine.IP
	if serviceType == "docker" {
		var svc model.DockerService
		if db.First(&svc, serviceID).Error == nil && svc.DockerSourceIP != "" {
			internalIP = svc.DockerSourceIP
		}
	}
	db.Model(&model.EgressMethod{}).
		Where("service_id = ? AND service_type = ?", serviceID, serviceType).
		Update("internal_ip", internalIP)

	rebuildProxyNamesByService(db, serviceID, serviceType)
}

func rebuildProxyNamesByService(db *gorm.DB, serviceID uint, serviceType string) {
	var methods []model.EgressMethod
	db.Where("service_id = ? AND service_type = ? AND is_direct = ?", serviceID, serviceType, true).Find(&methods)
	for i := range methods {
		rebuildProxyName(db, &methods[i], serviceType)
	}
}

func rebuildProxyNamesByEgressService(db *gorm.DB, egressServiceID uint) {
	var methods []model.EgressMethod
	db.Where("egress_service_id = ? AND is_direct = ?", egressServiceID, false).Find(&methods)
	for i := range methods {
		var svc model.DockerService
		if db.First(&svc, egressServiceID).Error != nil {
			continue
		}
		rebuildProxyNameWithMachine(db, &methods[i], svc.MachineID)
	}
}

func rebuildProxyName(db *gorm.DB, method *model.EgressMethod, serviceType string) {
	var machineID uint
	if serviceType == "docker" {
		var svc model.DockerService
		if db.First(&svc, method.ServiceID).Error != nil {
			return
		}
		machineID = svc.MachineID
	} else {
		var svc model.OtherService
		if db.First(&svc, method.ServiceID).Error != nil {
			return
		}
		machineID = svc.MachineID
	}
	rebuildProxyNameWithMachine(db, method, machineID)
}

func rebuildProxyNameWithMachine(db *gorm.DB, method *model.EgressMethod, machineID uint) {
	var machine model.Machine
	if db.First(&machine, machineID).Error != nil {
		return
	}
	if method.PublicPort > 0 {
		db.Model(method).Update("proxy_name", fmt.Sprintf("%s.%d", machine.Name, method.PublicPort))
	}
}

func getServiceIDsByMachine(db *gorm.DB, machineID uint) (dockerIDs, otherIDs []uint) {
	db.Model(&model.DockerService{}).Where("machine_id = ?", machineID).Pluck("id", &dockerIDs)
	db.Model(&model.OtherService{}).Where("machine_id = ?", machineID).Pluck("id", &otherIDs)
	return
}

func syncMachineEgressIP(db *gorm.DB, machineID uint, oldIP, newIP string) {
	dockerIDs, otherIDs := getServiceIDsByMachine(db, machineID)

	if len(dockerIDs) > 0 {
		db.Model(&model.EgressMethod{}).
			Where("service_id IN ? AND service_type = ?", dockerIDs, "docker").
			Where("internal_ip = ?", oldIP).
			Update("internal_ip", newIP)

		db.Model(&model.EgressMethod{}).
			Where("service_id IN ? AND service_type = ? AND is_direct = ?", dockerIDs, "docker", true).
			Where("public_ip = ?", oldIP).
			Update("public_ip", newIP)

		db.Model(&model.EgressMethod{}).
			Where("egress_service_id IN ? AND is_direct = ?", dockerIDs, false).
			Where("public_ip = ?", oldIP).
			Update("public_ip", newIP)
	}

	if len(otherIDs) > 0 {
		db.Model(&model.EgressMethod{}).
			Where("service_id IN ? AND service_type = ?", otherIDs, "other").
			Where("internal_ip = ?", oldIP).
			Update("internal_ip", newIP)

		db.Model(&model.EgressMethod{}).
			Where("service_id IN ? AND service_type = ? AND is_direct = ?", otherIDs, "other", true).
			Where("public_ip = ?", oldIP).
			Update("public_ip", newIP)
	}
}

func syncMachineEgressProxyName(db *gorm.DB, machineID uint, newMachineName string) {
	dockerIDs, otherIDs := getServiceIDsByMachine(db, machineID)

	var methods []model.EgressMethod
	if len(dockerIDs) > 0 {
		db.Where("service_id IN ? AND service_type = ?", dockerIDs, "docker").Find(&methods)
	}
	if len(otherIDs) > 0 {
		var otherMethods []model.EgressMethod
		db.Where("service_id IN ? AND service_type = ?", otherIDs, "other").Find(&otherMethods)
		methods = append(methods, otherMethods...)
	}

	for i := range methods {
		if methods[i].PublicPort > 0 {
			db.Model(&methods[i]).Update("proxy_name", fmt.Sprintf("%s.%d", newMachineName, methods[i].PublicPort))
		}
	}
}
