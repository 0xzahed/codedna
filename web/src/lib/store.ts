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
  theme: 'dark';
}

export const useAppStore = create<AppState>((set) => ({
  activeTab: 'overview',
  setActiveTab: (tab) => set({ activeTab: tab }),
  scanId: 1,
  setScanId: (id) => set({ scanId: id }),
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
  theme: 'dark',
}));
