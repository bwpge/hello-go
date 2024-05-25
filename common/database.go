package common

import (
	"database/sql"
	"embed"
	"io/fs"
	"os"

	"github.com/charmbracelet/log"
	gonanoid "github.com/matoous/go-nanoid/v2"
	_ "github.com/mattn/go-sqlite3"
)

const (
	DB_CONNECTION_STR = "db.sqlite"
	AUTH_STMT         = `SELECT salt, hash, count FROM users WHERE username = ? LIMIT 1`
	CREATE_USER_STMT  = `INSERT INTO users VALUES (?, ?, ?, ?)`
	CREATE_TOKEN_STMT = `INSERT INTO tokens VALUES (?, ?)`
	CHECK_TOKEN_STMT  = `SELECT COUNT(*) FROM tokens WHERE key = ?`
)

var Migrations embed.FS

type Database struct {
	db *sql.DB
}

type UserData struct {
	Name  string `json:"username"`
	Salt  string `json:"salt"`
	Hash  string `json:"hash"`
	Count uint32 `json:"iter"`
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
	log.Debugf("creating user `%s`", user)

	salt, hash, count := GenCreds(pass)
	_, err := d.db.Exec(CREATE_USER_STMT, user, salt, hash, count)
	if err != nil {
		return err
	}

	return nil
}

func (d *Database) CreateToken(user string) (string, error) {
	log.Debugf("creating token for `%s`", user)

	token, err := gonanoid.New(30)
	if err != nil {
		return "", err
	}

	_, err = d.db.Exec(CREATE_TOKEN_STMT, token, user)
	if err != nil {
		return "", err
	}

	return token, nil
}

func (d *Database) IsValidToken(token string) bool {
	if token == "" {
		return false
	}

	var count int
	err := d.db.QueryRow(CHECK_TOKEN_STMT, token).Scan(&count)
	if err != nil {
		log.Error(err)
		return false
	}

	return count == 1
}

func (d *Database) GetUsers() ([]string, error) {
	rows, err := d.db.Query(`SELECT username FROM users`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []string
	for rows.Next() {
		var s string
		if err = rows.Scan(&s); err != nil {
			return nil, err
		}
		users = append(users, s)
	}

	return users, nil
}

func (d *Database) UserInfo(user string) (*UserData, error) {
	var name, salt, hash string
	var count uint32
	err := d.db.QueryRow(`SELECT * FROM users WHERE username = ?`, user).
		Scan(&name, &salt, &hash, &count)

	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return &UserData{
		Name:  name,
		Salt:  salt,
		Hash:  hash,
		Count: count,
	}, nil
}

func (d *Database) AuthUser(user string, pass string) bool {
	log.Debugf("authenticating user `%s`", user)

	var saltStr, hash string
	var count uint32
	err := d.db.QueryRow(AUTH_STMT, user).Scan(&saltStr, &hash, &count)
	if err != nil {
		return false
	}

	salt, err := b64decode(saltStr)
	if err != nil {
		panic(err)
	}

	return HashPassword(pass, salt, count) == hash
}
