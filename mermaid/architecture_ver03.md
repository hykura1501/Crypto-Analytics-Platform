---
title: "Version 3: Horizontal Scaling"
description: "Load Balancer + Multiple Servers + Workers (1,000-10,000 users)"
---

flowchart LR
    Client["👤 Client<br/>Browsers"]
    
    LB["⚖️ Load Balancer<br/>Nginx/HAProxy"]
    
    subgraph AppTier["🖥️ Application Servers"]
        App1["App Server 1<br/>API + WebSocket"]
        App2["App Server 2<br/>API + WebSocket"]
        App3["App Server 3<br/>API + WebSocket"]
    end
    
    Queue["📨 Message Queue<br/>RabbitMQ/Kafka<br/>crawl_jobs<br/>ai_analysis_jobs"]
    
    subgraph WorkerTier["⚙️ Background Workers"]
        Crawler1["🕷️ Crawler 1"]
        Crawler2["🕷️ Crawler 2"]
        AI1["🤖 AI Worker 1"]
        AI2["🤖 AI Worker 2"]
    end
    
    Redis["⚡ Redis Cache<br/>Prices, News<br/>Sessions"]
    
    subgraph Database["💾 Database Cluster"]
        DBPrimary[("Primary<br/>Writes")]
        DBReplica1[("Replica 1<br/>Reads")]
        DBReplica2[("Replica 2<br/>Reads")]
    end
    
    Storage[("☁️ Object Storage<br/>HTML, Models")]
    
    Binance["🌍 Binance API"]
    News["🌍 News Sites"]
    
    Client --> LB
    LB --> App1
    LB --> App2
    LB --> App3
    
    App1 --> Redis
    App2 --> Redis
    App3 --> Redis
    
    App1 --> DBReplica1
    App2 --> DBReplica2
    App3 --> DBReplica1
    
    App1 --> Queue
    App2 --> Queue
    App3 --> Queue
    
    App1 --> Binance
    
    Queue --> Crawler1
    Queue --> Crawler2
    Queue --> AI1
    Queue --> AI2
    
    Crawler1 --> DBPrimary
    Crawler2 --> DBPrimary
    Crawler1 --> Storage
    Crawler2 --> Storage
    Crawler1 --> News
    Crawler2 --> News
    
    AI1 --> DBPrimary
    AI2 --> DBPrimary
    
    DBPrimary -.->|Replication| DBReplica1
    DBPrimary -.->|Replication| DBReplica2
