package model

type User struct {
	BaseModel
	Username string `gorm:"size:64;uniqueIndex;not null" json:"username"`
	Password string `gorm:"size:128;not null" json:"-"`
	Role     string `gorm:"size:16;default:'user';not null" json:"role"`
	Status   int8   `gorm:"default:1" json:"status"`
}

func (User) TableName() string {
	return "users"
}

const (
	RoleSuperAdmin = "super_admin"
	RoleUser       = "user"
)