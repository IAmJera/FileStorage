// Package general defines general functions
package general

import (
	"encoding/hex"
	"github.com/minio/minio-go"
	"golang.org/x/crypto/sha3"
	"io"
	"log"
)

type S3 interface {
	GetObject(bucketName string, objectName string, opts minio.GetObjectOptions) (*minio.Object, error)
	MakeBucket(bucketName string, location string) (err error)
	RemoveBucket(bucketName string) error
	RemoveObject(bucketName string, objectName string) error
	PutObject(bucketName string, objectName string, reader io.Reader, objectSize int64, opts minio.PutObjectOptions) (n int64, err error)
	FPutObject(bucketName string, objectName string, filePath string, opts minio.PutObjectOptions) (n int64, err error)
	ListObjects(bucketName string, objectPrefix string, recursive bool, doneCh <-chan struct{}) <-chan minio.ObjectInfo
}

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
func Hash(passwd, salt string) string {
	pwd := sha3.New256()
	pwd.Write([]byte(passwd))
	pwd.Write([]byte(salt))
	return hex.EncodeToString(pwd.Sum(nil))
}

// GetS3Objects gets S3 objects by prefix and returns a slice with them
func GetS3Objects(s3 S3, login, prefix string, recursive bool) []string {
	var filesList []string
	for message := range s3.ListObjects(login, prefix, recursive, make(chan struct{})) {
		filesList = append(filesList, message.Key)
	}
	return filesList
}
