// migrate-db is an offline, one-shot migration. It never starts the web server.
package main

import (
	"database/sql"
	"flag"
	"fmt"
	mysqlDriver "github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
	"mall-server/internal/app/config"
	"mall-server/internal/app/models"
	"mall-server/internal/dbmigration"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	_ "time/tzdata" // Migration also needs named timezones in minimal images.
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "migration:", err)
		os.Exit(1)
	}
}
func openSource(path string) (*gorm.DB, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	if _, err := os.Stat(abs); err != nil {
		return nil, err
	}
	u := url.URL{Scheme: "file", Path: abs, RawQuery: "mode=ro"}
	return models.OpenDatabase(&config.Config{Gorm: config.Gorm{DBType: "sqlite", MaxOpenConns: 1, MaxIdleConns: 1}, SQLite: config.SQLite{FilePath: u.String()}})
}
func closeDB(db *gorm.DB) {
	s, err := db.DB()
	if err == nil {
		s.Close()
	}
}
func run() error {
	cfgPath := flag.String("config", "configs/mysql.yaml", "target MySQL config")
	source := flag.String("source", "second-hand-mall.db", "source SQLite file")
	backup := flag.String("backup", "", "new SQLite snapshot file (required with --apply)")
	apply := flag.Bool("apply", false, "create target schema and import; stop all app writers first")
	create := flag.Bool("create-database", false, "create target database if missing (requires CREATE privilege)")
	info := flag.Bool("check-target", false, "inspect target database without modifying it")
	check := flag.Bool("check", false, "validate source only, without connecting to MySQL")
	stopped := flag.Bool("confirm-stopped", false, "confirm all services writing SQLite and MySQL have been stopped")
	flag.Parse()
	if *info {
		config.MustLoad(*cfgPath)
		target := *config.C
		target.Gorm.EnableAutoMigrate = false
		target.Gorm.Debug = false
		db, err := models.OpenDatabase(&target)
		if err != nil {
			return err
		}
		defer closeDB(db)
		var version string
		if err := db.Raw("SELECT VERSION()").Scan(&version).Error; err != nil {
			return err
		}
		fmt.Println("MySQL version:", version)
		tables, err := db.Migrator().GetTables()
		if err != nil {
			return err
		}
		for _, table := range tables {
			var n int64
			if err := db.Table(table).Count(&n).Error; err != nil {
				return err
			}
			fmt.Printf("%s: %d records\n", table, n)
		}
		return nil
	}
	if *check && *apply {
		return fmt.Errorf("choose --check or --apply")
	}
	if !*check && !*apply {
		return fmt.Errorf("choose --check or --apply")
	}
	if *apply && (!*stopped || *backup == "") {
		return fmt.Errorf("--apply requires --confirm-stopped and --backup; stop all application writers first")
	}
	src, err := openSource(*source)
	if err != nil {
		return err
	}
	defer closeDB(src)
	if *apply {
		abs, err := filepath.Abs(*backup)
		if err != nil {
			return err
		}
		f, err := os.OpenFile(abs, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
		if err != nil {
			return fmt.Errorf("create backup (must not exist): %w", err)
		}
		if err := f.Close(); err != nil {
			return err
		}
		if err := src.Exec("VACUUM INTO ?", abs).Error; err != nil {
			return fmt.Errorf("backup failed; remove incomplete backup before retry: %w", err)
		}
		snapshot, err := openSource(abs)
		if err != nil {
			return err
		}
		defer closeDB(snapshot)
		src = snapshot
		fmt.Println("SQLite snapshot created:", abs)
	}
	data, err := dbmigration.Read(src)
	if err != nil {
		return err
	}
	for _, table := range data {
		fmt.Printf("%s: %d records\n", table.Name, len(table.Rows))
	}
	if *check {
		fmt.Println("Source validation passed; no database modified.")
		return nil
	}
	config.MustLoad(*cfgPath)
	target := *config.C
	if target.Gorm.DBType != "mysql" {
		return fmt.Errorf("target db_type must be mysql")
	}
	target.Gorm.EnableAutoMigrate = false
	target.Gorm.Debug = false
	if *create {
		dsn, err := models.DatabaseDSN(&target)
		if err != nil {
			return err
		}
		mc, err := mysqlDriver.ParseDSN(dsn)
		if err != nil {
			return err
		}
		name := mc.DBName
		mc.DBName = ""
		server, err := sql.Open("mysql", mc.FormatDSN())
		if err != nil {
			return err
		}
		defer server.Close()
		if _, err := server.Exec("CREATE DATABASE IF NOT EXISTS `" + strings.ReplaceAll(name, "`", "``") + "` CHARACTER SET utf8mb4 COLLATE utf8mb4_bin"); err != nil {
			return err
		}
	}
	dst, err := models.OpenDatabase(&target)
	if err != nil {
		return err
	}
	defer closeDB(dst)
	// Refuse nonempty target before running schema changes.
	tables, err := dst.Migrator().GetTables()
	if err != nil {
		return err
	}
	for _, table := range tables {
		known := false
		for _, sourceTable := range data {
			if sourceTable.Name == table {
				known = true
			}
		}
		if !known {
			return fmt.Errorf("unexpected target table %s; use a dedicated empty database", table)
		}
		var n int64
		if err := dst.Table(table).Count(&n).Error; err != nil {
			return err
		}
		if n != 0 {
			return fmt.Errorf("target table %s is not empty", table)
		}
	}
	if err := models.MigrateSchema(dst); err != nil {
		return err
	}
	if err := dbmigration.Copy(dst, data); err != nil {
		return err
	}
	fmt.Println("Migration committed; all six tables verified (timestamps normalized to UTC microseconds).")
	return nil
}
