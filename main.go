package main

import (
	"errors"
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
					NewServer(uint16(ctx.Uint("port"))).Run()
					return nil
				},
			},
			{
				Name:  "client",
				Usage: "Start a client",
				Action: func(ctx *cli.Context) error {
					NewClient(uint16(ctx.Uint("port"))).Run()
					return nil
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
