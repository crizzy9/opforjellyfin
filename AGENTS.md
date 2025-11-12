# Agent Guidelines for OpforJellyfin

## Build/Test Commands
- **Build**: `go build -o opfor`
- **Run single test**: `go test -run TestFunctionName ./internal/package`
- **Run all tests**: `go test ./...`
- **Lint**: `golangci-lint run` (available in nix shell)
- **Format**: `gofmt -w .` or `go fmt ./...`

## Code Style
- **Language**: Go 1.24.4
- **Imports**: Group stdlib, then blank line, then local packages, then blank line, then third-party
  ```go
  import (
      "fmt"
      
      "opforjellyfin/internal/shared"
      
      "github.com/spf13/cobra"
  )
  ```
- **Naming**: PascalCase for exported, camelCase for unexported (Go conventions)
- **Error handling**: Use `fmt.Errorf` with context, check all errors explicitly
- **Types**: Prefer explicit types, use struct tags for JSON serialization
- **Comments**: Minimal - focus on "why" not "what", avoid redundant comments
- **Testing**: Use table-driven tests (see `internal/shared/parsers_test.go`)
- **Note**: Do not create random md files as guides or quick fix or implementation summaries. Only do it when specifically instructed to do so. Keep everything in the chat
