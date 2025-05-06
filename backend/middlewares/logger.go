package middlewares

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"launay-dot-one/utils"
)

func Logger() gin.HandlerFunc {
	logger := utils.GetLogger()
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		c.Next()

		status := c.Writer.Status()
		latency := time.Since(start)

		logger.WithFields(logrus.Fields{
			"method":  method,
			"path":    path,
			"status":  status,
			"latency": latency,
		}).Info("Request handled")
	}
}
