package app

import (
	"context"
	"mall-server/internal/app/config"
	"mall-server/pkg/logger"
)

type Option func(*options)

type options struct {
	ConfigFile string
	WWWDir     string
	Version    string
}

func Init(ctx context.Context, opts ...Option) (func(), error) {

	logger.WithContext(ctx).Printf("Start server,#mode,#version ,#pid")

	var o options
	for _, opt := range opts {
		opt(&o)
	}

	config.MustLoad(o.ConfigFile)

	return InitLogger()
}

func Run(ctx context.Context, opts ...Option) error {
	_, err := Init(ctx, opts...)
	return err
}

// SetConfigFile 设定配置文件
func SetConfigFile(s string) Option {
	return func(o *options) {
		o.ConfigFile = s
	}
}
