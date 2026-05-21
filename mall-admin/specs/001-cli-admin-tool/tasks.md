---
description: "Task list for CLI Admin Tool for Second-Hand Mall"
---

# Tasks: CLI Admin Tool for Second-Hand Mall

**Input**: Design documents from `/specs/001-cli-admin-tool/`
**Prerequisites**: plan.md ✅, spec.md ✅, research.md ✅, data-model.md ✅, contracts/cli-commands.md ✅

**Tests**: Not requested in feature specification — no test tasks generated.

**Organization**: Tasks are grouped by user story to enable independent
implementation and testing of each story.

## Format: `[ID] [P?] [Story?] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1–US4)
- Exact file paths are included in every task description

## Path Conventions

Source lives under `mall-admin/` (repo root for this project):
- Models: `internal/models/`
- DB plumbing: `internal/db/`
- CLI commands: `cmd/`
- Entrypoint: `main.go`

---

## Phase 1: Setup

**Purpose**: Initialize the Go module and project skeleton.

- [x] T001 Initialize Go module: run `go mod init mall-admin` in `mall-admin/` directory
- [x] T002 Add all required dependencies to `go.mod`: `gorm.io/gorm`, `github.com/glebarez/sqlite`, `github.com/spf13/cobra` (using stdlib `text/tabwriter` instead of external tablewriter)
- [x] T003 [P] Create directory structure: `cmd/`, `internal/db/`, `internal/models/` under `mall-admin/`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that ALL user story commands depend on.

**⚠️ CRITICAL**: No user story work can begin until this phase is complete.

- [x] T004 Create `internal/db/db.go`: implement `Open(path string) (*gorm.DB, error)` using `github.com/glebarez/sqlite`; open with `gorm.Open(sqlite.Open(path), &gorm.Config{})`
- [x] T005 Create `cmd/root.go`: define Cobra root command with persistent `--db-path` flag; read `MALL_DB_PATH` env var as fallback; compiled-in default path `/Users/yangzhongyu/Desktop/code/github/second-hand-mall/mall-server/second-hand-mall.db`; call `db.Open` and store `*gorm.DB` in a package-level var accessible to subcommands; add `--version` flag printing `mall-admin v1.0.0`
- [x] T006 Create `main.go`: single call to `cmd.Execute()`; exit with code 1 on error
- [x] T007 [P] Create `internal/models/user.go`: define `SysUser` struct mapping to `sys_user` table with all columns (use `gorm.DeletedAt` for soft-delete); implement `MaskPhone(phone string) string` that replaces chars 4–7 with `*` (e.g., `138****5678`)
- [x] T008 [P] Create `internal/models/product.go`: define `Product` struct mapping to `product` table with all columns (use `gorm.DeletedAt`); define `UserFavorite` struct mapping to `user_favorite` table; add `StatusLabel(status int) string` helper returning `"available"`, `"sold"`, or `"removed"`
- [x] T009 [P] Create `cmd/validate.go`: implement `parseID(s string) (uint, error)` (positive integer, error if not); implement `parseStatus(s string) (int, error)` (must be 0, 1, or 2)
- [x] T010 [P] Create `cmd/confirm.go`: implement `mustConfirm(prompt string) bool` that prints the prompt, reads a line from stdin, returns true only if input is exactly `"yes"`

**Checkpoint**: Foundation ready — all four user story phases can now begin in parallel.

---

## Phase 3: User Story 1 — List and Search Users (Priority: P1) 🎯 MVP

**Goal**: Admin can list all users, search by username/phone/nickname, and
view full profile details for any user.

**Independent Test**: Run `./mall-admin users list` and confirm a table of
users prints. Run `./mall-admin users search --username <known>` and confirm
filtered rows. Run `./mall-admin users show 1` and confirm all profile fields.

### Implementation for User Story 1

