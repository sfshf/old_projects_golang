package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
)

// Cors handler
func Download() gin.HandlerFunc {
	return func(c *gin.Context) {
		// fmt.Println("download mDownloadiddleware")
		if strings.HasPrefix(c.Request.URL.Path, "/download/") {
			c.Header("Content-Disposition", "attachment")
			// fmt.Println("method", c.Request.URL.Path)
		}

		c.Next()
	}
}
