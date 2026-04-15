package main

import (
	"context"
	"github.com/urfave/cli/v2"
	"log"
	"mall-server/internal/app"
	"mall-server/pkg/logger"
	"os"
)

func main() {
	ctx := logger.NewTagContext(context.Background(), "__main__")

	app := cli.NewApp()
	app.Name = "mall-server"
	app.Usage = "mall api Service"
	app.Commands = []*cli.Command{
		newWebCmd(ctx),
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Fatalf("应用启动失败: %v", err)
		panic("应用启动失败")
	}
}

func newWebCmd(ctx context.Context) *cli.Command {
	return &cli.Command{
		Name:  "web",
		Usage: "Run http server",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "config",
				Aliases:  []string{"c"},
				Usage:    "App configuration file(.json,.yaml,.toml)",
				Required: true,
			},
		},
		Action: func(c *cli.Context) error {
			return app.Run(ctx,
				app.SetConfigFile(c.String("config")))
		},
	}
}
