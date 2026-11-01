'use client';

import React, { useState, useEffect, Component, ErrorInfo, ReactNode } from 'react';
import Sidebar from './Sidebar';
import Header from './Header';
import SourceDrawer from './SourceDrawer';

interface ErrorBoundaryProps {
  children: ReactNode;
}

interface ErrorBoundaryState {
  hasError: boolean;
  error: Error | null;
}

class ErrorBoundary extends Component<ErrorBoundaryProps, ErrorBoundaryState> {
  constructor(props: ErrorBoundaryProps) {
    super(props);
    this.state = { hasError: false, error: null };
  }

  static getDerivedStateFromError(error: Error): ErrorBoundaryState {
    return { hasError: true, error };
  }

  componentDidCatch(error: Error, errorInfo: ErrorInfo) {
    console.error('CodeDNA View Error:', error, errorInfo);
  }

  render() {
    if (this.state.hasError) {
      return (
        <div className="p-8 text-center space-y-4 font-mono">
          <div className="p-5 bg-charcoal-900 border border-rose-500/30 rounded-xl max-w-lg mx-auto text-left space-y-2">
            <h3 className="text-sm font-bold text-rose-400">View Rendering Notice</h3>
            <p className="text-xs text-zinc-300">{this.state.error?.message || 'Unexpected client exception'}</p>
            <button
              onClick={() => this.setState({ hasError: false, error: null })}
              className="mt-2 px-3 py-1 bg-charcoal-800 hover:bg-charcoal-750 text-xs text-white rounded border border-charcoal-700 font-mono"
            >
              Reset View
            </button>
          </div>
        </div>
      );
    }
    return this.props.children;
  }
}

export default function AppLayout({ children }: { children: ReactNode }) {
  const [mounted, setMounted] = useState(false);

  useEffect(() => {
    setMounted(true);
  }, []);

  if (!mounted) {
    return (
      <div className="h-screen w-screen bg-[#090a0b] flex items-center justify-center text-zinc-400 font-mono text-xs">
        <div className="flex items-center gap-3">
          <div className="w-2 h-2 rounded-full bg-[#38bdf8] animate-ping" />
          <span>Initializing CodeDNA Intelligence Map...</span>
        </div>
      </div>
    );
  }

  return (
    <div className="flex h-screen w-screen overflow-hidden bg-charcoal-950 select-none">
      {/* Sidebar Navigation */}
      <Sidebar />

      {/* Main Content Area */}
      <div className="flex-1 flex flex-col min-w-0 overflow-hidden">
        {/* Top Header */}
        <Header />

        {/* View Content */}
        <main className="flex-1 overflow-y-auto">
          <ErrorBoundary>
            {children}
          </ErrorBoundary>
        </main>
      </div>

      {/* Source Code Inspector Drawer */}
      <SourceDrawer />
    </div>
  );
}
