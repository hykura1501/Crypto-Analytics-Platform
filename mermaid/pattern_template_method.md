flowchart TB
    subgraph Pipeline["Prediction Pipeline - Template Method"]
        S1["1. load_price_data()"]
        S2["2. load_news_data()"]
        S3["3. process_news()\nalign with price buckets"]
        S4["4. create_price_features()"]
        S5["5. train() / predict()"]
    end
    
    S1 --> S2
    S2 --> S3
    S3 --> S4
    S4 --> S5
