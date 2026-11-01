'use client';

import React, { useState } from 'react';
import { useAppStore } from '../lib/store';
import { triggerScan, fetchReport } from '../lib/api';
import { 
  RefreshCw, 
  Download, 
  GitBranch, 
  CheckCircle2, 
  FileText,
  Search
} from 'lucide-react';

interface HeaderProps {
  onSearch?: (q: string) => void;
}

export default function Header({ onSearch }: HeaderProps) {
  const { isScanning, setIsScanning, scanId } = useAppStore();
  const [downloading, setDownloading] = useState(false);
  const [searchVal, setSearchVal] = useState('');

  const handleRescan = async () => {
    try {
      setIsScanning(true);
      await triggerScan();
      // Reload page state
      window.location.reload();
    } catch (err) {
      console.error('Scan error:', err);
    } finally {
      setIsScanning(false);
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

  return (
    <header className="h-16 border-b border-charcoal-750 bg-charcoal-900 px-6 flex items-center justify-between gap-4 shrink-0">
      {/* Repo Context */}
      <div className="flex items-center gap-3">
        <div className="flex items-center gap-2 px-2.5 py-1 rounded-md bg-charcoal-800 border border-charcoal-700 text-xs font-mono text-zinc-300">
          <GitBranch className="w-3.5 h-3.5 text-dna-blue" />
          <span>main</span>
        </div>
        <div className="text-xs text-zinc-400 font-mono flex items-center gap-1.5">
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

        <button
          onClick={handleRescan}
          disabled={isScanning}
          className={`flex items-center gap-2 px-3.5 py-1.5 rounded-lg bg-dna-blue/15 border border-dna-blue/40 text-dna-blue hover:bg-dna-blue/25 text-xs font-medium font-mono transition-all shadow-sm ${
            isScanning ? 'opacity-60 cursor-not-allowed' : ''
          }`}
        >
          <RefreshCw className={`w-3.5 h-3.5 ${isScanning ? 'animate-spin' : ''}`} />
          <span>{isScanning ? 'Analyzing...' : 'Rescan Codebase'}</span>
        </button>
      </div>
    </header>
  );
}
