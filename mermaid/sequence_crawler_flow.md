---
title: "News Crawler Flow"
description: "Thu thập tin tức với Kafka"
---

sequenceDiagram
    participant Crawler as Crawler Service
    participant RSS as News Sources
    participant Kafka as Kafka
    participant AI as AI Service
    participant DB as Database

    Note over Crawler,DB: Scheduled Crawl (30 mins)
    
    Crawler->>DB: Get active sources
    Crawler->>RSS: Fetch RSS feeds (concurrent)
    RSS-->>Crawler: XML articles
    
    loop Each article
        Crawler->>Crawler: Check if exists
        alt New article
            Crawler->>RSS: Fetch full HTML
            Crawler->>Crawler: Extract content
            Crawler->>DB: INSERT article
            Crawler->>Kafka: Produce news_new_article
        end
    end
    
    Note over Crawler,DB: AI Processing
    
    Kafka->>AI: Consume message
    AI->>AI: Detect language
    
    alt English
        AI->>AI: FinBERT sentiment
    else Vietnamese  
        AI->>AI: PhoBERT sentiment
    end
    
    AI->>AI: Extract keywords
    AI->>DB: UPDATE article<br/>SET sentiment, keywords
