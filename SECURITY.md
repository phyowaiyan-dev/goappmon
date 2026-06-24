# Security Policy

## Supported Versions

This repository is currently in its foundation stage, so there are no stable supported runtime releases yet.

As the project matures, supported versions should be listed here.

## Reporting a Vulnerability

If you find a security issue:

1. Do not open a public issue.
2. Describe the problem clearly and include reproduction steps.
3. Share any relevant logs, impact details, and affected versions.

If a private reporting channel is added later, document it here.

## Security Expectations

- never commit secrets
- avoid logging sensitive data
- keep dependencies minimal and reviewed
- validate all untrusted input
- use explicit timeouts and cancellation for networked code

## Responsible Disclosure

Security fixes should be coordinated carefully so users can update before details are published.

