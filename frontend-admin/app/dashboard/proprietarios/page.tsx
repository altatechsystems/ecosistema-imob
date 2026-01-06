'use client';

import { useState, useEffect, useMemo, useCallback } from 'react';
import { useRouter } from 'next/navigation';
import { User, Search, Eye, Phone, Mail, MapPin, Home, Plus } from 'lucide-react';

interface Owner {
  id: string;
  name: string;
  phone?: string;
  email?: string;
  address?: string;
  city?: string;
  state?: string;
  data_source?: string;
  data_quality?: string;
  created_at?: string;
  updated_at?: string;
  property_count?: number;
}

export default function ProprietariosPage() {
  const router = useRouter();
  const [searchTerm, setSearchTerm] = useState('');
  const [owners, setOwners] = useState<Owner[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  const fetchOwners = useCallback(async () => {
    try {
      setLoading(true);
      const tenantId = localStorage.getItem('tenant_id');

      if (!tenantId) {
        setError('Tenant ID não encontrado');
        return;
      }

      // Get Firebase auth token
      const { auth } = await import('@/lib/firebase');
      const user = auth.currentUser;
      if (!user) {
        throw new Error('Usuário não autenticado');
      }
      const token = await user.getIdToken();

      const url = `${process.env.NEXT_PUBLIC_API_URL}/admin/${tenantId}/owners?limit=1000`;

      const response = await fetch(url, {
        headers: {
          'Authorization': `Bearer ${token}`,
        },
      });

      if (!response.ok) {
        throw new Error('Erro ao buscar proprietários');
      }

      const data = await response.json();

      setOwners(data.data || []);
    } catch (err: any) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchOwners();
  }, [fetchOwners]);

  const stats = useMemo(() => ({
    total: owners.length,
    withPhone: owners.filter(o => o.phone).length,
    withEmail: owners.filter(o => o.email).length,
    withAddress: owners.filter(o => o.address || o.city).length,
  }), [owners]);

  const filteredOwners = useMemo(() => {
    if (!searchTerm) return owners;

    const term = searchTerm.toLowerCase();
    return owners.filter(owner =>
      owner.name.toLowerCase().includes(term) ||
      owner.phone?.toLowerCase().includes(term) ||
      owner.email?.toLowerCase().includes(term) ||
      owner.city?.toLowerCase().includes(term)
    );
  }, [owners, searchTerm]);

  const getDataQualityBadge = (quality?: string) => {
    const badges = {
      'high': { label: 'Alta', color: 'bg-green-100 text-green-800' },
      'medium': { label: 'Média', color: 'bg-yellow-100 text-yellow-800' },
      'low': { label: 'Baixa', color: 'bg-red-100 text-red-800' },
    };

    const badge = badges[quality as keyof typeof badges] || badges.low;
    return (
      <span className={`px-2 py-0.5 text-xs font-medium rounded ${badge.color}`}>
        {badge.label}
      </span>
    );
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"></div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="text-center">
          <p className="text-red-600 mb-4">{error}</p>
          <button
            onClick={fetchOwners}
            className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700"
          >
            Tentar novamente
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="p-6">
      {/* Header */}
      <div className="mb-6">
        <h1 className="text-2xl font-bold text-gray-900 mb-2">Proprietários</h1>
        <p className="text-gray-600">Gerencie os proprietários dos imóveis</p>
      </div>

      {/* Stats */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4 mb-6">
        <div className="bg-white p-4 rounded-lg shadow-sm">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm text-gray-600">Total</p>
              <p className="text-2xl font-bold text-gray-900">{stats.total}</p>
            </div>
            <User className="w-8 h-8 text-blue-600" />
          </div>
        </div>
        <div className="bg-white p-4 rounded-lg shadow-sm">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm text-gray-600">Com Telefone</p>
              <p className="text-2xl font-bold text-gray-900">{stats.withPhone}</p>
            </div>
            <Phone className="w-8 h-8 text-green-600" />
          </div>
        </div>
        <div className="bg-white p-4 rounded-lg shadow-sm">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm text-gray-600">Com Email</p>
              <p className="text-2xl font-bold text-gray-900">{stats.withEmail}</p>
            </div>
            <Mail className="w-8 h-8 text-purple-600" />
          </div>
        </div>
        <div className="bg-white p-4 rounded-lg shadow-sm">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm text-gray-600">Com Endereço</p>
              <p className="text-2xl font-bold text-gray-900">{stats.withAddress}</p>
            </div>
            <MapPin className="w-8 h-8 text-orange-600" />
          </div>
        </div>
      </div>

      {/* Search */}
      <div className="mb-6">
        <div className="relative">
          <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400 w-5 h-5" />
          <input
            type="text"
            placeholder="Buscar por nome, telefone, email ou cidade..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            className="w-full pl-10 pr-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
          />
        </div>
      </div>

      {/* Owners List */}
      {filteredOwners.length === 0 ? (
        <div className="bg-white rounded-lg shadow-sm p-12 text-center">
          <User className="w-16 h-16 mx-auto mb-4 text-gray-300" />
          <h3 className="text-lg font-medium text-gray-900 mb-2">
            {searchTerm ? 'Nenhum proprietário encontrado' : 'Nenhum proprietário cadastrado'}
          </h3>
          <p className="text-gray-600 mb-4">
            {searchTerm ? 'Tente ajustar sua busca' : 'Comece importando dados ou cadastre um novo proprietário'}
          </p>
        </div>
      ) : (
        <div className="bg-white rounded-lg shadow-sm overflow-hidden">
          <div className="overflow-x-auto">
            <table className="min-w-full divide-y divide-gray-200">
              <thead className="bg-gray-50">
                <tr>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Proprietário
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Contato
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Localização
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Qualidade
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Ações
                  </th>
                </tr>
              </thead>
              <tbody className="bg-white divide-y divide-gray-200">
                {filteredOwners.map((owner) => (
                  <tr key={owner.id} className="hover:bg-gray-50">
                    <td className="px-6 py-4 whitespace-nowrap">
                      <div className="flex items-center">
                        <div className="flex-shrink-0 h-10 w-10 bg-blue-100 rounded-full flex items-center justify-center">
                          <User className="w-5 h-5 text-blue-600" />
                        </div>
                        <div className="ml-4">
                          <div className="text-sm font-medium text-gray-900">{owner.name}</div>
                          {owner.data_source && (
                            <div className="text-xs text-gray-500">Fonte: {owner.data_source}</div>
                          )}
                        </div>
                      </div>
                    </td>
                    <td className="px-6 py-4">
                      <div className="text-sm text-gray-900 space-y-1">
                        {owner.phone ? (
                          <div className="flex items-center gap-2">
                            <Phone className="w-3 h-3 text-gray-400" />
                            <span>{owner.phone}</span>
                          </div>
                        ) : (
                          <span className="text-gray-400 text-xs">Sem telefone</span>
                        )}
                        {owner.email ? (
                          <div className="flex items-center gap-2">
                            <Mail className="w-3 h-3 text-gray-400" />
                            <span className="text-xs">{owner.email}</span>
                          </div>
                        ) : (
                          <span className="text-gray-400 text-xs">Sem email</span>
                        )}
                      </div>
                    </td>
                    <td className="px-6 py-4">
                      <div className="text-sm text-gray-900">
                        {owner.city || owner.address ? (
                          <div className="flex items-center gap-2">
                            <MapPin className="w-3 h-3 text-gray-400" />
                            <span>{[owner.city, owner.state].filter(Boolean).join(' - ') || owner.address}</span>
                          </div>
                        ) : (
                          <span className="text-gray-400 text-xs">Não informado</span>
                        )}
                      </div>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      {getDataQualityBadge(owner.data_quality)}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm font-medium">
                      <button
                        onClick={() => router.push(`/dashboard/proprietarios/${owner.id}`)}
                        className="flex items-center gap-2 text-blue-600 hover:text-blue-900"
                      >
                        <Eye className="w-4 h-4" />
                        Ver detalhes
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>

          {/* Results count */}
          <div className="px-6 py-4 border-t border-gray-200 bg-gray-50">
            <p className="text-sm text-gray-600">
              Mostrando {filteredOwners.length} de {stats.total} proprietários
            </p>
          </div>
        </div>
      )}
    </div>
  );
}
