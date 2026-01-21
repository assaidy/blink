# AGENTS.md - Blink Realtime Chat Application

This document provides guidelines for AI agents working on the Blink codebase.

## Project Overview

Blink is a realtime chat application built with Go (1.25+), featuring:
- **Web Framework**: Fiber v2
- **Database**: PostgreSQL with sqlc for type-safe SQL generation
- **Cache/PubSub**: Valkey (Redis fork)
- **Migrations**: goose
- **Realtime**: WebSockets
- **Logging**: charmbracelet/log with slog backend

## Build Commands

```bash
# Build binary (generates sqlc code first)
make build
./bin/app

# Run locally with auto-regeneration
make run

# Watch mode (requires watchexec)
make watch

# Generate sqlc code only
make sqlc

# Clean build artifacts
make clean
```

## Database Commands

```bash
# Start infrastructure (Postgres, Valkey)
make comp-up

# Stop infrastructure
make comp-down

# Run migrations
make goose-up

# Rollback migration
make goose-down

# Reset database
make goose-reset

# Create new migration
make goose-migration name=migration_name
```

## Testing

```bash
# Run all tests
go test ./...

# Run specific test file
go test ./app/services/...

# Run specific test function
go test -v -run TestFunctionName ./...

# Run tests with coverage
go test -cover ./...

# Run tests for specific package with timeout
go test -timeout 30s ./app/handlers/...
```

## Code Style Guidelines

### Receiver Naming
Use `me` as the receiver name for all methods:
```go
func (me *AuthService) Register(params RegisterParams) error
```

### Imports
Group imports in the following order with blank lines between groups:
1. Standard library
2. Internal packages (`github.com/assaidy/blink/...`)
3. Third-party packages

```go
import (
	"context"
	"database/sql"
	"fmt"

	"github.com/assaidy/blink/app/repo"
	"github.com/assaidy/blink/app/utils"
	"github.com/gofiber/fiber/v2"
)
```

### Validation
Use `github.com/go-ozzo/ozzo-validation/v4` for input validation:
```go
func (me *CreateClientParams) cleanAndValidate() error {
	me.Platform = strings.TrimSpace(me.Platform)
	me.Os = strings.TrimSpace(me.Os)
	me.App = strings.TrimSpace(me.App)

	return validation.ValidateStruct(me,
		validation.Field(&me.Platform, validation.Required),
		validation.Field(&me.Os, validation.Required),
		validation.Field(&me.App, validation.Required),
	)
}
```

### Error Handling
Use the custom error utilities in `app/utils/errors.go`:

```go
// Define error kinds as package-level variables
var NotFound ErrorKind = "NotFound"

// Create structured errors
return "", utils.NewError(utils.InvalidData, err)
return nil, utils.NewError(utils.EmailNotFound, nil)

// Error kinds used in this project:
InvalidJson, InvalidData, NotFound, UsernameConflict, EmailConflict,
InternalFailure, ClientNotFound, EmailNotFound, InvalidOtp,
Unauthorized, InvalidCursor, InvalidEndpoint, WebscoketUpgradeRequired
```

### Context Usage
Use `context.Background()` in service methods when no context is provided from handlers. Always pass context as the first parameter:
```go
ctx := context.Background()
```

### Struct Tags
Use JSON tags for all public structs that are serialized:
```go
type CreateClientRequest struct {
	Platform string `json:"platform"`
	Os       string `json:"os"`
	App      string `json:"app"`
}
```

### ID Generation
Use ULID for ID generation:
```go
import "github.com/oklog/ulid/v2"

id := ulid.Make().String()
```

### Transaction Pattern
When using transactions, always defer rollback and commit explicitly:
```go
tx, err := me.db.BeginTx(ctx, nil)
if err != nil {
	return fmt.Errorf("failed to begin tx: %w", err)
}
defer tx.Rollback()
qtx := me.queries.WithTx(tx)
// ... operations ...
if err := tx.Commit(); err != nil {
	return fmt.Errorf("failed to commit tx: %w", err)
}
```

### Logging
Use the logger from the App struct, or create one via charmbracelet/log:
```go
logger := slog.New(log.NewWithOptions(os.Stderr, log.Options{ReportTimestamp: true}))
me.logger.Info("message", "key", value)
```

## Architecture

### Directory Structure
- `cmd/`: Entry points (app/, workers/)
- `app/handlers/`: HTTP controllers and WebSocket handlers
- `app/services/`: Business logic layer
- `app/repo/`: Data access layer (sqlc-generated)
- `app/db/`: SQL migrations and queries
- `app/utils/`: Shared utilities

### Dependency Injection
Inject dependencies via constructor functions:
```go
func NewAuthService(db *sql.DB) *AuthService {
	return &AuthService{
		db:      db,
		queries: repo.New(db),
	}
}
```

### Handler Pattern
1. Parse request body
2. Validate input
3. Call service method
4. Return appropriate HTTP response

## SQL and sqlc

Place SQL queries in `app/db/queries/` and migrations in `app/db/migrations/`. After modifying:
```bash
make sqlc  # Regenerates app/repo/
```

## WebSocket Handling

Use `github.com/gofiber/contrib/websocket`. WebSocket handlers should be wrapped with `WithWebsocket` middleware for authentication.

## Configuration

Environment variables are loaded via `godotenv` in `app/env` package. Never hardcode secrets.
