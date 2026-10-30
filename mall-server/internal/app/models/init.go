package models

import (
	"fmt"
	mysqlDriver "github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
	"mall-server/internal/app/config"
	"mall-server/internal/app/dao"
	"mall-server/internal/app/gormx"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

func NewDB() (*gorm.DB, error) { return OpenDatabase(config.C) }

// DatabaseDSN builds a driver DSN without exposing credentials in logs.
func DatabaseDSN(c *config.Config) (string, error) {
	switch strings.ToLower(c.Gorm.DBType) {
	case "sqlite":
		if c.SQLite.FilePath == "" {
			return "", fmt.Errorf("sqlite.file_path is required")
		}
		return buildDSN(&c.SQLite), nil
	case "mysql":
		m := c.MySQL
		if m.Host == "" || m.User == "" || m.DBName == "" {
			return "", fmt.Errorf("mysql host, user and db_name are required")
		}
		if m.Port == 0 {
			m.Port = 3306
		}
		if m.Port < 1 || m.Port > 65535 {
			return "", fmt.Errorf("invalid MySQL port")
		}
		password := m.Password
		if value, ok := os.LookupEnv("MYSQL_PASSWORD"); ok {
			password = value
		}
		loc, err := time.LoadLocation("Asia/Shanghai")
		if err != nil {
			return "", err
		}
		params, err := url.ParseQuery(m.Parameters)
		if err != nil {
			return "", fmt.Errorf("invalid mysql parameters")
		}
		for key, vals := range params {
			if len(vals) != 1 {
				return "", fmt.Errorf("duplicate mysql parameter %s", key)
			}
		}
		// The driver parses connection options as well as server session variables.
		params.Set("parseTime", "true")
		if !params.Has("charset") {
			params.Set("charset", "utf8mb4")
		}
		if !params.Has("loc") {
			params.Set("loc", loc.String())
		}
		cfg := mysqlDriver.NewConfig()
		cfg.User = m.User
		cfg.Passwd = password
		cfg.Net = "tcp"
		cfg.Addr = net.JoinHostPort(m.Host, strconv.Itoa(m.Port))
		cfg.DBName = m.DBName
		cfg.Timeout = 5 * time.Second
		cfg.ReadTimeout = 30 * time.Second
		cfg.WriteTimeout = 30 * time.Second
		base := cfg.FormatDSN()
		sep := "?"
		if strings.Contains(base, "?") {
			sep = "&"
		}
		parsed, err := mysqlDriver.ParseDSN(base + sep + params.Encode())
		if err != nil {
			return "", fmt.Errorf("invalid mysql connection parameters")
		}
		return parsed.FormatDSN(), nil
	default:
		return "", fmt.Errorf("unsupported database type %q", c.Gorm.DBType)
	}
}

func OpenDatabase(c *config.Config) (*gorm.DB, error) {
	dsn, err := DatabaseDSN(c)
	if err != nil {
		return nil, err
	}
	if c.Gorm.TablePrefix != "" {
		return nil, fmt.Errorf("table_prefix must be empty: application queries use fixed table names")
	}
	db, err := gormx.New(&gormx.Config{DBType: c.Gorm.DBType, DSN: dsn, Debug: c.Gorm.Debug,
		MaxLifetime: c.Gorm.MaxLifetime, MaxOpenConns: c.Gorm.MaxOpenConns, MaxIdleConns: c.Gorm.MaxIdleConns})
	if err != nil {
		return nil, fmt.Errorf("connect to %s: %w", c.Gorm.DBType, err)
	}
	if c.Gorm.EnableAutoMigrate {
		if err := MigrateSchema(db); err != nil {
			sqlDB, _ := db.DB()
			sqlDB.Close()
			return nil, err
		}
	}
	return db, nil
}

func MigrateSchema(db *gorm.DB) error {
	if db.Dialector.Name() == "mysql" {
		db = db.Set("gorm:table_options", "ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin")
	}
	if err := db.AutoMigrate(&dao.SysUser{}, &dao.AdminUser{}, &dao.Product{}, &dao.UserFavorite{}, &dao.Conversation{}, &dao.Message{}); err != nil {
		return fmt.Errorf("migrate schema: %w", err)
	}
	return nil
}

func buildDSN(cfg *config.SQLite) string {
	dsn := cfg.FilePath
	if cfg.Parameters != "" {
		if strings.Contains(dsn, "?") {
			dsn += "&"
		} else {
			dsn += "?"
		}
		dsn += cfg.Parameters
	}
	return dsn
}
