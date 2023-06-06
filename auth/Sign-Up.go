// Package auth defines the functions responsible for auth
package auth

import (
	"FileStorage/api"
	"FileStorage/app/general"
	"FileStorage/user"
	"context"
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
	"strings"
)

// SignUpHandler registers the user by writing his data to the database
func SignUpHandler(conn api.AuthClient) gin.HandlerFunc {
	fn := func(c *gin.Context) {
		var usr user.User
		if ok := usr.ParseCredentials(c); !ok {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		if ok := usr.CheckCredentials(); !ok {
			c.IndentedJSON(http.StatusOK, gin.H{"error": "credentials does not meet requirements"})
			return
		}

		u := api.User{Login: usr.Login, Password: general.Hash(usr.Password), Role: usr.Role}
		if _, err := conn.AddUser(context.Background(), &u); err != nil {
			if strings.Contains(err.Error(), "user already exist") {
				c.IndentedJSON(http.StatusOK, gin.H{"error": "user already exist"})
				return
			}
			c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": strings.Split(err.Error(), "= ")[2]})
			return
		}

		if err := os.Mkdir(os.Getenv("BASEDIR")+usr.Login, os.ModePerm); err != nil {
			c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "error creating user directory"})
			return
		}
		c.IndentedJSON(http.StatusCreated, gin.H{"message": "user created successfully"})
	}
	return fn
}
