package router

import (
	"github.com/gin-gonic/gin"

	"iot/internal/handler"
	"iot/internal/middleware"
)

type Router struct {
	deviceHandler *handler.DeviceHandler
}

func NewRouter(deviceHandler *handler.DeviceHandler) *Router {
	return &Router{
		deviceHandler: deviceHandler,
	}
}

func (r *Router) Setup(engine *gin.Engine) {
	// 全局中间件
	engine.Use(middleware.Recovery())
	engine.Use(middleware.Logger())
	engine.Use(middleware.CORS())

	// 健康检查
	engine.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// API v1
	v1 := engine.Group("/api/v1")
	{
		// 设备管理
		devices := v1.Group("/devices")
		{
			devices.POST("", r.deviceHandler.CreateDevice)
			devices.GET("", r.deviceHandler.ListDevices)
			devices.GET("/:id", r.deviceHandler.GetDevice)
			devices.PUT("/:id", r.deviceHandler.UpdateDevice)
			devices.DELETE("/:id", r.deviceHandler.DeleteDevice)
		}
	}
}
