package main

import (
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/urfave/cli/v2"
)

func main() {
	userPassFlags := []cli.Flag{
		&cli.StringFlag{
			Name:    "username",
			Usage:   "`USERNAME` to connect with",
			Aliases: []string{"u"},
		},
		&cli.StringFlag{
			Name:    "password",
			Usage:   "`PASSWORD` for the account",
			Aliases: []string{"p"},
		},
	}
	app := &cli.App{
		Name:  "hello-go",
		Usage: "a basic client-server application",
		Flags: []cli.Flag{
			&cli.UintFlag{
				Name:  "port",
				Value: 3000,
				Action: func(ctx *cli.Context, value uint) error {
					if value > 65535 {
						return errors.New("port must be within range 0-65535")
					}
					return nil
				},
				Usage: "`PORT` to serve or connect on",
			},
		},
		Commands: []*cli.Command{
			{
				Name:  "server",
				Usage: "Start a server",
				Action: func(ctx *cli.Context) error {
					s := NewServer(uint16(ctx.Uint("port")))
					log.Fatal(s.Run())
					return nil
				},
			},
			{
				Name:  "client",
				Usage: "Start a client",
				Flags: userPassFlags,
				Action: func(ctx *cli.Context) error {
					user := ctx.String("username")
					if user == "" {
						user = "guest"
					}
					NewClient(uint16(ctx.Uint("port"))).Run(user, ctx.String("password"))
					return nil
				},
			},
			{
				Name:    "register",
				Usage:   "Create a new user in the database",
				Flags:   userPassFlags,
				Aliases: []string{"r"},
				Action: func(ctx *cli.Context) error {
					user := ctx.String("username")
					if user == "" {
						return errors.New("username must not be empty")
					}
					if user == "guest" {
						return errors.New("`guest` is a reserved name and cannot be used")
					}

					pass := ctx.String("password")
					if pass == "" {
						return errors.New("password must not be empty")
					}

					db := DbConnect()
					return db.CreateUser(user, pass)
				},
			},
			{
				Name:    "gen-database",
				Usage:   "Create a new database (deletes existing one)",
				Aliases: []string{"db"},
				Action: func(ctx *cli.Context) error {
					CreateDb()
					fmt.Println("Database created!")
					return nil
				},
			},
			{
				Name:    "gen-password",
				Usage:   "Create salt and hash values for passwords",
				Aliases: []string{"pw"},
				Args:    true,
				Action: func(ctx *cli.Context) error {
					value := ctx.Args().First()
					if value == "" {
						return errors.New("password must not be empty")
					}

					salt := hex.EncodeToString(GenerateSalt())
					hash := HashPassword(value, salt)

					fmt.Printf("SALT: %v\nHASH: %v\n", salt, hash)

					return nil
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
