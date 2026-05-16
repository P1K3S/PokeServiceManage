package middleware

import (
	"service-manage/config"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func CORS() gin.HandlerFunc {
	cfg := cors.Config{
		AllowMethods:  []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:  []string{"Content-Type", "Authorization"},
		ExposeHeaders: []string{"Content-Length"},
	}
	if config.AppConfig.Cors.AllowAllOrigins() {
		cfg.AllowAllOrigins = true
	} else {
		cfg.AllowOrigins = config.AppConfig.Cors.Origins()
	}
	return cors.New(cfg)
}
