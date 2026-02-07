import { useNavigate, useLocation } from 'react-router-dom';
import { apiClient } from '../api/client';
import { useAuth } from '../contexts/AuthContext';

interface HeaderProps {
  showSymbolSelector?: boolean;
  selectedSymbol?: string;
  onSymbolChange?: (symbol: string) => void;
  extraActions?: React.ReactNode;
}

export default function Header({ 
  showSymbolSelector = false, 
  selectedSymbol,
  onSymbolChange,
  extraActions 
}: HeaderProps) {
  const navigate = useNavigate();
  const location = useLocation();
  const { canManageSources, canManageUsers, user } = useAuth();

  const handleLogout = async () => {
    await apiClient.logout();
    navigate('/login');
  };

  const symbols = (import.meta.env.VITE_SYMBOLS || 'BTCUSDT,ETHUSDT,BNBUSDT').split(',').map((s: string) => s.trim());

  const isActive = (path: string) => {
    return location.pathname === path;
  };

  return (
    <header className="bg-[#131722] border-b border-[#2a2e39] px-6 py-4 shadow-lg backdrop-blur-sm bg-opacity-95 sticky top-0 z-50">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <button
            onClick={() => navigate('/')}
            className="w-10 h-10 bg-gradient-to-br from-[#26a69a] to-[#1e7a6e] rounded-lg flex items-center justify-center shadow-lg hover:scale-105 transition-transform"
          >
            <span className="text-white font-bold text-lg">₿</span>
          </button>
          <h1 className="text-2xl font-bold bg-gradient-to-r from-[#d1d4dc] to-[#26a69a] bg-clip-text text-transparent">
            Crypto Analysis System
          </h1>
        </div>
        
        <div className="flex items-center gap-3">
          {/* Navigation */}
          <nav className="flex items-center gap-2">
            <button
              onClick={() => navigate('/')}
              className={`px-4 py-2 text-sm font-medium rounded-lg transition-all duration-200 ${
                isActive('/')
                  ? 'text-white bg-gradient-to-r from-[#26a69a] to-[#1e7a6e] shadow-lg'
                  : 'text-[#d1d4dc] bg-[#1e222d] border border-[#2a2e39] hover:bg-[#252936] hover:border-[#26a69a]/50'
              }`}
            >
              Dashboard
            </button>
            <button
              onClick={() => navigate('/news')}
              className={`px-4 py-2 text-sm font-medium rounded-lg transition-all duration-200 ${
                isActive('/news')
                  ? 'text-white bg-gradient-to-r from-[#26a69a] to-[#1e7a6e] shadow-lg'
                  : 'text-[#d1d4dc] bg-[#1e222d] border border-[#2a2e39] hover:bg-[#252936] hover:border-[#26a69a]/50'
              }`}
            >
              News
            </button>
            {canManageSources && (
              <button
                onClick={() => navigate('/sources')}
                className={`px-4 py-2 text-sm font-medium rounded-lg transition-all duration-200 ${
                  isActive('/sources') || location.pathname.startsWith('/sources/')
                    ? 'text-white bg-gradient-to-r from-[#26a69a] to-[#1e7a6e] shadow-lg'
                    : 'text-[#d1d4dc] bg-[#1e222d] border border-[#2a2e39] hover:bg-[#252936] hover:border-[#26a69a]/50'
                }`}
              >
                Sources
              </button>
            )}
            {canManageUsers && (
              <button
                onClick={() => navigate('/users')}
                className={`px-4 py-2 text-sm font-medium rounded-lg transition-all duration-200 ${
                  isActive('/users')
                    ? 'text-white bg-gradient-to-r from-[#26a69a] to-[#1e7a6e] shadow-lg'
                    : 'text-[#d1d4dc] bg-[#1e222d] border border-[#2a2e39] hover:bg-[#252936] hover:border-[#26a69a]/50'
                }`}
              >
                Users
              </button>
            )}
          </nav>

          {/* Symbol Selector (only on Dashboard) */}
          {showSymbolSelector && selectedSymbol && onSymbolChange && (
            <select
              value={selectedSymbol}
              onChange={(e) => onSymbolChange(e.target.value)}
              className="px-3 py-2 bg-[#1e222d] border border-[#2a2e39] rounded-lg text-sm text-[#d1d4dc] focus:outline-none focus:ring-2 focus:ring-[#26a69a] hover:bg-[#252936] transition-all cursor-pointer"
            >
              {symbols.map((sym) => (
                <option key={sym} value={sym} className="bg-[#1e222d] text-[#d1d4dc]">
                  {sym}
                </option>
              ))}
            </select>
          )}

          {/* Extra Actions */}
          {extraActions}

          {/* User Info */}
          {user && (
            <div className="flex items-center gap-2 px-3 py-1.5 bg-[#1e222d] border border-[#2a2e39] rounded-lg">
              <div className="w-8 h-8 rounded-full bg-gradient-to-br from-[#26a69a] to-[#1e7a6e] flex items-center justify-center text-white text-xs font-bold">
                {user.email.charAt(0).toUpperCase()}
              </div>
              <div className="hidden md:block">
                <div className="text-xs font-medium text-[#d1d4dc]">{user.email}</div>
                <div className="text-xs text-[#758696]">{user.role}</div>
              </div>
            </div>
          )}

          {/* Logout Button */}
          <button
            onClick={handleLogout}
            className="px-4 py-2 text-sm font-medium text-white bg-gradient-to-r from-[#ef5350] to-[#c62828] rounded-lg hover:from-[#f56565] hover:to-[#d32f2f] focus:outline-none focus:ring-2 focus:ring-[#ef5350] transition-all duration-200 shadow-lg"
          >
            Logout
          </button>
        </div>
      </div>
    </header>
  );
}
