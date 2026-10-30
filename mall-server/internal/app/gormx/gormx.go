package gormx

import (
	"fmt"
	"github.com/glebarez/sqlite"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"strings"
	"time"
)

type Config struct {
	Debug        bool
	DBType       string
	DSN          string
	MaxLifetime  int
	MaxOpenConns int
	MaxIdleConns int
	TablePrefix  string
}

func New(c *Config) (*gorm.DB, error) {
	var dialector gorm.Dialector
	switch strings.ToLower(c.DBType) {
	case "sqlite":
		dialector = sqlite.Open(c.DSN)
	case "mysql":
		precision := 6
		dialector = mysql.New(mysql.Config{DSN: c.DSN, DefaultDatetimePrecision: &precision})
	default:
		return nil, fmt.Errorf("unsupported database type %q", c.DBType)
	}
	level := logger.Silent
	if c.Debug {
		level = logger.Info
	}
	db, err := gorm.Open(dialector, &gorm.Config{
		NamingStrategy: schema.NamingStrategy{TablePrefix: c.TablePrefix, SingularTable: true},
		Logger:         logger.Default.LogMode(level),
	})
	if err != nil {
		return nil, err
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(c.MaxOpenConns)
	sqlDB.SetMaxIdleConns(c.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(time.Duration(c.MaxLifetime) * time.Second)
	return db, nil
}
