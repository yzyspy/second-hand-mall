package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/urfave/cli/v2"
	"golang.org/x/crypto/bcrypt"
	"mall-server/internal/app"
	"mall-server/internal/app/dao"
	"mall-server/internal/app/models"
	"mall-server/internal/app/router"
	"mall-server/pkg/logger"
)

func main() {
	ctx := logger.NewTagContext(context.Background(), "__main__")

	app := cli.NewApp()
	app.Name = "mall-server"
	app.Usage = "mall api Service"
	app.Commands = []*cli.Command{
		newWebCmd(ctx),
		newAdminCmd(ctx),
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Fatalf("应用启动失败: %v", err)
		panic("应用启动失败")
	}

	serviceContext := models.NewServiceContext(ctx)

	log.Println("启动完成成功,监听 8080 端口")
	logger.Infof("启动完成成功,监听 8080 端口")
	//启动gin
	r := router.App(ctx, serviceContext)
	r.Run(":8080")
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

func newAdminCmd(ctx context.Context) *cli.Command {
	return &cli.Command{
		Name:  "admin",
		Usage: "Admin management commands",
		Subcommands: []*cli.Command{
			{
				Name:  "create-admin",
				Usage: "Create a new admin user",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "config", Aliases: []string{"c"}, Usage: "Config file", Required: true},
					&cli.StringFlag{Name: "username", Aliases: []string{"u"}, Usage: "Admin username", Required: true},
					&cli.StringFlag{Name: "password", Aliases: []string{"p"}, Usage: "Admin password", Required: true},
				},
				Action: func(c *cli.Context) error {
					if err := app.Run(ctx, app.SetConfigFile(c.String("config"))); err != nil {
						return err
					}
					svc := models.NewServiceContext(ctx)
					hash, err := bcrypt.GenerateFromPassword([]byte(c.String("password")), bcrypt.DefaultCost)
					if err != nil {
						return fmt.Errorf("hash password: %w", err)
					}
					admin := dao.AdminUser{
						Username:     c.String("username"),
						PasswordHash: string(hash),
					}
					if err := svc.DB.Create(&admin).Error; err != nil {
						return fmt.Errorf("create admin: %w", err)
					}
					log.Printf("管理员 %q 创建成功 (ID=%d)\n", admin.Username, admin.ID)
					return nil
				},
			},
		},
	}
}
