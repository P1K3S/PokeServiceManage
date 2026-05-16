package middleware

import (
	"service-manage/config"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func CORS() gin.HandlerFunc {
	cfg := cors.DefaultConfig()
	if config.AppConfig.Cors.AllowAllOrigins() {
		cfg.AllowAllOrigins = true
	} else {
		cfg.AllowOrigins = config.AppConfig.Cors.Origins()
	}
	cfg.AllowHeaders = []string{"Content-Type", "Authorization"}
	return cors.New(cfg)
}
