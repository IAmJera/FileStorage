package general

import (
	"encoding/hex"
	"golang.org/x/crypto/sha3"
	"log"
	"os"
)

type Closer interface {
	Close() error
}

func CloseFile(c Closer) {
	if err := c.Close(); err != nil {
		log.Printf("CloseFile: %s", err)
	}
	return
}

func Hash(passwd string) string {
	pwd := sha3.New256()
	pwd.Write([]byte(passwd))
	pwd.Write([]byte(os.Getenv("SALT")))
	return hex.EncodeToString(pwd.Sum(nil))
}
