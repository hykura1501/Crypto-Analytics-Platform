flowchart TB
    subgraph Client["Client Layer"]
        WebApp["Web Application"]
        MobileApp["Mobile App"]
    end

    subgraph Gateway["API Gateway Layer"]
        APIGateway["API Gateway<br/>Nginx/Kong/Envoy<br/>AuthN/AuthZ Rate Limiting"]
    end

    subgraph Services["Microservices Layer"]
        subgraph MarketService["Market Data Service"]
            MarketAPI["REST API<br/>Price History"]
            MarketWS["WebSocket Server<br/>Realtime Price"]
            TimeSeriesDB[("Time-series DB")]
        end
        
        subgraph NewsService["News Service"]
            NewsAPI["REST API<br/>News CRUD"]
            CrawlerManager["Crawler Manager"]
            NewsDB[("PostgreSQL")]
            SearchEngine[("Elasticsearch")]
        end
        
        subgraph AIServiceMS["AI Service"]
            AIAPI["REST API<br/>Sentiment + Prediction"]
            ModelServer["Model Server"]
        end
        
        subgraph AccountService["Account Service"]
            AccountAPI["REST API<br/>User Auth"]
            AccountDB[("PostgreSQL")]
        end
    end

    subgraph MessageBus["Message Bus - Kafka"]
        TopicNewsRaw["Topic: news_raw"]
        TopicNewsParsed["Topic: news_parsed"]
        TopicPriceTicks["Topic: price_ticks"]
        TopicAISignals["Topic: ai_signals"]
    end

    subgraph External["External Services"]
        Binance["Binance API"]
        NewsSites["News Websites"]
    end

    WebApp --> APIGateway
    MobileApp --> APIGateway
    
    APIGateway -->|"/prices/*"| MarketAPI
    APIGateway -->|"/prices/ws"| MarketWS
    APIGateway -->|"/news/*"| NewsAPI
    APIGateway -->|"/ai/*"| AIAPI
    APIGateway -->|"/account/*"| AccountAPI
    
    MarketAPI --> TimeSeriesDB
    MarketWS --> TopicPriceTicks
    MarketWS --> Binance
    
    NewsAPI --> NewsDB
    NewsAPI --> SearchEngine
    CrawlerManager --> NewsSites
    CrawlerManager --> TopicNewsRaw
    
    TopicNewsRaw --> NewsAPI
    NewsAPI --> TopicNewsParsed
    TopicNewsParsed --> AIAPI
    
    AIAPI --> ModelServer
    AIAPI --> TopicAISignals
    
    AccountAPI --> AccountDB
    
    TopicAISignals --> MarketWS
