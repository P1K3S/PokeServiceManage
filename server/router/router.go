package router

import (
	"net/http"
	"os"
	"path/filepath"
	"service-manage/handler"
	"service-manage/middleware"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupRouter(db *gorm.DB) *gin.Engine {
	r := gin.Default()
	r.Use(middleware.CORS())

	authHandler := handler.NewAuthHandler(db)
	r.POST("/api/login", authHandler.Login)
	r.POST("/api/register", authHandler.Register)

	api := r.Group("/api")
	api.Use(middleware.Auth())
	{
		api.GET("/user-info", authHandler.GetUserInfo)

		overviewHandler := handler.NewMachineHandler(db)
		api.GET("/overview", overviewHandler.Overview)

		machineHandler := handler.NewMachineHandler(db)
		machines := api.Group("/machines")
		{
			machines.GET("", machineHandler.List)
			machines.GET("/:id", machineHandler.Get)
			machines.POST("", machineHandler.Create)
			machines.PUT("/:id", machineHandler.Update)
			machines.DELETE("/:id", machineHandler.Delete)
			machines.POST("/:id/check-ssh", machineHandler.CheckSSH)
			machines.POST("/:id/discover-services", machineHandler.DiscoverServices)
		}

		sshHandler := handler.NewSSHTerminalHandler(db)
		api.GET("/ssh-terminal/:id", sshHandler.HandleTerminal)

		sftpHandler := handler.NewSFTPHandler(db)
		sftp := api.Group("/sftp")
		{
			sftp.GET("/:id/list", sftpHandler.List)
			sftp.GET("/:id/download", sftpHandler.Download)
			sftp.GET("/:id/download-dir", sftpHandler.DownloadDir)
			sftp.POST("/:id/upload", sftpHandler.Upload)
			sftp.POST("/:id/mkdir", sftpHandler.Mkdir)
			sftp.DELETE("/:id/remove", sftpHandler.Remove)
			sftp.PUT("/:id/rename", sftpHandler.Rename)
			sftp.GET("/:id/read", sftpHandler.ReadFile)
			sftp.POST("/:id/write", sftpHandler.WriteFile)
			sftp.GET("/:id/stat", sftpHandler.Stat)
		}

		dockerServiceHandler := handler.NewDockerServiceHandler(db)
		dockerServices := api.Group("/docker-services")
		{
			dockerServices.GET("", dockerServiceHandler.List)
			dockerServices.POST("", dockerServiceHandler.Create)
			dockerServices.PUT("/:id", dockerServiceHandler.Update)
			dockerServices.DELETE("/:id", dockerServiceHandler.Delete)
			dockerServices.POST("/:id/check", dockerServiceHandler.Check)
		}

		otherServiceHandler := handler.NewOtherServiceHandler(db)
		otherServices := api.Group("/other-services")
		{
			otherServices.GET("", otherServiceHandler.List)
			otherServices.POST("", otherServiceHandler.Create)
			otherServices.PUT("/:id", otherServiceHandler.Update)
			otherServices.DELETE("/:id", otherServiceHandler.Delete)
		}

		egressHandler := handler.NewEgressMethodHandler(db)
		egress := api.Group("/egress-methods")
		{
			egress.GET("", egressHandler.List)
			egress.POST("", egressHandler.Create)
			egress.PUT("/:id", egressHandler.Update)
			egress.DELETE("/:id", egressHandler.Delete)
			egress.POST("/sync-firewall", egressHandler.SyncFirewall)
			egress.POST("/generate-frpc", egressHandler.GenerateFrpc)
			egress.PUT("/batch-status", egressHandler.BatchUpdateStatus)
			egress.DELETE("/batch", egressHandler.BatchDelete)
			egress.GET("/health-check", egressHandler.HealthCheck)
		}

		noticeHandler := handler.NewNoticeHandler(db)
		notices := api.Group("/notices")
		{
			notices.GET("", noticeHandler.GetNotice)
			notices.PUT("", noticeHandler.UpdateNotice)
		}

		logHandler := handler.NewOperationLogHandler(db)
		logs := api.Group("/operation-logs")
		{
			logs.GET("", logHandler.List)
		}

		configHandler := handler.NewConfigHandler(db)
		config := api.Group("/config")
		{
			config.GET("/export", configHandler.Export)
			config.POST("/import", configHandler.Import)
		}
	}

	distPath := "dist"
	if _, err := os.Stat(distPath); err == nil {
		r.NoRoute(func(c *gin.Context) {
			if strings.HasPrefix(c.Request.URL.Path, "/api") {
				c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "接口不存在"})
				return
			}
			c.File(filepath.Join(distPath, "index.html"))
		})

		r.Use(func(c *gin.Context) {
			if strings.HasPrefix(c.Request.URL.Path, "/api") {
				c.Next()
				return
			}
			filePath := filepath.Join(distPath, c.Request.URL.Path)
			if s, err := os.Stat(filePath); err == nil && !s.IsDir() {
				c.File(filePath)
				return
			}
			c.File(filepath.Join(distPath, "index.html"))
		})
	}

	return r
}
