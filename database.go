package main

import (
	"database/sql"
	"embed"
	"encoding/hex"
	"fmt"
	"io/fs"
	"os"

	"github.com/fatih/color"
	_ "github.com/mattn/go-sqlite3"
)

const DB_CONNECTION_STR = "db.sqlite"

const AUTH_STMT = `SELECT salt, hash FROM users WHERE username = ? LIMIT 1`

const CREATE_USER_STMT = `INSERT INTO users VALUES (?, ?, ?)`

//go:embed sql/*.sql
var scripts embed.FS

type Database struct {
	db *sql.DB
}

func CreateDb() {
	os.Remove("./foo.db")
	db, err := sql.Open("sqlite3", DB_CONNECTION_STR)
	if err != nil {
		panic(err)
	}

	// execute migration scripts
	err = fs.WalkDir(scripts, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

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

	db.Close()
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
		d.db.Close()
	}
}

func (d *Database) CreateUser(user string, pass string) error {
	fmt.Printf("Creating user `%s`", user)
	stmt, err := d.db.Prepare(CREATE_USER_STMT)
	if err != nil {
		return err
	}

	salt := hex.EncodeToString(GenerateSalt())
	hash := HashPassword(pass, salt)

	_, err = stmt.Exec(user, salt, hash)
	if err != nil {
		return err
	}

	return nil
}

func (d *Database) AuthUser(user string, pass string) bool {
	color.HiBlack("Authenticating user `%s`", user)

	stmt, err := d.db.Prepare(AUTH_STMT)
	if err != nil {
		panic(err)
	}

	var salt string
	var hash string
	err = stmt.QueryRow(user).Scan(&salt, &hash)
	if err != nil {
		return false
	}

	return HashPassword(pass, salt) == hash
}
