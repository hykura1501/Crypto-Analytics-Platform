import axios, { type AxiosInstance, type InternalAxiosRequestConfig } from 'axios';
// @ts-expect-error - js-cookie uses CommonJS export which works at runtime
import Cookies from 'js-cookie';
import { API_BASE_URL, COOKIE_NAMES } from '../config';
import { type ApiSuccessResponse, type Article, type AuthResponse, type LoginRequest, type MarketPrice, type RefreshTokenRequest } from '../types';

class ApiClient {
  private client: AxiosInstance;
  private isRefreshing = false;
  private refreshSubscribers: Array<(token: string) => void> = [];

  constructor() {
    this.client = axios.create({
      baseURL: API_BASE_URL,
      withCredentials: true,
      headers: {
        'Content-Type': 'application/json',
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

        // If error is 401 and we haven't tried to refresh yet
        if (error.response?.status === 401 && !originalRequest._retry) {
          if (this.isRefreshing) {
            // If already refreshing, wait for the new token
            return new Promise((resolve) => {
              this.refreshSubscribers.push((token: string) => {
                originalRequest.headers.Authorization = `Bearer ${token}`;
                resolve(this.client(originalRequest));
              });
            });
          }

          originalRequest._retry = true;
          this.isRefreshing = true;

          try {
            const refreshToken = Cookies.get(COOKIE_NAMES.REFRESH_TOKEN);
            if (!refreshToken) {
              throw new Error('No refresh token available');
            }

            const response = await this.refreshToken(refreshToken);
            const { access_token } = response.data;

            // Update cookie
            Cookies.set(COOKIE_NAMES.ACCESS_TOKEN, access_token, {
              expires: 7, // 7 days
              httpOnly: false, // Note: js-cookie can't set httpOnly, but backend should set it
            });

            // Notify all subscribers
            this.refreshSubscribers.forEach((cb) => cb(access_token));
            this.refreshSubscribers = [];

            // Retry original request
            originalRequest.headers.Authorization = `Bearer ${access_token}`;
            return this.client(originalRequest);
          } catch (refreshError) {
            // Refresh failed, redirect to login
            this.refreshSubscribers = [];
            Cookies.remove(COOKIE_NAMES.ACCESS_TOKEN);
            Cookies.remove(COOKIE_NAMES.REFRESH_TOKEN);
            window.location.href = '/login';
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
  async login(credentials: LoginRequest): Promise<AuthResponse> {
    const response = await this.client.post<AuthResponse>(
      '/auth/login',
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

  async refreshToken(refreshToken: string): Promise<ApiSuccessResponse<AuthResponse>> {
    const response = await this.client.post<ApiSuccessResponse<AuthResponse>>(
      '/auth/refresh',
      { refresh_token: refreshToken } as RefreshTokenRequest
    );
    
    if (response.data.data) {
      const authData = response.data.data as unknown as AuthResponse;
      Cookies.set(COOKIE_NAMES.ACCESS_TOKEN, authData.access_token, {
        expires: new Date(Date.now() + authData.expires_in * 1000),
        httpOnly: false,
      });
    }
    
    return response.data;
  }

  async logout(): Promise<void> {
    const refreshToken = Cookies.get(COOKIE_NAMES.REFRESH_TOKEN);
    if (refreshToken) {
      try {
        await this.client.post('/auth/logout', { refresh_token: refreshToken });
      } catch (error) {
        console.error('Logout error:', error);
      }
    }
    Cookies.remove(COOKIE_NAMES.ACCESS_TOKEN);
    Cookies.remove(COOKIE_NAMES.REFRESH_TOKEN);
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
      '/market/api/v1/market/history',
      { params }
    );
    return response.data;
  }

  // News endpoints
  async getNews(params?: {
    skip?: number;
    limit?: number;
    source_id?: string;
    event_type?: string;
    language?: string;
  }): Promise<Article[]> {
    const response = await this.client.get<Article[]>(
      '/news/articles/',
      { params }
    );
    return response.data;
  }

  getClient(): AxiosInstance {
    return this.client;
  }
}

export const apiClient = new ApiClient();