- [x] T011 Create `cmd/users.go`: define `usersCmd` Cobra command group; register it on the root command; do NOT add subcommands yet (added in T012–T014)
- [x] T012 [P] [US1] Add `users list` subcommand to `cmd/users.go`: `--page` flag (default 1), `--all` flag; query `db.Unscoped().Offset().Limit(20).Find(&[]models.SysUser{})`, render tablewriter table with columns `ID | Username | Nickname | Phone | Email | Role | Created | Status`; phone via `MaskPhone()`; status column shows `[DELETED]` when `DeletedAt.Valid`, blank otherwise; print `"No records found."` if empty
- [x] T013 [P] [US1] Add `users search` subcommand to `cmd/users.go`: `--username`, `--phone`, `--nickname` flags; require at least one flag (print usage error if none given); build GORM Where chain with LIKE `%str%` for each provided flag; same Unscoped paginated table output as `users list`
- [x] T014 [US1] Add `users show` subcommand to `cmd/users.go`: takes one positional `<id>` arg; parse via `parseID()`; `db.Unscoped().First(&user, id)` — print `"User not found."` to stderr + exit 1 if not found; print all fields except `Password` and `WxSessionKey` as a key-value panel; phone shown unmasked (detail view)

**Checkpoint**: User Story 1 fully functional — admin can read user data independently of all other stories.

---

## Phase 4: User Story 3 — List and Search Products (Priority: P1)

**Goal**: Admin can list all products with status/seller filters and view
full product details including contact info and favorites count.

**Independent Test**: Run `./mall-admin products list` and confirm a product
table prints. Run `./mall-admin products list --status 0` and confirm only
available products appear. Run `./mall-admin products show <id>` and confirm
all fields plus `Favorites: N`.

**Note**: US3 is also P1 — this phase can be worked in parallel with Phase 3
by a second developer once Phase 2 (Foundational) is complete.

### Implementation for User Story 3

- [x] T015 Create `cmd/products.go`: define `productsCmd` Cobra command group; register it on the root command
- [x] T016 [P] [US3] Add `products list` subcommand to `cmd/products.go`: `--status`, `--user-id`, `--page`, `--all` flags; Unscoped query; apply status/user-id WHERE clauses when flags provided; render tablewriter table with columns `ID | Title | Price | Status | Seller | Created | Deleted`; status via `StatusLabel()`; `Deleted` column shows `[DELETED]` when soft-deleted
- [x] T017 [US3] Add `products show` subcommand to `cmd/products.go`: takes one positional `<id>` arg; parse via `parseID()`; `db.Unscoped().First(&product, id)` — print `"Product not found."` + exit 1 if not found; count `UserFavorite` rows for the product; print all fields as key-value panel including `Favorites: N`

**Checkpoint**: User Story 3 fully functional — admin can read all product data independently.

---

## Phase 5: User Story 2 — Manage User Accounts (Priority: P2)

**Goal**: Admin can update a user's role or soft-delete a user, both with
mandatory confirmation prompts.

**Independent Test**: Run `users set-role <existing-id> 1` and confirm prompt;
answer "yes" and confirm DB update. Run `users delete <existing-id>`, answer
"no" and confirm no DB change.

**Depends on**: Phase 3 (User Story 1) complete — set-role/delete reuse the
same user lookup pattern established in `users show`.

### Implementation for User Story 2

- [x] T018 [P] [US2] Add `users set-role` subcommand to `cmd/users.go`: takes two positional args `<id> <role>`; parse id via `parseID()`, role as non-negative int; fetch user (Unscoped) — exit 1 + message if not found; call `mustConfirm()` with prompt showing current role → new role; on confirm: `db.Model(&user).Update("role_id", newRole)`; print success message
- [x] T019 [US2] Add `users delete` subcommand to `cmd/users.go`: takes one positional `<id>`; parse via `parseID()`; fetch user — exit 1 if not found; call `mustConfirm()` with prompt showing username; on confirm: `db.Delete(&user)` (GORM soft-delete sets `deleted_at`); print success message

**Checkpoint**: User Stories 1 and 2 both work independently — full user management is complete.

---

## Phase 6: User Story 4 — Manage Product Listings (Priority: P2)

**Goal**: Admin can update a product's status or soft-delete it, both with
confirmation prompts.

**Independent Test**: Run `products set-status <id> 2` and confirm; answer
"yes" and verify status updated. Run `products delete <id>`, answer "no" and
confirm no change.

**Depends on**: Phase 4 (User Story 3) complete.

### Implementation for User Story 4

- [x] T020 [P] [US4] Add `products set-status` subcommand to `cmd/products.go`: takes two positional args `<id> <status>`; parse id via `parseID()`, status via `parseStatus()`; fetch product (Unscoped) — exit 1 if not found; call `mustConfirm()` with prompt showing current status label → new label; on confirm: `db.Model(&product).Update("status", newStatus)`; print success message
- [x] T021 [US4] Add `products delete` subcommand to `cmd/products.go`: takes one positional `<id>`; parse via `parseID()`; fetch product — exit 1 if not found; call `mustConfirm()` with prompt showing title; on confirm: `db.Delete(&product)`; print success message

