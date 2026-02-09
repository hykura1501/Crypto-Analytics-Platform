# Kiến trúc Hệ thống - Crypto Analytics Platform

## Sơ đồ kiến trúc tổng quan

```mermaid
graph TB
    subgraph "Client Layer"
        FE[Frontend<br/>Vite + React + TypeScript<br/>Port: 3000]
    end

    subgraph "API Gateway Layer"
        API[Backend API<br/>Golang + Gin<br/>Port: 8080]
        WS[WebSocket Hub<br/>Realtime Price Updates]
    end

    subgraph "Business Logic Layer"
        CRAWLER[News Crawler Service<br/>Python<br/>Scheduled Jobs]
        AI[AI Analysis Service<br/>Python + Flask<br/>Port: 5000]
    end

    subgraph "Data Layer"
        PG[(PostgreSQL<br/>Primary Database)]
        REDIS[(Redis<br/>Cache Layer)]
    end

    subgraph "External Services"
        BINANCE[Binance API<br/>Price Data]
        NEWS_SOURCES[News Sources<br/>CoinTelegraph, CoinDesk, etc.]
    end

    FE -->|HTTP/REST| API
    FE -->|WebSocket| WS
    API -->|Read/Write| PG
    API -->|Cache| REDIS
    API -->|Fetch Price| BINANCE
    WS -->|Broadcast| FE
    
    CRAWLER -->|Crawl| NEWS_SOURCES
    CRAWLER -->|Save News| PG
    CRAWLER -->|Trigger Analysis| AI
    
    AI -->|Read News| PG
    AI -->|Read Price| PG
    AI -->|Save Analysis| PG
    AI -->|REST API| API

    style FE fill:#4CAF50
    style API fill:#2196F3
    style WS fill:#FF9800
    style CRAWLER fill:#9C27B0
    style AI fill:#F44336
    style PG fill:#607D8B
    style REDIS fill:#E91E63
    style BINANCE fill:#FFC107
    style NEWS_SOURCES fill:#00BCD4
```

## Luồng dữ liệu chi tiết

### 1. Luồng thu thập và phân tích tin tức

```mermaid
sequenceDiagram
    participant NS as News Sources
    participant C as Crawler Service
    participant DB as PostgreSQL
    participant AI as AI Service
    participant API as Backend API
    participant FE as Frontend

    Note over C: Chạy mỗi 30 phút
    C->>NS: Crawl tin tức
    NS-->>C: HTML Content
    C->>C: Parse & Extract
    C->>DB: Lưu news (chưa có sentiment)
    
    Note over AI: Phân tích sentiment
    AI->>DB: Lấy news chưa phân tích
    DB-->>AI: News list
    AI->>AI: Sentiment Analysis (VADER + TextBlob)
    AI->>DB: Cập nhật sentiment_score, sentiment_label
    
    Note over AI: Align với giá
    AI->>DB: Lấy price_history
    AI->>AI: Tính correlation
    AI->>DB: Lưu news_price_alignment
    
    API->>DB: GET /api/v1/news
    DB-->>API: News với sentiment
    API-->>FE: Display news feed
```

### 2. Luồng hiển thị giá realtime

```mermaid
sequenceDiagram
    participant FE as Frontend
    participant API as Backend API
    participant WS as WebSocket Hub
    participant BINANCE as Binance API
    participant REDIS as Redis Cache
    participant DB as PostgreSQL

    FE->>API: GET /api/v1/klines/BTCUSDT
    API->>REDIS: Check cache
    alt Cache hit
        REDIS-->>API: Return cached data
    else Cache miss
        API->>BINANCE: Fetch klines
        BINANCE-->>API: Historical data
        API->>REDIS: Cache data
        API->>DB: Save to price_history
    end
    API-->>FE: Klines data
    
    FE->>WS: WebSocket connect /ws/price/BTCUSDT
    WS-->>FE: Connection established
    
    loop Every second
        WS->>BINANCE: Fetch current price
        BINANCE-->>WS: Price update
        WS->>FE: Broadcast price update
        WS->>REDIS: Cache price
    end
```

### 3. Luồng phân tích và dự đoán AI

