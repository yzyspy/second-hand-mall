# Feature Specification: CLI Admin Tool for Second-Hand Mall

**Feature Branch**: `001-cli-admin-tool`
**Created**: 2026-05-19
**Status**: Draft
**Input**: User description: "读取/Users/yangzhongyu/Desktop/code/github/second-hand-mall/mall-server/second-hand-mall.db，通过命令行交互，可以进行用户管理，查询，商品查询，管理"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - List and Search Users (Priority: P1)

A platform administrator opens the CLI tool and queries the user list.
They can filter users by username, phone number, or nickname, and view
full profile details for any individual user.

**Why this priority**: User management is the most critical admin function.
Being able to look up and audit user accounts underlies all moderation work.

**Independent Test**: Run the CLI tool with a `users list` command and
confirm a paginated table of users is printed. Run `users search` with a
known username and confirm matching records appear. Run `users show <id>`
and confirm all user profile fields are displayed.

**Acceptance Scenarios**:

1. **Given** the admin runs `users list`, **When** the command executes,
   **Then** a table showing id, username, nickname, phone, email, and
   registration date is printed to the terminal, paginated at 20 rows.
2. **Given** the admin runs `users search --username alice`, **When** the
   command executes, **Then** only users whose username contains "alice"
   are shown.
3. **Given** the admin runs `users show 42`, **When** the command executes,
   **Then** all fields for user 42 are displayed, with the phone number
   partially masked (e.g., `138****5678`).

---

### User Story 2 - Manage User Accounts (Priority: P2)

An administrator needs to update a user's role or soft-delete a user who
has violated platform rules. The CLI prompts for confirmation before any
destructive or role-changing operation.

**Why this priority**: Account management (role changes, account suspension)
is essential for platform governance but less urgent than read access.

**Independent Test**: Run `users set-role <id> <role>` and confirm the
role_id field is updated in the database. Run `users delete <id>` with
confirmation "yes" and confirm the deleted_at timestamp is set. Run
`users delete <id>` with confirmation "no" and confirm no change occurs.

**Acceptance Scenarios**:

1. **Given** the admin runs `users set-role 42 1`, **When** the admin
   confirms the prompt, **Then** user 42's role is updated and a success
   message is printed.
2. **Given** the admin runs `users delete 42`, **When** the admin types
   "yes" at the confirmation prompt, **Then** the user is soft-deleted and
   the action is logged to the terminal.
3. **Given** the admin runs `users delete 42`, **When** the admin types
   anything other than "yes", **Then** no change is made and the tool
   prints "Aborted."

---

### User Story 3 - List and Search Products (Priority: P1)

An administrator lists all products on the platform and filters them by
status (available, sold, removed) or by seller. They can view full product
details including contact information and buyer identity.

**Why this priority**: Product moderation is the other core admin function
alongside user management. Administrators must be able to audit all listings.

**Independent Test**: Run `products list` and confirm a table with product
id, title, price, status, seller id, and created date is printed. Run
`products list --status 0` and confirm only active listings appear. Run
`products show <id>` and confirm all product fields are displayed.

**Acceptance Scenarios**:

1. **Given** the admin runs `products list`, **When** the command executes,
   **Then** a paginated table of all products is printed with key fields.
2. **Given** the admin runs `products list --status 0`, **When** the
   command executes, **Then** only products with status 0 (available) are
   shown.
3. **Given** the admin runs `products list --user-id 5`, **When** the
   command executes, **Then** only products published by user 5 are shown.
4. **Given** the admin runs `products show 10`, **When** the command
   executes, **Then** all fields for product 10 are displayed including
   description, images, contact type, contact value, and buyer id.

---

### User Story 4 - Manage Product Listings (Priority: P2)

An administrator updates a product's status (e.g., force-removes a
prohibited listing) or soft-deletes a record. All mutations require
confirmation.

**Why this priority**: Moderating product listings is important for platform
health but secondary to the read-access stories.

**Independent Test**: Run `products set-status <id> 2` with confirmation
"yes" and confirm the product status field is updated. Run `products delete
<id>` with confirmation and confirm the deleted_at timestamp is set.

**Acceptance Scenarios**:

