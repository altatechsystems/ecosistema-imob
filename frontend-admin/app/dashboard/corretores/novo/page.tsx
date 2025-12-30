'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import { ArrowLeft, Save, X, User, FileText, Award } from 'lucide-react';
import { Broker } from '@/types/broker';

export default function NewBrokerPage() {
  const router = useRouter();

  const [broker, setBroker] = useState<Partial<Broker>>({
    name: '',
    email: '',
    phone: '',
    creci: '',
    document: '',
    document_type: 'cpf',
    role: 'broker',
    is_active: true,
    bio: '',
    specialties: '',
    languages: 'Português',
    experience: 0,
    company: '',
    website: '',
  });

  const [saving, setSaving] = useState(false);
  const [error, setError] = useState('');

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!broker.name || !broker.email || !broker.creci) {
      setError('Por favor, preencha todos os campos obrigatórios');
      return;
    }

    try {
      setSaving(true);
      setError('');

      const tenantId = localStorage.getItem('tenant_id');

      if (!tenantId) {
        setError('Tenant ID não encontrado');
        return;
      }

      const { auth } = await import('@/lib/firebase');
      const user = auth.currentUser;

      if (!user) {
        setError('Usuário não autenticado');
        router.push('/login');
        return;
      }

      const token = await user.getIdToken(true);

      const response = await fetch(
        `${process.env.NEXT_PUBLIC_API_URL}/admin/${tenantId}/brokers`,
        {
          method: 'POST',
          headers: {
            'Authorization': `Bearer ${token}`,
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({
            ...broker,
            tenant_id: tenantId,
          }),
        }
      );

      if (!response.ok) {
        const data = await response.json();
        throw new Error(data.error || 'Erro ao cadastrar corretor');
      }

      const data = await response.json();

      // Redirecionar para página de detalhes do corretor criado
      router.push(`/dashboard/corretores/${data.data.id}`);
    } catch (err: any) {
      console.error('Erro ao cadastrar:', err);
      setError(err.message);
    } finally {
      setSaving(false);
    }
  };

  const handleChange = (field: keyof Broker, value: any) => {
    setBroker({ ...broker, [field]: value });
  };

  return (
    <div className="p-6">
      {/* Header */}
      <div className="mb-6">
        <button
          onClick={() => router.push('/dashboard/corretores')}
          className="flex items-center gap-2 text-gray-600 hover:text-gray-900 mb-4"
        >
          <ArrowLeft className="w-5 h-5" />
          Voltar para lista
        </button>

        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-2xl font-bold text-gray-900 mb-2">Novo Corretor</h1>
            <p className="text-gray-600">Cadastre um novo corretor na sua imobiliária</p>
          </div>
        </div>
      </div>

      {error && (
        <div className="bg-red-50 border border-red-200 rounded-lg p-4 mb-6">
          <p className="text-red-600">{error}</p>
        </div>
      )}

      <form onSubmit={handleSubmit}>
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
          {/* Main Content */}
          <div className="lg:col-span-2 space-y-6">
            {/* Informações Básicas */}
            <div className="bg-white rounded-lg shadow-sm p-6">
              <h2 className="text-xl font-bold text-gray-900 mb-4 flex items-center gap-2">
                <User className="w-5 h-5" />
                Informações Básicas
              </h2>

              <div className="grid grid-cols-2 gap-4">
                <div className="col-span-2">
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    Nome Completo *
                  </label>
                  <input
                    type="text"
                    value={broker.name || ''}
                    onChange={(e) => handleChange('name', e.target.value)}
                    required
                    placeholder="Digite o nome completo do corretor"
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                  />
                </div>

                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    Email *
                  </label>
                  <input
                    type="email"
                    value={broker.email || ''}
                    onChange={(e) => handleChange('email', e.target.value)}
                    required
                    placeholder="email@exemplo.com"
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                  />
                </div>

                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    Telefone
                  </label>
                  <input
                    type="tel"
                    value={broker.phone || ''}
                    onChange={(e) => handleChange('phone', e.target.value)}
                    placeholder="(11) 98765-4321"
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                  />
                </div>

                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    CRECI *
                  </label>
                  <input
                    type="text"
                    value={broker.creci || ''}
                    onChange={(e) => handleChange('creci', e.target.value)}
                    required
                    placeholder="12345-J/SP"
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                  />
                  <p className="text-xs text-gray-500 mt-1">
                    Formato: XXXXX-J/UF (ex: 12345-J/SP)
                  </p>
                </div>

                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    CPF/CNPJ
                  </label>
                  <input
                    type="text"
                    value={broker.document || ''}
                    onChange={(e) => handleChange('document', e.target.value)}
                    placeholder="000.000.000-00"
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                  />
                </div>

                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    Tipo de Documento
                  </label>
                  <select
                    value={broker.document_type || 'cpf'}
                    onChange={(e) => handleChange('document_type', e.target.value)}
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                  >
                    <option value="cpf">CPF</option>
                    <option value="cnpj">CNPJ</option>
                  </select>
                </div>
              </div>
            </div>

            {/* Perfil Profissional */}
            <div className="bg-white rounded-lg shadow-sm p-6">
              <h2 className="text-xl font-bold text-gray-900 mb-4 flex items-center gap-2">
                <FileText className="w-5 h-5" />
                Perfil Profissional
              </h2>

              <div className="space-y-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    Biografia
                  </label>
                  <textarea
                    value={broker.bio || ''}
                    onChange={(e) => handleChange('bio', e.target.value)}
                    rows={4}
                    placeholder="Conte sobre sua experiência, especialidades e o que o diferencia..."
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                  />
                  <p className="text-xs text-gray-500 mt-1">
                    Esta biografia será exibida no perfil público do corretor
                  </p>
                </div>

                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-2">
                      Especialidades
                    </label>
                    <input
                      type="text"
                      value={broker.specialties || ''}
                      onChange={(e) => handleChange('specialties', e.target.value)}
                      placeholder="Ex: Comprador, Vendedor, Aluguel"
                      className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                    />
                  </div>

                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-2">
                      Idiomas
                    </label>
                    <input
                      type="text"
                      value={broker.languages || ''}
                      onChange={(e) => handleChange('languages', e.target.value)}
                      placeholder="Ex: Português, Inglês, Espanhol"
                      className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                    />
                  </div>

                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-2">
                      Anos de Experiência
                    </label>
                    <input
                      type="number"
                      value={broker.experience || 0}
                      onChange={(e) => handleChange('experience', parseInt(e.target.value) || 0)}
                      min="0"
                      placeholder="0"
                      className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                    />
                  </div>

                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-2">
                      Empresa/Imobiliária
                    </label>
                    <input
                      type="text"
                      value={broker.company || ''}
                      onChange={(e) => handleChange('company', e.target.value)}
                      placeholder="Nome da empresa"
                      className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                    />
                  </div>
                </div>

                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    Website
                  </label>
                  <input
                    type="url"
                    value={broker.website || ''}
                    onChange={(e) => handleChange('website', e.target.value)}
                    placeholder="https://seusite.com.br"
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                  />
                </div>

                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    Certificações e Prêmios
                  </label>
                  <textarea
                    value={broker.certifications_awards || ''}
                    onChange={(e) => handleChange('certifications_awards', e.target.value)}
                    rows={3}
                    placeholder="Liste suas certificações e prêmios..."
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                  />
                </div>
              </div>
            </div>
          </div>

          {/* Sidebar */}
          <div className="space-y-6">
            {/* Status e Perfil */}
            <div className="bg-white rounded-lg shadow-sm p-6">
              <h3 className="font-bold text-gray-900 mb-4">Status e Perfil</h3>

              <div className="space-y-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    Perfil/Função
                  </label>
                  <select
                    value={broker.role || 'broker'}
                    onChange={(e) => handleChange('role', e.target.value)}
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                  >
                    <option value="broker">Corretor</option>
                    <option value="manager">Gerente</option>
                    <option value="broker_admin">Admin Imobiliária</option>
                    <option value="platform_admin">Admin Plataforma</option>
                  </select>
                  <p className="text-xs text-gray-500 mt-1">
                    Define as permissões do corretor no sistema
                  </p>
                </div>

                <div className="flex items-center">
                  <input
                    type="checkbox"
                    checked={broker.is_active || false}
                    onChange={(e) => handleChange('is_active', e.target.checked)}
                    className="w-4 h-4 text-blue-600 border-gray-300 rounded focus:ring-blue-500"
                  />
                  <label className="ml-2 text-sm text-gray-700">
                    Corretor ativo
                  </label>
                </div>
              </div>
            </div>

            {/* Informações */}
            <div className="bg-blue-50 border border-blue-200 rounded-lg p-4">
              <h4 className="font-medium text-blue-900 mb-2">📌 Informação Importante</h4>
              <p className="text-sm text-blue-800">
                Após cadastrar o corretor, você poderá adicionar mais informações como foto,
                redes sociais e outros dados que aparecerão no perfil público.
              </p>
            </div>

            {/* Ações */}
            <div className="bg-white rounded-lg shadow-sm p-6">
              <div className="space-y-2">
                <button
                  type="submit"
                  disabled={saving}
                  className="w-full flex items-center justify-center gap-2 px-4 py-3 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  <Save className="w-5 h-5" />
                  {saving ? 'Cadastrando...' : 'Cadastrar Corretor'}
                </button>
                <button
                  type="button"
                  onClick={() => router.push('/dashboard/corretores')}
                  className="w-full flex items-center justify-center gap-2 px-4 py-3 border border-gray-300 rounded-lg hover:bg-gray-50 transition-colors"
                >
                  <X className="w-5 h-5" />
                  Cancelar
                </button>
              </div>
            </div>
          </div>
        </div>
      </form>
    </div>
  );
}
