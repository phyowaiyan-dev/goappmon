# Release Guide

## Release Model

The project should use predictable, documented releases.

Recommended release inputs:

- tagged versions
- changelog entries
- CI validation
- documented breaking changes

## Release Checklist

- version is tagged consistently
- tests pass
- README and docs reflect the shipped behavior
- security-sensitive changes are reviewed
- example commands still work

## Versioning

Use a clear versioning policy once the codebase stabilizes.

Common options:

- semantic versioning
- date-based releases
- internal pre-release tags for early development

## Artifacts to Publish

Depending on the future project shape, releases may include:

- Go module versions
- CLI binaries for Linux, macOS, and Windows
- release notes

For `goappmon`, the GitHub release workflow builds runnable archives for:

- Linux `amd64`
- Linux `arm64`
- macOS `amd64`
- macOS `arm64`
- Windows `amd64`

That means production operators can download a release asset and run the binary directly on the target server without rebuilding from source.
