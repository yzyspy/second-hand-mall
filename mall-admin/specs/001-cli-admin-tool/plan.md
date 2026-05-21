# Implementation Plan: CLI Admin Tool for Second-Hand Mall

**Branch**: `001-cli-admin-tool` | **Date**: 2026-05-19 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/001-cli-admin-tool/spec.md`

## Summary

Build a standalone Go CLI tool (`mall-admin`) that connects directly to the
existing SQLite database and provides interactive commands for user management
(list, search, show, set-role, delete) and product management (list, show,
set-status, delete). All mutations require confirmation. The tool reads
existing `sys_user`, `product`, and `user_favorite` tables via GORM with
the `github.com/glebarez/sqlite` pure-Go driver.

## Technical Context

**Language/Version**: Go 1.22+
**Primary Dependencies**: GORM v2 (`gorm.io/gorm`), `github.com/glebarez/sqlite`,
  Cobra (`github.com/spf13/cobra`), tablewriter (`github.com/olekuznetsov/tablewriter`)
**Storage**: SQLite — existing `second-hand-mall.db` (read + write, no migrations)
**Testing**: `go test ./...` (unit tests for masking/pagination helpers; no DB
  integration tests required by spec)
**Target Platform**: macOS / Linux (local developer machine)
**Project Type**: CLI tool
**Performance Goals**: First result page within 2 seconds on local machine
**Constraints**: Pure-Go SQLite driver (no CGO); no network access required;
  single-user, no concurrent admin sessions
**Scale/Scope**: ~100–10,000 DB rows; single operator; 4 command groups

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Admin-First Design | ✅ PASS | Tool is exclusively for platform administrators |
| II. API-Driven Architecture | ⚠️ JUSTIFIED VIOLATION | CLI reads SQLite directly — see Complexity Tracking |
| III. Component-Based UI | N/A | CLI tool; no UI components |
| IV. Security and Access Control | ✅ PASS | Phone masking, confirmation prompts, non-zero exit on error |
| V. Simplicity and YAGNI | ✅ PASS | Cobra + GORM + tablewriter only; no extra abstractions |

**Post-design re-check**: All applicable principles satisfied. Violation II
is justified and documented in Complexity Tracking.

## Project Structure

### Documentation (this feature)

```text
specs/001-cli-admin-tool/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/           # Phase 1 output
│   └── cli-commands.md
└── tasks.md             # Phase 2 output (/speckit.tasks)
```

### Source Code (repository root)

```text
mall-admin/
├── main.go              # Entrypoint — calls cmd.Execute()
├── go.mod
├── go.sum
├── cmd/
│   ├── root.go          # Cobra root: --db-path flag, DB init, version
│   ├── users.go         # users subcommands (list, search, show, set-role, delete)
│   └── products.go      # products subcommands (list, show, set-status, delete)
└── internal/
    ├── db/
    │   └── db.go        # GORM open + singleton accessor
    └── models/
        ├── user.go      # SysUser GORM model + MaskPhone helper
        └── product.go   # Product + UserFavorite GORM models
```

**Structure Decision**: Single CLI project under `mall-admin/`. Source is
organized into `cmd/` for Cobra commands and `internal/` for reusable models
and DB plumbing. No separate `tests/` directory at root — tests live alongside
packages following Go conventions (`_test.go` files).

## Complexity Tracking

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| II. Direct SQLite access (bypasses mall-server API) | Admin operations (role change, force-remove, bulk query) are not exposed by the existing mall-server API | Adding admin endpoints to mall-server requires auth redesign, new endpoint spec, and backend deployment — out of scope for a local admin tool |
