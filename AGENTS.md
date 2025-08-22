# Agent Instructions for git-shippr

## Build/Lint/Test Commands

### Building

- **Build binary**: `go build -o shippr ./cmd/shippr`
- **Build all commands**: `go build ./cmd/...`
- **Clean build**: `go clean && go build -o shippr ./cmd/shippr`

### Testing

- **Run all tests**: `go test ./...`
- **Run tests with coverage**: `go test -cover ./...`
- **Run tests with verbose output**: `go test -v ./...`
- **Run single test**: `go test -run TestName ./internal/gh`
- **Run package tests**: `go test ./internal/gh`

### Linting & Formatting

- **Format code**: `gofmt -w .`
- **Check formatting**: `gofmt -d .`
- **Vet code**: `go vet ./...`
- **Run all checks**: `go vet ./... && gofmt -d .`

### Dependencies

- **Tidy modules**: `go mod tidy`
- **Download dependencies**: `go mod download`
- **Verify dependencies**: `go mod verify`

## Code Style Guidelines

### General Principles

- Follow standard Go conventions and idioms
- Use `gofmt` for consistent formatting
- Keep functions short and focused (single responsibility)
- Use descriptive names that reflect purpose
- Handle errors explicitly and return them with context
- Do not add code comments unless something is really complex to understand via code itself

### Imports

```go
import (
    "context"
    "fmt"
    "os"

    "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/bubbles/list"

    "git-shippr/internal/gh"
)
```

- Group imports: standard library, blank line, third-party, blank line, local
- Use full import paths, no dot imports
- Import local packages with module prefix

### Naming Conventions

- **Variables/Functions**: camelCase for unexported, PascalCase for exported
- **Types/Structs**: PascalCase for exported types
- **Constants**: PascalCase for exported, camelCase for unexported
- **Methods**: PascalCase, receiver names should be short (1-2 chars)
- **Files**: snake_case.go, test files as \*\_test.go

### Error Handling

```go
// Good: wrap errors with context
if err != nil {
    return fmt.Errorf("failed to list PRs: %w", err)
}

// Good: use context for cancellation
ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
defer cancel()
```

### Struct Tags and Types

```go
type PR struct {
    Number      int    `json:"number"`
    Title       string `json:"title"`
    HeadRefName string `json:"headRefName"`
}
```

- Use JSON struct tags for API responses
- Keep struct field names in PascalCase
- Use meaningful JSON tag names that match API

### Functions and Methods

- Use context.Context as first parameter for functions that may block
- Return errors as last return value
- Keep functions under 50 lines when possible
- Use early returns for error conditions

### Testing

```go
func TestSlug(t *testing.T) {
    if Slug("org", "repo") != "org/repo" {
        t.Fatalf("unexpected slug")
    }
}
```

- Use standard Go testing package
- Use t.Fatalf for test failures
- Test files named \*\_test.go
- Keep tests simple and focused

### Context Usage

- Always pass context through call chains
- Use context.WithTimeout for operations that might hang
- Cancel contexts in goroutines to prevent leaks

### Concurrency

- Use sync.WaitGroup for coordinating goroutines
- Use channels for communication between goroutines
- Limit concurrency with semaphores when needed
- Always close channels to prevent deadlocks

### Command Line Interface

- Use flag package for CLI arguments
- Provide helpful usage messages
- Exit with appropriate status codes (0 for success, 1 for error)
- Write errors to stderr, normal output to stdout
