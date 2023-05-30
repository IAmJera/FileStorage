// Package storage defines the functions that use the database and cache
package storage

import (
	"fmt"
	"github.com/bradfitz/gomemcache/memcache"
	"log"
)

var storage = InitStorages()

// GetUser takes the login and returns the password hash
func GetUser(login string) (string, error) {
	passwd, err := getFromCache(login)
	if err != nil {
		passwd, err = getFromDB(login)
		if err != nil {
			log.Printf("GetUser:getFromDB: %s", err)
			return "", err
		}
	}
	return passwd, nil
}

func getFromCache(login string) (string, error) {
	res, err := storage.Cache.Get("user_" + login)
	if err != nil {
		if err.Error() != "memcache: cache miss" {
			log.Printf("getFromCache:%s", err)
		}
		return "", err
	}
	return string(res.Value), nil
}

func getFromDB(login string) (string, error) {
	var password string
	query := "SELECT password FROM `users` WHERE login = ?"
	if err := storage.MySQL.QueryRow(query, login).Scan(&password); err != nil {
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
	query := "INSERT INTO `users` (`login`, `password`) VALUES (?, ?)"
	if _, err := storage.MySQL.Exec(query, login, password); err != nil {
		log.Printf("SetUser:Query: %s", err)
		return fmt.Errorf("impossible insert record: %s", err)
	}
	if err := cacheUser(login, password); err != nil {
		log.Printf("SetUser:cacheUser: %s", err)
		return err
	}
	return nil
}
