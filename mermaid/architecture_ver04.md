---
title: "Version 4: Microservices Event-Driven (CURRENT IMPLEMENTATION)"
description: "4 Services + Kafka + Docker Compose (10,000-100,000 users)"
---

flowchart TB
    subgraph Client["👤 Client Layer"]
        Frontend["React Frontend<br/><small>Vite + TypeScript + TailwindCSS<br/>Port 5173</small>"]
    end

    subgraph Gateway["🚪 API Gateway (Traefik)"]
        Traefik["Traefik Reverse Proxy<br/><small>Port 5000 (HTTP)<br/>Port 8080 (Dashboard)</small><br/>Routing + CORS + Health Checks"]
    end

    subgraph Services["🔧 Microservices Layer (Docker Compose)"]
        subgraph AuthSvc["🔐 Auth Service"]
            AuthAPI["Golang + Gin<br/><small>Port 8081</small><br/>JWT + Refresh Token<br/>Role-based Access"]
        end
        
        subgraph MarketSvc["📊 Market Service"]
            MarketAPI["Golang + Gin<br/><small>Port 8082</small><br/>REST API + WebSocket<br/>Hub Pattern"]
        end
        
        subgraph CrawlerSvc["🕷️ Crawler Service"]
            CrawlerAPI["Golang + Colly<br/><small>Port 8083</small><br/>RSS + HTML Parsing<br/>Concurrent Crawling"]
        end
        
        subgraph AISvc["🤖 AI Service"]
            AIAPI["Python + Flask<br/><small>Port 9001</small><br/>FinBERT + PhoBERT<br/>XGBoost + SHAP"]
        end
    end

    subgraph MessageBroker["📨 Message Broker (Kafka)"]
        Kafka["Apache Kafka 3.7.0<br/><small>KRaft mode (No Zookeeper)<br/>Port 9092</small>"]
        Topics["<b>Topics:</b><br/>• news_new_article<br/>• news_analyze_css_selector<br/>• news_analyze_rss_structure"]
    end

    subgraph DataLayer["💾 Data Layer"]
        Postgres[("PostgreSQL 15<br/><small>Port 5432</small><br/>users, articles, prices<br/>predictions, refresh_tokens")]
        RedisCache[("Redis 7<br/><small>Port 6379</small><br/>Cache + Sessions<br/>Token Blacklist")]
    end

    subgraph External["🌍 External Services"]
        Binance["Binance API<br/><small>WebSocket + REST</small>"]
        NewsRSS["News Sources<br/><small>CoinDesk, VNExpress<br/>CoinTelegraph, VnEconomy</small>"]
        GeminiAI["Google Gemini AI<br/><small>CSS Selector Analysis</small>"]
    end

    Frontend <-->|"HTTP/WebSocket"| Traefik
    
    Traefik -->|"/api/v1/auth/*"| AuthAPI
    Traefik -->|"/api/v1/market/*"| MarketAPI
    Traefik -->|"/ws/*"| MarketAPI
    Traefik -->|"/api/v1/news/*"| CrawlerAPI
    Traefik -->|"/api/v1/ai/*"| AIAPI
    
    AuthAPI <--> Postgres
    AuthAPI <--> RedisCache
    
    MarketAPI <--> Postgres
    MarketAPI <--> Binance
    
    CrawlerAPI -->|"Produce"| Kafka
    CrawlerAPI <--> Postgres
    CrawlerAPI --> NewsRSS
    
    Kafka -->|"Consume"| AIAPI
    AIAPI <--> Postgres
    AIAPI --> GeminiAI
    
    Topics -.->|"Part of"| Kafka
    
    note1[/"<b>🎯 CURRENT VERSION</b><br/>Deployed via docker-compose.yml<br/>4 microservices + Kafka<br/>Production-ready for demo"/]
