import './globals.css';
import type { Metadata } from 'next';

export const metadata: Metadata = {
  title: 'CodeDNA 2.0 | A Living Intelligence Map of Software',
  description: 'Understand Your Code. See Its DNA.',
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="en" className="dark">
      <body className="bg-charcoal-950 text-zinc-100 min-h-screen flex flex-col antialiased">
        {children}
      </body>
    </html>
  );
}
