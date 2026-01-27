flowchart TB
    subgraph DataSource["External Data Source"]
        BinanceWS["Binance WebSocket\nReal-time Price Stream"]
    end

    subgraph Publisher["Price Publisher"]
        PricePublisher["PricePublisher\n+Subscribe(observer)\n+Unsubscribe(observer)\n+Notify(update)\n+Start()"]
        PriceUpdate["PriceUpdate\n+Symbol\n+Price\n+Volume\n+Timestamp\n+Change24h"]
    end

    subgraph ObserverInterface["Observer Interface"]
        PriceObserver["PriceObserver\n+OnPriceUpdate(update)\n+GetName()"]
    end

    subgraph Observers["Concrete Observers"]
        DashboardObs["DashboardObserver\nPush to WebSocket clients"]
        AlertObs["AlertObserver\nCheck alert conditions"]
        AIObserver["AIObserver\nTrigger AI prediction"]
        DBObserver["DatabaseObserver\nBuffer and save to DB"]
    end

    subgraph Actions["Observer Actions"]
        WSClients["WebSocket Clients\nReal-time Dashboard"]
        AlertService["Alert Service\nEmail/SMS/Push"]
        AIPrediction["AI Prediction\nUP/DOWN signal"]
        Database[("PostgreSQL\nPrice History")]
    end

    BinanceWS --> PricePublisher
    PricePublisher --> PriceUpdate
    PricePublisher --> PriceObserver

    PriceObserver --> DashboardObs
    PriceObserver --> AlertObs
    PriceObserver --> AIObserver
    PriceObserver --> DBObserver

    DashboardObs --> WSClients
    AlertObs --> AlertService
    AIObserver --> AIPrediction
    DBObserver --> Database

    style PricePublisher fill:#e1f5fe
    style PriceObserver fill:#fff3e0
    style DashboardObs fill:#c8e6c9
    style AlertObs fill:#c8e6c9
    style AIObserver fill:#c8e6c9
    style DBObserver fill:#c8e6c9
