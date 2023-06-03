// Package general defines general functions
package general

import (
	"encoding/hex"
	"golang.org/x/crypto/sha3"
	"log"
	"os"
)

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
