package handler

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func getUserId(c *gin.Context) uint {
	id, _ := c.Get("userId")
	if id == nil {
		return 0
	}
	return id.(uint)
}

func getUsername(c *gin.Context) string {
	username, _ := c.Get("username")
	if username == nil {
		return ""
	}
	return username.(string)
}

func getLogUserInfo(c *gin.Context) (uint, string) {
	if uid := getUserId(c); uid > 0 {
		return uid, getUsername(c)
	}
	return 0, "system"
}

func isAdmin(c *gin.Context) bool {
	role, _ := c.Get("role")
	return role == "super_admin"
}

func userScope(c *gin.Context, db *gorm.DB) *gorm.DB {
	userId := getUserId(c)
	if userId == 0 {
		return db
	}
	if isAdmin(c) {
		return db
	}
	return db.Where("user_id = ?", userId)
}

func serviceScope(c *gin.Context, db *gorm.DB) *gorm.DB {
	if isAdmin(c) {
		return db
	}
	userId := getUserId(c)
	if userId == 0 {
		return db.Where("is_public = 1")
	}
	return db.Where("user_id = ? OR is_public = 1", userId)
}
