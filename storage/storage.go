// Package storage defines the functions that use the database and cache
package storage

import (
	"fmt"
	"github.com/bradfitz/gomemcache/memcache"
	"log"
	"strings"
)

// GetUser takes the login and returns the password hash
func GetUser(storages Storage, login string) (string, error) {
	passwd, err := GetFromCache(storages, login)
	if err != nil {
		passwd, err = GetFromDB(false, storages, login)
		if err != nil {
			log.Printf("GetUser:getFromDB: %s", err)
			return "", err
		}
		if err = storages.Cache.Set(&memcache.Item{Key: "user_" + login, Value: []byte(passwd)}); err != nil {
			log.Printf("SetUser:Set: %s", err)
			return "", err
		}
	}
	return passwd, nil
}

func GetFromCache(storages Storage, login string) (string, error) {
	password, err := storages.Cache.Get("user_" + login)
	if err != nil {
		if err.Error() != "memcache: cache miss" {
			log.Printf("getFromCache:%s", err)
		}
		return "", err
	}
	return string(password.Value), nil
}

func GetFromDB(test bool, storages Storage, login string) (string, error) {
	var password string
	if test {
		password = strings.Split(login, " ")[0]
		err := fmt.Errorf(strings.Split(login, " ")[1])
		return password, err
	} else {
		query := "SELECT password FROM users WHERE login = $1"
		if err := storages.PSQL.QueryRow(query, login).Scan(&password); err != nil {
			return "", err
		}
	}
	return password, nil
}

// SetUser writes user data to the database
func SetUser(storages Storage, login, password string) error {
	query := "INSERT INTO users (login, password) VALUES ($1, $2)"
	if _, err := storages.PSQL.Query(query, login, password); err != nil {
		if !strings.Contains(err.Error(), "duplicate key value violates unique constraint \"users_login_key\"") {
			log.Printf("SetUser:Query: %s", err)
		}
		return err
	}

	err := storages.Cache.Set(&memcache.Item{Key: "user_" + login, Value: []byte(password)})
	if err != nil {
		log.Printf("SetUser:Set: %s", err)
	}
	return err
}

func Close(storages Storage) {
	if err := storages.PSQL.Close(); err != nil {
		log.Printf("storage.Close: %s", err)
	}
}
