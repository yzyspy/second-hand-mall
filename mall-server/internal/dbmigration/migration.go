// Package dbmigration implements the small, offline SQLite-to-MySQL transfer.
package dbmigration

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"gorm.io/gorm"
	"mall-server/internal/app/dao"
	"math"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

type Table struct {
	Name string
	Rows []map[string]interface{}
}
type Data []Table

// Read includes soft-deleted records and rejects values MySQL would truncate.
// This tool intentionally loads all records: it is for this small offline database.
func Read(db *gorm.DB) (Data, error) {
	var result Data
	err := db.Transaction(func(tx *gorm.DB) error {
		readers := []func(*gorm.DB) (Table, error){readTable[dao.SysUser], readTable[dao.AdminUser], readTable[dao.Product], readTable[dao.UserFavorite], readTable[dao.Conversation], readTable[dao.Message]}
		for _, read := range readers {
			table, err := read(tx)
			if err != nil {
				return err
			}
			result = append(result, table)
		}
		tables, err := tx.Migrator().GetTables()
		if err != nil {
			return err
		}
		for _, name := range tables {
			known := strings.HasPrefix(name, "sqlite_")
			for _, table := range result {
				if name == table.Name {
					known = true
				}
			}
			if !known {
				return fmt.Errorf("unhandled source table %s; refusing incomplete migration", name)
			}
		}
		return checkReferences(result)
	})
	return result, err
}

func readTable[T any](db *gorm.DB) (Table, error) {
	var model T
	stmt := &gorm.Statement{DB: db}
	if err := stmt.Parse(&model); err != nil {
		return Table{}, err
	}
	table := Table{Name: stmt.Schema.Table}
	columns, err := db.Migrator().ColumnTypes(&model)
	if err != nil {
		return table, err
	}
	for _, col := range columns {
		if stmt.Schema.LookUpField(col.Name()) == nil {
			return table, fmt.Errorf("unhandled source column %s.%s", table.Name, col.Name())
		}
	}
	var items []T
	if err := db.Unscoped().Order("id ASC").Find(&items).Error; err != nil {
		return table, err
	}
	for _, item := range items {
		row := map[string]interface{}{}
		value := reflect.ValueOf(item)
		for _, field := range stmt.Schema.Fields {
			if field.DBName == "" {
				continue
			}
			v, _ := field.ValueOf(context.Background(), value)
			if valuer, ok := v.(driver.Valuer); ok {
				v, err = valuer.Value()
				if err != nil {
					return table, err
				}
			}
			if text, ok := v.(string); ok {
				max := field.Size
				typ := strings.ToLower(field.TagSettings["TYPE"])
				if strings.HasPrefix(typ, "varchar(") {
					size := strings.TrimSuffix(strings.TrimPrefix(typ, "varchar("), ")")
					max, _ = strconv.Atoi(size)
				}
				if max > 0 && utf8.RuneCountInString(text) > max {
					return table, fmt.Errorf("%s id=%v: %s exceeds %d characters", table.Name, row["id"], field.DBName, max)
				}
				if typ == "text" && len(text) > 65535 {
					return table, fmt.Errorf("%s.%s exceeds TEXT byte limit", table.Name, field.DBName)
				}
			}
			if t, ok := v.(time.Time); ok {
				if t.IsZero() || t.Year() < 1000 || t.Year() > 9999 {
					return table, fmt.Errorf("%s id=%v: invalid MySQL timestamp %s", table.Name, row["id"], field.DBName)
				}
				// DATETIME(6): keep instants, normalize timezone and declared precision.
				v = t.UTC().Truncate(time.Microsecond)
			}
			if field.DBName == "price" {
				f := v.(float64)
				if math.IsNaN(f) || math.IsInf(f, 0) || math.Abs(f) > 99999999.99 || math.Abs(f*100-math.Round(f*100)) > 0.00001 {
					return table, fmt.Errorf("product id=%v: price cannot fit decimal(10,2)", row["id"])
				}
			}
			if field.TagSettings["TYPE"] == "int" {
				rv := reflect.ValueOf(v)
				if rv.Kind() == reflect.Uint && rv.Uint() > math.MaxInt32 {
					return table, fmt.Errorf("%s.%s exceeds signed INT", table.Name, field.DBName)
				}
			}
			row[field.DBName] = v
		}
		table.Rows = append(table.Rows, row)
	}
	return table, nil
}

