// Package handlers determines the handlers of the application
package handlers

import (
	"FileStorage/app/general"
	"FileStorage/auth"
	"github.com/gin-gonic/gin"
	"github.com/minio/minio-go"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

var counter uint8 = 0

// DeleteObjectHandler sends a request to delete the user's file
func DeleteObjectHandler(s3 general.S3, secret *[]byte) gin.HandlerFunc {
	fn := func(c *gin.Context) {
		token, err := auth.ParseToken(strings.Split(c.GetHeader("Authorization"), " ")[1], secret)
		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}

		object := c.PostForm("objectpath")
		filesList := general.GetS3Objects(s3, token[0], object, true)
		for _, obj := range filesList {
			if err = s3.RemoveObject(token[0], obj); err != nil {
				c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err})
				return
			}
		}
		c.IndentedJSON(http.StatusOK, gin.H{"message": "remove successful"})
	}
	return fn
}

// DownloadFileHandler sends a request to download the user's file
func DownloadFileHandler(s3 general.S3, secret *[]byte) gin.HandlerFunc {
	fn := func(c *gin.Context) {
		token, err := auth.ParseToken(strings.Split(c.GetHeader("Authorization"), " ")[1], secret)
		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}
		objectPath := c.PostForm("objectpath")

		var byteFile []byte
		object, err := s3.GetObject(token[0], objectPath, minio.GetObjectOptions{})
		if err != nil {
			log.Printf("DownloadFileHandler:GetObject: %s", err)
			c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}
		defer general.CloseFile(object)

		for {
			chunk := make([]byte, 1024)
			n, err := object.Read(chunk)
			if err != nil {
				break
			}
			byteFile = append(byteFile, chunk[:n]...)
		}

		c.Header("Content-Disposition", "attachment; filename="+c.Param("file"))
		c.Data(http.StatusOK, "application/octet-stream", byteFile)
	}
	return fn
}

// CreateDirHandler creates dir or return error
func CreateDirHandler(s3 general.S3, secret *[]byte) gin.HandlerFunc {
	fn := func(c *gin.Context) {
		token, err := auth.ParseToken(strings.Split(c.GetHeader("Authorization"), " ")[1], secret)
		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}

		path := c.PostForm("path")
		if path == "" {
			c.IndentedJSON(http.StatusOK, gin.H{"error": "field 'path' must not be empty"})
			return
		} else if path[len(path)-1:] != "/" {
			c.IndentedJSON(http.StatusOK, gin.H{"error": "directory must contain '/' at the end"})
			return
		}

		if _, err = s3.PutObject(token[0], path, nil, 0, minio.PutObjectOptions{}); err != nil {
			log.Printf("CreateDirHandler:MakeBucket: %s", err)
			c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}
		c.IndentedJSON(http.StatusOK, gin.H{"message": "directory created successfully"})
	}
	return fn
}

// ListFilesHandler sends a request for a list of user files
func ListFilesHandler(s3 general.S3, secret *[]byte) gin.HandlerFunc {
	fn := func(c *gin.Context) {
		token, err := auth.ParseToken(strings.Split(c.GetHeader("Authorization"), " ")[1], secret)
		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": err})
			return
		}

		path := c.PostForm("path")
		filesList := general.GetS3Objects(s3, token[0], path, false)
		c.IndentedJSON(http.StatusOK, gin.H{"message": filesList})
	}
	return fn
}

// UploadFileHandler sends a request to download the user's file
func UploadFileHandler(s3 general.S3, secret *[]byte) gin.HandlerFunc {
	fn := func(c *gin.Context) {
		file, header, err := c.Request.FormFile("file")
		if err != nil {
			c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err})
			return
		}
		defer general.CloseFile(file)

		if err = putFile(c, s3, file, header.Filename, secret); err != nil {
			c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}
		c.IndentedJSON(http.StatusOK, gin.H{"message": "File uploaded successfully"})
	}
	return fn
}

func putFile(c *gin.Context, s3 general.S3, file multipart.File, filename string, secret *[]byte) error {
	token, err := auth.ParseToken(strings.Split(c.GetHeader("Authorization"), " ")[1], secret)
	if err != nil {
		return err
	}
	path := c.PostForm("path")

	counter += 1
	if counter == 255 {
		counter = 0
	}

	filepath := "/tmp/" + strconv.FormatInt(time.Now().UnixMicro(), 10) + string(counter) + token[0] + filename
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer general.CloseFile(out)

	if _, err = io.Copy(out, file); err != nil {
		return err
	}
	_, err = s3.FPutObject(token[0], path+filename, filepath, minio.PutObjectOptions{})
	if err != nil {
		return err
	}
	err = os.Remove(filepath)
	return err
}
