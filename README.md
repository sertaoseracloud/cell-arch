# cell-arch

A multicloud sidecar PoC written in Go, following a clean-architecture layout.

## Repository

https://github.com/sertaoseracloud/cell-arch

## Description

This project implements a multicloud sidecar proxy using Go, with support for AWS and Azure SDKs. It follows Clean Architecture principles with strict separation between domain logic and infrastructure concerns.

## Getting Started

- **Build**: `go build ./cmd/...`
- **Run Application**: `go run ./cmd/app`
- **Run Sidecar**: `go run ./cmd/sidecar`
- **Run Tests**: `go test ./...`

## Architecture

```
repo-root/
├─ cmd/
│   ├─ app/          # Main application entry point
│   └─ sidecar/      # Multicloud sidecar proxy
├─ internal/
│   ├─ domain/       # Entities and repository interfaces
│   ├─ usecase/      # Business logic
│   └─ infrastructure/ # Cloud implementations
├─ pkg/               # Reusable utilities
└─ infra/             # Terraform IaC
```
