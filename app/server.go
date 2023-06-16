// Package main defines the router and its methods
package main

import (
	"FileStorage/api"
	"FileStorage/app/handlers"
	"FileStorage/auth"
	"github.com/gin-gonic/gin"
	"github.com/minio/minio-go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"os"
)

func main() {
	var secret = []byte(os.Getenv("SIGNING_KEY"))
	conn, err := grpc.Dial(os.Getenv("GRPC_SERVER"), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}
	rpc := api.NewAuthClient(conn)

	s3, err := minio.New(os.Getenv("MINIO_ADDRESS"), os.Getenv("MINIO_KEY"),
		os.Getenv("MINIO_SECRET"), false)
	if err != nil {
		log.Panic(err)
	}

	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.POST("/auth/Sign-In", auth.SignInHandler(rpc))
	r.POST("/auth/Sign-Up", auth.SignUpHandler(rpc, s3))
	r.PUT("/auth/ChangePass", auth.Middleware(rpc), auth.ChangePasswordHandler(rpc))
	r.DELETE("/auth/DeleteUser", auth.Middleware(rpc), auth.DeleteUserHandler(rpc, s3))

	r.POST("/app/Upload", auth.Middleware(rpc, &secret), handlers.UploadFileHandler(s3, &secret))
	r.POST("/app/ListFiles", auth.Middleware(rpc, &secret), handlers.ListFilesHandler(s3, &secret))
	r.POST("/app/CreateDir", auth.Middleware(rpc, &secret), handlers.CreateDirHandler(s3, &secret))
	r.DELETE("/app/DeleteObject", auth.Middleware(rpc, &secret), handlers.DeleteObjectHandler(s3, &secret))
	r.POST("/app/Download", auth.Middleware(rpc, &secret), handlers.DownloadFileHandler(s3, &secret))

	if err = r.Run(":8080"); err != nil {
		log.Panicf("error occurred: %s", err)
	}
}
