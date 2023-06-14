package handlers_test

import (
	"FileStorage/app/handlers"
	"bytes"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/minio/minio-go"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

type User struct {
	jwt.StandardClaims
	Login    string
	Password string
}

type testS3 struct{}

func (s *testS3) GetObject(bucketName string, _ string, _ minio.GetObjectOptions) (*minio.Object, error) {
	if bucketName == "error" {
		return &minio.Object{}, fmt.Errorf("error")
	}
	return nil, nil
}
func (s *testS3) MakeBucket(bucketName string, location string) (err error) {
	return nil
}
func (s *testS3) RemoveBucket(bucketName string) error {
	return nil
}
func (s *testS3) RemoveObject(bucketName string, _ string) error {
	if bucketName == "error" {
		return fmt.Errorf("error")
	}
	return nil
}
func (s *testS3) PutObject(bucketName string, _ string, _ io.Reader, _ int64, _ minio.PutObjectOptions) (n int64, err error) {
	if bucketName == "error" {
		return 0, fmt.Errorf("error")
	}
	return 0, nil
}
func (s *testS3) FPutObject(bucketName string, objectName string, filePath string, opts minio.PutObjectOptions) (n int64, err error) {
	return 0, nil
}
func (s *testS3) ListObjects(_ string, objectPrefix string, _ bool, _ <-chan struct{}) <-chan minio.ObjectInfo {
	ch := make(chan minio.ObjectInfo)
	go func() {
		defer close(ch)
		select {
		case ch <- minio.ObjectInfo{Key: objectPrefix}:
		default:
		}
	}()
	return ch
}

func createToken(login string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &User{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 24).Unix(),
			IssuedAt:  time.Now().Unix(),
		},
		Login: login,
	})
	return token.SignedString([]byte(os.Getenv("SIGNING_KEY")))
}

func TestDeleteObjectHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	errToken, _ := createToken("error")
	token, _ := createToken("test")
	s3 := &testS3{}

	testCases := []struct {
		description string
		token       string
		path        string
		body        string
		expected    int
	}{
		{
			description: "valid token",
			token:       token,
			path:        "/path1",
			body:        "test1",
			expected:    http.StatusOK},
		{
			description: "error removing",
			token:       errToken,
			path:        "/path2",
			body:        "test2",
			expected:    http.StatusInternalServerError},
		{
			description: "invalid token",
			token:       "wrongToken",
			path:        "/path3",
			body:        "test3",
			expected:    http.StatusInternalServerError},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			req, err := http.NewRequest(http.MethodPost, tc.path, bytes.NewBufferString(tc.body))
			assert.NoError(t, err)
			req.Header.Add("Authorization", "Bearer "+tc.token)

			handler := handlers.DeleteObjectHandler(s3)
			router.POST(tc.path, handler)
			router.ServeHTTP(recorder, req)

			assert.Equal(t, tc.expected, recorder.Code)
		})
	}
}

func TestDownloadFileHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	errToken, _ := createToken("error")
	token, _ := createToken("test")
	s3 := &testS3{}

	testCases := []struct {
		description string
		token       string
		path        string
		expected    int
	}{
		{
			description: "valid token and form data",
			token:       token,
			path:        "/path1",
			expected:    http.StatusOK},
		{
			description: "error downloading",
			token:       errToken,
			path:        "/path2",
			expected:    http.StatusInternalServerError},
		{
			description: "invalid token",
			token:       "wrongToken",
			path:        "/path3",
			expected:    http.StatusInternalServerError},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			req, err := http.NewRequest(http.MethodPost, tc.path, nil)
			assert.NoError(t, err)
			req.Header.Add("Authorization", "Bearer "+tc.token)

			handler := handlers.DownloadFileHandler(s3)
			router.POST(tc.path, handler)
			router.ServeHTTP(recorder, req)

			assert.Equal(t, tc.expected, recorder.Code)
		})
	}
}

func TestCreateDirHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	errToken, _ := createToken("error")
	token, _ := createToken("test")
	s3 := &testS3{}

	testCases := []struct {
		description string
		token       string
		path        string
		body        string
		expected    int
	}{
		{
			description: "valid token and form data",
			token:       token,
			path:        "/path1",
			body:        "test/",
			expected:    http.StatusOK},
		{
			description: "error creating dir",
			token:       errToken,
			path:        "/path2",
			body:        "test",
			expected:    http.StatusOK},
		{
			description: "error creating dir",
			token:       errToken,
			path:        "/path3",
			body:        "",
			expected:    http.StatusOK},
		{
			description: "invalid token",
			token:       "wrongToken",
			path:        "/path4",
			body:        "",
			expected:    http.StatusInternalServerError},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			req, err := http.NewRequest(http.MethodPost, tc.path, bytes.NewBufferString(tc.body))
			assert.NoError(t, err)
			req.Header.Add("Authorization", "Bearer "+tc.token)

			handler := handlers.CreateDirHandler(s3)
			router.POST(tc.path, handler)
			router.ServeHTTP(recorder, req)

			assert.Equal(t, tc.expected, recorder.Code)
		})
	}
}

func TestListFilesHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	token, _ := createToken("test")
	s3 := &testS3{}

	testCases := []struct {
		description string
		token       string
		path        string
		expected    int
	}{
		{
			description: "valid token",
			token:       token,
			path:        "/path1",
			expected:    http.StatusOK},
		{
			description: "invalid token",
			token:       "wrongToken",
			path:        "/path3",
			expected:    http.StatusInternalServerError},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			req, err := http.NewRequest(http.MethodPost, tc.path, nil)
			assert.NoError(t, err)
			req.Header.Add("Authorization", "Bearer "+tc.token)

			handler := handlers.ListFilesHandler(s3)
			router.POST(tc.path, handler)
			router.ServeHTTP(recorder, req)

			assert.Equal(t, tc.expected, recorder.Code)
		})
	}
}
