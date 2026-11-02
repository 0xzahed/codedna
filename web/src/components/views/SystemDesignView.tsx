'use client';

import React, { useEffect, useState, useMemo } from 'react';
import { useAppStore } from '../../lib/store';
import { 
  fetchSystemDesign, 
  fetchFileContent,
  SystemDesignReport, 
  SystemTier, 
  DataFlowHop, 
  DetectedPattern,
  PatternCategory,
  ArchRecommendation
} from '../../lib/api';
import ArchitectureDiagramGraph from '../ArchitectureDiagramGraph';
import CodeEditorViewer from '../CodeEditorViewer';
import { 
  Workflow, 
  Layers, 
  ArrowRight, 
  ShieldCheck, 
  AlertTriangle, 
  CheckCircle2, 
  Cpu, 
  Database, 
  Globe, 
  Server, 
  Radio, 
  ExternalLink, 
  Sparkles, 
  RotateCw, 
  Code2, 
  Compass, 
  FileCode, 
  Zap, 
  X, 
  Maximize2,
  Lightbulb,
  CheckSquare,
  Square,
  TrendingUp,
  Sliders,
  Filter,
  Check
} from 'lucide-react';

const TIER_ICONS: Record<string, React.ElementType> = {
  client_gateway: Globe,
  api_presentation: Server,
  domain_service: Cpu,
  persistence_storage: Database,
  async_messaging: Radio,
  external_integration: ExternalLink,
};

const TIER_COLORS: Record<string, { bg: string; text: string; border: string; glow: string }> = {
  client_gateway: {
    bg: 'bg-sky-500/10',
    text: 'text-sky-400',
    border: 'border-sky-500/30',
    glow: 'hover:border-sky-400/50',
  },
  api_presentation: {
    bg: 'bg-indigo-500/10',
    text: 'text-indigo-400',
    border: 'border-indigo-500/30',
    glow: 'hover:border-indigo-400/50',
  },
  domain_service: {
    bg: 'bg-emerald-500/10',
    text: 'text-emerald-400',
    border: 'border-emerald-500/30',
    glow: 'hover:border-emerald-400/50',
  },
  persistence_storage: {
    bg: 'bg-amber-500/10',
    text: 'text-amber-400',
    border: 'border-amber-500/30',
    glow: 'hover:border-amber-400/50',
  },
  async_messaging: {
    bg: 'bg-purple-500/10',
    text: 'text-purple-400',
    border: 'border-purple-500/30',
    glow: 'hover:border-purple-400/50',
  },
  external_integration: {
    bg: 'bg-rose-500/10',
    text: 'text-rose-400',
    border: 'border-rose-500/30',
    glow: 'hover:border-rose-400/50',
  },
};

