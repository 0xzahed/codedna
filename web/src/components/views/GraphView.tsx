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
  ExternalLink
} from 'lucide-react';

// Custom Node Component for Functions, Methods, Classes & Files
function CustomGraphNode({ data }: { data: any }) {
  const gn: GraphNode = data.node;
  const isFocused = data.isFocused;
  const isCaller = data.isCaller;
  const isCallee = data.isCallee;
  const isDimmed = data.isDimmed;
  const onSelect = data.onSelect;
  const onOpenFile = data.onOpenFile;

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
    ringStyle = 'border-dna-blue ring-2 ring-dna-blue bg-[#141b2d] shadow-2xl shadow-dna-blue/30';
  } else if (isCaller) {
    ringStyle = 'border-emerald-400 ring-2 ring-emerald-400/80 bg-[#102419] shadow-xl shadow-emerald-400/20';
  } else if (isCallee) {
    ringStyle = 'border-sky-400 ring-2 ring-sky-400/80 bg-[#0e1f2b] shadow-xl shadow-sky-400/20';
  }

  // Extract clean package/dir name
  const pkgName = gn.file 
    ? (gn.file.includes('/') ? gn.file.split('/').slice(0, 2).join('/') : gn.file)
    : 'root';

  return (
    <div
      onClick={() => onSelect && onSelect(gn)}
      className={`w-72 rounded-xl p-3.5 transition-all duration-200 cursor-pointer shadow-lg select-none ${ringStyle} ${
        isDimmed ? 'opacity-25 hover:opacity-100' : 'opacity-100'
      }`}
    >
      {/* React Flow Connection Handles on all 4 sides for clean DAG routing */}
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
          <span className="text-[9px] font-mono font-bold px-1.5 py-0.5 rounded bg-dna-blue text-white shadow-sm">
            ★ FOCUSED
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
          <span className="flex items-center text-[10px] font-mono text-rose-400 font-bold" title={`Cyclomatic Complexity: ${complexity}`}>
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
  );
}

const nodeTypes = {
  customNode: CustomGraphNode,
};

