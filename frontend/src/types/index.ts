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

export interface AuthResponse {
  access_token: string;
  refresh_token: string;
  token_type: string;
  expires_in: number;
  user?: {
    id: number;
    email: string;
    first_name?: string;
    last_name?: string;
  };
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

