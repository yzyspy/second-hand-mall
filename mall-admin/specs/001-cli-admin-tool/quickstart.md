# Quickstart: mall-admin CLI Tool

**Phase 1 output** | **Date**: 2026-05-19 | **Branch**: `001-cli-admin-tool`

## Prerequisites

- Go 1.22 or later (`go version`)
- Access to the SQLite database file

No C compiler required — the tool uses a pure-Go SQLite driver.

---

## Build

```bash
cd /Users/yangzhongyu/Desktop/code/github/second-hand-mall/mall-admin
go build -o mall-admin .
```

Or to build without CGO explicitly (confirms the pure-Go driver is working):

```bash
CGO_ENABLED=0 go build -o mall-admin .
```

---

## Configure Database Path

The tool looks for the database in this order:

1. `--db-path` flag (highest priority)
2. `MALL_DB_PATH` environment variable
3. Compiled-in default:
   `/Users/yangzhongyu/Desktop/code/github/second-hand-mall/mall-server/second-hand-mall.db`

Set the environment variable to avoid typing the path each time:

```bash
export MALL_DB_PATH=/Users/yangzhongyu/Desktop/code/github/second-hand-mall/mall-server/second-hand-mall.db
```

---

## First Run Validation

```bash
# Confirm the tool connects and lists users
./mall-admin users list

# Expected output: table with user rows, or "No records found."
# If the DB path is wrong, you'll see: "Error: failed to open database: ..."
```

---

## Common Operations

### User Management

```bash
# List all users (paginated, 20 per page)
./mall-admin users list

# Page 2
./mall-admin users list --page 2

# Search by username
./mall-admin users search --username admin

# Search by phone
./mall-admin users search --phone 138

# View full user profile
./mall-admin users show 1

# Change user role
./mall-admin users set-role 5 1
# → prompts for confirmation

# Soft-delete a user
./mall-admin users delete 5
# → prompts for confirmation
```

### Product Management

```bash
# List all products
./mall-admin products list

# Filter by status (0=available, 1=sold, 2=removed)
./mall-admin products list --status 0

# Filter by seller
./mall-admin products list --user-id 3

# View full product details
./mall-admin products show 10

# Force-remove a listing
./mall-admin products set-status 10 2
# → prompts: "Change product #10 status from available → removed?"

# Soft-delete a product
./mall-admin products delete 10
# → prompts for confirmation
```

---

## Validation

To verify the tool is working correctly end-to-end:

1. Run `./mall-admin users list` — confirms DB connection and user table access.
2. Run `./mall-admin products list` — confirms product table access.
3. Run `./mall-admin users show 1` — confirms single-record detail view.
4. Run `./mall-admin users delete 9999` — confirm "User not found." error and exit code 1 (`echo $?`).
5. Run `./mall-admin products set-status 1 2` and type "no" at the prompt — confirm "Aborted." and no DB change.
