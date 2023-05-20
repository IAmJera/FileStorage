package app

import (
	"FileStorage/auth"
	"github.com/gin-gonic/gin"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
)

var mySigningKey = []byte(os.Getenv("SIGNINGKEY"))

func DeleteFileHandler() gin.HandlerFunc {
	fn := func(c *gin.Context) {
		token, err := auth.ParseToken(strings.Split(c.GetHeader("Authorization"), " ")[1], mySigningKey)
		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}
		if err := os.RemoveAll(os.Getenv("BASEDIR") + token[0] + "/" + c.Param("file")); err != nil {
			c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}
		c.IndentedJSON(http.StatusOK, gin.H{"message": "remove successful"})
	}
	return fn
}

func DownloadFileHandler() gin.HandlerFunc {
	fn := func(c *gin.Context) {
		token, err := auth.ParseToken(strings.Split(c.GetHeader("Authorization"), " ")[1], mySigningKey)
		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}
		byteFile, err := os.ReadFile(os.Getenv("BASEDIR") + token[0] + "/" + c.Param("file"))
		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err})
		}
		c.Header("Content-Disposition", "attachment; filename="+c.Param("file"))
		c.Data(http.StatusOK, "application/octet-stream", byteFile)
	}
	return fn
}

func ListFilesHandler() gin.HandlerFunc {
	fn := func(c *gin.Context) {
		token, err := auth.ParseToken(strings.Split(c.GetHeader("Authorization"), " ")[1], mySigningKey)
		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": err})
			return
		}
		files, err := os.ReadDir(os.Getenv("BASEDIR") + token[0])
		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": err})
			return
		}

		var fls []string
		for _, f := range files {
			fls = append(fls, f.Name())
		}
		c.IndentedJSON(http.StatusOK, gin.H{"message": fls})
	}
	return fn
}

func UploadFileHandler() gin.HandlerFunc {
	fn := func(c *gin.Context) {
		file, header, err := c.Request.FormFile("file")
		if err != nil {
			c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err})
		}
		defer CloseFile(file)

		if err := writeFile(c, file, header.Filename); err != nil {
			c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}
		c.IndentedJSON(http.StatusOK, gin.H{"message": "File uploaded successfully"})
	}
	return fn
}

func writeFile(c *gin.Context, file multipart.File, filename string) error {
	token, err := auth.ParseToken(strings.Split(c.GetHeader("Authorization"), " ")[1], mySigningKey)

	out, err := os.Create(token[0] + "/" + filename)
	if err != nil {
		return err
	}
	defer CloseFile(out)

	if _, err = io.Copy(out, file); err != nil {
		return err
	}

	return nil
}

type Closer interface {
	Close() error
}

func CloseFile(c Closer) {
	if err := c.Close(); err != nil {
		log.Println("error occurred: ", err)
	}
	return
}
