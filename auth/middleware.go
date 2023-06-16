// Package auth defines the functions responsible for auth
package auth

import (
	"FileStorage/api"
	"FileStorage/user"
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"net/http"
	"strings"
)

// ErrInvalidToken defines the error of the invalid token
var ErrInvalidToken = errors.New("invalid token")

// Middleware verifies the token and authorizes the user
func Middleware(rpc api.AuthClient, secret *[]byte) gin.HandlerFunc {
	fn := func(c *gin.Context) {
		status := http.StatusInternalServerError
		header := c.GetHeader("Authorization")
		if header == "" {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		headerPart := strings.Split(header, " ")
		if len(headerPart) != 2 || headerPart[0] != "Bearer" {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		token, err := ParseToken(headerPart[1], secret)
		if err != nil {
			status = http.StatusBadRequest
			if err == ErrInvalidToken {
				status = http.StatusUnauthorized
			}
			c.AbortWithStatus(status)
			return
		}
		resp, _ := rpc.UserExist(context.Background(), &api.User{Login: token[0]})
		if resp != nil {
			if resp.Value {
				return
			}
			status = http.StatusUnauthorized
		}
		c.AbortWithStatus(status)
	}
	return fn
}

// ParseToken parses the token and returns the fields of the custom structure
func ParseToken(accessToken string, key *[]byte) ([]string, error) {
	token, err := jwt.ParseWithClaims(accessToken, &user.User{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signin method: %v", token.Header["alg"])
		}
		return *key, nil
	})
	if err != nil {
		return []string{}, err
	}

	if claims, ok := token.Claims.(*user.User); ok && token.Valid {
		return []string{claims.Login, claims.Role}, nil
	}
	return []string{}, ErrInvalidToken
}