func checkReferences(data Data) error {
	ids := map[string]map[uint]bool{}
	for _, table := range data {
		ids[table.Name] = map[uint]bool{}
		for _, row := range table.Rows {
			ids[table.Name][row["id"].(uint)] = true
		}
	}
	links := map[string]map[string]string{
		"product":           {"user_id": "sys_user", "buy_uid": "sys_user"},
		"user_favorite":     {"user_id": "sys_user", "product_id": "product"},
		"chat_conversation": {"product_id": "product", "buyer_id": "sys_user", "seller_id": "sys_user"},
		"chat_message":      {"conversation_id": "chat_conversation", "sender_id": "sys_user"},
	}
	for _, table := range data {
		for _, row := range table.Rows {
			for field, target := range links[table.Name] {
				id := row[field].(uint)
				if field == "buy_uid" && id == 0 {
					continue
				}
				if !ids[target][id] {
					return fmt.Errorf("%s id=%v: %s=%d references missing %s", table.Name, row["id"], field, id, target)
				}
			}
		}
	}
	return nil
}

// Copy requires pre-created empty tables. DDL is deliberately outside this transaction.
// All inserts and full row verification succeed together, or no data is committed.
func Copy(dst *gorm.DB, data Data) error {
	return dst.Transaction(func(tx *gorm.DB) error {
		if tx.Dialector.Name() == "mysql" {
			for _, table := range data {
				var engine string
				if err := tx.Raw("SELECT ENGINE FROM information_schema.TABLES WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ?", table.Name).Scan(&engine).Error; err != nil {
					return err
				}
				if !strings.EqualFold(engine, "InnoDB") {
					return fmt.Errorf("target table %s must use InnoDB for rollback", table.Name)
				}
			}
		}
		for _, table := range data {
			var count int64
			if err := tx.Table(table.Name).Count(&count).Error; err != nil {
				return err
			}
			if count != 0 {
				return fmt.Errorf("target table %s is not empty", table.Name)
			}
		}
		for _, table := range data {
			if len(table.Rows) == 0 {
				continue
			}
			// GORM adds last-insert-id metadata to map inserts; keep source rows immutable.
			rows := make([]map[string]interface{}, len(table.Rows))
			for i, row := range table.Rows {
				rows[i] = map[string]interface{}{}
				for key, value := range row {
					rows[i][key] = value
				}
			}
			if err := tx.Table(table.Name).Session(&gorm.Session{SkipHooks: true}).CreateInBatches(rows, 100).Error; err != nil {
				return fmt.Errorf("copy %s: %w", table.Name, err)
			}
		}
		readers := []func(*gorm.DB) (Table, error){readTable[dao.SysUser], readTable[dao.AdminUser], readTable[dao.Product], readTable[dao.UserFavorite], readTable[dao.Conversation], readTable[dao.Message]}
		for i, read := range readers {
			got, err := read(tx)
			if err != nil {
				return err
			}
			expected, err := canonical(data[i])
			if err != nil {
				return err
			}
			actual, err := canonical(got)
			if err != nil {
				return err
			}
			if expected != actual {
				return fmt.Errorf("verification failed for %s; transaction rolled back", got.Name)
			}
		}
		return nil
	})
}
func canonical(table Table) (string, error) {
	// ORDER BY id is used for source and target; JSON sorts map keys.
	rows := append([]map[string]interface{}{}, table.Rows...)
	sort.Slice(rows, func(i, j int) bool { return rows[i]["id"].(uint) < rows[j]["id"].(uint) })
	b, err := json.Marshal(rows)
	return string(b), err
}
