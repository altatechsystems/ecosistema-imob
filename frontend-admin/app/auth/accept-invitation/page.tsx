'use client';

import { useEffect, useState } from 'react';
import { useRouter, useSearchParams } from 'next/navigation';
import { signInWithCustomToken } from 'firebase/auth';
import { auth } from '@/lib/firebase';
import { logger } from '@/lib/logger';

interface InvitationData {
  id: string;
  tenant_id: string;
  email: string;
  name: string;
  phone?: string;
  role: string;
  creci?: string;
  status: string;
  expires_at: string;
  created_at: string;
}

export default function AcceptInvitationPage() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const token = searchParams.get('token');

  const [loading, setLoading] = useState(true);
  const [verifying, setVerifying] = useState(true);
  const [accepting, setAccepting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [invitation, setInvitation] = useState<InvitationData | null>(null);
  const [password, setPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');

  // Verify invitation token on mount
  useEffect(() => {
    const verifyInvitation = async () => {
      if (!token) {
        setError('Token de convite não encontrado na URL');
        setVerifying(false);
        setLoading(false);
        return;
      }

      try {
        const apiUrl = process.env.NEXT_PUBLIC_API_URL;
        logger.dev('Verifying invitation token:', token);

        const response = await fetch(`${apiUrl}/invitations/${token}/verify`);
        const data = await response.json();

        logger.dev('Verification response:', data);

        if (!response.ok || !data.valid) {
          setError(data.message || 'Convite inválido ou expirado');
          setVerifying(false);
          setLoading(false);
          return;
        }

        setInvitation(data.invitation);
        setVerifying(false);
        setLoading(false);
      } catch (err) {
        logger.error('Error verifying invitation:', err);
        setError('Erro ao verificar convite. Tente novamente.');
        setVerifying(false);
        setLoading(false);
      }
    };

    verifyInvitation();
  }, [token]);

  const handleAcceptInvitation = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);

    // Validate password
    if (password.length < 8) {
      setError('A senha deve ter no mínimo 8 caracteres');
      return;
    }

    if (password !== confirmPassword) {
      setError('As senhas não coincidem');
      return;
    }

    setAccepting(true);

    try {
      const apiUrl = process.env.NEXT_PUBLIC_API_URL;
      logger.dev('Accepting invitation with token:', token);

      const response = await fetch(`${apiUrl}/invitations/${token}/accept`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ password }),
      });

      const data = await response.json();
      logger.dev('Accept invitation response:', data);

      if (!response.ok) {
        throw new Error(data.error || 'Erro ao aceitar convite');
      }

      // Sign in with custom token if provided
      if (data.firebase_token) {
        logger.dev('Signing in with custom token');
        await signInWithCustomToken(auth, data.firebase_token);
        logger.dev('Signed in successfully');

        // Store tenant_id in localStorage for auth context
        localStorage.setItem('tenant_id', data.tenant_id);

        // Redirect to dashboard
        router.push('/dashboard');
      } else {
        // No token, redirect to login
        router.push('/auth/login?message=Conta criada com sucesso! Faça login para continuar.');
      }
    } catch (err: any) {
      logger.error('Error accepting invitation:', err);
      setError(err.message || 'Erro ao aceitar convite. Tente novamente.');
      setAccepting(false);
    }
  };

  const getRoleLabel = (role: string) => {
    const roleLabels: Record<string, string> = {
      admin: 'Administrador',
      manager: 'Gerente',
      broker: 'Corretor',
      broker_admin: 'Corretor Administrador',
    };
    return roleLabels[role] || role;
  };

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-purple-50 to-blue-50">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-purple-600 mx-auto mb-4"></div>
          <p className="text-gray-600">Verificando convite...</p>
        </div>
      </div>
    );
  }

  if (error && !invitation) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-purple-50 to-blue-50 p-4">
        <div className="max-w-md w-full bg-white rounded-lg shadow-xl p-8 text-center">
          <div className="w-16 h-16 bg-red-100 rounded-full flex items-center justify-center mx-auto mb-4">
            <svg
              className="w-8 h-8 text-red-600"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M6 18L18 6M6 6l12 12"
              />
            </svg>
          </div>
          <h1 className="text-2xl font-bold text-gray-900 mb-2">Convite Inválido</h1>
          <p className="text-gray-600 mb-6">{error}</p>
          <button
            onClick={() => router.push('/auth/login')}
            className="px-6 py-2 bg-purple-600 text-white rounded-lg hover:bg-purple-700 transition-colors"
          >
            Ir para Login
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-purple-50 to-blue-50 p-4">
      <div className="max-w-md w-full bg-white rounded-lg shadow-xl p-8">
        <div className="text-center mb-8">
          <div className="w-16 h-16 bg-purple-100 rounded-full flex items-center justify-center mx-auto mb-4">
            <svg
              className="w-8 h-8 text-purple-600"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z"
              />
            </svg>
          </div>
          <h1 className="text-2xl font-bold text-gray-900 mb-2">
            Aceitar Convite
          </h1>
          <p className="text-gray-600">
            Você foi convidado para fazer parte da equipe
          </p>
        </div>

        {invitation && (
          <div className="bg-purple-50 border border-purple-200 rounded-lg p-4 mb-6">
            <div className="space-y-2 text-sm">
              <div className="flex justify-between">
                <span className="text-gray-600">Nome:</span>
                <span className="font-medium text-gray-900">{invitation.name}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-gray-600">Email:</span>
                <span className="font-medium text-gray-900">{invitation.email}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-gray-600">Função:</span>
                <span className="font-medium text-gray-900">
                  {getRoleLabel(invitation.role)}
                </span>
              </div>
              {invitation.creci && (
                <div className="flex justify-between">
                  <span className="text-gray-600">CRECI:</span>
                  <span className="font-medium text-gray-900">{invitation.creci}</span>
                </div>
              )}
            </div>
          </div>
        )}

        <form onSubmit={handleAcceptInvitation} className="space-y-4">
          <div>
            <label htmlFor="password" className="block text-sm font-medium text-gray-700 mb-1">
              Senha
            </label>
            <input
              type="password"
              id="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              required
              minLength={8}
              className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-purple-500 focus:border-transparent"
              placeholder="Mínimo 8 caracteres"
              disabled={accepting}
            />
          </div>

          <div>
            <label htmlFor="confirmPassword" className="block text-sm font-medium text-gray-700 mb-1">
              Confirmar Senha
            </label>
            <input
              type="password"
              id="confirmPassword"
              value={confirmPassword}
              onChange={(e) => setConfirmPassword(e.target.value)}
              required
              minLength={8}
              className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-purple-500 focus:border-transparent"
              placeholder="Digite a senha novamente"
              disabled={accepting}
            />
          </div>

          {error && (
            <div className="bg-red-50 border border-red-200 rounded-lg p-3 text-sm text-red-800">
              {error}
            </div>
          )}

          <button
            type="submit"
            disabled={accepting}
            className="w-full py-3 bg-gradient-to-r from-purple-600 to-blue-600 text-white rounded-lg font-medium hover:from-purple-700 hover:to-blue-700 transition-all disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center"
          >
            {accepting ? (
              <>
                <svg className="animate-spin -ml-1 mr-3 h-5 w-5 text-white" fill="none" viewBox="0 0 24 24">
                  <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                  <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                </svg>
                Criando conta...
              </>
            ) : (
              'Aceitar Convite e Criar Conta'
            )}
          </button>
        </form>

        <div className="mt-6 text-center text-sm text-gray-600">
          <p>
            Já tem uma conta?{' '}
            <a href="/auth/login" className="text-purple-600 hover:text-purple-700 font-medium">
              Faça login
            </a>
          </p>
        </div>
      </div>
    </div>
  );
}
