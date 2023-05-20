package storage

import (
	account "FileStorage/user"
	"github.com/bradfitz/gomemcache/memcache"
)

var cache = memcache.New("localhost:11211")

func SetUser(user account.User) error {
	//TODO: database support
	if err := cacheUser(user); err != nil {
		return err
	}
	return nil
}

func cacheUser(user account.User) error {
	if err := cache.Set(&memcache.Item{Key: "user_" + user.Login, Value: account.Hash(user.Password)}); err != nil {
		return err
	}
	return nil
}

func GetUser(login string) (account.User, error) {
	user := account.User{Login: login}
	var err error
	if user.Password, err = getFromCache(login); err != nil {
		if user.Password, err = getFromDB(login); err != nil {
			return account.User{}, err
		}
	}
	return user, nil
}

func getFromCache(login string) (string, error) {
	res, err := cache.Get("user_" + login)
	if err != nil {
		return "", err
	}
	return string(res.Value), nil
}

func getFromDB(login string) (string, error) {

	//cache()
	return "", nil
}
