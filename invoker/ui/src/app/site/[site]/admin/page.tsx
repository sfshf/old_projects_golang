'use client';
import React from 'react';
import { useRouter, useParams } from 'next/navigation';

export default function Page() {
  const router = useRouter();
  const params = useParams();
  React.useEffect(() => {
    if (params.site && typeof params.site === 'string') {
      router.push('/site/' + params.site + '/admin/category');
    }
  }, []);

  return <></>;
}
