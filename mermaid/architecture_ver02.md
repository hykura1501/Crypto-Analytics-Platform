flowchart TB
    subgraph Client["Client"]
        Browser["Browser/Mobile App"]
    end

    subgraph VPC["VPC"]
        subgraph PublicSubnet["Public Subnet"]
            EC2["EC2 App Server<br/>Monolith Application"]
        end
        
        subgraph PrivateSubnet["Private Subnet"]
            RDS[("RDS PostgreSQL/MySQL<br/>Managed Database<br/>users, news, prices, signals")]
        end
    end

    subgraph AWS["AWS Services"]
        S3[("S3 Object Storage<br/>Raw HTML News<br/>AI Model Files<br/>Logs and Backups")]
    end

    subgraph External["External Services"]
        Binance["Binance API"]
        NewsSites["News Sites"]
    end

    Browser <-->|"HTTP/HTTPS<br/>Port 80/443"| EC2
    
    EC2 -->|"Private Connection<br/>Security Group"| RDS
    EC2 -->|"S3 Endpoint"| S3
    
    EC2 --> Binance
    EC2 --> NewsSites
