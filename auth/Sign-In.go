package auth

import (
	"FileStorage/app/general"
	"FileStorage/user"
	"github.com/gin-gonic/gin"
	"net/http"
)

func SignInHandler() gin.HandlerFunc {
	fn := func(c *gin.Context) {
		var user account.User
		if ok := user.ParseCredentials(c); !ok {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		isExist, sameHash := user.Exist()
		if !isExist {
			c.IndentedJSON(http.StatusOK, gin.H{"error": "user does not exist"})
			return
		}

		if sameHash {
			user.Password = general.Hash(user.Password)
			token, err := user.SignIn()
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
