# Hướng dẫn cài đặt và chạy Crypto Analytics Platform

Hướng dẫn chi tiết cách cài đặt môi trường, cấu hình và chạy source code hệ thống phân tích tài chính AI thời gian thực.

---

## 1. Yêu cầu hệ thống

| Thành phần | Phiên bản |
|------------|-----------|
| **Docker** | 20.10+ |
| **Docker Compose** | 2.0+ |
| **Node.js** | 18+ (cho Frontend) |
| **npm** | 9+ |
| **RAM** | Tối thiểu 8GB (khuyến nghị 16GB cho AI Service) |

---

## 2. Chuẩn bị môi trường

### 2.1. Cài đặt Docker và Docker Compose

**Linux (Ubuntu/Debian):**
```bash
# Cài Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh

# Cài Docker Compose (nếu chưa có)
sudo apt-get update
sudo apt-get install docker-compose-plugin

# Kiểm tra
docker --version
docker compose version
```

**macOS / Windows:** Tải và cài [Docker Desktop](https://www.docker.com/products/docker-desktop/).

### 2.2. Cài đặt Node.js (cho Frontend)

```bash
# Ubuntu/Debian: dùng nvm hoặc NodeSource
curl -fsSL https://deb.nodesource.com/setup_20.x | sudo -E bash -
sudo apt-get install -y nodejs

# Kiểm tra
node --version  # v20.x
npm --version
```

---

## 3. Clone và cấu hình dự án

### 3.1. Clone repository

```bash
# HTTPS
git clone https://github.com/hykura1501/Crypto-Analytics-Platform.git
cd Crypto-Analytics-Platform

# Hoặc SSH
git clone git@github.com:hykura1501/Crypto-Analytics-Platform.git
cd Crypto-Analytics-Platform
```

### 3.2. Tạo file môi trường (.env) cho các service

Mỗi service cần file `.env` (copy từ `.env.example`):

**Auth Service:**
```bash
cp services/auth-service/.env.example services/auth-service/.env
```

**Crawler Service:**
```bash
cp services/crawler-service/.env.example services/crawler-service/.env
```

**Market Service:**
```bash
cp services/market-service/.env.example services/market-service/.env
```

**Frontend:**
Tạo file `frontend/.env` với nội dung:
```env
VITE_API_BASE_URL=http://localhost:5000
VITE_WS_BASE_URL=ws://localhost:5000
```
*(Nếu không có .env, frontend dùng mặc định localhost:5000)*

### 3.3. (Tùy chọn) Cấu hình Gemini AI

Nếu cần tính năng AI tự động học CSS selectors, thêm biến môi trường cho AI Service:

Tạo file `services/ai-service-python/.env` (hoặc thêm vào docker-compose):
```env
GEMINI_API_KEY=your-google-api-key
```

Lấy API key tại: [Google AI Studio](https://makersuite.google.com/app/apikey)

---

## 4. Chạy Backend (Docker Compose)

### 4.1. Khởi động tất cả services

```bash
docker compose up -d
```

Services sẽ được khởi động theo thứ tự: **PostgreSQL** → **Redis** → **Kafka** → **Traefik** → **Auth** → **Market** → **Crawler** → **AI Service**.

### 4.2. Kiểm tra trạng thái

```bash
docker compose ps
```

Tất cả services phải có trạng thái `running`. AI Service có thể mất 1–2 phút để khởi động (download models).

### 4.3. Xem logs

```bash
# Toàn bộ logs
docker compose logs -f

# Logs từng service
docker compose logs -f auth-service
docker compose logs -f market-service
docker compose logs -f crawler-service
docker compose logs -f ai-service
```

### 4.4. Các cổng sử dụng

| Service | Cổng | Mô tả |
|---------|------|-------|
| **Traefik (API Gateway)** | 5000 | REST API + WebSocket |
| **Traefik Dashboard** | 8080 | Giao diện quản lý |
| **PostgreSQL** | 5432 | Database |
| **Redis** | 6379 | Cache |
| **Kafka** | 29092 | Message broker |
| **AI Service** | 9001 | API AI trực tiếp (debug) |

---

## 5. Chạy Frontend

```bash
cd frontend
npm install
npm run dev
```

Frontend chạy tại **http://localhost:3000** (hoặc 5173 tùy cấu hình Vite).

**Lưu ý:** Frontend phải gọi API qua `http://localhost:5000` (Traefik). Đảm bảo `frontend/.env` có:
```env
VITE_API_BASE_URL=http://localhost:5000
VITE_WS_BASE_URL=ws://localhost:5000
```

---

## 6. Đăng ký tài khoản và đăng nhập

**Cách 1 – Đăng ký qua giao diện:**
- Mở **http://localhost:3000/register**
- Điền email, mật khẩu, họ tên và bấm đăng ký

**Cách 2 – Đăng ký qua API:**
```bash
curl -X POST http://localhost:5000/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"yourpassword","first_name":"Test","last_name":"User"}'
```

**Đăng nhập:** Mở **http://localhost:3000/login** và nhập email + password.

---

## 7. Một số lệnh hữu ích

| Lệnh | Mô tả |
|------|-------|
| `docker compose up -d` | Khởi động tất cả services |
| `docker compose down` | Dừng và xóa containers |
| `docker compose down -v` | Dừng và xóa cả volumes (mất dữ liệu) |
| `docker compose restart <service>` | Khởi động lại một service |
| `docker compose up -d --scale market-service=3` | Scale Market Service lên 3 instance |

---

## 8. Xử lý lỗi thường gặp

### 8.1. AI Service không khởi động / lỗi healthcheck

- AI Service cần tải models (FinBERT, PhoBERT) lần đầu (~2–3 phút).
- Kiểm tra logs: `docker compose logs ai-service`
- Nếu thiếu RAM, tăng cấu hình Docker (Settings → Resources).

### 8.2. Kafka chưa sẵn sàng

- Kafka cần ~40 giây để khởi động. Có thể chạy lại:
  ```bash
  docker compose restart crawler-service ai-service
  ```

### 8.3. Frontend không kết nối được API

- Kiểm tra Traefik: `curl http://localhost:5000/api/v1/auth/health` (nếu có endpoint health)
- Đảm bảo `VITE_API_BASE_URL=http://localhost:5000` trong `frontend/.env`

### 8.4. Port bị chiếm

Nếu port 5000, 5432, 6379 đã bị sử dụng, sửa trong `docker-compose.yml`:
```yaml
ports:
  - "5001:80"  # Thay 5000 bằng 5001
```

Và cập nhật `frontend/.env` tương ứng.

---

## 9. Cấu trúc thư mục dự án

```
crypto-analytics/
├── docker-compose.yml      # Định nghĩa toàn bộ stack
├── frontend/               # React + Vite frontend
│   ├── .env
│   └── src/
├── services/
│   ├── auth-service/       # Golang - JWT, RBAC
│   ├── market-service/     # Golang - WebSocket, Binance
│   ├── crawler-service/    # Golang - RSS, Kafka
│   └── ai-service-python/  # Python - Sentiment, XGBoost, SHAP
├── reports/                # LaTeX báo cáo
└── docs/                   # Tài liệu
```

---

## 10. Liên kết hữu ích

- **Repository:** https://github.com/hykura1501/Crypto-Analytics-Platform  
- **Traefik Dashboard:** http://localhost:8080  
- **Frontend:** http://localhost:3000 (sau khi chạy `npm run dev`)  
- **API Gateway:** http://localhost:5000  

---

*Crypto Analytics Platform — Hệ thống Phân tích Tài chính AI Thời gian Thực*
