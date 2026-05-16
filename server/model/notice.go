package model

type Notice struct {
	BaseModel
	Title   string `gorm:"size:128" json:"title"`
	Content string `gorm:"type:text" json:"content"`
	Status  int8   `gorm:"default:1" json:"status"`
}

func (Notice) TableName() string {
	return "notices"
}
