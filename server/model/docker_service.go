package model

type DockerService struct {
	BaseModel
	MachineID        uint   `gorm:"not null;index" json:"machineId"`
	Name             string `gorm:"size:64;not null" json:"name"`
	Port             int    `gorm:"not null" json:"port"`
	Protocol         string `gorm:"size:8;default:'TCP'" json:"protocol"`
	DockerSourceIP   string `gorm:"size:45" json:"dockerSourceIp"`
	DockerSourcePort int    `gorm:"default:0" json:"dockerSourcePort"`
	PortMappings     string `gorm:"type:text" json:"portMappings"`
	Status           int8   `gorm:"default:1" json:"status"`
	Locked           bool   `gorm:"default:false" json:"locked"`
	IsEgress         bool   `gorm:"default:false" json:"isEgress"`
	Remark           string `gorm:"type:text" json:"remark"`

	Machine Machine `gorm:"foreignKey:MachineID" json:"machine,omitempty"`
}

func (DockerService) TableName() string {
	return "docker_services"
}