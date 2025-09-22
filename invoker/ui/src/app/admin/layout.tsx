import type { Metadata } from 'next';
import { Inter } from 'next/font/google';
import '@fontsource/roboto/300.css';
import '@fontsource/roboto/400.css';
import '@fontsource/roboto/500.css';
import '@fontsource/roboto/700.css';
import React from 'react';
import Container from './container';

const inter = Inter({ subsets: ['latin'] });

export const metadata: Metadata = {
  title: 'Invoker Admin',
  description: 'Invoker Admin',
};

export default function AdminLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <main>
      <Container>
        <main>{children}</main>
      </Container>
    </main>
  );
}
