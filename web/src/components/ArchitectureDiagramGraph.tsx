'use client';

import React, { useMemo, useState, useCallback, useEffect } from 'react';
import ReactFlow, {
  Background,
  Controls,
  MiniMap,
  useNodesState,
  useEdgesState,
  Node,
  Edge,
  MarkerType,
  Position,
  ReactFlowProvider,
  Handle,
} from 'reactflow';
import { 
  SystemDesignReport, 
  SystemTier, 
  DataFlowHop, 
  TierKind 
} from '../lib/api';
import { 
  Globe, 
  Server, 
  Cpu, 
  Database, 
  Radio, 
  ExternalLink, 
  Maximize2, 
  Minimize2,
  RotateCcw, 
  Layers, 
  Share2, 
  Copy, 
  Check, 
  FileCode, 
  X,
  Compass,
  ArrowRight,
  Search,
  Filter,
  Sparkles
} from 'lucide-react';

interface ArchitectureDiagramGraphProps {
  report: SystemDesignReport;
  onSelectTier?: (tier: SystemTier) => void;
  onOpenFile?: (filePath: string) => void;
}

const TIER_ICONS: Record<string, React.ElementType> = {
  client_gateway: Globe,
  api_presentation: Server,
  domain_service: Cpu,
  persistence_storage: Database,
  async_messaging: Radio,
  external_integration: ExternalLink,
};

const TIER_PALETTES: Record<string, { border: string; bg: string; text: string; edge: string }> = {
  client_gateway: {
    border: '#38bdf8', // sky-400
    bg: 'rgba(56, 189, 248, 0.08)',
    text: '#38bdf8',
    edge: '#38bdf8',
  },
  api_presentation: {
    border: '#818cf8', // indigo-400
    bg: 'rgba(129, 140, 248, 0.08)',
    text: '#818cf8',
    edge: '#818cf8',
  },
  domain_service: {
    border: '#34d399', // emerald-400
    bg: 'rgba(52, 211, 153, 0.08)',
    text: '#34d399',
    edge: '#34d399',
  },
  persistence_storage: {
    border: '#fbbf24', // amber-400
    bg: 'rgba(251, 191, 36, 0.08)',
    text: '#fbbf24',
    edge: '#fbbf24',
  },
  async_messaging: {
    border: '#c084fc', // purple-400
    bg: 'rgba(192, 132, 252, 0.08)',
    text: '#c084fc',
    edge: '#c084fc',
  },
  external_integration: {
    border: '#fb7185', // rose-400
    bg: 'rgba(251, 113, 133, 0.08)',
    text: '#fb7185',
    edge: '#fb7185',
  },
};

type ProtocolFilterType = 'ALL' | 'HTTPS' | 'IN_MEMORY' | 'STORAGE' | 'ASYNC';

function matchesProtocolFilter(protocol: string, filter: ProtocolFilterType): boolean {
  if (filter === 'ALL') return true;
  const p = protocol.toLowerCase();
  if (filter === 'HTTPS') return p.includes('http') || p.includes('rest');
  if (filter === 'IN_MEMORY') return p.includes('memory') || p.includes('internal');
  if (filter === 'STORAGE') return p.includes('sql') || p.includes('storage');
  if (filter === 'ASYNC') return p.includes('async') || p.includes('channel') || p.includes('event');
  return true;
}

