import axios, {
  type AxiosInstance,
  type InternalAxiosRequestConfig,
} from "axios";
import Cookies from "js-cookie";
import { API_BASE_URL, API_VERSION_PREFIX, COOKIE_NAMES } from "../config";
import type {
  ApiSuccessResponse,
  Article,
  AuthResponse,
  LoginRequest,
  MarketPrice,
  PredictionRequest,
  PredictionResponse,
  PredictionResult,
  RefreshTokenRequest,
  RegisterRequest,
  Source,
  CreateSourceRequest,
  UpdateSourceRequest,
  User,
  UserRole,
} from "../types";

class ApiClient {
  private client: AxiosInstance;
  private isRefreshing = false;
  private refreshSubscribers: Array<(token: string) => void> = [];

  constructor() {
    this.client = axios.create({
      baseURL: `${API_BASE_URL}${API_VERSION_PREFIX}`,
      withCredentials: true,
      headers: {
        "Content-Type": "application/json",
      },
    });

    // Request interceptor to add access token
    this.client.interceptors.request.use(
      (config: InternalAxiosRequestConfig) => {
        const accessToken = Cookies.get(COOKIE_NAMES.ACCESS_TOKEN);
        if (accessToken && config.headers) {
          config.headers.Authorization = `Bearer ${accessToken}`;
        }
        return config;
      },
      (error) => {
        return Promise.reject(error);
      }
    );

    // Response interceptor to handle token refresh
    this.client.interceptors.response.use(
      (response) => response,
      async (error) => {
        const originalRequest = error.config;
        
        // Skip refresh for auth endpoints to avoid infinite loop
        const authEndpoints = ['/auth/login', '/auth/register', '/auth/refresh', '/auth/logout'];
        const isAuthEndpoint = authEndpoints.some(endpoint => 
          originalRequest?.url?.includes(endpoint)
        );
        
        // If error is 401 and we haven't tried to refresh yet
        if (error.response?.status === 401 && !originalRequest?._retry && !isAuthEndpoint) {
          originalRequest._retry = true;

          if (this.isRefreshing) {
            // If already refreshing, queue this request
            return new Promise((resolve, reject) => {
              this.refreshSubscribers.push((token: string) => {
                if (token) {
                  originalRequest.headers.Authorization = `Bearer ${token}`;
                  resolve(this.client(originalRequest));
                } else {
                  reject(error);
                }
              });
            });
          }

          this.isRefreshing = true;

          try {
            const refreshToken = Cookies.get(COOKIE_NAMES.REFRESH_TOKEN);
            if (!refreshToken) {
              throw new Error("No refresh token available");
            }

            // Call refresh endpoint directly without going through interceptor
            const response = await axios.post<AuthResponse>(
              `${API_BASE_URL}${API_VERSION_PREFIX}/auth/refresh`,
              { refresh_token: refreshToken },
              {
                headers: { 'Content-Type': 'application/json' },
                withCredentials: true,
              }
            );
            
            const { access_token, refresh_token: newRefreshToken, expires_in } = response.data;

            // Update access token cookie with correct expiry
            Cookies.set(COOKIE_NAMES.ACCESS_TOKEN, access_token, {
              expires: new Date(Date.now() + expires_in * 1000),
              httpOnly: false,
            });

            // Update refresh token if provided
            if (newRefreshToken) {
              Cookies.set(COOKIE_NAMES.REFRESH_TOKEN, newRefreshToken, {
                expires: 7, // 7 days
                httpOnly: false,
              });
            }

            // Notify all queued requests with new token
            this.refreshSubscribers.forEach((cb) => cb(access_token));
            this.refreshSubscribers = [];

            // Retry original request with new token
            originalRequest.headers.Authorization = `Bearer ${access_token}`;
            return this.client(originalRequest);
          } catch (refreshError) {
            // Refresh failed, notify queued requests and redirect to login
            this.refreshSubscribers.forEach((cb) => cb(''));
            this.refreshSubscribers = [];
            Cookies.remove(COOKIE_NAMES.ACCESS_TOKEN);
            Cookies.remove(COOKIE_NAMES.REFRESH_TOKEN);
            
            // Only redirect if we're not already on the login page
            if (window.location.pathname !== '/login') {
              window.location.href = "/login";
            }
            return Promise.reject(refreshError);
          } finally {
            this.isRefreshing = false;
          }
        }

        return Promise.reject(error);
      }
    );
  }

  // Auth endpoints
  async register(credentials: RegisterRequest): Promise<AuthResponse> {
    const response = await this.client.post<AuthResponse>(
      "/auth/register",
      credentials
    );

    // Store tokens in cookies (backend should set httpOnly cookies, but we also set them client-side)
    if (response.data) {
      const authData = response.data;
      Cookies.set(COOKIE_NAMES.ACCESS_TOKEN, authData.access_token, {
        expires: new Date(Date.now() + authData.expires_in * 1000),
        httpOnly: false,
      });
      Cookies.set(COOKIE_NAMES.REFRESH_TOKEN, authData.refresh_token, {
        expires: 7, // 7 days
        httpOnly: false,
      });
    }

    return response.data;
  }

