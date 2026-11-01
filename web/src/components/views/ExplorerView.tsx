'use client';

import React, { useEffect, useState } from 'react';
import { useAppStore } from '../../lib/store';
import { fetchFiles, fetchSymbols, fetchFileContent, FileItem, SymbolItem } from '../../lib/api';
import { 
  FolderTree, 
  FileCode, 
  Search, 
  Layers, 
  Flame, 
  Code2, 
  Tag, 
  ExternalLink,
  ChevronRight
} from 'lucide-react';

export default function ExplorerView() {
  const { scanId, setSelectedNode } = useAppStore();
  const [files, setFiles] = useState<FileItem[]>([]);
  const [symbols, setSymbols] = useState<SymbolItem[]>([]);
  const [selectedFile, setSelectedFile] = useState<FileItem | null>(null);
  const [fileContent, setFileContent] = useState<string>('');
  const [search, setSearch] = useState('');
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    setLoading(true);
    Promise.all([
      fetchFiles(scanId),
      fetchSymbols(scanId),
    ])
      .then(([fileList, symList]) => {
        setFiles(fileList || []);
        setSymbols(symList || []);
        if (fileList?.length > 0) {
          setSelectedFile(fileList[0]);
        } else {
          setSelectedFile(null);
          setFileContent('');
        }
      })
      .catch((err) => console.error(err))
      .finally(() => setLoading(false));
  }, [scanId]);

  useEffect(() => {
    if (!selectedFile) return;
    fetchFileContent(scanId, selectedFile.path)
      .then((res) => {
        setFileContent(res.content || '');
      })
      .catch((err) => console.error(err));
  }, [selectedFile, scanId]);

  const filteredFiles = (Array.isArray(files) ? files : []).filter((f) => 
    (f?.path || '').toLowerCase().includes(search.toLowerCase()) || 
    (f?.language || '').toLowerCase().includes(search.toLowerCase())
  );

  const fileSymbols = (Array.isArray(symbols) ? symbols : []).filter((s) => s?.file === selectedFile?.path);
  const codeLines = fileContent ? fileContent.split('\n') : [];

  return (
    <div className="h-[calc(100vh-4rem)] flex overflow-hidden">
      {/* Left Pane: Files List */}
      <div className="w-80 border-r border-charcoal-750 bg-charcoal-900 flex flex-col shrink-0">
        <div className="p-3 border-b border-charcoal-750 bg-charcoal-850">
          <div className="relative">
            <Search className="w-3.5 h-3.5 text-zinc-500 absolute left-3 top-1/2 -translate-y-1/2" />
            <input
              type="text"
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              placeholder="Filter files..."
              className="w-full bg-charcoal-900 border border-charcoal-750 rounded-lg pl-8 pr-3 py-1.5 text-xs text-white placeholder-zinc-500 focus:outline-none focus:border-dna-blue/50 font-mono"
            />
          </div>
          <div className="text-[11px] text-zinc-400 font-mono mt-2 px-1 flex items-center justify-between">
            <span>{filteredFiles.length} files tracked</span>
            <span>Total: {files.length}</span>
          </div>
        </div>

        <div className="flex-1 overflow-y-auto divide-y divide-charcoal-750/50">
          {filteredFiles.map((file) => {
            const isSelected = selectedFile?.path === file.path;
            return (
              <button
                key={file.path}
                onClick={() => setSelectedFile(file)}
                className={`w-full text-left px-3.5 py-2.5 transition-colors flex items-center justify-between group ${
                  isSelected ? 'bg-charcoal-750 border-l-2 border-dna-blue' : 'hover:bg-charcoal-800'
                }`}
              >
                <div className="truncate pr-2">
                  <div className="text-xs font-mono font-medium text-white truncate flex items-center gap-1.5">
                    <FileCode className={`w-3.5 h-3.5 shrink-0 ${isSelected ? 'text-dna-blue' : 'text-zinc-400'}`} />
                    <span className="truncate">{file.path}</span>
                  </div>
                  <div className="text-[10px] text-zinc-400 font-mono mt-0.5 flex items-center gap-2">
                    <span className="uppercase text-zinc-400">{file.language}</span>
                    <span>•</span>
                    <span>{file.loc} LOC</span>
                    {file.tags && file.tags.length > 0 && (
                      <span className="px-1 rounded bg-charcoal-700 text-[9px] text-zinc-300">
                        {file.tags[0]}
                      </span>
                    )}
                  </div>
                </div>
                <ChevronRight className={`w-3.5 h-3.5 text-zinc-600 group-hover:text-zinc-300 transition-colors ${isSelected ? 'text-white' : ''}`} />
              </button>
            );
          })}
        </div>
      </div>

      {/* Right Pane: File Outline & Source Inspector */}
      <div className="flex-1 flex flex-col bg-charcoal-950 overflow-hidden">
        {selectedFile ? (
          <>
            {/* File Header Bar */}
            <div className="h-14 border-b border-charcoal-750 bg-charcoal-900 px-6 flex items-center justify-between shrink-0">
              <div className="flex items-center gap-3">
                <FileCode className="w-4 h-4 text-dna-blue" />
                <span className="text-sm font-bold text-white font-mono">{selectedFile.path}</span>
                <span className="text-xs text-zinc-400 font-mono">({selectedFile.loc} lines, {selectedFile.size} bytes)</span>
              </div>
              <div className="flex items-center gap-2">
                <span className="text-xs text-zinc-400 font-mono bg-charcoal-800 px-2 py-1 rounded border border-charcoal-700">
                  {fileSymbols.length} Symbols
                </span>
              </div>
            </div>

            {/* Split view: Symbols Outline (top) & Code Preview (bottom) */}
            <div className="flex-1 flex flex-col overflow-hidden">
              {/* Symbols Ribbon */}
              {fileSymbols.length > 0 && (
                <div className="p-3 border-b border-charcoal-750 bg-charcoal-900/60 overflow-x-auto shrink-0 flex items-center gap-2">
                  <span className="text-[10px] uppercase font-mono text-zinc-400 pl-2">Symbols:</span>
                  {fileSymbols.map((sym) => (
                    <button
                      key={sym.id}
                      onClick={() => setSelectedNode({ id: sym.id, type: sym.kind, label: sym.name, file: sym.file, line: sym.line, complexity: sym.complexity, exported: sym.exported })}
                      className="px-2.5 py-1 rounded-md bg-charcoal-800 hover:bg-charcoal-750 border border-charcoal-700 text-xs font-mono text-zinc-200 hover:text-white flex items-center gap-1.5 shrink-0 transition-all"
                    >
                      <Layers className="w-3 h-3 text-dna-purple" />
                      <span>{sym.name}</span>
                      <span className="text-[10px] text-zinc-400">L{sym.line}</span>
                    </button>
                  ))}
                </div>
              )}

              {/* Code viewer */}
              <div className="flex-1 overflow-auto p-4 font-mono text-xs leading-relaxed select-text bg-charcoal-950">
                {codeLines.map((line, idx) => (
                  <div key={idx + 1} className="flex items-start hover:bg-charcoal-800/40 px-2 py-0.5 rounded">
                    <span className="w-12 shrink-0 text-right pr-4 text-zinc-600 select-none text-[11px]">
                      {idx + 1}
                    </span>
                    <pre className="flex-1 overflow-x-auto whitespace-pre font-mono text-[11px] text-zinc-300">
                      {line || ' '}
                    </pre>
                  </div>
                ))}
              </div>
            </div>
          </>
        ) : (
          <div className="h-full flex items-center justify-center text-zinc-500 font-mono text-xs">
            Select a file to inspect its source code and symbols.
          </div>
        )}
      </div>
    </div>
  );
}