```mermaid
sequenceDiagram
    participant FE as Frontend
    participant API as Backend API
    participant AI as AI Service
    participant DB as PostgreSQL

    FE->>API: POST /api/v1/analysis/predict
    Note over API: pair: BTCUSDT, time_horizon: 24h
    
    API->>AI: POST /predict/1?time_horizon=24h
    AI->>DB: SELECT news với sentiment
    AI->>DB: SELECT price_history
    AI->>AI: Calculate trend & correlation
    AI->>AI: Generate prediction
    AI->>DB: INSERT ai_analysis
    AI-->>API: Prediction result
    
    API-->>FE: Display prediction
```

## Kiến trúc Database Schema

```mermaid
erDiagram
    TRADING_PAIRS ||--o{ PRICE_HISTORY : has
    TRADING_PAIRS ||--o{ AI_ANALYSIS : analyzes
    TRADING_PAIRS ||--o{ NEWS_PRICE_ALIGNMENT : referenced
    TRADING_PAIRS ||--o{ WATCHLISTS : "in watchlist"
    
    NEWS_SOURCES ||--o{ NEWS : generates
    NEWS ||--o{ NEWS_PRICE_ALIGNMENT : aligned_with
    
    USERS ||--o{ WATCHLISTS : has

    TRADING_PAIRS {
        int id PK
        string symbol UK
        string base_asset
        string quote_asset
        string status
        timestamp created_at
    }
    
    PRICE_HISTORY {
        int id PK
        int pair_id FK
        decimal price
        decimal volume
        decimal high
        decimal low
        timestamp timestamp
        string interval
    }
    
    NEWS_SOURCES {
        int id PK
        string name UK
        string url
        text title_selector
        text content_selector
        string status
        timestamp last_crawled_at
    }
    
    NEWS {
        int id PK
        int source_id FK
        string title
        text content
        string url UK
        timestamp published_at
        decimal sentiment_score
        string sentiment_label
    }
    
    NEWS_PRICE_ALIGNMENT {
        int id PK
        int news_id FK
        int pair_id FK
        decimal price_before
        decimal price_after
        decimal price_change_percent
        decimal correlation_score
    }
    
    AI_ANALYSIS {
        int id PK
        int pair_id FK
        string analysis_type
        string prediction
        decimal confidence_score
        text reasoning
        string time_horizon
        timestamp created_at
    }
    
    USERS {
        int id PK
        string username UK
        string email UK
        string password_hash
    }
    
    WATCHLISTS {
        int id PK
        int user_id FK
        int pair_id FK
        timestamp created_at
    }
```

## Component Interaction Diagram

```mermaid
graph LR
    subgraph "Frontend (React)"
        A[Dashboard]
        B[Chart Page]
        C[News Page]
        D[Account Page]
    end

    subgraph "Backend Services"
        E[HTTP Handlers]
        F[WebSocket Hub]
        G[Binance Service]
        H[Database Layer]
    end

    subgraph "Python Services"
        I[Crawler]
        J[AI Service]
    end

    subgraph "Data Stores"
        K[(PostgreSQL)]
        L[(Redis)]
    end

    A --> E
    B --> E
    B --> F
    C --> E
    D --> E
    
    E --> G
    E --> H
    E --> L
    F --> L
    
    H --> K
    I --> K
    J --> K
    
    G -->|API Calls| M[Binance API]
    I -->|HTTP Requests| N[News Websites]
    
    J -.->|REST API| E

    style A fill:#4CAF50
    style B fill:#4CAF50
    style C fill:#4CAF50
    style D fill:#4CAF50
    style E fill:#2196F3
    style F fill:#FF9800
    style I fill:#9C27B0
    style J fill:#F44336
```

## Deployment Architecture

```mermaid
graph TB
    subgraph "Docker Compose"
        subgraph "Container 1: Backend"
            API[Golang API Server<br/>Port 8080]
        end
        
        subgraph "Container 2: Frontend"
            FE[Vite Dev Server<br/>Port 3000]
        end
        
        subgraph "Container 3: PostgreSQL"
            PG[(PostgreSQL 15<br/>Port 5432)]
        end
        
        subgraph "Container 4: Redis"
            REDIS[(Redis 7<br/>Port 6379)]
        end
        
        subgraph "Container 5: Crawler"
            CRAWLER[Python Crawler<br/>Scheduler]
        end
        
        subgraph "Container 6: AI Service"
            AI[Python AI Service<br/>Port 5000]
        end
    end

    subgraph "External"
        BINANCE[Binance API]
        NEWS[News Websites]
    end

    FE --> API
    API --> PG
    API --> REDIS
    API --> BINANCE
    CRAWLER --> PG
    CRAWLER --> NEWS
    AI --> PG
    AI --> API

    style API fill:#2196F3
    style FE fill:#4CAF50
    style PG fill:#607D8B
    style REDIS fill:#E91E63
    style CRAWLER fill:#9C27B0
    style AI fill:#F44336
```

