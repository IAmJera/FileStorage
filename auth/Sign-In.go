// Package auth defines the functions responsible for auth
package auth

import (
	"FileStorage/api"
	"FileStorage/app/general"
	"FileStorage/user"
	"context"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

// SignInHandler authenticate the user and returns a jwt token to him
func SignInHandler(conn api.AuthClient) gin.HandlerFunc {
	fn := func(c *gin.Context) {
		var usr user.User
		if ok := usr.ParseCredentials(c); !ok {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
		u := api.User{Login: usr.Login, Role: usr.Role, Password: general.Hash(usr.Password)}
		token, err := conn.SignIn(context.Background(), &u)
		if err != nil {
			switch {
			case strings.Contains(err.Error(), "user does not exist"):
				c.IndentedJSON(http.StatusOK, gin.H{"error": "user does not exist"})
				return
			case strings.Contains(err.Error(), "incorrect password"):
				c.IndentedJSON(http.StatusOK, gin.H{"error": "incorrect password"})
				return
			default:
				c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": strings.Split(err.Error(), "= ")[2]})
				return
			}
		}
		c.IndentedJSON(http.StatusOK, gin.H{"token": token.Token})
		return
	}
	return fn
}
