# Research: CLI Admin Tool for Second-Hand Mall

**Phase 0 output** | **Date**: 2026-05-19 | **Branch**: `001-cli-admin-tool`

## 1. Pure-Go SQLite with GORM

**Decision**: Use `github.com/glebarez/sqlite` as the GORM SQLite driver.

**Rationale**: The standard GORM SQLite driver (`gorm.io/driver/sqlite`) wraps
`mattn/go-sqlite3` which requires CGO and a C compiler. `glebarez/sqlite` is
a 100% pure-Go replacement (backed by `modernc.org/sqlite`), enabling builds
with `CGO_ENABLED=0` and simpler cross-compilation.

**Usage**:
```go
import (
    "github.com/glebarez/sqlite"
    "gorm.io/gorm"
)

db, err := gorm.Open(sqlite.Open("path/to/file.db"), &gorm.Config{})
```

**Alternatives considered**:
- `mattn/go-sqlite3` — requires CGO; eliminated for build simplicity
- `zombiezen/go-sqlite` — lower-level, no GORM integration

---

## 2. GORM Soft Delete

**Decision**: Use GORM's built-in `gorm.DeletedAt` field type for soft-delete
awareness. Mark models with `DeletedAt gorm.DeletedAt` (not `time.Time`).

**Rationale**: GORM automatically appends `WHERE deleted_at IS NULL` to all
queries when the model has a `gorm.DeletedAt` field. The admin tool needs to
see ALL records (including soft-deleted ones) for auditing, so queries use
`db.Unscoped()` when showing all rows.

**Pagination pattern**:
```go
db.Unscoped().Offset((page-1)*pageSize).Limit(pageSize).Find(&users)
```

**Alternatives considered**:
- Raw SQL queries — more verbose, forfeits GORM convenience and model safety

---

## 3. Cobra CLI Structure

**Decision**: Use `github.com/spf13/cobra` with subcommand groups `users`
and `products`. Root command accepts `--db-path` persistent flag.

**Pattern**:
```
mall-admin [--db-path <path>] <group> <subcommand> [flags]
```

Examples:
```
mall-admin users list [--page N]
mall-admin users search --username alice
mall-admin users show 42
mall-admin users set-role 42 1
mall-admin users delete 42
mall-admin products list [--status N] [--user-id N] [--page N]
mall-admin products show 10
mall-admin products set-status 10 2
mall-admin products delete 10
```

**Confirmation pattern** (for mutations):
```go
fmt.Print("Type 'yes' to confirm: ")
var answer string
fmt.Scanln(&answer)
if answer != "yes" {
    fmt.Println("Aborted.")
    return
}
```

**Alternatives considered**:
- `urfave/cli/v2` — already used in mall-server; Cobra preferred here for
  richer subcommand nesting and auto-generated `--help`

---

## 4. Table Output

**Decision**: Use `github.com/olekuznetsov/tablewriter` (or equivalent
`tablewriter`) to render aligned terminal tables.

**Rationale**: Plain `fmt.Printf` with manual spacing is brittle; tablewriter
handles column alignment, borders, and wrapping automatically.

**Alternatives considered**:
- `pterm` — heavier dependency with colors/spinners that are overkill here
- Manual `text/tabwriter` — stdlib, but limited to tab-aligned output

---

## 5. Environment Variable Override

**Decision**: DB path defaults to `MALL_DB_PATH` env var, then falls back to
`/Users/yangzhongyu/Desktop/code/github/second-hand-mall/mall-server/second-hand-mall.db`
if the env var is unset. The `--db-path` CLI flag takes highest precedence.

**Priority**: `--db-path` flag > `MALL_DB_PATH` env var > compiled-in default path

---

## 6. Phone Masking Algorithm

**Decision**: Mask characters 4–7 (zero-indexed) of a phone string with `*`.

```go
func MaskPhone(phone string) string {
    if len(phone) < 8 {
        return phone
    }
    runes := []rune(phone)
    for i := 3; i < 7 && i < len(runes); i++ {
        runes[i] = '*'
    }
    return string(runes)
}
// "13812345678" → "138****5678"
```

**Rationale**: Chinese mobile numbers are 11 digits; masking the middle 4
digits balances privacy and administrator usability.

---

## All NEEDS CLARIFICATION resolved

No open questions remain. Proceeding to Phase 1.
