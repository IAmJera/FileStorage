// Package user defines the structure and methods of the user
package user

import (
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"log"
	"os"
	"strings"
	"time"
)

// User defines user structure
type User struct {
	jwt.StandardClaims
	Login    string `json:"login"`
	Password string `json:"password"`
	Role     string //Not used. For the future
}

var (
	loginMin = 4
	loginMax = 20
	passMin  = 10
)

// CheckCredentials Checks whether the user data corresponds to the requirements
func (user *User) CheckCredentials() bool {
	if len(user.Login) < loginMin || len(user.Login) > loginMax {
		return false
	}
	if len(user.Password) < passMin {
		return false
	}
	return true
}

// ParseCredentials parses the query and fills in the user structure
func (user *User) ParseCredentials(c *gin.Context) bool {
	if err := c.Request.ParseForm(); err != nil {
		log.Printf("ParseCredentials: %s", err)
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

// SignIn creates and returns a jwt token
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
