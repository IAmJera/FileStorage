package user_test

import (
	"FileStorage/user"
	"github.com/gin-gonic/gin"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestUser_CheckCredentials(t *testing.T) {
	tests := []struct {
		name     string
		user     user.User
		wantBool bool
	}{
		{
			name:     "short_password",
			user:     user.User{Login: "qwerty", Password: "qwerty123"},
			wantBool: false},
		{
			name:     "short_login",
			user:     user.User{Login: "qwe", Password: "qwerty12345"},
			wantBool: false},
		{
			name:     "long_login",
			user:     user.User{Login: "qwerty123456789101011", Password: "qwerty12345"},
			wantBool: false},
		{
			name:     "success",
			user:     user.User{Login: "qwerty", Password: "qwerty12345"},
			wantBool: true},
	}
	for _, tc := range tests {
		gotBool := tc.user.CheckCredentials()
		if gotBool != tc.wantBool {
			t.Errorf("%s: user.CheckCredentials() gotBool = %v, want %v", tc.name, gotBool, tc.wantBool)
		}
	}
}

func TestUser_ParseCredentials(t *testing.T) {
	tests := []struct {
		name     string
		login    string
		password string
		wantUser user.User
		wantBool bool
	}{
		{
			name:     "right credentials",
			login:    "testUser",
			password: "testPassword",
			wantUser: user.User{Login: "testUser", Password: "testPassword"}},
		{
			name:     "wrong password",
			login:    "test1",
			password: "qwerty",
			wantUser: user.User{Login: "test1", Password: "qwerty1"}},
		{
			name:     "wrong user",
			login:    "test",
			password: "qwerty1",
			wantUser: user.User{Login: "test1", Password: "qwerty1"}},
		{
			name:     "wrong credentials",
			login:    "test",
			password: "qwerty",
			wantUser: user.User{Login: "test1", Password: "qwerty1"}},
	}
	for _, tc := range tests {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		form := strings.NewReader("login=" + tc.login + "&password=" + tc.password)
		c.Request, _ = http.NewRequest(http.MethodPost, "/", form)
		c.Request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		usr := user.User{}

		if ok := usr.ParseCredentials(c); !ok {
			t.Fatal("Failed to parse credentials")
		}
		if usr.Login != tc.login && usr.Password != tc.password {
			t.Errorf("%s: user.ParseCredentials() gotUser = %v, want %v", tc.name, usr, tc.wantUser)
		}
	}
}