  async login(credentials: LoginRequest): Promise<AuthResponse> {
    const response = await this.client.post<AuthResponse>(
      "/auth/login",
      credentials
    );

    // Store tokens in cookies (backend should set httpOnly cookies, but we also set them client-side)
    if (response.data) {
      const authData = response.data;
      Cookies.set(COOKIE_NAMES.ACCESS_TOKEN, authData.access_token, {
        expires: new Date(Date.now() + authData.expires_in * 1000),
        httpOnly: false,
      });
      Cookies.set(COOKIE_NAMES.REFRESH_TOKEN, authData.refresh_token, {
        expires: 7, // 7 days
        httpOnly: false,
      });
    }

    return response.data;
  }

  async refreshToken(
    refreshToken: string
  ): Promise<AuthResponse> {
    const response = await this.client.post<AuthResponse>(
      "/auth/refresh",
      { refresh_token: refreshToken } as RefreshTokenRequest
    );

    if (response.data) {
      const authData = response.data;
      Cookies.set(COOKIE_NAMES.ACCESS_TOKEN, authData.access_token, {
        expires: new Date(Date.now() + authData.expires_in * 1000),
        httpOnly: false,
      });
      if (authData.refresh_token) {
        Cookies.set(COOKIE_NAMES.REFRESH_TOKEN, authData.refresh_token, {
          expires: 7, // 7 days
          httpOnly: false,
        });
      }
    }

    return response.data;
  }

  async logout(): Promise<void> {
    const refreshToken = Cookies.get(COOKIE_NAMES.REFRESH_TOKEN);
    if (refreshToken) {
      try {
        await this.client.post("/auth/logout", { refresh_token: refreshToken });
      } catch (error) {
        console.error("Logout error:", error);
      }
    }
    Cookies.remove(COOKIE_NAMES.ACCESS_TOKEN);
    Cookies.remove(COOKIE_NAMES.REFRESH_TOKEN);
  }

  async getMe(): Promise<User> {
    const response = await this.client.get<User>("/auth/me");
    return response.data;
  }

  async getUsers(): Promise<User[]> {
    const response = await this.client.get<ApiSuccessResponse<User[]>>("/auth/users");
    return Array.isArray(response.data.data) ? response.data.data : [];
  }

  async updateUserRole(userId: number, role: UserRole): Promise<User> {
    const response = await this.client.patch<ApiSuccessResponse<User>>(
      `/auth/users/${userId}/role`,
      { role }
    );
    return response.data.data;
  }

  // Market endpoints
  async getHistory(params: {
    symbol: string;
    interval: string;
    limit?: number;
    from?: number;
    to?: number;
  }): Promise<ApiSuccessResponse<MarketPrice[]>> {
    const response = await this.client.get<ApiSuccessResponse<MarketPrice[]>>(
      "/market/history",
      { params }
    );
    return response.data;
  }

  // News endpoints
  async getNews(params?: {
    skip?: number;
    limit?: number;
    cursor?: number | string;
    source_id?: string;
    event_type?: string;
    language?: string;
    sort_by?: 'published_at' | 'created_at';
    sort_order?: 'ASC' | 'DESC';
  }): Promise<{ articles: Article[]; next_cursor: string | number; has_more: boolean }> {
    const response = await this.client.get<{ articles: Article[]; next_cursor: string | number; has_more: boolean }>("/news/articles", {
      params,
    });
    return response.data;
  }

  // Source endpoints
  async getSources(): Promise<Source[]> {
    const response = await this.client.get<{
      sources: Source[];
      total: number;
    }>("/news/sources");
    return response.data.sources;
  }

  async getSource(sourceId: string): Promise<Source> {
    const response = await this.client.get<Source>(`/news/sources/${sourceId}`);
    return response.data;
  }

  async createSource(data: CreateSourceRequest): Promise<Source> {
    const response = await this.client.post<Source>("/news/sources", data);
    return response.data;
  }

  async updateSource(sourceId: string, data: UpdateSourceRequest): Promise<Source> {
    const response = await this.client.put<Source>(`/news/sources/${sourceId}`, data);
    return response.data;
  }

  async deleteSource(sourceId: string): Promise<void> {
    await this.client.delete(`/news/sources/${sourceId}`);
  }

  async analyzeSource(sourceId: string): Promise<{ message: string; source_id: string }> {
    const response = await this.client.post<{ message: string; source_id: string }>(
      `/news/sources/${sourceId}/analyze`
    );
    return response.data;
  }

  // Prediction endpoints
  async trainModel(request: PredictionRequest): Promise<{ message: string; rmse?: number; mae?: number; directional_accuracy?: number }> {
    const response = await this.client.post<{ message: string; rmse?: number; mae?: number; directional_accuracy?: number }>(
      "/ai/prediction/train",
      request
    );
    return response.data;
  }

  async getPrediction(symbol: string, horizon_hours: number): Promise<PredictionResult> {
    const response = await this.client.get<PredictionResponse>(
      `/ai/prediction/predict/${symbol}/${horizon_hours}`
    );
    return response.data.result;
  }

  getClient(): AxiosInstance {
    return this.client;
  }
}

export const apiClient = new ApiClient();
