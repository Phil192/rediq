package rest

import (
	"github.com/gin-gonic/gin"
	"os"
	"fmt"
)

func TokenAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		//if !gin.IsDebugging() {
		fmt.Println("!!!!!!!!!!!!!!")
			token := c.Request.Header.Get("token")
			if token == "" {
				c.AbortWithStatus(401)
				return
			}
			if os.Getenv("token") != token {
				c.AbortWithStatus(401)
				return
			}
			c.Next()
		//}

	}
}

