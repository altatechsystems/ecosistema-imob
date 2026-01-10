'use client';

import React, { createContext, useContext, useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import {
  User as FirebaseUser,
  onAuthStateChanged,
  signInWithCustomToken,
  signOut as firebaseSignOut,
} from 'firebase/auth';
import { auth } from '@/lib/firebase';
import { User } from '@/types/user';
import { Tenant } from '@/types/tenant';
import { api } from '@/lib/api';

// Production API URL is set via NEXT_PUBLIC_API_URL environment variable
const API_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api/v1';

interface AuthContextValue {
  // Firebase auth state
  user: FirebaseUser | null;
  loading: boolean;

  // Enhanced user profile with tenant info
  userProfile: User | null;

  // Current tenant
  tenantId: string | null;

  // Authentication actions
  login: (email: string, password: string) => Promise<void>;
  logout: () => Promise<void>;

  // Tenant management (for future multi-tenant users)
  availableTenants: Tenant[];
  switchTenant: (tenantId: string) => Promise<void>;
}

const AuthContext = createContext<AuthContextValue | undefined>(undefined);

export function AuthProvider({ children }: { children: React.ReactNode }) {
  console.log('[AuthProvider] Initializing...');

  let router;
  try {
    router = useRouter();
    console.log('[AuthProvider] useRouter initialized');
  } catch (error) {
    console.error('[AuthProvider] ❌ useRouter failed:', error);
    throw error;
  }

  // Firebase user state
  const [user, setUser] = useState<FirebaseUser | null>(null);
  const [loading, setLoading] = useState(true);

  // Enhanced user profile with tenant info
  const [userProfile, setUserProfile] = useState<User | null>(null);
  const [tenantId, setTenantId] = useState<string | null>(null);

  // Tenant management (for future)
  const [availableTenants, setAvailableTenants] = useState<Tenant[]>([]);

  console.log('[AuthProvider] State initialized', { loading, user: !!user, tenantId });

  // Extract tenant_id and user profile from Firebase custom claims
  const extractUserInfoFromToken = async (firebaseUser: FirebaseUser) => {
    console.log('[AuthProvider] Extracting user info from token...', { uid: firebaseUser.uid });

    try {
      const tokenResult = await firebaseUser.getIdTokenResult();
      const claims = tokenResult.claims;
      console.log('[AuthProvider] Token claims:', claims);

      // Extract tenant_id from custom claims
      const extractedTenantId = claims.tenant_id as string | undefined;
      console.log('[AuthProvider] Extracted tenant_id:', extractedTenantId);

      if (extractedTenantId) {
        setTenantId(extractedTenantId);

        // Update API client with tenant
        api.setTenant(extractedTenantId);
        console.log('[AuthProvider] API client tenant set to:', extractedTenantId);

        // Store in sessionStorage as backup for UX (only in browser)
        if (typeof window !== 'undefined') {
          sessionStorage.setItem('tenant_id', extractedTenantId);
          console.log('[AuthProvider] Stored tenant_id in sessionStorage');
        }

        // Build user profile from claims and localStorage (legacy)
        // In the future, we should fetch full profile from backend
        const brokerName = typeof window !== 'undefined'
          ? localStorage.getItem('broker_name')
          : null;

        const profile: User = {
          id: (claims.user_id || claims.broker_id) as string,
          tenant_id: extractedTenantId,
          firebase_uid: firebaseUser.uid,
          name: brokerName || firebaseUser.displayName || 'User',
          email: firebaseUser.email || '',
          role: (claims.role as 'admin' | 'manager') || 'manager',
          is_active: true,
          created_at: new Date(),
          updated_at: new Date(),
        };

        setUserProfile(profile);
        console.log('[AuthProvider] User profile set:', profile);
      } else {
        console.error('[AuthProvider] ❌ No tenant_id found in custom claims');
        // Clear state if no tenant
        setTenantId(null);
        setUserProfile(null);
      }
    } catch (error) {
      console.error('[AuthProvider] ❌ Error extracting user info from token:', error);
      setTenantId(null);
      setUserProfile(null);
    }
  };

  // Listen to Firebase auth state changes
  useEffect(() => {
    console.log('[AuthProvider] useEffect running...');

    // Only run on client side
    if (typeof window === 'undefined') {
      console.log('[AuthProvider] ⚠️ Not on client side, skipping auth setup');
      setLoading(false);
      return;
    }

    if (!auth) {
      console.log('[AuthProvider] ⚠️ Firebase Auth not initialized');
      setLoading(false);
      return;
    }

    console.log('[AuthProvider] ✅ Setting up Firebase Auth listener...');

    const unsubscribe = onAuthStateChanged(auth, async (firebaseUser) => {
      console.log('[AuthProvider] Auth state changed:', firebaseUser ? `User: ${firebaseUser.email}` : 'No user');
      setUser(firebaseUser);

      if (firebaseUser) {
        // Extract tenant and user info from custom claims
        await extractUserInfoFromToken(firebaseUser);
      } else {
        console.log('[AuthProvider] Clearing auth state...');
        // Clear state on logout
        setUserProfile(null);
        setTenantId(null);
        setAvailableTenants([]);
        api.setTenant(''); // Clear API client tenant
        sessionStorage.removeItem('tenant_id');
      }

      setLoading(false);
      console.log('[AuthProvider] Loading complete');
    });

    return () => {
      console.log('[AuthProvider] Cleaning up auth listener');
      unsubscribe();
    };
  }, []);

  // Login function
  const login = async (email: string, password: string) => {
    console.log('[AuthProvider] Login attempt:', email);
    console.log('[AuthProvider] API_URL:', API_URL);
    console.log('[AuthProvider] NEXT_PUBLIC_API_URL:', process.env.NEXT_PUBLIC_API_URL);

    // Ensure we're on client side
    if (typeof window === 'undefined') {
      console.error('[AuthProvider] ❌ Login called on server side');
      throw new Error('Login can only be called on client side');
    }

    if (!auth) {
      console.error('[AuthProvider] ❌ Firebase Auth not initialized');
      throw new Error('Firebase Auth not initialized');
    }

    try {
      setLoading(true);
      console.log('[AuthProvider] Calling backend login API...');

      // Call backend API login endpoint
      const response = await fetch(`${API_URL}/auth/login`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ email, password }),
      });

      console.log('[AuthProvider] Backend response status:', response.status);

      if (!response.ok) {
        const errorData = await response.json();
        console.error('[AuthProvider] ❌ Backend login failed:', errorData);
        throw new Error(errorData.error || 'Erro ao fazer login');
      }

      const data = await response.json();
      console.log('[AuthProvider] ✅ Backend login successful:', { tenant_id: data.tenant_id, broker: data.broker });

      // Sign in with custom token from backend
      // This will trigger onAuthStateChanged above
      console.log('[AuthProvider] Signing in with custom token...');
      await signInWithCustomToken(auth, data.firebase_token);
      console.log('[AuthProvider] ✅ Firebase sign-in successful');

      // Store additional data in localStorage (legacy support)
      // TODO: Remove once we fetch full profile from backend
      localStorage.setItem('broker_id', data.broker.id);
      localStorage.setItem('broker_role', data.broker.role);
      localStorage.setItem('broker_name', data.broker.name);
      localStorage.setItem('is_platform_admin', data.is_platform_admin ? 'true' : 'false');
      console.log('[AuthProvider] Stored broker data in localStorage');

      // Redirect to dashboard
      console.log('[AuthProvider] Redirecting to dashboard...');
      router.push('/dashboard');
    } catch (error: any) {
      console.error('[AuthProvider] ❌ Login error:', error);
      throw error;
    } finally {
      setLoading(false);
    }
  };

  // Logout function
  const logout = async () => {
    // Ensure we're on client side
    if (typeof window === 'undefined') {
      throw new Error('Logout can only be called on client side');
    }

    if (!auth) {
      throw new Error('Firebase Auth not initialized');
    }

    try {
      await firebaseSignOut(auth);

      // Clear localStorage
      localStorage.removeItem('broker_id');
      localStorage.removeItem('broker_role');
      localStorage.removeItem('broker_name');
      localStorage.removeItem('is_platform_admin');
      localStorage.removeItem('tenant_id'); // legacy

      // Clear sessionStorage
      sessionStorage.removeItem('tenant_id');

      // State is cleared by onAuthStateChanged listener

      // Redirect to login
      router.push('/login');
    } catch (error) {
      console.error('Logout error:', error);
      throw error;
    }
  };

  // Switch tenant (for future multi-tenant users)
  const switchTenant = async (newTenantId: string) => {
    // TODO: Implement when backend supports multi-tenant users
    // This would:
    // 1. Call POST /api/v1/auth/switch-tenant with newTenantId
    // 2. Get new custom token with updated tenant_id claim
    // 3. Sign in with new token
    // 4. Update local state

    console.warn('switchTenant not implemented yet');
    throw new Error('Switching tenants is not yet supported');
  };

  const value: AuthContextValue = {
    user,
    loading,
    userProfile,
    tenantId,
    login,
    logout,
    availableTenants,
    switchTenant,
  };

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth(): AuthContextValue {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error('useAuth must be used within AuthProvider');
  }
  return context;
}
