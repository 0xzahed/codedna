'use client';

import React, { useState } from 'react';
import { 
  Settings as SettingsIcon, 
  ShieldCheck, 
  Sliders, 
  Terminal, 
  Save, 
  Check, 
  Database,
  Cpu
} from 'lucide-react';

export default function SettingsView() {
  const [ollamaHost, setOllamaHost] = useState('http://localhost:11434');
  const [ollamaModel, setOllamaModel] = useState('llama3.2');
  const [maxFileSize, setMaxFileSize] = useState('2MB');
  const [saved, setSaved] = useState(false);

  const handleSave = (e: React.FormEvent) => {
    e.preventDefault();
    setSaved(true);
    setTimeout(() => setSaved(false), 2000);
  };

  return (
    <div className="p-8 space-y-8 max-w-4xl mx-auto font-mono">
      {/* Header */}
      <div className="border-b border-charcoal-750 pb-6">
        <h1 className="text-2xl font-bold tracking-tight text-white flex items-center gap-3">
          <SettingsIcon className="w-6 h-6 text-zinc-400" />
          <span>Platform &amp; Engine Settings</span>
        </h1>
        <p className="text-sm text-zinc-400 mt-1">
          Configure scanner rules, local AI endpoints, and architectural boundaries.
        </p>
      </div>

      <form onSubmit={handleSave} className="space-y-6">
        {/* Scanner Engine Config */}
        <div className="bg-charcoal-850 border border-charcoal-750 rounded-xl p-6 space-y-4">
          <h3 className="text-sm font-bold text-white flex items-center gap-2 border-b border-charcoal-750 pb-3">
            <Sliders className="w-4 h-4 text-dna-blue" />
            <span>Scanner &amp; Parser Settings</span>
          </h3>

          <div className="space-y-4 text-xs">
            <div>
              <label className="text-zinc-400 block mb-1">Max Parsable File Size:</label>
              <input
                type="text"
                value={maxFileSize}
                onChange={(e) => setMaxFileSize(e.target.value)}
                className="w-full bg-charcoal-900 border border-charcoal-750 rounded-lg p-2 text-white focus:outline-none focus:border-dna-blue"
              />
              <span className="text-[10px] text-zinc-400 mt-1 block">Protects memory from scanning huge generated blobs or minified bundles.</span>
            </div>

            <div>
              <label className="text-zinc-400 block mb-1">Default Ignored Directories:</label>
              <div className="p-3 bg-charcoal-900 border border-charcoal-750 rounded-lg text-zinc-400 text-[11px] leading-relaxed">
                <code>node_modules, .git, vendor, dist, build, target, .venv, __pycache__, .cache, coverage, .next, .nuxt</code>
              </div>
            </div>
          </div>
        </div>

        {/* Local AI Configuration */}
        <div className="bg-charcoal-850 border border-charcoal-750 rounded-xl p-6 space-y-4">
          <h3 className="text-sm font-bold text-white flex items-center gap-2 border-b border-charcoal-750 pb-3">
            <Cpu className="w-4 h-4 text-dna-purple" />
            <span>Local AI Model (Ollama)</span>
          </h3>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-4 text-xs">
            <div>
              <label className="text-zinc-400 block mb-1">Ollama Host URL:</label>
              <input
                type="text"
                value={ollamaHost}
                onChange={(e) => setOllamaHost(e.target.value)}
                className="w-full bg-charcoal-900 border border-charcoal-750 rounded-lg p-2 text-white focus:outline-none focus:border-dna-blue"
              />
            </div>
            <div>
              <label className="text-zinc-400 block mb-1">Ollama Model Name:</label>
              <input
                type="text"
                value={ollamaModel}
                onChange={(e) => setOllamaModel(e.target.value)}
                className="w-full bg-charcoal-900 border border-charcoal-750 rounded-lg p-2 text-white focus:outline-none focus:border-dna-blue"
              />
            </div>
          </div>
        </div>

        {/* Storage Info */}
        <div className="bg-charcoal-850 border border-charcoal-750 rounded-xl p-6 space-y-4">
          <h3 className="text-sm font-bold text-white flex items-center gap-2 border-b border-charcoal-750 pb-3">
            <Database className="w-4 h-4 text-dna-emerald" />
            <span>Persistence &amp; Storage</span>
          </h3>

          <div className="text-xs space-y-2 text-zinc-300">
            <div className="flex justify-between py-1 border-b border-charcoal-800">
              <span className="text-zinc-400">Database Engine:</span>
              <span className="text-white">SQLite 3 (WAL Mode)</span>
            </div>
            <div className="flex justify-between py-1 border-b border-charcoal-800">
              <span className="text-zinc-400">Database Path:</span>
              <span className="text-zinc-400">~/.codedna/codedna.db</span>
            </div>
            <div className="flex justify-between py-1">
              <span className="text-zinc-400">Runtime Tracer Endpoint:</span>
              <span className="text-dna-blue">POST http://localhost:8080/api/runtime/events</span>
            </div>
          </div>
        </div>

        {/* Save button */}
        <div className="flex justify-end">
          <button
            type="submit"
            className="flex items-center gap-2 px-5 py-2.5 rounded-lg bg-dna-blue text-charcoal-950 font-bold text-xs hover:bg-sky-300 transition-colors shadow-md"
          >
            {saved ? <Check className="w-4 h-4" /> : <Save className="w-4 h-4" />}
            <span>{saved ? 'Settings Saved!' : 'Save Preferences'}</span>
          </button>
        </div>
      </form>
    </div>
  );
}
