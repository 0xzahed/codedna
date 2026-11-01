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
import { fetchGraph, GraphNode, GraphEdge } from '../../lib/api';
import { 
  Search, 
  Filter, 
  Layers, 
  Maximize2, 
  RotateCcw, 
  GitFork,
  Flame,
  AlertTriangle,
  ZoomIn
} from 'lucide-react';

export default function GraphView() {
  const { scanId, setSelectedNode } = useAppStore();
  const [graphData, setGraphData] = useState<{ nodes: GraphNode[]; edges: GraphEdge[] }>({ nodes: [], edges: [] });
  const [nodes, setNodes, onNodesChange] = useNodesState([]);
  const [edges, setEdges, onEdgesChange] = useEdgesState([]);
  const [search, setSearch] = useState('');
  const [selectedType, setSelectedType] = useState<string>('all');
  const [layoutMode, setLayoutMode] = useState<'grid' | 'hierarchical'>('hierarchical');
  const [highlightCycle, setHighlightCycle] = useState(false);
  const [loading, setLoading] = useState(true);

  // Load graph data from API
  useEffect(() => {
    setLoading(true);
    fetchGraph(scanId)
      .then((data) => {
        setGraphData(data || { nodes: [], edges: [] });
      })
      .catch((err) => console.error('Failed to fetch graph:', err))
      .finally(() => setLoading(false));
  }, [scanId]);

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
    const rfNodes: Node[] = filteredNodes.map((gn, idx) => {
      let x = 0;
      let y = 0;

      if (layoutMode === 'hierarchical') {
        // Hierarchical clustering by file / type
        const col = idx % 5;
        const row = Math.floor(idx / 5);
        x = col * 260 + 50;
        y = row * 140 + 50;
      } else {
        // Grid layout
        const col = idx % cols;
        const row = Math.floor(idx / cols);
        x = col * 240 + 50;
        y = row * 130 + 50;
      }

      const isHighComplexity = (gn.complexity || 0) > 15;

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
                {isHighComplexity && (
                  <span className="flex items-center text-[10px] font-mono text-dna-rose">
                    <Flame className="w-3 h-3 mr-0.5" />
                    {gn.complexity}
                  </span>
                )}
              </div>
              <div className="text-xs font-mono font-bold text-white truncate max-w-[180px]">
                {gn.label}
              </div>
              {gn.file && (
                <div className="text-[10px] font-mono text-zinc-400 truncate mt-0.5 max-w-[180px]">
                  {gn.file}:{gn.line}
                </div>
              )}
            </div>
          ),
          rawNode: gn,
        },
        style: {
          backgroundColor: '#141518',
          border: isHighComplexity ? '1px solid #f43f5e' : '1px solid #27292e',
          borderRadius: '8px',
          width: 220,
          color: '#ffffff',
          boxShadow: '0 4px 12px rgba(0,0,0,0.5)',
        },
      };
    });

    // Filter edges connecting existing nodes
    const rfEdges: Edge[] = (graphData.edges || [])
      .filter((e) => nodeIds.has(e.source) && nodeIds.has(e.target))
      .map((ge) => ({
        id: ge.id,
        source: ge.source,
        target: ge.target,
        animated: ge.isRuntimeObserved || ge.confidence > 0.9,
        style: {
          stroke: ge.isRuntimeObserved ? '#38bdf8' : (ge.confidence > 0.8 ? '#34d399' : '#71717a'),
          strokeWidth: ge.isRuntimeObserved ? 2 : 1.5,
        },
        markerEnd: {
          type: MarkerType.ArrowClosed,
          color: ge.isRuntimeObserved ? '#38bdf8' : (ge.confidence > 0.8 ? '#34d399' : '#71717a'),
        },
      }));

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
    </div>
  );
}
