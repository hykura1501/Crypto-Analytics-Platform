# Luồng hoạt động Crawler - Sơ đồ chi tiết

## 🔄 Luồng tổng quan

```mermaid
flowchart TD
    Start([Start Crawler Service]) --> Init[Khởi tạo NewsCrawler]
    Init --> WaitDB{Database<br/>sẵn sàng?}
    WaitDB -->|Chưa| Retry[Retry với<br/>exponential backoff]
    Retry --> WaitDB
    WaitDB -->|Sẵn sàng| Scheduler[Scheduler Start]
    
    Scheduler --> CrawlNow[Crawl ngay lập tức]
    CrawlNow --> GetSources[Lấy danh sách<br/>active sources]
    
    GetSources --> LoopSources{ còn<br/>source? }
    LoopSources -->|Có| GetSourceInfo[Lấy thông tin source<br/>từ database]
    
    GetSourceInfo --> FetchHomepage[Fetch homepage]
    FetchHomepage --> ParseHTML[Parse HTML với<br/>BeautifulSoup]
    ParseHTML --> FindLinks[Tìm article links<br/>bằng link_selector]
    
    FindLinks --> LoopArticles{ còn<br/>article? }
    LoopArticles -->|Có| CheckExist{URL đã<br/>tồn tại?}
    
    CheckExist -->|Có| Skip[Skip article]
    CheckExist -->|Chưa| FetchArticle[Fetch nội dung<br/>article]
    
    FetchArticle --> Extract[Extract:<br/>- Title<br/>- Content<br/>- Date]
    Extract --> SaveDB[Lưu vào database<br/>INSERT news]
    SaveDB --> Sleep1[Sleep 1 giây]
    Skip --> LoopArticles
    Sleep1 --> LoopArticles
    
    LoopArticles -->|Hết| UpdateTimestamp[Cập nhật<br/>last_crawled_at]
    UpdateTimestamp --> Sleep5[Sleep 5 giây]
    Sleep5 --> LoopSources
    
    LoopSources -->|Hết| Wait30[Đợi 30 phút]
    Wait30 --> CrawlNow
    
    style Start fill:#4CAF50
    style Scheduler fill:#2196F3
    style SaveDB fill:#FF9800
    style Wait30 fill:#9C27B0
```

## 📋 Quy trình chi tiết từng bước

### 1. Khởi tạo Service

```mermaid
sequenceDiagram
    participant Main
    participant Crawler
    participant Env as Environment
    participant DB as Database

    Main->>Crawler: NewsCrawler()
    Crawler->>Env: getenv('DATABASE_URL')
    Env-->>Crawler: postgres://.../cryptodb
    Crawler->>Crawler: Validate & fix DB URL
    Crawler->>Crawler: Create HTTP session
    Note over Crawler: User-Agent: Mozilla/5.0...
```

### 2. Kết nối Database với Retry

```mermaid
sequenceDiagram
    participant Crawler
    participant DB as PostgreSQL

    Crawler->>DB: psycopg2.connect()
    alt Connection Success
        DB-->>Crawler: Connection OK
        Crawler->>DB: SELECT 1 (test)
        DB-->>Crawler: Success
    else Connection Failed
        DB-->>Crawler: OperationalError
        Crawler->>Crawler: Wait 2 seconds
        Crawler->>DB: Retry connect
        alt Still Failed
            DB-->>Crawler: Error
            Crawler->>Crawler: Wait 4 seconds (2^2)
            Crawler->>DB: Retry again
            Note over Crawler: Exponential backoff: 2, 4, 8, 16s
        end
    end
```

### 3. Crawl một Source

```mermaid
sequenceDiagram
    participant Scheduler
    participant Crawler
    participant DB as Database
    participant Website as News Website

    Scheduler->>Crawler: crawl_source(source_id)
    Crawler->>DB: SELECT source info
    DB-->>Crawler: name, url, selectors
    
    Crawler->>Website: GET homepage
    Website-->>Crawler: HTML content
    Crawler->>Crawler: Parse HTML
    Crawler->>Crawler: Find article links (max 20)
    
    loop For each article link
        Crawler->>DB: Check URL exists
        DB-->>Crawler: Not found
        Crawler->>Website: GET article URL
        Website-->>Crawler: Article HTML
        Crawler->>Crawler: Extract title, content, date
        Crawler->>DB: INSERT news
        DB-->>Crawler: Success
        Crawler->>Crawler: Sleep 1 second
    end
    
    Crawler->>DB: UPDATE last_crawled_at
```

### 4. Scheduler Cycle

