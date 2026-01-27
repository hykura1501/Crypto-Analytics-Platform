flowchart TB
    subgraph Client["AI Service Client"]
        Director["PredictionDirector"]
        DirectUse["Direct Builder Usage"]
    end

    subgraph BuilderClass["PredictionModelBuilder"]
        Builder["PredictionModelBuilder\n+for_symbol()\n+with_horizon()\n+using_model()\n+with_features()\n+enable_shap_explanation()\n+train_with_years()\n+build()"]
    end

    subgraph Config["Configuration"]
        PredictionConfig["PredictionConfig\n-symbol: string\n-horizon: int\n-model_type: string\n-features: list\n-enable_shap: bool\n-train_years: int"]
    end

    subgraph Product["Built Product"]
        PredictionModel["PredictionModel\n+train()\n+predict()\n+get_config()"]
    end

    subgraph PresetConfigs["Director Presets"]
        BTCShortTerm["build_btc_short_term()\nBTCUSDT, 1h, XGBoost"]
        ETHLongTerm["build_eth_long_term()\nETHUSDT, 7d, XGBoost"]
        Custom["build_custom(symbol, horizon)"]
    end

    subgraph Output["Prediction Output"]
        Result["PredictionResult\n+symbol\n+prediction: UP/DOWN\n+confidence\n+explanation: SHAP values\n+timestamp"]
    end

    Director --> PresetConfigs
    DirectUse --> Builder
    PresetConfigs --> Builder
    
    Builder --> PredictionConfig
    Builder --> PredictionModel
    
    PredictionModel --> PredictionConfig
    PredictionModel --> Result

    style Builder fill:#e1f5fe
    style PredictionModel fill:#c8e6c9
    style PredictionConfig fill:#fff3e0
    style Result fill:#f3e5f5
