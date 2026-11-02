'use client';

import React, { useEffect, useState } from 'react';
import { useAppStore } from '../lib/store';
import { fetchFileContent } from '../lib/api';
import CodeEditorViewer from './CodeEditorViewer';
import { 
  X, 
  FileCode, 
  ExternalLink, 
  Copy, 
  Check, 
  Layers, 
  GitCommit,
  Flame
} from 'lucide-react';

export default function SourceDrawer() {
  const { selectedNode, setSelectedNode, scanId } = useAppStore();
  const [fileData, setFileData] = useState<{ path: string; content: string } | null>(null);
  const [loading, setLoading] = useState(false);
  const [copied, setCopied] = useState(false);

  useEffect(() => {
    if (!selectedNode?.file) {
      setFileData(null);
      return;
    }

    setLoading(true);
    const targetScanOrRepo = (selectedNode.data?.repoId as number) || scanId;
    fetchFileContent(targetScanOrRepo, selectedNode.file)
      .then((res) => {
        setFileData(res);
      })
      .catch((err) => {
        console.error('Failed to load file:', err);
      })
      .finally(() => {
        setLoading(false);
      });
  }, [selectedNode, scanId]);

  useEffect(() => {
    if (fileData && selectedNode?.line) {
      const timer = setTimeout(() => {
        const el = document.getElementById(`L${selectedNode.line}`);
        if (el) {
          el.scrollIntoView({ behavior: 'smooth', block: 'center' });
        }
      }, 50);
      return () => clearTimeout(timer);
    }
  }, [fileData, selectedNode?.line]);

  if (!selectedNode) return null;

  const handleCopy = () => {
    if (selectedNode.file) {
      navigator.clipboard.writeText(selectedNode.file);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    }
  };

  const lines = fileData?.content ? fileData.content.split('\n') : [];
  const targetLine = selectedNode.line || 1;

  return (
    <div className="fixed inset-y-0 right-0 w-[720px] max-w-[95vw] bg-charcoal-900 border-l border-charcoal-750 shadow-2xl z-50 flex flex-col animate-in slide-in-from-right duration-200">
      {/* Drawer Header */}
      <div className="h-16 border-b border-charcoal-750 px-6 flex items-center justify-between shrink-0 bg-charcoal-850">
        <div className="flex items-center gap-3 truncate">
          <div className="w-8 h-8 rounded-lg bg-dna-blue/10 border border-dna-blue/30 flex items-center justify-center text-dna-blue shrink-0">
            <FileCode className="w-4 h-4" />
          </div>
          <div className="truncate">
            <h3 className="text-sm font-bold text-white truncate font-mono">{selectedNode.label}</h3>
            <p className="text-[11px] text-zinc-400 font-mono truncate">{selectedNode.file}:{selectedNode.line}</p>
          </div>
        </div>

        <button
          onClick={() => setSelectedNode(null)}
          className="p-1.5 rounded-lg text-zinc-400 hover:text-white hover:bg-charcoal-750 transition-colors"
        >
          <X className="w-4 h-4" />
        </button>
      </div>

      {/* Symbol Metrics Ribbon */}
      <div className="px-6 py-2.5 border-b border-charcoal-750 bg-charcoal-950/60 flex items-center gap-4 text-xs font-mono">
        <div className="flex items-center gap-1.5 text-zinc-300">
          <Layers className="w-3.5 h-3.5 text-dna-purple" />
          <span className="capitalize">{selectedNode.type}</span>
        </div>
        <div className="flex items-center gap-1.5 text-zinc-300">
          <Flame className={`w-3.5 h-3.5 ${(selectedNode.complexity || 0) > 15 ? 'text-dna-rose' : 'text-dna-amber'}`} />
          <span>Complexity: {selectedNode.complexity || 1}</span>
        </div>
        {selectedNode.exported !== undefined && (
          <span className={`text-[10px] px-2 py-0.5 rounded border ${
            selectedNode.exported 
              ? 'bg-emerald-500/10 text-emerald-400 border-emerald-500/30' 
              : 'bg-zinc-800 text-zinc-400 border-zinc-700'
          }`}>
            {selectedNode.exported ? 'Exported' : 'Internal'}
          </span>
        )}
      </div>

      {/* Code Editor Viewer */}
      <div className="flex-1 overflow-hidden p-3 bg-charcoal-950">
        {loading ? (
          <div className="h-full flex items-center justify-center text-zinc-500 font-mono text-xs">
            <div className="flex items-center gap-2">
              <div className="w-3.5 h-3.5 border-2 border-dna-blue border-t-transparent rounded-full animate-spin" />
              <span>Loading source code...</span>
            </div>
          </div>
        ) : fileData?.content ? (
          <CodeEditorViewer
            filePath={selectedNode.file || 'source'}
            content={fileData.content}
            targetLine={selectedNode.line || 1}
            language={selectedNode.language}
            maxHeight="h-full"
          />
        ) : (
          <div className="h-full flex items-center justify-center text-zinc-500 font-mono text-xs">
            File preview unavailable.
          </div>
        )}
      </div>

      {/* Footer Info */}
      <div className="p-4 border-t border-charcoal-750 bg-charcoal-850 flex items-center justify-between text-xs font-mono">
        <button
          onClick={handleCopy}
          className="flex items-center gap-1.5 text-zinc-400 hover:text-white transition-colors"
        >
          {copied ? <Check className="w-3.5 h-3.5 text-emerald-400" /> : <Copy className="w-3.5 h-3.5" />}
          <span>{copied ? 'Copied Path!' : 'Copy Path'}</span>
        </button>

        <span className="text-[11px] text-zinc-500">
          Source Location Preserved
        </span>
      </div>
    </div>
  );
}
