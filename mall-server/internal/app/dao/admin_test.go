package dao

import (
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func newTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.AutoMigrate(&AdminUser{}, &SysUser{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

func TestAdminUserCreate(t *testing.T) {
	db := newTestDB(t)
	admin := AdminUser{Username: "admin", PasswordHash: "hash"}
	if err := db.Create(&admin).Error; err != nil {
		t.Fatalf("create admin: %v", err)
	}
	if admin.ID == 0 {
		t.Fatal("expected non-zero ID")
	}
}

func TestAdminUserUsernameUnique(t *testing.T) {
	db := newTestDB(t)
	db.Create(&AdminUser{Username: "dup", PasswordHash: "h1"})
	err := db.Create(&AdminUser{Username: "dup", PasswordHash: "h2"}).Error
	if err == nil {
		t.Fatal("expected unique constraint violation")
	}
}

func TestSysUserIsBanned(t *testing.T) {
	db := newTestDB(t)
	user := SysUser{UserName: "testuser", Password: "pw"}
	db.Create(&user)
	db.Model(&user).Update("is_banned", true)

	var result SysUser
	db.First(&result, user.ID)
	if !result.IsBanned {
		t.Fatal("expected is_banned=true")
	}
}
