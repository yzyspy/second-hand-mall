# Data Model: CLI Admin Tool for Second-Hand Mall

**Phase 1 output** | **Date**: 2026-05-19 | **Branch**: `001-cli-admin-tool`
**Source**: Existing SQLite schema in `second-hand-mall.db` — no migrations required.

---

## Entity: SysUser

**Table**: `sys_user`
**GORM model file**: `internal/models/user.go`

| Field | DB Column | Type | Notes |
|-------|-----------|------|-------|
| ID | id | uint (PK, autoincrement) | |
| CreatedAt | created_at | time.Time | |
| UpdatedAt | updated_at | time.Time | |
| DeletedAt | deleted_at | gorm.DeletedAt | Soft-delete; use `Unscoped()` to include deleted |
| Username | username | string (varchar 50) | |
| Password | password | string (varchar 100) | NEVER displayed; omit from all output |
| Phone | phone | string (varchar 20) | Masked in list views; full value shown in `show` |
| WxUserid | wx_userid | string (varchar 50) | |
| WxOpenid | wx_openid | string (varchar 50) | |
| Avatar | avatar | string (varchar 255) | |
| Sex | sex | string (varchar 20) | |
| Email | email | string (varchar 100) | |
| Remarks | remarks | string (varchar 255) | |
| RoleID | role_id | int | Raw integer; no roles table exists |
| WxSessionKey | wx_session_key | string (varchar 100) | Sensitive; NEVER displayed |
| WxUnionid | wx_unionid | string (varchar 100) | |
| NickName | nick_name | string (varchar 50) | |

### Display Rules

- **List view** (`users list`, `users search`): show id, username, nick_name,
  MaskPhone(phone), email, role_id, created_at.
- **Detail view** (`users show`): show all fields except `password` and
  `wx_session_key` (never displayed). Phone shown unmasked in detail view
  (admin has explicit intent to view a single record).
- **Soft-delete state**: When `deleted_at` is not NULL, the user is soft-deleted.
  List commands show all users (Unscoped) with a `[DELETED]` marker when
  `deleted_at` is set.

### State Transitions

```
active (deleted_at IS NULL)
  │
  └─ users delete <id> (confirm)
       │
       ▼
  soft-deleted (deleted_at IS NOT NULL)
```

Role changes (`users set-role`) do not affect soft-delete state.

---

## Entity: Product

**Table**: `product`
**GORM model file**: `internal/models/product.go`

| Field | DB Column | Type | Notes |
|-------|-----------|------|-------|
| ID | id | uint (PK, autoincrement) | |
| CreatedAt | created_at | time.Time | |
| UpdatedAt | updated_at | time.Time | |
| DeletedAt | deleted_at | gorm.DeletedAt | Soft-delete |
| Title | title | string (varchar 200) | |
| Description | description | string (text) | Shown only in detail view |
| Price | price | float64 (decimal 10,2) | |
| Images | images | string (varchar 1000) | Comma-separated or JSON URLs |
| Location | location | string (varchar 100) | |
| Status | status | int | 0=available, 1=sold, 2=force-removed |
| UserID | user_id | uint | Seller; foreign key to sys_user.id |
| BuyUID | buy_uid | uint | Buyer; 0 if unsold |
| ContactType | contact_type | string (varchar 10) | e.g., "phone", "wechat" |
| ContactValue | contact_value | string (varchar 100) | Contact string |

### Status Codes

| Value | Meaning | CLI Label |
|-------|---------|-----------|
| 0 | Available for purchase | `available` |
| 1 | Sold | `sold` |
| 2 | Force-removed by admin | `removed` |

### Display Rules

- **List view**: show id, title, price, status label, user_id, created_at.
  Include soft-deleted records (Unscoped) with a `[DELETED]` marker.
- **Detail view** (`products show`): show all fields including description,
  images, contact_type, contact_value, buy_uid.

### State Transitions

```
available (status=0)
  │
  ├─ products set-status <id> 1   → sold (status=1)
  │
  └─ products set-status <id> 2   → removed (status=2)
       │
       └─ products delete <id>    → soft-deleted (deleted_at set)
```

---

## Entity: UserFavorite

**Table**: `user_favorite`
**GORM model file**: `internal/models/product.go` (same file)

| Field | DB Column | Type | Notes |
|-------|-----------|------|-------|
| ID | id | uint (PK, autoincrement) | |
| UserID | user_id | uint | |
| ProductID | product_id | uint | |
| CreatedAt | created_at | time.Time | |

**Usage in admin tool**: Read-only. Favorite count is shown in
`products show <id>` output as `Favorites: N`.

---

## Validation Rules (enforced by CLI, not DB)

- `<id>` arguments MUST be positive integers; non-integer or negative values
  produce a validation error before any DB call.
- `<role>` in `users set-role` MUST be a non-negative integer.
- `<status>` in `products set-status` MUST be 0, 1, or 2.
- If a record matching `<id>` is not found (including Unscoped), the command
  prints "Not found." and exits with code 1.
