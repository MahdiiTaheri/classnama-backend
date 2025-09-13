# ClassNama

**ClassNama** is a backend application for managing school operations, developed in **Go** with **PostgreSQL** and **Redis**. It provides a robust and scalable API to manage students, teachers, executives, classrooms, and attendance, with built-in features such as rate limiting, authentication, and API documentation.

## Table of Contents

- [Features](#Features)
- [Tech Stack](#tech-stack)
- [Architecture](#Architecture)
- [Project Structure.](#project-struture)
  - [Prerequisites](#Prerequisites)
  - [Installation](#Installation)
  - [Running the Application](#running-the-application)
- [API Documentation](#api-documentation)
- [Database Migrations and Seeding](#database-migrations-and-seeding)
- [Rate Limiting](#rate-limiting)
- [Environment Variables](#environment-variables)

## Features

- Manage **students**, **teachers**, **executives**, and **classrooms**
- Record and track **attendance**
- **JWT-based authentication** for secure access
- **Rate limiting** to prevent abuse
- **Redis caching** for performance optimization
- Fully **documented API** with **Swagger**
- Database **migrations and seeding** for easy setup

## Tech Stack

- **Backend:** Go (Golang)
- **Database:** PostgreSQL
- **Caching:** Redis
- **Authentication:** JWT
- **API Documentation:** Swagger
- **Rate Limiting:** Token Bucket Algorithm

## Architecture

ClassNama follows a modular structure with separation of concerns:

- **cmd** – main application entry points and API handlers
- **internal** – core business logic
  - **auth** – authentication and JWT handling
  - **db** – database connection and seeding
  - **ratelimiter** – request rate limiting
  - **store** – database queries and cache logic
  - **utils** – helper functions
- **docs** – Swagger API documentation
- **scripts** – initial database scripts
  Redis is used to cache frequently accessed data (e.g., students, teachers, executives) to reduce database load. PostgreSQL handles persistent storage, ensuring ACID compliance for critical school data.

## Project Structure

```bash
cmd/
  api/         # API handlers and routes
  migrate/     # Database migrations and seeds
internal/
  auth/        # Authentication and JWT
  db/          # DB connection and seeding
  env/         # Environment config
  ratelimiter/ # Token bucket implementation
  store/       # DB queries, cache logic
  utils/       # Utility functions
docs/          # Swagger documentation
scripts/       # Database initialization scripts
```

## Getting Started

### Prerequisites

- Go >= 1.22
- PostgreSQL
- Redis
- Docker & Docker Compose (optional but recommended)
- Air (optional, for live-reload development)

### Installation

1. Clone the repository:
   ```bash
   git clone <your-repo-url>
   cd classnama
   ```
2. Create a `.env` file in `internal/env` with the following variables:

```sh
# Server configuration
ADDR=:8080
ENV=development
EXTERNAL_URL=http://localhost:8080

# Database configuration
DB_ADDR=postgres://postgres:password@localhost:5432/classnama?sslmode=disable
DB_MAX_OPEN_CONNS=30
DB_MAX_IDLE_CONNS=30
DB_MAX_IDLE_TIME=15m

# Authentication
AUTH_BASIC=admin
AUTH_BASIC_PASS=admin
AUTH_TOKEN_SECRET=your_jwt_secret

# Rate Limiter
RATE_LIMITER_REQUESTS_COUNT=100
RATE_LIMITER_TIME_FRAME=5s
RATE_LIMITER_ENABLED=true

# Redis configuration
REDIS_ADDR=localhost:6379
REDIS_PW=
REDIS_DB=0
REDIS_ENABLED=true
```

3. Install dependencies:

```bash
go mod tidy
```

### Running the Application

#### Using Docker Compose (Recommended)

```bash
docker-compose up --build
```

#### Using Air (Optional)

1.  Install Air for live-reload.
2.  Run the project:

```bash
air
```

The server will be available at:

```bash
http://localhost:8080
```

## API Documentation

Swagger API documentation is available at:

```bash
http://localhost:8080/v1/swagger/index.html
```

You can update the documentations via running the following command:

```bash
make gen-docs
```

## Database Migrations and Seeding

Run migrations first:

```bash
make migrate-up
```

Then seed the database:

```bash
go run cmd/migrate/seed/seed.go
```

## Rate Limiting

Implemented using a **Token Bucket algorithm** (`internal/ratelimiter/token-bucket.go`) to limit the number of API requests per client. This helps prevent abuse and ensures stable performance under load.

## Environment Variables

Key environment variables:

- **`ADDR`** – API server listen address (default :8080)
- **`ENV`** – Application environment (development, production, etc.)
- **`EXTERNAL_URL`** – External URL for API access and Swagger documentation
- **`DB_ADDR`** – PostgreSQL connection URI
- **`DB_MAX_OPEN_CONNS`** – Maximum open connections to the database
- **`DB_MAX_IDLE_CONNS`** – Maximum idle connections to the database
- **`DB_MAX_IDLE_TIME`** – Maximum idle time for connections
- **`AUTH_BASIC / AUTH_BASIC_PASS`** – Basic authentication credentials
- **`AUTH_TOKEN_SECRET`** – Secret key for JWT authentication
- **`RATE_LIMITER_REQUESTS_COUNT`** – Max allowed requests per timeframe
- **`RATE_LIMITER_TIME_FRAME`** – Timeframe for rate limiting
- **`RATE_LIMITER_ENABLED`** – Enable/disable rate limiting
- **`REDIS_ADDR / REDIS_PW / REDIS_DB`** – Redis connection settings
- **`REDIS_ENABLED`** – Enable/disable Redis caching

## Badges

[![Go](https://img.shields.io/badge/Language-Go-00ADD8?logo=go&logoColor=white)](https://golang.org/) [![PostgreSQL](https://img.shields.io/badge/Database-PostgreSQL-336791?logo=postgresql&logoColor=white)](https://www.postgresql.org/) [![Redis](https://img.shields.io/badge/Cache-Redis-DC382D?logo=redis&logoColor=white)](https://redis.io/) [![Docker](https://img.shields.io/badge/Container-Docker-2496ED?logo=docker&logoColor=white)](https://www.docker.com/) [![Swagger](https://img.shields.io/badge/API-Swagger-85EA2D?logo=swagger&logoColor=white)](https://swagger.io/) [![Go Version](https://img.shields.io/badge/Go-1.22-blue)](https://golang.org/doc/go1.25) [![MIT License](https://img.shields.io/badge/License-MIT-green.svg)](https://choosealicense.com/licenses/mit/) [![Build Status](https://img.shields.io/badge/Build-Passing-brightgreen)](#) <!-- replace # with your CI/CD URL if available -->
