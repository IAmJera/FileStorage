package auth

import (
	"FileStorage/api"
	"FileStorage/app/general"
	"FileStorage/user"
	"context"
	"github.com/gin-gonic/gin"
	"github.com/minio/minio-go"
	"log"
	"net/http"
	"os"
	"strings"
)

var mySigningKey = []byte(os.Getenv("SIGNING_KEY"))

// SignUpHandler registers the user by writing his data to the database
func SignUpHandler(rpc api.AuthClient, s3 *minio.Client) gin.HandlerFunc {
	fn := func(c *gin.Context) {
		var usr user.User
		if ok := usr.ParseCredentials(c); !ok {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		if ok := usr.CheckCredentials(); !ok {
			c.IndentedJSON(http.StatusOK, gin.H{"error": "credentials does not meet requirements"})
			return
		}

		u := api.User{Login: usr.Login, Password: general.Hash(usr.Password), Role: usr.Role}
		resp, err := rpc.AddUser(context.Background(), &u)
		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}
		if resp.Value == "user already exist" {
			c.IndentedJSON(http.StatusOK, gin.H{"error": resp.Value})
			return
		}

		if err = s3.MakeBucket(usr.Login, ""); err != nil {
			_, err = rpc.DelUser(context.Background(), &u)
			log.Printf("SignUpHandler:DelUser: %s", err)
			c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "error occurred while creating the user"})
			return
		}
		c.IndentedJSON(http.StatusCreated, gin.H{"message": "user created successfully"})
	}
	return fn
}

// ChangePasswordHandler changes user's password
func ChangePasswordHandler(rpc api.AuthClient) gin.HandlerFunc {
	fn := func(c *gin.Context) {
		token, err := ParseToken(strings.Split(c.GetHeader("Authorization"), " ")[1], &mySigningKey)
		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}

		var usr user.User
		if ok := usr.ParseCredentials(c); !ok {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
		u := api.User{Login: token[0], Role: usr.Role, Password: general.Hash(usr.Password)}
		res, err := rpc.UpdateUser(context.Background(), &u)
		if err != nil {
			c.IndentedJSON(http.StatusOK, gin.H{"error": err})
			return
		}
		if res.Value != "" {
			c.IndentedJSON(http.StatusOK, gin.H{"message": res.Value})
			return
		}
		c.IndentedJSON(http.StatusOK, gin.H{"message": "password changed successfully"})
		return
	}
	return fn
}

// DeleteUserHandler removes user from DB and S3
func DeleteUserHandler(rpc api.AuthClient, s3 *minio.Client) gin.HandlerFunc {
	fn := func(c *gin.Context) {
		token, err := ParseToken(strings.Split(c.GetHeader("Authorization"), " ")[1], &mySigningKey)
		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": err})
			return
		}

		res, err := rpc.DelUser(context.Background(), &api.User{Login: token[0]})
		if err != nil {
			c.IndentedJSON(http.StatusOK, gin.H{"error": err})
			return
		}
		if res.Value != "" {
			c.IndentedJSON(http.StatusOK, gin.H{"message": res.Value})
			return
		}

		filesList := general.GetS3Objects(s3, token[0], "", true)
		for _, obj := range filesList {
			if err = s3.RemoveObject(token[0], obj); err != nil {
				log.Printf("DeleteUserHandler:RemoveObject: %s", err)
			}
		}

		if err = s3.RemoveBucket(token[0]); err != nil {
			log.Printf("DeleteUserHandler:RemoveBucket: %s", err)
		}
		c.IndentedJSON(http.StatusOK, gin.H{"message": "user deleted successfully"})
		return
	}
	return fn
}

// SignInHandler authenticate the user and returns a jwt token to him
func SignInHandler(rpc api.AuthClient) gin.HandlerFunc {
	fn := func(c *gin.Context) {
		var usr user.User
		if ok := usr.ParseCredentials(c); !ok {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		u := api.User{Login: usr.Login, Role: usr.Role, Password: general.Hash(usr.Password)}
		token, err := rpc.GetToken(context.Background(), &u)
		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}
		if token.Error != "" {
			c.IndentedJSON(http.StatusOK, gin.H{"error": token.Error})
			return
		}

		c.IndentedJSON(http.StatusOK, gin.H{"token": token.Token})
		return
	}
	return fn
}
