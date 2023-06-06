// Package main defines the router and its methods
package main

import (
	"FileStorage/api"
	"FileStorage/app/handlers"
	"FileStorage/auth"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"os"
)

func main() {
	conn, err := grpc.Dial(os.Getenv("GRPC_SERVER"), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}
	c := api.NewAuthClient(conn)

	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.POST("/auth/Sign-In", auth.SignInHandler(c))
	r.POST("/auth/Sign-Up", auth.SignUpHandler(c))
	r.POST("/app/Upload", auth.Middleware(), handlers.UploadFileHandler())
	r.POST("/app/ListFiles", auth.Middleware(), handlers.ListFilesHandler())
	r.GET("/app/Delete/:file", auth.Middleware(), handlers.DeleteFileHandler())
	r.GET("/app/Download/:file", auth.Middleware(), handlers.DownloadFileHandler())

	if err := r.Run(":8080"); err != nil {
		log.Panicf("error occurred: %s", err)
	}
}
