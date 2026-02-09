---
title: "Version 2: Separated Data Layer"
description: "Tách Database + Object Storage (100-1,000 users)"
---

flowchart TB
    subgraph Client["👤 Client Layer"]
        Browser["Browser/Mobile App<br/><small>React Frontend</small>"]
    end

    subgraph Infrastructure["🏗️ Infrastructure"]
        subgraph AppLayer["Application Layer"]
            AppServer["🖥️ Application Server<br/>Monolith App<br/><small>Web + Crawler + AI</small>"]
        end
        
        subgraph DataLayer["Data Layer (Separated)"]
            Database[("💾 Managed Database<br/>PostgreSQL/MySQL<br/><small>users, news, prices, signals</small>")]
            ObjectStorage[("☁️ Object Storage<br/>S3-compatible/MinIO<br/><small>Raw HTML, Models, Logs</small>")]
        end
    end

    subgraph External["🌍 External Services"]
        Binance["Binance API<br/><small>Market Data</small>"]
        NewsSites["News Websites<br/><small>Content Sources</small>"]
    end

    Browser <-->|"HTTP/HTTPS<br/>Port 80/443"| AppServer
    
    AppServer -->|"SQL Queries<br/>Private Network"| Database
    AppServer -->|"File Upload/Download<br/>API"| ObjectStorage
    
    AppServer -->|"WebSocket/REST"| Binance
    AppServer -->|"HTTP Requests"| NewsSites
