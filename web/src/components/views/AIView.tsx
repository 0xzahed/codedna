'use client';

import React, { useEffect, useState } from 'react';
import { useAppStore } from '../../lib/store';
import { queryAI, fetchAIContext, fetchBusinessLogic } from '../../lib/api';
import { 
  Sparkles, 
  Send, 
  ShieldAlert, 
  Copy, 
  Check, 
  Download, 
  Cpu, 
  FileCode,
  Layers,
  HelpCircle
} from 'lucide-react';

export default function AIView() {
  const { scanId, setSelectedNode } = useAppStore();
  const [subTab, setSubTab] = useState<'chat' | 'business' | 'export'>('chat');
  const [query, setQuery] = useState('');
  const [messages, setMessages] = useState<Array<{ role: 'user' | 'assistant'; text: string; refs?: any[]; provider?: string }>>([
    {
      role: 'assistant',
      text: 'Hello! I am CodeDNA Assistant. Ask me anything about this codebase’s structure, architecture, call flows, or dependencies.',
      provider: 'deterministic-engine',
    },
  ]);
  const [loading, setLoading] = useState(false);
  const [consent, setConsent] = useState(false);
  const [provider, setProvider] = useState<'offline' | 'ollama' | 'openai'>('offline');
  const [copied, setCopied] = useState(false);
  const [contextExport, setContextExport] = useState('');
  const [businessDomains, setBusinessDomains] = useState<any[]>([]);

  useEffect(() => {
    fetchBusinessLogic(scanId)
      .then((res) => setBusinessDomains(res || []))
      .catch((err) => console.error(err));

    fetchAIContext('claude')
      .then((txt) => setContextExport(txt))
      .catch((err) => console.error(err));
  }, [scanId]);

  const handleSend = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!query.trim() || loading) return;

    const userText = query;
    setQuery('');
    setMessages((prev) => [...prev, { role: 'user', text: userText }]);
    setLoading(true);

    try {
      const res = await queryAI({
        query: userText,
        provider: provider,
        userConsent: consent,
      });

      if (res.error) {
        setMessages((prev) => [
          ...prev,
          {
            role: 'assistant',
            text: `⚠ Privacy Notice: ${res.error}`,
            provider: 'privacy-guard',
          },
        ]);
      } else {
        setMessages((prev) => [
          ...prev,
          {
            role: 'assistant',
            text: res.answer || 'No analysis available.',
            refs: res.sourceRefs,
            provider: res.provider,
          },
        ]);
      }
    } catch (err: any) {
      setMessages((prev) => [
        ...prev,
        {
          role: 'assistant',
          text: `Error querying intelligence engine: ${err?.message || 'Network error'}`,
        },
      ]);
    } finally {
      setLoading(false);
    }
  };

  const handleCopyExport = () => {
    navigator.clipboard.writeText(contextExport);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  return (
    <div className="p-6 md:p-8 space-y-6 w-full max-w-[1920px] mx-auto">
      {/* Header */}
      <div className="flex items-center justify-between border-b border-charcoal-750 pb-6">
        <div>
          <h1 className="text-2xl font-bold tracking-tight text-white flex items-center gap-3">
            <Sparkles className="w-6 h-6 text-dna-blue" />
            <span>AI Codebase Intelligence</span>
          </h1>
          <p className="text-sm text-zinc-400 mt-1">
            Grounded codebase Q&amp;A, business logic mapping, and structured context export.
          </p>
        </div>

        {/* View Mode Switcher */}
        <div className="flex items-center bg-charcoal-850 border border-charcoal-750 rounded-lg p-1 text-xs font-mono">
          <button
            onClick={() => setSubTab('chat')}
            className={`px-3 py-1.5 rounded-md transition-all ${
              subTab === 'chat' ? 'bg-charcoal-700 text-white font-medium' : 'text-zinc-400 hover:text-white'
            }`}
          >
            Codebase Q&amp;A
          </button>
          <button
            onClick={() => setSubTab('business')}
            className={`px-3 py-1.5 rounded-md transition-all ${
              subTab === 'business' ? 'bg-charcoal-700 text-white font-medium' : 'text-zinc-400 hover:text-white'
            }`}
          >
            Business Logic
          </button>
          <button
            onClick={() => setSubTab('export')}
            className={`px-3 py-1.5 rounded-md transition-all ${
              subTab === 'export' ? 'bg-charcoal-700 text-white font-medium' : 'text-zinc-400 hover:text-white'
            }`}
          >
            Context Export
          </button>
        </div>
      </div>

      {subTab === 'chat' && (
        <div className="grid grid-cols-1 lg:grid-cols-4 gap-6">
          {/* Main Chat Box */}
          <div className="lg:col-span-3 bg-charcoal-850 border border-charcoal-750 rounded-xl flex flex-col h-[600px] overflow-hidden">
            {/* Messages Area */}
            <div className="flex-1 overflow-y-auto p-5 space-y-4 font-mono text-xs">
              {messages.map((m, i) => (
                <div
                  key={i}
                  className={`flex flex-col ${
                    m.role === 'user' ? 'items-end' : 'items-start'
                  }`}
                >
                  <div
                    className={`max-w-[85%] rounded-xl p-4 leading-relaxed ${
                      m.role === 'user'
                        ? 'bg-dna-blue/15 text-white border border-dna-blue/30'
                        : 'bg-charcoal-900 text-zinc-200 border border-charcoal-750'
                    }`}
                  >
                    <div className="whitespace-pre-wrap">{m.text}</div>

                    {m.refs && m.refs.length > 0 && (
                      <div className="mt-3 pt-3 border-t border-charcoal-800 flex flex-wrap gap-1.5">
                        <span className="text-[10px] text-zinc-400">Sources:</span>
                        {m.refs.map((ref, rIdx) => (
                          <button
                            key={rIdx}
                            onClick={() => setSelectedNode({ id: ref.file, type: 'file', label: ref.file, file: ref.file, line: ref.line })}
                            className="text-[10px] px-2 py-0.5 rounded bg-charcoal-800 hover:bg-charcoal-750 text-dna-blue border border-charcoal-700 transition-colors"
                          >
                            {ref.file}:{ref.line}
                          </button>
                        ))}
                      </div>
                    )}
                  </div>
                  {m.provider && (
                    <span className="text-[9px] text-zinc-400 mt-1 px-1 font-mono">
                      via {m.provider}
                    </span>
                  )}
                </div>
              ))}
              {loading && (
                <div className="text-zinc-400 text-xs italic animate-pulse">
                  Analyzing AST knowledge graph...
                </div>
              )}
            </div>

            {/* Input Form */}
            <form onSubmit={handleSend} className="p-3 border-t border-charcoal-750 bg-charcoal-900 flex items-center gap-2">
              <input
                type="text"
                value={query}
                onChange={(e) => setQuery(e.target.value)}
                placeholder="Ask about relationships, execution paths, architecture, or functions..."
                className="flex-1 bg-charcoal-850 border border-charcoal-750 rounded-lg px-4 py-2.5 text-xs text-white placeholder-zinc-500 focus:outline-none focus:border-dna-blue font-mono"
              />
              <button
                type="submit"
                disabled={loading || !query.trim()}
                className="px-4 py-2.5 bg-dna-blue text-charcoal-950 font-bold rounded-lg hover:bg-sky-300 disabled:opacity-40 transition-colors text-xs font-mono flex items-center gap-2"
              >
                <span>Ask</span>
                <Send className="w-3.5 h-3.5" />
              </button>
            </form>
          </div>

          {/* Privacy & Provider Settings Sidebar */}
          <div className="bg-charcoal-850 border border-charcoal-750 rounded-xl p-5 space-y-5 font-mono text-xs">
            <h3 className="font-bold text-white uppercase text-[11px] tracking-wider border-b border-charcoal-750 pb-2 flex items-center gap-2">
              <ShieldAlert className="w-4 h-4 text-dna-amber" />
              <span>Privacy Safeguards</span>
            </h3>

            <div className="space-y-3">
              <label className="text-zinc-400 text-[11px] block">AI Engine Provider:</label>
              <select
                value={provider}
                onChange={(e: any) => setProvider(e.target.value)}
                className="w-full bg-charcoal-900 border border-charcoal-750 rounded-lg p-2 text-white text-xs font-mono focus:outline-none focus:border-dna-blue"
              >
                <option value="offline">Deterministic AST Graph (Offline)</option>
                <option value="ollama">Local Ollama (Local AI)</option>
                <option value="openai">OpenAI / Compatible (Remote)</option>
              </select>
            </div>

            <div className="p-3 bg-charcoal-900 rounded-lg border border-charcoal-750 space-y-2">
              <div className="flex items-center gap-2 text-zinc-300">
                <input
                  type="checkbox"
                  id="consent"
                  checked={consent}
                  onChange={(e) => setConsent(e.target.checked)}
                  className="rounded bg-charcoal-800 border-charcoal-700 text-dna-blue focus:ring-0"
                />
                <label htmlFor="consent" className="cursor-pointer text-[11px] select-none">
                  Explicit Remote Consent
                </label>
              </div>
              <p className="text-[10px] text-zinc-400 leading-normal">
                Never send source code to external AI providers without user consent.
              </p>
            </div>
          </div>
        </div>
      )}

      {subTab === 'business' && (
        <div className="space-y-4 font-mono">
          <p className="text-xs text-zinc-400">
            Discovered business functionality modules mapped from routes, controllers, and data models:
          </p>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            {(Array.isArray(businessDomains) ? businessDomains : []).map((dom, i) => (
              <div key={i} className="p-5 bg-charcoal-850 border border-charcoal-750 rounded-xl space-y-3">
                <div className="flex items-center justify-between">
                  <h3 className="text-sm font-bold text-white flex items-center gap-2">
                    <Layers className="w-4 h-4 text-dna-blue" />
                    <span>{dom.domain}</span>
                  </h3>
                  <span className="text-[10px] px-2 py-0.5 rounded bg-emerald-500/10 text-emerald-400 border border-emerald-500/30">
                    confidence: {(dom.confidence * 100).toFixed(0)}%
                  </span>
                </div>
                <p className="text-xs text-zinc-400 leading-relaxed">{dom.description}</p>
                <div className="pt-2 border-t border-charcoal-750 space-y-1">
                  <span className="text-[10px] text-zinc-400">Key Domain Symbols:</span>
                  <div className="flex flex-wrap gap-1.5">
                    {dom.keySymbols?.map((sym: string, sIdx: number) => (
                      <span key={sIdx} className="text-[10px] px-2 py-0.5 rounded bg-charcoal-900 border border-charcoal-700 text-zinc-300">
                        {sym}
                      </span>
                    ))}
                  </div>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      {subTab === 'export' && (
        <div className="bg-charcoal-850 border border-charcoal-750 rounded-xl p-6 space-y-4 font-mono">
          <div className="flex items-center justify-between border-b border-charcoal-750 pb-3">
            <div>
              <h3 className="text-sm font-bold text-white">Export Context for Developer AI Tools</h3>
              <p className="text-xs text-zinc-400 mt-0.5">High-fidelity compressed context for Claude Code, Cursor, and Codex.</p>
            </div>
            <button
              onClick={handleCopyExport}
              className="flex items-center gap-2 px-3 py-1.5 rounded-lg bg-charcoal-800 hover:bg-charcoal-750 border border-charcoal-700 text-xs text-white transition-colors"
            >
              {copied ? <Check className="w-3.5 h-3.5 text-emerald-400" /> : <Copy className="w-3.5 h-3.5" />}
              <span>{copied ? 'Copied Context!' : 'Copy Context'}</span>
            </button>
          </div>

          <pre className="p-4 bg-charcoal-950 rounded-lg text-xs text-zinc-300 font-mono overflow-auto max-h-[480px] whitespace-pre leading-relaxed select-text border border-charcoal-800">
            {contextExport}
          </pre>
        </div>
      )}
    </div>
  );
}
