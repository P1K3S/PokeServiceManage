package model

type OperationLog struct {
	BaseModel
	UserID   uint   `gorm:"not null;index" json:"userId"`
	Username string `gorm:"size:64;not null" json:"username"`
	Action   string `gorm:"size:32;not null;index" json:"action"`
	Target   string `gorm:"size:32;not null;index" json:"target"`
	TargetID uint   `gorm:"index" json:"targetId"`
	Detail   string `gorm:"type:text" json:"detail"`
}

func (OperationLog) TableName() string {
	return "operation_logs"
}
