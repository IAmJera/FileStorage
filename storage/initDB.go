package storage

import (
	"context"
	"database/sql"
	"github.com/bradfitz/gomemcache/memcache"
	"github.com/go-sql-driver/mysql"
	"log"
	"time"
)

type Storage struct {
	Cache *memcache.Client
	MySQL *sql.DB
}

func InitStorages() Storage {
	strg := Storage{}
	strg.Cache = memcache.New("localhost:11211")
	var err error
	if strg.MySQL, err = initMySQL(); err != nil {
		log.Fatal(err)
	}
	if err := prepareDB(strg.MySQL); err != nil {
		log.Fatal(err)
	}
	return strg
}

func initMySQL() (*sql.DB, error) {
	auth := mysql.Config{
		User:                 "root",    //os.Getenv("MYSQL_USER"),
		Passwd:               "example", //os.Getenv("MYSQL_PASSWORD"),
		Net:                  "tcp",
		Addr:                 "192.168.88.150:3306", //os.Getenv("MYSQL_ADDRESS") + ":" + os.Getenv("MYSQL_PORT"),
		DBName:               "storage",             //os.Getenv("MYSQL_DB"),
		AllowNativePasswords: true,
	}
	db, err := sql.Open("mysql", auth.FormatDSN())
	if err != nil {
		return nil, err
	}
	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)

	if err := db.Ping(); err != nil {
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

	query := "CREATE TABLE `users` ( `login` varchar(30), `password` varchar(255));"
	if _, err := db.ExecContext(context.Background(), query); err != nil {
		if err.Error() != "Error 1050 (42S01): Table 'users' already exists" {
			return err
		}
	}
	return nil
}

func tableExist(db *sql.DB) (bool, error) {
	if _, err := db.Query("SELECT * FROM `users`;"); err != nil {
		if err.Error() == "Error 1146 (42S02): Table 'storage.users' doesn't exist" {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
