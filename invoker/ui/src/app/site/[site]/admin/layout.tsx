import type { Metadata } from 'next';
import { Inter } from 'next/font/google';
import React from 'react';
import Container from './container';

const inter = Inter({ subsets: ['latin'] });

export const metadata: Metadata = {
  title: 'Invoker Site Admin',
  description: 'Invoker Site Admin',
};

export default function SiteAdminLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return <Container>{children}</Container>;
}
