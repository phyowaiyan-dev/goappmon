# Internal

This directory is reserved for private implementation details that should not be treated as part of the public API surface.

Use this folder for:

- internal services and adapters
- private helpers
- implementation notes that guide the codebase
- task lists that are only relevant to maintainers

Anything that should be consumed by other packages should live elsewhere, such as `pkg/` if the project later exposes reusable public packages.

