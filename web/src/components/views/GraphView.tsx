'use client';

import React, { useEffect, useState, useMemo, useCallback } from 'react';
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
  Handle,
  ReactFlowProvider,
  EdgeProps,
  getBezierPath,
  getSmoothStepPath,
} from 'reactflow';
import { useAppStore } from '../../lib/store';
import { 
  fetchGraph, 
  fetchWorkspaceGraph, 
  fetchExportGraph, 
  fetchFileContent,
  GraphNode, 
  GraphEdge 
} from '../../lib/api';
import CodeEditorViewer from '../CodeEditorViewer';
import { 
  Search, 
  Filter, 
  Layers, 
  Maximize2, 
  RotateCcw, 
  GitFork,
  Flame, 
  AlertTriangle, 
  ZoomIn, 
  Download, 
  Copy, 
  Check, 
  X, 
  FileCode, 
  Share2,
  Compass,
  ArrowRight,
  ArrowLeft,
  Code2,
  Activity,
  Eye,
  EyeOff,
  Sparkles,
  ChevronRight,
  Sliders,
  ExternalLink,
  ShieldCheck,
  Zap
} from 'lucide-react';

// ======================================================================
// 1. ELECTRON CIRCUIT EDGE COMPONENT (Dashed Circuit Line with Moving Electrons)
// ======================================================================
interface ElectronCircuitEdgeData {
  isFocused?: boolean;
  isIncoming?: boolean;
  isOutgoing?: boolean;
  isDimmed?: boolean;
  kind?: string;
  label?: string;
  wireRouting?: 'circuit' | 'smooth';
  enableElectrons?: boolean;
}

const ElectronCircuitEdge = React.memo(function ElectronCircuitEdge({
  id,
  sourceX,
  sourceY,
  targetX,
  targetY,
  sourcePosition = Position.Right,
  targetPosition = Position.Left,
  data,
}: EdgeProps<ElectronCircuitEdgeData>) {
  const isFocused = data?.isFocused;
  const isIncoming = data?.isIncoming;
  const isOutgoing = data?.isOutgoing;
  const isDimmed = data?.isDimmed;
  const kind = data?.kind || 'calls';
  const wireRouting = data?.wireRouting || 'circuit';
  const enableElectrons = data?.enableElectrons !== false;

  let edgePath: string;
  let labelX: number;
  let labelY: number;

  if (wireRouting === 'circuit') {
    [edgePath, labelX, labelY] = getSmoothStepPath({
      sourceX,
      sourceY,
      sourcePosition,
      targetX,
      targetY,
      targetPosition,
      borderRadius: 16,
    });
  } else {
    const dx = Math.abs(targetX - sourceX);
    const curvature = Math.min(0.35, Math.max(0.18, 120 / (dx + 1)));
    [edgePath, labelX, labelY] = getBezierPath({
      sourceX,
      sourceY,
      sourcePosition,
      targetX,
      targetY,
      targetPosition,
      curvature,
    });
  }

  // When dimmed, render single minimalist faint dashed line without animations
  // for maximum GPU/CPU performance during zoom & pan
  if (isDimmed) {
    return (
      <path
        d={edgePath}
        fill="none"
        stroke="#272a33"
        strokeWidth={1}
        strokeDasharray="4 4"
        strokeOpacity={0.12}
      />
    );
  }

  // Determine line style and electron colors
  let trackColor = '#475569'; // Muted dark circuit track
  let trackOpacity = 0.35;
  let electronColor = '#38bdf8'; // Cyan electron
  let markerId = 'circuit-marker-muted';
  let strokeDash = '6 4';
  let strokeWidth = 1.4;

  if (kind === 'imports') {
    trackColor = '#7c3aed';
    electronColor = '#c084fc';
    markerId = 'circuit-marker-purple';
    strokeDash = '4 4';
    trackOpacity = 0.5;
  } else if (kind === 'cross_service_api') {
    trackColor = '#0284c7';
    electronColor = '#38bdf8';
    markerId = 'circuit-marker-cyan';
    strokeDash = '6 3';
    trackOpacity = 0.55;
  }

  // When focused, spotlight caller / callee paths
  if (isFocused) {
    strokeWidth = 2.2;
    strokeDash = '6 3';
    if (isIncoming) {
      trackColor = '#10b981'; // Bright emerald incoming caller
      electronColor = '#34d399';
      markerId = 'circuit-marker-emerald';
      trackOpacity = 0.95;
    } else {
      trackColor = '#06b6d4'; // Bright cyan outgoing callee
      electronColor = '#38bdf8';
      markerId = 'circuit-marker-cyan';
      trackOpacity = 0.95;
    }
  }

  const speed = isFocused ? '1.2s' : '2.8s';
  const electronRadius = isFocused ? 2.5 : 2.0;

  return (
    <g className="electron-circuit-edge group cursor-pointer">
      {/* Invisible wider hit area for easy hover and click interaction */}
      <path
        d={edgePath}
        fill="none"
        stroke="transparent"
        strokeWidth={14}
        className="cursor-pointer"
      />

      {/* Layer 1: Dashed Circuit Wire Track (Kata Kata Line) */}
      <path
        id={id}
        d={edgePath}
        fill="none"
        stroke={trackColor}
        strokeWidth={strokeWidth}
        strokeDasharray={strokeDash}
        strokeOpacity={trackOpacity}
        strokeLinecap="round"
        markerEnd={`url(#${isFocused ? markerId : (kind === 'imports' ? 'circuit-marker-purple' : 'circuit-marker-muted')})`}
      />

      {/* Layer 2: Traveling Electron (Data Packet) */}
      {enableElectrons && (
        <g pointerEvents="none">
          {/* Electron Glow Aura */}
          <circle
            r={electronRadius + 2.5}
            fill={electronColor}
            opacity={isFocused ? 0.65 : 0.4}
          >
            <animateMotion dur={speed} repeatCount="indefinite" path={edgePath} />
          </circle>

          {/* High-energy white core */}
          <circle
            r={electronRadius}
            fill="#ffffff"
            opacity={isFocused ? 1 : 0.9}
          >
            <animateMotion dur={speed} repeatCount="indefinite" path={edgePath} />
          </circle>

          {/* Secondary trailing electron on focused active paths */}
          {isFocused && (
            <circle
              r={electronRadius * 0.75}
              fill={electronColor}
              opacity={0.7}
            >
              <animateMotion dur={speed} begin="0.6s" repeatCount="indefinite" path={edgePath} />
            </circle>
          )}
        </g>
      )}

      {/* Optional Interactive Wire Label on focus */}
      {data?.label && (
        <text
          x={labelX}
          y={labelY - 6}
          fill={isIncoming ? '#34d399' : '#38bdf8'}
          fontSize={10}
          fontFamily="monospace"
          fontWeight="bold"
          textAnchor="middle"
          className="select-none pointer-events-none"
        >
          {data.label}
        </text>
      )}
    </g>
  );
});

const edgeTypes = {
  neonLighting: ElectronCircuitEdge,
  electronCircuit: ElectronCircuitEdge,
};