**Checkpoint**: All four user stories fully functional and independently testable.

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: Ensure consistent output conventions, self-documenting help, and
end-to-end validation per quickstart.md.

- [x] T022 [P] Audit all subcommands in `cmd/users.go` and `cmd/products.go`: ensure every error path writes to `os.Stderr` (not `fmt.Println`) and calls `os.Exit(1)`; ensure all success output goes to `os.Stdout`
- [x] T023 [P] Add `Short` and `Long` descriptions to every Cobra command in `cmd/root.go`, `cmd/users.go`, `cmd/products.go` so `--help` is informative for each command and subcommand
- [x] T024 Run all five quickstart.md validation steps: `users list`, `products list`, `users show 1`, `users delete 9999` (expect "not found" + exit 1), `products set-status 1 2` with "no" answer (expect "Aborted." + no DB change)

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — start immediately
- **Foundational (Phase 2)**: Depends on Phase 1 — blocks all user stories
- **US1 (Phase 3)**: Depends on Phase 2 — no dependency on US3/US4
- **US3 (Phase 4)**: Depends on Phase 2 — no dependency on US1/US2/US4 (**can run in parallel with Phase 3**)
- **US2 (Phase 5)**: Depends on Phase 3 (US1) — reuses user model and lookup pattern
- **US4 (Phase 6)**: Depends on Phase 4 (US3) — reuses product model and lookup pattern
- **Polish (Phase 7)**: Depends on all story phases complete

### User Story Dependencies

- **US1 (P1)**: Independent after Phase 2 — no story dependencies
- **US3 (P1)**: Independent after Phase 2 — **can be worked in parallel with US1**
- **US2 (P2)**: Depends on US1 completion
- **US4 (P2)**: Depends on US3 completion

### Within Each Phase

- All `[P]`-marked tasks in the same phase can run concurrently
- Model tasks (T007, T008) before their respective command tasks
- `cmd/validate.go` and `cmd/confirm.go` (T009, T010) before mutation commands (T018–T021)

### Parallel Opportunities

```bash
# Phase 2: Foundational tasks — all [P] run together
Task: "Create internal/models/user.go with SysUser and MaskPhone"  # T007
Task: "Create internal/models/product.go with Product, UserFavorite, StatusLabel"  # T008
Task: "Create cmd/validate.go with parseID and parseStatus"  # T009
Task: "Create cmd/confirm.go with mustConfirm"  # T010

# Phase 3 + Phase 4 (run in parallel, both P1):
Developer A: Phase 3 (US1 — user read commands)
Developer B: Phase 4 (US3 — product read commands)

# Within Phase 3:
Task: "Add users list subcommand"  # T012
Task: "Add users search subcommand"  # T013
# (T012 and T013 can be done concurrently — same file but non-overlapping sections)
```

---

## Implementation Strategy

### MVP First (User Story 1 + 3 Only)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational (CRITICAL — blocks all stories)
3. Complete Phase 3: US1 (user list/search/show)
4. Complete Phase 4: US3 (product list/show)
5. **STOP and VALIDATE**: run `users list`, `products list`, `users show 1`
6. Ship MVP — administrators can audit users and products

### Incremental Delivery

1. Setup + Foundational → foundation ready
2. US1 + US3 (P1) → read-only admin tool (MVP)
3. US2 (P2) → user mutation commands
4. US4 (P2) → product mutation commands
5. Polish → fully hardened tool

### Parallel Team Strategy

With two developers:

1. Both complete Setup + Foundational together
2. Developer A: US1 (Phase 3) — user read commands
3. Developer B: US3 (Phase 4) — product read commands
4. Developer A: US2 (Phase 5) — user mutation commands
5. Developer B: US4 (Phase 6) — product mutation commands
6. Both: Polish (Phase 7)

---

## Notes

- `[P]` tasks = different files (or non-overlapping sections), no blocking dependencies
- `[Story]` label maps each task to a user story for traceability
- Each user story is independently completable and testable without the others
- No test tasks generated (not requested in spec)
- Commit after each phase or logical group
- Validation step (T024) is the acceptance gate before declaring the feature done
