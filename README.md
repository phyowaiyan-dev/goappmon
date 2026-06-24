# goappmon

`goappmon` is a lightweight application control center for mobile and web applications.

It ships as a single Go binary with SQLite storage, Gin HTTP handlers, bcrypt-backed admin auth, server-rendered HTML, and JSON APIs for app status, version policy, and feature flags.

## Start Here

- [Project overview](docs/project-overview.md)
- [Tech stack](docs/tech-stack.md)
- [Architecture](docs/architecture.md)
- [Development guide](docs/development.md)
- [Testing guide](docs/testing.md)
- [Release guide](docs/release.md)
- [Security policy](docs/security.md)
- [Contributing guide](docs/contributing.md)
- [Roadmap](docs/roadmap.md)
- [Docs index](docs/README.md)
- [Contributor guide](CONTRIBUTING.md)
- [Code of conduct](CODE_OF_CONDUCT.md)
- [Security reporting](SECURITY.md)
- [Changelog](CHANGELOG.md)

## Current Status

- Module path: `github.com/phyowaiyan-dev/goappmon`
- Go version: `1.26.1`
- License: MIT
- Storage: `storage/goappmon.sqlite`
- Server-rendered admin UI: yes
- Public JSON APIs: yes
- Source code: implemented MVP

## Repository Contents

- `go.mod` - module definition
- `LICENSE` - MIT license
- `README.md` - landing page and docs entry point
- `docs/` - production-oriented project documentation
- `CONTRIBUTING.md` - contributor workflow and standards
- `CODE_OF_CONDUCT.md` - community behavior policy
- `SECURITY.md` - vulnerability reporting and security contact guidance
- `CHANGELOG.md` - release history placeholder
- `internal/` - private implementation notes and work items

## Build

```bash
go build ./...
```

## Run

```bash
go run ./cmd/goappmon
```

## License

This project is licensed under the MIT License. See [LICENSE](LICENSE) for the full text.
