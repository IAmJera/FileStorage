package general

import "log"

type Closer interface {
	Close() error
}

func CloseFile(c Closer) {
	if err := c.Close(); err != nil {
		log.Println("error occurred: ", err)
	}
	return
}
