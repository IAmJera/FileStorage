package general

import (
	"golang.org/x/crypto/sha3"
	"log"
	"os"
)

type Closer interface {
	Close() error
}

func CloseFile(c Closer) {
	if err := c.Close(); err != nil {
		log.Println("error occurred: ", err)
	}
	return
}

func Hash(passwd string) []byte {
	pwd := sha3.New256()
	pwd.Write([]byte(passwd))
	pwd.Write([]byte(os.Getenv("SALT")))
	return pwd.Sum(nil)
}
