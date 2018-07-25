package rest

import (
	"github.com/gin-gonic/gin"
	"os"
)

func TokenAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !gin.IsDebugging() {
			token := c.Request.FormValue("api_token")

			if token == "" {
				c.AbortWithStatus(401)
				return
			}

			if token != os.Getenv("API_TOKEN") {
				c.AbortWithStatus(401)
				return
			}
			c.Next()
		}

	}
}
