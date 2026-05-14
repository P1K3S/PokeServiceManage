package model

import "gorm.io/gorm"

type EgressMethod struct {
	BaseModel
	ServiceID    uint   `gorm:"not null;index" json:"serviceId"`
	MethodType   string `gorm:"size:32;not null" json:"methodType"`
	ProxyName    string `gorm:"size:64;default:''" json:"proxyName"`
	PublicIP     string `gorm:"size:45;not null" json:"publicIp"`
	PublicPort   int    `gorm:"not null" json:"publicPort"`
	InternalIP   string `gorm:"size:45;not null" json:"internalIp"`
	InternalPort int    `gorm:"not null" json:"internalPort"`
	Protocol     string `gorm:"size:8;default:'TCP'" json:"protocol"`
	Status       int8   `gorm:"default:1" json:"status"`
	Remark       string `gorm:"type:text" json:"remark"`

	DockerService DockerService `gorm:"foreignKey:ServiceID" json:"dockerService,omitempty"`
}

func (EgressMethod) TableName() string {
	return "egress_methods"
}

func (e *EgressMethod) AfterFind(tx *gorm.DB) error {
	if e.Remark == "" {
		e.Remark = ""
	}
	return nil
}