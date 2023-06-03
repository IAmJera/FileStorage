// Package auth defines the functions responsible for auth
package auth

import (
	"FileStorage/app/general"
	"FileStorage/storage"
	"FileStorage/user"
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
	"strings"
)

// SignUpHandler registers the user by writing his data to the database
func SignUpHandler() gin.HandlerFunc {
	fn := func(c *gin.Context) {
		var usr user.User
		if ok := usr.ParseCredentials(c); !ok {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		//isExist, _ := usr.Exist()
		//if isExist {
		//	c.IndentedJSON(http.StatusOK, gin.H{"error": "user already exist"})
		//	return
		//}
		if ok := usr.CheckCredentials(); !ok {
			c.IndentedJSON(http.StatusOK, gin.H{"error": "credentials does not meet requirements"})
			return
		}
		if err := storage.SetUser(usr.Login, general.Hash(usr.Password)); err != nil {
			if strings.Contains(err.Error(), "duplicate key value violates unique constraint \"users_login_key\"") {
				c.IndentedJSON(http.StatusOK, gin.H{"error": "user already exist"})
				return
			}
			c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err})
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
