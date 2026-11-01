'use client';

import dynamic from 'next/dynamic';
import AppLayout from '../../components/AppLayout';

const GraphView = dynamic(() => import('../../components/views/GraphView'), {
  ssr: false,
  loading: () => (
    <div className="h-full flex items-center justify-center text-zinc-500 font-mono text-xs">
      Loading interactive DNA canvas...
    </div>
  ),
});

export default function GraphPage() {
  return (
    <AppLayout>
      <GraphView />
    </AppLayout>
  );
}
