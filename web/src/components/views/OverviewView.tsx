'use client';

import React, { useEffect, useState } from 'react';
import { useAppStore } from '../../lib/store';
import { 
  fetchRepos, 
  fetchIssues, 
  fetchArchitecture, 
  IssueItem, 
  ArchReport, 
  RepoStats,
  fetchDependencies 
} from '../../lib/api';
import { 
  Code2, 
  Files, 
  Share2, 
  AlertTriangle, 
  Flame, 
  ShieldCheck, 
  Layers, 
  ArrowRight,
  TrendingUp,
  Cpu
} from 'lucide-react';

export default function OverviewView() {
  const { scanId, setActiveTab, setSelectedNode } = useAppStore();
  const [stats, setStats] = useState<RepoStats | null>(null);
  const [issues, setIssues] = useState<IssueItem[]>([]);
  const [arch, setArch] = useState<ArchReport | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    setLoading(true);
    Promise.all([
      fetchDependencies(scanId),
      fetchIssues(scanId),
      fetchArchitecture(scanId),
    ])
      .then(([depData, issueData, archData]) => {
        setIssues(issueData || []);
        setArch(archData || null);
      })
      .catch((err) => console.error(err))
      .finally(() => setLoading(false));

    // Also fetch repos for stats
    fetchRepos()
      .then((repos) => {
        if (repos?.length > 0 && repos[0]?.git) {
          // stats populated if available
        }
      })
      .catch((err) => console.error(err));
  }, [scanId]);

  const cycles = (issues || []).filter((i) => i.kind === 'circular-dependency');
  const hotspots = (issues || []).filter((i) => i.kind === 'complexity-hotspot').slice(0, 8);
  const deadCode = (issues || []).filter((i) => i.kind === 'dead-code').slice(0, 8);

  const healthScore = Math.max(0, 100 - (cycles.length * 15) - (hotspots.length * 2) - (((arch?.violations?.length) || 0) * 4));

  return (
    <div className="p-8 space-y-8 max-w-7xl mx-auto">
      {/* Welcome Banner */}
      <div className="flex items-center justify-between border-b border-charcoal-750 pb-6">
        <div>
          <h1 className="text-2xl font-bold tracking-tight text-white flex items-center gap-3">
            Codebase Intelligence Overview
          </h1>
          <p className="text-sm text-zinc-400 mt-1">
            Real-time static topology, architectural boundaries, and software DNA metrics.
          </p>
        </div>
        <div className="flex items-center gap-3">
          <button
            onClick={() => setActiveTab('graph')}
            className="flex items-center gap-2 px-4 py-2 rounded-lg bg-dna-blue text-charcoal-950 font-semibold text-xs hover:bg-sky-300 transition-all shadow-md font-mono"
          >
            <span>Open DNA Graph</span>
            <ArrowRight className="w-4 h-4" />
          </button>
        </div>
      </div>

      {/* KPI Cards Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        {/* Health Card */}
        <div className="bg-charcoal-850 border border-charcoal-750 rounded-xl p-5 space-y-2 relative overflow-hidden group hover:border-charcoal-600 transition-all">
          <div className="flex items-center justify-between text-zinc-400">
            <span className="text-xs font-mono">System Health</span>
            <ShieldCheck className={`w-4 h-4 ${healthScore > 80 ? 'text-emerald-400' : 'text-amber-400'}`} />
          </div>
          <div className="text-3xl font-bold text-white font-mono">
            {healthScore.toFixed(0)} <span className="text-sm text-zinc-500 font-normal">/ 100</span>
          </div>
          <div className="text-[11px] text-zinc-400 flex items-center gap-1.5 pt-1">
            <TrendingUp className="w-3.5 h-3.5 text-emerald-400" />
            <span>Layer integrity: {arch?.layerHealthScore ?? 100}%</span>
          </div>
        </div>

        {/* Circular Dependencies Card */}
        <div className="bg-charcoal-850 border border-charcoal-750 rounded-xl p-5 space-y-2 relative overflow-hidden group hover:border-charcoal-600 transition-all">
          <div className="flex items-center justify-between text-zinc-400">
            <span className="text-xs font-mono">Circular Cycles</span>
            <AlertTriangle className={`w-4 h-4 ${cycles.length > 0 ? 'text-dna-rose' : 'text-emerald-400'}`} />
          </div>
          <div className="text-3xl font-bold text-white font-mono">
            {cycles.length}
          </div>
          <div className="text-[11px] text-zinc-400 pt-1">
            {cycles.length === 0 ? '✓ Zero cyclic dependencies' : `⚠ ${cycles.length} cycle(s) need decoupling`}
          </div>
        </div>

        {/* Complexity Hotspots Card */}
        <div className="bg-charcoal-850 border border-charcoal-750 rounded-xl p-5 space-y-2 relative overflow-hidden group hover:border-charcoal-600 transition-all">
          <div className="flex items-center justify-between text-zinc-400">
            <span className="text-xs font-mono">Complexity Hotspots</span>
            <Flame className="w-4 h-4 text-dna-amber" />
          </div>
          <div className="text-3xl font-bold text-white font-mono">
            {hotspots.length}
          </div>
          <div className="text-[11px] text-zinc-400 pt-1">
            Symbols with cyclomatic complexity &gt; 15
          </div>
        </div>

        {/* Architecture Violations Card */}
        <div className="bg-charcoal-850 border border-charcoal-750 rounded-xl p-5 space-y-2 relative overflow-hidden group hover:border-charcoal-600 transition-all">
          <div className="flex items-center justify-between text-zinc-400">
            <span className="text-xs font-mono">Layer Violations</span>
            <Layers className="w-4 h-4 text-dna-purple" />
          </div>
          <div className="text-3xl font-bold text-white font-mono">
            {arch?.violations?.length || 0}
          </div>
          <div className="text-[11px] text-zinc-400 pt-1">
            Against {arch?.ruleCount || 3} configured boundaries
          </div>
        </div>
      </div>

      {/* Main Content Grid: Hotspots & Cycles */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Top Complexity Hotspots */}
        <div className="bg-charcoal-850 border border-charcoal-750 rounded-xl p-6 space-y-4">
          <div className="flex items-center justify-between border-b border-charcoal-750 pb-3">
            <div className="flex items-center gap-2 text-white font-semibold text-sm">
              <Flame className="w-4 h-4 text-dna-amber" />
              <span>Complexity Hotspots</span>
            </div>
            <span className="text-xs text-zinc-400 font-mono">Top High-Branching Symbols</span>
          </div>

          {hotspots.length > 0 ? (
            <div className="divide-y divide-charcoal-750">
              {hotspots.map((h, i) => (
                <div
                  key={i}
                  onClick={() => setSelectedNode({ id: h.symbolId || '', type: 'function', label: h.title, file: h.file })}
                  className="py-2.5 flex items-center justify-between hover:bg-charcoal-800 px-2 rounded-lg cursor-pointer transition-all"
                >
                  <div className="truncate pr-4">
                    <div className="text-xs font-mono font-medium text-white truncate">{h.title}</div>
                    <div className="text-[11px] font-mono text-zinc-400 truncate">{h.file}</div>
                  </div>
                  <div className="flex items-center gap-2 shrink-0">
                    <span className="text-[10px] font-mono px-2 py-0.5 rounded bg-amber-500/10 text-amber-400 border border-amber-500/30">
                      score: {h.detail?.complexity ?? 15}
                    </span>
                    <ArrowRight className="w-3.5 h-3.5 text-zinc-500" />
                  </div>
                </div>
              ))}
            </div>
          ) : (
            <div className="py-8 text-center text-xs text-zinc-500 font-mono">
              No extreme complexity hotspots detected.
            </div>
          )}
        </div>

        {/* Circular Dependency Chains & Architecture */}
        <div className="bg-charcoal-850 border border-charcoal-750 rounded-xl p-6 space-y-4">
          <div className="flex items-center justify-between border-b border-charcoal-750 pb-3">
            <div className="flex items-center gap-2 text-white font-semibold text-sm">
              <AlertTriangle className="w-4 h-4 text-dna-rose" />
              <span>Circular Dependencies</span>
            </div>
            <span className="text-xs text-zinc-400 font-mono">Tarjan SCC Detection</span>
          </div>

          {cycles.length > 0 ? (
            <div className="space-y-3">
              {cycles.map((c, i) => (
                <div key={i} className="p-3 rounded-lg bg-charcoal-900 border border-rose-500/20 space-y-1.5 font-mono text-xs">
                  <div className="text-rose-400 font-semibold">{c.title}</div>
                  <div className="text-[11px] text-zinc-400 break-all bg-charcoal-950 p-2 rounded">
                    {c.detail?.chain || 'A -> B -> A'}
                  </div>
                </div>
              ))}
            </div>
          ) : (
            <div className="py-8 text-center space-y-2">
              <div className="w-10 h-10 rounded-full bg-emerald-500/10 text-emerald-400 flex items-center justify-center mx-auto">
                <ShieldCheck className="w-5 h-5" />
              </div>
              <p className="text-xs text-zinc-300 font-mono">No Circular Dependencies</p>
              <p className="text-[11px] text-zinc-500 max-w-xs mx-auto">
                All module dependencies form a clean Directed Acyclic Graph (DAG).
              </p>
            </div>
          )}
        </div>
      </div>

      {/* Dead Code Candidates Table */}
      <div className="bg-charcoal-850 border border-charcoal-750 rounded-xl p-6 space-y-4">
        <div className="flex items-center justify-between border-b border-charcoal-750 pb-3">
          <div className="flex items-center gap-2 text-white font-semibold text-sm">
            <Code2 className="w-4 h-4 text-zinc-400" />
            <span>Dead Code Candidates (Unreferenced Internals)</span>
          </div>
          <span className="text-xs text-zinc-400 font-mono">Uncalled internal functions with 0 inbound edges</span>
        </div>

        {deadCode.length > 0 ? (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-3">
            {deadCode.map((d, i) => (
              <div
                key={i}
                onClick={() => setSelectedNode({ id: d.symbolId || '', type: 'function', label: d.title, file: d.file })}
                className="p-3 bg-charcoal-900 border border-charcoal-750 rounded-lg cursor-pointer hover:border-charcoal-600 transition-all font-mono"
              >
                <div className="text-xs font-semibold text-zinc-200 truncate">{(d?.title || '').replace('Dead code candidate: ', '').replace('Unchecked reference: ', '')}</div>
                <div className="text-[10px] text-zinc-400 truncate mt-1">{d.file}</div>
              </div>
            ))}
          </div>
        ) : (
          <div className="py-6 text-center text-xs text-zinc-500 font-mono">
            No unreferenced internal functions found.
          </div>
        )}
      </div>
    </div>
  );
}
