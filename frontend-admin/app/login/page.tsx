'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import { signInWithCustomToken } from 'firebase/auth';
import { auth } from '@/lib/firebase';
import { Home } from 'lucide-react';
import { loginSchema } from '@/lib/validations';

const API_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api/v1';

export default function LoginPage() {
  const router = useRouter();
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setLoading(true);

    try {
      // Validate form data with Zod
      const validatedData = loginSchema.parse({ email, password });

      // Call backend API login endpoint
      const response = await fetch(`${API_URL}/auth/login`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(validatedData),
      });

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.error || 'Erro ao fazer login');
      }

      const data = await response.json();

      // Sign in with custom token from backend
      await signInWithCustomToken(auth, data.firebase_token);

      // Store tenant info in localStorage
      localStorage.setItem('tenant_id', data.tenant_id);
      localStorage.setItem('broker_id', data.broker.id);
      localStorage.setItem('broker_role', data.broker.role);
      localStorage.setItem('broker_name', data.broker.name);
      localStorage.setItem('is_platform_admin', data.is_platform_admin ? 'true' : 'false');

      router.push('/dashboard');
    } catch (err: any) {
      if (err.errors) {
        // Zod validation error
        setError(err.errors[0]?.message || 'Dados inválidos');
      } else {
        setError(err.message || 'Email ou senha inválidos. Tente novamente.');
      }
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen bg-gradient-to-br from-blue-600 to-blue-800 flex items-center justify-center p-4">
      <div className="w-full max-w-md">
        {/* Logo */}
        <div className="text-center mb-8">
          <div className="inline-flex items-center justify-center w-16 h-16 bg-white rounded-full mb-4">
            <Home className="w-8 h-8 text-blue-600" />
          </div>
          <h1 className="text-3xl font-bold text-white mb-2">
            Admin Imobiliária
          </h1>
          <p className="text-blue-100">
            Acesse o painel administrativo
          </p>
        </div>

        {/* Login Form */}
        <div className="bg-white rounded-lg shadow-xl p-8">
          <form onSubmit={handleSubmit} className="space-y-6">
            {error && (
              <div className="p-3 bg-red-50 border border-red-200 rounded-lg">
                <p className="text-sm text-red-600">{error}</p>
              </div>
            )}

            <div>
              <label htmlFor="email" className="block text-sm font-medium text-gray-700 mb-1">
                Email
              </label>
              <input
                id="email"
                type="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                placeholder="seu@email.com"
                required
              />
            </div>

            <div>
              <label htmlFor="password" className="block text-sm font-medium text-gray-700 mb-1">
                Senha
              </label>
              <input
                id="password"
                type="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                placeholder="••••••••"
                required
              />
            </div>

            <button
              type="submit"
              disabled={loading}
              className="w-full bg-blue-600 hover:bg-blue-700 text-white font-medium py-2 px-4 rounded-lg transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {loading ? 'Entrando...' : 'Entrar'}
            </button>
          </form>

          <div className="mt-6 space-y-3">
            <div className="text-center">
              <p className="text-sm text-gray-600">
                Esqueceu sua senha?{' '}
                <a href="#" className="text-blue-600 hover:underline">
                  Recuperar senha
                </a>
              </p>
            </div>

            <div className="relative">
              <div className="absolute inset-0 flex items-center">
                <div className="w-full border-t border-gray-300"></div>
              </div>
              <div className="relative flex justify-center text-sm">
                <span className="px-2 bg-white text-gray-500">ou</span>
              </div>
            </div>

            <div className="text-center">
              <p className="text-sm text-gray-600">
                Ainda não tem uma conta?{' '}
                <a href="/signup" className="text-blue-600 hover:underline font-medium">
                  Cadastre-se
                </a>
              </p>
            </div>
          </div>
        </div>

        <p className="text-center text-sm text-blue-100 mt-8">
          &copy; 2025 Ecosistema Imobiliário. Todos os direitos reservados.
        </p>
      </div>
    </div>
  );
}
