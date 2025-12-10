# Crypto Analysis Platform - Mono-Repo Structure

This is a microservices-based crypto analysis system built with Golang, Python, and modern infrastructure.

## Tech Stack
- **Backend Services**: Golang (Gin Framework)
- **AI/Crawler Services**: Python
- **Database**: PostgreSQL
- **Cache**: Redis
- **Message Broker**: Apache Kafka
- **Frontend**: React with Vite

## Project Structure

```
v4/
├── docker-compose.yml          # Infrastructure orchestration
├── services/                   # Microservices directory
│   ├── auth-service/          # Authentication & Authorization (Go)
│   │   ├── cmd/               # Application entry points
│   │   ├── internal/          # Private application code
│   │   │   ├── handler/       # HTTP handlers
│   │   │   ├── service/       # Business logic
│   │   │   ├── repository/    # Data access layer
│   │   │   └── model/         # Data models
│   │   ├── pkg/               # Public libraries
│   │   ├── config/            # Configuration files
│   │   ├── Dockerfile         # Container definition
│   │   ├── go.mod             # Go dependencies
│   │   └── README.md          # Service documentation
│   │
│   ├── market-service/        # Market Data Management (Go)
│   │   ├── cmd/
│   │   ├── internal/
│   │   │   ├── handler/
│   │   │   ├── service/
│   │   │   ├── repository/
│   │   │   └── model/
│   │   ├── pkg/
│   │   ├── config/
│   │   ├── Dockerfile
│   │   ├── go.mod
│   │   └── README.md
│   │
│   ├── crawler-service/       # Data Crawling Service (Python)
│   │   ├── src/               # Source code
│   │   │   ├── crawlers/      # Crawler implementations
│   │   │   ├── processors/    # Data processing
│   │   │   ├── models/        # Data models
│   │   │   └── utils/         # Utility functions
│   │   ├── config/            # Configuration files
│   │   ├── requirements.txt   # Python dependencies
│   │   ├── Dockerfile
│   │   └── README.md
│   │
│   ├── ai-service/            # AI Analysis Service (Python)
│   │   ├── src/
│   │   │   ├── models/        # ML models
│   │   │   ├── api/           # REST API
│   │   │   ├── training/      # Model training
│   │   │   └── utils/         # Utilities
│   │   ├── config/
│   │   ├── requirements.txt
│   │   ├── Dockerfile
│   │   └── README.md
│   │
│   └── api-gateway/           # API Gateway (Go)
│       ├── cmd/
│       ├── internal/
│       │   ├── router/        # Route definitions
│       │   ├── middleware/    # Middleware (auth, logging, etc.)
│       │   ├── proxy/         # Service proxying
│       │   └── config/        # Configuration
│       ├── pkg/
│       ├── Dockerfile
│       ├── go.mod
│       └── README.md
│
├── frontend/                   # React Frontend (Vite)
│   ├── public/                # Static assets
│   ├── src/
│   │   ├── components/        # React components
│   │   ├── pages/             # Page components
│   │   ├── services/          # API services
│   │   ├── hooks/             # Custom React hooks
│   │   ├── store/             # State management
│   │   ├── utils/             # Utility functions
│   │   ├── styles/            # CSS/styling
│   │   ├── App.tsx            # Root component
│   │   └── main.tsx           # Entry point
│   ├── Dockerfile
│   ├── package.json
│   ├── vite.config.ts
│   └── README.md
│
├── scripts/                    # Utility scripts
│   ├── init-db.sh             # Database initialization
│   ├── start-dev.sh           # Development startup
│   └── deploy.sh              # Deployment script
│
├── docs/                       # Documentation
│   ├── architecture.md        # Architecture overview
│   ├── api/                   # API documentation
│   └── deployment.md          # Deployment guide
│
└── README.md                   # Main project documentation
```

## Quick Start

### Prerequisites
- Docker & Docker Compose
- Go 1.21+ (for local development)
- Python 3.11+ (for local development)
- Node.js 18+ (for frontend development)

### Starting Infrastructure

```bash
# Start all services
docker-compose up -d

# View logs
docker-compose logs -f

# Stop all services
docker-compose down

# Stop and remove volumes
docker-compose down -v
```

### Service Ports

| Service         | Port  | Description                    |
|----------------|-------|--------------------------------|
| API Gateway    | 8080  | Main entry point               |
| Auth Service   | 8081  | Authentication service         |
| Market Service | 8082  | Market data service            |
| AI Service     | 8083  | AI analysis service            |
| Frontend       | 3000  | React application              |
| PostgreSQL     | 5432  | Database                       |
| Redis          | 6379  | Cache                          |
| Kafka          | 29092 | Message broker (external)      |
| Zookeeper      | 2181  | Kafka coordination             |

## Service Communication

```
┌─────────┐
│ Frontend│
└────┬────┘
     │
     ▼
┌─────────────┐
│ API Gateway │
└──────┬──────┘
       │
       ├──────► Auth Service
       │
       ├──────► Market Service ◄──── Crawler Service (via Kafka)
       │                                      │
       └──────► AI Service ◄──────────────────┘
                    │
                    └──────► Market Service (data retrieval)
```

## Kafka Topics

- `crypto.market.data` - Market data updates
- `crypto.crawler.events` - Crawler events
- `crypto.ai.predictions` - AI predictions
- `crypto.user.events` - User activity events

## Database Schema

The `crypto_db` database will contain tables for:
- **auth-service**: users, sessions, permissions
- **market-service**: coins, prices, market_data, trading_pairs
- **ai-service**: predictions, models, training_data

## Development Workflow

### Go Services (auth-service, market-service, api-gateway)

```bash
cd services/auth-service
go mod init auth-service
go mod tidy
go run cmd/main.go
```

### Python Services (crawler-service, ai-service)

```bash
cd services/crawler-service
python -m venv venv
source venv/bin/activate  # On Windows: venv\Scripts\activate
pip install -r requirements.txt
python src/main.py
```

### Frontend

```bash
cd frontend
npm install
npm run dev
```

## Next Steps

1. **Initialize Service Structure**: Create the internal folder structure for each service
2. **Database Migrations**: Set up database migration tools (e.g., golang-migrate, Alembic)
3. **Shared Libraries**: Create shared packages for common functionality
4. **CI/CD Pipeline**: Set up automated testing and deployment
5. **Monitoring**: Add Prometheus, Grafana for observability
6. **API Documentation**: Generate OpenAPI/Swagger docs

## Environment Variables

Each service requires environment variables for configuration. See individual service READMEs for specific requirements.

Common environment variables:
- `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`
- `REDIS_HOST`, `REDIS_PORT`
- `KAFKA_BROKER`

## License

[Your License Here]