// Custom Node Component for Architectural Tiers
function TierNodeComponent({ data }: { data: any }) {
  const tier: SystemTier = data.tier;
  const isSelected = data.isSelected;
  const isSearchMatch = data.isSearchMatch;
  const isDimmed = data.isDimmed;
  const matchedFiles: string[] = data.matchedFiles || [];
  const searchTerm: string = data.searchTerm || '';
  const onSelect = data.onSelect;
  const onOpenFile = data.onOpenFile;
  const Icon = TIER_ICONS[tier.id] || Layers;
  const palette = TIER_PALETTES[tier.id] || { border: '#71717a', bg: 'rgba(113, 113, 122, 0.08)', text: '#71717a', edge: '#71717a' };

  // Calculate files to display: if search matched, prioritize matched files
  const displayFiles = useMemo(() => {
    if (!tier.files) return [];
    if (!searchTerm) return tier.files;
    const matched = tier.files.filter(f => f.toLowerCase().includes(searchTerm.toLowerCase()));
    const unmatched = tier.files.filter(f => !f.toLowerCase().includes(searchTerm.toLowerCase()));
    return [...matched, ...unmatched];
  }, [tier.files, searchTerm]);

  return (
    <div
      onClick={() => onSelect && onSelect(tier)}
      className={`w-72 rounded-xl p-4 transition-all duration-200 cursor-pointer shadow-xl ${
        isDimmed ? 'opacity-30 hover:opacity-100' : 'opacity-100'
      } ${
        isSearchMatch
          ? 'ring-2 ring-emerald-400 shadow-emerald-400/20 bg-[#121c17]'
          : isSelected
          ? 'ring-2 ring-dna-blue shadow-dna-blue/20 bg-[#161a22]'
          : 'hover:border-zinc-500 bg-[#101217]'
      }`}
      style={{
        border: `1.5px solid ${isSearchMatch ? '#34d399' : isSelected ? '#3b82f6' : palette.border}88`,
        boxShadow: isSearchMatch
          ? '0 0 25px rgba(52, 211, 153, 0.25)'
          : isSelected
          ? `0 0 24px ${palette.border}33`
          : '0 10px 25px -5px rgba(0, 0, 0, 0.5)',
      }}
    >
      {/* React Flow Connection Handles */}
      <Handle type="target" position={Position.Top} className="!w-2 !h-2 !bg-dna-blue !border-none" />
      <Handle type="source" position={Position.Bottom} className="!w-2 !h-2 !bg-dna-blue !border-none" />
      <Handle type="target" position={Position.Left} className="!w-2 !h-2 !bg-dna-blue !border-none" />
      <Handle type="source" position={Position.Right} className="!w-2 !h-2 !bg-dna-blue !border-none" />

      {/* Header */}
      <div className="flex items-start justify-between gap-3">
        <div className="flex items-center gap-2.5">
          <div 
            className="w-9 h-9 rounded-lg flex items-center justify-center shrink-0 border"
            style={{ 
              backgroundColor: palette.bg,
              borderColor: `${palette.border}44`,
              color: palette.text
            }}
          >
            <Icon className="w-5 h-5" />
          </div>
          <div>
            <div className="flex items-center gap-1.5">
              <span className="text-[10px] font-mono px-1.5 py-0.5 rounded bg-charcoal-800 border border-charcoal-700 text-zinc-400">
                Tier {tier.order}
              </span>
              {isSearchMatch && (
                <span className="text-[9px] font-mono px-1.5 py-0.5 rounded bg-emerald-500/20 text-emerald-300 border border-emerald-500/40">
                  {matchedFiles.length} match{matchedFiles.length > 1 ? 'es' : ''}
                </span>
              )}
              {tier.fileCount === 0 && !isSearchMatch && (
                <span className="text-[9px] font-mono px-1 rounded bg-zinc-800 text-zinc-500">Unused</span>
              )}
            </div>
            <h4 className="text-xs font-bold text-white mt-0.5 truncate max-w-[170px]">{tier.name}</h4>
          </div>
        </div>
      </div>

      {/* Description */}
      <p className="text-[11px] text-zinc-400 mt-2.5 line-clamp-2 leading-relaxed">
        {tier.description}
      </p>

      {/* Stats Bar */}
      <div className="mt-3 pt-2.5 border-t border-charcoal-800 flex items-center justify-between text-[10px] font-mono text-zinc-400">
        <span className="text-zinc-300 font-semibold">{tier.fileCount} Files</span>
        <span>•</span>
        <span>{tier.symbolCount} Symbols</span>
      </div>

      {/* Sample Key Components Pills */}
      {displayFiles.length > 0 && (
        <div className="mt-2.5 space-y-1">
          <div className="text-[9px] font-mono uppercase tracking-wider text-zinc-500">Components:</div>
          <div className="flex flex-wrap gap-1 max-h-16 overflow-hidden">
            {displayFiles.slice(0, 3).map((f) => {
              const isMatch = searchTerm && f.toLowerCase().includes(searchTerm.toLowerCase());
              return (
                <span 
                  key={f}
                  onClick={(e) => {
                    e.stopPropagation();
                    onOpenFile && onOpenFile(f);
                  }}
                  className={`text-[10px] font-mono px-1.5 py-0.5 rounded truncate max-w-[240px] cursor-pointer transition-colors ${
                    isMatch
                      ? 'bg-emerald-500/20 text-emerald-300 border border-emerald-500/40 hover:bg-emerald-500/30'
                      : 'bg-charcoal-800 hover:bg-charcoal-750 text-zinc-300 hover:text-white border border-charcoal-700'
                  }`}
                  title={`Click to inspect ${f} in editor`}
                >
                  {f.split('/').pop()}
                </span>
              );
            })}
            {displayFiles.length > 3 && (
              <span className="text-[9px] font-mono px-1 py-0.5 text-zinc-500">
                +{displayFiles.length - 3} more
              </span>
            )}
          </div>
        </div>
      )}
    </div>
  );
}

