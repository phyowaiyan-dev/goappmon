# Testing Guide

## Testing Goals

The project should be testable from the start.

Testing should prove:

- core behavior works as intended
- regressions are caught early
- error paths are handled
- public examples remain valid

## Minimum Test Baseline

- unit tests for core packages
- table-driven tests where they fit naturally
- integration tests for external dependencies
- smoke tests for executable entry points, once added

## Commands

```bash
go test ./...
```

## Testing Standards

- prefer deterministic tests
- avoid real network calls unless explicitly testing integrations
- isolate time, randomness, and filesystem access where practical
- make failures easy to understand

## Release Gate

Do not ship a release until:

- the full test suite passes
- formatting and linting pass
- documentation matches the current behavior

