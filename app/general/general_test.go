package general_test

import (
	"FileStorage/app/general"
	"github.com/minio/minio-go"
	"github.com/stretchr/testify/assert"
	"io"
	"testing"
)

func TestHash(t *testing.T) {
	tests := []struct {
		name  string
		input string
		salt  string
		res   string
	}{
		{
			name:  "hash_password_1",
			input: "qwertyuiop",
			res:   "dd7548fbd04ea7da3eaca4a4de8c283cd810ae60da74f6a6eb104216e3295ca1",
			salt:  "salt"},
		{
			name:  "hash_password_2",
			input: "testPassword",
			res:   "e7ecf5b48eaa7312c6613c6da7520452a188abb0abe23204255dfd74de33d6c5",
			salt:  "salt"},
	}
	for _, tt := range tests {
		gotRes := general.Hash(tt.input, tt.salt)
		if gotRes != tt.res {
			t.Errorf("%s: Hash() gotRes = %v, want %v", tt.name, gotRes, tt.res)
		}
	}
}

type testS3 struct{}

func (s *testS3) GetObject(_ string, _ string, _ minio.GetObjectOptions) (*minio.Object, error) {
	return &minio.Object{}, nil
}
func (s *testS3) MakeBucket(_ string, _ string) (err error) {
	return nil
}
func (s *testS3) RemoveBucket(_ string) error {
	return nil
}
func (s *testS3) RemoveObject(_ string, _ string) error {
	return nil
}
func (s *testS3) PutObject(_ string, _ string, _ io.Reader, _ int64, _ minio.PutObjectOptions) (n int64, err error) {
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

func TestGetS3Objects(t *testing.T) {
	s3 := &testS3{}

	testCases := []struct {
		description string
		login       string
		path        string
		prefix      string
		expected    []string
	}{
		{
			description: "test1",
			login:       "test",
			prefix:      "test1",
			expected:    []string{"test1"}},
		{
			description: "test2",
			login:       "test",
			prefix:      "test2",
			expected:    []string{"test2"}},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			handler := general.GetS3Objects(s3, tc.login, tc.prefix, false)
			assert.Equal(t, tc.expected, handler)
		})
	}
}
