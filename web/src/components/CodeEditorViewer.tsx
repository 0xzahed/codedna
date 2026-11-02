'use client';

import React, { useState, useMemo, useEffect, useRef } from 'react';
import { 
  Copy, 
  Check, 
  Search, 
  WrapText, 
  FileCode, 
  Hash, 
  ChevronDown, 
  ChevronUp, 
  Terminal,
  Maximize2,
  Minimize2
} from 'lucide-react';

interface CodeEditorViewerProps {
  filePath: string;
  content: string;
  targetLine?: number;
  language?: string;
  onSymbolClick?: (symbolName: string) => void;
  className?: string;
  maxHeight?: string;
}

// Token types for syntax highlighting
type TokenType = 
  | 'comment'
  | 'string'
  | 'keyword'
  | 'type'
  | 'number'
  | 'boolean'
  | 'function'
  | 'operator'
  | 'punctuation'
  | 'plain';

interface Token {
  type: TokenType;
  text: string;
}

// Tokenizer supporting Go, TS/JS, Python, Java, PHP, Rust, C/C++, JSON, SQL
function tokenizeLine(line: string, lang: string): Token[] {
  if (!line) return [{ type: 'plain', text: ' ' }];

  const tokens: Token[] = [];
  let remaining = line;

  const keywordsGo = new Set([
    'package', 'import', 'func', 'return', 'type', 'struct', 'interface',
    'if', 'else', 'for', 'range', 'switch', 'case', 'default', 'select',
    'chan', 'go', 'defer', 'map', 'const', 'var', 'make', 'new', 'len',
    'append', 'cap', 'panic', 'recover', 'fallthrough', 'continue', 'break'
  ]);

  const keywordsJS = new Set([
    'import', 'export', 'from', 'default', 'const', 'let', 'var', 'function',
    'return', 'if', 'else', 'for', 'while', 'do', 'switch', 'case', 'break',
    'continue', 'try', 'catch', 'finally', 'throw', 'async', 'await', 'class',
    'extends', 'implements', 'interface', 'type', 'new', 'this', 'super',
    'typeof', 'instanceof', 'void', 'delete', 'in', 'of', 'yield'
  ]);

  const keywordsPy = new Set([
    'def', 'class', 'import', 'from', 'as', 'return', 'if', 'elif', 'else',
    'for', 'while', 'break', 'continue', 'pass', 'try', 'except', 'finally',
    'raise', 'with', 'yield', 'async', 'await', 'lambda', 'global', 'nonlocal',
    'assert', 'del', 'and', 'or', 'not', 'is', 'in'
  ]);

  const commonTypes = new Set([
    'string', 'int', 'int32', 'int64', 'uint', 'uint32', 'uint64', 'float32',
    'float64', 'bool', 'byte', 'rune', 'error', 'any', 'void', 'null', 'undefined',
    'Promise', 'Array', 'Record', 'Map', 'Set', 'Object', 'Boolean', 'Number',
    'String', 'Symbol', 'never', 'unknown', 'self', 'cls'
  ]);

  const isCommentStart = (str: string) => {
    return str.startsWith('//') || str.startsWith('#') || str.startsWith('/*') || str.startsWith('*');
  };

  while (remaining.length > 0) {
    // 1. Comments
    if (isCommentStart(remaining)) {
      tokens.push({ type: 'comment', text: remaining });
      break;
    }

    // 2. Strings: double quote, single quote, backtick
    if (remaining[0] === '"' || remaining[0] === "'" || remaining[0] === '`') {
      const quote = remaining[0];
      let endIdx = 1;
      let escaped = false;
      while (endIdx < remaining.length) {
        if (remaining[endIdx] === '\\' && !escaped) {
          escaped = true;
          endIdx++;
          continue;
        }
        if (remaining[endIdx] === quote && !escaped) {
          endIdx++;
          break;
        }
        escaped = false;
        endIdx++;
      }
      tokens.push({ type: 'string', text: remaining.slice(0, endIdx) });
      remaining = remaining.slice(endIdx);
      continue;
    }

    // 3. Numbers
    const numMatch = remaining.match(/^(0x[0-9a-fA-F]+|\d+(\.\d+)?([eE][+-]?\d+)?)\b/);
    if (numMatch) {
      tokens.push({ type: 'number', text: numMatch[0] });
      remaining = remaining.slice(numMatch[0].length);
      continue;
    }

    // 4. Identifiers / Words
    const wordMatch = remaining.match(/^[a-zA-Z_$][a-zA-Z0-9_$]*/);
    if (wordMatch) {
      const word = wordMatch[0];
      const lower = word.toLowerCase();

      if (word === 'true' || word === 'false' || word === 'nil' || word === 'null' || word === 'None' || word === 'True' || word === 'False') {
        tokens.push({ type: 'boolean', text: word });
      } else if (
        keywordsGo.has(word) || 
        keywordsJS.has(word) || 
        keywordsPy.has(word)
      ) {
        tokens.push({ type: 'keyword', text: word });
      } else if (commonTypes.has(word) || (/^[A-Z][a-zA-Z0-9]*$/.test(word) && !remaining.slice(word.length).trim().startsWith('('))) {
        tokens.push({ type: 'type', text: word });
      } else if (remaining.slice(word.length).trim().startsWith('(')) {
        tokens.push({ type: 'function', text: word });
      } else {
        tokens.push({ type: 'plain', text: word });
      }

      remaining = remaining.slice(word.length);
      continue;
    }

    // 5. Operators
    const opMatch = remaining.match(/^(:=|===|!==|==|!=|<=|>=|=>|->|\+\+|--|\+=|-=|\*=|\/=|&&|\|\||<<|>>|[+\-*\/%&|^!~=<>?])/);
    if (opMatch) {
      tokens.push({ type: 'operator', text: opMatch[0] });
      remaining = remaining.slice(opMatch[0].length);
      continue;
    }

    // 6. Punctuation / Brackets
    const punctMatch = remaining.match(/^([(){}\[\],;:.])/);
    if (punctMatch) {
      tokens.push({ type: 'punctuation', text: punctMatch[0] });
      remaining = remaining.slice(punctMatch[0].length);
      continue;
    }

    // 7. Whitespace or unhandled char
    tokens.push({ type: 'plain', text: remaining[0] });
    remaining = remaining.slice(1);
  }

  return tokens;
}

