# Auth Service

Authentication microservice for the Crypto Analysis Platform.

## Features

- ✅ User registration with email and password
- ✅ Password hashing using bcrypt
- ✅ JWT-based authentication (access + refresh tokens)
- ✅ Token validation middleware
- ✅ Token refresh mechanism
- ✅ User logout (refresh token invalidation)
- ✅ PostgreSQL database with GORM
- ✅ Clean architecture (handler → service → repository)

## Tech Stack

- **Framework**: Gin (Go web framework)
- **Database**: PostgreSQL with GORM
- **Authentication**: JWT (golang-jwt/jwt)
- **Password Hashing**: bcrypt

## API Endpoints

### Public Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/auth/register` | Register new user |
| POST | `/api/v1/auth/login` | Login and get tokens |
| POST | `/api/v1/auth/refresh` | Refresh access token |
| POST | `/api/v1/auth/logout` | Logout (invalidate refresh token) |

### Protected Endpoints (require Bearer token)

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/auth/validate` | Validate current token |
| GET | `/api/v1/auth/me` | Get current user info |

### Health Check

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/health` | Service health status |

## Request/Response Examples

### Register
```bash
curl -X POST http://localhost:8081/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "securepassword",
    "first_name": "John",
    "last_name": "Doe"
  }'
```

### Login
```bash
curl -X POST http://localhost:8081/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "securepassword"
  }'
```

Response:
```json
{
  "access_token": "eyJhbGc...",
  "refresh_token": "eyJhbGc...",
  "token_type": "Bearer",
  "expires_in": 900,
  "user": {
    "id": 1,
    "email": "user@example.com",
    "first_name": "John",
    "last_name": "Doe",
    "is_active": true,
    "created_at": "2024-01-01T00:00:00Z"
  }
}
```

### Validate Token
```bash
curl -X GET http://localhost:8081/api/v1/auth/validate \
  -H "Authorization: Bearer <access_token>"
```

### Refresh Token
```bash
curl -X POST http://localhost:8081/api/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{
    "refresh_token": "eyJhbGc..."
  }'
```

### Get Current User
```bash
curl -X GET http://localhost:8081/api/v1/auth/me \
  -H "Authorization: Bearer <access_token>"
```

### Logout
```bash
curl -X POST http://localhost:8081/api/v1/auth/logout \
  -H "Content-Type: application/json" \
  -d '{
    "refresh_token": "eyJhbGc..."
  }'
```

## Environment Variables

Create a `.env` file based on `.env.example`:

```env
DB_HOST=postgres
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=crypto_db
DB_SSLMODE=disable

REDIS_HOST=redis
REDIS_PORT=6379

JWT_SECRET=your-super-secret-jwt-key-change-this-in-production
JWT_ACCESS_EXPIRY=15m
JWT_REFRESH_EXPIRY=7d

SERVER_PORT=8081
GIN_MODE=debug
```

## Project Structure

```
auth-service/
├── cmd/
│   └── main.go                 # Application entry point
├── config/
│   └── config.go               # Configuration management
├── internal/
│   ├── handler/
│   │   └── auth_handler.go     # HTTP handlers
│   ├── middleware/
│   │   └── auth.go             # JWT middleware
│   ├── model/
│   │   ├── user.go             # Database models
│   │   └── dto.go              # Request/Response DTOs
│   ├── repository/
│   │   ├── user_repository.go
│   │   └── refresh_token_repository.go
│   ├── router/
│   │   └── router.go           # Route definitions
│   └── service/
│       └── auth_service.go     # Business logic
├── pkg/
│   └── utils/
│       ├── jwt.go              # JWT utilities
│       └── password.go         # Password hashing
├── .env.example
├── Dockerfile
├── go.mod
└── README.md
```

## Running Locally

### With Docker Compose (Recommended)

```bash
# From project root
docker-compose up -d postgres redis
docker-compose up auth-service
```

### Standalone

```bash
# Install dependencies
go mod download

# Run the service
go run cmd/main.go
```

## Database Migrations

The service automatically runs migrations on startup using GORM AutoMigrate.

Tables created:
- `users` - User accounts
- `refresh_tokens` - Refresh token storage

## Security Features

- ✅ Passwords hashed with bcrypt (cost factor 10)
- ✅ JWT tokens with configurable expiry
- ✅ Separate access and refresh tokens
- ✅ Refresh token stored in database for revocation
- ✅ Token validation middleware
- ✅ CORS enabled

## Development

### Adding New Endpoints

1. Add method to `AuthService` interface in `internal/service/auth_service.go`
2. Implement the method in `authService` struct
3. Add handler method in `internal/handler/auth_handler.go`
4. Register route in `internal/router/router.go`

### Running Tests

```bash
go test ./... -v
```

## Docker Build

```bash
docker build -t auth-service .
```

## Contributing

Follow the Go project layout and clean architecture principles.
