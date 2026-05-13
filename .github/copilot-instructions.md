# .github/copilot-instructions.md

Purpose: Short, machine-focused guidance for Copilot-style assistants to speed up local setup, testing, and code navigation in this repository.

## Quick commands

Backend (Go: mall-server)
- Build: cd mall-server && go build -o mall-server
- Run (CLI): ./mall-server web -config configs/config.yaml
- Run full tests: cd mall-server && go test ./...
- Run a single test: cd mall-server && go test -run TestName ./path/to/package
  - Example (existing): cd mall-server && go test -run TestDateDiff ./internal/app/models/...

Frontend (WeChat mini-program: mall-mini)
- Open `mall-mini/` in WeChat Developer Tools (TypeScript compilation done by the IDE). No manual build step required.

## High-level architecture (big picture)

- Root splits into two apps:
  1. mall-server/ — Go backend providing REST APIs, JWT auth, COS STS signature generation, and SQLite/GORM persistence.
  2. mall-mini/ — WeChat Mini Program (TypeScript) that talks to mall-server and uploads images directly to COS using STS.

- Backend components to review for flows:
  - main.go — CLI entry (uses urfave/cli).
  - internal/app/config — global config loader (multiconfig) exposed as config.C.
  - internal/app/models/init.go & servicecontext.go — DB (GORM) initialization and ServiceContext that holds DB.
  - internal/app/router/router.go & auth.go — Gin router and JWT AuthMiddleware.
  - internal/app/service/* — handlers (login, upload, etc.).
  - pkg/jwtx — JWT generation/parsing (24h expiry).

- Mini-program key files:
  - miniprogram/app.ts — globalData (token, baseUrl)
  - miniprogram/utils/request.ts — HTTP wrapper that auto-injects JWT and returns Promises
  - miniprogram/utils/cos-upload.ts — uses backend STS signature then uploads via wx.uploadFile

## Key conventions and patterns (repo-specific)

- Configuration: YAML files under mall-server/configs; loaded via multiconfig into a singleton config.C.
- CLI-first server: mall-server exposes CLI commands (e.g., `web -config configs/config.yaml`) — use the CLI binary rather than running a custom main wrapper.
- DB: SQLite + GORM. DB connection and migrations (if any) live under internal/app/models. ServiceContext holds the DB for handlers.
- Auth: JWT tokens via pkg/jwtx. Header: `Authorization: Bearer <token>`. Expiry is 24 hours by default.
- Uploads: Backend generates Tencent COS STS credentials (policy/signature) and frontend uploads directly with wx.uploadFile; see internal/app/service/upload.go and miniprogram/utils/cos-upload.ts.
- HTTP wrapper (mini): miniprogram/utils/request.ts automatically attaches JWT from local storage; prefer this wrapper for API calls.
- TypeScript path alias: mall-mini/tsconfig.json uses path alias @/* → miniprogram/*.
- CORS: router.CORSMiddleware controls allowed origins (defaults to allow http://localhost:5173).

## Files and places to read first (quick map)
- mall-server/main.go
- mall-server/configs/config.yaml (set credentials and DB path)
- mall-server/internal/app/router/auth.go
- mall-server/internal/app/models/init.go
- mall-server/pkg/jwtx/jwtx.go
- mall-server/internal/app/service/upload.go
- mall-mini/miniprogram/utils/request.ts
- mall-mini/miniprogram/utils/cos-upload.ts

## Existing assistant configs to incorporate
- CLAUDE.md (contains a concise architecture and run/test commands). Review it for additional run examples.

## Notes for Copilot sessions
- When making changes that touch auth or DB initialization, run unit tests in mall-server with `go test ./...` and target single packages with `-run` to iterate quickly.
- For frontend work, open mall-mini in WeChat Developer Tools and use its TypeScript compiler/watch.
- Secrets (COS keys, DB file paths) live in configs/config.yaml — do not commit secrets to the repo.

---

If you need additions (e.g., CI commands, linter setup, or common troubleshooting steps), say which area and examples to include.
