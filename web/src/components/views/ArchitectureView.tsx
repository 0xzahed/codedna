'use client';

import React, { useEffect, useState } from 'react';
import { useAppStore } from '../../lib/store';
import { fetchArchitecture, ArchReport } from '../../lib/api';
import { 
  ShieldCheck, 
  AlertTriangle, 
  CheckCircle2, 
  Layers, 
  ArrowRight,
  FileCode,
  Sliders
} from 'lucide-react';

export default function ArchitectureView() {
  const { scanId, setSelectedNode } = useAppStore();
  const [arch, setArch] = useState<ArchReport | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    setLoading(true);
    fetchArchitecture(scanId)
      .then((data) => setArch(data))
      .catch((err) => console.error(err))
      .finally(() => setLoading(false));
  }, [scanId]);

  return (
    <div className="p-8 space-y-8 max-w-7xl mx-auto">
      {/* Header */}
      <div className="border-b border-charcoal-750 pb-6">
        <h1 className="text-2xl font-bold tracking-tight text-white flex items-center gap-3">
          <ShieldCheck className="w-6 h-6 text-dna-purple" />
          <span>Architecture Health &amp; Drift Detection</span>
        </h1>
        <p className="text-sm text-zinc-400 mt-1">
          Enforce architectural boundaries, clean layers, and detect dependency violations.
        </p>
      </div>

      {/* Layer Health Status Ribbon */}
      <div className="bg-charcoal-850 border border-charcoal-750 rounded-xl p-6 grid grid-cols-1 md:grid-cols-3 gap-6 font-mono">
        <div>
          <div className="text-xs text-zinc-400">Layer Health Score</div>
          <div className="text-4xl font-bold text-white mt-1">
            {arch?.layerHealthScore.toFixed(1) ?? '100.0'} <span className="text-xs text-zinc-500 font-normal">/ 100</span>
          </div>
          <p className="text-[11px] text-zinc-400 mt-2">
            Weighted penalty per boundary violation
          </p>
        </div>

        <div>
          <div className="text-xs text-zinc-400">Active Rules</div>
          <div className="text-3xl font-bold text-dna-blue mt-1">
            {arch?.ruleCount ?? 3}
          </div>
          <p className="text-[11px] text-zinc-400 mt-2">
            Loaded from .codedna/architecture.json
          </p>
        </div>

        <div>
          <div className="text-xs text-zinc-400">Boundary Violations</div>
          <div className={`text-3xl font-bold mt-1 ${(arch?.violations.length || 0) > 0 ? 'text-dna-rose' : 'text-emerald-400'}`}>
            {arch?.violations.length ?? 0}
          </div>
          <p className="text-[11px] text-zinc-400 mt-2">
            {arch?.violations.length === 0 ? 'All layer rules respected' : 'Action required to decouple'}
          </p>
        </div>
      </div>

      {/* Detected Violations Table */}
      <div className="bg-charcoal-850 border border-charcoal-750 rounded-xl p-6 space-y-4 font-mono">
        <div className="flex items-center justify-between border-b border-charcoal-750 pb-3">
          <h3 className="text-sm font-semibold text-white flex items-center gap-2">
            <AlertTriangle className="w-4 h-4 text-dna-rose" />
            <span>Architectural Boundary Violations</span>
          </h3>
          <span className="text-xs text-zinc-400">{arch?.violations.length ?? 0} violations found</span>
        </div>

        {arch?.violations && arch.violations.length > 0 ? (
          <div className="divide-y divide-charcoal-750/50">
            {arch.violations.map((v, i) => (
              <div
                key={i}
                onClick={() => setSelectedNode({ id: v.fromFile, type: 'file', label: v.ruleName, file: v.fromFile, line: v.line })}
                className="py-3 flex items-start justify-between hover:bg-charcoal-800 px-3 rounded-lg cursor-pointer transition-all"
              >
                <div className="space-y-1 pr-4">
                  <div className="flex items-center gap-2">
                    <span className="text-xs font-bold text-white">{v.ruleName}</span>
                    <span className="text-[10px] px-2 py-0.5 rounded bg-rose-500/10 text-rose-400 border border-rose-500/30 uppercase">
                      {v.severity}
                    </span>
                  </div>
                  <div className="text-xs text-zinc-300">
                    <span className="text-dna-blue">{v.fromFile}:{v.line || 1}</span>
                    <span className="text-zinc-500 mx-2">→ illegally depends on →</span>
                    <span className="text-dna-rose">{v.toFile}</span>
                  </div>
                  <div className="text-[11px] text-zinc-400">{v.reason}</div>
                </div>

                <ArrowRight className="w-4 h-4 text-zinc-500 mt-2 shrink-0" />
              </div>
            ))}
          </div>
        ) : (
          <div className="py-10 text-center space-y-2">
            <CheckCircle2 className="w-8 h-8 text-emerald-400 mx-auto" />
            <p className="text-xs text-white">Clean Architecture Boundaries</p>
            <p className="text-[11px] text-zinc-500 max-w-sm mx-auto">
              No forbidden imports or layer boundary violations detected.
            </p>
          </div>
        )}
      </div>
    </div>
  );
}
