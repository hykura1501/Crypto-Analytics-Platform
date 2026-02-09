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
    
    // If no tokens at all, clear user
    if (!accessToken && !refreshToken) {
      setUser(null);
      setLoading(false);
      return;
    }
    
    // Fetch user - the API client interceptor will handle token refresh if needed
    setLoading(true);
    setError(null);
    try {
      const me = await apiClient.getMe();
      setUser(me);
    } catch (e) {
      // If getMe fails, it means tokens are invalid or refresh failed
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
    const handleAuthChange = () => {
      const accessToken = Cookies.get(COOKIE_NAMES.ACCESS_TOKEN);
      const refreshToken = Cookies.get(COOKIE_NAMES.REFRESH_TOKEN);

      if (!accessToken && !refreshToken && user) {
        setUser(null);
      }
    };

    // Listen for custom auth-state-changed events (fired from API client on login/logout)
    window.addEventListener("auth-state-changed", handleAuthChange);

    // Also listen for storage events from other tabs
    window.addEventListener("storage", handleAuthChange);

    // Fallback: check every 30s instead of 2s (primarily for cookie expiry detection)
    const t = setInterval(handleAuthChange, 30_000);

    return () => {
      window.removeEventListener("auth-state-changed", handleAuthChange);
      window.removeEventListener("storage", handleAuthChange);
      clearInterval(t);
    };
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
