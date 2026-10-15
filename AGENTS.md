# AGENTS.md - NutriControl API Operational Context

## Project Summary
NutriControl API provides patient management endpoints using Go and Echo framework. Currently in early development (Day 2/3) with basic CRUD operations for patient data.

## Current Architecture
- **Language**: Go 1.26.1
- **Web Framework**: Echo v4.16.0 (minimalist approach)
- **Structure**: Clean separation between models (domain) and handlers (HTTP layer)
- **Current State**: Mock data implementation, no persistence layer yet

## Key Commands
```bash
# Run the API server
go run ./...

# Project structure
./main.go           # Entry point and server setup
./handlers/         # HTTP handlers and routes
./models/           # Domain models
./go.mod           # Dependencies
```

## Established Conventions
- **Code Structure**: Separate packages for domain vs. presentation layers
- **JSON Tags**: Struct fields include `json:"field_name"` annotations
- **Error Handling**: Basic HTTP error responses with JSON bodies
- **Mocking**: Starting with mocked data before implementing persistence
- **Commit Style**: Conventional commits with clear, actionable titles

## Boundaries (No-touch zones without explicit permission)
- **Data Persistence**: No files in `data/`, `db/`, or `storage/` directories (to be decided)
- **Testing Strategy**: No existing test files - future tests must follow Go testing conventions
- **API Documentation**: No OpenAPI/Swagger files present (to be added in future)
- **Configuration**: No `.env`, `config/`, or environment-specific setups currently