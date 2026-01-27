flowchart TB
    subgraph Input["Order Input"]
        Order["Order\nValue: $50,000"]
    end

    subgraph BaseComponent["Base Component"]
        BaseFee["BaseFeeCalculator\nRate: 0.1%\nFee: $50"]
    end

    subgraph Decorators["Fee Decorators"]
        PeakHour["PeakHourFeeDecorator\n+0.05% during 8-10, 15-17\nAdditional: $25"]
        HighVol["HighVolatilityFeeDecorator\n+0.1% when volatility > 5%\nAdditional: $50"]
        LargeOrder["LargeOrderFeeDecorator\n+0.02% for orders > $100K\nAdditional: $0"]
        VIPDiscount["VIPDiscountDecorator\n-30% for Gold tier\nDiscount: -$37.50"]
    end

    subgraph WrappingOrder["Decorator Wrapping Order"]
        Wrap1["1. BaseFeeCalculator"]
        Wrap2["2. PeakHourFeeDecorator wraps Base"]
        Wrap3["3. HighVolatilityFeeDecorator wraps PeakHour"]
        Wrap4["4. VIPDiscountDecorator wraps HighVolatility"]
    end

    subgraph Calculation["Fee Calculation Flow"]
        Calc1["BaseFee = $50,000 × 0.1% = $50"]
        Calc2["+ PeakHour = $50,000 × 0.05% = $25"]
        Calc3["+ HighVol = $50,000 × 0.1% = $50"]
        Calc4["Subtotal = $125"]
        Calc5["- VIP Gold 30% = -$37.50"]
        Calc6["Total Fee = $87.50"]
    end

    subgraph Output["Final Output"]
        TotalFee["Total Fee: $87.50\nEffective Rate: 0.175%"]
    end

    Order --> BaseFee
    BaseFee --> PeakHour
    PeakHour --> HighVol
    HighVol --> VIPDiscount
    VIPDiscount --> TotalFee

    Wrap1 --> Wrap2 --> Wrap3 --> Wrap4
    Calc1 --> Calc2 --> Calc3 --> Calc4 --> Calc5 --> Calc6

    style BaseFee fill:#e1f5fe
    style PeakHour fill:#fff3e0
    style HighVol fill:#fff3e0
    style LargeOrder fill:#fff3e0
    style VIPDiscount fill:#c8e6c9
    style TotalFee fill:#f3e5f5
