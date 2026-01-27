flowchart TB
    subgraph GlobalCDN["Global CDN Layer"]
        CloudFront["CloudFront CDN<br/>Static Assets"]
    end

    subgraph Client["Client Layer"]
        WebApp["Web Application"]
        MobileApp["Mobile App"]
    end

    subgraph Gateway["API Gateway Layer"]
        APIGateway["API Gateway<br/>Kong/AWS API Gateway"]
    end

    subgraph Services["Microservices - Multi-AZ Auto Scaling"]
        subgraph MarketServiceASG["Market Data Service ASG"]
            Market1["Market Service 1"]
            Market2["Market Service 2"]
        end
        
        subgraph NewsServiceASG["News Service ASG"]
            News1["News Service 1"]
            News2["News Service 2"]
        end
        
        subgraph AIServiceASG["AI Service ASG - GPU"]
            AI1["AI Service 1"]
            AI2["AI Service 2"]
        end
        
        subgraph AccountServiceASG["Account Service ASG"]
            Account1["Account Service 1"]
            Account2["Account Service 2"]
        end
    end

    subgraph MessageBus["Message Bus - Kafka Cluster"]
        KafkaCluster["Kafka Cluster<br/>news_raw, news_parsed<br/>price_ticks, ai_signals"]
    end

    subgraph CacheLayer["Distributed Cache"]
        RedisCluster["Redis Cluster"]
    end

    subgraph DataLayer["Data Layer - Multi-AZ"]
        subgraph RDSAZ["RDS Multi-AZ"]
            RDSPrimary[("RDS Primary")]
            RDSStandby[("RDS Standby")]
            RDSReplica1[("Read Replica 1")]
        end
        
        TimeSeriesDB[("InfluxDB/TimescaleDB")]
        Elasticsearch[("Elasticsearch Cluster")]
    end

    subgraph DataLake["Data Warehouse and ML Platform"]
        S3DataLake[("S3 Data Lake")]
        Redshift[("Redshift/BigQuery")]
        FeatureStore[("Feature Store")]
        SageMaker["SageMaker"]
    end

    subgraph Observability["Observability Stack"]
        Prometheus["Prometheus"]
        Grafana["Grafana"]
        ELK["ELK Stack"]
        PagerDuty["PagerDuty"]
    end

    subgraph External["External Services"]
        Binance["Binance API"]
        NewsSites["News Websites"]
    end

    CloudFront --> WebApp
    WebApp --> APIGateway
    MobileApp --> APIGateway
    
    APIGateway --> MarketServiceASG
    APIGateway --> NewsServiceASG
    APIGateway --> AIServiceASG
    APIGateway --> AccountServiceASG
    
    MarketServiceASG --> KafkaCluster
    MarketServiceASG --> RedisCluster
    MarketServiceASG --> TimeSeriesDB
    MarketServiceASG --> Binance
    
    NewsServiceASG --> KafkaCluster
    NewsServiceASG --> RDSPrimary
    NewsServiceASG --> Elasticsearch
    NewsServiceASG --> NewsSites
    
    AIServiceASG --> KafkaCluster
    AIServiceASG --> RedisCluster
    AIServiceASG --> FeatureStore
    
    AccountServiceASG --> RDSPrimary
    AccountServiceASG --> RedisCluster
    
    RDSPrimary --> RDSStandby
    RDSPrimary --> RDSReplica1
    
    KafkaCluster --> S3DataLake
    S3DataLake --> Redshift
    S3DataLake --> SageMaker
    SageMaker --> FeatureStore
    
    MarketServiceASG --> Prometheus
    NewsServiceASG --> Prometheus
    AIServiceASG --> Prometheus
    AccountServiceASG --> Prometheus
    
    Prometheus --> Grafana
    Prometheus --> PagerDuty
    
    MarketServiceASG --> ELK
    NewsServiceASG --> ELK
    AIServiceASG --> ELK
    AccountServiceASG --> ELK
