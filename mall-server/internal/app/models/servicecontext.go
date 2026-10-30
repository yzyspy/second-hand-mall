package models

import (
	"context"
	"gorm.io/gorm"
	"mall-server/internal/app/config"
)

type ServiceContext struct {
	Config config.Config
	DB     *gorm.DB
}

func NewServiceContext(c context.Context) *ServiceContext {
	db, err := NewDB()
	if err != nil {
		panic(err)
	}
	//cache, _, _ := InitRedis(c)

	return &ServiceContext{
		DB: db,
		//	Cache: cache,
	}
}
