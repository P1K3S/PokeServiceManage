package handler

import (
	"service-manage/model"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ConfigHandler struct {
	DB *gorm.DB
}

func NewConfigHandler(db *gorm.DB) *ConfigHandler {
	return &ConfigHandler{DB: db}
}

type ExportData struct {
	Machines       []model.Machine       `json:"machines"`
	DockerServices []model.DockerService `json:"dockerServices"`
	OtherServices  []model.OtherService  `json:"otherServices"`
	EgressMethods  []model.EgressMethod  `json:"egressMethods"`
}

func (h *ConfigHandler) Export(c *gin.Context) {
	uid := getUserId(c)

	var data ExportData

	if isAdmin(c) {
		h.DB.Find(&data.Machines)
		h.DB.Find(&data.DockerServices)
		h.DB.Find(&data.OtherServices)
		h.DB.Find(&data.EgressMethods)
	} else {
		h.DB.Where("user_id = ?", uid).Find(&data.Machines)
		h.DB.Where("user_id = ? OR is_public = 1", uid).Find(&data.DockerServices)
		h.DB.Where("user_id = ? OR is_public = 1", uid).Find(&data.OtherServices)
		h.DB.Where("user_id = ?", uid).Find(&data.EgressMethods)
	}

	for i := range data.Machines {
		data.Machines[i].SSHPassword = "******"
	}

	c.Header("Content-Disposition", "attachment; filename=service-config.json")
	c.JSON(200, gin.H{"code": 0, "data": data})
}

func (h *ConfigHandler) Import(c *gin.Context) {
	if !isAdmin(c) {
		jsonError(c, "仅管理员可导入配置")
		return
	}

	var data ExportData
	if err := c.ShouldBindJSON(&data); err != nil {
		jsonError(c, "请求数据格式错误")
		return
	}

	tx := h.DB.Begin()

	for i := range data.Machines {
		data.Machines[i].ID = 0
		data.Machines[i].UserID = getUserId(c)
		if data.Machines[i].SSHPassword == "******" {
			data.Machines[i].SSHPassword = ""
		}
		if err := tx.Create(&data.Machines[i]).Error; err != nil {
			tx.Rollback()
			jsonError(c, "导入主机失败: "+err.Error())
			return
		}
	}

	for i := range data.DockerServices {
		data.DockerServices[i].ID = 0
		data.DockerServices[i].UserID = getUserId(c)
		if err := tx.Create(&data.DockerServices[i]).Error; err != nil {
			tx.Rollback()
			jsonError(c, "导入Docker服务失败: "+err.Error())
			return
		}
	}

	for i := range data.OtherServices {
		data.OtherServices[i].ID = 0
		data.OtherServices[i].UserID = getUserId(c)
		if err := tx.Create(&data.OtherServices[i]).Error; err != nil {
			tx.Rollback()
			jsonError(c, "导入其他服务失败: "+err.Error())
			return
		}
	}

	for i := range data.EgressMethods {
		data.EgressMethods[i].ID = 0
		data.EgressMethods[i].UserID = getUserId(c)
		if err := tx.Create(&data.EgressMethods[i]).Error; err != nil {
			tx.Rollback()
			jsonError(c, "导入出站方式失败: "+err.Error())
			return
		}
	}

	tx.Commit()

	uid, uname := getLogUserInfo(c)
	logOperation(h.DB, uid, uname, "import", "config", 0, "配置导入")

	jsonSuccess(c, gin.H{
		"machines":       len(data.Machines),
		"dockerServices": len(data.DockerServices),
		"otherServices":  len(data.OtherServices),
		"egressMethods":  len(data.EgressMethods),
	})
}
