package model

import (
	"gorm.io/gorm"
)

type Machine struct {
	BaseModel
	Name        string `gorm:"size:64;uniqueIndex;not null" json:"name"`
	IP          string `gorm:"size:45;not null" json:"ip"`
	MachineType string `gorm:"size:16;not null" json:"machineType"`
	CPU         string `gorm:"size:64;default:''" json:"cpu"`
	Memory      string `gorm:"size:32;default:''" json:"memory"`
	Disk        string `gorm:"size:64;default:''" json:"disk"`
	OS          string `gorm:"size:64;default:''" json:"os"`
	Status      int8   `gorm:"default:1" json:"status"`
	SSHPort     int    `gorm:"default:22" json:"sshPort"`
	SSHUser     string `gorm:"size:32;default:'root'" json:"sshUser"`
	SSHPassword string `gorm:"size:128" json:"sshPassword"`
	Remark      string `gorm:"type:text" json:"remark"`
}

func (Machine) TableName() string {
	return "machines"
}

func (m *Machine) AfterFind(tx *gorm.DB) error {
	if m.Remark == "" {
		m.Remark = ""
	}
	return nil
}
