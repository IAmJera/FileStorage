package account

import (
	"FileStorage/storage"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/sha3"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type User struct {
	jwt.StandardClaims
	Login    string `json:"login"`
	Password string `json:"password"`
	Role     string
}

func (user *User) Exist() (bool, bool) { // isExist, sameHash
	sameHash := false
	usr, err := storage.GetUser(user.Login)
	if err != nil || usr.Password == "" {
		return false, false
	}

	if usr.Password == string(Hash(user.Password)) {
		sameHash = true
	}
	return true, sameHash
}

func (user *User) CheckCredentials(c *gin.Context) bool {
	loginMin, _ := strconv.Atoi(os.Getenv("LOGINMINLEN"))
	if len(user.Login) < loginMin {
		c.IndentedJSON(http.StatusOK, gin.H{"error": "login must be minimum 4 characters"})
		return false
	}
	passMin, _ := strconv.Atoi(os.Getenv("PASSMINLEN"))
	if len(user.Password) < passMin {
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

func Hash(passwd string) []byte {
	pwd := sha3.New256()
	pwd.Write([]byte(passwd))
	pwd.Write([]byte(os.Getenv("SALT")))
	return pwd.Sum(nil)
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
