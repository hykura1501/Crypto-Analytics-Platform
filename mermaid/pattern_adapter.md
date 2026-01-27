flowchart TB
    subgraph Client["Client Code"]
        MarketService["MarketService"]
    end

    subgraph AdapterInterface["Exchange Adapter Interface"]
        ExchangeAdapter["ExchangeAdapter\n+FetchPrice(symbol)\n+GetExchangeName()"]
    end

    subgraph Adapters["Concrete Adapters"]
        BinanceAdapter["BinanceAdapter"]
        CoinGeckoAdapter["CoinGeckoAdapter"]
        CoinbaseAdapter["CoinbaseAdapter"]
        KrakenAdapter["KrakenAdapter"]
    end

    subgraph ExternalAPIs["External APIs"]
        BinanceAPI["Binance API\n{symbol, price, timestamp}"]
        CoinGeckoAPI["CoinGecko API\n{id, current_price, last_updated}"]
        CoinbaseAPI["Coinbase API\n{base, currency, amount}"]
        KrakenAPI["Kraken API\n{pair, c: [price], ...]"]
    end

    subgraph StandardOutput["Standardized Output"]
        PriceData["PriceData Interface\n+GetSymbol()\n+GetPrice()\n+GetTimestamp()\n+GetExchange()"]
    end

    MarketService --> ExchangeAdapter
    ExchangeAdapter --> BinanceAdapter
    ExchangeAdapter --> CoinGeckoAdapter
    ExchangeAdapter --> CoinbaseAdapter
    ExchangeAdapter --> KrakenAdapter

    BinanceAdapter --> BinanceAPI
    CoinGeckoAdapter --> CoinGeckoAPI
    CoinbaseAdapter --> CoinbaseAPI
    KrakenAdapter --> KrakenAPI

    BinanceAdapter --> PriceData
    CoinGeckoAdapter --> PriceData
    CoinbaseAdapter --> PriceData
    KrakenAdapter --> PriceData

    style ExchangeAdapter fill:#e1f5fe
    style PriceData fill:#c8e6c9
    style BinanceAdapter fill:#fff3e0
    style CoinGeckoAdapter fill:#fff3e0
    style CoinbaseAdapter fill:#fff3e0
    style KrakenAdapter fill:#fff3e0
