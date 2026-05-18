package model

type Notice struct {
	BaseModel
	Title     string `gorm:"size:128" json:"title"`
	Content   string `gorm:"type:text" json:"content"`
	Status    int8   `gorm:"default:1" json:"status"`
	SortOrder int    `gorm:"default:0" json:"sortOrder"`
	Pinned    bool   `gorm:"default:false" json:"pinned"`
}

func (Notice) TableName() string {
	return "notices"
}