```mermaid
stateDiagram-v2
    [*] --> Initialize: Service Start
    Initialize --> FirstCrawl: Database Ready
    FirstCrawl --> CrawlingAll: Get Sources
    CrawlingAll --> Processing: For each source
    Processing --> Fetching: Get homepage
    Fetching --> Extracting: Parse HTML
    Extracting --> Saving: Extract articles
    Saving --> Waiting: Save to DB
    Waiting --> Processing: Next source
    Processing --> Scheduled: All done
    Scheduled --> Waiting30min: Schedule next run
    Waiting30min --> CrawlingAll: 30 minutes later
    
    Fetching --> Error: Network error
    Extracting --> Error: Parse error
    Saving --> Error: DB error
    Error --> Processing: Continue with next
```

## 🎯 Các điểm then chốt

### 1. Retry Mechanism

```
Attempt 1: ──────────────> [Fail] ──┐
                                    │
Attempt 2: ────wait 2s────> [Fail] ─┤
                                    │
Attempt 3: ────wait 4s────> [Fail] ─┤ Exponential
                                    │ Backoff
Attempt 4: ────wait 8s────> [Fail] ─┤
                                    │
Attempt 5: ────wait 16s───> [Success] ✅
```

### 2. Duplicate Prevention Flow

```
Article URL
    │
    ▼
Check Database: SELECT id FROM news WHERE url = ?
    │
    ├───[Exists]───> Skip ❌
    │
    └───[Not Found]───> Fetch & Save ✅
```

### 3. Rate Limiting

```
Source 1: [Article 1] ──1s──> [Article 2] ──1s──> [Article 3]
    │
    └───5s───> Source 2: [Article 1] ──1s──> [Article 2]
                    │
                    └───5s───> Source 3: ...
```

## 📊 Dữ liệu flow

```
┌─────────────────────────────────────────────────────────────┐
│                    News Source (Database)                    │
│  id, name, url, title_selector, content_selector, ...       │
└──────────────────────┬──────────────────────────────────────┘
                       │
                       ▼
┌─────────────────────────────────────────────────────────────┐
│                    Homepage HTML                             │
│  <html>                                                      │
│    <a href="/article1">Title 1</a>                          │
│    <a href="/article2">Title 2</a>                          │
│  </html>                                                     │
└──────────────────────┬──────────────────────────────────────┘
                       │
                       ▼
┌─────────────────────────────────────────────────────────────┐
│                Extracted Article Links                       │
│  [                                                           │
│    {url: ".../article1", title: "Title 1"},                 │
│    {url: ".../article2", title: "Title 2"}                  │
│  ]                                                           │
└──────────────────────┬──────────────────────────────────────┘
                       │
                       ▼ (For each article)
┌─────────────────────────────────────────────────────────────┐
│                  Article HTML                                │
│  <h1>Article Title</h1>                                     │
│  <div class="content">Article content...</div>              │
│  <time>2025-11-25</time>                                    │
└──────────────────────┬──────────────────────────────────────┘
                       │
                       ▼
┌─────────────────────────────────────────────────────────────┐
│              Extracted Data                                  │
│  {                                                            │
│    title: "Article Title",                                   │
│    content: "Article content...",                            │
│    url: ".../article1",                                      │
│    published_at: "2025-11-25 10:00:00"                      │
│  }                                                            │
└──────────────────────┬──────────────────────────────────────┘
                       │
                       ▼
┌─────────────────────────────────────────────────────────────┐
│                    News Table (Database)                     │
│  INSERT INTO news (...) VALUES (...)                        │
└─────────────────────────────────────────────────────────────┘
```

## ⏱️ Timeline thực tế

```
00:00 ──────── Start crawler service
00:01 ──────── Database connected
00:02 ──────── Start first crawl
00:03 ──────── Crawling Source 1 (CoinTelegraph)
00:05 ──────── Found 20 articles
00:06 ──────── Processing article 1/20
00:07 ──────── Processing article 2/20
      ...
00:25 ──────── Finished Source 1 (20 articles)
00:30 ──────── Start Source 2 (CoinDesk)
00:32 ──────── Processing articles...
00:50 ──────── Finished all sources
30:00 ──────── Next scheduled crawl (30 minutes later)
```

## 🛡️ Error Handling Strategy

```
┌─────────────────┐
│   Try Action    │
└────────┬────────┘
         │
         ▼
    ┌─────────┐
    │ Success?│
    └────┬────┘
         │
    ┌────┴────┐
    │         │
   Yes       No
    │         │
    ▼         ▼
  Continue  Log Error
    │         │
    │         ▼
    │    Continue Next
    │    (Don't crash)
    │
    ▼
  Next Step
```

## 💡 Best Practices được áp dụng

1. ✅ **Exponential Backoff** - Tránh spam khi retry
2. ✅ **Rate Limiting** - Lịch sự với server
3. ✅ **Transaction Safety** - Rollback khi lỗi
4. ✅ **Duplicate Prevention** - Check trước khi insert
5. ✅ **Error Isolation** - Một article lỗi không ảnh hưởng toàn bộ
6. ✅ **Resource Cleanup** - Luôn close connection
7. ✅ **Timeout Protection** - 30s timeout cho HTTP requests


