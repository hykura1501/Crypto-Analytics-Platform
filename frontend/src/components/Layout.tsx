import { type ReactNode } from 'react';
import Header from './Header';

interface LayoutProps {
  children: ReactNode;
  showSymbolSelector?: boolean;
  selectedSymbol?: string;
  onSymbolChange?: (symbol: string) => void;
  extraActions?: React.ReactNode;
}

export default function Layout({ 
  children, 
  showSymbolSelector,
  selectedSymbol,
  onSymbolChange,
  extraActions 
}: LayoutProps) {
  return (
    <div className="h-screen bg-gradient-to-br from-[#0a0e27] via-[#131722] to-[#0a0e27] flex flex-col overflow-hidden">
      <Header 
        showSymbolSelector={showSymbolSelector}
        selectedSymbol={selectedSymbol}
        onSymbolChange={onSymbolChange}
        extraActions={extraActions}
      />
      <div className="flex-1 overflow-y-auto">
        {children}
      </div>
    </div>
  );
}
