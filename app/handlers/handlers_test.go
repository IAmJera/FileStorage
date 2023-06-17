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
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
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

func (s *testS3) MakeBucket(_ string, _ string) (err error) {
	return nil
}

func (s *testS3) RemoveBucket(_ string) error {
	return nil
}

func (s *testS3) RemoveObject(bucketName string, _ string) error {
	if bucketName == "error" {
		return fmt.Errorf("error")
	}
	return nil
}

func (s *testS3) PutObject(bucketName string, _ string, _ io.Reader, _ int64, _ minio.PutObjectOptions) (n int64, err error) {
	if bucketName == "error" || bucketName == "s3error" {
		return 0, fmt.Errorf("error")
	}
	return 0, nil
}

func (s *testS3) FPutObject(bucketName string, _ string, _ string, _ minio.PutObjectOptions) (n int64, err error) {
	if bucketName == "error" {
		return 0, fmt.Errorf("error")
	}
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

func createToken(login string, secret *[]byte) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &User{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 24).Unix(),
			IssuedAt:  time.Now().Unix(),
		},
		Login: login,
	})
	return token.SignedString(*secret)
}

func TestDeleteObjectHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	secret := []byte("secret")
	router := gin.Default()
	errToken, _ := createToken("error", &secret)
	token, _ := createToken("test", &secret)
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
			body:        "objectPath=test1",
			expected:    http.StatusOK},
		{
			description: "error removing",
			token:       errToken,
			path:        "/path2",
			body:        "objectPath=test2",
			expected:    http.StatusInternalServerError},
		{
			description: "invalid token",
			token:       "wrongToken",
			path:        "/path3",
			body:        "objectPath=test3",
			expected:    http.StatusInternalServerError},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, err := http.NewRequest(http.MethodPost, tc.path, strings.NewReader(tc.body))
			assert.NoError(t, err)
			req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
			req.Header.Add("Authorization", "Bearer "+tc.token)

			handler := handlers.DeleteObjectHandler(s3, &secret)
			router.POST(tc.path, handler)
			router.ServeHTTP(w, req)

			assert.Equal(t, tc.expected, w.Code)
		})
	}
}

func TestDownloadFileHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	secret := []byte("secret")
	router := gin.Default()
	errToken, _ := createToken("error", &secret)
	token, _ := createToken("test", &secret)
	s3 := &testS3{}

	testCases := []struct {
		name     string
		token    string
		path     string
		expected int
	}{
		{
			name:     "valid token and form data",
			token:    token,
			path:     "/path1",
			expected: http.StatusOK},
		{
			name:     "error downloading",
			token:    errToken,
			path:     "/path2",
			expected: http.StatusInternalServerError},
		{
			name:     "invalid token",
			token:    "wrongToken",
			path:     "/path3",
			expected: http.StatusInternalServerError},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, err := http.NewRequest(http.MethodPost, tc.path, nil)
			assert.NoError(t, err)
			req.Header.Add("Authorization", "Bearer "+tc.token)

			handler := handlers.DownloadFileHandler(s3, &secret)
			router.POST(tc.path, handler)
			router.ServeHTTP(w, req)

			assert.Equal(t, tc.expected, w.Code)
		})
	}
}

func TestCreateDirHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	secret := []byte("secret")
	router := gin.Default()
	errToken, _ := createToken("error", &secret)
	errToken2, _ := createToken("s3error", &secret)
	token, _ := createToken("test", &secret)
	s3 := &testS3{}

	testCases := []struct {
		name     string
		token    string
		path     string
		body     string
		expected int
	}{
		{
			name:     "valid token and form data",
			token:    token,
			path:     "/path1",
			body:     "path=test/",
			expected: http.StatusOK},
		{
			name:     "error creating dir",
			token:    errToken,
			path:     "/path2",
			body:     "path=test",
			expected: http.StatusOK},
		{
			name:     "error creating dir",
			token:    errToken,
			path:     "/path3",
			body:     "path=",
			expected: http.StatusOK},
		{
			name:     "invalid token",
			token:    "wrongToken",
			path:     "/path4",
			body:     "path=",
			expected: http.StatusInternalServerError},
		{
			name:     "s3 error",
			token:    errToken2,
			path:     "/path5",
			body:     "path=/test/",
			expected: http.StatusInternalServerError},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, err := http.NewRequest(http.MethodPost, tc.path, bytes.NewBufferString(tc.body))
			assert.NoError(t, err)
			req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
			req.Header.Add("Authorization", "Bearer "+tc.token)

			handler := handlers.CreateDirHandler(s3, &secret)
			router.POST(tc.path, handler)
			router.ServeHTTP(w, req)

			assert.Equal(t, tc.expected, w.Code)
		})
	}
}

func TestListFilesHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	secret := []byte("secret")
	router := gin.Default()
	token, _ := createToken("test", &secret)
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
			w := httptest.NewRecorder()
			req, err := http.NewRequest(http.MethodPost, tc.path, nil)
			assert.NoError(t, err)
			req.Header.Add("Authorization", "Bearer "+tc.token)

			handler := handlers.ListFilesHandler(s3, &secret)
			router.POST(tc.path, handler)
			router.ServeHTTP(w, req)

			assert.Equal(t, tc.expected, w.Code)
		})
	}
}

func TestUploadFileHandler(t *testing.T) {
	secret := []byte("secret")
	token1, _ := createToken("noFile", &secret)
	token2, _ := createToken("error", &secret)
	token3, _ := createToken("test", &secret)
	s3 := testS3{}

	testCases := []struct {
		description string
		token       string
		path        string
		file        bool
		expected    int
	}{
		{
			description: "without file",
			token:       token1,
			path:        "/path1",
			file:        false,
			expected:    http.StatusBadRequest},
		{
			description: "s3 error",
			token:       token2,
			path:        "/path2",
			file:        true,
			expected:    http.StatusInternalServerError},
		{
			description: "success",
			token:       token3,
			path:        "/path3",
			file:        true,
			expected:    http.StatusOK},
	}

	for _, tc := range testCases {
		w := httptest.NewRecorder()
		var req *http.Request
		var router *gin.Engine

		if tc.file {
			tempFile, _ := os.CreateTemp("", "test-file")
			bodyBuf := &bytes.Buffer{}
			bodyWriter := multipart.NewWriter(bodyBuf)
			fileWriter, _ := bodyWriter.CreateFormFile("file", tempFile.Name())
			f, _ := os.Open(tempFile.Name())
			_, _ = io.Copy(fileWriter, f)
			contentType := bodyWriter.FormDataContentType()
			_ = bodyWriter.Close()

			req, _ = http.NewRequest("POST", tc.path, bodyBuf)
			req.Header.Add("Content-Type", contentType)
			_, router = gin.CreateTestContext(w)
			router.POST(tc.path, handlers.UploadFileHandler(&s3, &secret))

			_ = os.Remove(tempFile.Name())
			_ = f.Close()
		} else {
			req, _ = http.NewRequest("POST", tc.path, nil)
			_, router = gin.CreateTestContext(w)
			router.POST(tc.path, handlers.UploadFileHandler(&s3, &secret))
		}

		req.Header.Add("Authorization", "Bearer "+tc.token)
		router.ServeHTTP(w, req)

		assert.Equal(t, tc.expected, w.Code)
	}
}
