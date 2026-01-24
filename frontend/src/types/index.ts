// Market Data Types
export interface MarketPrice {
  symbol: string;
  time: string;
  interval: string;
  open: number;
  high: number;
  low: number;
  close: number;
  volume: number;
}

export interface HistoryRequest {
  symbol: string;
  interval: string;
  limit?: number;
  from?: number;
  to?: number;
}

// News Types
export interface Article {
  id: number;
  source_id: string;
  url: string;
  title: string;
  author?: string;
  published_at: string;
  crawled_at: string;
  content_text: string;
  language: string;
  tags?: string[];
  summary?: string;
  entities?: {
    tickers?: string[];
    coins?: string[];
    people?: string[];
    orgs?: string[];
  };
  sentiment_score?: number;
  confidence?: number;
  event_time?: string;
  event_type?: string;
  created_at: string;
}

// Auth Types
export interface LoginRequest {
  email: string;
  password: string;
}

export interface RegisterRequest {
  email: string;
  password: string;
  first_name?: string;
  last_name?: string;
}

export type UserRole = 'ADMIN' | 'NORMAL' | 'VIP';

export interface User {
  id: number;
  email: string;
  first_name?: string;
  last_name?: string;
  role: UserRole;
  is_active?: boolean;
  created_at?: string;
  updated_at?: string;
}

export interface AuthResponse {
  access_token: string;
  refresh_token: string;
  token_type: string;
  expires_in: number;
  user?: User;
}

export interface RefreshTokenRequest {
  refresh_token: string;
}

// API Response Types
export interface ApiSuccessResponse<T> {
  success: boolean;
  message?: string;
  data: T;
}

export interface ApiErrorResponse {
  error: string;
  message?: string;
}

// Source Types
export interface Source {
  source_id: string;
  rss_url: string;
  title_tag?: string;
  link_tag?: string;
  pub_date_tag?: string;
  summary_selector?: string;
  content_selector?: string;
  author_selector?: string;
  tags_selector?: string;
  created_at: string;
  updated_at: string;
}

export interface CreateSourceRequest {
  source_id: string;
  rss_url: string;
}

export interface UpdateSourceRequest {
  rss_url?: string;
  title_tag?: string;
  link_tag?: string;
  pub_date_tag?: string;
  summary_selector?: string;
  content_selector?: string;
  author_selector?: string;
  tags_selector?: string;
}

// Prediction Types
export interface PredictionRequest {
  symbol: string;
  horizon_hours: number;
  years?: number;
}

export interface PredictionResult {
  prediction_horizon: string;
  predicted_price: string;
  current_price: string;
  predicted_change_pct: string;
  confidence_score: string;
  metadata: {
    model_key: string;
    base_prediction?: string;
    sentiment_adjustment?: string;
    features_count: number;
    news_features_count: number;
    tech_features_count: number;
    recent_news_analyzed: number;
  };
  top_influential_features: string[];
  feature_analysis: {
    top_news_features: Array<{ feature: string; importance: number }>;
    top_technical_indicators: Array<{ feature: string; importance: number }>;
    shap_contributions: Record<string, number>;
  };
  top_news_articles: Array<{
    id: number;
    title: string;
    url: string;
    published_at: string;
    sentiment_score: number;
    language: string;
    keywords: string[];
  }>;
  explanation: {
    primary_factors: {
      most_important_feature: string;
      is_news_important: boolean;
      news_vs_technical: string;
      news_influence_pct: string;
    };
  };
}

export interface PredictionResponse {
  message: string;
  result: PredictionResult;
}