const TOKEN_STYLES: Record<TokenType, string> = {
  comment: 'text-zinc-500 italic',
  string: 'text-emerald-300',
  keyword: 'text-purple-400 font-semibold',
  type: 'text-sky-300 font-medium',
  number: 'text-amber-300',
  boolean: 'text-amber-400 font-semibold',
  function: 'text-yellow-200',
  operator: 'text-pink-400',
  punctuation: 'text-zinc-400',
  plain: 'text-zinc-200',
};

export default function CodeEditorViewer({
  filePath,
  content,
  targetLine = 1,
  language = '',
  className = '',
  maxHeight = 'h-full',
}: CodeEditorViewerProps) {
  const [copied, setCopied] = useState(false);
  const [wrapLines, setWrapLines] = useState(false);
  const [searchQuery, setSearchQuery] = useState('');
  const [showSearch, setShowSearch] = useState(false);
  const [currentMatchIndex, setCurrentMatchIndex] = useState(0);
  const [fullscreen, setFullscreen] = useState(false);
  const activeLineRef = useRef<HTMLDivElement | null>(null);

  const lines = useMemo(() => {
    if (!content) return [];
    return content.split('\n');
  }, [content]);

  const inferredLang = useMemo(() => {
    if (language) return language.toLowerCase();
    const ext = filePath.split('.').pop()?.toLowerCase();
    switch (ext) {
      case 'go': return 'go';
      case 'ts':
      case 'tsx': return 'typescript';
      case 'js':
      case 'jsx': return 'javascript';
      case 'py': return 'python';
      case 'java': return 'java';
      case 'php': return 'php';
      case 'rs': return 'rust';
      case 'json': return 'json';
      case 'sql': return 'sql';
      default: return 'text';
    }
  }, [filePath, language]);

  // Search matches
  const matchLines = useMemo(() => {
    if (!searchQuery.trim()) return [];
    const query = searchQuery.toLowerCase();
    const matches: number[] = [];
    lines.forEach((line, idx) => {
      if (line.toLowerCase().includes(query)) {
        matches.push(idx + 1);
      }
    });
    return matches;
  }, [searchQuery, lines]);

  // Auto-scroll to target line on load or line change
  useEffect(() => {
    if (targetLine && targetLine > 0) {
      const timer = setTimeout(() => {
        const el = document.getElementById(`editor-line-${targetLine}`);
        if (el) {
          el.scrollIntoView({ behavior: 'smooth', block: 'center' });
        }
      }, 100);
      return () => clearTimeout(timer);
    }
  }, [targetLine, filePath]);

  const handleCopyCode = () => {
    navigator.clipboard.writeText(content);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  const handleNextMatch = () => {
    if (matchLines.length === 0) return;
    const nextIdx = (currentMatchIndex + 1) % matchLines.length;
    setCurrentMatchIndex(nextIdx);
    const target = matchLines[nextIdx];
    document.getElementById(`editor-line-${target}`)?.scrollIntoView({ behavior: 'smooth', block: 'center' });
  };

  const handlePrevMatch = () => {
    if (matchLines.length === 0) return;
    const prevIdx = (currentMatchIndex - 1 + matchLines.length) % matchLines.length;
    setCurrentMatchIndex(prevIdx);
    const target = matchLines[prevIdx];
    document.getElementById(`editor-line-${target}`)?.scrollIntoView({ behavior: 'smooth', block: 'center' });
  };

  return (
    <div className={`flex flex-col rounded-xl border border-charcoal-750 bg-[#0d0f12] overflow-hidden select-text ${fullscreen ? 'fixed inset-4 z-50 shadow-2xl' : maxHeight} ${className}`}>
      {/* Editor Title Bar & Controls (VS Code / JetBrains Style) */}
      <div className="h-10 border-b border-charcoal-800 bg-[#13161b] px-4 flex items-center justify-between shrink-0 select-none">
        {/* Left: Tab with Language Badge */}
        <div className="flex items-center gap-2.5 truncate">
          <div className="flex items-center gap-2 bg-[#1b1f26] px-3 py-1.5 rounded-t-md border-t-2 border-dna-blue text-xs font-mono font-medium text-zinc-200">
            <FileCode className="w-3.5 h-3.5 text-dna-blue" />
            <span className="truncate max-w-[280px]">{filePath.split('/').pop() || 'Untitled'}</span>
            <span className="text-[10px] uppercase font-mono px-1 py-0.2 rounded bg-charcoal-800 text-zinc-400 border border-charcoal-700">
              {inferredLang}
            </span>
          </div>

          <span className="text-[11px] text-zinc-500 font-mono hidden sm:inline">
            {lines.length} lines • {(new TextEncoder().encode(content).length / 1024).toFixed(1)} KB
          </span>
        </div>

        {/* Right: Search, Wrap, Copy, Fullscreen */}
        <div className="flex items-center gap-1.5 text-xs">
          {/* Search Trigger */}
          <button
            onClick={() => setShowSearch(!showSearch)}
            className={`p-1.5 rounded hover:bg-charcoal-750 transition-colors ${showSearch ? 'text-dna-blue bg-charcoal-800' : 'text-zinc-400 hover:text-white'}`}
            title="Search in file (Ctrl+F)"
          >
            <Search className="w-3.5 h-3.5" />
          </button>

          {/* Wrap Toggle */}
          <button
            onClick={() => setWrapLines(!wrapLines)}
            className={`p-1.5 rounded hover:bg-charcoal-750 transition-colors ${wrapLines ? 'text-dna-blue bg-charcoal-800' : 'text-zinc-400 hover:text-white'}`}
            title="Toggle Word Wrap"
          >
            <WrapText className="w-3.5 h-3.5" />
          </button>

          {/* Copy Button */}
          <button
            onClick={handleCopyCode}
            className="flex items-center gap-1.5 px-2 py-1 rounded bg-[#1c2028] hover:bg-[#252b36] border border-charcoal-700 text-zinc-300 hover:text-white transition-all text-xs font-mono"
            title="Copy entire code"
          >
            {copied ? <Check className="w-3.5 h-3.5 text-emerald-400" /> : <Copy className="w-3.5 h-3.5" />}
            <span className="hidden md:inline">{copied ? 'Copied!' : 'Copy'}</span>
          </button>

          {/* Fullscreen Toggle */}
          <button
            onClick={() => setFullscreen(!fullscreen)}
            className="p-1.5 rounded text-zinc-400 hover:text-white hover:bg-charcoal-750 transition-colors"
            title={fullscreen ? 'Exit Fullscreen' : 'Fullscreen Editor'}
          >
            {fullscreen ? <Minimize2 className="w-3.5 h-3.5" /> : <Maximize2 className="w-3.5 h-3.5" />}
          </button>
        </div>
      </div>

      {/* Embedded Search Bar */}
      {showSearch && (
        <div className="bg-[#181c23] border-b border-charcoal-800 px-4 py-2 flex items-center justify-between gap-3 text-xs font-mono animate-in slide-in-from-top-1 duration-150">
          <div className="flex items-center gap-2 flex-1 max-w-md">
            <Search className="w-3.5 h-3.5 text-zinc-400 shrink-0" />
            <input
              type="text"
              value={searchQuery}
              onChange={(e) => {
                setSearchQuery(e.target.value);
                setCurrentMatchIndex(0);
              }}
              placeholder="Find in file..."
              className="w-full bg-[#0d0f12] border border-charcoal-750 rounded px-2 py-1 text-xs text-white focus:outline-none focus:border-dna-blue"
              autoFocus
            />
          </div>

          <div className="flex items-center gap-2">
            <span className="text-[11px] text-zinc-400">
              {matchLines.length > 0
                ? `${currentMatchIndex + 1} of ${matchLines.length} matches`
                : searchQuery
                ? 'No matches'
                : ''}
            </span>

            {matchLines.length > 0 && (
              <div className="flex items-center gap-1">
                <button
                  onClick={handlePrevMatch}
                  className="p-1 rounded bg-[#222731] hover:bg-[#2c3340] text-zinc-300"
                  title="Previous match"
                >
                  <ChevronUp className="w-3 h-3" />
                </button>
                <button
                  onClick={handleNextMatch}
                  className="p-1 rounded bg-[#222731] hover:bg-[#2c3340] text-zinc-300"
                  title="Next match"
                >
                  <ChevronDown className="w-3 h-3" />
                </button>
              </div>
            )}
          </div>
        </div>
      )}

      {/* Code Editor Body */}
      <div className="flex-1 overflow-auto font-mono text-[12px] leading-[1.65] bg-[#0c0e12]">
        {lines.length > 0 ? (
          <div className="py-2 min-w-full inline-block">
            {lines.map((lineText, idx) => {
              const lineNum = idx + 1;
              const isTarget = lineNum === targetLine;
              const isSearchMatch = matchLines.includes(lineNum);
              const isCurrentMatch = matchLines[currentMatchIndex] === lineNum;
              const tokens = tokenizeLine(lineText, inferredLang);

              return (
                <div
                  key={lineNum}
                  id={`editor-line-${lineNum}`}
                  ref={isTarget ? activeLineRef : null}
                  className={`flex items-start group transition-colors duration-100 ${
                    isTarget
                      ? 'bg-dna-blue/15 border-l-2 border-dna-blue'
                      : isCurrentMatch
                      ? 'bg-amber-500/20 border-l-2 border-amber-400'
                      : isSearchMatch
                      ? 'bg-amber-500/10'
                      : 'hover:bg-[#161a21]'
                  }`}
                >
                  {/* Line Number Gutter */}
                  <div className="w-12 shrink-0 text-right pr-4 text-zinc-600 group-hover:text-zinc-400 select-none text-[11px] font-mono border-r border-charcoal-850/60 mr-3">
                    {lineNum}
                  </div>

                  {/* Highlighted Code Line */}
                  <pre
                    className={`flex-1 pr-4 font-mono ${
                      wrapLines ? 'whitespace-pre-wrap break-all' : 'whitespace-pre overflow-x-visible'
                    }`}
                  >
                    {tokens.map((token, tIdx) => (
                      <span key={tIdx} className={TOKEN_STYLES[token.type]}>
                        {token.text}
                      </span>
                    ))}
                  </pre>
                </div>
              );
            })}
          </div>
        ) : (
          <div className="h-full flex items-center justify-center text-zinc-500 text-xs py-12">
            No source content available.
          </div>
        )}
      </div>

      {/* Editor Status Bar (VS Code / JetBrains Style) */}
      <div className="h-6 border-t border-charcoal-850 bg-[#0f1217] px-3 flex items-center justify-between text-[10px] font-mono text-zinc-400 select-none">
        <div className="flex items-center gap-3">
          <span className="flex items-center gap-1 text-dna-blue">
            <Terminal className="w-3 h-3" />
            <span>Ln {targetLine}, Col 1</span>
          </span>
          <span>UTF-8</span>
          <span>LF</span>
        </div>

        <div className="flex items-center gap-3">
          <span className="capitalize">{inferredLang}</span>
          <span>Spaces: 2</span>
        </div>
      </div>
    </div>
  );
}
