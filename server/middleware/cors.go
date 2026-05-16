package middleware

import (
	"service-manage/config"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func CORS() gin.HandlerFunc {
	origins := config.AppConfig.Cors.Origins()
	allowAll := config.AppConfig.Cors.AllowAllOrigins()

	if allowAll || len(origins) == 0 {
		return cors.New(cors.Config{
			AllowAllOrigins: true,
			AllowMethods:    []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowHeaders:    []string{"Content-Type", "Authorization"},
			ExposeHeaders:   []string{"Content-Length"},
		})
	}

	return cors.New(cors.Config{
		AllowOrigins:  origins,
		AllowMethods:  []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:  []string{"Content-Type", "Authorization"},
		ExposeHeaders: []string{"Content-Length"},
	})
}
