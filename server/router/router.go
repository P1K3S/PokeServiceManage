package router

import (
	"service-manage/handler"
	"service-manage/middleware"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupRouter(db *gorm.DB) *gin.Engine {
	r := gin.Default()
	r.Use(middleware.CORS())

	api := r.Group("/api")
	{
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
		}
	}

	return r
}