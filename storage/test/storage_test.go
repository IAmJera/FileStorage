package storage_test

import (
	"FileStorage/storage"
	"database/sql"
	"fmt"
	"github.com/bradfitz/gomemcache/memcache"
	"testing"
)

type testCache struct {
	status string
}

func (tc *testCache) Get(_ string) (*memcache.Item, error) {
	switch tc.status {
	case "miss":
		return nil, fmt.Errorf("memcache: cache miss")
	case "nil":
		return &memcache.Item{Value: []byte("password")}, nil
	}
	return nil, nil
}

func (tc *testCache) Set(_ *memcache.Item) error {
	switch tc.status {
	case "error":
		return fmt.Errorf("error")
	case "nil":
		return nil
	}
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

func (td *testDB) QueryRow(query string, args ...any) *sql.Row {
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
				t.Errorf("%s: TableExist() gotErrStr = %v, want %v", tt.name, gotErr, tt.wantErrStr)
			}
		}
		if gotExists != tt.wantExists {
			t.Errorf("%s: TableExist() gotExists = %v, want %v", tt.name, gotExists, tt.wantExists)
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
				t.Errorf("%s: CreateTable() gotErrStr = %v, want %v", tt.name, gotErr, tt.wantErr)
			}
		}
	}
}

func TestSetUser(t *testing.T) {
	tests := []struct {
		name    string
		strg    storage.Storage
		wantErr error
	}{
		{
			name:    "duplicate",
			strg:    storage.Storage{Cache: &testCache{}, PSQL: &testDB{status: "duplicate"}},
			wantErr: nil},
		{
			name:    "other",
			strg:    storage.Storage{Cache: &testCache{}, PSQL: &testDB{status: "other"}},
			wantErr: fmt.Errorf("error")},
		{
			name:    "cache_error",
			strg:    storage.Storage{Cache: &testCache{status: "error"}, PSQL: &testDB{"no error"}},
			wantErr: fmt.Errorf("error")},
		{
			name:    "cache_nil",
			strg:    storage.Storage{Cache: &testCache{status: "nil"}, PSQL: &testDB{status: "no error"}},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		gotErr := storage.SetUser(tt.strg, "login", "password")
		if gotErr != nil {
			if gotErr.Error() != tt.wantErr.Error() {
				t.Errorf("%s: SetUser() gotErrStr = %v, want %v", tt.name, gotErr, tt.wantErr)
			}
		}
	}
}

func TestGetFromCache(t *testing.T) {
	tests := []struct {
		name    string
		strg    storage.Storage
		wantErr error
		wantRes string
	}{
		{
			name:    "cache_error",
			strg:    storage.Storage{Cache: &testCache{status: "miss"}, PSQL: &testDB{}},
			wantErr: fmt.Errorf("memcache: cache miss"),
			wantRes: ""},
		{
			name:    "cache_nil",
			strg:    storage.Storage{Cache: &testCache{status: "nil"}, PSQL: &testDB{}},
			wantErr: nil,
			wantRes: "password"},
	}
	for _, tt := range tests {
		gotRes, gotErr := storage.GetFromCache(tt.strg, "login")
		if gotErr != nil {
			if gotErr.Error() != tt.wantErr.Error() {
				t.Errorf("%s: GetFromCache() gotErr = %v, want %v", tt.name, gotErr, tt.wantErr)
			}
		} else if gotRes != "password" {
			t.Errorf("%s: GetFromCache() gotRes = %v, want %v", tt.name, gotRes, tt.wantRes)
		}
	}
}

//TODO: GetFromDB

//TODO: GetUser