// ======================================================================
// 2. CUSTOM GRAPH NODE COMPONENT (High-Tech Function & Symbol Card)
// ======================================================================
const CustomGraphNode = React.memo(function CustomGraphNode({ data }: { data: any }) {
  const gn: GraphNode = data.node;
  const isFocused = data.isFocused;
  const isCaller = data.isCaller;
  const isCallee = data.isCallee;
  const isDimmed = data.isDimmed;
  const onSelect = data.onSelect;
  const onOpenFile = data.onOpenFile;
  const onIsolate = data.onIsolate;

  const isFuncOrMethod = gn.type === 'function' || gn.type === 'method';
  const isMethod = gn.type === 'method' || !!gn.data?.parentClassName;
  const parentClass = gn.data?.parentClassName || gn.data?.parentClass;
  const complexity = gn.complexity || 0;
  const inCount = gn.data?.inDegree ?? 0;
  const outCount = gn.data?.outDegree ?? 0;
  const connections = gn.data?.connections ?? (inCount + outCount);

  // Border & glow styling based on focus / caller / callee relationship
  let ringStyle = 'border-charcoal-750 hover:border-zinc-500 bg-[#12141a]';
  if (isFocused) {
    ringStyle = 'border-dna-blue ring-2 ring-dna-blue bg-[#141b2d] shadow-2xl shadow-dna-blue/40';
  } else if (isCaller) {
    ringStyle = 'border-emerald-400 ring-2 ring-emerald-400/80 bg-[#102419] shadow-xl shadow-emerald-400/25';
  } else if (isCallee) {
    ringStyle = 'border-sky-400 ring-2 ring-sky-400/80 bg-[#0e1f2b] shadow-xl shadow-sky-400/25';
  }

  return (
    <div
      onClick={() => onSelect && onSelect(gn)}
      className={`w-72 rounded-xl p-3.5 transition-all duration-200 cursor-pointer shadow-lg select-none ${ringStyle} ${
        isDimmed ? 'opacity-25 hover:opacity-100' : 'opacity-100'
      }`}
    >
      {/* React Flow Connection Handles on Left and Right for clean DAG flow */}
      <Handle type="target" position={Position.Left} className="!w-2.5 !h-2.5 !bg-emerald-400 !border-none" />
      <Handle type="source" position={Position.Right} className="!w-2.5 !h-2.5 !bg-sky-400 !border-none" />
      <Handle type="target" position={Position.Top} className="!w-2 !h-2 !bg-zinc-500 !border-none" />
      <Handle type="source" position={Position.Bottom} className="!w-2 !h-2 !bg-zinc-500 !border-none" />

      {/* Top Header */}
      <div className="flex items-center justify-between gap-2 mb-1.5">
        <div className="flex items-center gap-1.5 flex-wrap">
          <span className={`text-[10px] font-mono uppercase px-1.5 py-0.5 rounded font-bold border ${
            gn.type === 'file'
              ? 'bg-blue-500/15 text-blue-300 border-blue-500/30'
              : gn.type === 'class'
              ? 'bg-amber-500/15 text-amber-300 border-amber-500/30'
              : isMethod 
              ? 'bg-purple-500/15 text-purple-300 border-purple-500/30'
              : 'bg-emerald-500/15 text-emerald-300 border-emerald-500/30'
          }`}>
            {isMethod ? 'method' : gn.type}
          </span>

          {parentClass && (
            <span className="text-[10px] font-mono text-zinc-400 truncate max-w-[85px]" title={parentClass}>
              ({parentClass})
            </span>
          )}
        </div>

        {/* Relationship Status Badge */}
        {isFocused ? (
          <span className="text-[9px] font-mono font-bold px-1.5 py-0.5 rounded bg-dna-blue text-white shadow-sm flex items-center gap-1">
            <Zap className="w-2.5 h-2.5" />
            <span>FOCUSED</span>
          </span>
        ) : isCaller ? (
          <span className="text-[9px] font-mono font-bold px-1.5 py-0.5 rounded bg-emerald-500/20 text-emerald-300 border border-emerald-500/40">
            ▲ CALLER
          </span>
        ) : isCallee ? (
          <span className="text-[9px] font-mono font-bold px-1.5 py-0.5 rounded bg-sky-500/20 text-sky-300 border border-sky-500/40">
            ▼ CALLEE
          </span>
        ) : complexity > 15 ? (
          <span className="flex items-center text-[10px] font-mono text-rose-400 font-bold" title={`Complexity: ${complexity}`}>
            <Flame className="w-3 h-3 mr-0.5 text-rose-500" />
            {complexity}
          </span>
        ) : null}
      </div>

      {/* Symbol / File Name */}
      <div className="text-xs font-mono font-bold text-white truncate max-w-[260px] flex items-center gap-1.5">
        <Code2 className="w-3.5 h-3.5 text-dna-blue shrink-0" />
        <span className="truncate">{gn.label}</span>
      </div>

      {/* Package & File Path */}
      {gn.file && (
        <div className="text-[10px] font-mono text-zinc-400 truncate mt-1 flex items-center justify-between">
          <span className="truncate" title={gn.file}>
            {gn.file}{gn.line ? `:${gn.line}` : ''}
          </span>
        </div>
      )}

      {/* Stats Bar & Quick Actions */}
      <div className="mt-2.5 pt-2 border-t border-charcoal-800 flex items-center justify-between text-[10px] font-mono text-zinc-400">
        <div className="flex items-center gap-2">
          {isFuncOrMethod ? (
            <>
              <span className={inCount > 0 ? 'text-emerald-400 font-semibold' : 'text-zinc-500'} title="Callers (Incoming invocations)">
                ↓ {inCount} in
              </span>
              <span>•</span>
              <span className={outCount > 0 ? 'text-sky-400 font-semibold' : 'text-zinc-500'} title="Calls (Outgoing invocations)">
                ↑ {outCount} out
              </span>
            </>
          ) : (
            <span>{connections} relationships</span>
          )}
        </div>

        <div className="flex items-center gap-1">
          {isFocused && (
            <button
              onClick={(e) => {
                e.stopPropagation();
                onIsolate && onIsolate(gn);
              }}
              className="px-1.5 py-0.5 rounded bg-dna-blue/20 hover:bg-dna-blue/30 text-dna-blue border border-dna-blue/40 text-[9px] font-mono transition-colors"
              title="Isolate this call chain"
            >
              Isolate
            </button>
          )}

          {gn.file && (
            <button
              onClick={(e) => {
                e.stopPropagation();
                onOpenFile && onOpenFile(gn.file, gn.line);
              }}
              className="px-1.5 py-0.5 rounded bg-charcoal-800 hover:bg-charcoal-700 text-zinc-300 hover:text-white border border-charcoal-700 text-[9px] font-mono transition-colors"
              title="Inspect source code in editor"
            >
              Code
            </button>
          )}
        </div>
      </div>
    </div>
  );
});

const nodeTypes = {
  customNode: CustomGraphNode,
};

// ======================================================================
// 3. TOPOLOGICAL LEVEL DAG CALCULATOR (Left ➔ Right Call Pipeline)
// ======================================================================
function computeCallDagLevels(
  nodes: GraphNode[],
  edges: GraphEdge[]
): { levels: Map<string, number>; maxLevel: number } {
  const nodeSet = new Set(nodes.map((n) => n.id));
  const validEdges = edges.filter((e) => nodeSet.has(e.source) && nodeSet.has(e.target));

  const adj = new Map<string, string[]>();
  const inDeg = new Map<string, number>();

  nodes.forEach((n) => {
    adj.set(n.id, []);
    inDeg.set(n.id, 0);
  });

  validEdges.forEach((e) => {
    adj.get(e.source)?.push(e.target);
    inDeg.set(e.target, (inDeg.get(e.target) || 0) + 1);
  });

  const levels = new Map<string, number>();
  const queue: Array<{ id: string; level: number }> = [];

  // Seed root callers (in-degree == 0)
  nodes.forEach((n) => {
    if ((inDeg.get(n.id) || 0) === 0) {
      levels.set(n.id, 0);
      queue.push({ id: n.id, level: 0 });
    }
  });

  if (queue.length === 0 && nodes.length > 0) {
    levels.set(nodes[0].id, 0);
    queue.push({ id: nodes[0].id, level: 0 });
  }

  let maxLevel = 0;
  const visited = new Set<string>();

  while (queue.length > 0) {
    const { id, level } = queue.shift()!;
    visited.add(id);
    if (level > maxLevel) maxLevel = level;

    const neighbors = adj.get(id) || [];
    for (const nxt of neighbors) {
      const nextLevel = Math.max(levels.get(nxt) ?? 0, level + 1);
      // Cap at 8 depth levels to prevent infinite loops from cyclic calls
      if (nextLevel < 8) {
        levels.set(nxt, nextLevel);
        if (!visited.has(nxt)) {
          queue.push({ id: nxt, level: nextLevel });
        }
      }
    }
  }

  nodes.forEach((n) => {
    if (!levels.has(n.id)) {
      levels.set(n.id, 0);
    }
  });

  return { levels, maxLevel };
}

