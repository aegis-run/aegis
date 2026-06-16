set shell := ["bash", "-euo", "pipefail", "-c"]
set dotenv-load

[private]
default:
    @just --list --unsorted

# ------------------------------------------------------------------ #
# Infrastructure                                                     #
# ------------------------------------------------------------------ #

# Start Postgres and Redis in the background
[group("infrastructure")]
up:
    docker compose up -d

# Stop and remove the database/cache containers
[group("infrastructure")]
down:
    docker compose down

# Destroy all local data (useful if you need a clean slate)
[group("infrastructure")]
nuke:
    docker compose down -v

# ------------------------------------------------------------------ #
# Development                                                          #
# ------------------------------------------------------------------ #

# Run the server
[group("development")]
run: up
    go run ./cmd/aegis serve | hl -P

# Format all code
[group("development")]
fmt:
    go tool goimports -w .
    gofmt -w .

# Check formatting without modifying files
[group("development")]
fmt-check:
    @test -z "$(gofmt -l .)" || (echo "Code is not formatted. Run 'just fmt'" && exit 1)

# Run linter
[group("development")]
lint:
    golangci-lint run ./...

# Run all checks (fmt + lint)
[group("development")]
check: fmt-check lint

# Run tests
[group("development")]
test *args:
    go test ./... {{ args }}

# Run tests with race detector
[group("development")]
test-race *args:
    go test -race ./... {{ args }}

# Generate all code (OpenAPI, SQL, etc)
[group("development")]
gen:
    go generate ./...
    rm pkg/db/postgres/delete_me.go

# ------------------------------------------------------------------ #
# Database                                                           #
# ------------------------------------------------------------------ #

db_dir := "pkg/db/postgres/migrations"
db_url := env("AEGIS_DB_URI", "postgres://user:password@localhost:5432/aegis?sslmode=disable")
db_schema := env("AEGIS_DB_SCHEMA_NAME", "public")
db_table := env("AEGIS_DB_MIGRATIONS_TABLE", "aegis_schema_migrations")

# Create a new migration file (Usage: just migrate create_users_table)
[group("database")]
migrate name:
    go tool goose -dir {{ db_dir }} create {{ name }} sql

# Check migration status internally securely navigating the namespace
[group("database")]
db-status:
    go tool goose -table {{ db_table }} -dir {{ db_dir }} postgres "{{ db_url }}&search_path={{ db_schema }}" status

# Rollback the last migration step cleanly
[group("database")]
db-down:
    go tool goose -table {{ db_table }} -dir {{ db_dir }} postgres "{{ db_url }}&search_path={{ db_schema }}" down

# Reset the isolated database components completely
[group("database")]
db-reset:
    go tool goose -table {{ db_table }} -dir {{ db_dir }} postgres "{{ db_url }}&search_path={{ db_schema }}" down-to 0

# ------------------------------------------------------------------ #
# Security                                                             #
# ------------------------------------------------------------------ #

# Run vulnerability check
[group("security")]
vuln:
    go tool govulncheck ./...

# Run static application security testing
[group("security")]
sast:
    go tool gosec -exclude-generated ./...

# Run all security checks
[group("security")]
security: vuln sast

# ------------------------------------------------------------------ #
# Build                                                                #
# ------------------------------------------------------------------ #

# Debug build
[group("build")]
build:
    go build -o bin/ ./...

# Install globally
[group("build")]
install:
    go install ./...

# ------------------------------------------------------------------ #
# Release                                                              #
# ------------------------------------------------------------------ #

# Tag and push — run after bump PR is merged
[group("release")]
release version:
    git tag -a "v{{ version }}" -m "Release v{{ version }}"
    git push origin "v{{ version }}"
    @echo "✓ Tagged and pushed v{{ version }}"

# ------------------------------------------------------------------ #
# Dependency management                                                #
# ------------------------------------------------------------------ #

# Tidy go.mod
[group("dependency management")]
mod-tidy:
    go mod tidy

# Verify dependencies
[group("dependency management")]
mod-verify:
    go mod verify

# Download all dependencies
[group("dependency management")]
mod-download:
    go mod download

# ------------------------------------------------------------------ #
# CI                                                                   #
# ------------------------------------------------------------------ #

# Run the full CI pipeline locally
[group("ci")]
ci: check test-race security
    @echo "✓ CI pipeline passed"

# ------------------------------------------------------------------ #
# Housekeeping                                                         #
# ------------------------------------------------------------------ #

# Remove build artifacts
[group("misc")]
clean:
    go clean ./...
    rm -rf bin/ coverage.out result

# Remove go cache
[group("misc")]
clean-cache:
    go clean -cache -modcache

# Remove everything
[group("misc")]
clean-all: clean clean-cache

# Print tool versions
[group("misc")]
versions:
    @echo "go:            $(go version)"
    @echo "gopls:         $(go tool gopls version | head -1)"
    @echo "golangci-lint: $(golangci-lint version --short)"
    @echo "gosec:         $(go tool gosec --version 2>&1)"
    @echo "govulncheck:   $(go tool govulncheck -version)"
    @echo "just:          $(just --version)"
