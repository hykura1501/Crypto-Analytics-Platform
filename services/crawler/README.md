# News Crawler Service - Cơ chế hoạt động

## 📋 Tổng quan

Crawler service là một hệ thống tự động thu thập tin tức từ nhiều nguồn khác nhau, lưu vào database và có khả năng tự học cấu trúc HTML của từng trang web.

## 🏗️ Cấu trúc chính

### 1. Class NewsCrawler

```python
class NewsCrawler:
    - db_url: Connection string đến PostgreSQL
    - session: HTTP session với headers giả lập browser
```

## 🔄 Luồng hoạt động chi tiết

### Bước 1: Khởi tạo (Initialization)

```python
def __init__(self):
```

**Nhiệm vụ:**
1. ✅ Lấy `DATABASE_URL` từ environment variable
2. ✅ Đảm bảo database name đúng (`cryptodb`)
3. ✅ Thêm `?sslmode=disable` nếu chưa có
4. ✅ Tạo HTTP session với User-Agent giả lập browser
   - Giúp tránh bị chặn bởi các website chống bot

### Bước 2: Kết nối Database với Retry Logic

```python
def get_db_connection(self, max_retries=5):
```

**Cơ chế:**
- 🔁 **Retry với Exponential Backoff:**
  - Lần 1: retry ngay
  - Lần 2: đợi 2 giây (2^1)
  - Lần 3: đợi 4 giây (2^2)
  - Lần 4: đợi 8 giây (2^3)
  - Lần 5: đợi 16 giây (2^4)

**Lý do:** Database có thể chưa sẵn sàng ngay khi container start, cần đợi migrations chạy xong.

### Bước 3: Học cấu trúc HTML (Structure Learning)

```python
def learn_structure(self, html_content, sample_data):
```

**Mục đích:** Tự động phát hiện các selector phù hợp để trích xuất:
- **Title selectors:** `h1`, `h2`, `h3`, `[class*="title"]`
- **Content selectors:** `article`, `[class*="content"]`
- **Link selectors:** `a[href]`, `h2 a`

**Cách hoạt động:**
1. Parse HTML bằng BeautifulSoup
2. Thử các selector pattern phổ biến
3. Trả về danh sách selector có thể dùng

> ⚠️ **Lưu ý:** Function này hiện tại chỉ trả về các pattern mặc định. Có thể mở rộng để tự động phân tích HTML và tìm selector tốt nhất.

### Bước 4: Trích xuất tin tức (News Extraction)

```python
def extract_news(self, url, selectors):
```

**Quy trình:**
1. 📥 Fetch HTML từ URL
2. 🔍 Parse với BeautifulSoup
3. 📝 Extract:
   - Title (thử các selector cho đến khi tìm được)
   - Content (lấy nội dung chính)
   - Links (tìm các link bài viết)
4. ✅ Trả về dictionary với dữ liệu đã extract

**Xử lý lỗi:** 
- Try-catch để không crash khi một URL lỗi
- Return `None` nếu không extract được

### Bước 5: Crawl một nguồn cụ thể

```python
def crawl_source(self, source_id):
```

**Luồng hoạt động:**

```
┌─────────────────────────────────────────┐
│  1. Lấy thông tin source từ database    │
│     - name, url, selectors              │
└───────────────┬─────────────────────────┘
                │
                ▼
┌─────────────────────────────────────────┐
│  2. Fetch trang chủ (homepage)          │
│     - Parse HTML                        │
│     - Tìm các link bài viết             │
└───────────────┬─────────────────────────┘
                │
                ▼
┌─────────────────────────────────────────┐
│  3. Với mỗi article link:               │
│     a. Kiểm tra đã tồn tại chưa?        │
│     b. Fetch nội dung bài viết          │
│     c. Extract title, content, date     │
│     d. Lưu vào database                 │
└───────────────┬─────────────────────────┘
                │
                ▼
┌─────────────────────────────────────────┐
│  4. Cập nhật last_crawled_at            │
│  5. Commit transaction                  │
└─────────────────────────────────────────┘
```

**Chi tiết từng bước:**

#### 3a. Kiểm tra duplicate
```sql
SELECT id FROM news WHERE url = %s
```
- Tránh crawl lại bài đã có
- Sử dụng `ON CONFLICT (url) DO NOTHING` để an toàn

#### 3b. Fetch bài viết
- Timeout: 30 giây
- Parse với BeautifulSoup

#### 3c. Extract dữ liệu
- **Title:** Ưu tiên title từ selector, nếu không có thì dùng title từ link
- **Content:** Lấy 5000 ký tự đầu (limit để không quá dài)
- **Date:** Thử parse, nếu fail thì dùng `datetime.now()`

#### 3d. Lưu database
```sql
INSERT INTO news (source_id, title, content, url, published_at)
VALUES (%s, %s, %s, %s, %s)
ON CONFLICT (url) DO NOTHING
```

**Rate limiting:** 
- `time.sleep(1)` sau mỗi bài viết (để "lịch sự" với server)
- `time.sleep(5)` giữa các sources

