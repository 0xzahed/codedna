export const API_BASE = typeof window !== 'undefined'
  ? (window.location.port === '3000' ? 'http://localhost:8080' : '')
  : 'http://localhost:8080';

export interface GraphNode {
  id: string;
  type: string;
  label: string;
  file?: string;
  line?: number;
  complexity?: number;
  exported?: boolean;
  language?: string;
  data?: Record<string, any>;
}

export interface GraphEdge {
  id: string;
  source: string;
  target: string;
  kind: string;
  confidence: number;
  method?: string;
  locations?: Array<{ file: string; line: number; col: number }>;
  isRuntimeObserved?: boolean;
}

export interface GraphData {
  nodes: GraphNode[];
  edges: GraphEdge[];
}

export interface SymbolItem {
  id: string;
  name: string;
  qualifiedName: string;
  kind: string;
  file: string;
  language: string;
  line: number;
  column: number;
  endLine: number;
  exported: boolean;
  complexity: number;
  parameterCount: number;
  lineCount: number;
  parentClassName?: string;
}

export interface FileItem {
  path: string;
  language: string;
  size: number;
  loc: number;
  tags?: string[];
}

export interface IssueItem {
  kind: string;
  severity: 'info' | 'low' | 'medium' | 'high';
  title: string;
  file?: string;
  symbolId?: string;
  detail: any;
}

export interface RepoStats {
  filesScanned: number;
  filesParsed: number;
  filesSkipped: number;
  parseErrors: number;
  symbolsFound: number;
  relationships: number;
  resolvedCalls: number;
  unresolvedCalls: number;
  loc: number;
  languages: Record<string, number>;
  locByLanguage?: Record<string, number>;
  frameworks?: string[];
  directDeps?: Record<string, string[]>;
  duration: string;
}

export interface ImpactReport {
  baseRef: string;
  targetRef: string;
  changedFiles: string[];
  changedSymbols: Array<{
    symbol: SymbolItem;
    changeType: string;
    changedLines: number[];
  }>;
  upstreamAffected: string[];
  downstreamCalls: string[];
  affectedEndpoints: string[];
  affectedTests: string[];
  riskScore: number;
  riskLevel: 'info' | 'low' | 'medium' | 'high';
  summary: string;
}

export interface ArchRule {
  id: string;
  name: string;
  fromPattern: string;
  toPattern: string;
  forbidden: boolean;
  description: string;
  severity: 'info' | 'low' | 'medium' | 'high';
}

export interface ArchReport {
  violations: Array<{
    ruleId: string;
    ruleName: string;
    severity: 'info' | 'low' | 'medium' | 'high';
    fromFile: string;
    toFile: string;
    importPath?: string;
    reason: string;
    line?: number;
  }>;
  circularCount: number;
  cycles: Array<{ files: string[]; chain: string }>;
  layerHealthScore: number;
  ruleCount: number;
  rules?: ArchRule[];
}

export interface RefactoringProposal {
  id: string;
  issueType: 'circular-dependency' | 'layer-violation';
  target: string;
  severity: 'info' | 'low' | 'medium' | 'high';
  decouplingSteps: string[];
  pattern: string;
  proposedFix: string;
  explanation: string;
}

export interface RefactorReport {
  scanId: number;
  layerHealthScore: number;
  ruleCount: number;
  circularCount: number;
  proposals: RefactoringProposal[];
  totalProposals: number;
}

export interface AuthorStats {
  name: string;
  email: string;
  commits: number;
  linesAdded: number;
  linesDeleted: number;
  filesTouched: number;
  primaryFileCount: number;
  knowledgeShare: number;
}

export interface FileOwnership {
  filePath: string;
  primaryAuthor: string;
  contributorsCount: number;
  topContributors: Array<{
    name: string;
    email: string;
    linesAdded: number;
    percentage: number;
  }>;
  busFactorRisk: 'info' | 'low' | 'medium' | 'high';
}

export interface RepoKnowledgeMap {
  isGitRepo: boolean;
  totalAuthors: number;
  totalCommits: number;
  highRiskFiles: number;
  busFactorScore: number;
  topContributors: AuthorStats[];
  files: FileOwnership[];
}

export interface RuntimeSummary {
  totalEvents: number;
  uniqueFunctions: number;
  frequentFunctions: Array<{
    function: string;
    file: string;
    callCount: number;
    totalTimeUs: number;
    avgTimeUs: number;
  }>;
  executionPaths: Array<{
    traceId: string;
    root: string;
    steps: string[];
    totalTimeUs: number;
  }>;
  avgDurationUs: number;
  comparison: {
    staticEdgeCount: number;
    runtimeEdgeCount: number;
    observedStaticEdges: number;
    unexecutedStaticEdges: number;
    dynamicOnlyEdges: number;
    coveragePercentage: number;
  };
}

