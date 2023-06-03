// Package main defines the router and its methods
package main

import (
	"FileStorage/app/handlers"
	"FileStorage/auth"
	"github.com/gin-gonic/gin"
	"log"
)

func main() {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.POST("/auth/Sign-In", auth.SignInHandler())
	r.POST("/auth/Sign-Up", auth.SignUpHandler())
	r.POST("/app/Upload", auth.Middleware(), handlers.UploadFileHandler())
	r.POST("/app/ListFiles", auth.Middleware(), handlers.ListFilesHandler())
	r.GET("/app/Delete/:file", auth.Middleware(), handlers.DeleteFileHandler())
	r.GET("/app/Download/:file", auth.Middleware(), handlers.DownloadFileHandler())

	if err := r.Run(":8080"); err != nil {
		log.Panicf("error occurred: %s", err)
	}
}
