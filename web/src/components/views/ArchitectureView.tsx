'use client';

import React, { useEffect, useState } from 'react';
import { useAppStore } from '../../lib/store';
import { 
  fetchArchitecture, 
  addArchRule, 
  deleteArchRule, 
  resetArchRules, 
  ArchReport, 
  ArchRule 
} from '../../lib/api';
import { 
  ShieldCheck, 
  ShieldAlert,
  AlertTriangle, 
  CheckCircle2, 
  Layers, 
  ArrowRight, 
  FileCode, 
  Sliders, 
  Plus, 
  Trash2, 
  RotateCcw, 
  Ban, 
  Sparkles,
  RefreshCw,
  Info
} from 'lucide-react';

export default function ArchitectureView() {
  const { scanId, setSelectedNode } = useAppStore();
  const [arch, setArch] = useState<ArchReport | null>(null);
  const [loading, setLoading] = useState(true);
  const [actionLoading, setActionLoading] = useState(false);
  const [showAddModal, setShowAddModal] = useState(false);

  // New Rule Form State
  const [name, setName] = useState('');
  const [fromPattern, setFromPattern] = useState('');
  const [toPattern, setToPattern] = useState('');
  const [forbidden, setForbidden] = useState(true);
  const [severity, setSeverity] = useState<'high' | 'medium' | 'low'>('high');
  const [description, setDescription] = useState('');
  const [formError, setFormError] = useState('');

  const loadData = () => {
    setLoading(true);
    fetchArchitecture(scanId)
      .then((data) => setArch(data))
      .catch((err) => console.error(err))
      .finally(() => setLoading(false));
  };

  useEffect(() => {
    loadData();
  }, [scanId]);

  const handleApplyPreset = (preset: {
    name: string;
    from: string;
    to: string;
    sev: 'high' | 'medium' | 'low';
    desc: string;
  }) => {
    setName(preset.name);
    setFromPattern(preset.from);
    setToPattern(preset.to);
    setSeverity(preset.sev);
    setDescription(preset.desc);
    setForbidden(true);
    setFormError('');
  };

  const handleAddRule = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!name.trim()) {
      setFormError('Rule name is required');
      return;
    }
    if (!fromPattern.trim() || !toPattern.trim()) {
      setFormError('Both source and target pattern are required');
      return;
    }

    try {
      setActionLoading(true);
      setFormError('');
      const updated = await addArchRule(scanId, {
        name: name.trim(),
        fromPattern: fromPattern.trim(),
        toPattern: toPattern.trim(),
        forbidden,
        severity,
        description: description.trim() || `${fromPattern} must not depend on ${toPattern}`,
      });
      setArch(updated);
      setShowAddModal(false);
      // Reset form
      setName('');
      setFromPattern('');
      setToPattern('');
      setDescription('');
    } catch (err: any) {
      setFormError(err.message || 'Failed to save rule');
    } finally {
      setActionLoading(false);
    }
  };

  const handleDeleteRule = async (ruleId: string) => {
    if (!confirm(`Delete architecture rule "${ruleId}"?`)) return;
    try {
      setActionLoading(true);
      const updated = await deleteArchRule(scanId, ruleId);
      setArch(updated);
    } catch (err) {
      console.error(err);
    } finally {
      setActionLoading(false);
    }
  };

  const handleResetRules = async () => {
    if (!confirm('Reset all architecture rules to Clean Architecture defaults?')) return;
    try {
      setActionLoading(true);
      const updated = await resetArchRules(scanId);
      setArch(updated);
    } catch (err) {
      console.error(err);
    } finally {
      setActionLoading(false);
    }
  };

  return (
    <div className="p-8 space-y-8 max-w-7xl mx-auto">
      {/* Header */}
      <div className="border-b border-charcoal-750 pb-6 flex flex-col md:flex-row md:items-center justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold tracking-tight text-white flex items-center gap-3">
            <ShieldCheck className="w-6 h-6 text-dna-purple" />
            <span>Architecture Health &amp; Drift Detection</span>
          </h1>
          <p className="text-sm text-zinc-400 mt-1">
            Enforce architectural boundaries, clean layers, and detect dependency violations across your codebase.
          </p>
        </div>

        <div className="flex items-center gap-3">
          <button
            onClick={handleResetRules}
            disabled={actionLoading || loading}
            title="Restore default clean architecture rules"
            className="flex items-center gap-2 px-3 py-1.5 text-xs font-medium text-zinc-400 hover:text-white bg-charcoal-850 hover:bg-charcoal-750 border border-charcoal-750 rounded-lg transition-colors"
          >
            <RotateCcw className="w-3.5 h-3.5" />
            <span>Reset Rules</span>
          </button>

          <button
            onClick={() => setShowAddModal(true)}
            className="flex items-center gap-2 px-4 py-2 text-xs font-semibold text-white bg-dna-blue hover:bg-dna-blue/90 rounded-lg shadow-sm transition-all"
          >
            <Plus className="w-4 h-4" />
            <span>Add Custom Rule</span>
          </button>
        </div>
      </div>

      {/* Layer Health Status Ribbon */}
      <div className="bg-charcoal-850 border border-charcoal-750 rounded-xl p-6 grid grid-cols-1 md:grid-cols-4 gap-6 font-mono">
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
          <div className="text-xs text-zinc-400">Active Rules Enforced</div>
          <div className="text-3xl font-bold text-dna-blue mt-1">
            {arch?.ruleCount ?? arch?.rules?.length ?? 3}
          </div>
          <p className="text-[11px] text-zinc-400 mt-2">
            Persisted in .codedna/architecture.json
          </p>
        </div>

        <div>
          <div className="text-xs text-zinc-400">Boundary Violations</div>
          <div className={`text-3xl font-bold mt-1 ${(arch?.violations?.length || 0) > 0 ? 'text-dna-rose' : 'text-emerald-400'}`}>
            {arch?.violations?.length ?? 0}
          </div>
          <p className="text-[11px] text-zinc-400 mt-2">
            {(arch?.violations?.length || 0) === 0 ? 'All layer rules respected' : 'Action required to decouple'}
          </p>
        </div>

        <div>
          <div className="text-xs text-zinc-400">Circular Dependencies</div>
          <div className={`text-3xl font-bold mt-1 ${(arch?.circularCount || 0) > 0 ? 'text-amber-400' : 'text-emerald-400'}`}>
            {arch?.circularCount ?? 0}
          </div>
          <p className="text-[11px] text-zinc-400 mt-2">
            {(arch?.circularCount || 0) === 0 ? 'No cyclic graph clusters' : 'Tarjan SCC cycles detected'}
          </p>
        </div>
      </div>

      {/* Circular Dependencies Section (if any) */}
      {arch?.cycles && arch.cycles.length > 0 && (
        <div className="bg-amber-950/20 border border-amber-500/30 rounded-xl p-6 space-y-4 font-mono">
          <div className="flex items-center gap-2 text-amber-400">
            <ShieldAlert className="w-5 h-5 shrink-0" />
            <h3 className="text-sm font-semibold">Detected Circular Dependency Clusters</h3>
          </div>
          <div className="space-y-2">
            {arch.cycles.map((cyc, idx) => (
              <div key={idx} className="bg-charcoal-900/80 border border-charcoal-750 p-3 rounded-lg text-xs">
                <div className="text-zinc-400 mb-1">Cycle #{idx + 1} ({cyc.files?.length || 0} files):</div>
                <div className="text-amber-300 break-all font-mono leading-relaxed">{cyc.chain}</div>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Active Rules Section */}
      <div className="bg-charcoal-850 border border-charcoal-750 rounded-xl p-6 space-y-4 font-mono">
        <div className="flex items-center justify-between border-b border-charcoal-750 pb-3">
          <div className="flex items-center gap-2">
            <Sliders className="w-4 h-4 text-dna-cyan" />
            <h3 className="text-sm font-semibold text-white">Active Architecture Boundary Rules</h3>
          </div>
          <span className="text-xs text-zinc-400">{arch?.rules?.length ?? 0} rules active</span>
        </div>

        {arch?.rules && arch.rules.length > 0 ? (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4 pt-2">
            {arch.rules.map((r) => (
              <div
                key={r.id}
                className="bg-charcoal-900 border border-charcoal-750 hover:border-zinc-600 rounded-lg p-4 flex flex-col justify-between transition-colors group"
              >
                <div className="space-y-2">
                  <div className="flex items-center justify-between">
                    <span className="text-xs font-bold text-white truncate pr-2">{r.name}</span>
                    <span
                      className={`text-[10px] px-2 py-0.5 rounded font-mono uppercase font-semibold ${
                        r.severity === 'high'
                          ? 'bg-rose-500/10 text-rose-400 border border-rose-500/30'
                          : r.severity === 'medium'
                          ? 'bg-amber-500/10 text-amber-400 border border-amber-500/30'
                          : 'bg-blue-500/10 text-blue-400 border border-blue-500/30'
                      }`}
                    >
                      {r.severity}
                    </span>
                  </div>

                  {/* Flow Pattern */}
                  <div className="bg-charcoal-950 p-2 rounded border border-charcoal-800 text-[11px] flex items-center justify-between gap-1 overflow-x-auto">
                    <span className="text-dna-blue font-bold truncate max-w-[45%]">{r.fromPattern}</span>
                    <span className="text-rose-400 px-1 font-bold shrink-0">↛</span>
                    <span className="text-dna-rose font-bold truncate max-w-[45%]">{r.toPattern}</span>
                  </div>

                  <p className="text-[11px] text-zinc-400 line-clamp-2 leading-relaxed">
                    {r.description || 'No custom description'}
                  </p>
                </div>

                <div className="mt-4 pt-3 border-t border-charcoal-800 flex items-center justify-between text-[11px]">
                  <span className="text-zinc-500 text-[10px] truncate max-w-[150px]">{r.id}</span>
                  <button
                    onClick={() => handleDeleteRule(r.id)}
                    title="Delete rule"
                    className="text-zinc-500 hover:text-rose-400 p-1 rounded transition-colors"
                  >
                    <Trash2 className="w-3.5 h-3.5" />
                  </button>
                </div>
              </div>
            ))}
          </div>
        ) : (
          <div className="py-6 text-center text-xs text-zinc-500">
            No architecture rules currently configured. Click "Add Custom Rule" to define boundaries.
          </div>
        )}
      </div>

      {/* Detected Violations Table */}
      <div className="bg-charcoal-850 border border-charcoal-750 rounded-xl p-6 space-y-4 font-mono">
        <div className="flex items-center justify-between border-b border-charcoal-750 pb-3">
          <h3 className="text-sm font-semibold text-white flex items-center gap-2">
            <AlertTriangle className="w-4 h-4 text-dna-rose" />
            <span>Architectural Boundary Violations</span>
          </h3>
          <span className="text-xs text-zinc-400">{arch?.violations?.length ?? 0} violations found</span>
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

      {/* Add Custom Rule Modal */}
      {showAddModal && (
        <div className="fixed inset-0 z-50 bg-black/70 backdrop-blur-sm flex items-center justify-center p-4">
          <div className="bg-charcoal-900 border border-charcoal-700 rounded-2xl max-w-xl w-full p-6 shadow-2xl font-mono space-y-6 animate-in fade-in zoom-in-95 duration-150">
            <div className="flex items-center justify-between border-b border-charcoal-750 pb-4">
              <div className="flex items-center gap-2 text-white">
                <Sliders className="w-5 h-5 text-dna-blue" />
                <h3 className="font-bold text-base">New Architecture Boundary Rule</h3>
              </div>
              <button
                onClick={() => setShowAddModal(false)}
                className="text-zinc-400 hover:text-white text-xs px-2 py-1 rounded bg-charcoal-800"
              >
                ✕ Esc
              </button>
            </div>

            {/* Quick Presets */}
            <div className="space-y-2">
              <div className="text-[11px] text-zinc-400 flex items-center gap-1.5 font-sans">
                <Sparkles className="w-3.5 h-3.5 text-dna-cyan" />
                <span>Quick Presets:</span>
              </div>
              <div className="flex flex-wrap gap-2">
                <button
                  type="button"
                  onClick={() =>
                    handleApplyPreset({
                      name: 'Domain Isolation',
                      from: '**/domain/**',
                      to: '**/controllers/**',
                      sev: 'high',
                      desc: 'Domain business logic must never import web presentation controllers',
                    })
                  }
                  className="px-2.5 py-1 text-[11px] bg-charcoal-800 hover:bg-charcoal-750 border border-charcoal-700 rounded text-zinc-300"
                >
                  Domain ↛ Controllers
                </button>
                <button
                  type="button"
                  onClick={() =>
                    handleApplyPreset({
                      name: 'Direct Store Bypass',
                      from: '**/controllers/**',
                      to: '**/store/**',
                      sev: 'medium',
                      desc: 'Controllers must communicate through service layer rather than raw storage',
                    })
                  }
                  className="px-2.5 py-1 text-[11px] bg-charcoal-800 hover:bg-charcoal-750 border border-charcoal-700 rounded text-zinc-300"
                >
                  Controllers ↛ Store
                </button>
                <button
                  type="button"
                  onClick={() =>
                    handleApplyPreset({
                      name: 'UI Purity',
                      from: '**/components/**',
                      to: '**/db/**',
                      sev: 'high',
                      desc: 'Frontend UI components must never access database or persistence layer directly',
                    })
                  }
                  className="px-2.5 py-1 text-[11px] bg-charcoal-800 hover:bg-charcoal-750 border border-charcoal-700 rounded text-zinc-300"
                >
                  Components ↛ DB
                </button>
              </div>
            </div>

            {/* Rule Form */}
            <form onSubmit={handleAddRule} className="space-y-4">
              {formError && (
                <div className="p-3 bg-rose-500/10 border border-rose-500/30 rounded-lg text-rose-400 text-xs">
                  {formError}
                </div>
              )}

              <div>
                <label className="block text-xs text-zinc-300 mb-1">Rule Name</label>
                <input
                  type="text"
                  value={name}
                  onChange={(e) => setName(e.target.value)}
                  placeholder="e.g. Services Cannot Call Views"
                  className="w-full bg-charcoal-950 border border-charcoal-750 rounded-lg px-3 py-2 text-xs text-white focus:outline-none focus:border-dna-blue"
                  required
                />
              </div>

              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="block text-xs text-zinc-300 mb-1">From Pattern (Glob)</label>
                  <input
                    type="text"
                    value={fromPattern}
                    onChange={(e) => setFromPattern(e.target.value)}
                    placeholder="**/services/**"
                    className="w-full bg-charcoal-950 border border-charcoal-750 rounded-lg px-3 py-2 text-xs text-white focus:outline-none focus:border-dna-blue"
                    required
                  />
                </div>

                <div>
                  <label className="block text-xs text-zinc-300 mb-1">To Pattern (Glob)</label>
                  <input
                    type="text"
                    value={toPattern}
                    onChange={(e) => setToPattern(e.target.value)}
                    placeholder="**/views/**"
                    className="w-full bg-charcoal-950 border border-charcoal-750 rounded-lg px-3 py-2 text-xs text-white focus:outline-none focus:border-dna-blue"
                    required
                  />
                </div>
              </div>

              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="block text-xs text-zinc-300 mb-1">Constraint Type</label>
                  <select
                    value={forbidden ? 'forbidden' : 'allowed'}
                    onChange={(e) => setForbidden(e.target.value === 'forbidden')}
                    className="w-full bg-charcoal-950 border border-charcoal-750 rounded-lg px-3 py-2 text-xs text-white focus:outline-none focus:border-dna-blue"
                  >
                    <option value="forbidden">Forbidden (Illegal dependency)</option>
                    <option value="allowed">Allowed Exception</option>
                  </select>
                </div>

                <div>
                  <label className="block text-xs text-zinc-300 mb-1">Severity</label>
                  <select
                    value={severity}
                    onChange={(e) => setSeverity(e.target.value as any)}
                    className="w-full bg-charcoal-950 border border-charcoal-750 rounded-lg px-3 py-2 text-xs text-white focus:outline-none focus:border-dna-blue"
                  >
                    <option value="high">High (15 penalty points)</option>
                    <option value="medium">Medium (8 penalty points)</option>
                    <option value="low">Low (3 penalty points)</option>
                  </select>
                </div>
              </div>

              <div>
                <label className="block text-xs text-zinc-300 mb-1">Description / Decoupling Guideline</label>
                <input
                  type="text"
                  value={description}
                  onChange={(e) => setDescription(e.target.value)}
                  placeholder="Explain why this boundary is enforced and how to decouple"
                  className="w-full bg-charcoal-950 border border-charcoal-750 rounded-lg px-3 py-2 text-xs text-white focus:outline-none focus:border-dna-blue"
                />
              </div>

              <div className="flex items-center justify-end gap-3 pt-3 border-t border-charcoal-750">
                <button
                  type="button"
                  onClick={() => setShowAddModal(false)}
                  className="px-4 py-2 text-xs text-zinc-400 hover:text-white bg-charcoal-800 rounded-lg transition-colors"
                >
                  Cancel
                </button>

                <button
                  type="submit"
                  disabled={actionLoading}
                  className="px-5 py-2 text-xs font-semibold text-white bg-dna-blue hover:bg-dna-blue/90 disabled:opacity-50 rounded-lg transition-all shadow-sm flex items-center gap-2"
                >
                  {actionLoading ? (
                    <>
                      <RefreshCw className="w-3.5 h-3.5 animate-spin" />
                      <span>Evaluating...</span>
                    </>
                  ) : (
                    <span>Save &amp; Enforce Rule</span>
                  )}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}
