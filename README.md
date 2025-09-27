# playable-ads-prototype

A SaaS platform that allows users to create and render playable ads (videos + interactive templates).

---

## Features

- User management and authentication
- Upload and manage assets (videos, templates)
- Render playable ads combining video and interactive elements
- REST API backend written in Go
- PostgreSQL for data persistence
- Redis for caching and session management
- NGINX as reverse proxy and static file server
- Dockerized development and deployment

---

## Prerequisites

- [Go 1.24.5+](https://golang.org/dl/)
- [Docker](https://www.docker.com/get-started)
- [Docker Compose](https://docs.docker.com/compose/install/)

---

## Quick Start

### 1. Setup environment

```bash
make setup
````

Installs Go module dependencies and necessary tools (e.g., `swag` for API docs).

### 2. Run locally (without Docker)

```bash
make run
```

Starts the backend server directly; requires Postgres and Redis running.

### 3. Build the binary

```bash
make build
```

Generates the `backend` executable.

---

## Docker-based development

Bring up the full stack (Postgres, Redis, backend app, NGINX):

```bash
make up
```

Stop and remove containers:

```bash
make down
```

Follow backend logs:

```bash
make logs
```

---

## API Documentation

Generate Swagger docs:

```bash
make docs
```

Access the docs (if served by backend or separate doc server):

```
http://localhost:8080/swagger/docs.json
```

---

## Project Structure

```
cmd/api/main.go         - Backend entrypoint
migrations/             - Database migration scripts
local/uploads           - User uploaded assets (mounted volume)
local/outputs           - Rendered output files (mounted volume)
nginx/                  - NGINX configuration
web/                    - Frontend static files
Dockerfile              - Backend Dockerfile
docker-compose.yml      - Compose stack definition
Makefile                - Convenience commands
```

---

## Configuration

Environment variables for backend (via Docker Compose or locally via .env):

* `SERVER_PORT`: Backend server port (default: 8080)
* `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`, `DB_SSL_MODE`: Postgres connection
* `JWT_SECRET`: Secret key for JWT signing
* `JWT_ACCESS_TOKEN_TTL`, `JWT_REFRESH_TOKEN_TTL`: Token expiration settings
* `REDIS_ADDR`, `REDIS_PASSWORD`, `REDIS_DB`: Redis connection
* `MAX_FILE_SIZE_MB`: Max upload file size
* `ENVIRONMENT`: App environment (`Production`(Doesn't load .env), `Development`)
