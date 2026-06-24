# Security Policy

## Security Goals

`goappmon` should avoid introducing avoidable operational or supply-chain risk.

## Baseline Practices

- do not commit secrets
- document required environment variables clearly
- validate external input
- minimize dependency count
- keep logs free of sensitive data

## Reporting Vulnerabilities

If a vulnerability process is added later, document:

- where to report issues privately
- expected response time
- triage and fix process
- disclosure timing

## Secure Development Rules

- review new dependencies carefully
- prefer built-in Go packages when sufficient
- avoid unsafe defaults
- keep transport and authentication settings explicit

