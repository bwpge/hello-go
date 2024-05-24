package common

import (
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"os"

	"github.com/charmbracelet/log"
	_ "github.com/mattn/go-sqlite3"
)

const DB_CONNECTION_STR = "db.sqlite"

const AUTH_STMT = `SELECT salt, hash, count FROM users WHERE username = ? LIMIT 1`

const CREATE_USER_STMT = `INSERT INTO users VALUES (?, ?, ?, ?)`

var Migrations embed.FS

type Database struct {
	db *sql.DB
}

func CreateDb() {
	log.Infof("creating new database `%s`", DB_CONNECTION_STR)
	os.Remove(DB_CONNECTION_STR)
	db, err := sql.Open("sqlite3", DB_CONNECTION_STR)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	// execute migration scripts
	err = fs.WalkDir(Migrations, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		log.Debugf("executing migration `%s`", path)
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		_, err = db.Exec(string(data))
		if err != nil {
			panic(err)
		}

		return nil
	})
	if err != nil {
		panic(err)
	}

	log.Info("successfully created database")
}

func DbConnect() *Database {
	db, err := sql.Open("sqlite3", DB_CONNECTION_STR)
	if err != nil {
		panic(err)
	}

	return &Database{db: db}
}

func (d *Database) Close() {
	if d.db != nil {
		log.Info("closing database connection")
		d.db.Close()
	}
}

func (d *Database) CreateUser(user string, pass string) error {
	fmt.Printf("Creating user `%s`", user)
	stmt, err := d.db.Prepare(CREATE_USER_STMT)
	if err != nil {
		return err
	}

	salt, hash, count := GenCreds(pass)
	_, err = stmt.Exec(user, salt, hash, count)
	if err != nil {
		return err
	}

	return nil
}

func (d *Database) AuthUser(user string, pass string) bool {
	log.Debugf("authenticating user `%s`", user)

	stmt, err := d.db.Prepare(AUTH_STMT)
	if err != nil {
		panic(err)
	}

	var saltStr string
	var hash string
	var count uint32
	err = stmt.QueryRow(user).Scan(&saltStr, &hash, &count)
	if err != nil {
		return false
	}
	salt, err := b64decode(saltStr)
	if err != nil {
		panic(err)
	}

	return HashPassword(pass, salt, count) == hash
}
