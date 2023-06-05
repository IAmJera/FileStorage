package user_test

import (
	"FileStorage/user"
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
			user:     user.User{Login: "qwerty", Password: "qwertyuio"},
			wantBool: false},
		{
			name:     "short_login",
			user:     user.User{Login: "qwe", Password: "qwertyuiopa"},
			wantBool: false},
		{
			name:     "long_login",
			user:     user.User{Login: "qwertyuiopasdfghjklzx", Password: "qwertyuiopa"},
			wantBool: false},
		{
			name:     "success",
			user:     user.User{Login: "qwerty", Password: "qwertyuiopa"},
			wantBool: true},
	}
	for _, tt := range tests {
		gotBool := tt.user.CheckCredentials()
		if gotBool != tt.wantBool {
			t.Errorf("%s: user.CheckCredentials() gotBool = %v, want %v", tt.name, gotBool, tt.wantBool)
		}
	}
}
