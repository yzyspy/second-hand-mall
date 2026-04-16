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
	dsn := buildDSN(&c.VpnMySQL)
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
		panic(fmt.Sprintf("db init error dsn : %s %v", dsn, err))
	}
	return con, nil
}

func buildDSN(cfg *config.MySQL) string {
	//c := mysql.NewConfig()
	//c.User = cfg.User
	//c.Passwd = cfg.Password
	//c.Net = "tcp"
	//c.Addr = fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	//c.DBName = cfg.DBName
	//// 解析参数字符串为 map（如果需要）
	//if cfg.Parameters != "" {
	//	c.Params = parseParams(cfg.Parameters)
	//}
	//return c.FormatDSN()
	return ""
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
