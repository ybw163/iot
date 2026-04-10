package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"iot/pkg/logger"
)

// Logger 请求日志中间件
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		fields := []zap.Field{
			zap.Int("status", status),
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.String("query", query),
			zap.String("ip", c.ClientIP()),
			zap.Duration("latency", latency),
		}

		if len(c.Errors) > 0 {
			logger.Log.Error(c.Errors.ByType(gin.ErrorTypePrivate).String(), fields...)
		} else {
			logger.Log.Info("request", fields...)
		}
	}
}

// CORS 跨域中间件
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,PATCH,OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin,Content-Type,Accept,Authorization,Token,X-Requested-With")
		c.Header("Access-Control-Expose-Headers", "Content-Length,Access-Control-Allow-Origin,Access-Control-Allow-Headers")
		c.Header("Access-Control-Max-Age", "86400")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// Recovery 异常恢复中间件
func Recovery() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, err interface{}) {
		logger.Log.Error("panic recovered", zap.Any("error", err))
		c.AbortWithStatusJSON(500, gin.H{
			"code":    500,
			"message": "internal server error",
		})
	})
}
