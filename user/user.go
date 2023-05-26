package account

import (
	"FileStorage/app/general"
	"FileStorage/storage"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"net/http"
	"os"
	"strings"
	"time"
)

type User struct {
	jwt.StandardClaims
	Login    string `json:"login"`
	Password string `json:"password"`
	Role     string
}

var (
	loginMin = 4
	loginMax = 30
	passMin  = 10
	passMax  = 255
)

func (user *User) Exist() (bool, bool) { // isExist, sameHash
	sameHash := false
	passwd, err := storage.GetUser(user.Login)
	if err != nil || passwd == "" {
		return false, false
	}

	if passwd == string(general.Hash(user.Password)) {
		sameHash = true
	}
	return true, sameHash
}

func (user *User) CheckCredentials(c *gin.Context) bool {
	if len(user.Login) < loginMin && len(user.Login) > loginMax {
		c.IndentedJSON(http.StatusOK, gin.H{"error": "login must be minimum 4 characters"})
		return false
	}
	if len(user.Password) < passMin && len(user.Password) > passMax {
		c.IndentedJSON(http.StatusOK, gin.H{"error": "password must be minimum 10 characters"})
		return false
	}
	return true
}

func (user *User) ParseCredentials(c *gin.Context) bool {
	if err := c.Request.ParseForm(); err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err})
		return false
	}

	for key, val := range c.Request.PostForm {
		switch key {
		case "login":
			user.Login = strings.Join(val, "")
		case "password":
			user.Password = strings.Join(val, "")
		}
	}
	return true
}

func (user *User) SignIn() (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &User{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 24).Unix(),
			IssuedAt:  time.Now().Unix(),
		},
		Login: user.Login,
		Role:  "user",
	})
	return token.SignedString([]byte(os.Getenv("SIGNINGKEY")))
}
