// Package auth defines the functions responsible for auth
package auth

import (
	"FileStorage/app/general"
	"FileStorage/storage"
	"FileStorage/user"
	"github.com/gin-gonic/gin"
	"net/http"
)

// SignInHandler authenticate the user and returns a jwt token to him
func SignInHandler(storages storage.Storage) gin.HandlerFunc {
	fn := func(c *gin.Context) {
		var usr user.User
		if ok := usr.ParseCredentials(c); !ok {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		isExist, sameHash := usr.Exist(storages)
		if !isExist {
			c.IndentedJSON(http.StatusOK, gin.H{"error": "user does not exist"})
			return
		}

		if sameHash {
			usr.Password = general.Hash(usr.Password)
			token, err := usr.SignIn()
			if err != nil {
				c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err})
				return
			}
			c.IndentedJSON(http.StatusOK, gin.H{"token": token})
			return
		}
		c.IndentedJSON(http.StatusOK, gin.H{"error": "incorrect password"})
	}
	return fn
}
