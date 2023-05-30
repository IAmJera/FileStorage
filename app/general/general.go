// Package general defines general functions
package general

import (
	"FileStorage/user"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/sha3"
	"log"
	"os"
)

// ErrInvalidToken defines the error of the invalid token
var ErrInvalidToken = errors.New("invalid token")

// Closer defines the interface to which all objects with the Close method correspond
type Closer interface {
	Close() error
}

// CloseFile closes the object that satisfies the Closer interface
func CloseFile(c Closer) {
	if err := c.Close(); err != nil {
		log.Printf("CloseFile: %s", err)
	}
	return
}

// Hash hashes a given string with the addition of salt
func Hash(passwd string) string {
	pwd := sha3.New256()
	pwd.Write([]byte(passwd))
	pwd.Write([]byte(os.Getenv("SALT")))
	return hex.EncodeToString(pwd.Sum(nil))
}

// ParseToken parses the token and returns the fields of the custom structure
func ParseToken(accessToken string, key []byte) ([]string, error) {
	token, err := jwt.ParseWithClaims(accessToken, &user.User{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signin method: %v", token.Header["alg"])
		}
		return key, nil
	})
	if err != nil {
		return []string{}, err
	}

	if claims, ok := token.Claims.(*user.User); ok && token.Valid {
		return []string{claims.Login, claims.Role}, nil
	}
	return []string{}, ErrInvalidToken
}
