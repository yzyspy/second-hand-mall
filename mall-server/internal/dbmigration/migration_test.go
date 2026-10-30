package dbmigration

import (
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"mall-server/internal/app/dao"
	"mall-server/internal/app/models"
	"path/filepath"
	"testing"
	"time"
)

func testDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "test.db")), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s, _ := db.DB(); s.Close() })
	if err := models.MigrateSchema(db); err != nil {
		t.Fatal(err)
	}
	return db
}
func TestCopyPreservesDeletedRowsAndRejectsNonempty(t *testing.T) {
	src, dst := testDB(t), testDB(t)
	user := dao.SysUser{Model: gorm.Model{ID: 41, CreatedAt: time.Now()}, UserName: "旧用户", IsBanned: true}
	if err := src.Create(&user).Error; err != nil {
		t.Fatal(err)
	}
	if err := src.Delete(&user).Error; err != nil {
		t.Fatal(err)
	}
	data, err := Read(src)
	if err != nil {
		t.Fatal(err)
	}
	if err := Copy(dst, data); err != nil {
		t.Fatal(err)
	}
	var got dao.SysUser
	if err := dst.Unscoped().First(&got, 41).Error; err != nil {
		t.Fatal(err)
	}
	if !got.DeletedAt.Valid || !got.IsBanned || got.UserName != "旧用户" {
		t.Fatal("lost data")
	}
	if err := Copy(dst, data); err == nil {
		t.Fatal("accepted nonempty target")
	}
}
func TestReadRejectsLongStrings(t *testing.T) {
	src := testDB(t)
	user := dao.SysUser{UserName: string(make([]byte, 51))}
	if err := src.Create(&user).Error; err != nil {
		t.Fatal(err)
	}
	if _, err := Read(src); err == nil {
		t.Fatal("accepted oversized username")
	}
}
func TestCopyRollsBack(t *testing.T) {
	src, dst := testDB(t), testDB(t)
	src.Create(&dao.SysUser{UserName: "one"})
	src.Create(&dao.AdminUser{Username: "admin", PasswordHash: "hash"})
	data, err := Read(src)
	if err != nil {
		t.Fatal(err)
	}
	dst.Exec("CREATE TRIGGER reject_admin BEFORE INSERT ON admin_user BEGIN SELECT RAISE(ABORT, 'test rejection'); END")
	if err := Copy(dst, data); err == nil {
		t.Fatal("expected failure")
	}
	var count int64
	dst.Unscoped().Model(&dao.SysUser{}).Count(&count)
	if count != 0 {
		t.Fatal("partial copy committed")
	}
}