export default function SystemDesignView() {
  const { scanId } = useAppStore();
  const [report, setReport] = useState<SystemDesignReport | null>(null);
  const [loading, setLoading] = useState(true);
  const [selectedTier, setSelectedTier] = useState<SystemTier | null>(null);
  const [patternFilter, setPatternFilter] = useState<'All' | PatternCategory>('All');
  const [expandedHop, setExpandedHop] = useState<number | null>(null);
  const [viewMode, setViewMode] = useState<'diagram' | 'breakdown' | 'recommendations'>('diagram');

  // Optimization recommendations filter & action state
  const [recSeverityFilter, setRecSeverityFilter] = useState<'all' | 'high' | 'medium' | 'low'>('all');
  const [recCategoryFilter, setRecCategoryFilter] = useState<string>('all');
  const [checkedActions, setCheckedActions] = useState<Record<string, boolean>>({});

  // Code inspection in editor
  const [inspectedFile, setInspectedFile] = useState<{ path: string; content: string } | null>(null);
  const [loadingFile, setLoadingFile] = useState(false);

  const loadData = () => {
    setLoading(true);
    fetchSystemDesign(scanId)
      .then((data) => {
        setReport(data);
        if (data.tiers && data.tiers.length > 0) {
          const firstActive = data.tiers.find(t => t.fileCount > 0) || data.tiers[0];
          setSelectedTier(firstActive);
        }
      })
      .catch((err) => console.error('Failed to load system design report:', err))
      .finally(() => setLoading(false));
  };

  useEffect(() => {
    loadData();
  }, [scanId]);

  const handleOpenFile = async (filePath: string) => {
    setLoadingFile(true);
    try {
      const data = await fetchFileContent(scanId, filePath);
      setInspectedFile(data);
    } catch (err) {
      console.error('Failed to load file content:', err);
    } finally {
      setLoadingFile(false);
    }
  };

  const toggleAction = (recId: string, idx: number) => {
    const key = `${recId}-${idx}`;
    setCheckedActions(prev => ({
      ...prev,
      [key]: !prev[key]
    }));
  };

  const activeTiers = report?.tiers.filter((t) => t.fileCount > 0) || [];

  const filteredPatterns = (report?.detectedPatterns || []).filter((p) => {
    if (patternFilter === 'All') return true;
    return p.category === patternFilter;
  });

  const recommendations = report?.recommendations || [];

  const filteredRecs = useMemo(() => {
    return recommendations.filter((r) => {
      if (recSeverityFilter !== 'all' && r.severity !== recSeverityFilter) return false;
      if (recCategoryFilter !== 'all' && r.category !== recCategoryFilter) return false;
      return true;
    });
  }, [recommendations, recSeverityFilter, recCategoryFilter]);

  const highSeverityRecs = recommendations.filter(r => r.severity === 'high');

  return (
    <div className="p-6 md:p-8 space-y-8 w-full max-w-[1920px] mx-auto">
      {/* Header */}
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 border-b border-charcoal-750 pb-6">
        <div>
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 rounded-xl bg-dna-blue/10 border border-dna-blue/30 flex items-center justify-center text-dna-blue">
              <Workflow className="w-5 h-5" />
            </div>
            <div>
              <h1 className="text-2xl font-bold tracking-tight text-white flex items-center gap-3">
                System Design &amp; Architecture Flow
              </h1>
              <p className="text-sm text-zinc-400 mt-0.5">
                Multi-tier structural breakdown, protocol data flow hops, pattern recognition, and architectural optimization engine.
              </p>
            </div>
          </div>
        </div>

        <div className="flex items-center gap-3">
          {/* View Mode Toggle */}
          <div className="flex items-center gap-1 bg-charcoal-900 p-1 rounded-xl border border-charcoal-750 font-mono text-xs">
            <button
              onClick={() => setViewMode('diagram')}
              className={`flex items-center gap-2 px-3.5 py-1.5 rounded-lg transition-all ${
                viewMode === 'diagram' ? 'bg-dna-blue text-white shadow-sm' : 'text-zinc-400 hover:text-white'
              }`}
            >
              <Compass className="w-4 h-4" />
              <span>Diagram Graph</span>
            </button>
            <button
              onClick={() => setViewMode('breakdown')}
              className={`flex items-center gap-2 px-3.5 py-1.5 rounded-lg transition-all ${
                viewMode === 'breakdown' ? 'bg-dna-blue text-white shadow-sm' : 'text-zinc-400 hover:text-white'
              }`}
            >
              <Layers className="w-4 h-4" />
              <span>Topology &amp; Patterns</span>
            </button>
            <button
              onClick={() => setViewMode('recommendations')}
              className={`flex items-center gap-2 px-3.5 py-1.5 rounded-lg transition-all relative ${
                viewMode === 'recommendations' ? 'bg-dna-blue text-white shadow-sm' : 'text-zinc-400 hover:text-white'
              }`}
            >
              <Lightbulb className="w-4 h-4 text-amber-400" />
              <span>Optimizations</span>
              {recommendations.length > 0 && (
                <span className={`text-[10px] font-mono px-1.5 py-0.2 rounded-full font-bold ${
                  highSeverityRecs.length > 0 ? 'bg-rose-500 text-white' : 'bg-charcoal-700 text-zinc-300'
                }`}>
                  {recommendations.length}
                </span>
              )}
            </button>
          </div>

          <button
            onClick={loadData}
            disabled={loading}
            className="flex items-center gap-2 px-3.5 py-2 rounded-lg bg-charcoal-800 hover:bg-charcoal-750 border border-charcoal-700 text-xs font-medium text-zinc-300 hover:text-white transition-all disabled:opacity-50"
          >
            <RotateCw className={`w-3.5 h-3.5 ${loading ? 'animate-spin' : ''}`} />
            <span>Re-analyze</span>
          </button>
        </div>
      </div>

      {/* Metrics Banner */}
      {report && (
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
          <div className="bg-charcoal-900 border border-charcoal-750 rounded-xl p-5">
            <div className="text-xs font-medium text-zinc-400 uppercase tracking-wider font-mono">Architecture Style</div>
            <div className="text-sm font-semibold text-white mt-2 line-clamp-2">{report.architectureStyle}</div>
            <div className="text-xs text-zinc-500 mt-1">{activeTiers.length} of {report.tiers.length} Tiers Active</div>
          </div>

          <div className="bg-charcoal-900 border border-charcoal-750 rounded-xl p-5">
            <div className="text-xs font-medium text-zinc-400 uppercase tracking-wider font-mono">Reliability Score</div>
            <div className="flex items-baseline gap-2 mt-2">
              <span className={`text-2xl font-bold font-mono ${
                report.reliabilityScore >= 80 ? 'text-emerald-400' :
                report.reliabilityScore >= 60 ? 'text-amber-400' : 'text-rose-400'
              }`}>
                {report.reliabilityScore.toFixed(1)}
              </span>
              <span className="text-xs text-zinc-500">/ 100</span>
            </div>
            <div className="text-xs text-zinc-500 mt-1">Based on boundary cohesion &amp; isolation</div>
          </div>

          <div className="bg-charcoal-900 border border-charcoal-750 rounded-xl p-5">
            <div className="text-xs font-medium text-zinc-400 uppercase tracking-wider font-mono">Recognized Patterns</div>
            <div className="text-2xl font-bold font-mono text-dna-blue mt-2">{report.detectedPatterns.length}</div>
            <div className="text-xs text-zinc-500 mt-1">Architectural, structural &amp; creational</div>
          </div>

          <div 
            onClick={() => setViewMode('recommendations')}
            className="bg-charcoal-900 border border-charcoal-750 hover:border-amber-500/50 rounded-xl p-5 cursor-pointer transition-all group"
          >
            <div className="text-xs font-medium text-zinc-400 uppercase tracking-wider font-mono flex items-center justify-between">
              <span>Optimization Actions</span>
              <ArrowRight className="w-3.5 h-3.5 text-zinc-500 group-hover:text-amber-400 transition-colors" />
            </div>
            <div className="flex items-baseline gap-2 mt-2">
              <span className="text-2xl font-bold font-mono text-amber-400">
                {recommendations.length}
              </span>
              <span className="text-xs text-zinc-400">improvements</span>
            </div>
            <div className="text-xs text-zinc-500 mt-1">
              {highSeverityRecs.length > 0 ? (
                <span className="text-rose-400 font-medium">{highSeverityRecs.length} high priority risk</span>
              ) : (
                <span>Boundary integrity validated</span>
              )}
            </div>
          </div>
        </div>
      )}

      {/* VIEW MODE 1: Interactive Architecture Diagram Graph */}
      {viewMode === 'diagram' && report && (
        <div className="space-y-6 animate-in fade-in duration-200">
          <div className="flex items-center justify-between">
            <div>
              <h2 className="text-lg font-semibold text-white flex items-center gap-2">
                <Compass className="w-5 h-5 text-dna-blue" />
                <span>Architecture Diagram Flow Canvas</span>
              </h2>
              <p className="text-xs text-zinc-400 mt-0.5">
                Drag tiers, search components, filter protocol pipes, and click any component to open it in the code editor.
              </p>
            </div>
            <span className="text-xs font-mono text-zinc-400">Animated Protocol Pipes • Interactive Nodes</span>
          </div>

          {/* Interactive React Flow Architecture Graph Canvas */}
          <ArchitectureDiagramGraph 
            report={report} 
            onSelectTier={setSelectedTier} 
            onOpenFile={handleOpenFile} 
          />

          {/* Selected Tier Inspection Drawer directly below diagram */}
          {selectedTier && (
            <div className="bg-charcoal-900 border border-charcoal-750 rounded-xl p-6 space-y-5 shadow-lg">
              <div className="flex items-center justify-between border-b border-charcoal-800 pb-4">
                <div className="flex items-center gap-3">
                  <span className="text-xs font-mono px-2 py-0.5 rounded bg-charcoal-800 text-dna-blue border border-dna-blue/30">
                    Tier {selectedTier.order} Inspection
                  </span>
                  <h4 className="text-base font-semibold text-white">{selectedTier.name}</h4>
                </div>
                <div className="text-xs font-mono text-zinc-400">
                  {selectedTier.files?.length || 0} Files • {selectedTier.symbolCount} Symbols
                </div>
              </div>

              <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                {/* Files in Tier */}
                <div>
                  <h5 className="text-xs font-medium text-zinc-300 uppercase tracking-wider font-mono flex items-center gap-2 mb-3">
                    <FileCode className="w-4 h-4 text-dna-blue" />
                    <span>Component Files (Click to inspect in Editor)</span>
                  </h5>
                  <div className="bg-charcoal-950 border border-charcoal-800 rounded-lg p-2 max-h-60 overflow-y-auto space-y-1">
                    {selectedTier.files && selectedTier.files.length > 0 ? (
                      selectedTier.files.map((file) => (
                        <div 
                          key={file} 
                          onClick={() => handleOpenFile(file)}
                          className="text-xs font-mono text-zinc-300 hover:text-white flex items-center justify-between py-1.5 px-2 rounded hover:bg-charcoal-850 cursor-pointer border-b border-charcoal-850/60 last:border-0 group transition-colors"
                        >
                          <div className="flex items-center gap-2 truncate">
                            <span className="text-zinc-600">/</span>
                            <span className="truncate group-hover:text-dna-blue">{file}</span>
                          </div>
                          <span className="text-[10px] text-dna-blue opacity-0 group-hover:opacity-100 flex items-center gap-1 font-sans">
                            <span>Open in Editor</span>
                            <ArrowRight className="w-3 h-3" />
                          </span>
                        </div>
                      ))
                    ) : (
                      <div className="text-xs text-zinc-500 py-3 text-center">No files classified in this tier.</div>
                    )}
                  </div>
                </div>

                {/* Key Symbols in Tier */}
                <div>
                  <h5 className="text-xs font-medium text-zinc-300 uppercase tracking-wider font-mono flex items-center gap-2 mb-3">
                    <Code2 className="w-4 h-4 text-emerald-400" />
                    <span>Representative Symbols ({selectedTier.keySymbols?.length || 0})</span>
                  </h5>
                  <div className="bg-charcoal-950 border border-charcoal-800 rounded-lg p-3 max-h-60 overflow-y-auto space-y-1.5">
                    {selectedTier.keySymbols && selectedTier.keySymbols.length > 0 ? (
                      selectedTier.keySymbols.map((sym) => (
                        <div key={sym} className="text-xs font-mono text-emerald-300/90 py-1 border-b border-charcoal-850/60 last:border-0 flex items-center gap-2">
                          <span className="text-emerald-500/50">ƒ</span>
                          <span className="truncate">{sym}</span>
                        </div>
                      ))
                    ) : (
                      <div className="text-xs text-zinc-500 py-3 text-center">No symbols registered in this tier.</div>
                    )}
                  </div>
                </div>
              </div>
            </div>
          )}

          {/* Quick recommendations callout card */}
          {recommendations.length > 0 && (
            <div className="bg-charcoal-900 border border-charcoal-750 hover:border-charcoal-600 rounded-xl p-5 flex flex-col sm:flex-row sm:items-center justify-between gap-4 transition-all">
              <div className="flex items-center gap-3.5">
                <div className="w-10 h-10 rounded-xl bg-amber-500/10 border border-amber-500/30 flex items-center justify-center text-amber-400 shrink-0">
                  <Lightbulb className="w-5 h-5" />
                </div>
                <div>
                  <h4 className="text-sm font-semibold text-white flex items-center gap-2">
                    <span>{recommendations.length} Architectural Optimization Recommendations Available</span>
                    {highSeverityRecs.length > 0 && (
                      <span className="text-[10px] font-mono px-2 py-0.5 rounded bg-rose-500/20 text-rose-300 border border-rose-500/40">
                        {highSeverityRecs.length} High Priority
                      </span>
                    )}
                  </h4>
                  <p className="text-xs text-zinc-400 mt-0.5">
                    Explore concrete implementation steps for boundary decoupling, storage caching, and asynchronous task offloading.
                  </p>
                </div>
              </div>

              <button
                onClick={() => setViewMode('recommendations')}
                className="flex items-center gap-2 px-4 py-2 rounded-lg bg-dna-blue hover:bg-blue-600 text-xs font-semibold text-white transition-all shadow-md shrink-0"
              >
                <span>View Recommendations</span>
                <ArrowRight className="w-3.5 h-3.5" />
              </button>
            </div>
          )}
        </div>
      )}

      {/* VIEW MODE 2: Topology Cards & Pattern Recognition Gallery */}
      {viewMode === 'breakdown' && (
        <div className="space-y-8 animate-in fade-in duration-200">
          {/* Tiers Visual Architecture Grid */}
          <div className="space-y-4">
            <div className="flex items-center justify-between">
              <h2 className="text-lg font-semibold text-white flex items-center gap-2">
                <Layers className="w-5 h-5 text-dna-blue" />
                <span>Multi-Tier Architecture Breakdown</span>
              </h2>
              <span className="text-xs text-zinc-400">Click a tier to inspect files &amp; key symbols</span>
            </div>

            {report && (
              <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                {report.tiers.map((tier) => {
                  const Icon = TIER_ICONS[tier.id] || Layers;
                  const style = TIER_COLORS[tier.id] || {
                    bg: 'bg-zinc-800',
                    text: 'text-zinc-300',
                    border: 'border-zinc-700',
                    glow: 'hover:border-zinc-500',
                  };
                  const isSelected = selectedTier?.id === tier.id;

                  return (
                    <div
                      key={tier.id}
                      onClick={() => setSelectedTier(tier)}
                      className={`cursor-pointer rounded-xl p-5 border transition-all duration-200 ${
                        isSelected
                          ? `bg-charcoal-850 ${style.border} ring-1 ring-dna-blue/40 shadow-lg shadow-dna-blue/5`
                          : `bg-charcoal-900 border-charcoal-750 ${style.glow} hover:bg-charcoal-850/60`
                      }`}
                    >
                      <div className="flex items-start justify-between">
                        <div className="flex items-center gap-3">
                          <div className={`w-10 h-10 rounded-lg ${style.bg} border ${style.border} flex items-center justify-center ${style.text}`}>
                            <Icon className="w-5 h-5" />
                          </div>
                          <div>
                            <div className="flex items-center gap-2">
                              <span className="text-[10px] font-mono font-bold px-1.5 py-0.5 rounded bg-charcoal-800 border border-charcoal-700 text-zinc-400">
                                Tier {tier.order}
                              </span>
                              <h3 className="text-sm font-semibold text-white">{tier.name}</h3>
                            </div>
                            <div className="text-[11px] text-zinc-400 mt-1 flex items-center gap-3 font-mono">
                              <span>{tier.fileCount} files</span>
                              <span>•</span>
                              <span>{tier.symbolCount} symbols</span>
                            </div>
                          </div>
                        </div>
                      </div>

                      <p className="text-xs text-zinc-400 mt-3 line-clamp-2 leading-relaxed">
                        {tier.description}
                      </p>
                    </div>
                  );
                })}
              </div>
            )}

            {/* Selected Tier Drill-down Drawer */}
            {selectedTier && (
              <div className="bg-charcoal-900 border border-charcoal-750 rounded-xl p-6 mt-4 space-y-5">
                <div className="flex items-center justify-between border-b border-charcoal-800 pb-4">
                  <div className="flex items-center gap-3">
                    <span className="text-xs font-mono px-2 py-0.5 rounded bg-charcoal-800 text-dna-blue border border-dna-blue/30">
                      Tier {selectedTier.order} Inspection
                    </span>
                    <h4 className="text-base font-semibold text-white">{selectedTier.name}</h4>
                  </div>
                  <div className="text-xs font-mono text-zinc-400">
                    {selectedTier.files?.length || 0} Files • {selectedTier.symbolCount} Symbols
                  </div>
                </div>

                <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                  {/* Files in Tier */}
                  <div>
                    <h5 className="text-xs font-medium text-zinc-300 uppercase tracking-wider font-mono flex items-center gap-2 mb-3">
                      <FileCode className="w-4 h-4 text-dna-blue" />
                      <span>Component Files (Click to inspect in Editor)</span>
                    </h5>
                    <div className="bg-charcoal-950 border border-charcoal-800 rounded-lg p-2 max-h-56 overflow-y-auto space-y-1">
                      {selectedTier.files && selectedTier.files.length > 0 ? (
                        selectedTier.files.map((file) => (
                          <div 
                            key={file} 
                            onClick={() => handleOpenFile(file)}
                            className="text-xs font-mono text-zinc-300 hover:text-white flex items-center justify-between py-1.5 px-2 rounded hover:bg-charcoal-850 cursor-pointer border-b border-charcoal-850/60 last:border-0 group"
                          >
                            <div className="flex items-center gap-2 truncate">
                              <span className="text-zinc-600">/</span>
                              <span className="truncate group-hover:text-dna-blue">{file}</span>
                            </div>
                            <span className="text-[10px] text-dna-blue opacity-0 group-hover:opacity-100 flex items-center gap-1 font-sans">
                              <span>Open in Editor</span>
                              <ArrowRight className="w-3 h-3" />
                            </span>
                          </div>
                        ))
                      ) : (
                        <div className="text-xs text-zinc-500 py-3 text-center">No files classified in this tier.</div>
                      )}
                    </div>
                  </div>

                  {/* Key Symbols in Tier */}
                  <div>
                    <h5 className="text-xs font-medium text-zinc-300 uppercase tracking-wider font-mono flex items-center gap-2 mb-3">
                      <Code2 className="w-4 h-4 text-emerald-400" />
                      <span>Representative Symbols ({selectedTier.keySymbols?.length || 0})</span>
                    </h5>
                    <div className="bg-charcoal-950 border border-charcoal-800 rounded-lg p-3 max-h-56 overflow-y-auto space-y-1.5">
                      {selectedTier.keySymbols && selectedTier.keySymbols.length > 0 ? (
                        selectedTier.keySymbols.map((sym) => (
                          <div key={sym} className="text-xs font-mono text-emerald-300/90 py-1 border-b border-charcoal-850/60 last:border-0 flex items-center gap-2">
                            <span className="text-emerald-500/50">ƒ</span>
                            <span className="truncate">{sym}</span>
                          </div>
                        ))
                      ) : (
                        <div className="text-xs text-zinc-500 py-3 text-center">No symbols registered in this tier.</div>
                      )}
                    </div>
                  </div>
                </div>
              </div>
            )}
          </div>

          {/* End-to-End Data Flow Pipeline */}
          <div className="space-y-4">
            <div className="flex items-center justify-between">
              <h2 className="text-lg font-semibold text-white flex items-center gap-2">
                <Compass className="w-5 h-5 text-indigo-400" />
                <span>End-to-End Data Flow &amp; Inter-Tier Protocols</span>
              </h2>
              <span className="text-xs text-zinc-400">Reconstructed from AST call graph edges</span>
            </div>

            {report && report.dataFlow.length > 0 ? (
              <div className="bg-charcoal-900 border border-charcoal-750 rounded-xl divide-y divide-charcoal-800 overflow-hidden">
                {report.dataFlow.map((hop, idx) => {
                  const isExpanded = expandedHop === idx;
                  return (
                    <div key={idx} className="p-4 hover:bg-charcoal-850/40 transition-colors">
                      <div 
                        onClick={() => setExpandedHop(isExpanded ? null : idx)}
                        className="flex flex-col sm:flex-row sm:items-center justify-between gap-3 cursor-pointer"
                      >
                        <div className="flex items-center gap-3">
                          <span className="text-xs font-mono px-2 py-0.5 rounded bg-charcoal-800 text-zinc-300 border border-charcoal-700 capitalize">
                            {hop.fromTier.replace('_', ' ')}
                          </span>
                          <ArrowRight className="w-4 h-4 text-dna-blue shrink-0" />
                          <span className="text-xs font-mono px-2 py-0.5 rounded bg-charcoal-800 text-zinc-300 border border-charcoal-700 capitalize">
                            {hop.toTier.replace('_', ' ')}
                          </span>
                        </div>

                        <div className="flex items-center gap-3">
                          <span className="text-xs font-mono px-2.5 py-0.5 rounded bg-indigo-500/10 text-indigo-400 border border-indigo-500/30">
                            {hop.protocol}
                          </span>
                          <span className="text-xs font-mono font-bold text-white bg-charcoal-800 px-2 py-0.5 rounded border border-charcoal-700">
                            {hop.callCount} calls
                          </span>
                        </div>
                      </div>

                      {/* Expandable Sample Calls */}
                      {isExpanded && hop.sampleCalls && hop.sampleCalls.length > 0 && (
                        <div className="mt-3 pt-3 border-t border-charcoal-800/80 pl-2">
                          <div className="text-[11px] font-mono text-zinc-400 uppercase tracking-wider mb-2">Sample Cross-Boundary Invocations:</div>
                          <div className="space-y-1">
                            {hop.sampleCalls.map((call, cIdx) => (
                              <div key={cIdx} className="text-xs font-mono text-zinc-300 bg-charcoal-950 px-3 py-1.5 rounded border border-charcoal-800">
                                {call}
                              </div>
                            ))}
                          </div>
                        </div>
                      )}
                    </div>
                  );
                })}
              </div>
            ) : (
              <div className="bg-charcoal-900 border border-charcoal-750 rounded-xl p-8 text-center text-zinc-500 text-sm">
                No cross-tier data flows detected in the current codebase graph.
              </div>
            )}
          </div>

          {/* Software Design Patterns Gallery */}
          <div className="space-y-4">
            <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
              <h2 className="text-lg font-semibold text-white flex items-center gap-2">
                <Sparkles className="w-5 h-5 text-amber-400" />
                <span>Recognized Software Design Patterns</span>
              </h2>

              {/* Filter Pills */}
              <div className="flex items-center gap-1.5 bg-charcoal-900 p-1 rounded-lg border border-charcoal-750 text-xs font-medium">
                {(['All', 'Architectural', 'Structural', 'Creational', 'Behavioral'] as const).map((cat) => (
                  <button
                    key={cat}
                    onClick={() => setPatternFilter(cat)}
                    className={`px-3 py-1 rounded-md transition-all ${
                      patternFilter === cat
                        ? 'bg-dna-blue text-white shadow-sm'
                        : 'text-zinc-400 hover:text-zinc-200'
                    }`}
                  >
                    {cat}
                  </button>
                ))}
              </div>
            </div>

            {filteredPatterns.length > 0 ? (
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                {filteredPatterns.map((pat) => (
                  <div
                    key={pat.id}
                    className="bg-charcoal-900 border border-charcoal-750 rounded-xl p-5 space-y-3 hover:border-charcoal-600 transition-colors"
                  >
                    <div className="flex items-start justify-between gap-3">
                      <div>
                        <div className="flex items-center gap-2">
                          <span className={`text-[10px] font-mono uppercase px-2 py-0.5 rounded font-bold border ${
                            pat.category === 'Architectural' ? 'bg-purple-500/10 text-purple-400 border-purple-500/30' :
                            pat.category === 'Structural' ? 'bg-sky-500/10 text-sky-400 border-sky-500/30' :
                            pat.category === 'Creational' ? 'bg-emerald-500/10 text-emerald-400 border-emerald-500/30' :
                            'bg-amber-500/10 text-amber-400 border-amber-500/30'
                          }`}>
                            {pat.category}
                          </span>
                          <h3 className="text-sm font-semibold text-white">{pat.name}</h3>
                        </div>
                      </div>

                      <div className="text-right shrink-0">
                        <span className="text-xs font-mono font-bold text-dna-blue">
                          {(pat.confidence * 100).toFixed(0)}%
                        </span>
                        <span className="text-[10px] text-zinc-500 block">confidence</span>
                      </div>
                    </div>

                    <p className="text-xs text-zinc-300 leading-relaxed">
                      {pat.description}
                    </p>

                    {/* Evidence List */}
                    <div className="bg-charcoal-950 border border-charcoal-800 rounded-lg p-3 space-y-1.5">
                      <div className="text-[10px] font-mono text-zinc-400 uppercase tracking-wider">Identified Evidence:</div>
                      {pat.evidence.map((ev, eIdx) => (
                        <div key={eIdx} className="text-xs text-zinc-300 flex items-start gap-2">
                          <CheckCircle2 className="w-3.5 h-3.5 text-emerald-400 shrink-0 mt-0.5" />
                          <span>{ev}</span>
                        </div>
                      ))}
                    </div>

                    {/* Trade-offs */}
                    {pat.tradeoffs && (
                      <div className="text-xs text-zinc-400 border-t border-charcoal-800/80 pt-2.5">
                        <span className="font-semibold text-zinc-300">Engineering Trade-offs: </span>
                        <span>{pat.tradeoffs}</span>
                      </div>
                    )}
                  </div>
                ))}
              </div>
            ) : (
              <div className="bg-charcoal-900 border border-charcoal-750 rounded-xl p-8 text-center text-zinc-500 text-sm">
                No design patterns match the selected filter.
              </div>
            )}
          </div>

          {/* Architectural Bottleneck Hotspots & Key Entrypoints */}
          {report && (
            <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
              {/* Bottlenecks */}
              <div className="bg-charcoal-900 border border-charcoal-750 rounded-xl p-5 space-y-3">
                <h3 className="text-sm font-semibold text-white flex items-center gap-2">
                  <AlertTriangle className="w-4 h-4 text-amber-400" />
                  <span>Architectural Bottleneck &amp; Hotspot Diagnostics</span>
                </h3>
                <div className="space-y-2">
                  {report.bottleneckHotspots.map((b, idx) => (
                    <div
                      key={idx}
                      className={`text-xs p-3 rounded-lg border font-mono ${
                        b.includes('Risk')
                          ? 'bg-rose-500/10 text-rose-300 border-rose-500/30'
                          : b.includes('Warning') || b.includes('Traffic')
                          ? 'bg-amber-500/10 text-amber-300 border-amber-500/30'
                          : 'bg-emerald-500/10 text-emerald-300 border-emerald-500/30'
                      }`}
                    >
                      {b}
                    </div>
                  ))}
                </div>
              </div>

              {/* Key Entrypoints */}
              <div className="bg-charcoal-900 border border-charcoal-750 rounded-xl p-5 space-y-3">
                <h3 className="text-sm font-semibold text-white flex items-center gap-2">
                  <Zap className="w-4 h-4 text-dna-blue" />
                  <span>Primary Application Entrypoints</span>
                </h3>
                <div className="space-y-2">
                  {report.keyEntrypoints.length > 0 ? (
                    report.keyEntrypoints.map((ep, idx) => (
                      <div key={idx} className="text-xs font-mono text-zinc-300 bg-charcoal-950 p-2.5 rounded border border-charcoal-800 flex items-center gap-2">
                        <span className="text-dna-blue">➔</span>
                        <span className="truncate">{ep}</span>
                      </div>
                    ))
                  ) : (
                    <div className="text-xs text-zinc-500 py-3">No entrypoints automatically detected.</div>
                  )}
                </div>
              </div>
            </div>
          )}
        </div>
      )}

      {/* VIEW MODE 3: Architectural Optimization Recommendations Engine */}
      {viewMode === 'recommendations' && (
        <div className="space-y-6 animate-in fade-in duration-200">
          <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 border-b border-charcoal-750 pb-4">
            <div>
              <h2 className="text-lg font-semibold text-white flex items-center gap-2.5">
                <Lightbulb className="w-5 h-5 text-amber-400" />
                <span>Architectural Optimization &amp; Modernization Engine</span>
              </h2>
              <p className="text-xs text-zinc-400 mt-1">
                Concrete, high-impact recommendations to eliminate bottlenecks, decouple layer boundaries, and optimize scalability.
              </p>
            </div>

            {/* Filter Toolbar */}
            <div className="flex items-center gap-3 font-mono text-xs flex-wrap">
              {/* Category Filter */}
              <div className="flex items-center gap-1 bg-charcoal-900 p-1 rounded-lg border border-charcoal-750">
                {(['all', 'Decoupling', 'Caching', 'Scalability', 'Modularity'] as const).map((cat) => (
                  <button
                    key={cat}
                    onClick={() => setRecCategoryFilter(cat)}
                    className={`px-2.5 py-1 rounded capitalize transition-all ${
                      recCategoryFilter === cat ? 'bg-dna-blue text-white shadow-sm font-semibold' : 'text-zinc-400 hover:text-white'
                    }`}
                  >
                    {cat}
                  </button>
                ))}
              </div>

              {/* Severity Filter */}
              <div className="flex items-center gap-1 bg-charcoal-900 p-1 rounded-lg border border-charcoal-750">
                {(['all', 'high', 'medium', 'low'] as const).map((sev) => (
                  <button
                    key={sev}
                    onClick={() => setRecSeverityFilter(sev)}
                    className={`px-2 py-1 rounded capitalize transition-all ${
                      recSeverityFilter === sev
                        ? sev === 'high' ? 'bg-rose-500 text-white font-bold'
                        : sev === 'medium' ? 'bg-amber-500 text-black font-bold'
                        : sev === 'low' ? 'bg-blue-500 text-white font-bold'
                        : 'bg-dna-blue text-white font-bold'
                        : 'text-zinc-400 hover:text-white'
                    }`}
                  >
                    {sev}
                  </button>
                ))}
              </div>
            </div>
          </div>

          {/* Recommendations List */}
          {filteredRecs.length > 0 ? (
            <div className="grid grid-cols-1 gap-5">
              {filteredRecs.map((rec) => {
                const completedCount = rec.actionSteps.filter((_, idx) => checkedActions[`${rec.id}-${idx}`]).length;
                const isAllCompleted = completedCount === rec.actionSteps.length && rec.actionSteps.length > 0;

                return (
                  <div
                    key={rec.id}
                    className={`bg-charcoal-900 border rounded-xl p-6 space-y-4 transition-all shadow-lg ${
                      rec.severity === 'high' ? 'border-rose-500/40 hover:border-rose-400/60' :
                      rec.severity === 'medium' ? 'border-amber-500/40 hover:border-amber-400/60' :
                      'border-charcoal-750 hover:border-charcoal-600'
                    }`}
                  >
                    {/* Header */}
                    <div className="flex flex-col sm:flex-row sm:items-start justify-between gap-3">
                      <div>
                        <div className="flex items-center gap-2.5 flex-wrap">
                          <span className={`text-[10px] font-mono uppercase px-2.5 py-0.5 rounded-full font-bold border ${
                            rec.severity === 'high' ? 'bg-rose-500/10 text-rose-400 border-rose-500/30' :
                            rec.severity === 'medium' ? 'bg-amber-500/10 text-amber-400 border-amber-500/30' :
                            'bg-blue-500/10 text-blue-400 border-blue-500/30'
                          }`}>
                            {rec.severity} priority
                          </span>

                          <span className="text-[10px] font-mono px-2 py-0.5 rounded bg-charcoal-800 text-zinc-300 border border-charcoal-700">
                            {rec.category}
                          </span>

                          {isAllCompleted && (
                            <span className="text-[10px] font-mono px-2 py-0.5 rounded bg-emerald-500/20 text-emerald-300 border border-emerald-500/40 flex items-center gap-1 font-semibold">
                              <Check className="w-3 h-3" />
                              <span>Plan Ready</span>
                            </span>
                          )}
                        </div>

                        <h3 className="text-base font-bold text-white mt-2">
                          {rec.title}
                        </h3>
                      </div>

                      <div className="text-right shrink-0">
                        <span className="text-xs font-mono text-zinc-400">
                          {completedCount} of {rec.actionSteps.length} Actions Checked
                        </span>
                      </div>
                    </div>

                    {/* Issue Description */}
                    <p className="text-xs text-zinc-300 leading-relaxed font-sans bg-charcoal-950 p-3.5 rounded-lg border border-charcoal-800">
                      {rec.description}
                    </p>

                    {/* Actionable Implementation Protocol */}
                    <div className="space-y-2 pt-1">
                      <div className="text-[11px] font-mono text-zinc-400 uppercase tracking-wider flex items-center gap-2">
                        <Sliders className="w-3.5 h-3.5 text-dna-blue" />
                        <span>Actionable Optimization Protocol (Click to track progress):</span>
                      </div>

                      <div className="space-y-1.5">
                        {rec.actionSteps.map((step, idx) => {
                          const isChecked = !!checkedActions[`${rec.id}-${idx}`];
                          return (
                            <div
                              key={idx}
                              onClick={() => toggleAction(rec.id, idx)}
                              className={`flex items-start gap-3 p-2.5 rounded-lg border text-xs cursor-pointer transition-all select-none ${
                                isChecked
                                  ? 'bg-emerald-500/10 border-emerald-500/30 text-emerald-300'
                                  : 'bg-charcoal-950 border-charcoal-800 text-zinc-200 hover:border-charcoal-700 hover:bg-charcoal-850'
                              }`}
                            >
                              <div className="mt-0.5 shrink-0">
                                {isChecked ? (
                                  <CheckSquare className="w-4 h-4 text-emerald-400" />
                                ) : (
                                  <Square className="w-4 h-4 text-zinc-500" />
                                )}
                              </div>
                              <span className={`leading-relaxed ${isChecked ? 'line-through text-zinc-400' : ''}`}>
                                {step}
                              </span>
                            </div>
                          );
                        })}
                      </div>
                    </div>

                    {/* Architectural Impact */}
                    <div className="bg-emerald-950/20 border border-emerald-500/30 rounded-lg p-3.5 flex items-start gap-3 text-xs text-emerald-300">
                      <TrendingUp className="w-4 h-4 text-emerald-400 shrink-0 mt-0.5" />
                      <div>
                        <span className="font-bold uppercase tracking-wider font-mono mr-1.5 text-emerald-400">Architectural Payoff:</span>
                        <span>{rec.impact}</span>
                      </div>
                    </div>
                  </div>
                );
              })}
            </div>
          ) : (
            <div className="bg-charcoal-900 border border-charcoal-750 rounded-xl p-12 text-center text-zinc-400 space-y-2">
              <CheckCircle2 className="w-8 h-8 text-emerald-400 mx-auto" />
              <h4 className="text-sm font-semibold text-white">No optimization warnings in this category!</h4>
              <p className="text-xs text-zinc-500">All boundary rules and performance benchmarks in this filter passed cleanly.</p>
            </div>
          )}
        </div>
      )}

      {/* Full-Featured Embedded Code Editor Modal */}
      {inspectedFile && (
        <div className="fixed inset-0 z-50 bg-black/80 backdrop-blur-sm flex items-center justify-center p-4">
          <div className="bg-[#0f1217] border border-charcoal-700 rounded-2xl w-full max-w-5xl h-[80vh] flex flex-col overflow-hidden shadow-2xl animate-in zoom-in-95 duration-150">
            {/* Modal Header */}
            <div className="h-12 border-b border-charcoal-800 bg-[#141820] px-5 flex items-center justify-between shrink-0 select-none">
              <div className="flex items-center gap-2 font-mono text-xs">
                <FileCode className="w-4 h-4 text-dna-blue" />
                <span className="text-white font-bold">{inspectedFile.path}</span>
                <span className="text-zinc-500">• Code Editor View</span>
              </div>

              <button
                onClick={() => setInspectedFile(null)}
                className="p-1 rounded text-zinc-400 hover:text-white hover:bg-charcoal-750 transition-colors"
                title="Close Editor (Esc)"
              >
                <X className="w-4 h-4" />
              </button>
            </div>

            {/* Code Editor Body */}
            <div className="flex-1 p-3 overflow-hidden bg-charcoal-950">
              <CodeEditorViewer
                filePath={inspectedFile.path}
                content={inspectedFile.content}
                targetLine={1}
                maxHeight="h-full"
              />
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
