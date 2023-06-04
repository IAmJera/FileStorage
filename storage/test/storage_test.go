package storage_test

import (
	"FileStorage/storage"
	"database/sql"
	"fmt"
	"github.com/bradfitz/gomemcache/memcache"
	"testing"
)

type testCache struct{}

func (tc *testCache) Get(key string) (*memcache.Item, error) {
	return &memcache.Item{Key: key}, nil
}

func (tc *testCache) Set(_ *memcache.Item) error {
	return nil
}

type testDB struct {
	status string
}

func (td *testDB) Query(query string, _ ...any) (*sql.Rows, error) {
	switch td.status {
	case "exist":
		return nil, fmt.Errorf("pq: relation \"users\" already exists")
	case "not exist":
		return nil, fmt.Errorf("pq: relation \"users\" does not exist")
	case "other":
		return nil, fmt.Errorf("error")
	case "no error":
		return nil, nil
	}

	if query == "1" {
		return nil, fmt.Errorf("error")
	}
	return nil, nil
}

func (td *testDB) QueryRow(_ string, _ ...any) *sql.Row {
	return nil
}

func (td *testDB) Close() error {
	return nil
}

func TestTableExist(t *testing.T) {
	type args struct {
		db storage.Database
	}
	tests := []struct {
		name       string
		args       args
		wantErrStr error
		wantExists bool
	}{
		{args: args{db: &testDB{status: "no error"}}, name: "no error", wantErrStr: nil, wantExists: true},
		{args: args{db: &testDB{status: "not exist"}}, name: "not exist", wantErrStr: nil, wantExists: false},
		{args: args{db: &testDB{status: "other"}}, name: "other", wantErrStr: fmt.Errorf("error"), wantExists: false},
	}
	for _, tt := range tests {
		gotExists, gotErr := storage.TableExist(tt.args.db)
		if gotErr != nil {
			if gotErr.Error() != tt.wantErrStr.Error() {
				t.Errorf("%s: tableExist() gotErrStr = %v, want %v", tt.name, gotErr, tt.wantErrStr)
			}
		}
		if gotExists != tt.wantExists {
			t.Errorf("%s: tableExist() gotExists = %v, want %v", tt.name, gotExists, tt.wantExists)
		}
	}
}

func TestCreateTable(t *testing.T) {
	type args struct {
		db storage.Database
	}
	tests := []struct {
		name    string
		args    args
		wantErr error
	}{
		{args: args{db: &testDB{status: "no error"}}, name: "no error", wantErr: nil},
		{args: args{db: &testDB{status: "exist"}}, name: "exist", wantErr: nil},
		{args: args{db: &testDB{status: "other"}}, name: "other", wantErr: fmt.Errorf("error")},
	}
	for _, tt := range tests {
		gotErr := storage.CreateTable(tt.args.db)
		if gotErr != nil {
			if gotErr.Error() != tt.wantErr.Error() {
				t.Errorf("%s: tableExist() gotErrStr = %v, want %v", tt.name, gotErr, tt.wantErr)
			}
		}
	}
}

func TestSetUser(t *testing.T) { //TODO: uncompleted test
	storages := storage.Storage{&testCache{}, &testDB{}}
	tests := []struct {
		name     string
		login    string
		password string
		wantErr  error
	}{
		{name: "no error", login: "test", password: "test", wantErr: nil},
		{name: "exist", wantErr: nil},
		{name: "other", wantErr: fmt.Errorf("error")},
	}
	for _, tt := range tests {
		gotErr := storage.SetUser(storages, tt.login, tt.password)
		if gotErr != nil {
			if gotErr.Error() != tt.wantErr.Error() {
				t.Errorf("%s: tableExist() gotErrStr = %v, want %v", tt.name, gotErr, tt.wantErr)
			}
		}
	}

}
