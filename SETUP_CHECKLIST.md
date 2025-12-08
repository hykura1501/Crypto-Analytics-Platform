# Quick Setup Checklist - New Features

## ✅ Prerequisites

- [ ] Go 1.21+ installed
- [ ] Node.js 18+ installed
- [ ] Python 3.10+ installed
- [ ] PostgreSQL 15+ running
- [ ] Redis 7+ running
- [ ] Docker (optional, for containerized deployment)

## 🚀 Setup Backend với WebSocket Scaling

### 1. Install Go dependencies:

```bash
cd backend
go mod tidy
```

Expected new packages:
- `github.com/golang-jwt/jwt/v5` - JWT authentication
- `golang.org/x/time` - Rate limiting

### 2. Set environment variables:

```bash
# .env file in backend/
DATABASE_URL=postgres://crypto_user:crypto_pass@localhost:5432/cryptodb?sslmode=disable
REDIS_URL=redis://localhost:6379
JWT_SECRET=your-super-secret-key-change-in-production
PORT=8080
```

### 3. Start Redis:

```bash
# Option 1: Docker
docker run -d -p 6379:6379 redis:alpine

# Option 2: Local
redis-server
```

### 4. Run backend instance(s):

```bash
# Single instance for development
go run main.go

# Multiple instances for testing scaling
PORT=8081 go run main.go &
PORT=8082 go run main.go &
PORT=8083 go run main.go &
```

### 5. (Optional) Run Nginx load balancer:

```bash
# Update nginx.conf with your backend ports
# Then start nginx
nginx -c /path/to/crypto-analytics-platform/nginx.conf

# Or with Docker
docker run -d -p 80:80 \
  -v $(pwd)/nginx.conf:/etc/nginx/nginx.conf:ro \
  nginx:alpine
```

## 🤖 Setup AI-Enhanced Crawler

### 1. Install Python dependencies:

```bash
cd services/crawler
pip install -r requirements.txt
```

New packages:
- `openai==1.12.0` - OpenAI GPT integration
- `anthropic==0.18.1` - Anthropic Claude integration
- `playwright==1.41.0` - JavaScript site crawling

### 2. Install Playwright browsers:

```bash
playwright install chromium
```

### 3. Set API keys:

```bash
# .env file in services/crawler/
DATABASE_URL=postgres://crypto_user:crypto_pass@localhost:5432/cryptodb?sslmode=disable
OPENAI_API_KEY=sk-...your-key-here...
ANTHROPIC_API_KEY=sk-ant-...your-key-here...
```

**Note:** API keys are optional. Crawler will work with CSS selectors only if keys not provided.

### 4. Test LLM parser:

```bash
python llm_parser.py
```

### 5. Run crawler with LLM support:

```bash
# Crawl all sources
python main.py --all

# Crawl specific source
python main.py --source 1
```

## 🔐 Setup Authentication & Security

### 1. Generate JWT secret:

```bash
# Generate a secure random secret
openssl rand -hex 32

# Add to .env
JWT_SECRET=<generated-secret>
```

### 2. Update database schema (if needed):

The `users` table should already exist from migrations. Verify:

```sql
SELECT * FROM users LIMIT 1;
```

### 3. Test authentication endpoints:

```bash
# Register a user
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "username": "testuser",
    "password": "SecurePass123!"
  }'

# Login
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "SecurePass123!"
  }'

# Should return access_token and refresh_token
```

### 4. Test rate limiting:

```bash
# Run this multiple times quickly
for i in {1..105}; do
  curl http://localhost:8080/api/v1/pairs
  echo "Request $i"
done

# Should get 429 Too Many Requests after 100 requests
```

## 📊 Verification Checklist

### Backend:

- [ ] Backend starts without errors
- [ ] Redis connection successful (check logs: "Started Redis subscription")
- [ ] Price updater service running (check logs: "Starting price updater service")
- [ ] WebSocket connections work (`wscat -c ws://localhost:8080/ws/price/BTCUSDT`)
- [ ] Health check returns 200 (`curl http://localhost:8080/health`)

### Crawler:

- [ ] Python dependencies installed
- [ ] Database connection successful
- [ ] CSS extraction works for basic sites
- [ ] LLM fallback works (if API keys provided)
- [ ] Statistics printed correctly

### Security:

- [ ] User registration works
- [ ] Login returns JWT tokens
- [ ] Protected endpoints require Authorization header
- [ ] Rate limiting blocks excess requests (429 status)
- [ ] Audit logs appear in console

## 🧪 Testing

### WebSocket Scaling Test:

```bash
# Install wscat if not already
npm install -g wscat

# Connect multiple clients
wscat -c ws://localhost:8080/ws/price/BTCUSDT &
wscat -c ws://localhost:8080/ws/price/ETHUSDT &
wscat -c ws://localhost:8080/ws/price/BNBUSDT &

# Check health endpoint to see connection count
curl http://localhost:8080/health
# Should show: {"status":"healthy","connections":3}
```

### AI Crawler Test:

