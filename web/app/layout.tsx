import type { Metadata, Viewport } from 'next';
import ServiceWorkerRegistrar from '../components/ServiceWorkerRegistrar';
import './globals.css';

export const metadata: Metadata = {
  title: 'missless — Virtual Reunion',
  description: 'Meet the person you miss through AI-powered virtual reunion experience',
  manifest: '/manifest.json',
  appleWebApp: {
    capable: true,
    statusBarStyle: 'black-translucent',
    title: 'missless',
  },
  openGraph: {
    title: 'missless — Virtual Reunion',
    description: 'Meet the person you miss through AI-powered virtual reunion experience',
    url: 'https://missless.co',
    siteName: 'missless',
    type: 'website',
  },
  icons: {
    icon: '/icons/icon-192.png',
    apple: '/icons/apple-touch-icon.png',
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
      <body>
        {children}
        <ServiceWorkerRegistrar />
      </body>
    </html>
  );
}
