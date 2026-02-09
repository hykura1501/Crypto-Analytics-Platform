import { useEffect, useState } from 'react';
import { Navigate } from 'react-router-dom';
import Cookies from 'js-cookie';
import { COOKIE_NAMES } from '../config';
import { useAuth } from '../contexts/AuthContext';

interface ProtectedRouteProps {
  children: React.ReactNode;
  adminOnly?: boolean;
}

export default function ProtectedRoute({ children, adminOnly = false }: ProtectedRouteProps) {
  const { user, loading: authLoading, isAdmin } = useAuth();

  // Show loading while auth is loading (includes refresh token check)
  if (authLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="text-gray-600">Loading...</div>
      </div>
    );
  }

  // Check if we have tokens (AuthContext will handle refresh if needed)
  const accessToken = Cookies.get(COOKIE_NAMES.ACCESS_TOKEN);
  const refreshToken = Cookies.get(COOKIE_NAMES.REFRESH_TOKEN);
  
  // Only redirect to login if we have neither token AND no user
  // If we have refresh token, AuthContext will try to refresh
  if (!accessToken && !refreshToken && !user) {
    return <Navigate to="/login" replace />;
  }

  if (adminOnly && !isAdmin) {
    return <Navigate to="/" replace />;
  }

  return <>{children}</>;
}

