---
title: "Microservices Communication"
description: "Services interaction qua Kafka"
---

sequenceDiagram
    participant Client as Client
    participant Gateway as Traefik
    participant Auth as Auth Service
    participant Market as Market Service
    participant Crawler as Crawler Service
    participant Kafka as Kafka
    participant AI as AI Service
    participant DB as Database

    Note over Client,DB: User accesses system
    
    Client->>Gateway: Request with JWT
    Gateway->>Auth: Verify token
    Auth->>DB: Check user & role
    Auth-->>Gateway: Valid (Free/VIP)
    
    Note over Client,DB: Get market data
    
    Gateway->>Market: /market/prices
    Market->>DB: SELECT prices
    Market-->>Client: Price history
    
    Note over Client,DB: Background crawling
    
    Crawler->>Crawler: Scheduled crawl
    Crawler->>DB: INSERT articles
    Crawler->>Kafka: Produce news_new_article
    
    Note over Client,DB: Async AI processing
    
    Kafka->>AI: Consume message
    AI->>AI: Sentiment analysis
    AI->>DB: UPDATE article sentiment
    
    Note over Client,DB: VIP prediction
    
    Client->>Gateway: /ai/predict (VIP only)
    Gateway->>Auth: Check VIP role
    Auth-->>Gateway: Authorized
    Gateway->>AI: Forward request
    AI->>DB: Load data
    AI->>AI: XGBoost + SHAP
    AI-->>Client: Prediction + explanation
