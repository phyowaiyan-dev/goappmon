# Development Guide

## Local Setup

1. Install Go `1.26.1` or newer.
2. Clone the repository.
3. Run `go mod tidy` after dependencies or source files are added.

## Common Commands

```bash
go fmt ./...
go test ./...
go build ./...
```

## Workflow

1. Make the smallest useful change.
2. Add or update tests.
3. Keep docs in sync with behavior.
4. Verify formatting and tests locally.
5. Open a pull request with a clear description.

## Code Expectations

- prefer small, composable packages
- keep public APIs minimal and stable
- use descriptive names
- avoid hidden global state when possible
- handle errors explicitly

## Branch Hygiene

- use feature branches for all changes
- keep commits focused
- avoid mixing unrelated refactors into feature work