## Technology Stack

```mermaid
mindmap
  root((Crypto Analytics<br/>Platform))
    Frontend
      React
      TypeScript
      Vite
      Lightweight Charts
      Axios
    Backend
      Golang
      Gin Framework
      WebSocket
      Gorilla WebSocket
    Python Services
      Crawler
        BeautifulSoup
        Requests
        Schedule
      AI Service
        Flask
        VADER Sentiment
        TextBlob
        Transformers
    Database
      PostgreSQL
        Primary DB
        Migrations
      Redis
        Caching
        Session Store
    External APIs
      Binance API
        REST API
        WebSocket Streams
      News Sources
        CoinTelegraph
        CoinDesk
        Others
```

## API Endpoints Map

```mermaid
graph TD
    API["Backend API<br/>/api/v1"] --> PAIRS["GET /pairs"]
    API --> PRICE["GET /price/:pair"]
    API --> KLINES["GET /klines/:pair"]
    API --> WS["WS /ws/price/:pair"]
    API --> NEWS["GET /news"]
    API --> ANALYSIS["GET /analysis/:pair"]
    API --> AUTH["POST /auth/register<br/>POST /auth/login"]
    API --> ACCOUNT["GET /account/*"]
    
    PAIRS --> DB[(PostgreSQL)]
    PRICE --> BINANCE[Binance API]
    PRICE --> REDIS[(Redis Cache)]
    KLINES --> BINANCE
    WS --> HUB[WebSocket Hub]
    NEWS --> DB
    ANALYSIS --> DB
    ANALYSIS --> AI[AI Service]
    AUTH --> DB
    ACCOUNT --> DB
    
    style API fill:#2196F3
    style DB fill:#607D8B
    style REDIS fill:#E91E63
    style BINANCE fill:#FFC107
    style AI fill:#F44336
    style HUB fill:#FF9800
```

## Scalability Architecture

```mermaid
graph TB
    subgraph "Load Balancer"
        LB[Nginx/HAProxy]
    end

    subgraph "Frontend Cluster"
        FE1[Frontend Instance 1]
        FE2[Frontend Instance 2]
        FE3[Frontend Instance N]
    end

    subgraph "Backend API Cluster"
        API1[API Instance 1]
        API2[API Instance 2]
        API3[API Instance N]
    end

    subgraph "WebSocket Cluster"
        WS1[WS Hub 1]
        WS2[WS Hub 2]
        WS3[WS Hub N]
    end

    subgraph "Python Services"
        CRAWLER1[Crawler Instance 1]
        CRAWLER2[Crawler Instance 2]
        AI1[AI Service Instance 1]
        AI2[AI Service Instance 2]
    end

    subgraph "Database Cluster"
        PG_MASTER[(PostgreSQL<br/>Master)]
        PG_REPLICA1[(PostgreSQL<br/>Replica 1)]
        PG_REPLICA2[(PostgreSQL<br/>Replica 2)]
    end

    subgraph "Cache Cluster"
        REDIS1[(Redis<br/>Primary)]
        REDIS2[(Redis<br/>Replica)]
    end

    LB --> FE1
    LB --> FE2
    LB --> FE3
    
    FE1 --> API1
    FE2 --> API2
    FE3 --> API3
    
    API1 --> WS1
    API2 --> WS2
    API3 --> WS3
    
    API1 --> PG_MASTER
    API2 --> PG_REPLICA1
    API3 --> PG_REPLICA2
    
    API1 --> REDIS1
    API2 --> REDIS2
    
    CRAWLER1 --> PG_MASTER
    CRAWLER2 --> PG_MASTER
    AI1 --> PG_MASTER
    AI2 --> PG_REPLICA1
    
    PG_MASTER -.->|Replication| PG_REPLICA1
    PG_MASTER -.->|Replication| PG_REPLICA2
    REDIS1 -.->|Replication| REDIS2

    style LB fill:#FF5722
    style PG_MASTER fill:#607D8B
    style REDIS1 fill:#E91E63
```

