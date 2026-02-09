# Sửa lỗi Docker Compose

## Các lỗi đã sửa:

### 1. Backend container không tìm thấy Go
**Lỗi:** `exec: "go": executable file not found in $PATH`

**Nguyên nhân:** Dockerfile build binary nhưng docker-compose.yml lại dùng `go run` (cần Go runtime)

**Đã sửa:**
- Bỏ `command: go run main.go` trong docker-compose.yml
- Backend sẽ dùng binary đã build từ Dockerfile

### 2. Database name không đúng
**Lỗi:** `database "crypto_user" does not exist`

**Nguyên nhân:** Database name là `cryptodb`, không phải `crypto_user`

**Đã sửa:** Đảm bảo DATABASE_URL đúng: `postgres://crypto_user:crypto_pass@postgres:5432/cryptodb`

## Cách sử dụng:

### Production (dùng binary đã build):
```bash
docker-compose up --build
```

### Development (hot reload với Go):
```bash
docker-compose -f docker-compose.dev.yml up
```

## Rebuild containers:

Nếu vẫn gặp lỗi, rebuild lại:
```bash
# Dừng và xóa containers
docker-compose down

# Xóa volumes (cẩn thận - sẽ mất data)
docker-compose down -v

# Rebuild và chạy
docker-compose up --build
```

## Kiểm tra:

1. **PostgreSQL:**
```bash
docker-compose exec postgres psql -U crypto_user -d cryptodb -c "\dt"
```

2. **Backend logs:**
```bash
docker-compose logs backend
```

3. **Redis:**
```bash
docker-compose exec redis redis-cli ping
```

