# Market Service

Real-time and historical cryptocurrency market data service for the Crypto Analysis Platform.

## Features

- ✅ Historical candlestick (KLINE) data from Binance REST API
- ✅ Real-time market data via Binance WebSocket
- ✅ Kafka integration (publish market updates)
- ✅ WebSocket streaming to frontend clients
- ✅ PostgreSQL storage with TimescaleDB-optimized indexes
- ✅ Composite primary key (symbol, time) for efficient time-series queries

## Tech Stack

- **Framework**: Gin (Go web framework)
- **Database**: PostgreSQL (GORM) with TimescaleDB concepts
- **Real-time**: Gorilla WebSocket
- **Message Queue**: Kafka (segmentio/kafka-go)
- **External API**: Binance Public API

## Architecture

```
Binance WebSocket → Market Service → [Kafka, WebSocket Clients, PostgreSQL]
                          ↓
                    REST API (Historical Data)
```

## API Endpoints

### Health Check
```
GET /health
```

### Get Historical Data (from Binance)
```
GET /api/v1/market/history?symbol=BTCUSDT&interval=1h&limit=100
```

**Parameters:**
- `symbol` (required): Trading pair (e.g., BTCUSDT, ETHUSDT)
- `interval` (required): Kline interval
  - `1m`, `3m`, `5m`, `15m`, `30m`
  - `1h`, `2h`, `4h`, `6h`, `8h`, `12h`
  - `1d`, `3d`, `1w`, `1M`
- `limit` (optional): Number of records (default: 100, max: 1000)
- `from` (optional): Start time in milliseconds
- `to` (optional): End time in milliseconds

**Response:**
```json
{
  "success": true,
  "data": [
    {
      "symbol": "BTCUSDT",
      "time": "2024-01-01T00:00:00Z",
      "interval": "1h",
      "open": 42000.50,
      "high": 42500.00,
      "low": 41800.00,
      "close": 42200.00,
      "volume": 1234.56
    }
  ]
}
```

### Get Data from Database
```
GET /api/v1/market/data?symbol=BTCUSDT&interval=1h&limit=50
```

**Parameters:**
- `symbol` (required): Trading pair
- `interval` (required): Kline interval
- `limit` (optional): Number of records (default: 100)
- `from` (optional): Start time (RFC3339 format)
- `to` (optional): End time (RFC3339 format)

### WebSocket - Real-time Prices
```
ws://localhost:8082/ws/prices
```

Connect to receive real-time market updates for all tracked symbols.

**Message Format:**
```json
{
  "symbol": "BTCUSDT",
  "time": "2024-01-01T12:00:00Z",
  "interval": "1m",
  "open": 42100.00,
  "high": 42150.00,
  "low": 42050.00,
  "close": 42120.00,
  "volume": 15.23
}
```

## Configuration

Environment variables (`.env`):

```env
# Database
DB_HOST=postgres
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=crypto_db
DB_SSLMODE=disable

# Kafka
KAFKA_BROKER=kafka:9092
KAFKA_TOPIC=market_price_updates

# Binance
BINANCE_API_URL=https://api.binance.com
BINANCE_WS_URL=wss://stream.binance.com:9443

# Server
SERVER_PORT=8082
GIN_MODE=debug

# Symbols to track (comma-separated)
DEFAULT_SYMBOLS=BTCUSDT,ETHUSDT,BNBUSDT
```

## Database Schema

### Table: `market_prices`

| Column | Type | Description |
|--------|------|-------------|
| symbol | VARCHAR(20) | Trading pair symbol (PK) |
| time | TIMESTAMP | Kline open time (PK) |
| interval | VARCHAR(10) | Time interval |
| open | DECIMAL(20,8) | Opening price |
| high | DECIMAL(20,8) | Highest price |
| low | DECIMAL(20,8) | Lowest price |
| close | DECIMAL(20,8) | Closing price |
| volume | DECIMAL(20,8) | Trading volume |

**Indexes:**
- Composite primary key: `(symbol, time)`
- Index: `(symbol, time DESC)` - for efficient time-series queries
- Index: `(time DESC)` - for global time-based queries

## Kafka Integration

The service publishes all real-time market updates to the Kafka topic `market_price_updates`.

**Message Key**: Symbol (e.g., "BTCUSDT")

**Message Value**: JSON representation of MarketPrice

Other services can consume this topic to:
- Trigger AI analysis
- Update dashboards
- Send notifications
- Store in different formats

## Real-time Data Flow

```
1. Binance WebSocket → Service receives kline update
2. Service processes → Parses to MarketPrice
3. Three actions (concurrent):
   a) Publish to Kafka topic
   b) Broadcast to WebSocket clients
   c) Save to PostgreSQL
```

## Running the Service

### With Docker Compose
```bash
# From project root
docker-compose up market-service
```

### Standalone
```bash
# Install dependencies
go mod download

# Run
go run cmd/main.go
```

## Testing

### Test Historical Data
```bash
curl "http://localhost:8082/api/v1/market/history?symbol=BTCUSDT&interval=1h&limit=10"
```

### Test WebSocket
```javascript
const ws = new WebSocket('ws://localhost:8082/ws/prices');

ws.onmessage = (event) => {
  const data = JSON.parse(event.data);
  console.log('Price update:', data);
};

ws.onerror = (error) => {
  console.error('WebSocket error:', error);
};
```

### Test with wscat
```bash
npm install -g wscat
wscat -c ws://localhost:8082/ws/prices
```

## TimescaleDB Optimization (Optional)

For better performance with time-series data, you can convert the `market_prices` table to a TimescaleDB hypertable:

```sql
-- After creating the table
SELECT create_hypertable('market_prices', 'time', 
  chunk_time_interval => INTERVAL '1 day',
  if_not_exists => TRUE
);

-- Enable compression
ALTER TABLE market_prices SET (
  timescaledb.compress,
  timescaledb.compress_segmentby = 'symbol, interval'
);

-- Add compression policy (compress data older than 7 days)
SELECT add_compression_policy('market_prices', INTERVAL '7 days');
```

## Project Structure

```
market-service/
├── cmd/
│   └── main.go                      # Entry point
├── config/
│   └── config.go                    # Configuration
├── internal/
│   ├── handler/
│   │   └── market_handler.go        # HTTP + WebSocket handlers
│   ├── model/
│   │   ├── market.go                # Data models
│   │   └── dto.go                   # DTOs
│   ├── repository/
│   │   └── market_repository.go     # Database layer
│   ├── router/
│   │   └── router.go                # Route definitions
│   └── service/
│       └── market_service.go        # Business logic
├── pkg/
│   ├── binance/
│   │   ├── client.go                # Binance REST API
│   │   └── websocket.go             # Binance WebSocket
│   ├── kafka/
│   │   └── producer.go              # Kafka producer
│   └── websocket/
│       └── hub.go                   # WebSocket hub for clients
├── .env.example
├── Dockerfile
└── go.mod
```

## Dependencies

- `github.com/gin-gonic/gin` - Web framework
- `github.com/gorilla/websocket` - WebSocket implementation
- `github.com/segmentio/kafka-go` - Kafka client
- `gorm.io/gorm` - ORM
- `gorm.io/driver/postgres` - PostgreSQL driver

## Next Steps

- Add rate limiting for Binance API calls
- Implement data backfilling
- Add metrics and monitoring (Prometheus)
- Implement graceful reconnection for WebSocket errors
- Add unit tests
- Add integration tests
- Support multiple intervals per symbol
