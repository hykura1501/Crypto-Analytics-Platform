flowchart TB
    subgraph Subject["Subject"]
        Hub["WebSocket Hub"]
    end
    
    subgraph Observers["Observers"]
        C1["Client 1\n(topic: btc@1m)"]
        C2["Client 2\n(topic: btc@1m)"]
        C3["Client 3\n(topic: eth@5m)"]
    end
    
    Binance[Binance WS] -->|price update| Hub
    Hub -->|broadcast by topic| C1
    Hub -->|broadcast by topic| C2
    Hub -->|broadcast by topic| C3
