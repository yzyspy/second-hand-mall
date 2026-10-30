package dbmigration

import (
	"database/sql"
	"fmt"
	mysqlDriver "github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
	"mall-server/internal/app/config"
	"mall-server/internal/app/dao"
	"mall-server/internal/app/models"
	"os"
	"testing"
	"time"
)

// Uses a fresh database, never the configured application database.
func TestMySQLMigrationIntegration(t *testing.T) {
	path := os.Getenv("MYSQL_TEST_CONFIG")
	if path == "" {
		t.Skip("set MYSQL_TEST_CONFIG to an ignored MySQL config to run integration")
	}
	config.MustLoad(path)
	cfg := *config.C
	cfg.Gorm.DBType = "mysql"
	cfg.Gorm.EnableAutoMigrate = false
	cfg.Gorm.Debug = false
	dsn, err := models.DatabaseDSN(&cfg)
	if err != nil {
		t.Fatal(err)
	}
	mc, err := mysqlDriver.ParseDSN(dsn)
	if err != nil {
		t.Fatal(err)
	}
	mc.DBName = ""
	server, err := sql.Open("mysql", mc.FormatDSN())
	if err != nil {
		t.Fatal(err)
	}
	defer server.Close()
	name := fmt.Sprintf("mall_migration_test_%d", time.Now().UnixNano())
	if _, err := server.Exec("CREATE DATABASE `" + name + "` CHARACTER SET utf8mb4 COLLATE utf8mb4_bin"); err != nil {
		t.Fatal(err)
	}
	defer server.Exec("DROP DATABASE `" + name + "`")
	cfg.MySQL.DBName = name
	dst, err := models.OpenDatabase(&cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { db, _ := dst.DB(); db.Close() }()
	if err := models.MigrateSchema(dst); err != nil {
		t.Fatal(err)
	}
	src := testDB(t)
	stamp := time.Date(2026, 10, 30, 12, 34, 56, 123456789, time.FixedZone("CST", 8*3600))
	users := []dao.SysUser{{Model: gorm.Model{ID: 10, CreatedAt: stamp, UpdatedAt: stamp}, UserName: "buyer", NickName: "中文🙂"}, {Model: gorm.Model{ID: 25, CreatedAt: stamp, UpdatedAt: stamp}, UserName: "seller"}}
	if err := src.Create(&users).Error; err != nil {
		t.Fatal(err)
	}
	deleted := dao.SysUser{Model: gorm.Model{ID: 50}, UserName: "deleted"}
	src.Create(&deleted)
	src.Delete(&deleted)
	src.Create(&dao.AdminUser{Username: "admin", PasswordHash: "original-hash"})
	p := dao.Product{Model: gorm.Model{ID: 30}, UserId: 25, Title: "商品🙂", Price: 12.34, Description: ""}
	if err := src.Create(&p).Error; err != nil {
		t.Fatal(err)
	}
	src.Create(&dao.UserFavorite{UserID: 10, ProductID: 30})
	conv := dao.Conversation{ProductID: 30, BuyerID: 10, SellerID: 25, LastAt: stamp}
	src.Create(&conv)
	src.Create(&dao.Message{ConversationID: conv.ID, SenderID: 25, Content: "你好🙂", IsRead: false})
	data, err := Read(src)
	if err != nil {
		t.Fatal(err)
	}
	// A MyISAM target cannot promise rollback: reject it before any insert.
	if err := dst.Exec("ALTER TABLE admin_user ENGINE=MyISAM").Error; err != nil {
		t.Fatal(err)
	}
	if err := Copy(dst, data); err == nil {
		t.Fatal("accepted nontransactional table")
	}
	var partial int64
	dst.Unscoped().Model(&dao.SysUser{}).Count(&partial)
	if partial != 0 {
		t.Fatal("inserted before engine check")
	}
	if err := dst.Exec("ALTER TABLE admin_user ENGINE=InnoDB").Error; err != nil {
		t.Fatal(err)
	}
	// Simulate a failure after the first table to prove real MySQL rollback.
	if err := dst.Exec("CREATE TRIGGER reject_admin BEFORE INSERT ON admin_user FOR EACH ROW SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT = 'test rejection'").Error; err != nil {
		t.Fatal(err)
	}
	if err := Copy(dst, data); err == nil {
		t.Fatal("expected MySQL insertion failure")
	}
	dst.Unscoped().Model(&dao.SysUser{}).Count(&partial)
	if partial != 0 {
		t.Fatal("MySQL rollback left partial rows")
	}
	if err := dst.Exec("DROP TRIGGER reject_admin").Error; err != nil {
		t.Fatal(err)
	}
	if err := Copy(dst, data); err != nil {
		t.Fatal(err)
	}
	if err := Copy(dst, data); err == nil {
		t.Fatal("nonempty MySQL accepted")
	}
	rows, err := dao.ListConversations(dst, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || rows[0].UnreadCount != 1 {
		t.Fatalf("incorrect chat aggregation: %v", rows)
	}
	next := dao.SysUser{UserName: "new"}
	if err := dst.Create(&next).Error; err != nil {
		t.Fatal(err)
	}
	if next.ID <= 50 {
		t.Fatal("auto increment did not advance")
	}
	var got dao.SysUser
	if err := dst.First(&got, 10).Error; err != nil {
		t.Fatal(err)
	}
	if !got.CreatedAt.Equal(stamp.Truncate(time.Microsecond)) {
		t.Fatal("time or timezone changed")
	}
	var dropped int64
	dst.Unscoped().Model(&dao.SysUser{}).Where("id = ? AND deleted_at IS NOT NULL", 50).Count(&dropped)
	if dropped != 1 {
		t.Fatal("soft deleted row lost")
	}
}
