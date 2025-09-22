'use client';
import React, { useEffect } from 'react';
import { useRouter } from 'next/navigation';

export default function Home() {
  const router = useRouter();
  useEffect(() => {
    router.push('/site', { scroll: false });
  }, []);

  return <></>;
}
