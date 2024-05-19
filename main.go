package main

import (
	"embed"
	"errors"
	"fmt"
	"hello-go/client"
	"hello-go/common"
	"hello-go/server"
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/urfave/cli/v2"
)

//go:embed sql/*.sql
var migrations embed.FS

func main() {
	setupLogging()

	// see: https://stackoverflow.com/a/67357103
	common.Migrations = migrations

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
				Name:    "server",
				Usage:   "Start a server",
				Aliases: []string{"s"},
				Action: func(ctx *cli.Context) error {
					s := server.New(uint16(ctx.Uint("port")))
					log.Fatal(s.Run())
					return nil
				},
			},
			{
				Name:    "client",
				Usage:   "Start a client",
				Aliases: []string{"c"},
				Flags:   userPassFlags,
				Action: func(ctx *cli.Context) error {
					user := ctx.String("username")
					if user == "" {
						user = "guest"
					}
					c := client.New(uint16(ctx.Uint("port")))
					c.Run(user, ctx.String("password"))
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

					db := common.DbConnect()
					return db.CreateUser(user, pass)
				},
			},
			{
				Name:    "gen-database",
				Usage:   "Create a new database (deletes existing one)",
				Aliases: []string{"gendb"},
				Action: func(ctx *cli.Context) error {
					common.CreateDb()
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

					salt, hash, count := common.GenCreds(value)
					fmt.Printf("SALT:  %v\nHASH:  %v\nCOUNT: %v\n", salt, hash, count)

					return nil
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func setupLogging() {
	log.SetLevel(log.DebugLevel)
	styles := log.DefaultStyles()
	styles.Levels[log.DebugLevel].
		Foreground(lipgloss.Color("6")).
		UnsetMaxWidth()
	styles.Levels[log.InfoLevel].
		Foreground(lipgloss.Color("4")).
		UnsetMaxWidth().Padding(0, 1, 0, 0)
	styles.Levels[log.WarnLevel].
		Foreground(lipgloss.Color("3")).
		UnsetMaxWidth().
		Padding(0, 1, 0, 0)
	styles.Levels[log.ErrorLevel].
		Foreground(lipgloss.Color("1")).
		UnsetMaxWidth()
	styles.Levels[log.FatalLevel].
		Foreground(lipgloss.Color("1")).
		Reverse(true).
		UnsetMaxWidth()
	log.SetStyles(styles)
}
