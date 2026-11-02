'use client';

import React, { useEffect, useState } from 'react';
import { useAppStore } from '../../lib/store';
import { fetchDependencies, fetchSymbols, SymbolItem } from '../../lib/api';
import { 
  Boxes, 
  Package, 
  ExternalLink, 
  Search, 
  Layers, 
  ShieldCheck, 
  CheckCircle2,
  Terminal
} from 'lucide-react';

export default function DependencyView() {
  const { scanId } = useAppStore();
  const [data, setData] = useState<{ dependencies?: Record<string, string[]>; frameworks?: string[] }>({});
  const [search, setSearch] = useState('');
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    setLoading(true);
    fetchDependencies(scanId)
      .then((res) => setData(res || {}))
      .catch((err) => console.error(err))
      .finally(() => setLoading(false));
  }, [scanId]);

  const deps = data.dependencies || {};
  const frameworks = data.frameworks || [];

  return (
    <div className="p-6 md:p-8 space-y-8 w-full max-w-[1920px] mx-auto">
      {/* Header */}
      <div className="border-b border-charcoal-750 pb-6">
        <h1 className="text-2xl font-bold tracking-tight text-white flex items-center gap-3">
          <Boxes className="w-6 h-6 text-dna-blue" />
          <span>Dependency & Framework Intelligence</span>
        </h1>
        <p className="text-sm text-zinc-400 mt-1">
          Manifest discovery across npm, pip, go.mod, Composer, and Maven.
        </p>
      </div>

      {/* Detected Frameworks Banner */}
      {frameworks.length > 0 && (
        <div className="bg-charcoal-850 border border-charcoal-750 rounded-xl p-5 space-y-3">
          <div className="text-xs font-mono text-zinc-400 uppercase tracking-wider flex items-center gap-2">
            <CheckCircle2 className="w-4 h-4 text-emerald-400" />
            <span>Detected Framework Stacks</span>
          </div>
          <div className="flex flex-wrap gap-2">
            {frameworks.map((fw) => (
              <span
                key={fw}
                className="px-3 py-1.5 rounded-lg bg-dna-blue/10 border border-dna-blue/30 text-dna-blue font-mono text-xs font-semibold flex items-center gap-2"
              >
                <Terminal className="w-3.5 h-3.5" />
                {fw}
              </span>
            ))}
          </div>
        </div>
      )}

      {/* Search Filter */}
      <div className="relative max-w-md">
        <Search className="w-3.5 h-3.5 text-zinc-500 absolute left-3 top-1/2 -translate-y-1/2" />
        <input
          type="text"
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          placeholder="Search external packages..."
          className="w-full bg-charcoal-850 border border-charcoal-750 rounded-lg pl-9 pr-4 py-2 text-xs text-white placeholder-zinc-500 focus:outline-none focus:border-dna-blue/50 font-mono"
        />
      </div>

      {/* Dependencies by Ecosystem */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        {Object.entries(deps).map(([ecosystem, pkgList]) => {
          const list = Array.isArray(pkgList) ? pkgList : [];
          const filtered = list.filter((p) => p.toLowerCase().includes(search.toLowerCase()));
          return (
            <div key={ecosystem} className="bg-charcoal-850 border border-charcoal-750 rounded-xl p-6 space-y-4">
              <div className="flex items-center justify-between border-b border-charcoal-750 pb-3">
                <div className="flex items-center gap-2.5">
                  <Package className="w-4 h-4 text-dna-emerald" />
                  <h3 className="font-bold text-white text-sm uppercase font-mono">{ecosystem} Packages</h3>
                </div>
                <span className="text-xs font-mono text-zinc-400">{filtered.length} declared</span>
              </div>

              {filtered.length > 0 ? (
                <div className="divide-y divide-charcoal-750/50 max-h-80 overflow-y-auto">
                  {filtered.map((pkg, i) => (
                    <div key={i} className="py-2.5 flex items-center justify-between font-mono text-xs">
                      <span className="text-zinc-200">{pkg}</span>
                      <span className="text-[10px] text-zinc-500 bg-charcoal-800 px-2 py-0.5 rounded border border-charcoal-700">
                        direct
                      </span>
                    </div>
                  ))}
                </div>
              ) : (
                <div className="py-6 text-center text-xs text-zinc-500 font-mono">
                  No packages match search.
                </div>
              )}
            </div>
          );
        })}
      </div>
    </div>
  );
}
