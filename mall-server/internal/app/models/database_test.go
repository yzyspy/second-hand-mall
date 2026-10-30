package models

import (
	mysqlDriver "github.com/go-sql-driver/mysql"
	"mall-server/internal/app/config"
	"testing"
)

func TestMySQLDSN(t *testing.T) {
	t.Setenv("MYSQL_PASSWORD", "secret:@/?")
	c := config.Config{Gorm: config.Gorm{DBType: "mysql"}, MySQL: config.MySQL{Host: "127.0.0.1", Port: 3306, User: "mall", DBName: "mall_test", Password: "old"}}
	dsn, err := DatabaseDSN(&c)
	if err != nil {
		t.Fatal(err)
	}
	got, err := mysqlDriver.ParseDSN(dsn)
	if err != nil {
		t.Fatal(err)
	}
	if got.Passwd != "secret:@/?" || !got.ParseTime || got.Loc.String() != "Asia/Shanghai" || got.DBName != "mall_test" {
		t.Fatal("incorrect connection options")
	}
}
func TestOpenDatabaseMigrationSwitch(t *testing.T) {
	c := config.Config{Gorm: config.Gorm{DBType: "sqlite", MaxOpenConns: 1, MaxIdleConns: 1}, SQLite: config.SQLite{FilePath: ":memory:"}}
	db, err := OpenDatabase(&c)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { sqlDB, _ := db.DB(); sqlDB.Close() }()
	if db.Migrator().HasTable("product") {
		t.Fatal("unexpected automatic schema migration")
	}
	if err := MigrateSchema(db); err != nil {
		t.Fatal(err)
	}
	if !db.Migrator().HasTable("product") {
		t.Fatal("missing schema")
	}
	c.Gorm.DBType = "unknown"
	if _, err := OpenDatabase(&c); err == nil {
		t.Fatal("unsupported database accepted")
	}
}
