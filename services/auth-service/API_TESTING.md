# API Testing Guide

## Import vào Apidog/Postman

### Cách 1: Import Postman Collection
1. Mở Apidog/Postman
2. Click **Import**
3. Chọn file `Auth-Service.postman_collection.json`
4. Collection sẽ được import với tất cả endpoints

### Cách 2: Import OpenAPI Specification
1. Mở Apidog
2. Click **Import** → **OpenAPI**
3. Chọn file `openapi.yaml`
4. Apidog sẽ tự động tạo collection từ OpenAPI spec

## Environment Variables

Collection đã có sẵn các biến môi trường:

| Variable | Default Value | Description |
|----------|--------------|-------------|
| `base_url` | `http://localhost:8081` | Base URL của auth service |
| `user_email` | `test@example.com` | Email để test |
| `user_password` | `password123` | Password để test |
| `access_token` | (auto) | Access token (tự động lưu sau login) |
| `refresh_token` | (auto) | Refresh token (tự động lưu sau login) |

## Testing Flow

### 1. Health Check
```
GET {{base_url}}/health
```
Kiểm tra service có đang chạy không.

### 2. Register (Đăng ký)
```
POST {{base_url}}/api/v1/auth/register
Content-Type: application/json

{
  "email": "{{user_email}}",
  "password": "{{user_password}}",
  "first_name": "John",
  "last_name": "Doe"
}
```

### 3. Login (Đăng nhập)
```
POST {{base_url}}/api/v1/auth/login
Content-Type: application/json

{
  "email": "{{user_email}}",
  "password": "{{user_password}}"
}
```

**Response:**
```json
{
  "access_token": "eyJhbGc...",
  "refresh_token": "eyJhbGc...",
  "token_type": "Bearer",
  "expires_in": 900,
  "user": {
    "id": 1,
    "email": "test@example.com",
    "first_name": "John",
    "last_name": "Doe"
  }
}
```

Tokens sẽ **tự động được lưu** vào biến môi trường.

### 4. Validate Token (Xác thực token)
```
GET {{base_url}}/api/v1/auth/validate
Authorization: Bearer {{access_token}}
```

### 5. Get Current User (Lấy thông tin user)
```
GET {{base_url}}/api/v1/auth/me
Authorization: Bearer {{access_token}}
```

### 6. Refresh Token (Làm mới token)
```
POST {{base_url}}/api/v1/auth/refresh
Content-Type: application/json

{
  "refresh_token": "{{refresh_token}}"
}
```

Token mới sẽ **tự động thay thế** token cũ.

### 7. Logout (Đăng xuất)
```
POST {{base_url}}/api/v1/auth/logout
Content-Type: application/json

{
  "refresh_token": "{{refresh_token}}"
}
```

## Scripts Tự Động

Collection có sẵn **Test Scripts** để tự động:
- Lưu `access_token` sau khi login/refresh
- Lưu `refresh_token` sau khi login/refresh
- Log thông tin tokens ra console

## Quick Test với cURL

### Register
```bash
curl -X POST http://localhost:8081/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "password123",
    "first_name": "John",
    "last_name": "Doe"
  }'
```

### Login
```bash
curl -X POST http://localhost:8081/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "password123"
  }'
```

### Validate (thay YOUR_TOKEN bằng token thật)
```bash
curl -X GET http://localhost:8081/api/v1/auth/validate \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

## Error Responses

### 400 Bad Request
```json
{
  "error": "Invalid request",
  "message": "Email is required"
}
```

### 401 Unauthorized
```json
{
  "error": "Invalid credentials",
  "message": "Email or password is incorrect"
}
```

### 409 Conflict
```json
{
  "error": "User already exists",
  "message": "A user with this email already exists"
}
```

## Tips

1. **Chạy service trước**: Đảm bảo auth-service đang chạy (`go run cmd/main.go`)
2. **Đăng ký trước**: Phải register user trước khi login
3. **Token expiry**: Access token hết hạn sau 15 phút, dùng refresh token để lấy token mới
4. **CORS**: Service đã enable CORS cho tất cả origins

## Production Setup

Khi deploy production, đổi `base_url` thành:
- `http://your-api-gateway:8080` (qua API Gateway)
- hoặc `https://api.yourdomain.com`
