'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { Building2, Plus, Search, CheckCircle, XCircle, Shield, Loader2 } from 'lucide-react';
import { Tenant, TenantStats } from '@/types/tenant';

export default function TenantsAdminPage() {
  const router = useRouter();
  const [tenants, setTenants] = useState<Tenant[]>([]);
  const [stats, setStats] = useState<TenantStats>({
    total: 0,
    active: 0,
    inactive: 0,
    platformAdmins: 0,
  });
  const [loading, setLoading] = useState(true);
  const [searchTerm, setSearchTerm] = useState('');
  const [filterStatus, setFilterStatus] = useState<'all' | 'active' | 'inactive'>('all');

  useEffect(() => {
    fetchTenants();
  }, []);

  const fetchTenants = async () => {
    try {
      setLoading(true);

      const { auth } = await import('@/lib/firebase');
      const user = auth.currentUser;

      if (!user) {
        router.push('/login');
        return;
      }

      const token = await user.getIdToken(true);
      const url = `${process.env.NEXT_PUBLIC_API_URL?.replace('/api/v1', '')}/tenants`;

      const response = await fetch(url, {
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
      });

      if (!response.ok) {
        throw new Error('Erro ao carregar tenants');
      }

      const data = await response.json();
      const tenantsList = data.data || [];
      setTenants(tenantsList);

      // Calculate stats
      const active = tenantsList.filter((t: Tenant) => t.is_active).length;
      const inactive = tenantsList.filter((t: Tenant) => !t.is_active).length;
      const platformAdmins = tenantsList.filter((t: Tenant) => t.is_platform_admin).length;

      setStats({
        total: tenantsList.length,
        active,
        inactive,
        platformAdmins,
      });
    } catch (error) {
      console.error('Erro ao carregar tenants:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleToggleActive = async (tenant: Tenant) => {
    try {
      const { auth } = await import('@/lib/firebase');
      const user = auth.currentUser;

      if (!user) {
        router.push('/login');
        return;
      }

      const token = await user.getIdToken(true);
      const action = tenant.is_active ? 'deactivate' : 'activate';
      const url = `${process.env.NEXT_PUBLIC_API_URL?.replace('/api/v1', '')}/tenants/${tenant.id}/${action}`;

      const response = await fetch(url, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
      });

      if (!response.ok) {
        throw new Error(`Erro ao ${action === 'activate' ? 'ativar' : 'desativar'} tenant`);
      }

      // Reload tenants
      await fetchTenants();
    } catch (error) {
      console.error('Erro:', error);
      alert(error instanceof Error ? error.message : 'Erro ao atualizar tenant');
    }
  };

  const filteredTenants = tenants.filter(tenant => {
    // Filter by status
    if (filterStatus === 'active' && !tenant.is_active) return false;
    if (filterStatus === 'inactive' && tenant.is_active) return false;

    // Filter by search
    if (searchTerm) {
      const search = searchTerm.toLowerCase();
      return (
        tenant.name.toLowerCase().includes(search) ||
        tenant.slug?.toLowerCase().includes(search) ||
        tenant.document?.toLowerCase().includes(search) ||
        tenant.email?.toLowerCase().includes(search)
      );
    }

    return true;
  });

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <Loader2 className="h-8 w-8 animate-spin text-blue-600" />
      </div>
    );
  }

  return (
    <div className="p-6">
      {/* Header */}
      <div className="mb-8">
        <div className="flex items-center justify-between">
          <div>
            <div className="flex items-center gap-3 mb-2">
              <Building2 className="h-8 w-8 text-blue-600" />
              <h1 className="text-3xl font-bold text-gray-900">Gerenciar Tenants</h1>
            </div>
            <p className="text-gray-600">Gerencie todas as imobiliárias do sistema</p>
          </div>
          <button
            onClick={() => router.push('/admin/tenants/novo')}
            className="flex items-center gap-2 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors"
          >
            <Plus className="h-5 w-5" />
            Novo Tenant
          </button>
        </div>
      </div>

      {/* Stats Cards */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-6 mb-8">
        <div className="bg-white rounded-lg shadow p-6">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm font-medium text-gray-600">Total</p>
              <p className="text-3xl font-bold text-gray-900 mt-2">{stats.total}</p>
            </div>
            <Building2 className="h-12 w-12 text-gray-400" />
          </div>
        </div>

        <div className="bg-white rounded-lg shadow p-6">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm font-medium text-gray-600">Ativos</p>
              <p className="text-3xl font-bold text-green-600 mt-2">{stats.active}</p>
            </div>
            <CheckCircle className="h-12 w-12 text-green-400" />
          </div>
        </div>

        <div className="bg-white rounded-lg shadow p-6">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm font-medium text-gray-600">Inativos</p>
              <p className="text-3xl font-bold text-red-600 mt-2">{stats.inactive}</p>
            </div>
            <XCircle className="h-12 w-12 text-red-400" />
          </div>
        </div>

        <div className="bg-white rounded-lg shadow p-6">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm font-medium text-gray-600">Platform Admins</p>
              <p className="text-3xl font-bold text-purple-600 mt-2">{stats.platformAdmins}</p>
            </div>
            <Shield className="h-12 w-12 text-purple-400" />
          </div>
        </div>
      </div>

      {/* Filters */}
      <div className="bg-white rounded-lg shadow p-6 mb-6">
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          {/* Search */}
          <div className="relative">
            <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-5 w-5 text-gray-400" />
            <input
              type="text"
              placeholder="Buscar por nome, slug, CNPJ ou email..."
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              className="w-full pl-10 pr-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
            />
          </div>

          {/* Status Filter */}
          <div className="flex gap-2">
            <button
              onClick={() => setFilterStatus('all')}
              className={`flex-1 px-4 py-2 rounded-lg font-medium transition-colors ${
                filterStatus === 'all'
                  ? 'bg-blue-600 text-white'
                  : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
              }`}
            >
              Todos ({stats.total})
            </button>
            <button
              onClick={() => setFilterStatus('active')}
              className={`flex-1 px-4 py-2 rounded-lg font-medium transition-colors ${
                filterStatus === 'active'
                  ? 'bg-green-600 text-white'
                  : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
              }`}
            >
              Ativos ({stats.active})
            </button>
            <button
              onClick={() => setFilterStatus('inactive')}
              className={`flex-1 px-4 py-2 rounded-lg font-medium transition-colors ${
                filterStatus === 'inactive'
                  ? 'bg-red-600 text-white'
                  : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
              }`}
            >
              Inativos ({stats.inactive})
            </button>
          </div>
        </div>
      </div>

      {/* Tenants List */}
      <div className="bg-white rounded-lg shadow overflow-hidden">
        <div className="overflow-x-auto">
          <table className="w-full">
            <thead className="bg-gray-50 border-b border-gray-200">
              <tr>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Tenant
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Slug
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Contato
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  CNPJ
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Status
                </th>
                <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Ações
                </th>
              </tr>
            </thead>
            <tbody className="bg-white divide-y divide-gray-200">
              {filteredTenants.length === 0 ? (
                <tr>
                  <td colSpan={6} className="px-6 py-12 text-center text-gray-500">
                    Nenhum tenant encontrado
                  </td>
                </tr>
              ) : (
                filteredTenants.map((tenant) => (
                  <tr key={tenant.id} className="hover:bg-gray-50 transition-colors">
                    <td className="px-6 py-4 whitespace-nowrap">
                      <div className="flex items-center">
                        <div className="flex-shrink-0 h-10 w-10 bg-blue-100 rounded-full flex items-center justify-center">
                          <Building2 className="h-5 w-5 text-blue-600" />
                        </div>
                        <div className="ml-4">
                          <div className="text-sm font-medium text-gray-900">{tenant.name}</div>
                          {tenant.is_platform_admin && (
                            <div className="flex items-center gap-1 mt-1">
                              <Shield className="h-3 w-3 text-purple-600" />
                              <span className="text-xs text-purple-600 font-medium">Platform Admin</span>
                            </div>
                          )}
                        </div>
                      </div>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <code className="text-sm text-gray-600 bg-gray-100 px-2 py-1 rounded">
                        {tenant.slug}
                      </code>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <div className="text-sm text-gray-900">{tenant.email || '-'}</div>
                      <div className="text-sm text-gray-500">{tenant.phone || '-'}</div>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <div className="text-sm text-gray-900">{tenant.document || '-'}</div>
                      {tenant.creci && (
                        <div className="text-xs text-gray-500">CRECI: {tenant.creci}</div>
                      )}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <span className={`px-3 py-1 inline-flex text-xs leading-5 font-semibold rounded-full ${
                        tenant.is_active
                          ? 'bg-green-100 text-green-800'
                          : 'bg-red-100 text-red-800'
                      }`}>
                        {tenant.is_active ? 'Ativo' : 'Inativo'}
                      </span>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                      <div className="flex items-center justify-end gap-2">
                        <button
                          onClick={() => router.push(`/admin/tenants/${tenant.id}`)}
                          className="text-blue-600 hover:text-blue-900 transition-colors"
                        >
                          Editar
                        </button>
                        <button
                          onClick={() => handleToggleActive(tenant)}
                          className={`transition-colors ${
                            tenant.is_active
                              ? 'text-red-600 hover:text-red-900'
                              : 'text-green-600 hover:text-green-900'
                          }`}
                        >
                          {tenant.is_active ? 'Desativar' : 'Ativar'}
                        </button>
                      </div>
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>
      </div>

      {/* Results count */}
      {filteredTenants.length > 0 && (
        <div className="mt-4 text-sm text-gray-600 text-center">
          Mostrando {filteredTenants.length} de {tenants.length} tenants
        </div>
      )}
    </div>
  );
}
