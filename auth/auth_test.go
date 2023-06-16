package auth_test

import (
	"FileStorage/api"
	"FileStorage/auth"
	"FileStorage/user"
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/wrapperspb"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

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

func (m *testRPC) AddUser(ctx context.Context, in *api.User, opts ...grpc.CallOption) (*wrapperspb.StringValue, error) {
	return &wrapperspb.StringValue{}, nil
}

func (m *testRPC) UpdateUser(ctx context.Context, in *api.User, opts ...grpc.CallOption) (*wrapperspb.StringValue, error) {
	return &wrapperspb.StringValue{}, nil
}

func (m *testRPC) DelUser(ctx context.Context, in *api.User, opts ...grpc.CallOption) (*wrapperspb.StringValue, error) {
	return &wrapperspb.StringValue{}, nil
}

func (m *testRPC) GetToken(ctx context.Context, in *api.User, opts ...grpc.CallOption) (*api.Token, error) {
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
			token:    "qwertyerligerhikgrebk",
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
			if gotErr.Error() != tc.wantErr.Error() {
				t.Errorf("%s: user.ParseToken() gotErr = %v, wantErr %v", tc.name, gotErr, tc.wantErr)
			}
		} else if gotUser[0] != tc.wantUser {
			t.Errorf("%s: user.ParseToken() getUser = %v, wantUser = %v", tc.name, gotUser[0], tc.wantUser)
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
			header:   "Bearer neriuogiobhnerg erhoig;erioh;",
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
			recorder := httptest.NewRecorder()
			req, err := http.NewRequest(http.MethodPost, tc.path, nil)
			assert.NoError(t, err)
			req.Header.Add("Authorization", tc.header)

			handler := auth.Middleware(&rpc, &secret)
			router.POST(tc.path, handler)
			router.ServeHTTP(recorder, req)

			assert.Equal(t, tc.expected, recorder.Code)
		})
	}
}
