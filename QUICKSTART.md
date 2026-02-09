# Hướng dẫn nhanh - Quick Start

## 🚀 Chạy nhanh với Docker

```bash
# 1. Khởi động tất cả services
docker-compose up -d

# 2. Chạy frontend (terminal mới)
cd frontend
npm install
npm run dev

# 3. Truy cập ứng dụng
# Frontend: http://localhost:3000
# Backend API: http://localhost:8080
# AI Service: http://localhost:5000
```

## 📋 Checklist sau khi khởi động

### 1. Kiểm tra Backend API
```bash
curl http://localhost:8080/api/v1/pairs
```

### 2. Kiểm tra Database
```bash
# Kết nối PostgreSQL
psql -U crypto_user -d cryptodb -h localhost

# Kiểm tra tables
\dt

# Xem trading pairs
SELECT * FROM trading_pairs;
```

### 3. Test Crawler
```bash
# Chạy crawler một lần
cd services/crawler
python main.py --all

# Hoặc chạy crawler cho một source cụ thể
python main.py --source 1
```

### 4. Test AI Service
```bash
# Phân tích sentiment
curl -X POST http://localhost:5000/analyze/sentiment

# Dự đoán xu hướng cho BTCUSDT
curl http://localhost:5000/predict/1?time_horizon=24h
```

## 🔧 Cấu hình cơ bản

### Thêm nguồn tin tức mới

1. Vào database:
```sql
INSERT INTO news_sources (name, url, title_selector, content_selector, link_selector)
VALUES (
    'CoinTelegraph',
    'https://cointelegraph.com',
    'h2.post-card-inline__title a',
    'div.post-card-inline__text',
    'h2.post-card-inline__title a'
);
```

2. Crawler sẽ tự động crawl từ nguồn này

### Thêm cặp tiền mới

```sql
INSERT INTO trading_pairs (symbol, base_asset, quote_asset)
VALUES ('ETHUSDT', 'ETH', 'USDT');
```

## 📊 Các tính năng chính

### 1. Thu thập tin tức tự động
- Crawler chạy mỗi 30 phút
- Tự động phân tích sentiment
- Lưu vào database

### 2. Biểu đồ giá realtime
- Hiển thị candlestick chart
- WebSocket realtime updates
- Nhiều khung thời gian (1m, 5m, 15m, 1h, 4h, 1d)

### 3. Phân tích AI
- Sentiment analysis cho tin tức
- Dự đoán xu hướng giá
- Align tin tức với giá lịch sử

### 4. Dashboard
- Xem tất cả cặp tiền
- Giá realtime
- Link đến chart chi tiết

## 🐛 Troubleshooting nhanh

**Backend không chạy:**
```bash
cd backend
go mod tidy
go run main.go
```

**Frontend lỗi:**
```bash
cd frontend
rm -rf node_modules package-lock.json
npm install
npm run dev
```

**Database connection error:**
- Kiểm tra PostgreSQL đang chạy: `pg_isready`
- Kiểm tra trong docker: `docker ps | grep postgres`

**Crawler không crawl được:**
- Kiểm tra internet connection
- Kiểm tra selectors trong database có đúng không
- Xem logs: `docker logs da-crawler-1`

## 📚 Tài liệu chi tiết

Xem file `SETUP.md` để biết hướng dẫn setup chi tiết hơn.

