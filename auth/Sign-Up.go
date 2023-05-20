package auth

import (
	"FileStorage/storage"
	account "FileStorage/user"
	"github.com/bradfitz/gomemcache/memcache"
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
)

var cache = memcache.New("localhost:11211")

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
		if err := storage.SetUser(user); err != nil {
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
