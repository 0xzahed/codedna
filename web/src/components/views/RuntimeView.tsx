'use client';

import React, { useEffect, useState } from 'react';
import { useAppStore } from '../../lib/store';
import { fetchRuntime, RuntimeSummary, API_BASE } from '../../lib/api';
import { 
  Activity, 
  Clock, 
  Zap, 
  GitMerge, 
  CheckCircle2, 
  Flame, 
  Terminal,
  PlayCircle
} from 'lucide-react';

export default function RuntimeView() {
  const { scanId } = useAppStore();
  const [runtimeData, setRuntimeData] = useState<RuntimeSummary | null>(null);
  const [loading, setLoading] = useState(true);
  const [simulating, setSimulating] = useState(false);

  const loadData = () => {
    setLoading(true);
    fetchRuntime(scanId)
      .then((data) => setRuntimeData(data))
      .catch((err) => console.error(err))
      .finally(() => setLoading(false));
  };

  useEffect(() => {
    loadData();
  }, [scanId]);

  // Allows running a local test trace event to verify runtime ingestion
  const handleSimulateTrace = async () => {
    setSimulating(true);
    try {
      const mockEvents = [
        {
          runId: 'demo-run-' + Date.now(),
          functionName: 'Execute',
          file: 'main.go',
          line: 3,
          durationUs: 450,
          timestampNs: Date.now() * 1000000,
          callerFunction: 'main',
          callerFile: 'cmd/codedna/main.go',
        },
      ];
      await fetch(`${API_BASE}/api/runtime/events`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(mockEvents),
      });
      loadData();
    } catch (err) {
      console.error(err);
    } finally {
      setSimulating(false);
    }
  };

  const comp = runtimeData?.comparison;

  return (
    <div className="p-8 space-y-8 max-w-7xl mx-auto">
      {/* Header */}
      <div className="flex items-center justify-between border-b border-charcoal-750 pb-6">
        <div>
          <h1 className="text-2xl font-bold tracking-tight text-white flex items-center gap-3">
            <Activity className="w-6 h-6 text-emerald-400" />
            <span>Runtime DNA &amp; Execution Timeline</span>
          </h1>
          <p className="text-sm text-zinc-400 mt-1">
            Opt-in function tracing, runtime execution paths, and static vs runtime graph reconciliation.
          </p>
        </div>

        <button
          onClick={handleSimulateTrace}
          disabled={simulating}
          className="flex items-center gap-2 px-3.5 py-1.5 rounded-lg bg-emerald-500/15 border border-emerald-500/40 text-emerald-400 hover:bg-emerald-500/25 text-xs font-mono font-medium transition-all"
        >
          <PlayCircle className="w-3.5 h-3.5" />
          <span>{simulating ? 'Recording...' : 'Send Test Trace Event'}</span>
        </button>
      </div>

      {/* Static vs Runtime Reconciliation KPIs */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4 font-mono">
        <div className="bg-charcoal-850 border border-charcoal-750 rounded-xl p-5 space-y-1">
          <div className="text-xs text-zinc-400">Runtime Verified Coverage</div>
          <div className="text-3xl font-bold text-white">
            {comp?.coveragePercentage ?? 0}%
          </div>
          <div className="text-[11px] text-zinc-400">
            {comp?.observedStaticEdges ?? 0} of {comp?.staticEdgeCount ?? 0} static edges hit
          </div>
        </div>

        <div className="bg-charcoal-850 border border-charcoal-750 rounded-xl p-5 space-y-1">
          <div className="text-xs text-zinc-400">Captured Events</div>
          <div className="text-3xl font-bold text-emerald-400">
            {runtimeData?.totalEvents ?? 0}
          </div>
          <div className="text-[11px] text-zinc-400">
            {runtimeData?.uniqueFunctions ?? 0} unique functions called
          </div>
        </div>

        <div className="bg-charcoal-850 border border-charcoal-750 rounded-xl p-5 space-y-1">
          <div className="text-xs text-zinc-400">Average Duration</div>
          <div className="text-3xl font-bold text-dna-blue">
            {runtimeData?.avgDurationUs ?? 0} <span className="text-sm text-zinc-500 font-normal">µs</span>
          </div>
          <div className="text-[11px] text-zinc-400">Microseconds per function</div>
        </div>

        <div className="bg-charcoal-850 border border-charcoal-750 rounded-xl p-5 space-y-1">
          <div className="text-xs text-zinc-400">Dynamic-Only Calls</div>
          <div className="text-3xl font-bold text-dna-amber">
            {comp?.dynamicOnlyEdges ?? 0}
          </div>
          <div className="text-[11px] text-zinc-400">Discovered at runtime</div>
        </div>
      </div>

      {/* Main Grid: Hot Functions & Execution Paths */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Frequently Executed Functions */}
        <div className="bg-charcoal-850 border border-charcoal-750 rounded-xl p-6 space-y-4">
          <div className="flex items-center justify-between border-b border-charcoal-750 pb-3">
            <div className="flex items-center gap-2 text-white font-semibold text-sm">
              <Zap className="w-4 h-4 text-dna-amber" />
              <span>Frequently Executed Functions</span>
            </div>
            <span className="text-xs text-zinc-400 font-mono">By Call Frequency</span>
          </div>

          {runtimeData?.frequentFunctions && runtimeData.frequentFunctions.length > 0 ? (
            <div className="divide-y divide-charcoal-750/50">
              {runtimeData.frequentFunctions.map((fn, i) => (
                <div key={i} className="py-2.5 flex items-center justify-between font-mono text-xs">
                  <div className="truncate pr-4">
                    <div className="font-semibold text-white truncate">{fn.function}</div>
                    <div className="text-[11px] text-zinc-400 truncate">{fn.file}</div>
                  </div>
                  <div className="text-right shrink-0">
                    <div className="text-dna-blue font-bold">{fn.callCount} calls</div>
                    <div className="text-[10px] text-zinc-400">{fn.avgTimeUs} µs avg</div>
                  </div>
                </div>
              ))}
            </div>
          ) : (
            <div className="py-8 text-center text-xs text-zinc-500 font-mono space-y-2">
              <p>No runtime traces recorded yet.</p>
              <p className="text-[11px] text-zinc-400">
                Run <code>codedna runtime enable</code> or send events to <code>/api/runtime/events</code>.
              </p>
            </div>
          )}
        </div>

        {/* Execution Paths */}
        <div className="bg-charcoal-850 border border-charcoal-750 rounded-xl p-6 space-y-4">
          <div className="flex items-center justify-between border-b border-charcoal-750 pb-3">
            <div className="flex items-center gap-2 text-white font-semibold text-sm">
              <Clock className="w-4 h-4 text-dna-blue" />
              <span>Execution Path Timeline</span>
            </div>
            <span className="text-xs text-zinc-400 font-mono">Sequential Call Traces</span>
          </div>

          {runtimeData?.executionPaths && runtimeData.executionPaths.length > 0 ? (
            <div className="space-y-3 max-h-80 overflow-y-auto">
              {runtimeData.executionPaths.map((p, i) => (
                <div key={i} className="p-3 bg-charcoal-900 border border-charcoal-750 rounded-lg space-y-2 font-mono text-xs">
                  <div className="flex items-center justify-between text-zinc-400 border-b border-charcoal-800 pb-1.5">
                    <span className="text-white font-semibold">{p.traceId}</span>
                    <span className="text-[11px]">{p.totalTimeUs} µs total</span>
                  </div>
                  <div className="space-y-1 pl-2 border-l border-dna-blue/40">
                    {p.steps.map((st, sIdx) => (
                      <div key={sIdx} className="text-zinc-300 text-[11px]">
                        → {st}
                      </div>
                    ))}
                  </div>
                </div>
              ))}
            </div>
          ) : (
            <div className="py-8 text-center text-xs text-zinc-500 font-mono">
              No active execution paths logged.
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
