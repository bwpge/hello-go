package main

import (
	"database/sql"
	"embed"
	"io/fs"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

const DB_CONNECTION_STR = "db.sqlite"

const AUTH_USER_STMT = `
SELECT username FROM users WHERE
    username = ? AND password = ?;`

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

func (d *Database) AuthUser(user string, pass string) bool {
	stmt, err := d.db.Prepare(AUTH_USER_STMT)
	if err != nil {
		panic(err)
	}

	var result string
	err = stmt.QueryRow(user, pass).Scan(&result)
	if err == sql.ErrNoRows {
		return false
	}
	if err != nil {
		panic(err)
	}

	return result == user
}
