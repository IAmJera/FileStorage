package main

import (
	"FileStorage/app"
	"FileStorage/auth"
	"github.com/gin-gonic/gin"
	"log"
	"os"
)

func main() {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.POST("/auth/Sign-In", auth.SignInHandler())
	r.POST("/auth/Sign-Up", auth.SignUpHandler())
	r.POST("/app/Upload", auth.Middleware(), app.UploadFileHandler())
	r.POST("/app/ListFiles", auth.Middleware(), app.ListFilesHandler())
	r.GET("/app/Delete/:file", auth.Middleware(), app.DeleteFileHandler())
	r.GET("/app/Download/:file", auth.Middleware(), app.DownloadFileHandler())

	if err := r.Run(":8080"); err != nil {
		log.Printf("error occurred: %s", err)
		os.Exit(1)
	}
}
