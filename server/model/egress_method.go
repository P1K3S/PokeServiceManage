package model

type EgressMethod struct {
	BaseModel
	UserID       uint   `gorm:"not null;index" json:"userId"`
	ServiceID       uint   `gorm:"not null;index" json:"serviceId"`
	ServiceType     string `gorm:"size:16;not null;default:'docker'" json:"serviceType"`
	EgressServiceID uint   `gorm:"index;default:0" json:"egressServiceId"`
	IsDirect        bool   `gorm:"default:false" json:"isDirect"`
	ProxyName       string `gorm:"size:64;default:''" json:"proxyName"`
	PublicIP        string `gorm:"size:45;not null" json:"publicIp"`
	PublicPort      int    `gorm:"not null" json:"publicPort"`
	InternalIP      string `gorm:"size:45;not null" json:"internalIp"`
	InternalPort    int    `gorm:"not null" json:"internalPort"`
	Protocol        string `gorm:"size:8;default:'TCP'" json:"protocol"`
	Status          int8   `gorm:"default:1" json:"status"`
	Remark          string `gorm:"type:text" json:"remark"`
}

func (EgressMethod) TableName() string {
	return "egress_methods"
}
