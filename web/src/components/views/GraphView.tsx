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
  ReactFlowProvider,
} from 'reactflow';
import { useAppStore } from '../../lib/store';
import { fetchGraph, fetchWorkspaceGraph, fetchExportGraph, GraphNode, GraphEdge } from '../../lib/api';
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
  Share2
} from 'lucide-react';

export default function GraphView() {
  const { scanId, setSelectedNode, workspaceId, setWorkspaceId, refreshKey } = useAppStore();
  const [graphData, setGraphData] = useState<{ nodes: GraphNode[]; edges: GraphEdge[] }>({ nodes: [], edges: [] });
  const [nodes, setNodes, onNodesChange] = useNodesState([]);
  const [edges, setEdges, onEdgesChange] = useEdgesState([]);
  const [search, setSearch] = useState('');
  const [selectedType, setSelectedType] = useState<string>('all');
  const [layoutMode, setLayoutMode] = useState<'grid' | 'hierarchical'>('hierarchical');
  const [highlightCycle, setHighlightCycle] = useState(false);
  const [loading, setLoading] = useState(true);

  // Diagram Export Modal state
  const [showExportModal, setShowExportModal] = useState(false);
  const [exportFormat, setExportFormat] = useState<'mermaid' | 'c4'>('mermaid');
  const [exportCode, setExportCode] = useState('');
  const [exportLoading, setExportLoading] = useState(false);
  const [copiedExport, setCopiedExport] = useState(false);

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

  // Load graph data from API (either single-repo or multi-repo workspace)
  useEffect(() => {
    setLoading(true);
    const req = workspaceId ? fetchWorkspaceGraph(workspaceId) : fetchGraph(scanId);
    req
      .then((data) => {
        setGraphData(data || { nodes: [], edges: [] });
      })
      .catch((err) => console.error('Failed to fetch graph:', err))
      .finally(() => setLoading(false));
  }, [scanId, workspaceId, refreshKey]);

  // Compute layout & React Flow nodes/edges
  useEffect(() => {
    if (!graphData.nodes || graphData.nodes.length === 0) {
      setNodes([]);
      setEdges([]);
      return;
    }

    // Filter nodes
    let filteredNodes = graphData.nodes;
    if (selectedType !== 'all') {
      filteredNodes = filteredNodes.filter((n) => n.type === selectedType);
    }
    if (search.trim()) {
      const q = search.toLowerCase();
      filteredNodes = filteredNodes.filter((n) => 
        n.label.toLowerCase().includes(q) || (n.file && n.file.toLowerCase().includes(q))
      );
    }

    const nodeIds = new Set(filteredNodes.map((n) => n.id));

    // Calculate layout positions
    const cols = Math.ceil(Math.sqrt(filteredNodes.length)) || 4;

    // Build directory groups for hierarchical architectural layout
    const groupMap: { [key: string]: number } = {};
    const groupItemCounts: { [key: string]: number } = {};
    filteredNodes.forEach((gn) => {
      const dir = gn.file ? (gn.file.includes('/') ? gn.file.split('/')[0] : 'root') : 'root';
      if (groupMap[dir] === undefined) {
        groupMap[dir] = Object.keys(groupMap).length;
      }
    });

    const rfNodes: Node[] = filteredNodes.map((gn, idx) => {
      let x = 0;
      let y = 0;

      if (layoutMode === 'hierarchical') {
        const dir = gn.file ? (gn.file.includes('/') ? gn.file.split('/')[0] : 'root') : 'root';
        const col = groupMap[dir] || 0;
        const row = groupItemCounts[dir] || 0;
        groupItemCounts[dir] = row + 1;
        x = col * 280 + 50;
        y = row * 130 + 50;
      } else {
        // Grid layout
        const col = idx % cols;
        const row = Math.floor(idx / cols);
        x = col * 260 + 50;
        y = row * 130 + 50;
      }

      const isHighComplexity = (gn.complexity || 0) > 15;
      const connCount = gn.data?.connections as number | undefined;

      return {
        id: gn.id,
        position: { x, y },
        data: {
          label: (
            <div className="p-3 select-none">
              <div className="flex items-center justify-between gap-2 mb-1">
                <span className={`text-[10px] font-mono uppercase px-1.5 py-0.5 rounded border ${
                  gn.type === 'file' 
                    ? 'bg-blue-500/10 text-blue-400 border-blue-500/30'
                    : gn.type === 'class'
                    ? 'bg-purple-500/10 text-purple-400 border-purple-500/30'
                    : 'bg-emerald-500/10 text-emerald-400 border-emerald-500/30'
                }`}>
                  {gn.type}
                </span>
                <div className="flex items-center gap-1.5">
                  {connCount !== undefined && connCount > 0 && (
                    <span className="text-[10px] font-mono px-1.5 py-0.2 rounded bg-charcoal-800 text-zinc-300 border border-charcoal-700">
                      {connCount} rel
                    </span>
                  )}
                  {gn.data?.service && (
                    <span className="text-[10px] font-mono px-1.5 py-0.5 rounded bg-sky-500/15 text-sky-300 border border-sky-500/30 truncate max-w-[85px]">
                      {String(gn.data.service)}
                    </span>
                  )}
                  {isHighComplexity && (
                    <span className="flex items-center text-[10px] font-mono text-dna-rose">
                      <Flame className="w-3 h-3 mr-0.5" />
                      {gn.complexity}
                    </span>
                  )}
                </div>
              </div>
              <div className="text-xs font-mono font-bold text-white truncate max-w-[190px]">
                {gn.label}
              </div>
              {gn.file && (
                <div className="text-[10px] font-mono text-zinc-400 truncate mt-0.5 max-w-[190px]">
                  {gn.file}{gn.line ? `:${gn.line}` : ''}
                </div>
              )}
            </div>
          ),
          rawNode: gn,
        },
        style: {
          backgroundColor: '#141518',
          border: isHighComplexity ? '1px solid #f43f5e' : (connCount && connCount > 0 ? '1px solid #3b82f6' : '1px solid #27292e'),
          borderRadius: '8px',
          width: 230,
          color: '#ffffff',
          boxShadow: '0 4px 12px rgba(0,0,0,0.5)',
        },
      };
    });

    // Filter edges connecting existing nodes
    const rfEdges: Edge[] = (graphData.edges || [])
      .filter((e) => nodeIds.has(e.source) && nodeIds.has(e.target))
      .map((ge) => {
        const isCross = ge.kind === 'cross_service_api';
        const isImport = ge.kind === 'imports';
        const isCall = ge.kind === 'calls';

        let strokeColor = '#818cf8'; // default indigo
        if (isCross) strokeColor = '#38bdf8'; // sky blue
        else if (isCall) strokeColor = '#34d399'; // emerald green
        else if (isImport) strokeColor = '#a78bfa'; // violet/purple

        const labelText = isCross ? (ge.method || 'API') : (ge.method || ge.kind || 'relates');

        return {
          id: ge.id,
          source: ge.source,
          target: ge.target,
          animated: isCross || ge.isRuntimeObserved || isImport,
          label: labelText,
          labelStyle: { fill: strokeColor, fontSize: 10, fontFamily: 'monospace', fontWeight: 600 },
          labelBgStyle: { fill: '#141518', fillOpacity: 0.9 },
          labelBgPadding: [4, 2] as [number, number],
          labelBgBorderRadius: 4,
          style: {
            stroke: strokeColor,
            strokeWidth: isCross ? 2.5 : 1.5,
            strokeDasharray: isCross ? '5,5' : (isImport ? '4,4' : undefined),
          },
          markerEnd: {
            type: MarkerType.ArrowClosed,
            color: strokeColor,
          },
        };
      });

    setNodes(rfNodes);
    setEdges(rfEdges);
  }, [graphData, search, selectedType, layoutMode]);

  const onNodeClick = useCallback((_: any, node: Node) => {
    if (node.data?.rawNode) {
      setSelectedNode(node.data.rawNode);
    }
  }, [setSelectedNode]);

  return (
    <div className="h-[calc(100vh-4rem)] flex flex-col bg-charcoal-950 relative overflow-hidden">
      {/* Top Controls Toolbar */}
      <div className="h-14 border-b border-charcoal-750 bg-charcoal-900/90 backdrop-blur px-6 flex items-center justify-between gap-4 z-10 shrink-0">
        <div className="flex items-center gap-3">
          <div className="relative w-64">
            <Search className="w-3.5 h-3.5 text-zinc-500 absolute left-3 top-1/2 -translate-y-1/2" />
            <input
              type="text"
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              placeholder="Search graph nodes..."
              className="w-full bg-charcoal-850 border border-charcoal-750 rounded-lg pl-8 pr-3 py-1.5 text-xs text-white placeholder-zinc-500 focus:outline-none focus:border-dna-blue/50 font-mono"
            />
          </div>

          {/* Node Type Filters */}
          <div className="flex items-center bg-charcoal-850 border border-charcoal-750 rounded-lg p-0.5 text-xs font-mono">
            {['all', 'function', 'class', 'file'].map((t) => (
              <button
                key={t}
                onClick={() => setSelectedType(t)}
                className={`px-3 py-1 rounded-md capitalize transition-all ${
                  selectedType === t ? 'bg-charcoal-700 text-white font-medium' : 'text-zinc-400 hover:text-zinc-200'
                }`}
              >
                {t}
              </button>
            ))}
          </div>

          {workspaceId && (
            <div className="flex items-center gap-2 px-2.5 py-1 rounded-lg bg-sky-500/10 border border-sky-500/30 text-xs font-mono text-sky-300">
              <span className="w-1.5 h-1.5 rounded-full bg-sky-400 animate-ping" />
              <span>Full-System Architecture Active</span>
              <button
                onClick={() => setWorkspaceId(null)}
                className="ml-1 px-1 text-zinc-400 hover:text-white rounded hover:bg-charcoal-700"
                title="Return to single repo view"
              >
                ✕
              </button>
            </div>
          )}
        </div>

        {/* View Layout Controls */}
        <div className="flex items-center gap-2">
          <div className="flex items-center bg-charcoal-850 border border-charcoal-750 rounded-lg p-0.5 text-xs font-mono">
            <button
              onClick={() => setLayoutMode('hierarchical')}
              className={`px-2.5 py-1 rounded-md transition-all ${
                layoutMode === 'hierarchical' ? 'bg-charcoal-700 text-white' : 'text-zinc-400'
              }`}
            >
              DAG Layout
            </button>
            <button
              onClick={() => setLayoutMode('grid')}
              className={`px-2.5 py-1 rounded-md transition-all ${
                layoutMode === 'grid' ? 'bg-charcoal-700 text-white' : 'text-zinc-400'
              }`}
            >
              Grid
            </button>
          </div>

          <div className="text-xs font-mono text-zinc-400 px-3 py-1 bg-charcoal-850 border border-charcoal-750 rounded-lg">
            Nodes: <span className="text-white font-semibold">{nodes.length}</span> | Edges: <span className="text-white font-semibold">{edges.length}</span>
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

      {/* React Flow Container */}
      <div className="flex-1 w-full h-full relative">
        {loading ? (
          <div className="h-full flex items-center justify-center text-zinc-500 font-mono text-xs">
            Rendering Living Intelligence Map...
          </div>
        ) : nodes.length === 0 ? (
          <div className="h-full flex items-center justify-center text-zinc-500 font-mono text-xs">
            No graph nodes match current filters.
          </div>
        ) : (
          <ReactFlowProvider>
            <ReactFlow
              nodes={nodes}
              edges={edges}
              onNodesChange={onNodesChange}
              onEdgesChange={onEdgesChange}
              onNodeClick={onNodeClick}
              fitView
              minZoom={0.2}
              maxZoom={2}
            >
              <Background color="#1b1c20" gap={20} size={1} />
              <Controls />
              <MiniMap 
                nodeColor={() => '#38bdf8'}
                maskColor="rgba(12, 13, 14, 0.7)"
              />
            </ReactFlow>
          </ReactFlowProvider>
        )}
      </div>

      {/* Export Diagram Modal */}
      {showExportModal && (
        <div className="fixed inset-0 bg-black/75 backdrop-blur-sm z-50 flex items-center justify-center p-4 animate-in fade-in duration-200">
          <div className="bg-charcoal-900 border border-charcoal-750 rounded-xl w-full max-w-2xl shadow-2xl flex flex-col max-h-[85vh] overflow-hidden">
            {/* Modal Header */}
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

            {/* Format Tabs & Toolbar */}
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

            {/* Code Body */}
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

            {/* Footer */}
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
