import React, {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useState,
} from "react";
import Cookies from "js-cookie";
import { apiClient } from "../api/client";
import { COOKIE_NAMES } from "../config";
import type { User } from "../types";

interface AuthContextType {
  user: User | null;
  loading: boolean;
  error: string | null;
  refreshUser: () => Promise<void>;
  isAdmin: boolean;
  canViewPrediction: boolean;
  canManageSources: boolean;
  canManageUsers: boolean;
}

const AuthContext = createContext<AuthContextType | null>(null);

export function useAuth() {
  const ctx = useContext(AuthContext);
  if (!ctx) {
    throw new Error("useAuth must be used within AuthProvider");
  }
  return ctx;
}

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const refreshUser = useCallback(async () => {
    const accessToken = Cookies.get(COOKIE_NAMES.ACCESS_TOKEN);
    const refreshToken = Cookies.get(COOKIE_NAMES.REFRESH_TOKEN);
    
    // If no access token but have refresh token, try to refresh first
    if (!accessToken && refreshToken) {
      try {
        setLoading(true);
        setError(null);
        await apiClient.refreshToken(refreshToken);
        // Token refreshed, now fetch user
        const me = await apiClient.getMe();
        setUser(me);
        setLoading(false);
        return;
      } catch (e) {
        // Refresh failed, clear everything
        setUser(null);
        setError((e as Error).message);
        setLoading(false);
        return;
      }
    }
    
    // If no token at all
    if (!accessToken) {
      setUser(null);
      setLoading(false);
      return;
    }
    
    // Fetch user with existing access token
    setLoading(true);
    setError(null);
    try {
      const me = await apiClient.getMe();
      setUser(me);
    } catch (e) {
      setUser(null);
      setError((e as Error).message);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    const accessToken = Cookies.get(COOKIE_NAMES.ACCESS_TOKEN);
    const refreshToken = Cookies.get(COOKIE_NAMES.REFRESH_TOKEN);
    
    // If we have neither token, clear user
    if (!accessToken && !refreshToken) {
      setUser(null);
      setLoading(false);
      return;
    }
    
    // refreshUser will handle refreshing if needed
    refreshUser();
  }, [refreshUser]);

  useEffect(() => {
    const checkToken = () => {
      const accessToken = Cookies.get(COOKIE_NAMES.ACCESS_TOKEN);
      const refreshToken = Cookies.get(COOKIE_NAMES.REFRESH_TOKEN);
      
      // Only clear user if we have neither token
      // If refresh token exists, interceptor will handle refresh
      if (!accessToken && !refreshToken && user) {
        setUser(null);
      }
    };
    const t = setInterval(checkToken, 2000);
    return () => clearInterval(t);
  }, [user]);

  const role = user?.role ?? "NORMAL";
  const isAdmin = role === "ADMIN";
  const canViewPrediction = role === "ADMIN" || role === "VIP";
  const canManageSources = isAdmin;
  const canManageUsers = isAdmin;

  return (
    <AuthContext.Provider
      value={{
        user,
        loading,
        error,
        refreshUser,
        isAdmin,
        canViewPrediction,
        canManageSources,
        canManageUsers,
      }}
    >
      {children}
    </AuthContext.Provider>
  );
}
