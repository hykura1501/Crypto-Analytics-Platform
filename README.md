# Crypto Trading Analytics Platform

## Mục đích dự án

Hệ thống phân tích tài chính tiền điện tử (cryptocurrency) kết hợp:
- **Thu thập tin tức tự động**: Crawl tin tức từ nhiều nguồn khác nhau
- **Biểu đồ giá realtime**: Hiển thị giá theo thời gian thực như TradingView
- **Phân tích AI**: Sử dụng AI để phân tích sentiment và dự đoán xu hướng
- **Quản lý tài khoản**: Theo dõi portfolio và giao dịch

## Tech Stack

- **Backend API**: Golang (Gin framework)
- **Frontend**: Vite + React/TypeScript
- **Python Services**: Crawler + AI Analysis
- **Database**: PostgreSQL
- **Realtime**: WebSocket cho giá realtime

## Kiến trúc hệ thống

📊 **Tài liệu kiến trúc:**

- **[ARCHITECTURE.md](./ARCHITECTURE.md)** - Kiến trúc ban đầu với Mermaid diagrams
- **[SITUATION_ANALYSIS.md](./SITUATION_ANALYSIS.md)** - Phân tích chi tiết 3 tình huống và giải pháp kỹ thuật
- **[ARCHITECTURE_DIAGRAMS.md](./ARCHITECTURE_DIAGRAMS.md)** - Sơ đồ kiến trúc sau khi nâng cấp
- **[IMPLEMENTATION_GUIDE.md](./IMPLEMENTATION_GUIDE.md)** - Hướng dẫn sử dụng các tính năng mới
- **[SUMMARY.md](./SUMMARY.md)** - Tổng kết implementation

### 🎯 Giải pháp cho 3 tình huống đồ án:

#### 1. ✅ WebSocket Scaling (Redis Pub/Sub)
- Scale từ 5,000 → 15,000+ concurrent connections
- Sticky session với IP Hash
- Zero downtime deployment

#### 2. ✅ AI-Enhanced Crawler (LLM Parser)
- Tự động fallback từ CSS → LLM khi cần
- Success rate: 60% → 90%+
- Chi phí tối ưu: $150/month

#### 3. ✅ Security Architecture (JWT + Rate Limiting)
- JWT Authentication với refresh token
- IP-based rate limiting (100 req/min)
- Audit logging đầy đủ

**Chi tiết:** Xem [SITUATION_ANALYSIS.md](./SITUATION_ANALYSIS.md)

## Cấu trúc dự án

```
DA/
├── backend/          # Golang API server
├── frontend/         # Vite React app
├── services/
│   ├── crawler/      # Python news crawler
│   └── ai-service/   # Python AI analysis
├── docker-compose.yml
├── README.md
├── ARCHITECTURE.md   # Kiến trúc hệ thống (Mermaid diagrams)
├── SETUP.md          # Hướng dẫn setup chi tiết
└── QUICKSTART.md     # Hướng dẫn nhanh
```

## Cặp tiền (Trading Pairs)

**BTCUSDT**: Bitcoin vs USDT (Tether)
- BTC là đồng tiền cơ sở (base currency)
- USDT là đồng tiền định giá (quote currency)
- Giá BTCUSDT = 50,000 có nghĩa là 1 BTC = 50,000 USDT

Các cặp tiền phổ biến khác: ETHUSDT, BNBUSDT, SOLUSDT, v.v.

## Các nguồn tin tức đề xuất

1. **CoinTelegraph** - https://cointelegraph.com
2. **CoinDesk** - https://www.coindesk.com
3. **Binance News** - https://www.binance.com/en/blog
4. **TradingView News** - https://www.tradingview.com/news/
5. **CryptoCompare** - https://www.cryptocompare.com/news/
6. **Reddit** - r/CryptoCurrency, r/Bitcoin
7. **Twitter/X** - API để theo dõi influencers

## Cài đặt và chạy

### ⚡ Quick Start (Khuyến nghị)

Xem file [QUICKSTART.md](./QUICKSTART.md) để bắt đầu nhanh!

### Prerequisites
- Docker & Docker Compose (khuyến nghị)
- Hoặc cài đặt thủ công: Go 1.21+, Node.js 18+, Python 3.10+, PostgreSQL 15+, Redis 7+

### Chạy với Docker Compose
```bash
# Khởi động tất cả services
docker-compose up -d

# Xem logs
docker-compose logs -f

# Dừng services
docker-compose down
```

### Chạy development thủ công
Xem file [SETUP.md](./SETUP.md) để biết hướng dẫn chi tiết.

## API Endpoints

- `GET /api/v1/pairs` - Danh sách cặp tiền
- `GET /api/v1/price/:pair` - Giá hiện tại
- `WS /ws/price/:pair` - WebSocket giá realtime
- `GET /api/v1/news` - Danh sách tin tức
- `POST /api/v1/analyze` - Phân tích AI

