package user_test

import (
	"FileStorage/user"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
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
		assert.Equal(t, tc.wantBool, gotBool)
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
		assert.Equal(t, tc.password, usr.Password)
		assert.Equal(t, tc.login, usr.Login)
	}
}
