package auth

import (
	"FileStorage/user"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"log"
	"net/http"
	"os"
	"strings"
)

var InvalidToken = errors.New("invalid token")

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

		if _, err := ParseToken(headerPart[1], []byte(os.Getenv("SIGNINGKEY"))); err != nil {
			status := http.StatusBadRequest
			if err == InvalidToken {
				status = http.StatusUnauthorized
			}
			c.AbortWithStatus(status)
			return
		}
	}
	return fn
}

func ParseToken(accessToken string, key []byte) ([]string, error) {
	token, err := jwt.ParseWithClaims(accessToken, &account.User{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signin method: %v", token.Header["alg"])
		}
		return key, nil
	})
	if err != nil {
		return []string{}, err
	}

	if claims, ok := token.Claims.(*account.User); ok && token.Valid {
		return []string{claims.Login, claims.Role}, nil
	}
	return []string{}, InvalidToken
}
