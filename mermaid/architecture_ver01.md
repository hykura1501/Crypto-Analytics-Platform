---
title: "Version 1: Single Box Monolith"
description: "MVP Architecture - All components on 1 server (10-100 users)"
---

flowchart TB
    subgraph Client["👤 Client Layer"]
        Browser["Browser/Mobile App<br/><small>React Frontend</small>"]
    end

    subgraph SingleServer["🖥️ Single Server (VPS/Local Machine)"]
        subgraph Monolith["Monolithic Application"]
            WebServer["🌐 Web Server<br/>REST API + WebSocket<br/><small>Express/Flask/Golang</small>"]
            Crawler["🕷️ Crawler Module<br/>Background Jobs<br/><small>Cron/Celery/Worker</small>"]
            AIService["🤖 AI Module<br/>Local Models<br/><small>.pt/.h5 files</small>"]
            AccountModule["👥 Account Module<br/>Auth + Profile<br/><small>JWT/Session</small>"]
        end
        
        subgraph LocalData["Local Data Storage"]
            PostgreSQL[("💾 PostgreSQL/MySQL<br/><small>users, news, prices,<br/>signals, logs</small>")]
            Files[("📁 Local Disk<br/><small>Raw HTML<br/>AI Models<br/>Log files</small>")]
        end
    end

    subgraph External["🌍 External Services"]
        Binance["Binance API<br/><small>REST + WebSocket</small>"]
        NewsSites["News Websites<br/><small>RSS/HTML Crawling</small>"]
    end

    Browser <-->|"HTTP/WebSocket<br/>Port 80/443"| WebServer
    WebServer <--> Crawler
    WebServer <--> AIService
    WebServer <--> AccountModule
    
    Monolith --> PostgreSQL
    Monolith --> Files
    
    WebServer -->|"Price Data"| Binance
    Crawler -->|"News Content"| NewsSites
