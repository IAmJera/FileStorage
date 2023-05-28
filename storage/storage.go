package storage

import (
	"FileStorage/app/general"
	"context"
	"fmt"
	"github.com/bradfitz/gomemcache/memcache"
	"log"
)

var storage = InitStorages()

func GetUser(login string) (string, error) {
	passwd, err := getFromCache(login)
	if err != nil {
		passwd, err = getFromDB(login)
		if err != nil {
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
	rows, err := storage.MySQL.Query("SELECT * FROM `users` WHERE `login` = ?", login)
	if err != nil {
		log.Printf("getFromDB:Query: %s", err)
		return "", err
	}
	defer general.CloseFile(rows)
	if err = rows.Err(); err != nil {
		log.Printf("getFromDB:Err: %s", err)
		return "", err
	}

	var username, password string
	if rows.Next() {
		if err = rows.Scan(&username, &password); err != nil {
			log.Printf("getFromDB:Next: %s", err)
			return "", err
		}
	}
	if err = cacheUser(username, password); err != nil {
		log.Printf("getFromDB:cacheUser: %s", err)
		return "", err
	}
	return password, nil
}

func cacheUser(login, password string) error {
	if err := storage.Cache.Set(&memcache.Item{Key: "user_" + login, Value: []byte(password)}); err != nil {
		return err
	}
	return nil
}

func SetUser(login, password string) error {
	query := "INSERT INTO `users` (`login`, `password`) VALUES (?, ?)"
	if _, err := storage.MySQL.ExecContext(context.Background(), query, login, password); err != nil {
		log.Printf("SetUser:Query: %s", err)
		return fmt.Errorf("impossible insert record: %s", err)
	}
	if err := cacheUser(login, password); err != nil {
		log.Printf("SetUser:cacheUser: %s", err)
		return err
	}
	return nil
}
