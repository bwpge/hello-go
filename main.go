package main

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/urfave/cli/v2"
)

func main() {
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
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "username",
						Usage:   "`USERNAME` to connect with",
						Aliases: []string{"u"},
						Value:   "guest",
					},
					&cli.StringFlag{
						Name:    "password",
						Usage:   "`PASSWORD` for the account",
						Aliases: []string{"p"},
					},
				},
				Action: func(ctx *cli.Context) error {
					NewClient(uint16(ctx.Uint("port"))).Run(ctx.String("username"), ctx.String("password"))
					return nil
				},
			},
			{
				Name:    "database",
				Usage:   "manage the application database",
				Aliases: []string{"db"},
				Subcommands: []*cli.Command{
					{
						Name:  "create",
						Usage: "Create a new database (deletes existing one)",
						Action: func(ctx *cli.Context) error {
							CreateDb()
							fmt.Println("Database created!")
							return nil
						},
					},
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
