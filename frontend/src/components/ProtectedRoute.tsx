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
  const [isAuthenticated, setIsAuthenticated] = useState<boolean | null>(null);
  const { user, loading: authLoading, isAdmin } = useAuth();

  useEffect(() => {
    const accessToken = Cookies.get(COOKIE_NAMES.ACCESS_TOKEN);
    setIsAuthenticated(!!accessToken);
  }, []);

  if (isAuthenticated === null || (isAuthenticated && authLoading && !user)) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="text-gray-600">Loading...</div>
      </div>
    );
  }

  if (!isAuthenticated) {
    return <Navigate to="/login" replace />;
  }

  if (adminOnly && !isAdmin) {
    return <Navigate to="/" replace />;
  }

  return <>{children}</>;
}

