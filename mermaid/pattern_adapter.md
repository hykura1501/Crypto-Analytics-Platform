flowchart LR
    subgraph External["External API"]
        Binance["Binance API\n[][]interface{}"]
    end
    
    subgraph Adapter["Adapter"]
        parseKline["parseKline()"]
    end
    
    subgraph Internal["Internal Model"]
        MarketPrice["MarketPrice\nstruct"]
    end
    
    Binance -->|raw klines| parseKline
    parseKline -->|convert| MarketPrice
