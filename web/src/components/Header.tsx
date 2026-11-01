'use client';

import React, { useState, useEffect } from 'react';
import { useAppStore } from '../lib/store';
import { triggerScan, fetchReport, fetchRepos } from '../lib/api';
import { 
  RefreshCw, 
  Download, 
  CheckCircle2, 
  FileText,
  Search,
  FolderOpen,
  X,
  AlertCircle,
  ChevronDown
} from 'lucide-react';

interface HeaderProps {
  onSearch?: (q: string) => void;
}

export default function Header({ onSearch }: HeaderProps) {
  const { 
    isScanning, 
    setIsScanning, 
    scanId, 
    setScanId, 
    activeRepoName, 
    setActiveRepoName 
  } = useAppStore();

  const [downloading, setDownloading] = useState(false);
  const [searchVal, setSearchVal] = useState('');
  const [repos, setRepos] = useState<any[]>([]);
  
  // Open / Switch Project Modal state
  const [showModal, setShowModal] = useState(false);
  const [targetInput, setTargetInput] = useState('');
  const [scanStatus, setScanStatus] = useState<string | null>(null);
  const [scanProgress, setScanProgress] = useState<{ done: number; total: number; file: string } | null>(null);
  const [scanError, setScanError] = useState<string | null>(null);

  // Load repositories on mount and when scanId changes
  useEffect(() => {
    fetchRepos()
      .then((data) => {
        if (Array.isArray(data)) {
          setRepos(data);
          const active = data.find((r: any) => r.latestScanId === scanId);
          if (active) {
            setActiveRepoName(active.name);
          } else if (data.length > 0 && !localStorage.getItem('codedna_scan_id')) {
            if (data[0].latestScanId) {
              setScanId(data[0].latestScanId);
              setActiveRepoName(data[0].name);
            }
          }
        }
      })
      .catch(() => {});
  }, [scanId]);

  const handleRescan = async () => {
    try {
      setIsScanning(true);
      const res = await triggerScan();
      if (res.scanId) {
        setScanId(res.scanId);
      }
      const data = await fetchRepos();
      if (Array.isArray(data)) setRepos(data);
    } catch (err) {
      console.error('Scan error:', err);
    } finally {
      setIsScanning(false);
    }
  };

  const handleOpenTarget = async (path: string) => {
    if (!path.trim()) return;
    let eventSource: EventSource | null = null;
    try {
      setScanError(null);
      setIsScanning(true);
      setScanStatus('Starting intelligence pipeline...');
      setScanProgress(null);

      // Listen to SSE progress
      if (typeof window !== 'undefined') {
        eventSource = new EventSource('/api/events');
        eventSource.onmessage = (event) => {
          try {
            const data = JSON.parse(event.data);
            if (data.type === 'scan_progress') {
              setScanProgress({ done: data.done, total: data.total, file: data.file });
              setScanStatus(`Analyzing AST symbols in ${data.file}...`);
            }
          } catch {}
        };
      }

      const res = await triggerScan(path.trim());
      if (res.error) {
        throw new Error(res.error);
      }

      if (res.scanId) {
        setScanId(res.scanId);
      }
      if (res.repo?.name) {
        setActiveRepoName(res.repo.name);
      }

      // Refresh repos list
      const updatedRepos = await fetchRepos();
      if (Array.isArray(updatedRepos)) {
        setRepos(updatedRepos);
      }

      setShowModal(false);
    } catch (err: any) {
      setScanError(err.message || 'Failed to scan target path');
    } finally {
      if (eventSource) {
        eventSource.close();
      }
      setIsScanning(false);
      setScanStatus(null);
      setScanProgress(null);
    }
  };

  const handleExport = async (format: 'markdown' | 'json') => {
    try {
      setDownloading(true);
      const data = await fetchReport(scanId, format);
      const blob = new Blob([typeof data === 'string' ? data : JSON.stringify(data, null, 2)], {
        type: format === 'json' ? 'application/json' : 'text/markdown',
      });
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `codedna-report.${format === 'json' ? 'json' : 'md'}`;
      a.click();
      URL.revokeObjectURL(url);
    } catch (err) {
      console.error('Export error:', err);
    } finally {
      setDownloading(false);
    }
  };

  const progressPercent = scanProgress && scanProgress.total > 0
    ? Math.min(100, Math.round((scanProgress.done / scanProgress.total) * 100))
    : null;

  return (
    <>
      <header className="h-16 border-b border-charcoal-750 bg-charcoal-900 px-6 flex items-center justify-between gap-4 shrink-0">
        {/* Repo Selector & Status */}
        <div className="flex items-center gap-3">
          <div className="relative">
            <select
              value={scanId}
              onChange={(e) => {
                const newId = Number(e.target.value);
                const match = repos.find((r) => r.latestScanId === newId);
                setScanId(newId);
                if (match) setActiveRepoName(match.name);
              }}
              className="bg-charcoal-800 border border-charcoal-700 hover:border-dna-blue/50 text-xs font-mono text-zinc-200 rounded-lg pl-3 pr-8 py-1.5 focus:outline-none cursor-pointer appearance-none transition-colors"
            >
              {repos.length > 0 ? (
                repos.map((r) => (
                  <option key={r.id} value={r.latestScanId || r.id}>
                    📁 {r.name} {r.git?.branch ? `(${r.git.branch})` : ''}
                  </option>
                ))
              ) : (
                <option value={scanId}>📁 {activeRepoName}</option>
              )}
            </select>
            <ChevronDown className="w-3.5 h-3.5 text-zinc-400 absolute right-2.5 top-1/2 -translate-y-1/2 pointer-events-none" />
          </div>

          <div className="text-xs text-zinc-400 font-mono flex items-center gap-1.5 hidden sm:flex">
            <CheckCircle2 className="w-3.5 h-3.5 text-emerald-400" />
            <span>Engine Active</span>
          </div>
        </div>

        {/* Global Search Bar */}
        <div className="flex-1 max-w-md relative">
          <Search className="w-3.5 h-3.5 text-zinc-500 absolute left-3 top-1/2 -translate-y-1/2" />
          <input
            type="text"
            value={searchVal}
            onChange={(e) => {
              setSearchVal(e.target.value);
              onSearch?.(e.target.value);
            }}
            placeholder="Search symbols, files, endpoints, or issues... (Ctrl+K)"
            className="w-full bg-charcoal-850 border border-charcoal-750 rounded-lg pl-9 pr-4 py-1.5 text-xs text-white placeholder-zinc-500 focus:outline-none focus:border-dna-blue/50 focus:ring-1 focus:ring-dna-blue/30 font-mono transition-all"
          />
        </div>

        {/* Action Buttons */}
        <div className="flex items-center gap-2">
          {/* Open / Switch Project Button */}
          <button
            onClick={() => setShowModal(true)}
            className="flex items-center gap-2 px-3 py-1.5 rounded-lg border border-charcoal-700 bg-charcoal-800 hover:bg-charcoal-750 text-xs text-zinc-200 hover:text-white transition-all font-mono"
            title="Scan any local folder or GitHub repository"
          >
            <FolderOpen className="w-3.5 h-3.5 text-dna-blue" />
            <span>Open Codebase</span>
          </button>

          {/* Export Report Dropdown */}
          <div className="relative group">
            <button
              disabled={downloading}
              className="flex items-center gap-2 px-3 py-1.5 rounded-lg border border-charcoal-700 bg-charcoal-800 text-xs text-zinc-300 hover:text-white hover:bg-charcoal-750 transition-all font-mono"
            >
              <Download className="w-3.5 h-3.5 text-zinc-400" />
              <span>Export Report</span>
            </button>
            <div className="hidden group-hover:flex absolute right-0 top-full pt-1 z-50 flex-col">
              <div className="bg-charcoal-850 border border-charcoal-700 rounded-lg shadow-xl py-1 min-w-[140px]">
                <button
                  onClick={() => handleExport('markdown')}
                  className="w-full text-left px-3 py-1.5 text-xs text-zinc-300 hover:text-white hover:bg-charcoal-750 flex items-center gap-2"
                >
                  <FileText className="w-3.5 h-3.5 text-dna-blue" />
                  Markdown (.md)
                </button>
                <button
                  onClick={() => handleExport('json')}
                  className="w-full text-left px-3 py-1.5 text-xs text-zinc-300 hover:text-white hover:bg-charcoal-750 flex items-center gap-2"
                >
                  <Download className="w-3.5 h-3.5 text-dna-emerald" />
                  JSON (.json)
                </button>
              </div>
            </div>
          </div>

          {/* Rescan Button */}
          <button
            onClick={handleRescan}
            disabled={isScanning}
            className={`flex items-center gap-2 px-3.5 py-1.5 rounded-lg bg-dna-blue/15 border border-dna-blue/40 text-dna-blue hover:bg-dna-blue/25 text-xs font-medium font-mono transition-all shadow-sm ${
              isScanning ? 'opacity-60 cursor-not-allowed' : ''
            }`}
          >
            <RefreshCw className={`w-3.5 h-3.5 ${isScanning ? 'animate-spin' : ''}`} />
            <span>{isScanning ? 'Analyzing...' : 'Rescan'}</span>
          </button>
        </div>
      </header>

      {/* Open Codebase Modal */}
      {showModal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/70 backdrop-blur-sm p-4 font-mono">
          <div className="w-full max-w-xl bg-charcoal-900 border border-charcoal-700 rounded-xl shadow-2xl p-6 space-y-5">
            <div className="flex items-center justify-between border-b border-charcoal-750 pb-4">
              <div className="flex items-center gap-2.5">
                <FolderOpen className="w-5 h-5 text-dna-blue" />
                <h2 className="text-sm font-bold text-white">Analyze Any Codebase</h2>
              </div>
              <button
                onClick={() => {
                  if (!isScanning) {
                    setShowModal(false);
                    setScanError(null);
                  }
                }}
                disabled={isScanning}
                className="text-zinc-400 hover:text-white disabled:opacity-30"
              >
                <X className="w-4 h-4" />
              </button>
            </div>

            <p className="text-xs text-zinc-400 leading-relaxed">
              Enter any absolute/relative local directory path or a public GitHub repository URL.
              CodeDNA parses AST symbols, maps caller trees, and visualizes system architecture in real-time.
            </p>

            <form
              onSubmit={(e) => {
                e.preventDefault();
                handleOpenTarget(targetInput);
              }}
              className="space-y-4"
            >
              <div>
                <label className="block text-[11px] text-zinc-300 font-semibold mb-1.5">
                  Local Folder Path or GitHub Repo URL
                </label>
                <input
                  type="text"
                  value={targetInput}
                  onChange={(e) => setTargetInput(e.target.value)}
                  disabled={isScanning}
                  placeholder="/path/to/project, ~/repos/api, or https://github.com/org/repo"
                  className="w-full bg-charcoal-850 border border-charcoal-700 rounded-lg px-3 py-2 text-xs text-white placeholder-zinc-500 focus:outline-none focus:border-dna-blue disabled:opacity-50"
                  autoFocus
                />
              </div>

              {/* Quick suggestions */}
              <div className="space-y-1.5">
                <span className="text-[10px] text-zinc-500 uppercase tracking-wider">Quick Presets:</span>
                <div className="flex flex-wrap gap-2">
                  <button
                    type="button"
                    disabled={isScanning}
                    onClick={() => setTargetInput('/home/zahed/Downloads/i-edu-backend-master')}
                    className="px-2.5 py-1 rounded bg-charcoal-800 hover:bg-charcoal-750 border border-charcoal-700 text-[11px] text-zinc-300 disabled:opacity-50"
                  >
                    📂 i-edu-backend-master
                  </button>
                  <button
                    type="button"
                    disabled={isScanning}
                    onClick={() => setTargetInput('./examples/ecommerce-service')}
                    className="px-2.5 py-1 rounded bg-charcoal-800 hover:bg-charcoal-750 border border-charcoal-700 text-[11px] text-zinc-300 disabled:opacity-50"
                  >
                    📦 ecommerce-service
                  </button>
                  <button
                    type="button"
                    disabled={isScanning}
                    onClick={() => setTargetInput('.')}
                    className="px-2.5 py-1 rounded bg-charcoal-800 hover:bg-charcoal-750 border border-charcoal-700 text-[11px] text-zinc-300 disabled:opacity-50"
                  >
                    ⚡ CodeDNA Engine (.)
                  </button>
                </div>
              </div>

              {scanError && (
                <div className="p-3 bg-rose-500/10 border border-rose-500/30 rounded-lg flex items-start gap-2 text-xs text-rose-300">
                  <AlertCircle className="w-4 h-4 shrink-0 text-rose-400 mt-0.5" />
                  <span>{scanError}</span>
                </div>
              )}

              {/* Real-time progress bar feedback */}
              {isScanning && (
                <div className="p-4 bg-charcoal-850 border border-charcoal-700 rounded-lg space-y-2.5">
                  <div className="flex items-center justify-between text-xs">
                    <span className="text-dna-blue flex items-center gap-2">
                      <RefreshCw className="w-3.5 h-3.5 animate-spin text-dna-blue" />
                      <span>{scanStatus || 'Parsing AST symbols...'}</span>
                    </span>
                    {progressPercent !== null && (
                      <span className="text-zinc-300 font-bold">{progressPercent}%</span>
                    )}
                  </div>

                  {progressPercent !== null && (
                    <div className="w-full bg-charcoal-800 rounded-full h-2 overflow-hidden border border-charcoal-700">
                      <div 
                        className="bg-dna-blue h-full transition-all duration-200" 
                        style={{ width: `${progressPercent}%` }} 
                      />
                    </div>
                  )}

                  {scanProgress && (
                    <div className="flex items-center justify-between text-[11px] text-zinc-500">
                      <span>Parsed {scanProgress.done} of {scanProgress.total} files</span>
                      <span className="truncate max-w-[200px] text-zinc-400">[{scanProgress.file}]</span>
                    </div>
                  )}
                </div>
              )}

              <div className="flex justify-end gap-2 pt-2 border-t border-charcoal-750">
                <button
                  type="button"
                  disabled={isScanning}
                  onClick={() => setShowModal(false)}
                  className="px-3 py-1.5 bg-charcoal-800 hover:bg-charcoal-750 text-xs text-zinc-300 rounded-lg border border-charcoal-700 disabled:opacity-50"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  disabled={!targetInput.trim() || isScanning}
                  className="px-4 py-1.5 bg-dna-blue text-charcoal-950 font-bold hover:bg-dna-blue/90 text-xs rounded-lg transition-all disabled:opacity-50 flex items-center gap-2"
                >
                  {isScanning ? (
                    <>
                      <RefreshCw className="w-3.5 h-3.5 animate-spin" />
                      <span>Scanning...</span>
                    </>
                  ) : (
                    <span>Scan &amp; Open</span>
                  )}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </>
  );
}