export async function fetchHealth() {
  const res = await fetch(`${API_BASE}/api/health`);
  return res.json();
}

export async function fetchActive(): Promise<{ scanId: number; repoRoot: string }> {
  const res = await fetch(`${API_BASE}/api/active`);
  return res.json();
}

export async function fetchRepos() {
  const res = await fetch(`${API_BASE}/api/repos`);
  return res.json();
}

export async function triggerScan(path?: string) {
  const res = await fetch(`${API_BASE}/api/repos/scan`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ path: path || '.' }),
  });
  return res.json();
}

export async function fetchGraph(scanId: number = 1, params?: { type?: string; q?: string }): Promise<GraphData> {
  const qs = new URLSearchParams();
  if (params?.type) qs.set('type', params.type);
  if (params?.q) qs.set('q', params.q);
  const res = await fetch(`${API_BASE}/api/scans/${scanId}/graph?${qs.toString()}`);
  return res.json();
}

export async function fetchSymbols(scanId: number = 1, params?: { q?: string; kind?: string; file?: string }): Promise<SymbolItem[]> {
  const qs = new URLSearchParams();
  if (params?.q) qs.set('q', params.q);
  if (params?.kind) qs.set('kind', params.kind);
  if (params?.file) qs.set('file', params.file);
  const res = await fetch(`${API_BASE}/api/scans/${scanId}/symbols?${qs.toString()}`);
  return res.json();
}

export async function fetchFiles(scanId: number = 1): Promise<FileItem[]> {
  const res = await fetch(`${API_BASE}/api/scans/${scanId}/files`);
  return res.json();
}

export async function fetchFileContent(scanId: number = 1, path: string): Promise<{ path: string; content: string; size: number }> {
  const res = await fetch(`${API_BASE}/api/scans/${scanId}/file?path=${encodeURIComponent(path)}`);
  return res.json();
}

export async function fetchIssues(scanId: number = 1): Promise<IssueItem[]> {
  const res = await fetch(`${API_BASE}/api/scans/${scanId}/issues`);
  return res.json();
}

export async function fetchDependencies(scanId: number = 1) {
  const res = await fetch(`${API_BASE}/api/scans/${scanId}/dependencies`);
  return res.json();
}

export async function fetchExportGraph(scanId: number = 1, format: 'mermaid' | 'c4' = 'mermaid'): Promise<{ format: string; content: string; repo: string }> {
  const res = await fetch(`${API_BASE}/api/scans/${scanId}/export?format=${format}`);
  return res.json();
}

export async function fetchImpact(scanId: number = 1, base: string = 'main'): Promise<ImpactReport> {
  const res = await fetch(`${API_BASE}/api/scans/${scanId}/impact?base=${encodeURIComponent(base)}`);
  return res.json();
}

export async function fetchArchitecture(scanId: number = 1): Promise<ArchReport> {
  const res = await fetch(`${API_BASE}/api/scans/${scanId}/architecture`);
  return res.json();
}

export async function addArchRule(scanId: number = 1, rule: Partial<ArchRule>): Promise<ArchReport> {
  const res = await fetch(`${API_BASE}/api/scans/${scanId}/architecture/rules`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(rule),
  });
  if (!res.ok) {
    throw new Error('Failed to add architectural rule');
  }
  return res.json();
}

export async function deleteArchRule(scanId: number = 1, ruleId: string): Promise<ArchReport> {
  const res = await fetch(`${API_BASE}/api/scans/${scanId}/architecture/rules/${encodeURIComponent(ruleId)}`, {
    method: 'DELETE',
  });
  if (!res.ok) {
    throw new Error('Failed to delete architectural rule');
  }
  return res.json();
}

export async function resetArchRules(scanId: number = 1): Promise<ArchReport> {
  const res = await fetch(`${API_BASE}/api/scans/${scanId}/architecture/rules/reset`, {
    method: 'POST',
  });
  if (!res.ok) {
    throw new Error('Failed to reset architectural rules');
  }
  return res.json();
}

export async function fetchRefactorProposals(scanId: number = 1): Promise<RefactorReport> {
  const res = await fetch(`${API_BASE}/api/scans/${scanId}/refactor`);
  if (!res.ok) {
    throw new Error('Failed to fetch architectural refactoring proposals');
  }
  return res.json();
}

