'use client';

import React, { useEffect, useState } from 'react';
import { useAppStore } from '../../lib/store';
import { fetchImpact, ImpactReport } from '../../lib/api';
import { 
  GitCommit, 
  AlertTriangle, 
  ShieldAlert, 
  CheckCircle2, 
  ArrowRight, 
  Terminal, 
  FileCode,
  Share2
} from 'lucide-react';

export default function ImpactView() {
  const { scanId, setSelectedNode } = useAppStore();
  const [report, setReport] = useState<ImpactReport | null>(null);
  const [baseRef, setBaseRef] = useState('main');
  const [loading, setLoading] = useState(true);

  const loadImpact = (base: string) => {
    setLoading(true);
    fetchImpact(scanId, base)
      .then((data) => setReport(data))
      .catch((err) => console.error(err))
      .finally(() => setLoading(false));
  };

  useEffect(() => {
    loadImpact(baseRef);
  }, [scanId]);

  const handleApplyBase = (e: React.FormEvent) => {
    e.preventDefault();
    loadImpact(baseRef);
  };

  const riskColor = 
    report?.riskLevel === 'high' ? 'text-dna-rose' :
    report?.riskLevel === 'medium' ? 'text-dna-amber' : 'text-emerald-400';

  return (
    <div className="p-8 space-y-8 max-w-7xl mx-auto">
      {/* Header & Diff Selector */}
      <div className="flex items-center justify-between border-b border-charcoal-750 pb-6">
        <div>
          <h1 className="text-2xl font-bold tracking-tight text-white flex items-center gap-3">
            <GitCommit className="w-6 h-6 text-dna-blue" />
            <span>Change Impact Analyzer</span>
          </h1>
          <p className="text-sm text-zinc-400 mt-1">
            Blast radius, upstream caller propagation, affected APIs, and test mitigation.
          </p>
        </div>

        {/* Base ref form */}
        <form onSubmit={handleApplyBase} className="flex items-center gap-2">
          <span className="text-xs font-mono text-zinc-400">Diff Base:</span>
          <input
            type="text"
            value={baseRef}
            onChange={(e) => setBaseRef(e.target.value)}
            placeholder="e.g. main, HEAD~1"
            className="w-36 bg-charcoal-850 border border-charcoal-750 rounded-lg px-3 py-1.5 text-xs text-white font-mono focus:outline-none focus:border-dna-blue"
          />
          <button
            type="submit"
            className="px-3 py-1.5 bg-charcoal-800 hover:bg-charcoal-750 border border-charcoal-700 text-xs font-mono text-white rounded-lg transition-colors"
          >
            Analyze
          </button>
        </form>
      </div>

      {/* Risk Assessment Score Card */}
      <div className="bg-charcoal-850 border border-charcoal-750 rounded-xl p-6 grid grid-cols-1 md:grid-cols-4 gap-6 font-mono">
        <div>
          <div className="text-xs text-zinc-400">Risk Assessment</div>
          <div className={`text-4xl font-bold mt-1 ${riskColor}`}>
            {report?.riskScore != null ? report.riskScore.toFixed(1) : '0.0'}
            <span className="text-xs text-zinc-500 font-normal uppercase ml-2">
              ({report?.riskLevel ?? 'info'})
            </span>
          </div>
          <p className="text-[11px] text-zinc-400 mt-2">
            Based on blast radius, cyclomatic complexity, and test coverage.
          </p>
        </div>

        <div>
          <div className="text-xs text-zinc-400">Changed Files</div>
          <div className="text-2xl font-bold text-white mt-1">
            {report?.changedFiles?.length ?? 0}
          </div>
          <div className="text-[11px] text-zinc-400 mt-2">
            {report?.changedSymbols?.length ?? 0} AST symbols modified
          </div>
        </div>

        <div>
          <div className="text-xs text-zinc-400">Upstream Blast Radius</div>
          <div className="text-2xl font-bold text-dna-blue mt-1">
            {report?.upstreamAffected?.length ?? 0}
          </div>
          <div className="text-[11px] text-zinc-400 mt-2">
            Dependent callers impacted
          </div>
        </div>

        <div>
          <div className="text-xs text-zinc-400">Affected APIs &amp; Tests</div>
          <div className="text-2xl font-bold text-dna-amber mt-1">
            {report?.affectedEndpoints?.length ?? 0} / {report?.affectedTests?.length ?? 0}
          </div>
          <div className="text-[11px] text-zinc-400 mt-2">
            Endpoints / relevant tests
          </div>
        </div>
      </div>

      {/* Main Grid: Modified Entities & Blast Propagation */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Modified Symbols */}
        <div className="bg-charcoal-850 border border-charcoal-750 rounded-xl p-6 space-y-4 font-mono">
          <div className="flex items-center justify-between border-b border-charcoal-750 pb-3">
            <h3 className="text-sm font-semibold text-white flex items-center gap-2">
              <FileCode className="w-4 h-4 text-dna-blue" />
              <span>Directly Modified Symbols</span>
            </h3>
            <span className="text-xs text-zinc-400">{report?.changedSymbols?.length ?? 0} found</span>
          </div>

          {report?.changedSymbols && report.changedSymbols.length > 0 ? (
            <div className="divide-y divide-charcoal-750/50">
              {report.changedSymbols.map((cs, i) => (
                <div
                  key={i}
                  onClick={() => setSelectedNode({
                    id: cs.symbol.id,
                    type: cs.symbol.kind,
                    label: cs.symbol.name,
                    file: cs.symbol.file,
                    line: cs.symbol.line,
                    complexity: cs.symbol.complexity,
                    exported: cs.symbol.exported
                  })}
                  className="py-2.5 flex items-center justify-between hover:bg-charcoal-800 px-2 rounded-lg cursor-pointer transition-all"
                >
                  <div className="truncate pr-4">
                    <div className="text-xs font-semibold text-white truncate">{cs.symbol.qualifiedName}</div>
                    <div className="text-[11px] text-zinc-400 truncate">{cs.symbol.file}:{cs.symbol.line}</div>
                  </div>
                  <span className="text-[10px] px-2 py-0.5 rounded bg-blue-500/10 text-blue-400 border border-blue-500/30">
                    {cs.changeType}
                  </span>
                </div>
              ))}
            </div>
          ) : (
            <div className="py-8 text-center text-xs text-zinc-500">
              No code modifications detected against base ref '{baseRef}'.
            </div>
          )}
        </div>

        {/* Affected APIs and Recommended Tests */}
        <div className="bg-charcoal-850 border border-charcoal-750 rounded-xl p-6 space-y-6 font-mono">
          {/* Affected APIs */}
          <div className="space-y-3">
            <h3 className="text-sm font-semibold text-white flex items-center gap-2 border-b border-charcoal-750 pb-2">
              <AlertTriangle className="w-4 h-4 text-dna-amber" />
              <span>Affected API Endpoints</span>
            </h3>
            {report?.affectedEndpoints && report.affectedEndpoints.length > 0 ? (
              <div className="space-y-1.5">
                {report.affectedEndpoints.map((ep, i) => (
                  <div key={i} className="text-xs p-2 bg-charcoal-900 border border-amber-500/20 text-amber-300 rounded">
                    ⚠ {ep}
                  </div>
                ))}
              </div>
            ) : (
              <div className="text-xs text-zinc-500">No public API endpoints in direct blast radius.</div>
            )}
          </div>

          {/* Recommended Tests */}
          <div className="space-y-3">
            <h3 className="text-sm font-semibold text-white flex items-center gap-2 border-b border-charcoal-750 pb-2">
              <CheckCircle2 className="w-4 h-4 text-emerald-400" />
              <span>Recommended Tests to Execute</span>
            </h3>
            {report?.affectedTests && report.affectedTests.length > 0 ? (
              <div className="space-y-1.5">
                {report.affectedTests.map((t, i) => (
                  <div key={i} className="text-xs p-2 bg-charcoal-900 border border-emerald-500/20 text-emerald-300 rounded">
                    ✓ {t}
                  </div>
                ))}
              </div>
            ) : (
              <div className="text-xs text-zinc-500">No relevant test suites identified.</div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}
