'use client';

import { ReactNode, useEffect, useState } from 'react';
import { AuthProvider } from '@/contexts/auth-context';

/**
 * Wrapper that ensures AuthProvider only renders on client side
 * Prevents SSR hydration issues with Firebase Auth
 */
export function ClientOnlyAuthProvider({ children }: { children: ReactNode }) {
  const [isMounted, setIsMounted] = useState(false);

  useEffect(() => {
    setIsMounted(true);
  }, []);

  // During SSR or initial render, show loading
  // We can't render children without AuthProvider because they may use useAuth()
  if (!isMounted) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-50">
        <div className="text-center">
          <div className="w-16 h-16 border-4 border-blue-600 border-t-transparent rounded-full animate-spin mx-auto mb-4"></div>
          <p className="text-gray-600">Inicializando...</p>
        </div>
      </div>
    );
  }

  // Only render AuthProvider with children after client-side mount
  return <AuthProvider>{children}</AuthProvider>;
}
