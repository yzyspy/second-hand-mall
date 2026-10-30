package main

import (
	"mall-server/internal/app/config"
	"mall-server/internal/app/models"
	"testing"
)

// Run this test in a scratch container to verify startup without OS timezone files.
func TestMySQLTimezoneWithoutSystemFiles(t *testing.T) {
	t.Setenv("MYSQL_PASSWORD", "")
	_, err := models.DatabaseDSN(&config.Config{
		Gorm:  config.Gorm{DBType: "mysql"},
		MySQL: config.MySQL{Host: "127.0.0.1", Port: 3306, User: "test", DBName: "test"},
	})
	if err != nil {
		t.Fatalf("database startup must work without OS timezone files: %v", err)
	}
}