### Bước 6: Crawl tất cả sources

```python
def crawl_all_sources(self):
```

**Cách hoạt động:**
1. 🔍 Query tất cả sources có `status = 'active'`
2. 🔄 Loop qua từng source
3. 📥 Gọi `crawl_source()` cho mỗi source
4. ⏸️ Đợi 5 giây giữa các sources

### Bước 7: Scheduler (Lập lịch)

```python
def run_scheduler(self):
```

**Cơ chế:**
- ⏰ Sử dụng thư viện `schedule`
- 🔄 Chạy mỗi 30 phút: `schedule.every(30).minutes`
- 🚀 Crawl ngay lập tức khi start
- ♾️ Chạy vòng lặp vô hạn:
  ```python
  while True:
      schedule.run_pending()
      time.sleep(60)  # Check mỗi phút
  ```

## 📊 Luồng dữ liệu tổng thể

```mermaid
sequenceDiagram
    participant Scheduler
    participant Crawler
    participant DB as PostgreSQL
    participant NewsSite as News Website

    Scheduler->>Crawler: run_scheduler()
    Crawler->>DB: Get active sources
    DB-->>Crawler: List of sources
    
    loop For each source
        Crawler->>NewsSite: GET homepage
        NewsSite-->>Crawler: HTML content
        Crawler->>Crawler: Parse & find article links
        
        loop For each article
            Crawler->>DB: Check if exists
            alt Not exists
                Crawler->>NewsSite: GET article URL
                NewsSite-->>Crawler: Article HTML
                Crawler->>Crawler: Extract title, content, date
                Crawler->>DB: INSERT news
            end
        end
        
        Crawler->>DB: UPDATE last_crawled_at
    end
    
    Note over Scheduler: Wait 30 minutes
    Scheduler->>Crawler: Run again
```

## 🔑 Các tính năng quan trọng

### 1. Retry Logic với Exponential Backoff
```python
wait_time = 2 ** attempt  # 2, 4, 8, 16 giây
```
- Tăng thời gian chờ sau mỗi lần thất bại
- Tránh spam database khi chưa sẵn sàng

### 2. Duplicate Prevention
- Check URL trước khi insert
- Dùng `ON CONFLICT` để an toàn
- Unique constraint trên `url` trong database

### 3. Rate Limiting
- 1 giây giữa các bài viết
- 5 giây giữa các sources
- Tránh bị chặn bởi server

### 4. Error Handling
- Try-catch ở mọi level
- Log lỗi nhưng không crash
- Rollback transaction nếu có lỗi

### 5. Connection Pooling (tiềm năng)
- Mỗi request tạo connection mới
- Có thể tối ưu bằng connection pool

## 🎯 Các điểm cần lưu ý

### ✅ Điểm mạnh
1. **Tự động retry** khi database chưa sẵn sàng
2. **Tránh duplicate** bằng URL check
3. **Rate limiting** để không quá tải server
4. **Error handling** tốt, không crash service
5. **Scheduled** tự động chạy định kỳ

### ⚠️ Cần cải thiện
1. **Connection pooling:** Hiện tại mỗi lần tạo connection mới
2. **Selector learning:** Function `learn_structure()` chưa được sử dụng thực sự
3. **Date parsing:** Cần cải thiện để parse nhiều format date hơn
4. **Content cleaning:** Có thể loại bỏ ads, footer, header
5. **Parallel crawling:** Có thể crawl nhiều sources song song

## 🚀 Cách sử dụng

### Chạy scheduler (tự động):
```bash
python main.py
```

### Crawl một lần tất cả:
```bash
python main.py --all
```

### Crawl một source cụ thể:
```bash
python main.py --source 1
```

## 📝 Database Schema liên quan

```sql
news_sources:
  - id
  - name
  - url
  - title_selector
  - content_selector
  - link_selector
  - status
  - last_crawled_at

news:
  - id
  - source_id (FK)
  - title
  - content
  - url (UNIQUE)
  - published_at
  - crawled_at
```

## 🔧 Cấu hình

### Environment Variables:
```env
DATABASE_URL=postgres://crypto_user:crypto_pass@postgres:5432/cryptodb?sslmode=disable
```

### Thêm source mới:
```sql
INSERT INTO news_sources (name, url, link_selector, title_selector, content_selector)
VALUES (
    'CoinTelegraph',
    'https://cointelegraph.com',
    'h2.post-card-inline__title a',
    'h2.post-card-inline__title a',
    'div.post-card-inline__text'
);
```

## 🎓 Kết luận

Crawler service là một hệ thống **robust, tự động và có khả năng mở rộng**:
- ✅ Tự động thu thập tin tức định kỳ
- ✅ Xử lý lỗi tốt
- ✅ Tránh duplicate
- ✅ Có thể mở rộng thêm nhiều sources

Để cải thiện thêm, có thể:
- Thêm parallel processing
- Cải thiện date parsing
- Tự động học selector tốt hơn
- Thêm monitoring và logging