// Architectural Tier Ordering for Domain Cluster Layout
const DOMAIN_TIER_ORDER: Record<string, number> = {
  'cmd/codedna': 0,
  'internal/api': 1,
  'internal/engine': 2,
  'internal/parser': 2,
  'internal/scanner': 2,
  'internal/sysdesign': 3,
  'internal/architecture': 3,
  'internal/analysis': 3,
  'internal/impact': 3,
  'internal/ai': 3,
  'internal/store': 4,
  'internal/watcher': 4,
  'internal/runtime': 4,
  'internal/blame': 4,
  'internal/report': 4,
  'web/src': 5,
};

// ======================================================================
// 4. MAIN GRAPH VIEW COMPONENT
// ======================================================================
export default function GraphView() {
  const { scanId, setSelectedNode, workspaceId, setWorkspaceId, refreshKey } = useAppStore();
  const [graphData, setGraphData] = useState<{ nodes: GraphNode[]; edges: GraphEdge[] }>({ nodes: [], edges: [] });
  const [nodes, setNodes, onNodesChange] = useNodesState([]);
  const [edges, setEdges, onEdgesChange] = useEdgesState([]);

  // Filters & layout state
  const [search, setSearch] = useState('');
  const [selectedType, setSelectedType] = useState<string>('function');
  const [selectedLanguage, setSelectedLanguage] = useState<'all' | 'go' | 'typescript'>('all');
  const [selectedPackage, setSelectedPackage] = useState<string>('all');
  const [layoutMode, setLayoutMode] = useState<'dag' | 'domain' | 'grid'>('dag');
  const [densityMode, setDensityMode] = useState<'core' | 'all'>('core');
  const [wireRouting, setWireRouting] = useState<'circuit' | 'smooth'>('circuit');
  const [enableElectrons, setEnableElectrons] = useState(true);
  const [connectedOnly, setConnectedOnly] = useState(true);
  const [hideTests, setHideTests] = useState(true);
  const [focusedNodeId, setFocusedNodeId] = useState<string | null>(null);
  const [isolatedSubgraph, setIsolatedSubgraph] = useState(false);
  const [loading, setLoading] = useState(true);

  // Embedded Code Editor modal
  const [inspectedFile, setInspectedFile] = useState<{ path: string; content: string; targetLine?: number } | null>(null);
  const [loadingFile, setLoadingFile] = useState(false);

  // Diagram Export Modal state
  const [showExportModal, setShowExportModal] = useState(false);
  const [exportFormat, setExportFormat] = useState<'mermaid' | 'c4'>('mermaid');
  const [exportCode, setExportCode] = useState('');
  const [exportLoading, setExportLoading] = useState(false);
  const [copiedExport, setCopiedExport] = useState(false);

  // Open file in code editor
  const handleOpenFile = async (filePath: string, line?: number) => {
    setLoadingFile(true);
    try {
      const data = await fetchFileContent(scanId, filePath);
      setInspectedFile({ path: data.path, content: data.content, targetLine: line || 1 });
    } catch (err) {
      console.error('Failed to load file content:', err);
    } finally {
      setLoadingFile(false);
    }
  };

  const handleOpenExport = async () => {
    setShowExportModal(true);
    setExportLoading(true);
    try {
      const data = await fetchExportGraph(scanId, exportFormat);
      setExportCode(data.content || '');
    } catch (err) {
      console.error(err);
      setExportCode('// Failed to generate diagram export');
    } finally {
      setExportLoading(false);
    }
  };

  const handleFormatChange = async (fmt: 'mermaid' | 'c4') => {
    setExportFormat(fmt);
    setExportLoading(true);
    try {
      const data = await fetchExportGraph(scanId, fmt);
      setExportCode(data.content || '');
    } catch (err) {
      console.error(err);
    } finally {
      setExportLoading(false);
    }
  };

  const handleCopyExport = () => {
    navigator.clipboard.writeText(exportCode);
    setCopiedExport(true);
    setTimeout(() => setCopiedExport(false), 2000);
  };

  const handleDownloadExport = () => {
    const ext = exportFormat === 'mermaid' ? 'mmd' : 'c4';
    const blob = new Blob([exportCode], { type: 'text/plain' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `architecture-${exportFormat}.${ext}`;
    a.click();
    URL.revokeObjectURL(url);
  };

  // Load graph data with type param from API
  useEffect(() => {
    setLoading(true);
    const params = selectedType !== 'all' ? { type: selectedType } : undefined;
    const req = workspaceId ? fetchWorkspaceGraph(workspaceId) : fetchGraph(scanId, params);
    req
      .then((data) => {
        setGraphData(data || { nodes: [], edges: [] });
      })
      .catch((err) => console.error('Failed to fetch graph:', err))
      .finally(() => setLoading(false));
  }, [scanId, workspaceId, selectedType, refreshKey]);

  // Extract distinct package list from loaded nodes
  const packageList = useMemo(() => {
    const pkgs = new Set<string>();
    (graphData.nodes || []).forEach((n) => {
      if (n.file && n.file.includes('/')) {
        const parts = n.file.split('/');
        pkgs.add(parts.slice(0, 2).join('/'));
      }
    });
    return Array.from(pkgs).sort();
  }, [graphData.nodes]);

  // Compute Caller and Callee sets when a node is focused
  const { callerNodeIds, calleeNodeIds, focusedNode } = useMemo(() => {
    if (!focusedNodeId) {
      return { callerNodeIds: new Set<string>(), calleeNodeIds: new Set<string>(), focusedNode: null };
    }
    const callers = new Set<string>();
    const callees = new Set<string>();

    (graphData.edges || []).forEach((e) => {
      if (e.target === focusedNodeId) {
        callers.add(e.source);
      }
      if (e.source === focusedNodeId) {
        callees.add(e.target);
      }
    });

    const target = (graphData.nodes || []).find((n) => n.id === focusedNodeId) || null;
    return { callerNodeIds: callers, calleeNodeIds: callees, focusedNode: target };
  }, [focusedNodeId, graphData]);

  // Layout & React Flow graph computation
  useEffect(() => {
    if (!graphData.nodes || graphData.nodes.length === 0) {
      setNodes([]);
      setEdges([]);
      return;
    }

    // 1. Initial Filtering
    let filtered = graphData.nodes;

    // Type filter
    if (selectedType !== 'all') {
      if (selectedType === 'function') {
        filtered = filtered.filter((n) => n.type === 'function' || n.type === 'method');
      } else {
        filtered = filtered.filter((n) => n.type === selectedType);
      }
    }

    // Language filter
    if (selectedLanguage !== 'all') {
      if (selectedLanguage === 'go') {
        filtered = filtered.filter((n) => n.file?.endsWith('.go'));
      } else if (selectedLanguage === 'typescript') {
        filtered = filtered.filter((n) => n.file?.endsWith('.ts') || n.file?.endsWith('.tsx') || n.file?.endsWith('.js'));
      }
    }

    // Hide tests filter
    if (hideTests) {
      filtered = filtered.filter((n) => {
        if (!n.file) return true;
        const f = n.file.toLowerCase();
        return !f.includes('_test.go') && !f.includes('.test.') && !f.includes('.spec.');
      });
    }

    // Package filter: include package functions + external callers/callees to keep connections alive!
    if (selectedPackage !== 'all') {
      const pkgNodeIds = new Set(
        filtered.filter((n) => n.file && n.file.startsWith(selectedPackage)).map((n) => n.id)
      );
      // Keep package nodes + direct boundary callers & callees
      const boundaryNodeIds = new Set<string>();
      (graphData.edges || []).forEach((e) => {
        if (pkgNodeIds.has(e.source)) boundaryNodeIds.add(e.target);
        if (pkgNodeIds.has(e.target)) boundaryNodeIds.add(e.source);
      });
      filtered = filtered.filter((n) => pkgNodeIds.has(n.id) || boundaryNodeIds.has(n.id));
    }

    // Search query filter
    if (search.trim()) {
      const q = search.toLowerCase();
      filtered = filtered.filter((n) => 
        n.label.toLowerCase().includes(q) || (n.file && n.file.toLowerCase().includes(q))
      );
    }

    // Isolate Subgraph mode: show ONLY focused node + its direct callers & callees
    if (isolatedSubgraph && focusedNodeId) {
      filtered = filtered.filter(
        (n) => n.id === focusedNodeId || callerNodeIds.has(n.id) || calleeNodeIds.has(n.id)
      );
    }

    const filteredIds = new Set(filtered.map((n) => n.id));

    // 2. Filter edges strictly within visible nodes
    const validEdges = (graphData.edges || []).filter(
      (e) => filteredIds.has(e.source) && filteredIds.has(e.target)
    );

    // Compute active degree within this visible subgraph
    const activeDegree = new Map<string, number>();
    validEdges.forEach((e) => {
      activeDegree.set(e.source, (activeDegree.get(e.source) || 0) + 1);
      activeDegree.set(e.target, (activeDegree.get(e.target) || 0) + 1);
    });

    // If connectedOnly is true and viewing functions, hide isolated zero-degree nodes
    if (connectedOnly && (selectedType === 'function' || selectedType === 'all') && !isolatedSubgraph) {
      filtered = filtered.filter((n) => (activeDegree.get(n.id) || 0) > 0 || n.id === focusedNodeId);
    }

    // Density filter: In 'core' mode, focus on the architectural backbone (functions with degree >= 2, high complexity, or focused)
    // This eliminates visual spaghetti while preserving key calling flows
    if (densityMode === 'core' && selectedType === 'function' && !isolatedSubgraph && selectedPackage === 'all' && !search.trim()) {
      filtered = filtered.filter((n) => {
        const deg = activeDegree.get(n.id) || 0;
        const totalConn = (n.data?.connections as number) || deg;
        return deg >= 2 || totalConn >= 3 || (n.complexity || 0) > 8 || n.id === focusedNodeId || callerNodeIds.has(n.id) || calleeNodeIds.has(n.id);
      });
    }

    const finalNodeIds = new Set(filtered.map((n) => n.id));
    const finalEdges = validEdges.filter(
      (e) => finalNodeIds.has(e.source) && finalNodeIds.has(e.target)
    );

    // 3. Compute Positions based on selected Layout Mode
    const positions = new Map<string, { x: number; y: number }>();

    if (isolatedSubgraph && focusedNodeId) {
      // Clean 3-tier Focal Subgraph Layout:
      // Left: Callers (Incoming) ➔ Center: Target Function ➔ Right: Callees (Outgoing)
      const callersList = filtered.filter((n) => callerNodeIds.has(n.id));
      const calleesList = filtered.filter((n) => calleeNodeIds.has(n.id));
      const targetList = filtered.filter((n) => n.id === focusedNodeId);

      const maxH = Math.max(callersList.length, calleesList.length, 1);
      const centerY = (maxH * 140) / 2 + 50;

      // Position Target in center
      targetList.forEach((gn) => {
        positions.set(gn.id, { x: 420, y: centerY });
      });

      // Position Callers on left
      callersList.forEach((gn, idx) => {
        positions.set(gn.id, { x: 60, y: idx * 140 + 70 });
      });

      // Position Callees on right
      calleesList.forEach((gn, idx) => {
        positions.set(gn.id, { x: 780, y: idx * 140 + 70 });
      });

    } else if (layoutMode === 'dag') {
      // Topological Level DAG layout with Sub-Column Wrapping (Prevents 30,000px towers!)
      const { levels } = computeCallDagLevels(filtered, finalEdges);
      const levelBuckets = new Map<number, GraphNode[]>();

      filtered.forEach((gn) => {
        const lvl = levels.get(gn.id) ?? 0;
        if (!levelBuckets.has(lvl)) levelBuckets.set(lvl, []);
        levelBuckets.get(lvl)!.push(gn);
      });

      // Sort each level by package and complexity
      levelBuckets.forEach((bucket) => {
        bucket.sort((a, b) => {
          const pkgA = a.file ? a.file.split('/').slice(0, 2).join('/') : '';
          const pkgB = b.file ? b.file.split('/').slice(0, 2).join('/') : '';
          if (pkgA !== pkgB) return pkgA.localeCompare(pkgB);
          const degA = (a.data?.connections as number) || 0;
          const degB = (b.data?.connections as number) || 0;
          return degB - degA;
        });
      });

      const MAX_ROWS_PER_COL = 7;
      let currentX = 60;
      const sortedLevels = Array.from(levelBuckets.keys()).sort((a, b) => a - b);

      sortedLevels.forEach((lvl) => {
        const bucket = levelBuckets.get(lvl) || [];
        const subCols = Math.ceil(bucket.length / MAX_ROWS_PER_COL) || 1;

        bucket.forEach((gn, idx) => {
          const subCol = Math.floor(idx / MAX_ROWS_PER_COL);
          const row = idx % MAX_ROWS_PER_COL;
          const x = currentX + subCol * 320;
          const y = row * 145 + 70;
          positions.set(gn.id, { x, y });
        });

        currentX += subCols * 320 + 80;
      });

    } else if (layoutMode === 'domain') {
      // Domain Clusters Layout: Group by Architecture Tier
      const pkgMap = new Map<string, GraphNode[]>();
      filtered.forEach((gn) => {
        const pkg = gn.file ? (gn.file.includes('/') ? gn.file.split('/').slice(0, 2).join('/') : 'root') : 'root';
        if (!pkgMap.has(pkg)) pkgMap.set(pkg, []);
        pkgMap.get(pkg)!.push(gn);
      });

      const sortedPkgs = Array.from(pkgMap.keys()).sort((a, b) => {
        const tierA = DOMAIN_TIER_ORDER[a] ?? 10;
        const tierB = DOMAIN_TIER_ORDER[b] ?? 10;
        if (tierA !== tierB) return tierA - tierB;
        return a.localeCompare(b);
      });

      let clusterX = 60;
      sortedPkgs.forEach((pkg) => {
        const list = pkgMap.get(pkg) || [];
        list.sort((a, b) => (b.complexity || 0) - (a.complexity || 0));

        const subCols = Math.ceil(list.length / 6) || 1;
        list.forEach((gn, idx) => {
          const subCol = Math.floor(idx / 6);
          const row = idx % 6;
          positions.set(gn.id, {
            x: clusterX + subCol * 320,
            y: row * 145 + 80,
          });
        });
        clusterX += subCols * 320 + 90;
      });

    } else {
      // Compact Grid layout
      const cols = Math.ceil(Math.sqrt(filtered.length)) || 4;
      filtered.forEach((gn, idx) => {
        const col = idx % cols;
        const row = Math.floor(idx / cols);
        positions.set(gn.id, {
          x: col * 320 + 60,
          y: row * 145 + 70,
        });
      });
    }

    // 4. Transform to React Flow Nodes
    const rfNodes: Node[] = filtered.map((gn) => {
      const pos = positions.get(gn.id) || { x: 50, y: 50 };
      const isFocused = focusedNodeId === gn.id;
      const isCaller = callerNodeIds.has(gn.id);
      const isCallee = calleeNodeIds.has(gn.id);
      const hasFocus = !!focusedNodeId;
      const isDimmed = hasFocus && !isFocused && !isCaller && !isCallee;

      return {
        id: gn.id,
        type: 'customNode',
        position: pos,
        data: {
          node: gn,
          isFocused,
          isCaller,
          isCallee,
          isDimmed,
          onSelect: (selected: GraphNode) => {
            if (focusedNodeId === selected.id) {
              setFocusedNodeId(null);
              setIsolatedSubgraph(false);
            } else {
              setFocusedNodeId(selected.id);
            }
            setSelectedNode(selected);
          },
          onOpenFile: handleOpenFile,
          onIsolate: () => {
            setIsolatedSubgraph(true);
          },
        },
      };
    });

    // 5. Transform to React Flow Edges with Electron Circuit Traces
    const rfEdges: Edge[] = finalEdges.map((ge) => {
      const isConnectedToFocus = focusedNodeId === ge.source || focusedNodeId === ge.target;
      const isIncomingToFocus = focusedNodeId === ge.target;
      const isOutgoingFromFocus = focusedNodeId === ge.source;
      const hasFocus = !!focusedNodeId;
      const isDimmed = hasFocus && !isConnectedToFocus;

      return {
        id: ge.id,
        source: ge.source,
        target: ge.target,
        type: 'electronCircuit',
        data: {
          isFocused: isConnectedToFocus,
          isIncoming: isIncomingToFocus,
          isOutgoing: isOutgoingFromFocus,
          isDimmed,
          kind: ge.kind,
          wireRouting,
          enableElectrons,
          label: hasFocus && isConnectedToFocus ? (isIncomingToFocus ? 'calls ➔' : '➔ invokes') : undefined,
        },
      };
    });

    setNodes(rfNodes);
    setEdges(rfEdges);
  }, [
    graphData, 
    search, 
    selectedType, 
    selectedLanguage,
    selectedPackage, 
    layoutMode, 
    densityMode,
    wireRouting,
    enableElectrons,
    connectedOnly, 
    hideTests,
    focusedNodeId, 
    isolatedSubgraph,
    callerNodeIds, 
    calleeNodeIds,
    setSelectedNode
  ]);

  const onNodeClick = useCallback((_: any, node: Node) => {
    if (node.data?.node) {
      const gn = node.data.node as GraphNode;
      setFocusedNodeId(gn.id === focusedNodeId ? null : gn.id);
      setSelectedNode(gn);
    }
  }, [focusedNodeId, setSelectedNode]);

  // Extract caller and callee nodes for Inspector drawer
  const callerNodes = useMemo(() => {
    return (graphData.nodes || []).filter((n) => callerNodeIds.has(n.id));
  }, [graphData.nodes, callerNodeIds]);

  const calleeNodes = useMemo(() => {
    return (graphData.nodes || []).filter((n) => calleeNodeIds.has(n.id));
  }, [graphData.nodes, calleeNodeIds]);

  return (
    <div className="h-[calc(100vh-4rem)] flex flex-col bg-charcoal-950 relative overflow-hidden select-none">
      {/* Hidden SVG Definitions for Neon & Circuit Arrowheads */}
      <svg className="absolute w-0 h-0 pointer-events-none" aria-hidden="true">
        <defs>
          <marker
            id="circuit-marker-emerald"
            viewBox="0 0 10 10"
            refX="8"
            refY="5"
            markerWidth="7"
            markerHeight="7"
            orient="auto-start-reverse"
          >
            <path d="M 0 1.5 L 8 5 L 0 8.5 z" fill="#10b981" />
          </marker>
          <marker
            id="circuit-marker-cyan"
            viewBox="0 0 10 10"
            refX="8"
            refY="5"
            markerWidth="7"
            markerHeight="7"
            orient="auto-start-reverse"
          >
            <path d="M 0 1.5 L 8 5 L 0 8.5 z" fill="#06b6d4" />
          </marker>
          <marker
            id="circuit-marker-purple"
            viewBox="0 0 10 10"
            refX="8"
            refY="5"
            markerWidth="7"
            markerHeight="7"
            orient="auto-start-reverse"
          >
            <path d="M 0 1.5 L 8 5 L 0 8.5 z" fill="#a855f7" />
          </marker>
          <marker
            id="circuit-marker-muted"
            viewBox="0 0 10 10"
            refX="8"
            refY="5"
            markerWidth="6"
            markerHeight="6"
            orient="auto-start-reverse"
          >
            <path d="M 0 1.5 L 8 5 L 0 8.5 z" fill="#64748b" />
          </marker>
          <marker
            id="circuit-marker-white"
            viewBox="0 0 10 10"
            refX="8"
            refY="5"
            markerWidth="7"
            markerHeight="7"
            orient="auto-start-reverse"
          >
            <path d="M 0 1.5 L 8 5 L 0 8.5 z" fill="#ffffff" />
          </marker>

          <marker
            id="neon-marker-emerald"
            viewBox="0 0 10 10"
            refX="8"
            refY="5"
            markerWidth="8"
            markerHeight="8"
            orient="auto-start-reverse"
          >
            <path d="M 0 1.5 L 8 5 L 0 8.5 z" fill="#34d399" />
          </marker>
          <marker
            id="neon-marker-cyan"
            viewBox="0 0 10 10"
            refX="8"
            refY="5"
            markerWidth="8"
            markerHeight="8"
            orient="auto-start-reverse"
          >
            <path d="M 0 1.5 L 8 5 L 0 8.5 z" fill="#38bdf8" />
          </marker>
          <marker
            id="neon-marker-purple"
            viewBox="0 0 10 10"
            refX="8"
            refY="5"
            markerWidth="8"
            markerHeight="8"
            orient="auto-start-reverse"
          >
            <path d="M 0 1.5 L 8 5 L 0 8.5 z" fill="#c084fc" />
          </marker>
          <marker
            id="neon-marker-white"
            viewBox="0 0 10 10"
            refX="8"
            refY="5"
            markerWidth="8"
            markerHeight="8"
            orient="auto-start-reverse"
          >
            <path d="M 0 1.5 L 8 5 L 0 8.5 z" fill="#ffffff" />
          </marker>
        </defs>
      </svg>

      {/* Top Controls Toolbar */}
      <div className="h-14 border-b border-charcoal-750 bg-charcoal-900/95 backdrop-blur px-6 flex items-center justify-between gap-4 z-10 shrink-0 flex-wrap">
        <div className="flex items-center gap-2.5 flex-wrap">
          {/* Search bar */}
          <div className="relative w-56">
            <Search className="w-3.5 h-3.5 text-zinc-500 absolute left-3 top-1/2 -translate-y-1/2" />
            <input
              type="text"
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              placeholder="Search functions &amp; files..."
              className="w-full bg-charcoal-850 border border-charcoal-750 rounded-lg pl-8 pr-7 py-1.5 text-xs text-white placeholder-zinc-500 focus:outline-none focus:border-dna-blue/50 font-mono"
            />
            {search && (
              <button
                onClick={() => setSearch('')}
                className="absolute right-2.5 top-1/2 -translate-y-1/2 text-zinc-500 hover:text-zinc-300"
              >
                <X className="w-3 h-3" />
              </button>
            )}
          </div>

          {/* Node Type Selector */}
          <div className="flex items-center bg-charcoal-850 border border-charcoal-750 rounded-lg p-0.5 text-xs font-mono">
            {[
              { id: 'function', label: 'Functions' },
              { id: 'file', label: 'Files' },
              { id: 'class', label: 'Classes' },
              { id: 'all', label: 'All' },
            ].map((t) => (
              <button
                key={t.id}
                onClick={() => {
                  setSelectedType(t.id);
                  setFocusedNodeId(null);
                  setIsolatedSubgraph(false);
                }}
                className={`px-3 py-1 rounded-md transition-all ${
                  selectedType === t.id ? 'bg-dna-blue text-white font-medium shadow-sm' : 'text-zinc-400 hover:text-zinc-200'
                }`}
              >
                {t.label}
              </button>
            ))}
          </div>

          {/* Language Selector */}
          <div className="flex items-center bg-charcoal-850 border border-charcoal-750 rounded-lg p-0.5 text-xs font-mono">
            {[
              { id: 'all', label: 'All Lang' },
              { id: 'go', label: 'Go' },
              { id: 'typescript', label: 'TS/JS' },
            ].map((l) => (
              <button
                key={l.id}
                onClick={() => setSelectedLanguage(l.id as any)}
                className={`px-2.5 py-1 rounded-md transition-all ${
                  selectedLanguage === l.id ? 'bg-charcoal-700 text-white font-semibold' : 'text-zinc-400 hover:text-zinc-200'
                }`}
              >
                {l.label}
              </button>
            ))}
          </div>

          {/* Package Filter Dropdown */}
          {packageList.length > 0 && (
            <div className="flex items-center gap-1.5 bg-charcoal-850 border border-charcoal-750 rounded-lg px-2.5 py-1 text-xs font-mono text-zinc-300">
              <Filter className="w-3 h-3 text-dna-blue shrink-0" />
              <select
                value={selectedPackage}
                onChange={(e) => {
                  setSelectedPackage(e.target.value);
                  setFocusedNodeId(null);
                  setIsolatedSubgraph(false);
                }}
                className="bg-transparent text-xs text-zinc-200 focus:outline-none cursor-pointer max-w-[130px] truncate"
              >
                <option value="all" className="bg-charcoal-900 text-white">All Packages</option>
                {packageList.map((pkg) => (
                  <option key={pkg} value={pkg} className="bg-charcoal-900 text-white">
                    {pkg}
                  </option>
                ))}
              </select>
            </div>
          )}

          {/* Hide Tests Toggle */}
          <button
            onClick={() => setHideTests(!hideTests)}
            className={`flex items-center gap-1.5 px-2.5 py-1 rounded-lg border text-xs font-mono transition-all ${
              hideTests
                ? 'bg-amber-500/15 text-amber-300 border-amber-500/30 font-semibold'
                : 'bg-charcoal-850 text-zinc-400 border-charcoal-750 hover:text-zinc-200'
            }`}
            title="Exclude test functions from the architectural call graph"
          >
            <ShieldCheck className="w-3.5 h-3.5" />
            <span>{hideTests ? 'Prod Only' : 'With Tests'}</span>
          </button>

          {/* Connected Only Filter Toggle */}
          <button
            onClick={() => setConnectedOnly(!connectedOnly)}
            className={`flex items-center gap-1.5 px-2.5 py-1 rounded-lg border text-xs font-mono transition-all ${
              connectedOnly
                ? 'bg-emerald-500/15 text-emerald-300 border-emerald-500/30 font-semibold'
                : 'bg-charcoal-850 text-zinc-400 border-charcoal-750 hover:text-zinc-200'
            }`}
            title="Filter out orphan helper symbols with 0 connections"
          >
            <Activity className="w-3.5 h-3.5" />
            <span>Connected</span>
          </button>

          {/* Density Mode: Core Flow vs All Symbols */}
          <div className="flex items-center bg-charcoal-850 border border-charcoal-750 rounded-lg p-0.5 text-xs font-mono">
            <button
              onClick={() => setDensityMode('core')}
              className={`px-2.5 py-1 rounded-md transition-all flex items-center gap-1 ${
                densityMode === 'core'
                  ? 'bg-emerald-500/20 text-emerald-300 border border-emerald-500/40 font-semibold'
                  : 'text-zinc-400 hover:text-zinc-200'
              }`}
              title="Show architectural core flow (Clean & fast, removes spaghetti)"
            >
              <Sparkles className="w-3 h-3 text-emerald-400" />
              <span>Core Flow</span>
            </button>
            <button
              onClick={() => setDensityMode('all')}
              className={`px-2.5 py-1 rounded-md transition-all ${
                densityMode === 'all'
                  ? 'bg-dna-blue text-white font-semibold'
                  : 'text-zinc-400 hover:text-zinc-200'
              }`}
              title="Show all function symbols across the entire codebase"
            >
              <span>All Symbols</span>
            </button>
          </div>
        </div>

        {/* View Layout Controls & Export */}
        <div className="flex items-center gap-3">
          {/* Wire Routing Style */}
          <div className="flex items-center bg-charcoal-850 border border-charcoal-750 rounded-lg p-0.5 text-xs font-mono">
            <button
              onClick={() => setWireRouting('circuit')}
              className={`px-2 py-1 rounded-md transition-all ${
                wireRouting === 'circuit'
                  ? 'bg-charcoal-700 text-white font-semibold'
                  : 'text-zinc-400 hover:text-white'
              }`}
              title="Circuit board 90-degree step wires with dashed traces"
            >
              Circuit Step
            </button>
            <button
              onClick={() => setWireRouting('smooth')}
              className={`px-2 py-1 rounded-md transition-all ${
                wireRouting === 'smooth'
                  ? 'bg-charcoal-700 text-white font-semibold'
                  : 'text-zinc-400 hover:text-white'
              }`}
              title="Smooth bezier curves"
            >
              Smooth Wave
            </button>
          </div>

          {/* Electrons Animation Toggle */}
          <button
            onClick={() => setEnableElectrons(!enableElectrons)}
            className={`flex items-center gap-1 px-2.5 py-1 rounded-lg border text-xs font-mono transition-all ${
              enableElectrons
                ? 'bg-sky-500/15 text-sky-300 border-sky-500/30 font-semibold'
                : 'bg-charcoal-850 text-zinc-500 border-charcoal-750 hover:text-zinc-300'
            }`}
            title="Toggle animated traveling electrons along circuit tracks"
          >
            <Zap className={`w-3.5 h-3.5 ${enableElectrons ? 'text-sky-400' : 'text-zinc-600'}`} />
            <span>Electrons</span>
          </button>

          {/* Layout Mode Switcher */}
          <div className="flex items-center bg-charcoal-850 border border-charcoal-750 rounded-lg p-0.5 text-xs font-mono">
            <button
              onClick={() => setLayoutMode('dag')}
              className={`px-2.5 py-1 rounded-md transition-all ${
                layoutMode === 'dag' ? 'bg-dna-blue text-white shadow-sm font-semibold' : 'text-zinc-400 hover:text-white'
              }`}
              title="Left-to-Right Topological Call Graph (Callers ➔ Callees)"
            >
              Call Flow DAG
            </button>
            <button
              onClick={() => setLayoutMode('domain')}
              className={`px-2.5 py-1 rounded-md transition-all ${
                layoutMode === 'domain' ? 'bg-dna-blue text-white shadow-sm font-semibold' : 'text-zinc-400 hover:text-white'
              }`}
              title="Group columns by architectural domain tier"
            >
              Domain Clusters
            </button>
            <button
              onClick={() => setLayoutMode('grid')}
              className={`px-2.5 py-1 rounded-md transition-all ${
                layoutMode === 'grid' ? 'bg-dna-blue text-white shadow-sm font-semibold' : 'text-zinc-400 hover:text-white'
              }`}
              title="Matrix grid by connectivity"
            >
              Grid
            </button>
          </div>

          <div className="text-xs font-mono text-zinc-400 px-3 py-1 bg-charcoal-850 border border-charcoal-750 rounded-lg hidden sm:block">
            <span className="text-white font-semibold">{nodes.length}</span> nodes • <span className="text-white font-semibold">{edges.length}</span> calls
          </div>

          <button
            onClick={handleOpenExport}
            className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg bg-dna-blue/15 hover:bg-dna-blue/25 text-dna-blue border border-dna-blue/30 text-xs font-mono font-medium transition-all"
            title="Export architecture diagram as Mermaid / C4"
          >
            <Download className="w-3.5 h-3.5" />
            <span>Export</span>
          </button>
        </div>
      </div>

      {/* Focus & Subgraph Ribbon (when a function is selected) */}
      {focusedNode && (
        <div className="bg-[#141b2d] border-b border-dna-blue/30 px-6 py-2 flex items-center justify-between text-xs font-mono z-10 shrink-0">
          <div className="flex items-center gap-2.5 flex-wrap">
            <span className="text-[10px] uppercase font-bold px-2 py-0.5 rounded bg-dna-blue text-white flex items-center gap-1 shadow-sm">
              <Zap className="w-3 h-3" />
              <span>{isolatedSubgraph ? 'Isolated Subgraph Flow' : 'Call Path Active'}</span>
            </span>
            <span className="text-white font-bold">{focusedNode.label}</span>
            <span className="text-zinc-500">•</span>
            <span className="text-emerald-400 font-semibold">{callerNodes.length} Callers (Incoming)</span>
            <span className="text-zinc-500">•</span>
            <span className="text-sky-400 font-semibold">{calleeNodes.length} Calls (Outgoing)</span>
          </div>

          <div className="flex items-center gap-2">
            {!isolatedSubgraph ? (
              <button
                onClick={() => setIsolatedSubgraph(true)}
                className="flex items-center gap-1.5 px-2.5 py-1 rounded bg-dna-blue/20 hover:bg-dna-blue/30 border border-dna-blue/40 text-dna-blue text-[11px] font-semibold transition-all"
                title="Isolate this exact call workflow"
              >
                <Sparkles className="w-3 h-3" />
                <span>Isolate Subgraph</span>
              </button>
            ) : (
              <button
                onClick={() => setIsolatedSubgraph(false)}
                className="flex items-center gap-1.5 px-2.5 py-1 rounded bg-charcoal-800 hover:bg-charcoal-700 text-zinc-300 text-[11px] font-semibold transition-all"
              >
                <span>Show Full Graph</span>
              </button>
            )}

            <button
              onClick={() => {
                setFocusedNodeId(null);
                setIsolatedSubgraph(false);
              }}
              className="flex items-center gap-1 text-zinc-400 hover:text-white px-2 py-1 rounded hover:bg-charcoal-800 transition-colors"
            >
              <span>Reset</span>
              <X className="w-3.5 h-3.5" />
            </button>
          </div>
        </div>
      )}

      {/* React Flow Graph Canvas */}
      <div className="flex-1 w-full h-full relative">
        {loading ? (
          <div className="h-full flex flex-col items-center justify-center text-zinc-400 font-mono text-xs gap-3">
            <div className="w-6 h-6 border-2 border-dna-blue border-t-transparent rounded-full animate-spin" />
            <span>Resolving AST Call Graph &amp; Glowing Circuit Lines...</span>
          </div>
        ) : nodes.length === 0 ? (
          <div className="h-full flex flex-col items-center justify-center text-zinc-400 font-mono text-xs gap-2">
            <Activity className="w-8 h-8 text-zinc-600" />
            <p>No symbols match the current filters ({selectedType}, {selectedPackage}).</p>
            {connectedOnly && (
              <button
                onClick={() => setConnectedOnly(false)}
                className="mt-2 text-dna-blue underline"
              >
                Disable "Connected Only" to show orphan symbols
              </button>
            )}
          </div>
        ) : (
          <ReactFlowProvider>
            <ReactFlow
              nodes={nodes}
              edges={edges}
              nodeTypes={nodeTypes}
              edgeTypes={edgeTypes}
              onNodesChange={onNodesChange}
              onEdgesChange={onEdgesChange}
              onNodeClick={onNodeClick}
              fitView
              fitViewOptions={{ padding: 0.15 }}
              minZoom={0.06}
              maxZoom={2.8}
              zoomOnScroll={true}
              zoomOnPinch={true}
              panOnScroll={false}
              panOnDrag={true}
              preventScrolling={true}
              nodeDragThreshold={1}
              onlyRenderVisibleElements={true}
              proOptions={{ hideAttribution: true }}
            >
              <Background color="#161820" gap={20} size={1} />
              <Controls className="!bg-[#14171e] !border-charcoal-700 !fill-zinc-300 !rounded-xl" />
              <MiniMap 
                nodeColor={(n) => {
                  if (n.id === focusedNodeId) return '#3b82f6';
                  if (callerNodeIds.has(n.id)) return '#34d399';
                  if (calleeNodeIds.has(n.id)) return '#38bdf8';
                  return '#52525b';
                }}
                maskColor="rgba(10, 11, 14, 0.75)"
                className="!bg-[#12141c] !border-charcoal-700 !rounded-xl"
              />
            </ReactFlow>
          </ReactFlowProvider>
        )}

        {/* Dynamic Circuit Track & Electron Legend Toolbar */}
        <div className="absolute bottom-4 left-4 z-10 pointer-events-auto hidden md:flex items-center gap-3 bg-[#14171e]/90 backdrop-blur-md px-3.5 py-2 rounded-xl border border-charcoal-700 text-[11px] font-mono text-zinc-300 shadow-xl">
          <span className="text-zinc-500 font-bold flex items-center gap-1">
            <Zap className="w-3.5 h-3.5 text-amber-400" />
            <span>Circuit Flow:</span>
          </span>
          <span className="flex items-center gap-1.5">
            <span className="w-3 h-0.5 border-b-2 border-dashed border-emerald-400" />
            <span className="w-2 h-2 rounded-full bg-emerald-400" />
            <span>Caller Electron (In)</span>
          </span>
          <span className="flex items-center gap-1.5">
            <span className="w-3 h-0.5 border-b-2 border-dashed border-sky-400" />
            <span className="w-2 h-2 rounded-full bg-sky-400" />
            <span>Callee Electron (Out)</span>
          </span>
          <span className="flex items-center gap-1.5">
            <span className="w-3 h-0.5 border-b-2 border-dashed border-purple-400" />
            <span>Import</span>
          </span>
          <span className="text-zinc-600">|</span>
          <span className="text-zinc-400">Click any symbol to isolate path</span>
        </div>
      </div>

      {/* Function Call Hierarchy Inspector Drawer (When focused) */}
      {focusedNode && (
        <div className="absolute top-14 right-0 bottom-0 w-80 md:w-96 bg-[#10131a] border-l border-charcoal-750 shadow-2xl flex flex-col z-20 animate-in slide-in-from-right duration-200">
          {/* Header */}
          <div className="p-4 border-b border-charcoal-800 bg-[#141822] flex items-start justify-between gap-3">
            <div>
              <div className="flex items-center gap-2">
                <span className="text-[10px] font-mono uppercase px-1.5 py-0.5 rounded font-bold bg-dna-blue/20 text-dna-blue border border-dna-blue/40">
                  {focusedNode.type}
                </span>
                <span className="text-xs font-mono text-zinc-400">Call Hierarchy</span>
              </div>
              <h3 className="text-sm font-bold font-mono text-white mt-1.5 truncate max-w-[260px]">
                {focusedNode.label}
              </h3>
              {focusedNode.file && (
                <p className="text-[10px] font-mono text-zinc-400 mt-0.5 truncate">
                  {focusedNode.file}:{focusedNode.line || 1}
                </p>
              )}
            </div>

            <button
              onClick={() => {
                setFocusedNodeId(null);
                setIsolatedSubgraph(false);
              }}
              className="p-1 rounded text-zinc-400 hover:text-white hover:bg-charcoal-750"
              title="Close Inspector"
            >
              <X className="w-4 h-4" />
            </button>
          </div>

          {/* Quick Metrics Bar */}
          <div className="grid grid-cols-3 gap-2 p-3 border-b border-charcoal-800 bg-charcoal-950 font-mono text-center">
            <div className="bg-charcoal-900 p-2 rounded border border-charcoal-800">
              <div className="text-[10px] text-zinc-500 uppercase">Callers</div>
              <div className="text-xs font-bold text-emerald-400">{callerNodes.length}</div>
            </div>
            <div className="bg-charcoal-900 p-2 rounded border border-charcoal-800">
              <div className="text-[10px] text-zinc-500 uppercase">Callees</div>
              <div className="text-xs font-bold text-sky-400">{calleeNodes.length}</div>
            </div>
            <div className="bg-charcoal-900 p-2 rounded border border-charcoal-800">
              <div className="text-[10px] text-zinc-500 uppercase">Complexity</div>
              <div className={`text-xs font-bold ${
                (focusedNode.complexity || 0) > 15 ? 'text-rose-400' : 'text-white'
              }`}>
                {focusedNode.complexity || 0}
              </div>
            </div>
          </div>

          {/* Action Buttons: Isolate Subgraph & Code Editor */}
          <div className="p-3 border-b border-charcoal-800 bg-charcoal-950 space-y-2">
            {!isolatedSubgraph ? (
              <button
                onClick={() => setIsolatedSubgraph(true)}
                className="w-full flex items-center justify-center gap-2 py-2 rounded-lg bg-dna-blue/20 hover:bg-dna-blue/30 text-dna-blue border border-dna-blue/40 text-xs font-mono font-semibold transition-all shadow-md"
              >
                <Sparkles className="w-4 h-4" />
                <span>Isolate Call Subgraph (Clean Flow)</span>
              </button>
            ) : (
              <button
                onClick={() => setIsolatedSubgraph(false)}
                className="w-full flex items-center justify-center gap-2 py-2 rounded-lg bg-charcoal-800 hover:bg-charcoal-750 text-zinc-300 text-xs font-mono font-semibold transition-all"
              >
                <span>Restore Full Architecture Graph</span>
              </button>
            )}

            {focusedNode.file && (
              <button
                onClick={() => handleOpenFile(focusedNode.file, focusedNode.line)}
                className="w-full flex items-center justify-center gap-2 py-2 rounded-lg bg-charcoal-850 hover:bg-charcoal-800 text-xs font-mono font-semibold text-white border border-charcoal-700 transition-all shadow-md"
              >
                <FileCode className="w-4 h-4 text-dna-blue" />
                <span>Inspect in Code Editor (Line {focusedNode.line || 1})</span>
              </button>
            )}
          </div>

          {/* Callers & Callees Lists */}
          <div className="flex-1 overflow-y-auto p-4 space-y-5">
            {/* Direct Callers */}
            <div>
              <div className="flex items-center justify-between text-xs font-mono font-bold text-emerald-400 mb-2">
                <span className="flex items-center gap-1.5">
                  <span className="w-2 h-2 rounded-full bg-emerald-400 shadow-[0_0_6px_#34d399]" />
                  <span>Callers (Who invokes this):</span>
                </span>
                <span className="text-[10px] text-zinc-500 font-normal">{callerNodes.length} functions</span>
              </div>

              {callerNodes.length > 0 ? (
                <div className="space-y-1.5">
                  {callerNodes.map((cn) => (
                    <div
                      key={cn.id}
                      onClick={() => {
                        setFocusedNodeId(cn.id);
                        setSelectedNode(cn);
                      }}
                      className="p-2.5 rounded-lg bg-charcoal-900 hover:bg-charcoal-850 border border-charcoal-800 hover:border-emerald-500/40 cursor-pointer transition-all group font-mono text-xs"
                    >
                      <div className="flex items-center justify-between">
                        <span className="font-bold text-white group-hover:text-emerald-300 truncate max-w-[200px]">
                          {cn.label}
                        </span>
                        <ChevronRight className="w-3.5 h-3.5 text-zinc-500 group-hover:text-emerald-400" />
                      </div>
                      <div className="text-[10px] text-zinc-400 truncate mt-0.5">
                        {cn.file?.split('/').pop()}:{cn.line || 1}
                      </div>
                    </div>
                  ))}
                </div>
              ) : (
                <div className="p-3 rounded-lg bg-charcoal-900/50 border border-charcoal-800 text-center text-zinc-500 text-xs font-mono">
                  No direct callers in current view. (Entrypoint / Public API)
                </div>
              )}
            </div>

            {/* Direct Callees */}
            <div>
              <div className="flex items-center justify-between text-xs font-mono font-bold text-sky-400 mb-2">
                <span className="flex items-center gap-1.5">
                  <span className="w-2 h-2 rounded-full bg-sky-400 shadow-[0_0_6px_#38bdf8]" />
                  <span>Callees (What this invokes):</span>
                </span>
                <span className="text-[10px] text-zinc-500 font-normal">{calleeNodes.length} functions</span>
              </div>

              {calleeNodes.length > 0 ? (
                <div className="space-y-1.5">
                  {calleeNodes.map((cn) => (
                    <div
                      key={cn.id}
                      onClick={() => {
                        setFocusedNodeId(cn.id);
                        setSelectedNode(cn);
                      }}
                      className="p-2.5 rounded-lg bg-charcoal-900 hover:bg-charcoal-850 border border-charcoal-800 hover:border-sky-500/40 cursor-pointer transition-all group font-mono text-xs"
                    >
                      <div className="flex items-center justify-between">
                        <span className="font-bold text-white group-hover:text-sky-300 truncate max-w-[200px]">
                          {cn.label}
                        </span>
                        <ChevronRight className="w-3.5 h-3.5 text-zinc-500 group-hover:text-sky-400" />
                      </div>
                      <div className="text-[10px] text-zinc-400 truncate mt-0.5">
                        {cn.file?.split('/').pop()}:{cn.line || 1}
                      </div>
                    </div>
                  ))}
                </div>
              ) : (
                <div className="p-3 rounded-lg bg-charcoal-900/50 border border-charcoal-800 text-center text-zinc-500 text-xs font-mono">
                  No outgoing calls detected. (Leaf / Terminal Function)
                </div>
              )}
            </div>
          </div>
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
                <span className="text-zinc-500">• Line {inspectedFile.targetLine || 1}</span>
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
                targetLine={inspectedFile.targetLine || 1}
                maxHeight="h-full"
              />
            </div>
          </div>
        </div>
      )}

      {/* Export Diagram Modal */}
      {showExportModal && (
        <div className="fixed inset-0 bg-black/75 backdrop-blur-sm z-50 flex items-center justify-center p-4 animate-in fade-in duration-200">
          <div className="bg-charcoal-900 border border-charcoal-750 rounded-xl w-full max-w-2xl shadow-2xl flex flex-col max-h-[85vh] overflow-hidden">
            <div className="h-14 border-b border-charcoal-750 px-6 flex items-center justify-between shrink-0 bg-charcoal-850">
              <div className="flex items-center gap-2.5 text-white font-mono font-bold text-sm">
                <Share2 className="w-4 h-4 text-dna-blue" />
                <span>Export Architecture Diagram</span>
              </div>
              <button
                onClick={() => setShowExportModal(false)}
                className="text-zinc-400 hover:text-white transition-colors"
              >
                <X className="w-4 h-4" />
              </button>
            </div>

            <div className="p-4 border-b border-charcoal-750 bg-charcoal-950 flex items-center justify-between gap-4">
              <div className="flex items-center bg-charcoal-850 border border-charcoal-750 rounded-lg p-0.5 text-xs font-mono">
                <button
                  onClick={() => handleFormatChange('mermaid')}
                  className={`px-3 py-1 rounded-md transition-all ${
                    exportFormat === 'mermaid' ? 'bg-dna-blue text-white font-medium' : 'text-zinc-400 hover:text-white'
                  }`}
                >
                  Mermaid.js Flowchart
                </button>
                <button
                  onClick={() => handleFormatChange('c4')}
                  className={`px-3 py-1 rounded-md transition-all ${
                    exportFormat === 'c4' ? 'bg-dna-blue text-white font-medium' : 'text-zinc-400 hover:text-white'
                  }`}
                >
                  C4 Component Diagram
                </button>
              </div>

              <div className="flex items-center gap-2">
                <button
                  onClick={handleCopyExport}
                  className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg bg-charcoal-800 hover:bg-charcoal-750 border border-charcoal-700 text-xs font-mono text-zinc-200 hover:text-white transition-all"
                >
                  {copiedExport ? <Check className="w-3.5 h-3.5 text-emerald-400" /> : <Copy className="w-3.5 h-3.5" />}
                  <span>{copiedExport ? 'Copied!' : 'Copy Code'}</span>
                </button>
                <button
                  onClick={handleDownloadExport}
                  className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg bg-dna-blue text-white hover:bg-dna-blue/90 text-xs font-mono font-medium transition-all"
                >
                  <Download className="w-3.5 h-3.5" />
                  <span>Download file</span>
                </button>
              </div>
            </div>

            <div className="flex-1 overflow-auto p-4 bg-charcoal-950 font-mono text-xs select-text">
              {exportLoading ? (
                <div className="h-64 flex items-center justify-center text-zinc-500 gap-2">
                  <div className="w-3.5 h-3.5 border-2 border-dna-blue border-t-transparent rounded-full animate-spin" />
                  <span>Generating diagram specification...</span>
                </div>
              ) : (
                <pre className="text-zinc-300 whitespace-pre leading-relaxed">
                  {exportCode}
                </pre>
              )}
            </div>

            <div className="p-3 border-t border-charcoal-750 bg-charcoal-850 px-6 flex items-center justify-between text-[11px] font-mono text-zinc-400">
              <span>Ready for GitHub README, Notion, Obsidian &amp; Architecture Docs</span>
              <span>Format: {exportFormat === 'mermaid' ? 'Mermaid Flowchart' : 'C4 Container'}</span>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
