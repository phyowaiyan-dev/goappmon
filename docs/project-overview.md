# Project Overview

## Purpose

`goappmon` is a production-oriented Go project for application monitoring and observability.

It is currently implemented as a lightweight application control center for mobile and web apps.

The project can grow into one or more of the following shapes:

- a reusable Go library
- a CLI for monitoring and diagnostics
- an agent or service for health and telemetry collection
- integrations for dashboards, alerting, or external observability platforms

## Current State

The repository currently contains the MVP application, docs, templates, and GitHub workflow scaffolding.

Current facts:

- module path: `github.com/phyowaiyan-dev/goappmon`
- Go version: `1.26.1`
- license: MIT
- database: SQLite at `storage/goappmon.sqlite`
- auth: bcrypt + secure cookie session
- web UI: html/template with Tailwind CDN
- public API: health, status, version, config, feature flags

## Principles

The project should be built with the following principles:

- production safety first
- clear defaults and minimal surprise
- testability and maintainability
- explicit configuration
- good observability of the observability tooling itself
- OSS-friendly contribution and release flow

## Non-Goals

The MVP does not include:

- Docker packaging
- Redis
- PostgreSQL
- .env-based configuration
- a JavaScript frontend framework

