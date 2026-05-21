# CLI Command Contract: mall-admin

**Phase 1 output** | **Date**: 2026-05-19 | **Branch**: `001-cli-admin-tool`

This document is the authoritative contract for all commands exposed by the
`mall-admin` CLI tool. Implementation MUST conform to this contract.

---

## Root Command

```
mall-admin [--db-path <path>] [--help] [--version] <command>
```

| Flag | Default | Description |
|------|---------|-------------|
| `--db-path` | `$MALL_DB_PATH` env var, then compiled-in default path | Path to the SQLite database file |
| `--help` | — | Print usage summary |
| `--version` | — | Print tool version |

**Exit codes**: 0 = success, 1 = error (not found, validation, DB failure)

---

## Command Group: `users`

### `users list`

```
mall-admin users list [--page N] [--all]
```

| Flag | Default | Description |
|------|---------|-------------|
| `--page` | 1 | Page number (20 rows per page) |
| `--all` | false | Show all pages at once |

**Output**: Table with columns: `ID | Username | Nickname | Phone | Email | Role | Created | Status`
- Phone is masked (138****5678)
- Status column shows `[DELETED]` when `deleted_at` is set, blank otherwise
- Always includes soft-deleted records (Unscoped query)

**Exit**: 0 on success. Prints "No records found." if table is empty.

---

### `users search`

```
mall-admin users search [--username <str>] [--phone <str>] [--nickname <str>] [--page N]
```

| Flag | Description |
|------|-------------|
| `--username` | Filter by username (LIKE %str%) |
| `--phone` | Filter by phone (LIKE %str%) |
| `--nickname` | Filter by nick_name (LIKE %str%) |
| `--page` | Page number (default 1) |

At least one filter flag MUST be provided; prints usage error if none given.

**Output**: Same table format as `users list`.

---

### `users show`

```
mall-admin users show <id>
```

**Output**: Key-value detail panel. All user fields except `password` and
`wx_session_key`. Phone shown in full (not masked — admin has explicit intent).

**Exit**: 1 + "User not found." if id not in DB (Unscoped).

---

### `users set-role`

```
mall-admin users set-role <id> <role>
```

**Arguments**:
- `<id>`: positive integer
- `<role>`: non-negative integer

**Interaction**:
```
Update role for user #42 from 0 → 1?
Type 'yes' to confirm: _
```

**Output on confirm**: "✓ User #42 role updated to 1."
**Output on abort**: "Aborted."
**Exit**: 1 if user not found or validation fails.

---

### `users delete`

```
mall-admin users delete <id>
```

**Interaction**:
```
Soft-delete user #42 (username: alice)?
Type 'yes' to confirm: _
```

**Output on confirm**: "✓ User #42 deleted."
**Output on abort**: "Aborted."
**Exit**: 1 if user not found.

---

## Command Group: `products`

### `products list`

```
mall-admin products list [--status N] [--user-id N] [--page N] [--all]
```

| Flag | Description |
|------|-------------|
| `--status` | Filter by status (0, 1, or 2) |
| `--user-id` | Filter by seller user id |
| `--page` | Page number (default 1) |
| `--all` | Show all pages at once |

**Output**: Table with columns: `ID | Title | Price | Status | Seller | Created | Deleted`
- Status column: `available`, `sold`, `removed`
- Deleted column: `[DELETED]` if soft-deleted, blank otherwise
- Always includes soft-deleted records (Unscoped)

**Exit**: 0 on success. Prints "No records found." if table is empty.

---

### `products show`

```
mall-admin products show <id>
```

**Output**: Key-value detail panel. All product fields including description,
images, contact_type, contact_value, buy_uid, and `Favorites: N` (count from
`user_favorite` table).

**Exit**: 1 + "Product not found." if id not in DB.

---

### `products set-status`

```
mall-admin products set-status <id> <status>
```

**Arguments**:
- `<id>`: positive integer
- `<status>`: 0, 1, or 2

**Interaction**:
```
Change product #10 status from available → removed?
Type 'yes' to confirm: _
```

**Output on confirm**: "✓ Product #10 status updated to removed."
**Output on abort**: "Aborted."
**Exit**: 1 if not found or invalid status.

---

### `products delete`

```
mall-admin products delete <id>
```

**Interaction**:
```
Soft-delete product #10 (title: "iPhone 14 pro")?
Type 'yes' to confirm: _
```

**Output on confirm**: "✓ Product #10 deleted."
**Output on abort**: "Aborted."
**Exit**: 1 if not found.

---

## Error Output Convention

All error messages go to **stderr**. Success output goes to **stdout**.
This allows piping (`mall-admin users list | grep alice`) without error noise.
