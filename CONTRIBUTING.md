# Contributing

Thanks for your interest in `goappmon`.

This project is intended to be maintained as a production-quality open-source repository. Contributions are welcome as long as they stay focused, testable, and aligned with the project documentation.

## Before You Start

Please read:

- [Project overview](docs/project-overview.md)
- [Tech stack](docs/tech-stack.md)
- [Architecture](docs/architecture.md)
- [Development guide](docs/development.md)
- [Testing guide](docs/testing.md)
- [Security policy](SECURITY.md)

## Good Contributions

- bug fixes
- tests
- documentation improvements
- dependency or build fixes
- code cleanup that improves maintainability

## Contribution Workflow

1. Open an issue or check existing work first.
2. Keep the change narrow and easy to review.
3. Add or update tests when behavior changes.
4. Update docs when commands, configuration, or output changes.
5. Run formatting and verification locally.
6. Open a pull request with a clear description of the change.

## Local Checks

```bash
go fmt ./...
go test ./...
go build ./...
```

## Pull Request Expectations

- explain what changed and why
- call out breaking changes clearly
- include screenshots, logs, or examples when relevant
- keep unrelated cleanup out of feature PRs unless necessary

## Code Style

- use standard Go formatting
- keep functions and packages small when possible
- prefer readability over cleverness
- avoid introducing dependencies unless they solve a real problem

## Documentation Rule

If the code changes behavior, the docs should change too.

