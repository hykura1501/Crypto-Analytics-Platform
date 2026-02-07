---
title: "WebSocket Real-time Flow"
description: "Luồng kết nối WebSocket"
---

sequenceDiagram
    participant Client as Client
    participant Market as Market Service
    participant Hub as WebSocket Hub
    participant Binance as Binance API

    Client->>Market: WebSocket Connect /ws
    Market->>Hub: Register Client
    Hub-->>Client: Connected
    
    Client->>Hub: Subscribe BTCUSDT
    Hub->>Hub: topics[BTCUSDT][client]=true
    
    Market->>Binance: Connect WebSocket
    
    loop Real-time Updates
        Binance-->>Market: Price Update
        Market->>Hub: Broadcast(BTCUSDT, price)
        Hub->>Client: Send price data
        Client->>Client: Update Chart
    end
    
    Client->>Market: Disconnect
    Market->>Hub: Unregister Client