const nodeTypes = {
  tierNode: TierNodeComponent,
};

function ArchitectureDiagramGraphInner({
  report,
  onSelectTier,
  onOpenFile,
}: ArchitectureDiagramGraphProps) {
  const [selectedTierId, setSelectedTierId] = useState<string | null>(null);
  const [layoutMode, setLayoutMode] = useState<'pipeline' | 'horizontal'>('pipeline');
  const [protocolFilter, setProtocolFilter] = useState<ProtocolFilterType>('ALL');
  const [searchTerm, setSearchTerm] = useState('');
  const [copiedMermaid, setCopiedMermaid] = useState(false);
  const [showExportModal, setShowExportModal] = useState(false);
  const [isFullscreen, setIsFullscreen] = useState(false);

  // Compute tier matching status based on search query
  const tierSearchStats = useMemo(() => {
    const q = searchTerm.trim().toLowerCase();
    const map: Record<string, { isMatch: boolean; matchedFiles: string[] }> = {};
    if (!report?.tiers) return map;

    report.tiers.forEach((t) => {
      if (!q) {
        map[t.id] = { isMatch: false, matchedFiles: [] };
        return;
      }
      const matchedFiles = (t.files || []).filter((f) => f.toLowerCase().includes(q));
      const matchedSymbols = (t.keySymbols || []).filter((s) => s.toLowerCase().includes(q));
      const isNameMatch = t.name.toLowerCase().includes(q);
      const isMatch = matchedFiles.length > 0 || matchedSymbols.length > 0 || isNameMatch;
      map[t.id] = { isMatch, matchedFiles };
    });
    return map;
  }, [report, searchTerm]);

  const hasSearchFilter = searchTerm.trim().length > 0;

  // Compute Layout Positions
  const initialNodes: Node[] = useMemo(() => {
    if (!report?.tiers) return [];

    // Coordinates for Pipeline layout (Top to bottom flow)
    const pipelineCoords: Record<string, { x: number; y: number }> = {
      client_gateway: { x: 380, y: 30 },
      api_presentation: { x: 380, y: 240 },
      domain_service: { x: 380, y: 460 },
      persistence_storage: { x: 80, y: 690 },
      async_messaging: { x: 420, y: 690 },
      external_integration: { x: 760, y: 690 },
    };

    // Coordinates for Horizontal layout (Left to right flow)
    const horizontalCoords: Record<string, { x: number; y: number }> = {
      client_gateway: { x: 30, y: 240 },
      api_presentation: { x: 360, y: 240 },
      domain_service: { x: 690, y: 240 },
      persistence_storage: { x: 1040, y: 60 },
      async_messaging: { x: 1040, y: 260 },
      external_integration: { x: 1040, y: 460 },
    };

    const coords = layoutMode === 'pipeline' ? pipelineCoords : horizontalCoords;

    return report.tiers.map((t) => {
      const pos = coords[t.id] || { x: 300, y: 300 };
      const searchInfo = tierSearchStats[t.id] || { isMatch: false, matchedFiles: [] };
      const isDimmed = hasSearchFilter && !searchInfo.isMatch;

      return {
        id: t.id,
        type: 'tierNode',
        position: pos,
        data: {
          tier: t,
          isSelected: selectedTierId === t.id,
          isSearchMatch: searchInfo.isMatch,
          isDimmed,
          matchedFiles: searchInfo.matchedFiles,
          searchTerm: searchTerm.trim(),
          onSelect: (selected: SystemTier) => {
            setSelectedTierId(selected.id);
            if (onSelectTier) onSelectTier(selected);
          },
          onOpenFile,
        },
      };
    });
  }, [report, selectedTierId, layoutMode, tierSearchStats, hasSearchFilter, searchTerm, onSelectTier, onOpenFile]);

  // Compute protocol counts for filter pills
  const protocolCounts = useMemo(() => {
    const counts = { ALL: 0, HTTPS: 0, IN_MEMORY: 0, STORAGE: 0, ASYNC: 0 };
    if (!report?.dataFlow) return counts;
    counts.ALL = report.dataFlow.length;
    report.dataFlow.forEach((h) => {
      if (matchesProtocolFilter(h.protocol, 'HTTPS')) counts.HTTPS++;
      if (matchesProtocolFilter(h.protocol, 'IN_MEMORY')) counts.IN_MEMORY++;
      if (matchesProtocolFilter(h.protocol, 'STORAGE')) counts.STORAGE++;
      if (matchesProtocolFilter(h.protocol, 'ASYNC')) counts.ASYNC++;
    });
    return counts;
  }, [report]);

  // Compute Cross-Tier Directed Edges with protocol filtering
  const initialEdges: Edge[] = useMemo(() => {
    if (!report?.dataFlow) return [];

    return report.dataFlow.map((hop, idx) => {
      const isMatch = matchesProtocolFilter(hop.protocol, protocolFilter);
      const fromPalette = TIER_PALETTES[hop.fromTier] || { edge: '#60a5fa' };
      const edgeColor = isMatch ? fromPalette.edge : '#2d3340';

      return {
        id: `flow-${hop.fromTier}-${hop.toTier}-${idx}`,
        source: hop.fromTier,
        target: hop.toTier,
        type: 'smoothstep',
        animated: isMatch,
        style: {
          stroke: edgeColor,
          strokeWidth: isMatch ? Math.min(4, Math.max(1.8, Math.log2(hop.callCount + 1))) : 1,
          strokeDasharray: isMatch ? '5,5' : '2,2',
          opacity: isMatch ? 1 : 0.2,
        },
        markerEnd: {
          type: MarkerType.ArrowClosed,
          color: edgeColor,
          width: isMatch ? 14 : 8,
          height: isMatch ? 14 : 8,
        },
        label: isMatch ? `${hop.protocol} (${hop.callCount})` : '',
        labelStyle: {
          fill: '#e4e4e7',
          fontWeight: 600,
          fontFamily: 'monospace',
          fontSize: 10,
        },
        labelBgStyle: {
          fill: '#181b22',
          fillOpacity: 0.95,
          rx: 6,
          ry: 6,
          stroke: '#272a33',
          strokeWidth: 1,
        },
      };
    });
  }, [report, protocolFilter]);

  const [nodes, setNodes, onNodesChange] = useNodesState(initialNodes);
  const [edges, setEdges, onEdgesChange] = useEdgesState(initialEdges);

  // Sync state if initialNodes changes
  useEffect(() => {
    setNodes(initialNodes);
  }, [initialNodes, setNodes]);

  useEffect(() => {
    setEdges(initialEdges);
  }, [initialEdges, setEdges]);

  // Generate Mermaid Diagram specification
  const mermaidDiagram = useMemo(() => {
    let mmd = 'flowchart TD\n';
    mmd += '  %% CodeDNA System Design Architecture Flow\n';

    report.tiers.forEach((t) => {
      const safeName = t.name.replace(/["&]/g, '');
      mmd += `  ${t.id}["<b>${safeName}</b><br/>(${t.fileCount} files, ${t.symbolCount} symbols)"]\n`;
    });

    mmd += '\n';

    report.dataFlow.forEach((hop) => {
      const proto = hop.protocol.replace(/["&]/g, '');
      mmd += `  ${hop.fromTier} -->|"${proto} (${hop.callCount} calls)"| ${hop.toTier}\n`;
    });

    return mmd;
  }, [report]);

  const handleCopyMermaid = () => {
    navigator.clipboard.writeText(mermaidDiagram);
    setCopiedMermaid(true);
    setTimeout(() => setCopiedMermaid(false), 2000);
  };

  return (
    <div className={`transition-all duration-200 border border-charcoal-750 bg-[#090b0e] overflow-hidden shadow-2xl flex flex-col ${
      isFullscreen 
        ? 'fixed inset-0 z-50 w-screen h-screen rounded-none' 
        : 'relative w-full h-[850px] min-h-[750px] rounded-2xl'
    }`}>
      {/* Canvas Top Bar */}
      <div className="z-10 bg-[#12151c]/95 backdrop-blur-md px-4 py-3 border-b border-charcoal-800 flex flex-wrap items-center justify-between gap-3 shrink-0">
        {/* Left: Title & Search Bar */}
        <div className="flex items-center gap-3 flex-1 min-w-[280px]">
          <div className="flex items-center gap-2 font-mono text-xs text-white shrink-0">
            <Compass className="w-4 h-4 text-dna-blue" />
            <span className="font-bold">Architecture Graph</span>
          </div>

          {/* Component Search Input */}
          <div className="relative flex-1 max-w-sm">
            <Search className="w-3.5 h-3.5 absolute left-3 top-1/2 -translate-y-1/2 text-zinc-400" />
            <input
              type="text"
              placeholder="Filter tiers by file or symbol..."
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              className="w-full pl-8 pr-7 py-1.5 bg-charcoal-900 border border-charcoal-700 focus:border-dna-blue rounded-lg text-xs font-mono text-zinc-200 placeholder-zinc-500 focus:outline-none transition-colors"
            />
            {searchTerm && (
              <button
                onClick={() => setSearchTerm('')}
                className="absolute right-2.5 top-1/2 -translate-y-1/2 text-zinc-500 hover:text-zinc-300"
              >
                <X className="w-3 h-3" />
              </button>
            )}
          </div>
        </div>

        {/* Right: Protocol Filters & Export */}
        <div className="flex items-center gap-2 font-mono text-xs flex-wrap">
          {/* Protocol Filter Pills */}
          <div className="flex items-center gap-1 bg-charcoal-900 p-0.5 rounded-lg border border-charcoal-750 text-[11px]">
            <button
              onClick={() => setProtocolFilter('ALL')}
              className={`px-2.5 py-1 rounded transition-all ${
                protocolFilter === 'ALL' ? 'bg-dna-blue text-white shadow-sm font-semibold' : 'text-zinc-400 hover:text-white'
              }`}
            >
              All ({protocolCounts.ALL})
            </button>
            <button
              onClick={() => setProtocolFilter('HTTPS')}
              className={`px-2 py-1 rounded transition-all ${
                protocolFilter === 'HTTPS' ? 'bg-sky-500/20 text-sky-300 border border-sky-500/40 font-semibold' : 'text-zinc-400 hover:text-white'
              }`}
            >
              HTTPS ({protocolCounts.HTTPS})
            </button>
            <button
              onClick={() => setProtocolFilter('IN_MEMORY')}
              className={`px-2 py-1 rounded transition-all ${
                protocolFilter === 'IN_MEMORY' ? 'bg-emerald-500/20 text-emerald-300 border border-emerald-500/40 font-semibold' : 'text-zinc-400 hover:text-white'
              }`}
            >
              In-Memory ({protocolCounts.IN_MEMORY})
            </button>
            <button
              onClick={() => setProtocolFilter('STORAGE')}
              className={`px-2 py-1 rounded transition-all ${
                protocolFilter === 'STORAGE' ? 'bg-amber-500/20 text-amber-300 border border-amber-500/40 font-semibold' : 'text-zinc-400 hover:text-white'
              }`}
            >
              Storage ({protocolCounts.STORAGE})
            </button>
            <button
              onClick={() => setProtocolFilter('ASYNC')}
              className={`px-2 py-1 rounded transition-all ${
                protocolFilter === 'ASYNC' ? 'bg-purple-500/20 text-purple-300 border border-purple-500/40 font-semibold' : 'text-zinc-400 hover:text-white'
              }`}
            >
              Async ({protocolCounts.ASYNC})
            </button>
          </div>

          {/* Layout Switcher */}
          <div className="flex items-center gap-1 bg-charcoal-900 p-0.5 rounded-lg border border-charcoal-750 text-[11px]">
            <button
              onClick={() => setLayoutMode('pipeline')}
              className={`px-2 py-1 rounded transition-all ${
                layoutMode === 'pipeline' ? 'bg-charcoal-750 text-white font-semibold' : 'text-zinc-400 hover:text-white'
              }`}
            >
              Pipeline
            </button>
            <button
              onClick={() => setLayoutMode('horizontal')}
              className={`px-2 py-1 rounded transition-all ${
                layoutMode === 'horizontal' ? 'bg-charcoal-750 text-white font-semibold' : 'text-zinc-400 hover:text-white'
              }`}
            >
              Layered
            </button>
          </div>

          <button
            onClick={() => setShowExportModal(true)}
            className="flex items-center gap-1.5 px-2.5 py-1.5 rounded-lg bg-charcoal-800 hover:bg-charcoal-750 border border-charcoal-700 text-zinc-200 hover:text-white transition-all text-xs"
            title="Export architecture diagram as Mermaid.js"
          >
            <Share2 className="w-3.5 h-3.5 text-dna-purple" />
            <span>Mermaid</span>
          </button>

          {/* Fullscreen Canvas Toggle */}
          <button
            onClick={() => setIsFullscreen(!isFullscreen)}
            className="flex items-center gap-1.5 px-2.5 py-1.5 rounded-lg bg-charcoal-800 hover:bg-charcoal-750 border border-charcoal-700 text-zinc-200 hover:text-white transition-all text-xs"
            title={isFullscreen ? "Exit Fullscreen (Esc)" : "Expand Canvas to Fullscreen"}
          >
            {isFullscreen ? <Minimize2 className="w-3.5 h-3.5 text-dna-blue" /> : <Maximize2 className="w-3.5 h-3.5 text-dna-blue" />}
            <span>{isFullscreen ? 'Exit' : 'Fullscreen'}</span>
          </button>
        </div>
      </div>

      {/* React Flow Diagram Canvas */}
      <div className="flex-1 relative">
        <ReactFlow
          nodes={nodes}
          edges={edges}
          nodeTypes={nodeTypes}
          onNodesChange={onNodesChange}
          onEdgesChange={onEdgesChange}
          fitView
          fitViewOptions={{ padding: 0.15 }}
          minZoom={0.2}
          maxZoom={1.8}
          attributionPosition="bottom-right"
        >
          <Background color="#1a1e27" gap={20} size={1} />
          <Controls className="!bg-[#14171e] !border-charcoal-700 !fill-zinc-300 !rounded-xl" />
          <MiniMap
            nodeColor={(n) => {
              const palette = TIER_PALETTES[n.id];
              return palette ? palette.border : '#52525b';
            }}
            className="!bg-[#14171e] !border-charcoal-700 !rounded-xl"
          />
        </ReactFlow>

        {/* Dynamic Legend Indicator */}
        <div className="absolute bottom-4 left-4 z-10 pointer-events-auto hidden md:flex items-center gap-3 bg-[#14171e]/90 backdrop-blur-md px-3.5 py-2 rounded-xl border border-charcoal-700 text-[11px] font-mono text-zinc-300 shadow-lg">
          <span className="text-zinc-500">Legend:</span>
          <span className="flex items-center gap-1.5">
            <span className="w-2 h-2 rounded-full bg-sky-400" />
            <span>HTTPS/REST</span>
          </span>
          <span className="flex items-center gap-1.5">
            <span className="w-2 h-2 rounded-full bg-emerald-400" />
            <span>In-Memory</span>
          </span>
          <span className="flex items-center gap-1.5">
            <span className="w-2 h-2 rounded-full bg-amber-400" />
            <span>SQL/Storage</span>
          </span>
          <span className="flex items-center gap-1.5">
            <span className="w-2 h-2 rounded-full bg-purple-400" />
            <span>Async Event</span>
          </span>
        </div>
      </div>

      {/* Mermaid Export Modal */}
      {showExportModal && (
        <div className="fixed inset-0 z-50 bg-black/80 backdrop-blur-sm flex items-center justify-center p-4">
          <div className="bg-[#12151b] border border-charcoal-700 rounded-2xl w-full max-w-2xl overflow-hidden shadow-2xl space-y-4 p-6">
            <div className="flex items-center justify-between border-b border-charcoal-800 pb-4">
              <div className="flex items-center gap-2.5">
                <Share2 className="w-5 h-5 text-dna-blue" />
                <h3 className="text-base font-bold text-white">System Architecture Flow Diagram (Mermaid)</h3>
              </div>
              <button
                onClick={() => setShowExportModal(false)}
                className="p-1 rounded text-zinc-400 hover:text-white hover:bg-charcoal-800 transition-colors"
              >
                <X className="w-4 h-4" />
              </button>
            </div>

            <p className="text-xs text-zinc-400">
              Copy this standard Mermaid.js flowchart specification to embed in GitHub READMEs, RFCs, and markdown documentation.
            </p>

            <div className="bg-[#090b0e] border border-charcoal-800 rounded-xl p-4 font-mono text-xs text-zinc-300 max-h-72 overflow-y-auto">
              <pre>{mermaidDiagram}</pre>
            </div>

            <div className="flex items-center justify-end gap-3 pt-2">
              <button
                onClick={() => setShowExportModal(false)}
                className="px-4 py-2 rounded-lg bg-charcoal-800 hover:bg-charcoal-750 text-xs font-mono text-zinc-300 hover:text-white"
              >
                Close
              </button>
              <button
                onClick={handleCopyMermaid}
                className="flex items-center gap-2 px-4 py-2 rounded-lg bg-dna-blue hover:bg-blue-600 text-xs font-mono text-white font-semibold transition-all shadow-md shadow-dna-blue/20"
              >
                {copiedMermaid ? <Check className="w-4 h-4 text-emerald-300" /> : <Copy className="w-4 h-4" />}
                <span>{copiedMermaid ? 'Copied to Clipboard!' : 'Copy Mermaid Code'}</span>
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

export default function ArchitectureDiagramGraph(props: ArchitectureDiagramGraphProps) {
  return (
    <ReactFlowProvider>
      <ArchitectureDiagramGraphInner {...props} />
    </ReactFlowProvider>
  );
}
