// Package storage defines the functions that use the database and cache
package storage

import (
	"github.com/bradfitz/gomemcache/memcache"
	"log"
	"strings"
)

var storage = InitStorages()

// GetUser takes the login and returns the password hash
func GetUser(login string) (string, error) {
	passwd, err := getFromCache(login)
	if err != nil {
		passwd, err = getFromDB(login)
		if err != nil {
			if err != nil {
				log.Printf("GetUser:getFromDB: %s", err)
			}
			return "", err
		}
	}
	return passwd, nil
}

func getFromCache(login string) (string, error) {
	password, err := storage.Cache.Get("user_" + login)
	if err != nil {
		if err.Error() != "memcache: cache miss" {
			log.Printf("getFromCache:%s", err)
		}
		return "", err
	}
	return string(password.Value), nil
}

func getFromDB(login string) (string, error) {
	var password string
	query := "SELECT password FROM users WHERE login = $1"
	if err := storage.PSQL.QueryRow(query, login).Scan(&password); err != nil {
		return "", err
	}
	return password, nil
}

func cacheUser(login, password string) error {
	err := storage.Cache.Set(&memcache.Item{Key: "user_" + login, Value: []byte(password)})
	return err
}

// SetUser writes user data to the database
func SetUser(login, password string) error {
	query := "INSERT INTO users (login, password) VALUES ($1, $2)"
	if _, err := storage.PSQL.Query(query, login, password); err != nil {
		if !strings.Contains(err.Error(), "duplicate key value violates unique constraint \"users_login_key\"") {
			log.Printf("SetUser:Query: %s", err)
		}
		return err
	}
	if err := cacheUser(login, password); err != nil {
		log.Printf("SetUser:cacheUser: %s", err)
		return err
	}
	return nil
}

func Close() {
	if err := storage.PSQL.Close(); err != nil {
		log.Printf("storage.Close: %s", err)
	}
}
