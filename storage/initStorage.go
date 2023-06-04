// Package storage defines the functions that use the database and cache
package storage

import (
	"database/sql"
	"fmt"
	"github.com/bradfitz/gomemcache/memcache"
	_ "github.com/lib/pq"
	"log"
	"os"
	"time"
)

type Database interface {
	Query(query string, args ...any) (*sql.Rows, error)
	QueryRow(query string, args ...any) *sql.Row
	Close() error
}

type Memcached interface {
	Get(key string) (*memcache.Item, error)
	Set(*memcache.Item) error
}

// Storage stores the cache and database fields
type Storage struct {
	Cache Memcached
	PSQL  Database
}

// InitStorages initializes all storages and returns the structure with them
func InitStorages() Storage {
	store := Storage{}
	store.Cache = memcache.New(os.Getenv("CACHE_ADDRESS"))
	var err error
	if store.PSQL, err = initPSQL(); err != nil {
		log.Fatal(err)
	}

	exist, err := TableExist(store.PSQL)
	if err != nil {
		log.Panicf("InitStorages:TableExist: %s", err)
	}
	if !exist {
		if err = CreateTable(store.PSQL); err != nil {
			log.Panicf("InitStorages:CreateTable: %s", err)
		}
	}
	return store
}

func initPSQL() (*sql.DB, error) {
	addr := os.Getenv("PSQL_ADDRESS")
	port := os.Getenv("PSQL_PORT")
	login := os.Getenv("POSTGRES_USER")
	passwd := os.Getenv("POSTGRES_PASSWORD")
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=storage sslmode=disable",
		addr, port, login, passwd)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}
	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)

	if err = db.Ping(); err != nil {
		return nil, err
	}
	return db, err
}

func TableExist(db Database) (bool, error) {
	if _, err := db.Query("SELECT * FROM users;"); err != nil {
		if err.Error() == "pq: relation \"users\" does not exist" {
			return false, nil
		}
		log.Printf("tableExist:Query: %s", err)
		return false, err
	}
	return true, nil
}

func CreateTable(db Database) error {
	query := "CREATE TABLE users ( login VARCHAR(30) UNIQUE NOT NULL, password VARCHAR (64) NOT NULL);"
	if _, err := db.Query(query); err != nil {
		if err.Error() == "pq: relation \"users\" already exists" {
			return nil
		}
		log.Printf("prepareDB:Query: %s", err)
		return err
	}
	return nil
}
