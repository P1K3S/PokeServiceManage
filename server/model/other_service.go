package model

type OtherService struct {
	BaseModel
	UserID    uint   `gorm:"not null;index" json:"userId"`
	MachineID uint   `gorm:"not null;index" json:"machineId"`
	Name      string `gorm:"size:128;not null" json:"name"`
	Port      int    `gorm:"not null" json:"port"`
	Protocol  string `gorm:"size:8;default:'TCP'" json:"protocol"`
	Status    int8   `gorm:"default:1" json:"status"`
	IsPublic  bool   `gorm:"default:false" json:"isPublic"`
	Remark    string `gorm:"type:text" json:"remark"`
}

func (OtherService) TableName() string {
	return "other_services"
}
