'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { Building2, Save, Loader2, MapPin, Phone, Mail, FileText, CheckCircle } from 'lucide-react';
import { Tenant, UpdateTenantRequest } from '@/types/tenant';

export default function ConfiguracoesPage() {
  const router = useRouter();
  const [tenant, setTenant] = useState<Tenant | null>(null);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [success, setSuccess] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const [formData, setFormData] = useState<UpdateTenantRequest>({
    name: '',
    email: '',
    phone: '',
    document: '',
    document_type: 'cnpj',
    business_type: 'imobiliaria',
    creci: '',
    street: '',
    number: '',
    complement: '',
    neighborhood: '',
    city: '',
    state: '',
    zip_code: '',
    country: 'BR',
  });

  useEffect(() => {
    fetchTenant();
  }, []);

  const fetchTenant = async () => {
    try {
      setLoading(true);
      setError(null);
      const tenantId = localStorage.getItem('tenant_id');

      if (!tenantId) {
        setError('Tenant ID não encontrado');
        return;
      }

      const { auth } = await import('@/lib/firebase');
      const user = auth.currentUser;

      if (!user) {
        router.push('/login');
        return;
      }

      const token = await user.getIdToken(true);
      const url = `${process.env.NEXT_PUBLIC_API_URL?.replace('/api/v1', '')}/tenants/${tenantId}`;

      const response = await fetch(url, {
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
      });

      if (!response.ok) {
        throw new Error('Erro ao carregar dados do tenant');
      }

      const data = await response.json();
      const tenantData = data.data;
      setTenant(tenantData);

      // Populate form
      setFormData({
        name: tenantData.name || '',
        email: tenantData.email || '',
        phone: tenantData.phone || '',
        document: tenantData.document || '',
        document_type: tenantData.document_type || 'cnpj',
        business_type: tenantData.business_type || 'imobiliaria',
        creci: tenantData.creci || '',
        street: tenantData.street || '',
        number: tenantData.number || '',
        complement: tenantData.complement || '',
        neighborhood: tenantData.neighborhood || '',
        city: tenantData.city || '',
        state: tenantData.state || '',
        zip_code: tenantData.zip_code || '',
        country: tenantData.country || 'BR',
      });
    } catch (err) {
      console.error('Erro ao carregar tenant:', err);
      setError(err instanceof Error ? err.message : 'Erro desconhecido');
    } finally {
      setLoading(false);
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    try {
      setSaving(true);
      setError(null);
      setSuccess(false);

      const tenantId = localStorage.getItem('tenant_id');
      if (!tenantId) {
        throw new Error('Tenant ID não encontrado');
      }

      const { auth } = await import('@/lib/firebase');
      const user = auth.currentUser;

      if (!user) {
        router.push('/login');
        return;
      }

      const token = await user.getIdToken(true);
      const url = `${process.env.NEXT_PUBLIC_API_URL?.replace('/api/v1', '')}/tenants/${tenantId}`;

      const response = await fetch(url, {
        method: 'PUT',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(formData),
      });

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.error || 'Erro ao salvar configurações');
      }

      setSuccess(true);

      // Atualizar nome no localStorage se mudou
      if (formData.name) {
        localStorage.setItem('tenant_name', formData.name);
      }

      // Recarregar dados
      await fetchTenant();

      // Limpar mensagem de sucesso após 3s
      setTimeout(() => setSuccess(false), 3000);
    } catch (err) {
      console.error('Erro ao salvar:', err);
      setError(err instanceof Error ? err.message : 'Erro ao salvar configurações');
    } finally {
      setSaving(false);
    }
  };

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { name, value } = e.target;
    setFormData(prev => ({ ...prev, [name]: value }));
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <Loader2 className="h-8 w-8 animate-spin text-blue-600" />
      </div>
    );
  }

  return (
    <div className="p-6 max-w-4xl mx-auto">
      {/* Header */}
      <div className="mb-8">
        <div className="flex items-center gap-3 mb-2">
          <Building2 className="h-8 w-8 text-blue-600" />
          <h1 className="text-3xl font-bold text-gray-900">Configurações do Tenant</h1>
        </div>
        <p className="text-gray-600">Gerencie as informações da sua empresa</p>
      </div>

      {/* Success Message */}
      {success && (
        <div className="mb-6 p-4 bg-green-50 border border-green-200 rounded-lg flex items-center gap-3">
          <CheckCircle className="h-5 w-5 text-green-600" />
          <p className="text-green-800 font-medium">Configurações salvas com sucesso!</p>
        </div>
      )}

      {/* Error Message */}
      {error && (
        <div className="mb-6 p-4 bg-red-50 border border-red-200 rounded-lg">
          <p className="text-red-800">{error}</p>
        </div>
      )}

      {/* Form */}
      <form onSubmit={handleSubmit} className="space-y-8">
        {/* Informações Básicas */}
        <div className="bg-white rounded-lg shadow p-6">
          <h2 className="text-xl font-semibold text-gray-900 mb-4 flex items-center gap-2">
            <FileText className="h-5 w-5" />
            Informações Básicas
          </h2>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            <div className="md:col-span-2">
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Razão Social *
              </label>
              <input
                type="text"
                name="name"
                value={formData.name}
                onChange={handleInputChange}
                required
                className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                placeholder="Ex: ALTATECH Imóveis Ltda"
              />
              <p className="mt-1 text-xs text-gray-500">
                Nome oficial registrado no CNPJ ou nome completo do CPF
              </p>
            </div>

            <div className="md:col-span-2">
              <label className="block text-sm font-medium text-gray-700 mb-3">
                Tipo de Pessoa
              </label>
              <div className="flex gap-6 mb-4">
                <label className="flex items-center gap-2 cursor-pointer">
                  <input
                    type="radio"
                    name="document_type"
                    value="cpf"
                    checked={formData.document_type === 'cpf'}
                    onChange={handleInputChange}
                    className="w-4 h-4 text-blue-600 focus:ring-blue-500"
                  />
                  <span className="text-sm text-gray-700">Pessoa Física (CPF)</span>
                </label>
                <label className="flex items-center gap-2 cursor-pointer">
                  <input
                    type="radio"
                    name="document_type"
                    value="cnpj"
                    checked={formData.document_type === 'cnpj'}
                    onChange={handleInputChange}
                    className="w-4 h-4 text-blue-600 focus:ring-blue-500"
                  />
                  <span className="text-sm text-gray-700">Pessoa Jurídica (CNPJ)</span>
                </label>
              </div>
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                {formData.document_type === 'cpf' ? 'CPF' : 'CNPJ'}
              </label>
              <input
                type="text"
                name="document"
                value={formData.document}
                onChange={handleInputChange}
                className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                placeholder={formData.document_type === 'cpf' ? '000.000.000-00' : '00.000.000/0000-00'}
              />
              <p className="mt-1 text-xs text-gray-500">
                {formData.document_type === 'cpf'
                  ? 'Informe o CPF com 11 dígitos'
                  : 'Informe o CNPJ com 14 dígitos'}
              </p>
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                {formData.document_type === 'cpf' ? 'CRECI Pessoa Física' : 'CRECI Pessoa Jurídica'}
              </label>
              <input
                type="text"
                name="creci"
                value={formData.creci}
                onChange={handleInputChange}
                className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                placeholder={formData.document_type === 'cpf' ? 'Ex: 12345-F/SP' : 'Ex: 12345-J/SP'}
              />
            </div>

            <div className="md:col-span-2">
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Tipo de Negócio
              </label>
              <select
                name="business_type"
                value={formData.business_type}
                onChange={handleInputChange}
                className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
              >
                <option value="imobiliaria">Imobiliária</option>
                <option value="incorporadora">Incorporadora</option>
                <option value="loteadora">Loteadora</option>
                <option value="construtora">Construtora</option>
                <option value="corretor_autonomo">Corretor Autônomo</option>
              </select>
              <p className="mt-1 text-xs text-gray-500">
                Selecione o tipo de negócio que melhor descreve sua operação
              </p>
            </div>
          </div>
        </div>

        {/* Contato */}
        <div className="bg-white rounded-lg shadow p-6">
          <h2 className="text-xl font-semibold text-gray-900 mb-4 flex items-center gap-2">
            <Phone className="h-5 w-5" />
            Contato
          </h2>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2 flex items-center gap-2">
                <Mail className="h-4 w-4" />
                Email
              </label>
              <input
                type="email"
                name="email"
                value={formData.email}
                onChange={handleInputChange}
                className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                placeholder="contato@empresa.com.br"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2 flex items-center gap-2">
                <Phone className="h-4 w-4" />
                Telefone
              </label>
              <input
                type="tel"
                name="phone"
                value={formData.phone}
                onChange={handleInputChange}
                className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                placeholder="(00) 0000-0000"
              />
            </div>
          </div>
        </div>

        {/* Endereço */}
        <div className="bg-white rounded-lg shadow p-6">
          <h2 className="text-xl font-semibold text-gray-900 mb-4 flex items-center gap-2">
            <MapPin className="h-5 w-5" />
            Endereço
          </h2>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                CEP
              </label>
              <input
                type="text"
                name="zip_code"
                value={formData.zip_code}
                onChange={handleInputChange}
                className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                placeholder="00000-000"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Rua
              </label>
              <input
                type="text"
                name="street"
                value={formData.street}
                onChange={handleInputChange}
                className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                placeholder="Rua Exemplo"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Número
              </label>
              <input
                type="text"
                name="number"
                value={formData.number}
                onChange={handleInputChange}
                className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                placeholder="123"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Complemento
              </label>
              <input
                type="text"
                name="complement"
                value={formData.complement}
                onChange={handleInputChange}
                className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                placeholder="Sala 101"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Bairro
              </label>
              <input
                type="text"
                name="neighborhood"
                value={formData.neighborhood}
                onChange={handleInputChange}
                className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                placeholder="Centro"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Cidade
              </label>
              <input
                type="text"
                name="city"
                value={formData.city}
                onChange={handleInputChange}
                className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                placeholder="São Paulo"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Estado (UF)
              </label>
              <input
                type="text"
                name="state"
                value={formData.state}
                onChange={handleInputChange}
                maxLength={2}
                className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                placeholder="SP"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                País
              </label>
              <input
                type="text"
                name="country"
                value={formData.country}
                onChange={handleInputChange}
                maxLength={2}
                className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                placeholder="BR"
              />
            </div>
          </div>
        </div>

        {/* Status */}
        {tenant && (
          <div className="bg-white rounded-lg shadow p-6">
            <h2 className="text-xl font-semibold text-gray-900 mb-4">Status</h2>
            <div className="flex items-center gap-3">
              <span className="text-sm font-medium text-gray-700">Status do Tenant:</span>
              <span className={`px-3 py-1 rounded-full text-sm font-medium ${
                tenant.is_active
                  ? 'bg-green-100 text-green-800'
                  : 'bg-red-100 text-red-800'
              }`}>
                {tenant.is_active ? 'Ativo' : 'Inativo'}
              </span>
            </div>
            {tenant.slug && (
              <div className="mt-3">
                <span className="text-sm font-medium text-gray-700">Slug:</span>
                <code className="ml-2 px-2 py-1 bg-gray-100 text-gray-800 rounded text-sm">
                  {tenant.slug}
                </code>
              </div>
            )}
          </div>
        )}

        {/* Actions */}
        <div className="flex justify-end gap-4">
          <button
            type="button"
            onClick={() => router.back()}
            className="px-6 py-2 border border-gray-300 rounded-lg text-gray-700 hover:bg-gray-50 transition-colors"
          >
            Cancelar
          </button>
          <button
            type="submit"
            disabled={saving}
            className="px-6 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors flex items-center gap-2 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {saving ? (
              <>
                <Loader2 className="h-4 w-4 animate-spin" />
                Salvando...
              </>
            ) : (
              <>
                <Save className="h-4 w-4" />
                Salvar Configurações
              </>
            )}
          </button>
        </div>
      </form>
    </div>
  );
}
