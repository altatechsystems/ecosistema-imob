'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { Plus, Search, Users, UserCheck, UserX, Shield } from 'lucide-react';
import { Broker, BrokerStats } from '@/types/broker';

export default function BrokersPage() {
  const router = useRouter();
  const [brokers, setBrokers] = useState<Broker[]>([]);
  const [stats, setStats] = useState<BrokerStats>({
    total: 0,
    active: 0,
    inactive: 0,
    byRole: {
      platform_admin: 0,
      broker_admin: 0,
      broker: 0,
      manager: 0,
    },
  });
  const [loading, setLoading] = useState(true);
  const [searchTerm, setSearchTerm] = useState('');
  const [filterStatus, setFilterStatus] = useState<'all' | 'active' | 'inactive'>('all');
  const [filterRole, setFilterRole] = useState<string>('all');

  useEffect(() => {
    fetchBrokers();
  }, []);

  const fetchBrokers = async () => {
    try {
      setLoading(true);
      const tenantId = localStorage.getItem('tenant_id');

      if (!tenantId) {
        return;
      }

      const { auth } = await import('@/lib/firebase');
      const user = auth.currentUser;

      if (!user) {
        router.push('/login');
        return;
      }

      const token = await user.getIdToken(true);
      const url = `${process.env.NEXT_PUBLIC_API_URL}/admin/${tenantId}/brokers`;

      const response = await fetch(url, {
        headers: {
          'Authorization': `Bearer ${token}`,
        },
      });

      if (!response.ok) {
        const errorData = await response.json().catch(() => ({}));
        throw new Error(errorData.error || 'Erro ao buscar corretores');
      }

      const data = await response.json();

      const brokersData = data.data || [];

      setBrokers(brokersData);
      calculateStats(brokersData);
    } catch (err: any) {
      setError(err.message || 'Erro ao buscar corretores');
    } finally {
      setLoading(false);
    }
  };

  const calculateStats = (brokersData: Broker[]) => {
    const active = brokersData.filter(b => b.is_active).length;
    const inactive = brokersData.filter(b => !b.is_active).length;

    const byRole = {
      platform_admin: brokersData.filter(b => b.role === 'platform_admin').length,
      broker_admin: brokersData.filter(b => b.role === 'broker_admin').length,
      broker: brokersData.filter(b => b.role === 'broker').length,
      manager: brokersData.filter(b => b.role === 'manager').length,
    };

    setStats({
      total: brokersData.length,
      active,
      inactive,
      byRole,
    });
  };

  const filteredBrokers = brokers.filter(broker => {
    const matchesSearch = !searchTerm ||
      broker.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
      broker.email.toLowerCase().includes(searchTerm.toLowerCase()) ||
      broker.creci?.toLowerCase().includes(searchTerm.toLowerCase());

    const matchesStatus = filterStatus === 'all' ||
      (filterStatus === 'active' && broker.is_active) ||
      (filterStatus === 'inactive' && !broker.is_active);

    const matchesRole = filterRole === 'all' || broker.role === filterRole;

    return matchesSearch && matchesStatus && matchesRole;
  });

  const getRoleLabel = (role?: string) => {
    switch (role) {
      case 'platform_admin':
        return 'Admin Plataforma';
      case 'broker_admin':
        return 'Admin Imobiliária';
      case 'broker':
        return 'Corretor';
      case 'manager':
        return 'Gerente';
      default:
        return 'Corretor';
    }
  };

  const getRoleBadgeColor = (role?: string) => {
    switch (role) {
      case 'platform_admin':
        return 'bg-purple-100 text-purple-800';
      case 'broker_admin':
        return 'bg-blue-100 text-blue-800';
      case 'manager':
        return 'bg-green-100 text-green-800';
      default:
        return 'bg-gray-100 text-gray-800';
    }
  };

  return (
    <div className="p-4 md:p-6">
      {/* Header */}
      <div className="mb-6">
        <div className="flex flex-col md:flex-row md:items-center md:justify-between gap-4">
          <div>
            <h1 className="text-2xl md:text-3xl font-bold text-gray-900">Corretores</h1>
            <p className="text-sm md:text-base text-gray-600 mt-1">
              Gerencie os corretores da sua imobiliária (CRECI obrigatório)
            </p>
          </div>
          <button
            onClick={() => router.push('/dashboard/corretores/novo')}
            className="flex items-center justify-center gap-2 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors"
          >
            <Plus className="w-5 h-5" />
            <span>Novo Corretor</span>
          </button>
        </div>

        {/* Info Box */}
        <div className="mt-4 bg-blue-50 border border-blue-200 rounded-lg p-4">
          <div className="flex items-start gap-3">
            <Shield className="w-5 h-5 text-blue-600 mt-0.5 flex-shrink-0" />
            <div className="flex-1">
              <h3 className="font-semibold text-blue-900 mb-2">
                Sobre os Corretores
              </h3>
              <p className="text-sm text-blue-800 mb-3">
                Esta página lista apenas <strong>corretores credenciados</strong> com CRECI obrigatório.
              </p>
              <div className="flex items-center gap-2 bg-white border border-blue-200 rounded-lg p-3">
                <Shield className="w-4 h-4 text-blue-600 flex-shrink-0" />
                <p className="text-sm text-blue-900">
                  Para gerenciar <strong>usuários administrativos</strong> (sem CRECI), acesse{' '}
                  <a
                    href="/dashboard/equipe"
                    className="inline-flex items-center gap-1 font-bold text-blue-700 hover:text-blue-900 underline"
                  >
                    Equipe no menu lateral
                  </a>
                </p>
              </div>
            </div>
          </div>
        </div>
      </div>

      {/* Stats Cards */}
      <div className="grid grid-cols-2 lg:grid-cols-4 gap-4 mb-6">
        <div className="bg-white rounded-lg shadow-sm p-4 md:p-6">
          <div className="flex items-center gap-3">
            <div className="p-2 bg-blue-100 rounded-lg">
              <Users className="w-6 h-6 text-blue-600" />
            </div>
            <div>
              <p className="text-xs md:text-sm text-gray-600">Total</p>
              <p className="text-xl md:text-2xl font-bold text-gray-900">{stats.total}</p>
            </div>
          </div>
        </div>

        <div className="bg-white rounded-lg shadow-sm p-4 md:p-6">
          <div className="flex items-center gap-3">
            <div className="p-2 bg-green-100 rounded-lg">
              <UserCheck className="w-6 h-6 text-green-600" />
            </div>
            <div>
              <p className="text-xs md:text-sm text-gray-600">Ativos</p>
              <p className="text-xl md:text-2xl font-bold text-gray-900">{stats.active}</p>
            </div>
          </div>
        </div>

        <div className="bg-white rounded-lg shadow-sm p-4 md:p-6">
          <div className="flex items-center gap-3">
            <div className="p-2 bg-red-100 rounded-lg">
              <UserX className="w-6 h-6 text-red-600" />
            </div>
            <div>
              <p className="text-xs md:text-sm text-gray-600">Inativos</p>
              <p className="text-xl md:text-2xl font-bold text-gray-900">{stats.inactive}</p>
            </div>
          </div>
        </div>

        <div className="bg-white rounded-lg shadow-sm p-4 md:p-6">
          <div className="flex items-center gap-3">
            <div className="p-2 bg-purple-100 rounded-lg">
              <Shield className="w-6 h-6 text-purple-600" />
            </div>
            <div>
              <p className="text-xs md:text-sm text-gray-600">Admins</p>
              <p className="text-xl md:text-2xl font-bold text-gray-900">
                {stats.byRole.platform_admin + stats.byRole.broker_admin}
              </p>
            </div>
          </div>
        </div>
      </div>

      {/* Filters */}
      <div className="bg-white rounded-lg shadow-sm p-4 mb-6">
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          {/* Search */}
          <div className="md:col-span-1">
            <div className="relative">
              <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400 w-5 h-5" />
              <input
                type="text"
                placeholder="Buscar por nome, email ou CRECI..."
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
                className="w-full pl-10 pr-4 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 text-sm"
              />
            </div>
          </div>

          {/* Status Filter */}
          <div>
            <select
              value={filterStatus}
              onChange={(e) => setFilterStatus(e.target.value as any)}
              className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 text-sm"
            >
              <option value="all">Todos os status</option>
              <option value="active">Ativos</option>
              <option value="inactive">Inativos</option>
            </select>
          </div>

          {/* Role Filter */}
          <div>
            <select
              value={filterRole}
              onChange={(e) => setFilterRole(e.target.value)}
              className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 text-sm"
            >
              <option value="all">Todos os perfis</option>
              <option value="platform_admin">Admin Plataforma</option>
              <option value="broker_admin">Admin Imobiliária</option>
              <option value="manager">Gerente</option>
              <option value="broker">Corretor</option>
            </select>
          </div>
        </div>
      </div>

      {/* Brokers List */}
      {loading ? (
        <div className="bg-white rounded-lg shadow-sm overflow-hidden">
          <div className="animate-pulse">
            <div className="h-12 bg-gray-100 border-b border-gray-200"></div>
            {[...Array(5)].map((_, i) => (
              <div key={i} className="border-b border-gray-200 p-4">
                <div className="flex items-center gap-4">
                  <div className="w-10 h-10 bg-gray-200 rounded-full"></div>
                  <div className="flex-1 space-y-2">
                    <div className="h-4 bg-gray-200 rounded w-1/4"></div>
                    <div className="h-3 bg-gray-200 rounded w-1/6"></div>
                  </div>
                </div>
              </div>
            ))}
          </div>
        </div>
      ) : filteredBrokers.length === 0 ? (
        <div className="bg-white rounded-lg shadow-sm p-12 text-center">
          <Users className="w-16 h-16 text-gray-400 mx-auto mb-4" />
          <h3 className="text-lg font-semibold text-gray-900 mb-2">
            {searchTerm || filterStatus !== 'all' || filterRole !== 'all'
              ? 'Nenhum corretor encontrado'
              : 'Nenhum corretor cadastrado'}
          </h3>
          <p className="text-gray-600 mb-6">
            {searchTerm || filterStatus !== 'all' || filterRole !== 'all'
              ? 'Tente ajustar os filtros de busca'
              : 'Comece cadastrando seu primeiro corretor'}
          </p>
          {!searchTerm && filterStatus === 'all' && filterRole === 'all' && (
            <button
              onClick={() => router.push('/dashboard/corretores/novo')}
              className="inline-flex items-center gap-2 px-6 py-3 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors"
            >
              <Plus className="w-5 h-5" />
              Cadastrar Corretor
            </button>
          )}
        </div>
      ) : (
        <div className="bg-white rounded-lg shadow-sm overflow-hidden">
          <table className="min-w-full divide-y divide-gray-200">
            <thead className="bg-gray-50">
              <tr>
                <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Corretor
                </th>
                <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Contato
                </th>
                <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  CRECI
                </th>
                <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Imóveis
                </th>
                <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Status
                </th>
                <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Perfil
                </th>
              </tr>
            </thead>
            <tbody className="bg-white divide-y divide-gray-200">
              {filteredBrokers.map((broker) => (
                <tr
                  key={broker.id}
                  onClick={() => router.push(`/dashboard/corretores/${broker.id}`)}
                  className="hover:bg-gray-50 cursor-pointer transition-colors"
                >
                  <td className="px-6 py-4 whitespace-nowrap">
                    <div className="flex items-center">
                      {broker.photo_url ? (
                        <img
                          src={broker.photo_url}
                          alt={broker.name}
                          className="w-10 h-10 rounded-full object-cover"
                        />
                      ) : (
                        <div className="w-10 h-10 rounded-full bg-blue-100 flex items-center justify-center">
                          <span className="text-lg font-bold text-blue-600">
                            {broker.name.charAt(0).toUpperCase()}
                          </span>
                        </div>
                      )}
                      <div className="ml-4">
                        <div className="text-sm font-medium text-gray-900">{broker.name}</div>
                        {broker.experience && broker.experience > 0 && (
                          <div className="text-sm text-gray-500">
                            {broker.experience} {broker.experience === 1 ? 'ano' : 'anos'} de experiência
                          </div>
                        )}
                      </div>
                    </div>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <div className="text-sm text-gray-900">{broker.email}</div>
                    {broker.phone && (
                      <div className="text-sm text-gray-500">{broker.phone}</div>
                    )}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <div className="text-sm text-gray-900">{broker.creci || '-'}</div>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                    {broker.total_listings || 0}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <span
                      className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${
                        broker.is_active
                          ? 'bg-green-100 text-green-800'
                          : 'bg-red-100 text-red-800'
                      }`}
                    >
                      {broker.is_active ? '✓ Ativo' : '✗ Inativo'}
                    </span>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <span
                      className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${getRoleBadgeColor(
                        broker.role
                      )}`}
                    >
                      {getRoleLabel(broker.role)}
                    </span>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}
