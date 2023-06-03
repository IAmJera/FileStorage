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

// Storage stores the cache and database fields
type Storage struct {
	Cache *memcache.Client
	PSQL  *sql.DB
}

// InitStorages initializes all storages and returns the structure with them
func InitStorages() Storage {
	store := Storage{}
	store.Cache = memcache.New(os.Getenv("CACHE_ADDRESS"))
	var err error
	if store.PSQL, err = initPSQL(); err != nil {
		log.Fatal(err)
	}
	if err = prepareDB(store.PSQL); err != nil {
		log.Fatal(err)
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

func prepareDB(db *sql.DB) error {
	if exist, err := tableExist(db); err != nil {
		return err
	} else if exist {
		return nil
	}

	query := "CREATE TABLE users ( login VARCHAR(30) UNIQUE NOT NULL, password VARCHAR (64) NOT NULL);"
	if _, err := db.Query(query); err != nil {
		log.Printf("prepareDB:Query: %s", err)
		if err.Error() != "pq: relation \"users\" already exists" {
			return err
		}
	}
	return nil
}

func tableExist(db *sql.DB) (bool, error) {
	if _, err := db.Query("SELECT * FROM users;"); err != nil {
		log.Printf("tableExist:Query: %s", err)
		if err.Error() == "pq: relation \"users\" does not exist" {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
