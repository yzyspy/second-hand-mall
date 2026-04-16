package models

import (
	"fmt"
	"gorm.io/gorm"
	"mall-server/internal/app/config"
	"mall-server/internal/app/gormx"
	"strings"
)

func NewDB() (*gorm.DB, error) {
	c := config.C
	dsn := buildDSN(&c.SQLite)
	con, err := gormx.New(&gormx.Config{
		DBType:       c.Gorm.DBType,
		DSN:          dsn,
		Debug:        true,
		MaxLifetime:  100000,
		MaxOpenConns: 60,
		MaxIdleConns: 20,
		TablePrefix:  "",
	})
	if err != nil {
		panic(fmt.Sprintf("db init error dsn ,path: %v error:%v", c.SQLite.FilePath, err))
	}
	return con, nil
}

func buildDSN(cfg *config.SQLite) string {
	dsn := cfg.FilePath
	if cfg.Parameters != "" {
		// 根据已有参数决定连接符
		if strings.Contains(dsn, "?") {
			dsn += "&" + cfg.Parameters
		} else {
			dsn += "?" + cfg.Parameters
		}
	}
	return dsn
}

// 简单的参数字符串解析，生产环境建议使用 url.ParseQuery
func parseParams(paramStr string) map[string]string {
	params := make(map[string]string)
	pairs := strings.Split(paramStr, "&")
	for _, pair := range pairs {
		kv := strings.SplitN(pair, "=", 2)
		if len(kv) == 2 {
			params[kv[0]] = kv[1]
		}
	}
	return params
}
