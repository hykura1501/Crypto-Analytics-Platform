flowchart TB
    subgraph Singleton["Singleton Pattern"]
        Hub["WebSocket Hub\n(single instance)"]
    end
    
    C1[Client 1]
    C2[Client 2]
    C3[Client N]
    
    Hub --> C1
    Hub --> C2
    Hub --> C3
    
    Redis["Redis Cache\n(single instance)"]
    Redis --> AIService[AI Service]