export async function fetchKnowledge(scanId: number = 1): Promise<RepoKnowledgeMap> {
  const res = await fetch(`${API_BASE}/api/scans/${scanId}/knowledge`);
  if (!res.ok) {
    throw new Error('Failed to fetch team knowledge map');
  }
  return res.json();
}

export async function fetchRuntime(scanId: number = 1): Promise<RuntimeSummary> {
  const res = await fetch(`${API_BASE}/api/scans/${scanId}/runtime`);
  return res.json();
}

export async function queryAI(req: { query: string; provider?: string; userConsent: boolean }) {
  const res = await fetch(`${API_BASE}/api/ai/query`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(req),
  });
  return res.json();
}

export async function fetchAIContext(format: string = 'markdown'): Promise<string> {
  const res = await fetch(`${API_BASE}/api/ai/context?format=${encodeURIComponent(format)}`);
  return res.text();
}

export async function fetchBusinessLogic(scanId: number = 1) {
  const res = await fetch(`${API_BASE}/api/ai/business-logic`);
  return res.json();
}

export async function fetchReport(scanId: number = 1, format: 'markdown' | 'json' = 'markdown') {
  const res = await fetch(`${API_BASE}/api/scans/${scanId}/report?format=${format}`);
  if (format === 'json') return res.json();
  return res.text();
}

export interface CrossServiceLink {
  id: string;
  sourceRepoId: number;
  sourceRepoName: string;
  sourceFile: string;
  sourceSymbol?: string;
  targetRepoId: number;
  targetRepoName: string;
  targetFile: string;
  targetSymbol?: string;
  httpMethod: string;
  routePath: string;
  confidence: number;
  description: string;
}

export interface WorkspaceItem {
  id: number;
  name: string;
  repoIds: number[];
  repos?: any[];
  links?: CrossServiceLink[];
  createdAt: string;
}

export async function fetchWorkspaces(): Promise<WorkspaceItem[]> {
  const res = await fetch(`${API_BASE}/api/workspaces`);
  return res.json();
}

export async function createWorkspace(name: string, repoIds: number[]): Promise<WorkspaceItem> {
  const res = await fetch(`${API_BASE}/api/workspaces`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ name, repoIds }),
  });
  return res.json();
}

export async function fetchWorkspace(id: number): Promise<WorkspaceItem> {
  const res = await fetch(`${API_BASE}/api/workspaces/${id}`);
  return res.json();
}

export async function fetchWorkspaceGraph(id: number): Promise<GraphData> {
  const res = await fetch(`${API_BASE}/api/workspaces/${id}/graph`);
  return res.json();
}

export type TierKind =
  | 'client_gateway'
  | 'api_presentation'
  | 'domain_service'
  | 'persistence_storage'
  | 'async_messaging'
  | 'external_integration';

export interface SystemTier {
  id: TierKind;
  name: string;
  description: string;
  order: number;
  fileCount: number;
  symbolCount: number;
  files?: string[];
  keySymbols?: string[];
}

export interface DataFlowHop {
  fromTier: TierKind;
  toTier: TierKind;
  protocol: string;
  callCount: number;
  sampleCalls?: string[];
}

export type PatternCategory = 'Architectural' | 'Structural' | 'Creational' | 'Behavioral';

export interface DetectedPattern {
  id: string;
  name: string;
  category: PatternCategory;
  confidence: number;
  description: string;
  evidence: string[];
  files: string[];
  tradeoffs: string;
}

export interface SystemDesignReport {
  architectureStyle: string;
  summary: string;
  tiers: SystemTier[];
  dataFlow: DataFlowHop[];
  detectedPatterns: DetectedPattern[];
  keyEntrypoints: string[];
  bottleneckHotspots: string[];
  reliabilityScore: number;
}

export async function fetchSystemDesign(scanId: number = 1): Promise<SystemDesignReport> {
  const res = await fetch(`${API_BASE}/api/scans/${scanId}/system-design`);
  if (!res.ok) {
    throw new Error(`Failed to fetch system design report: ${res.statusText}`);
  }
  return res.json();
}

export async function fetchPatterns(scanId: number = 1): Promise<{ scanId: number; patterns: DetectedPattern[]; total: number }> {
  const res = await fetch(`${API_BASE}/api/scans/${scanId}/patterns`);
  if (!res.ok) {
    throw new Error(`Failed to fetch patterns: ${res.statusText}`);
  }
  return res.json();
}
