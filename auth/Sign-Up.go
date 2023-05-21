package auth

import (
	"FileStorage/storage"
	"FileStorage/user"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
)

func SignUpHandler() gin.HandlerFunc {
	fn := func(c *gin.Context) {
		var user account.User
		if ok := user.ParseCredentials(c); !ok {
			return
		}

		isExist, _ := user.Exist()
		if isExist {
			c.IndentedJSON(http.StatusOK, gin.H{"error": "user already exist"})
			return
		}
		if ok := user.CheckCredentials(c); !ok {
			c.IndentedJSON(http.StatusOK, gin.H{"error": "credentials does not meet requirements"})
			return
		}
		if err := storage.SetUser(user.Login, user.Password); err != nil {
			fmt.Println("aaa")
			c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}

		if err := os.Mkdir(os.Getenv("BASEDIR")+user.Login, os.ModePerm); err != nil {
			c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "error creating user directory"})
			return
		}
		c.IndentedJSON(http.StatusOK, gin.H{"message": "user created successfully"})
	}
	return fn
}
