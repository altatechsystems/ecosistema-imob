'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import axios from 'axios';
import { signInWithCustomToken } from 'firebase/auth';
import { auth } from '@/lib/firebase';
import { Building2, Loader2, Eye, EyeOff } from 'lucide-react';
import { signupSchema } from '@/lib/validations';

interface SignupFormProps {
  onSuccess?: () => void;
  redirectTo?: string;
  variant?: 'standalone' | 'embedded';
}

export function SignupForm({
  onSuccess,
  redirectTo = '/dashboard',
  variant = 'standalone'
}: SignupFormProps) {
  const router = useRouter();
  const [formData, setFormData] = useState({
    tenant_name: '',
    name: '',
    email: '',
    phone: '',
    password: '',
    confirmPassword: '',
  });
  const [showPassword, setShowPassword] = useState(false);
  const [showConfirmPassword, setShowConfirmPassword] = useState(false);
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setFormData({
      ...formData,
      [e.target.name]: e.target.value,
    });
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');

    setLoading(true);

    try {
      // Validate password confirmation
      if (formData.password !== formData.confirmPassword) {
        setError('As senhas não coincidem');
        setLoading(false);
        return;
      }

      // Validate with Zod
      const validatedData = signupSchema.parse({
        name: formData.name,
        email: formData.email,
        password: formData.password,
        phone: formData.phone,
        tenant_name: formData.tenant_name,
      });
      // 1. Criar tenant e usuário no backend
      const signupResponse = await axios.post(
        `${process.env.NEXT_PUBLIC_API_URL}/api/v1/auth/signup`,
        validatedData
      );

      const data = signupResponse.data;

      // 2. Sign in with custom token from backend
      await signInWithCustomToken(auth, data.firebase_token);

      // 3. Store tenant info in localStorage
      localStorage.setItem('tenant_id', data.tenant_id);
      localStorage.setItem('broker_id', data.broker_id);
      localStorage.setItem('broker_role', data.user.role);
      localStorage.setItem('broker_name', data.user.name);

      // 4. Callback de sucesso (se fornecido)
      if (onSuccess) {
        onSuccess();
      }

      // 5. Redirecionar
      router.push(redirectTo);
    } catch (err: any) {
      if (err.errors) {
        // Zod validation error
        setError(err.errors[0]?.message || 'Dados inválidos');
      } else if (err.response?.data?.error) {
        setError(err.response.data.error);
      } else if (err.response?.status === 409) {
        setError('Email já cadastrado. Faça login ou use outro email.');
      } else if (err.response?.status === 400) {
        setError('Dados inválidos. Verifique os campos e tente novamente.');
      } else {
        setError('Erro ao criar conta. Tente novamente.');
      }
    } finally {
      setLoading(false);
    }
  };

  const containerClasses = variant === 'standalone'
    ? 'min-h-screen bg-gradient-to-br from-blue-600 to-blue-800 flex items-center justify-center p-4'
    : '';

  const cardClasses = variant === 'standalone'
    ? 'bg-white rounded-2xl shadow-2xl p-8 w-full max-w-md'
    : 'w-full';

  return (
    <div className={containerClasses}>
      <div className={cardClasses}>
        {variant === 'standalone' && (
          <div className="text-center mb-8">
            <div className="inline-flex items-center justify-center w-16 h-16 bg-blue-100 rounded-full mb-4">
              <Building2 className="w-8 h-8 text-blue-600" />
            </div>
            <h1 className="text-2xl font-bold text-gray-900 mb-2">
              Cadastre-se
            </h1>
            <p className="text-gray-600">
              Comece a gerenciar seus imóveis hoje mesmo
            </p>
          </div>
        )}

        {error && (
          <div className="mb-6 p-4 bg-red-50 border border-red-200 rounded-lg">
            <p className="text-sm text-red-600">{error}</p>
          </div>
        )}

        <form onSubmit={handleSubmit} className="space-y-4">
          {/* Nome da Imobiliária */}
          <div>
            <label htmlFor="tenant_name" className="block text-sm font-medium text-gray-700 mb-1">
              Nome da Imobiliária *
            </label>
            <input
              id="tenant_name"
              name="tenant_name"
              type="text"
              required
              value={formData.tenant_name}
              onChange={handleChange}
              disabled={loading}
              placeholder="Ex: Imobiliária XYZ"
              className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent disabled:bg-gray-100 disabled:cursor-not-allowed"
            />
          </div>

          {/* Nome do Usuário */}
          <div>
            <label htmlFor="name" className="block text-sm font-medium text-gray-700 mb-1">
              Seu Nome Completo *
            </label>
            <input
              id="name"
              name="name"
              type="text"
              required
              value={formData.name}
              onChange={handleChange}
              disabled={loading}
              placeholder="João Silva"
              className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent disabled:bg-gray-100 disabled:cursor-not-allowed"
            />
          </div>

          {/* Email */}
          <div>
            <label htmlFor="email" className="block text-sm font-medium text-gray-700 mb-1">
              Email *
            </label>
            <input
              id="email"
              name="email"
              type="email"
              required
              value={formData.email}
              onChange={handleChange}
              disabled={loading}
              placeholder="seu@email.com"
              className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent disabled:bg-gray-100 disabled:cursor-not-allowed"
            />
          </div>

          {/* Telefone */}
          <div>
            <label htmlFor="phone" className="block text-sm font-medium text-gray-700 mb-1">
              Telefone (WhatsApp) *
            </label>
            <input
              id="phone"
              name="phone"
              type="tel"
              required
              value={formData.phone}
              onChange={handleChange}
              disabled={loading}
              placeholder="+5511999999999"
              className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent disabled:bg-gray-100 disabled:cursor-not-allowed"
            />
            <p className="mt-1 text-xs text-gray-500">
              Formato internacional: +55 (código do país) + DDD + número
            </p>
          </div>

          {/* Senha */}
          <div>
            <label htmlFor="password" className="block text-sm font-medium text-gray-700 mb-1">
              Senha *
            </label>
            <div className="relative">
              <input
                id="password"
                name="password"
                type={showPassword ? 'text' : 'password'}
                required
                value={formData.password}
                onChange={handleChange}
                disabled={loading}
                placeholder="Mínimo 6 caracteres"
                className="w-full px-4 py-2 pr-10 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent disabled:bg-gray-100 disabled:cursor-not-allowed"
              />
              <button
                type="button"
                onClick={() => setShowPassword(!showPassword)}
                className="absolute right-3 top-1/2 -translate-y-1/2 text-gray-500 hover:text-gray-700"
              >
                {showPassword ? <EyeOff className="w-5 h-5" /> : <Eye className="w-5 h-5" />}
              </button>
            </div>
          </div>

          {/* Confirmar Senha */}
          <div>
            <label htmlFor="confirmPassword" className="block text-sm font-medium text-gray-700 mb-1">
              Confirmar Senha *
            </label>
            <div className="relative">
              <input
                id="confirmPassword"
                name="confirmPassword"
                type={showConfirmPassword ? 'text' : 'password'}
                required
                value={formData.confirmPassword}
                onChange={handleChange}
                disabled={loading}
                placeholder="Digite a senha novamente"
                className="w-full px-4 py-2 pr-10 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent disabled:bg-gray-100 disabled:cursor-not-allowed"
              />
              <button
                type="button"
                onClick={() => setShowConfirmPassword(!showConfirmPassword)}
                className="absolute right-3 top-1/2 -translate-y-1/2 text-gray-500 hover:text-gray-700"
              >
                {showConfirmPassword ? <EyeOff className="w-5 h-5" /> : <Eye className="w-5 h-5" />}
              </button>
            </div>
          </div>

          {/* Submit Button */}
          <button
            type="submit"
            disabled={loading}
            className="w-full bg-blue-600 text-white py-3 px-4 rounded-lg font-medium hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 disabled:bg-gray-400 disabled:cursor-not-allowed transition-colors flex items-center justify-center"
          >
            {loading ? (
              <>
                <Loader2 className="w-5 h-5 mr-2 animate-spin" />
                Criando conta...
              </>
            ) : (
              'Criar Conta'
            )}
          </button>
        </form>

        {variant === 'standalone' && (
          <div className="mt-6 text-center">
            <p className="text-sm text-gray-600">
              Já tem uma conta?{' '}
              <a href="/login" className="text-blue-600 hover:text-blue-700 font-medium">
                Faça login
              </a>
            </p>
          </div>
        )}

        {variant === 'standalone' && (
          <div className="mt-6 pt-6 border-t border-gray-200">
            <p className="text-xs text-gray-500 text-center">
              Ao criar uma conta, você concorda com nossos{' '}
              <a href="/termos" className="text-blue-600 hover:underline">
                Termos de Uso
              </a>{' '}
              e{' '}
              <a href="/privacidade" className="text-blue-600 hover:underline">
                Política de Privacidade
              </a>
            </p>
          </div>
        )}
      </div>
    </div>
  );
}
