// Package auth defines the functions responsible for auth
package auth

import (
	"FileStorage/app/general"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"os"
	"strings"
)

// Middleware verifies the token and authorizes the user
func Middleware() gin.HandlerFunc {
	fn := func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" {
			log.Printf("Middleware: empty header")
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		headerPart := strings.Split(header, " ")
		if len(headerPart) != 2 || headerPart[0] != "Bearer" {
			log.Printf("Middleware: wrong header")
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		if _, err := general.ParseToken(headerPart[1], []byte(os.Getenv("SIGNINGKEY"))); err != nil {
			status := http.StatusBadRequest
			if err == general.ErrInvalidToken {
				status = http.StatusUnauthorized
			}
			c.AbortWithStatus(status)
			return
		}
	}
	return fn
}