// Topological Level DAG Calculator for Call Graphs
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
      // Cap at 10 depth levels to prevent infinite loops from cyclic calls
      if (nextLevel < 10) {
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

export default function GraphView() {
  const { scanId, setSelectedNode, workspaceId, setWorkspaceId, refreshKey } = useAppStore();
  const [graphData, setGraphData] = useState<{ nodes: GraphNode[]; edges: GraphEdge[] }>({ nodes: [], edges: [] });
  const [nodes, setNodes, onNodesChange] = useNodesState([]);
  const [edges, setEdges, onEdgesChange] = useEdgesState([]);
  const [search, setSearch] = useState('');
  const [selectedType, setSelectedType] = useState<string>('function');
  const [selectedPackage, setSelectedPackage] = useState<string>('all');
  const [layoutMode, setLayoutMode] = useState<'dag' | 'swimlane' | 'grid'>('dag');
  const [connectedOnly, setConnectedOnly] = useState(true);
  const [focusedNodeId, setFocusedNodeId] = useState<string | null>(null);
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

  // Compute Layout & React Flow nodes/edges
  useEffect(() => {
    if (!graphData.nodes || graphData.nodes.length === 0) {
      setNodes([]);
      setEdges([]);
      return;
    }

    // 1. Initial filter
    let filtered = graphData.nodes;

    if (selectedType !== 'all') {
      if (selectedType === 'function') {
        filtered = filtered.filter((n) => n.type === 'function' || n.type === 'method');
      } else {
        filtered = filtered.filter((n) => n.type === selectedType);
      }
    }

    if (selectedPackage !== 'all') {
      filtered = filtered.filter((n) => n.file && n.file.startsWith(selectedPackage));
    }

    if (search.trim()) {
      const q = search.toLowerCase();
      filtered = filtered.filter((n) => 
        n.label.toLowerCase().includes(q) || (n.file && n.file.toLowerCase().includes(q))
      );
    }

    const filteredIds = new Set(filtered.map((n) => n.id));

    // 2. Filter edges strictly within current nodes
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
    if (connectedOnly && (selectedType === 'function' || selectedType === 'all')) {
      filtered = filtered.filter((n) => (activeDegree.get(n.id) || 0) > 0 || n.id === focusedNodeId);
    }

    const finalNodeIds = new Set(filtered.map((n) => n.id));
    const finalEdges = validEdges.filter(
      (e) => finalNodeIds.has(e.source) && finalNodeIds.has(e.target)
    );

    // 3. Compute Positions based on selected Layout Mode
    const positions = new Map<string, { x: number; y: number }>();

    if (layoutMode === 'dag') {
      // Topological Level DAG layout (Left to Right call progression)
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
          const pkgA = a.file?.split('/')[0] || '';
          const pkgB = b.file?.split('/')[0] || '';
          if (pkgA !== pkgB) return pkgA.localeCompare(pkgB);
          return (b.complexity || 0) - (a.complexity || 0);
        });
      });

      levelBuckets.forEach((bucket, lvl) => {
        bucket.forEach((gn, row) => {
          positions.set(gn.id, {
            x: lvl * 360 + 60,
            y: row * 140 + 60,
          });
        });
      });
    } else if (layoutMode === 'swimlane') {
      // Package Swimlanes layout
      const pkgMap = new Map<string, GraphNode[]>();
      filtered.forEach((gn) => {
        const pkg = gn.file ? (gn.file.includes('/') ? gn.file.split('/').slice(0, 2).join('/') : 'root') : 'root';
        if (!pkgMap.has(pkg)) pkgMap.set(pkg, []);
        pkgMap.get(pkg)!.push(gn);
      });

      const sortedPkgs = Array.from(pkgMap.keys()).sort();
      sortedPkgs.forEach((pkg, col) => {
        const list = pkgMap.get(pkg) || [];
        list.sort((a, b) => (b.complexity || 0) - (a.complexity || 0));
        list.forEach((gn, row) => {
          positions.set(gn.id, {
            x: col * 340 + 60,
            y: row * 140 + 80,
          });
        });
      });
    } else {
      // Compact Grid layout
      const cols = Math.ceil(Math.sqrt(filtered.length)) || 4;
      filtered.forEach((gn, idx) => {
        const col = idx % cols;
        const row = Math.floor(idx / cols);
        positions.set(gn.id, {
          x: col * 320 + 60,
          y: row * 140 + 60,
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
            setFocusedNodeId(selected.id === focusedNodeId ? null : selected.id);
            setSelectedNode(selected);
          },
          onOpenFile: handleOpenFile,
        },
      };
    });

    // 5. Transform to React Flow Edges
    const rfEdges: Edge[] = finalEdges.map((ge) => {
      const isConnectedToFocus = focusedNodeId === ge.source || focusedNodeId === ge.target;
      const isIncomingToFocus = focusedNodeId === ge.target;
      const isOutgoingFromFocus = focusedNodeId === ge.source;
      const hasFocus = !!focusedNodeId;

      let strokeColor = '#34d399'; // default emerald for call
      if (ge.kind === 'imports') strokeColor = '#a78bfa'; // violet for import
      if (ge.kind === 'cross_service_api') strokeColor = '#38bdf8'; // sky blue for api

      if (hasFocus) {
        if (isIncomingToFocus) strokeColor = '#34d399'; // bright emerald for incoming
        else if (isOutgoingFromFocus) strokeColor = '#38bdf8'; // bright sky blue for outgoing
        else strokeColor = '#272a33'; // dimmed
      }

      const opacity = hasFocus ? (isConnectedToFocus ? 1 : 0.12) : 0.85;
      const strokeWidth = isConnectedToFocus ? 3 : 1.6;

      return {
        id: ge.id,
        source: ge.source,
        target: ge.target,
        type: 'smoothstep',
        animated: isConnectedToFocus || ge.kind === 'cross_service_api',
        label: hasFocus && isConnectedToFocus ? (isIncomingToFocus ? 'calls ➔' : '➔ invokes') : undefined,
        labelStyle: { fill: strokeColor, fontSize: 10, fontFamily: 'monospace', fontWeight: 600 },
        labelBgStyle: { fill: '#141518', fillOpacity: 0.95 },
        style: {
          stroke: strokeColor,
          strokeWidth,
          opacity,
          strokeDasharray: ge.kind === 'imports' ? '4,4' : undefined,
        },
        markerEnd: {
          type: MarkerType.ArrowClosed,
          color: strokeColor,
          width: isConnectedToFocus ? 14 : 10,
          height: isConnectedToFocus ? 14 : 10,
        },
      };
    });

    setNodes(rfNodes);
    setEdges(rfEdges);
  }, [
    graphData, 
    search, 
    selectedType, 
    selectedPackage, 
    layoutMode, 
    connectedOnly, 
    focusedNodeId, 
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

  // Extract caller and callee nodes for the Inspector drawer
  const callerNodes = useMemo(() => {
    return (graphData.nodes || []).filter((n) => callerNodeIds.has(n.id));
  }, [graphData.nodes, callerNodeIds]);

  const calleeNodes = useMemo(() => {
    return (graphData.nodes || []).filter((n) => calleeNodeIds.has(n.id));
  }, [graphData.nodes, calleeNodeIds]);

  return (
    <div className="h-[calc(100vh-4rem)] flex flex-col bg-charcoal-950 relative overflow-hidden select-none">
      {/* Top Controls Toolbar */}
      <div className="h-14 border-b border-charcoal-750 bg-charcoal-900/95 backdrop-blur px-6 flex items-center justify-between gap-4 z-10 shrink-0 flex-wrap">
        <div className="flex items-center gap-3 flex-wrap">
          {/* Search bar */}
          <div className="relative w-60">
            <Search className="w-3.5 h-3.5 text-zinc-500 absolute left-3 top-1/2 -translate-y-1/2" />
            <input
              type="text"
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              placeholder="Search functions & files..."
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
                }}
                className={`px-3 py-1 rounded-md transition-all ${
                  selectedType === t.id ? 'bg-dna-blue text-white font-medium shadow-sm' : 'text-zinc-400 hover:text-zinc-200'
                }`}
              >
                {t.label}
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
                }}
                className="bg-transparent text-xs text-zinc-200 focus:outline-none cursor-pointer max-w-[140px] truncate"
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
            <span>Connected Only</span>
          </button>
        </div>

        {/* View Layout Controls & Export */}
        <div className="flex items-center gap-3">
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
              onClick={() => setLayoutMode('swimlane')}
              className={`px-2.5 py-1 rounded-md transition-all ${
                layoutMode === 'swimlane' ? 'bg-dna-blue text-white shadow-sm font-semibold' : 'text-zinc-400 hover:text-white'
              }`}
              title="Group columns by package/module"
            >
              Swimlanes
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

      {/* Focus Indicator Ribbon (when a function is selected) */}
      {focusedNode && (
        <div className="bg-[#141b2d] border-b border-dna-blue/30 px-6 py-2 flex items-center justify-between text-xs font-mono z-10 shrink-0">
          <div className="flex items-center gap-2.5 flex-wrap">
            <span className="text-[10px] uppercase font-bold px-1.5 py-0.5 rounded bg-dna-blue text-white">
              Call Path Focus
            </span>
            <span className="text-white font-bold">{focusedNode.label}</span>
            <span className="text-zinc-500">•</span>
            <span className="text-emerald-400 font-semibold">{callerNodes.length} Callers (Incoming)</span>
            <span className="text-zinc-500">•</span>
            <span className="text-sky-400 font-semibold">{calleeNodes.length} Calls (Outgoing)</span>
          </div>

          <button
            onClick={() => setFocusedNodeId(null)}
            className="flex items-center gap-1 text-zinc-400 hover:text-white px-2 py-0.5 rounded hover:bg-charcoal-800 transition-colors"
          >
            <span>Reset View</span>
            <X className="w-3.5 h-3.5" />
          </button>
        </div>
      )}

      {/* React Flow Graph Canvas */}
      <div className="flex-1 w-full h-full relative">
        {loading ? (
          <div className="h-full flex flex-col items-center justify-center text-zinc-400 font-mono text-xs gap-3">
            <div className="w-6 h-6 border-2 border-dna-blue border-t-transparent rounded-full animate-spin" />
            <span>Resolving AST Call Graph &amp; Functions...</span>
          </div>
        ) : nodes.length === 0 ? (
          <div className="h-full flex flex-col items-center justify-center text-zinc-400 font-mono text-xs gap-2">
            <Activity className="w-8 h-8 text-zinc-600" />
            <p>No symbols match the current filter ({selectedType}, {selectedPackage}).</p>
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
              onNodesChange={onNodesChange}
              onEdgesChange={onEdgesChange}
              onNodeClick={onNodeClick}
              fitView
              fitViewOptions={{ padding: 0.15 }}
              minZoom={0.15}
              maxZoom={2.2}
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

        {/* Legend Toolbar */}
        <div className="absolute bottom-4 left-4 z-10 pointer-events-auto hidden md:flex items-center gap-3 bg-[#14171e]/90 backdrop-blur-md px-3.5 py-2 rounded-xl border border-charcoal-700 text-[11px] font-mono text-zinc-300 shadow-xl">
          <span className="text-zinc-500 font-bold">Call Graph:</span>
          <span className="flex items-center gap-1.5">
            <span className="w-2.5 h-2.5 rounded-full bg-emerald-400" />
            <span>Caller (Incoming)</span>
          </span>
          <span className="flex items-center gap-1.5">
            <span className="w-2.5 h-2.5 rounded-full bg-sky-400" />
            <span>Callee (Outgoing)</span>
          </span>
          <span className="flex items-center gap-1.5">
            <span className="w-2.5 h-2.5 rounded-full bg-purple-400" />
            <span>Imports</span>
          </span>
          <span className="text-zinc-600">|</span>
          <span className="text-zinc-400">Click any function to isolate its call path</span>
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
              onClick={() => setFocusedNodeId(null)}
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

          {/* Action Button: Open in Code Editor */}
          {focusedNode.file && (
            <div className="p-3 border-b border-charcoal-800 bg-charcoal-950">
              <button
                onClick={() => handleOpenFile(focusedNode.file, focusedNode.line)}
                className="w-full flex items-center justify-center gap-2 py-2 rounded-lg bg-dna-blue hover:bg-blue-600 text-xs font-mono font-semibold text-white transition-all shadow-md"
              >
                <FileCode className="w-4 h-4" />
                <span>Inspect in Code Editor (Line {focusedNode.line || 1})</span>
              </button>
            </div>
          )}

          {/* Callers & Callees Lists */}
          <div className="flex-1 overflow-y-auto p-4 space-y-5">
            {/* Direct Callers */}
            <div>
              <div className="flex items-center justify-between text-xs font-mono font-bold text-emerald-400 mb-2">
                <span className="flex items-center gap-1.5">
                  <span className="w-2 h-2 rounded-full bg-emerald-400" />
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
                  No direct callers in current codebase. (Entrypoint / Public API)
                </div>
              )}
            </div>

            {/* Direct Callees */}
            <div>
              <div className="flex items-center justify-between text-xs font-mono font-bold text-sky-400 mb-2">
                <span className="flex items-center gap-1.5">
                  <span className="w-2 h-2 rounded-full bg-sky-400" />
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
              <span>Ready for GitHub README, Notion, Obsidian & Architecture Docs</span>
              <span>Format: {exportFormat === 'mermaid' ? 'Mermaid Flowchart' : 'C4 Container'}</span>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
