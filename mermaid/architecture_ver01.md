flowchart TB
    subgraph Client["Client"]
        Browser["Browser/Mobile App"]
    end

    subgraph VPC["VPC - Single EC2 Instance"]
        subgraph Monolith["Monolithic Application"]
            WebServer["Web Server<br/>REST API + WebSocket Server"]
            Crawler["Crawler Service<br/>Cron Job/Background Task"]
            AIService["AI Service<br/>Local Model .pt/.h5"]
            AccountModule["Account Module<br/>Auth + Profile + Watchlist"]
        end
        
        subgraph LocalDB["Local Database"]
            PostgreSQL[("PostgreSQL/MySQL<br/>users, news, prices,<br/>signals, logs")]
        end
        
        subgraph LocalStorage["Local Disk Storage"]
            Files[("Raw HTML<br/>AI Models<br/>Logs")]
        end
    end

    subgraph External["External Services"]
        Binance["Binance API<br/>REST + WebSocket"]
        NewsSites["News Sites<br/>HTTP Crawling"]
    end

    Browser <-->|"HTTP/WebSocket<br/>Port 80/443"| WebServer
    WebServer <--> Crawler
    WebServer <--> AIService
    WebServer <--> AccountModule
    
    Monolith --> PostgreSQL
    Monolith --> Files
    
    WebServer -->|"REST: History<br/>WS: Realtime"| Binance
    Crawler -->|"HTTP Parse HTML"| NewsSites
