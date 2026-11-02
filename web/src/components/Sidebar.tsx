'use client';

import React from 'react';
import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { useAppStore, NavTab } from '../lib/store';
import { 
  LayoutDashboard, 
  FolderTree, 
  GitFork, 
  Boxes, 
  Workflow,
  Activity, 
  GitCommit, 
  ShieldCheck, 
  Sparkles, 
  Settings,
  Dna,
  ChevronLeft,
  ChevronRight
} from 'lucide-react';

interface NavItem {
  id: NavTab;
  label: string;
  href: string;
  icon: React.ElementType;
  badge?: string;
}

const navItems: NavItem[] = [
  { id: 'overview', label: 'Overview', href: '/overview', icon: LayoutDashboard },
  { id: 'explorer', label: 'Repository Explorer', href: '/explorer', icon: FolderTree },
  { id: 'graph', label: 'DNA Graph', href: '/graph', icon: GitFork },
  { id: 'deps', label: 'Dependencies', href: '/dependencies', icon: Boxes },
  { id: 'design', label: 'System Design', href: '/design', icon: Workflow, badge: 'New' },
  { id: 'runtime', label: 'Runtime DNA', href: '/runtime', icon: Activity, badge: 'Live' },
  { id: 'impact', label: 'Change Impact', href: '/impact', icon: GitCommit },
  { id: 'arch', label: 'Architecture Health', href: '/architecture', icon: ShieldCheck },
  { id: 'ai', label: 'AI Assistant', href: '/ai', icon: Sparkles },
  { id: 'settings', label: 'Settings', href: '/settings', icon: Settings },
];

export default function Sidebar() {
  const { activeTab, setActiveTab, sidebarCollapsed, toggleSidebar } = useAppStore();
  const pathname = usePathname();

  return (
    <aside 
      className={`border-r border-charcoal-750 bg-charcoal-900 flex flex-col shrink-0 select-none transition-all duration-200 ${
        sidebarCollapsed ? 'w-16' : 'w-64'
      }`}
    >
      {/* Brand Header */}
      <div className="h-16 border-b border-charcoal-750 px-3 flex items-center justify-between">
        <Link href="/" className="flex items-center gap-3 overflow-hidden">
          <div className="w-9 h-9 rounded-lg bg-gradient-to-br from-dna-blue/20 to-dna-purple/20 border border-dna-blue/40 flex items-center justify-center text-dna-blue shrink-0">
            <Dna className="w-5 h-5 animate-pulse" />
          </div>
          {!sidebarCollapsed && (
            <div className="overflow-hidden">
              <div className="font-bold tracking-tight text-white flex items-center gap-1.5 truncate">
                <span>CodeDNA</span>
                <span className="text-[10px] bg-dna-blue/10 text-dna-blue px-1 py-0.2 rounded font-mono border border-dna-blue/30">2.4</span>
              </div>
              <div className="text-[11px] text-zinc-400 truncate">Software Intelligence</div>
            </div>
          )}
        </Link>

        {/* Sidebar Collapse Toggle Button */}
        <button
          onClick={toggleSidebar}
          className="p-1 rounded-lg text-zinc-400 hover:text-white hover:bg-charcoal-800 transition-colors"
          title={sidebarCollapsed ? "Expand Sidebar (Wider Navigation)" : "Collapse Sidebar (Maximum Diagram Width)"}
        >
          {sidebarCollapsed ? <ChevronRight className="w-4 h-4" /> : <ChevronLeft className="w-4 h-4" />}
        </button>
      </div>

      {/* Navigation Links */}
      <nav className="p-2 space-y-1 flex-1 overflow-y-auto">
        {!sidebarCollapsed && (
          <div className="px-3 py-2 text-[10px] font-semibold tracking-wider text-zinc-400 uppercase font-mono">
            Intelligence Views
          </div>
        )}
        {navItems.map((item) => {
          const Icon = item.icon;
          const isCurrentRoute = 
            pathname === item.href || 
            (pathname?.startsWith(item.href) && item.href !== '/') ||
            (item.id === 'overview' && (pathname === '/' || pathname === '')) ||
            (item.id === 'deps' && (pathname === '/deps' || pathname === '/deps/')) ||
            (item.id === 'arch' && (pathname === '/arch' || pathname === '/arch/'));
          const isActive = isCurrentRoute || activeTab === item.id;

          return (
            <Link
              key={item.id}
              href={item.href}
              onClick={() => setActiveTab(item.id)}
              className={`w-full flex items-center ${sidebarCollapsed ? 'justify-center p-2.5' : 'justify-between px-3 py-2.5'} rounded-lg text-xs font-medium transition-all ${
                isActive
                  ? 'bg-charcoal-750 text-white border border-charcoal-600 shadow-sm'
                  : 'text-zinc-400 hover:text-zinc-200 hover:bg-charcoal-800'
              }`}
              title={sidebarCollapsed ? item.label : undefined}
            >
              <div className="flex items-center gap-3">
                <Icon className={`w-4 h-4 shrink-0 ${isActive ? 'text-dna-blue' : 'text-zinc-400'}`} />
                {!sidebarCollapsed && <span className="truncate">{item.label}</span>}
              </div>
              {!sidebarCollapsed && item.badge && (
                <span className="text-[9px] font-mono px-1.5 py-0.5 rounded bg-emerald-500/10 text-emerald-400 border border-emerald-500/30 shrink-0">
                  {item.badge}
                </span>
              )}
            </Link>
          );
        })}
      </nav>

      {/* Local Environment Footer */}
      <div className="p-3 border-t border-charcoal-750 bg-charcoal-950/40">
        {sidebarCollapsed ? (
          <div className="flex justify-center" title="Engine Status: Online (SQLite)">
            <span className="w-2 h-2 rounded-full bg-emerald-400 animate-ping" />
          </div>
        ) : (
          <div>
            <div className="flex items-center justify-between text-[11px] text-zinc-400 font-mono mb-0.5">
              <span>Engine Status</span>
              <span className="flex items-center gap-1.5 text-emerald-400">
                <span className="w-1.5 h-1.5 rounded-full bg-emerald-400 animate-ping"></span>
                Online
              </span>
            </div>
            <div className="text-[10px] text-zinc-400 truncate">
              Local-first SQLite Storage
            </div>
          </div>
        )}
      </div>
    </aside>
  );
}
