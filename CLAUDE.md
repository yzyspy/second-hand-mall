# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a second-hand marketplace platform (二手交易平台) consisting of:
- **mall-server/**: Go backend service with JWT auth, RESTful APIs, and Tencent COS file upload
- **mall-mini/**: WeChat mini-program frontend (TypeScript)

## Build and Run Commands

### Backend (mall-server)

```bash
# Build the server
cd mall-server && go build -o mall-server

# Run the server (requires config file)
./mall-server web -config configs/config.yaml

# Run tests
cd mall-server && go test ./...

# Run a specific test
cd mall-server && go test -run TestDateDiff ./internal/app/models/...
```

### Mini-program (mall-mini)

Open the `mall-mini/` directory in WeChat Developer Tools. The project uses TypeScript compilation built into the IDE. No manual build commands required.

## Architecture

### Backend Structure (mall-server/)

```
mall-server/
├── main.go                    # Entry point, CLI commands (urfave/cli)
├── configs/config.yaml        # YAML configuration
├── internal/app/
│   ├── app.go                 # App initialization, config loading
│   ├── config/config.go       # Config structs, multiconfig loader → global singleton config.C
│   ├── dao/
│   │   ├── user.entity.go     # GORM entity definitions
│   │   └── user.repo.go       # Repository functions
│   ├── models/
│   │   ├── init.go            # Database connection (GORM/SQLite)
│   │   └── servicecontext.go  # ServiceContext holding DB
│   ├── router/
│   │   ├── router.go          # Gin routes, CORS middleware
│   │   └── auth.go            # JWT authentication middleware
│   ├── service/
│   │   ├── login.go           # User login/registration handlers
│   │   ├── upload.go          # Tencent COS STS signature generation
│   │   └── types.go           # Request/response DTOs
│   └── gormx/gormx.go         # GORM database wrapper
└── pkg/
    ├── jwtx/jwtx.go           # JWT token generation/parsing (24h expiry)
    └── logger/logger.go       # Logrus wrapper with context support
```

### Mini-program Structure (mall-mini/)

```
mall-mini/
├── project.config.json        # WeChat project config (appid, TypeScript enabled)
├── tsconfig.json              # TypeScript config, path alias @/* → miniprogram/*
└── miniprogram/
    ├── app.ts                 # App entry, globalData: { token, baseUrl }
    ├── app.json               # Pages, tabBar config
    ├── app.wxss               # Global styles
    ├── pages/
    │   ├── home/              # Product listing (mock data, pull-to-refresh)
    │   ├── publish/           # Publish product, COS image upload
    │   └── my/                # User profile, login/logout
    └── utils/
        ├── request.ts         # HTTP wrapper: auto-inject JWT, Promise-based
        └── cos-upload.ts      # COS upload using STS temp credentials
```

## Key Patterns

### Backend Patterns

- **Configuration**: YAML config loaded via `multiconfig` into `config.C` global singleton
- **Database**: SQLite with GORM, connection in `ServiceContext`
- **Routing**: Gin framework with route groups; public routes vs authenticated (via `AuthMiddleware`)
- **Auth**: JWT tokens via `pkg/jwtx`, 24-hour expiry, header: `Authorization: Bearer <token>`
- **File Upload**: Backend generates COS STS temp credentials with policy/signature; frontend uploads directly

### Mini-program Patterns

- **HTTP Requests**: Use `utils/request.ts` - wraps `wx.request` with Promise, auto-injects JWT token from storage
- **Image Upload**: Use `utils/cos-upload.ts` - gets STS signature from backend, uploads via `wx.uploadFile`
- **Page Data**: TypeScript interface for page data state, mock data in pages (API integration pending)
- **Styling**: WXSS with rpx units, global styles in `app.wxss`

## API Endpoints

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| POST | /user/save | No | Create user |
| POST | /user/login | No | Login, returns JWT |
| POST | /api/upload/cos-signature-v2 | No | Get COS STS signature for upload |
| GET | /ping | Yes | Health check (JWT required) |
| GET | /actuator/health/readiness | No | K8s readiness probe |
| GET | /actuator/health/liveness | No | K8s liveness probe |

## CORS

Server allows `http://localhost:5173` (Vue dev server). Update `CORSMiddleware` in `router.go` for production. Mini-program uses `localhost:8080` as API base URL (see `app.ts` and `request.ts`).

## Dependencies

### Backend
- **Gin** - HTTP framework
- **GORM** - ORM with SQLite driver
- **urfave/cli/v2** - CLI framework
- **logrus** - Structured logging
- **jwt/v5** - JWT handling
- **Tencent COS STS SDK** - Cloud storage temp credentials

### Mini-program
- TypeScript (compiled by WeChat Developer Tools)
- WeChat Mini Program APIs (wx.request, wx.uploadFile, wx.chooseMedia)
