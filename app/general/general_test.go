package general_test

import (
	"FileStorage/app/general"
	"testing"
)

func TestHash(t *testing.T) {
	tests := []struct {
		name  string
		input string
		res   string
	}{
		{
			name:  "hash_password_1",
			input: "qwertyuiop",
			res:   "dd7548fbd04ea7da3eaca4a4de8c283cd810ae60da74f6a6eb104216e3295ca1"},
		{
			name:  "hash_password_2",
			input: "testPassword",
			res:   "e7ecf5b48eaa7312c6613c6da7520452a188abb0abe23204255dfd74de33d6c5"},
	}
	for _, tt := range tests {
		gotRes := general.Hash(tt.input)
		if gotRes != tt.res {
			t.Errorf("%s: Hash() gotRes = %v, want %v", tt.name, gotRes, tt.res)
		}
	}
}
