# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a second-hand marketplace backend service (二手交易平台) built with Go. It provides user authentication, COS file upload signatures, and RESTful APIs for a mobile/web frontend.

## Build and Run Commands

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

## Architecture

```
mall-server/
├── main.go              # Entry point, CLI command definitions (urfave/cli)
├── lib.go               # Blank imports to pin dependencies in go.mod
├── configs/config.yaml  # Configuration file (YAML)
├── internal/app/
│   ├── app.go           # App initialization (config loading)
│   ├── logger.go        # Logger initialization with file rotation
│   ├── config/config.go # Configuration structs and loader (multiconfig)
│   ├── dao/             # Data Access Layer
│   │   ├── user.entity.go  # GORM entity definitions
│   │   └── user.repo.go    # Repository functions
│   ├── models/
│   │   ├── init.go           # Database connection setup (GORM/SQLite)
│   │   └── servicecontext.go # ServiceContext holding DB and config
│   ├── router/
│   │   ├── router.go     # Gin routes and CORS middleware
│   │   └── auth.go       # JWT authentication middleware
│   ├── service/
│   │   ├── login.go      # User login and registration handlers
│   │   ├── upload.go     # Tencent COS upload signature generation
│   │   └── types.go     # Request/response DTOs
│   └── gormx/
│       └── gormx.go      # GORM database wrapper
└── pkg/
    ├── jwtx/jwtx.go      # JWT token generation and parsing
    └── logger/logger.go  # Logrus wrapper with context support
```

## Key Patterns

- **Configuration**: YAML config loaded via `multiconfig` into `config.C` global singleton
- **Database**: SQLite with GORM, connection managed in `ServiceContext`
- **Routing**: Gin framework with route groups; public routes vs authenticated routes (via `AuthMiddleware`)
- **Auth**: JWT tokens generated/parsed by `pkg/jwtx`, 24-hour expiry, passed as `Authorization: Bearer <token>`

## Dependencies

- **Gin** (github.com/gin-gonic/gin) - HTTP web framework
- **GORM** (gorm.io/gorm) - ORM with SQLite driver
- **urfave/cli/v2** - CLI framework for subcommands
- **logrus** - Structured logging with file rotation
- **jwt/v5** - JWT token handling
- **Tencent COS STS SDK** - Cloud object storage temporary credentials

## API Endpoints

| Method | Path | Auth Required | Description |
|--------|------|---------------|-------------|
| POST | /user/save | No | Create user |
| POST | /user/login | No | Login, returns JWT |
| POST | /api/upload/cos-signature | No | Get COS upload signature (fixed key) |
| POST | /api/upload/cos-signature-v2 | No | Get COS upload signature (STS temp key) |
| GET | /ping | Yes | Health check (JWT required) |
| GET | /actuator/health/readiness | No | K8s readiness probe |
| GET | /actuator/health/liveness | No | K8s liveness probe |

## CORS

Server allows requests from `http://localhost:5173` (frontend dev server). Update `CORSMiddleware` in `router.go` for production.
