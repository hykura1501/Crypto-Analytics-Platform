# Crawler Service

Automated news crawler for crypto-related articles from CoinDesk and CoinTelegraph.

## Features

- ✅ RSS feed parsing from CoinDesk and CoinTelegraph
- ✅ Auto-extraction using newspaper3k (handles dynamic HTML)
- ✅ Scheduled crawling every 10 minutes (APScheduler)
- ✅ Kafka integration (publishes to `news_new_article` topic)
- ✅ PostgreSQL storage with SQLAlchemy
- ✅ FastAPI REST endpoints

## Tech Stack

- **Framework**: FastAPI
- **Crawler**: newspaper3k + feedparser
- **Database**: PostgreSQL (SQLAlchemy ORM)
- **Message Queue**: Kafka
- **Scheduler**: APScheduler
- **Runtime**: Python 3.11

## API Endpoints

### Get News List
```
GET /news?skip=0&limit=20&source=CoinDesk
```

**Parameters:**
- `skip` (int): Pagination offset (default: 0)
- `limit` (int): Max results (default: 20, max: 100)
- `source` (string, optional): Filter by source (CoinDesk, CoinTelegraph)

**Response:**
```json
[
  {
    "id": 1,
    "source": "CoinDesk",
    "title": "Bitcoin Reaches New High",
    "content": "Full article text...",
    "published_at": "2024-01-01T12:00:00",
    "url": "https://...",
    "sentiment_score": null,
    "created_at": "2024-01-01T12:05:00"
  }
]
```

### Get News by ID
```
GET /news/{news_id}
```

### Get Statistics
```
GET /news/stats/summary
```

Returns total articles and breakdown by source.

### Health Check
```
GET /health
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

# Kafka
KAFKA_BROKER=kafka:9092
KAFKA_TOPIC=news_new_article

# RSS Feeds
COINDESK_RSS_URL=https://www.coindesk.com/arc/outboundfeeds/rss/
COINTELEGRAPH_RSS_URL=https://cointelegraph.com/rss

# Crawler
CRAWL_INTERVAL_MINUTES=10
MAX_ARTICLES_PER_RUN=20

# Server
SERVER_PORT=8083
```

## Database Model

```python
class News:
    id: int  # Primary key
    source: str  # CoinDesk, CoinTelegraph
    title: str
    content: str  # Auto-extracted full text
    published_at: datetime (nullable)
    url: str  # Unique constraint
    sentiment_score: float (nullable)  # For AI service
    created_at: datetime
```

## How It Works

### 1. Scheduled Crawling
APScheduler runs every 10 minutes:

```python
# Runs on startup + every 10 minutes
scheduler.add_job(run_crawler_job, IntervalTrigger(minutes=10))
```

### 2. RSS → newspaper3k Pipeline

```
1. Fetch RSS feed (CoinDesk/CoinTelegraph)
   ↓
2. Extract article URLs from feed
   ↓
3. Use newspaper3k to auto-extract:
   - Title
   - Full text content
   - Publish date
   ↓
4. Save to PostgreSQL
   ↓
5. Publish to Kafka topic
```

### 3. Auto-Extraction Benefits

**newspaper3k** solves dynamic HTML issues:
- Automatically identifies article content
- Filters out ads, navigation, sidebars
- Works on most news sites without custom parsers
- Extracts publish dates intelligently

### 4. Kafka Integration

After saving each article:
```python
kafka_producer.send(
    topic="news_new_article",
    key=news_id,
    value={
        "news_id": 1,
        "title": "...",
        "content": "..."  # Truncated to 1000 chars
    }
)
```

**Consumers** (AI Service) can:
- Perform sentiment analysis
- Extract keywords
- Categorize articles
- Update `sentiment_score` field

## Running Locally

```bash
# Install dependencies
pip install -r requirements.txt

# Set environment variables
cp .env.example .env

# Run
python -m uvicorn app.main:app --reload --port 8083
```

## Run with Docker

```bash
docker build -t crawler-service .
docker run -p 8083:8083 --env-file .env crawler-service
```

## Testing

```bash
# Get latest news
curl http://localhost:8083/news?limit=5

# Get CoinDesk articles only
curl http://localhost:8083/news?source=CoinDesk

# Get stats
curl http://localhost:8083/news/stats/summary

# Health check
curl http://localhost:8083/health
```

## Project Structure

```
crawler-service/
├── app/
│   ├── main.py              # FastAPI app + lifespan
│   ├── config.py            # Settings + env vars
│   ├── models.py            # News SQLAlchemy model
│   ├── database.py          # DB connection
│   ├── routes.py            # API endpoints
│   ├── crawler.py           # RSS + newspaper3k logic
│   ├── kafka_producer.py    # Kafka client
│   └── scheduler.py         # APScheduler jobs
├── requirements.txt
├── Dockerfile
└── .env.example
```

## Logs Example

```
2024-01-01 12:00:00 - INFO - 🚀 Starting Crawler Service...
2024-01-01 12:00:01 - INFO - Database tables created successfully
2024-01-01 12:00:01 - INFO - Starting scheduler (interval: 10 minutes)
2024-01-01 12:00:02 - INFO - 🕐 Starting scheduled crawl job...
2024-01-01 12:00:03 - INFO - Fetching RSS feed from CoinDesk
2024-01-01 12:00:04 - INFO - Found 20 articles from CoinDesk
2024-01-01 12:00:10 - INFO - ✅ Saved article: Bitcoin Reaches New High...
2024-01-01 12:00:10 - INFO - ✅ Published news #1 to Kafka
2024-01-01 12:00:30 - INFO - Crawl job completed. Saved 15 articles.
```

## Error Handling

- **RSS parse fails**: Logs error, skips source
- **newspaper3k extraction fails**: Logs warning, skips article  
- **Duplicate URL**: Skips (unique constraint)
- **Kafka unavailable**: Logs warning, continues saving to DB
- **Database error**: Rollback, log error

## Next Steps

- Add support for more news sources
- Implement retry logic for failed extractions
- Add article deduplication (title similarity)
- Create admin API to trigger manual crawls
- Add metrics/monitoring

## Integration with Other Services

**AI Service** subscribes to `news_new_article`:
```python
# AI service consumes Kafka
for message in consumer:
    news_id = message.value["news_id"]
    content = message.value["content"]
    
    # Analyze sentiment
    score = analyze_sentiment(content)
    
    # Update database
    update_news_sentiment(news_id, score)
```

Crawler service is **ready to run**! 🎉
