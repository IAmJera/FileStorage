package account

import (
	"FileStorage/app/general"
	"FileStorage/storage"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"log"
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
)

func (user *User) Exist() (bool, bool) { // isExist, sameHash
	sameHash := false
	passwd, err := storage.GetUser(user.Login)
	if err != nil || passwd == "" {
		return false, false
	}

	if passwd == general.Hash(user.Password) {
		sameHash = true
	}
	return true, sameHash
}

func (user *User) CheckCredentials() bool {
	if len(user.Login) < loginMin && len(user.Login) > loginMax {
		return false
	}
	if len(user.Password) < passMin {
		return false
	}
	return true
}

func (user *User) ParseCredentials(c *gin.Context) bool {
	if err := c.Request.ParseForm(); err != nil {
		log.Printf("ParseCredentials: %s", err)
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