```bash
cd services/crawler

# Test with a known crypto news URL
python -c "
from llm_parser import SmartCrawler
import requests

session = requests.Session()
session.headers.update({
    'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36'
})

crawler = SmartCrawler(session, None)

result = crawler.crawl_article(
    'https://cointelegraph.com/news',
    {
        'title_selectors': ['h1', 'h2'],
        'content_selectors': ['article', '.post-content']
    }
)

if result:
    print('✅ Success!')
    print(f'Title: {result[\"title\"][:50]}...')
    print(f'Content length: {len(result[\"content\"])} chars')
else:
    print('❌ Failed')
"
```

### Authentication Test:

```bash
# Full auth flow test script
#!/bin/bash

echo "1. Registering user..."
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@test.com","username":"test","password":"password123"}'

echo -e "\n\n2. Logging in..."
RESPONSE=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@test.com","password":"password123"}')

TOKEN=$(echo $RESPONSE | jq -r '.access_token')
echo "Got token: ${TOKEN:0:20}..."

echo -e "\n\n3. Accessing protected endpoint..."
curl http://localhost:8080/api/v1/account/profile \
  -H "Authorization: Bearer $TOKEN"

echo -e "\n\n4. Testing invalid token..."
curl http://localhost:8080/api/v1/account/profile \
  -H "Authorization: Bearer invalid-token"

echo -e "\n\nDone!"
```

## 🐛 Troubleshooting

### Backend won't start:

```bash
# Check if required services are running
redis-cli ping  # Should return PONG
psql -U crypto_user -d cryptodb -c "SELECT 1"  # Should return 1

# Check for port conflicts
lsof -i :8080
lsof -i :6379
lsof -i :5432

# View full error logs
go run main.go 2>&1 | tee backend.log
```

### Crawler errors:

```bash
# Test database connection
python -c "
import psycopg2
import os
conn = psycopg2.connect(os.getenv('DATABASE_URL'))
print('✅ DB connection OK')
"

# Test LLM API keys
python -c "
import openai
import os
openai.api_key = os.getenv('OPENAI_API_KEY')
response = openai.models.list()
print('✅ OpenAI API OK')
"

# Check for missing dependencies
pip install -r requirements.txt --upgrade
```

### WebSocket not working:

```bash
# Test WebSocket connection manually
wscat -c ws://localhost:8080/ws/price/BTCUSDT

# Check if Redis pub/sub is working
redis-cli
> SUBSCRIBE price_updates
# Should see messages coming in

# Check backend logs for Redis connection errors
grep -i redis backend.log
```

### Rate limiting not working:

```bash
# Check Redis connection
redis-cli KEYS "rate_limit:*"
# Should show rate limit keys

# Test rate limit directly
curl -v http://localhost:8080/api/v1/pairs
# Check for X-RateLimit-* headers

# View rate limit logs
grep "Rate limit" backend.log
```

## 📈 Monitoring

### Check system status:

```bash
# Backend health
curl http://localhost:8080/health | jq

# Redis status
redis-cli INFO stats

# PostgreSQL connections
psql -U crypto_user -d cryptodb -c "
SELECT count(*) as total_connections 
FROM pg_stat_activity 
WHERE datname = 'cryptodb';
"

# Check WebSocket connections
curl http://localhost:8080/health | jq '.connections'
```

### View logs:

```bash
# Backend logs (real-time)
tail -f backend.log | grep -E "AUDIT|ERROR|WebSocket"

# Crawler logs
tail -f crawler.log

# Redis logs
redis-cli MONITOR
```

## 🎉 Success Criteria

Your setup is complete when:

- [x] Backend starts and connects to Redis + PostgreSQL
- [x] WebSocket connections receive real-time price updates
- [x] Crawler successfully extracts news (with CSS or LLM)
- [x] User registration and login work
- [x] Protected endpoints require valid JWT
- [x] Rate limiting blocks excess requests
- [x] Health check endpoint returns 200
- [x] Audit logs appear in console

## 🚀 Next Steps

1. **Frontend Integration:**
   - Update frontend to use new auth endpoints
   - Implement token refresh logic
   - Add WebSocket reconnection

2. **Production Deployment:**
   - Set up Kubernetes manifests
   - Configure proper secrets management
   - Set up monitoring (Prometheus + Grafana)

3. **Optimizations:**
   - Implement bcrypt password hashing
   - Add token blacklist with Redis
   - Set up distributed tracing

## 📚 Documentation

- [SITUATION_ANALYSIS.md](./SITUATION_ANALYSIS.md) - Detailed technical analysis
- [IMPLEMENTATION_GUIDE.md](./IMPLEMENTATION_GUIDE.md) - Usage guide
- [ARCHITECTURE_DIAGRAMS.md](./ARCHITECTURE_DIAGRAMS.md) - Visual diagrams
- [SUMMARY.md](./SUMMARY.md) - Executive summary

## ❓ Need Help?

Check the troubleshooting section above or refer to the detailed documentation.

**Common issues:**
- "Redis connection failed" → Make sure Redis is running on port 6379
- "Database error" → Check DATABASE_URL and PostgreSQL status
- "LLM extraction failed" → Verify OPENAI_API_KEY is set correctly
- "Rate limit not working" → Redis must be running
