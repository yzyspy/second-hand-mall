package dao

import (
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func setupFavoriteTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&UserFavorite{}, &Product{}, &SysUser{}); err != nil {
		t.Fatal(err)
	}
	return db
}

func TestToggleFavorite_AddThenRemove(t *testing.T) {
	db := setupFavoriteTestDB(t)

	// 第一次：添加收藏
	isFav, err := ToggleFavorite(db, 1, 10)
	assert.NoError(t, err)
	assert.True(t, isFav)

	// 第二次：取消收藏
	isFav, err = ToggleFavorite(db, 1, 10)
	assert.NoError(t, err)
	assert.False(t, isFav)
}

func TestIsFavorited(t *testing.T) {
	db := setupFavoriteTestDB(t)

	isFav, err := IsFavorited(db, 1, 10)
	assert.NoError(t, err)
	assert.False(t, isFav)

	_, _ = ToggleFavorite(db, 1, 10)

	isFav, err = IsFavorited(db, 1, 10)
	assert.NoError(t, err)
	assert.True(t, isFav)
}

func TestGetFavoriteList_OnlyInSale(t *testing.T) {
	db := setupFavoriteTestDB(t)

	// 插入一个在售商品（status=0）和一个已售商品（status=1）
	p1 := Product{Title: "在售商品", Price: 100, Status: 0, UserId: 99}
	p2 := Product{Title: "已售商品", Price: 200, Status: 1, UserId: 99}
	db.Create(&p1)
	db.Create(&p2)

	// 都收藏
	db.Create(&UserFavorite{UserID: 1, ProductID: p1.ID})
	db.Create(&UserFavorite{UserID: 1, ProductID: p2.ID})

	results, total, err := GetFavoriteList(db, 1, 1, 10)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, results, 1)
	assert.Equal(t, "在售商品", results[0].Title)
}
