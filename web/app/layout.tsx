import type { Metadata, Viewport } from 'next';
import './globals.css';

export const metadata: Metadata = {
  title: 'missless — Virtual Reunion',
  description: 'Meet the person you miss through AI-powered virtual reunion experience',
  openGraph: {
    title: 'missless — Virtual Reunion',
    description: 'Meet the person you miss through AI-powered virtual reunion experience',
    url: 'https://missless.co',
    siteName: 'missless',
    type: 'website',
  },
};

export const viewport: Viewport = {
  width: 'device-width',
  initialScale: 1,
  maximumScale: 1,
  themeColor: '#1a1a2e',
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="en">
      <body>{children}</body>
    </html>
  );
}
