flowchart TB
    subgraph Client["Crawler Service"]
        CrawlerContext["CrawlerContext\n+RegisterStrategy()\n+Crawl(source)\n+CrawlAll(sources)"]
    end

    subgraph StrategyInterface["Crawl Strategy Interface"]
        CrawlStrategy["CrawlStrategy\n+Crawl(source) []Article\n+GetType() string\n+CanHandle(source) bool"]
    end

    subgraph Strategies["Concrete Strategies"]
        RSSStrategy["RSSStrategy\ntype: rss"]
        HTMLStrategy["HTMLScrapingStrategy\ntype: html"]
        APIStrategy["APIStrategy\ntype: api"]
        SocialStrategy["SocialMediaStrategy\ntype: social"]
    end

    subgraph Sources["News Sources"]
        CoinDesk["CoinDesk\nRSS Feed"]
        CryptoNews["CryptoNews\nHTML Scraping"]
        NewsAPI["NewsAPI\nREST API"]
        Twitter["Twitter/X\nSocial API"]
    end

    subgraph Output["Standardized Output"]
        Article["Article\n+Title\n+Content\n+URL\n+PublishedAt\n+Source\n+Author"]
    end

    CrawlerContext --> CrawlStrategy
    CrawlStrategy --> RSSStrategy
    CrawlStrategy --> HTMLStrategy
    CrawlStrategy --> APIStrategy
    CrawlStrategy --> SocialStrategy

    RSSStrategy --> CoinDesk
    HTMLStrategy --> CryptoNews
    APIStrategy --> NewsAPI
    SocialStrategy --> Twitter

    RSSStrategy --> Article
    HTMLStrategy --> Article
    APIStrategy --> Article
    SocialStrategy --> Article

    style CrawlStrategy fill:#e1f5fe
    style Article fill:#c8e6c9
    style RSSStrategy fill:#fff3e0
    style HTMLStrategy fill:#fff3e0
    style APIStrategy fill:#fff3e0
    style SocialStrategy fill:#fff3e0
