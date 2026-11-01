import { create } from 'zustand';
import { GraphNode, RepoStats } from './api';

export type NavTab = 
  | 'overview'
  | 'explorer'
  | 'graph'
  | 'deps'
  | 'runtime'
  | 'impact'
  | 'arch'
  | 'ai'
  | 'settings';

interface AppState {
  activeTab: NavTab;
  setActiveTab: (tab: NavTab) => void;
  scanId: number;
  setScanId: (id: number) => void;
  activeRepoName: string;
  setActiveRepoName: (name: string) => void;
  workspaceId: number | null;
  setWorkspaceId: (id: number | null) => void;
  isScanning: boolean;
  setIsScanning: (scanning: boolean) => void;
  repoStats: RepoStats | null;
  setRepoStats: (stats: RepoStats | null) => void;
  selectedNode: GraphNode | null;
  setSelectedNode: (node: GraphNode | null) => void;
  selectedFile: string | null;
  setSelectedFile: (file: string | null) => void;
  fileContent: string | null;
  setFileContent: (content: string | null) => void;
  refreshKey: number;
  triggerRefresh: () => void;
  lastLiveUpdate: { message: string; timestamp: number } | null;
  setLastLiveUpdate: (update: { message: string; timestamp: number } | null) => void;
  theme: 'dark';
}

export const useAppStore = create<AppState>((set) => ({
  activeTab: 'overview',
  setActiveTab: (tab) => set({ activeTab: tab }),
  scanId: typeof window !== 'undefined' ? Number(localStorage.getItem('codedna_scan_id')) || 1 : 1,
  setScanId: (id) => {
    if (typeof window !== 'undefined') {
      localStorage.setItem('codedna_scan_id', id.toString());
    }
    set({ scanId: id });
  },
  activeRepoName: typeof window !== 'undefined' ? localStorage.getItem('codedna_repo_name') || 'codedna' : 'codedna',
  setActiveRepoName: (name) => {
    if (typeof window !== 'undefined') {
      localStorage.setItem('codedna_repo_name', name);
    }
    set({ activeRepoName: name });
  },
  workspaceId: null,
  setWorkspaceId: (id) => set({ workspaceId: id }),
  isScanning: false,
  setIsScanning: (scanning) => set({ isScanning: scanning }),
  repoStats: null,
  setRepoStats: (stats) => set({ repoStats: stats }),
  selectedNode: null,
  setSelectedNode: (node) => set({ selectedNode: node }),
  selectedFile: null,
  setSelectedFile: (file) => set({ selectedFile: file }),
  fileContent: null,
  setFileContent: (content) => set({ fileContent: content }),
  refreshKey: 0,
  triggerRefresh: () => set((state) => ({ refreshKey: state.refreshKey + 1 })),
  lastLiveUpdate: null,
  setLastLiveUpdate: (update) => set({ lastLiveUpdate: update }),
  theme: 'dark',
}));
