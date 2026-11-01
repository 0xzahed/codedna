'use client';

import React, { useState, useEffect, Component, ErrorInfo, ReactNode } from 'react';
import dynamic from 'next/dynamic';
import { useAppStore } from '../lib/store';
import Sidebar from '../components/Sidebar';
import Header from '../components/Header';
import SourceDrawer from '../components/SourceDrawer';

import OverviewView from '../components/views/OverviewView';
import ExplorerView from '../components/views/ExplorerView';
import DependencyView from '../components/views/DependencyView';
import RuntimeView from '../components/views/RuntimeView';
import ImpactView from '../components/views/ImpactView';
import ArchitectureView from '../components/views/ArchitectureView';
import AIView from '../components/views/AIView';
import SettingsView from '../components/views/SettingsView';

// Dynamically load React Flow on client-side only
const GraphView = dynamic(() => import('../components/views/GraphView'), {
  ssr: false,
  loading: () => (
    <div className="h-full flex items-center justify-center text-zinc-500 font-mono text-xs">
      Loading interactive DNA canvas...
    </div>
  ),
});

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
          <div className="p-4 bg-charcoal-900 border border-rose-500/30 rounded-xl max-w-lg mx-auto text-left space-y-2">
            <h3 className="text-sm font-bold text-rose-400">View Rendering Notice</h3>
            <p className="text-xs text-zinc-300">{this.state.error?.message || 'Unexpected client exception'}</p>
            <button
              onClick={() => this.setState({ hasError: false, error: null })}
              className="mt-2 px-3 py-1 bg-charcoal-800 hover:bg-charcoal-750 text-xs text-white rounded border border-charcoal-700"
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

export default function Home() {
  const { activeTab } = useAppStore();
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

        {/* Dynamic Viewport */}
        <main className="flex-1 overflow-y-auto">
          <ErrorBoundary>
            {activeTab === 'overview' && <OverviewView />}
            {activeTab === 'explorer' && <ExplorerView />}
            {activeTab === 'graph' && <GraphView />}
            {activeTab === 'deps' && <DependencyView />}
            {activeTab === 'runtime' && <RuntimeView />}
            {activeTab === 'impact' && <ImpactView />}
            {activeTab === 'arch' && <ArchitectureView />}
            {activeTab === 'ai' && <AIView />}
            {activeTab === 'settings' && <SettingsView />}
          </ErrorBoundary>
        </main>
      </div>

      {/* Source Code Inspector Drawer */}
      <SourceDrawer />
    </div>
  );
}
