# Hướng dẫn Setup và Chạy Dự án

## Yêu cầu hệ thống

- **Docker & Docker Compose** (khuyến nghị)
- Hoặc cài đặt thủ công:
  - Go 1.21+
  - Node.js 18+
  - Python 3.10+
  - PostgreSQL 15+
  - Redis 7+

## Cách 1: Chạy với Docker Compose (Khuyến nghị)

### Bước 1: Clone và vào thư mục dự án

```bash
cd /home/hykura/Desktop/Workspace/Study/KTPM/DA
```

### Bước 2: Tạo file .env cho backend

```bash
cd backend
cp .env.example .env
# Chỉnh sửa .env nếu cần
```

### Bước 3: Chạy với Docker Compose

```bash
docker-compose up -d
```

Lệnh này sẽ tự động:
- Khởi động PostgreSQL
- Khởi động Redis
- Khởi động Backend API (port 8080)
- Khởi động Python Crawler Service
- Khởi động Python AI Service (port 5000)

### Bước 4: Chạy Frontend

Mở terminal mới:

```bash
cd frontend
npm install
npm run dev
```

Frontend sẽ chạy tại: http://localhost:3000

## Cách 2: Chạy thủ công (Development)

### Bước 1: Setup Database

```bash
# Tạo database
createdb cryptodb

# Hoặc dùng PostgreSQL CLI
psql -U postgres
CREATE DATABASE cryptodb;
CREATE USER crypto_user WITH PASSWORD 'crypto_pass';
GRANT ALL PRIVILEGES ON DATABASE cryptodb TO crypto_user;
\q
```

### Bước 2: Setup Redis

```bash
# Ubuntu/Debian
sudo apt-get install redis-server
redis-server

# Hoặc dùng Docker
docker run -d -p 6379:6379 redis:7-alpine
```

### Bước 3: Chạy Backend

```bash
cd backend
go mod download
go run main.go
```

Backend sẽ chạy tại: http://localhost:8080

### Bước 4: Chạy Python Crawler

```bash
cd services/crawler
python -m venv venv
source venv/bin/activate  # Trên Windows: venv\Scripts\activate
pip install -r requirements.txt
python main.py --all  # Crawl một lần
# hoặc
python main.py  # Chạy scheduler tự động
```

### Bước 5: Chạy AI Service

```bash
cd services/ai-service
python -m venv venv
source venv/bin/activate
pip install -r requirements.txt
python main.py
```

AI Service sẽ chạy tại: http://localhost:5000

### Bước 6: Chạy Frontend

```bash
cd frontend
npm install
npm run dev
```

## Cấu trúc dự án

```
DA/
├── backend/              # Golang API Server
│   ├── main.go
│   ├── internal/
│   │   ├── database/     # Database connection & migrations
│   │   ├── handlers/     # HTTP & WebSocket handlers
│   │   ├── models/       # Data models
│   │   ├── services/     # External services (Binance)
│   │   └── websocket/    # WebSocket hub
│   └── go.mod
│
├── frontend/             # React + Vite Frontend
│   ├── src/
│   │   ├── components/   # React components
│   │   ├── pages/        # Page components
│   │   ├── services/     # API clients
│   │   └── hooks/        # Custom hooks
│   └── package.json
│
├── services/
│   ├── crawler/          # Python News Crawler
│   │   └── main.py
│   └── ai-service/       # Python AI Analysis
│       └── main.py
│
└── docker-compose.yml    # Docker orchestration
```

## API Endpoints

### Trading Pairs
- `GET /api/v1/pairs` - Danh sách cặp tiền
- `GET /api/v1/pairs/:pair` - Thông tin cặp tiền cụ thể

### Price Data
- `GET /api/v1/price/:pair` - Giá hiện tại
- `GET /api/v1/price/:pair/history` - Lịch sử giá
- `GET /api/v1/klines/:pair` - Dữ liệu candlestick
- `WS /ws/price/:pair` - WebSocket cho giá realtime

### News
- `GET /api/v1/news` - Danh sách tin tức
- `GET /api/v1/news/:id` - Chi tiết tin tức
- `GET /api/v1/news/sources` - Danh sách nguồn tin

### AI Analysis
- `GET /api/v1/analysis/:pair` - Phân tích cho cặp tiền
- `POST /api/v1/analysis/predict` - Dự đoán xu hướng

## Cấu hình môi trường

### Backend (.env)

```env
PORT=8080
DATABASE_URL=postgres://crypto_user:crypto_pass@localhost:5432/cryptodb?sslmode=disable
REDIS_URL=redis://localhost:6379
```

### Python Services

Tạo file `.env` trong `services/crawler/` và `services/ai-service/`:

```env
DATABASE_URL=postgres://crypto_user:crypto_pass@localhost:5432/cryptodb?sslmode=disable
```

## Troubleshooting

### Database connection error
- Kiểm tra PostgreSQL đang chạy: `pg_isready`
- Kiểm tra credentials trong `.env`

### Redis connection error
- Kiểm tra Redis đang chạy: `redis-cli ping`
- Nếu dùng Docker, kiểm tra container: `docker ps`

### Python dependencies error
- Đảm bảo dùng Python 3.10+
- Cài đặt lại: `pip install -r requirements.txt --upgrade`

### Frontend build error
- Xóa node_modules và cài lại: `rm -rf node_modules && npm install`
- Kiểm tra Node.js version: `node --version` (cần >= 18)

## Next Steps

1. **Cấu hình nguồn tin tức**: Chỉnh sửa `news_sources` table để thêm nguồn mới
2. **Tích hợp Binance WebSocket**: Thêm WebSocket client để nhận giá realtime từ Binance
3. **Hoàn thiện AI models**: Tích hợp mô hình ML nâng cao hơn
4. **Thêm authentication**: Implement JWT cho user management
5. **Deploy**: Setup CI/CD và deploy lên production

## Tài liệu tham khảo

- [Binance API Docs](https://binance-docs.github.io/apidocs/)
- [TradingView Charting Library](https://github.com/tradingview/lightweight-charts)
- [Gin Framework](https://gin-gonic.com/docs/)
- [Vite Documentation](https://vitejs.dev/)

