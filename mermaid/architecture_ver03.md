flowchart TB
    subgraph Client["Client Layer"]
        Browser["Browser/Mobile App"]
    end

    subgraph VPC["VPC"]
        subgraph PublicSubnet["Public Subnet"]
            ALB["Application Load Balancer<br/>HTTP/HTTPS + WebSocket"]
        end
        
        subgraph AppTier["App Tier - Auto Scaling Group"]
            App1["App Server 1<br/>API + WebSocket + GUI"]
            App2["App Server 2<br/>API + WebSocket + GUI"]
            App3["App Server N"]
        end
        
        subgraph WorkerTier["Worker Tier - Auto Scaling Group"]
            Crawler1["Crawler Worker 1"]
            Crawler2["Crawler Worker 2"]
            AI1["AI Worker 1"]
            AI2["AI Worker 2"]
        end
        
        subgraph MessageQueue["Message Queue"]
            SQS["SQS/Kafka<br/>News Fetch Queue<br/>AI Analysis Queue"]
        end
        
        subgraph CacheTier["Cache Layer"]
            Redis["Redis/ElastiCache<br/>Price Cache<br/>Top News Cache<br/>AI Signals Cache"]
        end
        
        subgraph DataTier["Data Tier"]
            RDSMaster[("RDS Master<br/>Write Operations")]
            RDSReplica1[("RDS Read Replica 1")]
            RDSReplica2[("RDS Read Replica 2")]
        end
    end

    subgraph AWS["AWS Services"]
        S3[("S3<br/>Raw HTML<br/>AI Models")]
    end

    subgraph External["External"]
        Binance["Binance API"]
        NewsSites["News Sites"]
    end

    Browser <--> ALB
    ALB --> App1
    ALB --> App2
    ALB --> App3
    
    App1 --> Redis
    App2 --> Redis
    App3 --> Redis
    
    App1 --> RDSReplica1
    App2 --> RDSReplica1
    App3 --> RDSReplica2
    
    App1 --> SQS
    App2 --> SQS
    App3 --> SQS
    
    SQS --> Crawler1
    SQS --> Crawler2
    SQS --> AI1
    SQS --> AI2
    
    Crawler1 --> RDSMaster
    Crawler2 --> RDSMaster
    Crawler1 --> S3
    Crawler2 --> S3
    Crawler1 --> NewsSites
    Crawler2 --> NewsSites
    
    AI1 --> RDSMaster
    AI2 --> RDSMaster
    
    App1 --> Binance
    
    RDSMaster --> RDSReplica1
    RDSMaster --> RDSReplica2
