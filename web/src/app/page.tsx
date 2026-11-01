'use client';

import React from 'react';
import { useAppStore } from '../lib/store';
import Sidebar from '../components/Sidebar';
import Header from '../components/Header';
import SourceDrawer from '../components/SourceDrawer';

import OverviewView from '../components/views/OverviewView';
import ExplorerView from '../components/views/ExplorerView';
import GraphView from '../components/views/GraphView';
import DependencyView from '../components/views/DependencyView';
import RuntimeView from '../components/views/RuntimeView';
import ImpactView from '../components/views/ImpactView';
import ArchitectureView from '../components/views/ArchitectureView';
import AIView from '../components/views/AIView';
import SettingsView from '../components/views/SettingsView';

export default function Home() {
  const { activeTab } = useAppStore();

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
          {activeTab === 'overview' && <OverviewView />}
          {activeTab === 'explorer' && <ExplorerView />}
          {activeTab === 'graph' && <GraphView />}
          {activeTab === 'deps' && <DependencyView />}
          {activeTab === 'runtime' && <RuntimeView />}
          {activeTab === 'impact' && <ImpactView />}
          {activeTab === 'arch' && <ArchitectureView />}
          {activeTab === 'ai' && <AIView />}
          {activeTab === 'settings' && <SettingsView />}
        </main>
      </div>

      {/* Source Code Inspector Drawer */}
      <SourceDrawer />
    </div>
  );
}
