# Code Standards

## General
- File naming: kebab-case
- Max file size: 200 lines
- Principles: YAGNI, KISS, DRY

## Go (server/)
- Standard project layout: cmd/, internal/, pkg/
- Use `internal/` for private packages
- Error handling: wrap with context using `fmt.Errorf("...: %w", err)`
- HTTP handlers: accept `http.ResponseWriter, *http.Request`
- Use chi/v5 for routing
- Use pgx/v5 for database access
- Linter: golangci-lint + gosec

## TypeScript (dashboard/)
- Strict mode, no `any`
- React Server Components by default, `"use client"` only when needed
- shadcn/ui + Tailwind for styling
- BFF pattern: API routes in `app/api/` for auth token management
- Linter: ESLint via `next lint`

## Rust (mcp-server/)
- Edition 2021
- `cargo clippy -- -D warnings` must pass
- No `unsafe` unless justified with comment
- Error handling: `thiserror` for library errors, `anyhow` for app errors
- Linter: clippy

## API Conventions
- REST, JSON request/response
- Versioned: `/api/v1/`
- Pagination: `?page=1&limit=20`
- Errors: `{"error": {"code": "...", "message": "..."}}`
- Timestamps: RFC3339 UTC

## Git
- Conventional commits: `feat:`, `fix:`, `refactor:`, `docs:`, `test:`, `chore:`
- One logical change per commit
- No secrets in commits
