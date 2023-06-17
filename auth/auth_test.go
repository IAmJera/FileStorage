package auth_test

import (
	"FileStorage/api"
	"FileStorage/auth"
	"FileStorage/user"
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/minio/minio-go"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/wrapperspb"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type testS3 struct{}

func (s *testS3) GetObject(bucketName string, _ string, _ minio.GetObjectOptions) (*minio.Object, error) {
	if bucketName == "error" {
		return &minio.Object{}, fmt.Errorf("error")
	}
	return nil, nil
}
func (s *testS3) MakeBucket(bucketName string, _ string) (err error) {
	if bucketName == "bucketError" {
		return fmt.Errorf("error")
	}
	return nil
}
func (s *testS3) RemoveBucket(_ string) error {
	return fmt.Errorf("error")
}
func (s *testS3) RemoveObject(bucketName string, _ string) error {
	if bucketName == "error" || bucketName == "test" {
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
func (s *testS3) FPutObject(_ string, _ string, _ string, _ minio.PutObjectOptions) (n int64, err error) {
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

type testRPC struct{}

func (m *testRPC) UserExist(_ context.Context, in *api.User, _ ...grpc.CallOption) (*wrapperspb.BoolValue, error) {
	switch in.Login {
	case "nil":
		return nil, nil
	case "true":
		return &wrapperspb.BoolValue{Value: true}, nil
	case "false":
		return &wrapperspb.BoolValue{Value: false}, nil
	}
	return &wrapperspb.BoolValue{}, nil
}

func (m *testRPC) AddUser(_ context.Context, in *api.User, _ ...grpc.CallOption) (*wrapperspb.StringValue, error) {
	switch in.Login {
	case "exist":
		return &wrapperspb.StringValue{Value: "user already exist"}, nil
	case "error":
		return &wrapperspb.StringValue{}, fmt.Errorf("error")
	}
	return &wrapperspb.StringValue{}, nil
}

func (m *testRPC) UpdateUser(_ context.Context, in *api.User, _ ...grpc.CallOption) (*wrapperspb.StringValue, error) {
	switch in.Login {
	case "resError":
		return &wrapperspb.StringValue{Value: "error"}, nil
	case "error":
		return &wrapperspb.StringValue{}, fmt.Errorf("error")
	}
	return &wrapperspb.StringValue{}, nil
}

func (m *testRPC) DelUser(_ context.Context, in *api.User, _ ...grpc.CallOption) (*wrapperspb.StringValue, error) {
	switch in.Login {
	case "resError":
		return &wrapperspb.StringValue{Value: "error"}, nil
	case "error":
		return &wrapperspb.StringValue{}, fmt.Errorf("error")
	}
	return &wrapperspb.StringValue{}, nil
}

func (m *testRPC) GetToken(_ context.Context, in *api.User, _ ...grpc.CallOption) (*api.Token, error) {
	switch in.Login {
	case "getError":
		return &api.Token{}, fmt.Errorf("error")
	case "tokenError":
		return &api.Token{Error: "error"}, nil
	}
	return &api.Token{}, nil
}

func createToken(login string, secret *[]byte) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &user.User{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 24).Unix(),
			IssuedAt:  time.Now().Unix(),
		},
		Login: login,
	})
	return token.SignedString(*secret)
}

func TestParseToken(t *testing.T) {
	key := []byte("secret")
	token, _ := createToken("test1", &key)
	token2, _ := createToken("test3", &key)
	tests := []struct {
		name     string
		token    string
		secret   []byte
		wantUser string
		wantErr  error
	}{
		{
			name:     "valid token and secret",
			wantUser: "test1",
			token:    token,
			secret:   []byte("secret"),
			wantErr:  nil},
		{
			name:     "invalid token",
			wantUser: "test2",
			token:    "qwerty12345qwerty",
			secret:   []byte("secret"),
			wantErr:  fmt.Errorf("token contains an invalid number of segments")},
		{
			name:     "valid token, invalid secret",
			wantUser: "test3",
			token:    token2,
			secret:   []byte("qwerty"),
			wantErr:  fmt.Errorf("signature is invalid")},
	}
	for _, tc := range tests {
		gotUser, gotErr := auth.ParseToken(tc.token, &tc.secret)
		if gotErr != nil {
			assert.Equal(t, tc.wantErr.Error(), gotErr.Error())
		} else {
			assert.Equal(t, tc.wantUser, gotUser[0])
		}
	}
}

func TestMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	secret := []byte("secret")
	router := gin.Default()
	rpc := testRPC{}
	token1, _ := createToken("nil", &secret)
	token2, _ := createToken("true", &secret)
	token3, _ := createToken("false", &secret)

	testCases := []struct {
		name     string
		header   string
		path     string
		expected int
	}{
		{
			name:     "empty header",
			header:   "",
			path:     "/path1",
			expected: http.StatusUnauthorized},
		{
			name:     "too much substrings",
			header:   "Bearer qwerty12345 qwerty12345;",
			path:     "/path2",
			expected: http.StatusUnauthorized},
		{
			name:     "nil response",
			header:   "Bearer " + token1,
			path:     "/path3",
			expected: http.StatusInternalServerError},
		{
			name:     "false response",
			header:   "Bearer " + token3,
			path:     "/path4",
			expected: http.StatusUnauthorized},
		{
			name:     "true response",
			header:   "Bearer " + token2,
			path:     "/path5",
			expected: http.StatusOK},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, err := http.NewRequest(http.MethodPost, tc.path, nil)
			assert.NoError(t, err)
			req.Header.Add("Authorization", tc.header)

			handler := auth.Middleware(&rpc, &secret)
			router.POST(tc.path, handler)
			router.ServeHTTP(w, req)

			assert.Equal(t, tc.expected, w.Code)
		})
	}
}

func TestSignUpHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	rpc := testRPC{}
	s3 := testS3{}

	testCases := []struct {
		name     string
		login    string
		password string
		token    string
		path     string
		expected int
	}{
		{
			name:     "rpc error",
			login:    "error",
			password: "qwerty12345",
			path:     "/path1",
			expected: http.StatusInternalServerError},
		{
			name:     "user already exist",
			login:    "exist",
			password: "qwerty12345",
			path:     "/path2",
			expected: http.StatusOK},
		{
			name:     "bucket error",
			login:    "bucketError",
			password: "qwerty12345",
			path:     "/path3",
			expected: http.StatusInternalServerError},
		{
			name:     "success",
			login:    "success",
			password: "qwerty12345",
			path:     "/path4",
			expected: http.StatusCreated},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			form := strings.NewReader("login=" + tc.login + "&password=" + tc.password)
			req, _ := http.NewRequest(http.MethodPost, tc.path, form)
			req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

			handler := auth.SignUpHandler(&rpc, &s3)
			router.POST(tc.path, handler)
			router.ServeHTTP(w, req)

			assert.Equal(t, tc.expected, w.Code)
		})
	}
}

func TestChangePasswordHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	secret := []byte("secret")
	token1, _ := createToken("error", &secret)
	token2, _ := createToken("resError", &secret)
	token3, _ := createToken("test", &secret)
	rpc := testRPC{}

	testCases := []struct {
		name     string
		login    string
		password string
		token    string
		path     string
		expected int
	}{
		{
			name:     "rpc error",
			login:    "error",
			password: "qwerty12345",
			token:    token1,
			path:     "/path1",
			expected: http.StatusOK},
		{
			name:     "response error",
			login:    "resError",
			password: "qwerty12345",
			token:    token2,
			path:     "/path2",
			expected: http.StatusOK},
		{
			name:     "success",
			login:    "test",
			password: "qwerty12345",
			token:    token3,
			path:     "/path3",
			expected: http.StatusOK},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			form := strings.NewReader("login=" + tc.login + "&password=" + tc.password)
			req, _ := http.NewRequest(http.MethodPost, tc.path, form)
			req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
			req.Header.Add("Authorization", "Bearer "+tc.token)

			handler := auth.ChangePasswordHandler(&rpc, &secret)
			router.POST(tc.path, handler)
			router.ServeHTTP(w, req)

			assert.Equal(t, tc.expected, w.Code)
		})
	}
}

func TestDeleteUserHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	secret := []byte("secret")
	token1, _ := createToken("error", &secret)
	token2, _ := createToken("resError", &secret)
	token3, _ := createToken("test", &secret)
	rpc := testRPC{}
	s3 := testS3{}

	testCases := []struct {
		name     string
		login    string
		password string
		token    string
		path     string
		expected int
	}{
		{
			name:     "rpc error",
			login:    "error",
			password: "qwerty12345",
			token:    token1,
			path:     "/path1",
			expected: http.StatusOK},
		{
			name:     "response error",
			login:    "resError",
			password: "qwerty12345",
			token:    token2,
			path:     "/path2",
			expected: http.StatusOK},
		{
			name:     "success",
			login:    "test",
			password: "qwerty12345",
			token:    token3,
			path:     "/path3",
			expected: http.StatusOK},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodPost, tc.path, nil)
			req.Header.Add("Authorization", "Bearer "+tc.token)

			handler := auth.DeleteUserHandler(&rpc, &s3, &secret)
			router.POST(tc.path, handler)
			router.ServeHTTP(w, req)

			assert.Equal(t, tc.expected, w.Code)
		})
	}
}

func TestSignInHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	rpc := testRPC{}

	testCases := []struct {
		name     string
		login    string
		password string
		path     string
		expected int
	}{
		{
			name:     "get token error",
			login:    "getError",
			password: "qwerty12345",
			path:     "/path1",
			expected: http.StatusInternalServerError},
		{
			name:     "token error",
			login:    "tokenError",
			password: "qwerty12345",
			path:     "/path2",
			expected: http.StatusOK},
		{
			name:     "success",
			login:    "test",
			password: "qwerty12345",
			path:     "/path3",
			expected: http.StatusOK},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			form := strings.NewReader("login=" + tc.login + "&password=" + tc.password)
			req, _ := http.NewRequest(http.MethodPost, tc.path, form)
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			handler := auth.SignInHandler(&rpc)
			router.POST(tc.path, handler)
			router.ServeHTTP(w, req)

			assert.Equal(t, tc.expected, w.Code)
		})
	}
}