1. **Given** the admin runs `products set-status 10 2`, **When** the admin
   confirms, **Then** product 10's status is set to 2 and a success message
   is printed.
2. **Given** the admin runs `products delete 10` and confirms "yes",
   **Then** product 10's deleted_at timestamp is set.
3. **Given** a non-existent product id is supplied, **When** the command
   runs, **Then** the tool prints "Product not found." and exits with a
   non-zero code.

---

### Edge Cases

- What happens when the database file path is invalid or the file is
  missing? → Tool prints a clear error and exits immediately.
- What happens when a query returns zero results? → Tool prints
  "No records found." rather than an empty table.
- What happens when the admin supplies a non-integer id? → Tool prints
  a validation error without touching the database.
- What happens when the admin aborts a confirmation prompt (Ctrl+C)? →
  Tool exits cleanly with no changes made.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The tool MUST accept the database file path as a CLI argument
  or environment variable, defaulting to the project's database path.
- **FR-002**: The tool MUST provide a `users list` command displaying all
  non-deleted users in a table, paginated at 20 rows by default.
- **FR-003**: The tool MUST provide a `users search` command with optional
  filters `--username`, `--phone`, and `--nickname`.
- **FR-004**: The tool MUST provide a `users show <id>` command displaying
  all profile fields for one user, with phone middle digits masked.
- **FR-005**: The tool MUST provide a `users set-role <id> <role>` command
  updating role_id after explicit confirmation.
- **FR-006**: The tool MUST provide a `users delete <id>` command that
  soft-deletes the user (sets deleted_at) after explicit confirmation.
- **FR-007**: The tool MUST provide a `products list` command with optional
  filters `--status` and `--user-id`, paginated at 20 rows by default.
- **FR-008**: The tool MUST provide a `products show <id>` command
  displaying all fields for one product.
- **FR-009**: The tool MUST provide a `products set-status <id> <status>`
  command updating the product's status field after explicit confirmation.
- **FR-010**: The tool MUST provide a `products delete <id>` command that
  soft-deletes the product after explicit confirmation.
- **FR-011**: All mutation commands MUST prompt the admin to type "yes"
  before executing any database write.
- **FR-012**: The tool MUST exit with a non-zero status code on any error
  (invalid id, record not found, database unreachable).
- **FR-013**: Every command MUST display usage information when invoked
  with `--help`.

### Key Entities *(include if feature involves data)*

- **User** (`sys_user`): Platform account with profile, WeChat identity
  fields (openid, unionid, session key), role, and soft-delete state.
  Phone numbers are sensitive and MUST be partially masked in list output.
- **Product** (`product`): Marketplace listing with title, price, image
  URLs, location, status (0=available, 1=sold, 2=removed), contact info,
  seller user id, and buyer user id.
- **UserFavorite** (`user_favorite`): Read-only association between a user
  and a favorited product. Surfaced in `products show` for context only;
  not directly manageable via this tool.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: An administrator can locate any user or product by id or
  search filter within 30 seconds of launching the tool.
- **SC-002**: All mutation commands produce a clear success or failure
  message; zero silent failures occur.
- **SC-003**: Confirmation prompts prevent accidental changes; no
  unintended deletions or role changes happen during normal operation.
- **SC-004**: The tool launches and returns the first page of results
  within 2 seconds on the local machine with the current database.
- **SC-005**: Every command's usage is fully self-documented via `--help`;
  no external documentation is required to operate the tool.

## Assumptions

- The tool is for local, single-admin use on the same machine as the
  database; no network access or multi-user session management is needed.
- The database file is located at
  `/Users/yangzhongyu/Desktop/code/github/second-hand-mall/mall-server/second-hand-mall.db`
  by default; a CLI argument override is sufficient.
- Product status codes are: 0 = available, 1 = sold, 2 = force-removed.
- User role_id values are raw integers; the tool displays them as numbers
  since no roles table exists in the current schema.
- No authentication is required to launch the admin tool; physical access
  to the machine is considered sufficient authorization.
- All delete operations are soft-deletes (setting deleted_at). Hard
  deletion is out of scope for v1.
- The `user_favorite` table is read-only context; it is not manageable
  through this tool.
