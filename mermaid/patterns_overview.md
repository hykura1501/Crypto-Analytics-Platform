flowchart TB
    subgraph ClientLayer["Client Layer"]
        WebApp["Web Application"]
        MobileApp["Mobile App"]
    end

    subgraph Gateway["API Gateway"]
        APIGateway["API Gateway\n• Rate Limiting\n• Authentication\n• Routing"]
    end

    subgraph MarketServiceBox["Market Service"]
        subgraph AdapterPattern["ADAPTER Pattern"]
            MarketService["MarketService"]
            BinanceAdapter["BinanceAdapter"]
            CoinGeckoAdapter["CoinGeckoAdapter"]
            CoinbaseAdapter["CoinbaseAdapter"]
        end
        
        subgraph ObserverPattern["OBSERVER Pattern"]
            PricePublisher["PricePublisher"]
            DashboardObs["DashboardObserver"]
            AlertObs["AlertObserver"]
            DBObserver["DatabaseObserver"]
        end
    end

    subgraph CrawlerServiceBox["Crawler Service"]
        subgraph StrategyPattern["STRATEGY Pattern"]
            CrawlerContext["CrawlerContext"]
            RSSStrategy["RSSStrategy"]
            HTMLStrategy["HTMLStrategy"]
            APIStrategy["APIStrategy"]
        end
    end

    subgraph AIServiceBox["AI Service"]
        subgraph BuilderPattern["BUILDER Pattern"]
            PredictionDirector["PredictionDirector"]
            PredictionBuilder["PredictionModelBuilder"]
            PredictionModel["PredictionModel"]
        end
    end

    subgraph TradingServiceBox["Trading Service"]
        subgraph DecoratorPattern["DECORATOR Pattern"]
            BaseFee["BaseFeeCalculator"]
            PeakHourDec["PeakHourDecorator"]
            VolatilityDec["VolatilityDecorator"]
            VIPDec["VIPDiscountDecorator"]
        end
    end

    subgraph Infrastructure["Infrastructure"]
        subgraph SingletonPattern["SINGLETON Pattern"]
            ConfigMgr["ConfigManager"]
            WSHub["WebSocketHub"]
            DBPool["DatabasePool"]
        end
    end

    subgraph ExternalSystems["External Systems"]
        Binance["Binance API"]
        CoinGecko["CoinGecko API"]
        NewsFeeds["News RSS Feeds"]
    end

    subgraph Storage["Storage"]
        PostgreSQL[("PostgreSQL")]
        Redis[("Redis Cache")]
        Kafka["Kafka Topics"]
    end

    ClientLayer --> Gateway
    Gateway --> MarketServiceBox
    Gateway --> CrawlerServiceBox
    Gateway --> AIServiceBox
    Gateway --> TradingServiceBox

    BinanceAdapter --> Binance
    CoinGeckoAdapter --> CoinGecko
    CrawlerContext --> NewsFeeds

    MarketServiceBox --> Storage
    CrawlerServiceBox --> Storage
    AIServiceBox --> Storage
    TradingServiceBox --> Storage

    Infrastructure --> MarketServiceBox
    Infrastructure --> CrawlerServiceBox
    Infrastructure --> AIServiceBox
    Infrastructure --> TradingServiceBox

    style AdapterPattern fill:#e3f2fd
    style ObserverPattern fill:#e8f5e9
    style StrategyPattern fill:#fff3e0
    style BuilderPattern fill:#fce4ec
    style DecoratorPattern fill:#f3e5f5
    style SingletonPattern fill:#e0f7fa
