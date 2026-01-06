'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import api from '@/lib/api';
import { User, getRoleDisplayName, hasPermission, STANDARD_PERMISSIONS } from '@/types/user';
import { Plus, UserCog, Trash2, Shield, ShieldOff, MoreVertical, Mail, Phone, FileText } from 'lucide-react';

export default function EquipePage() {
  const router = useRouter();
  const [users, setUsers] = useState<User[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [showActiveOnly, setShowActiveOnly] = useState(false);

  useEffect(() => {
    loadUsers();
  }, [showActiveOnly]);

  const loadUsers = async () => {
    try {
      setLoading(true);
      setError(null);

      const tenantId = localStorage.getItem('tenant_id');
      if (!tenantId) {
        setError('Tenant ID não encontrado');
        setLoading(false);
        return;
      }

      const { auth } = await import('@/lib/firebase');
      const user = auth.currentUser;

      if (!user) {
        console.error('Usuário não autenticado');
        router.push('/login');
        return;
      }

      const token = await user.getIdToken(true);
      const params = new URLSearchParams();
      if (showActiveOnly) params.append('active', 'true');

      const url = `${process.env.NEXT_PUBLIC_API_URL}/admin/${tenantId}/users${params.toString() ? '?' + params.toString() : ''}`;

      const response = await fetch(url, {
        headers: {
          'Authorization': `Bearer ${token}`,
        },
      });

      if (!response.ok) {
        const errorData = await response.json().catch(() => ({}));
        throw new Error(errorData.error || 'Erro ao buscar usuários');
      }

      const data = await response.json();
      setUsers(data || []);
    } catch (err: any) {
      console.error('Error loading users:', err);
      setError(err.message || 'Erro ao carregar equipe');
    } finally {
      setLoading(false);
    }
  };

  const handleDeleteUser = async (userId: string) => {
    if (!confirm('Tem certeza que deseja excluir este usuário?')) {
      return;
    }

    try {
      const tenantId = localStorage.getItem('tenant_id');
      if (!tenantId) {
        alert('Tenant ID não encontrado');
        return;
      }

      const { auth } = await import('@/lib/firebase');
      const user = auth.currentUser;

      if (!user) {
        router.push('/login');
        return;
      }

      const token = await user.getIdToken(true);
      const url = `${process.env.NEXT_PUBLIC_API_URL}/admin/${tenantId}/users/${userId}`;

      const response = await fetch(url, {
        method: 'DELETE',
        headers: {
          'Authorization': `Bearer ${token}`,
        },
      });

      if (!response.ok) {
        const errorData = await response.json().catch(() => ({}));
        throw new Error(errorData.error || 'Erro ao excluir usuário');
      }

      await loadUsers();
    } catch (err: any) {
      console.error('Error deleting user:', err);
      alert(err.message || 'Erro ao excluir usuário');
    }
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto"></div>
          <p className="mt-4 text-gray-600">Carregando equipe...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="container mx-auto px-4 py-6 md:py-8">
      {/* Header */}
      <div className="mb-6 md:mb-8">
        <div className="flex flex-col md:flex-row md:justify-between md:items-center gap-4 mb-4">
          <div>
            <h1 className="text-2xl md:text-3xl lg:text-4xl font-bold text-gray-900">Equipe</h1>
            <p className="text-sm md:text-base text-gray-600 mt-1">
              Gerencie usuários administrativos da plataforma
            </p>
          </div>
          <button
            onClick={() => router.push('/dashboard/equipe/novo')}
            className="flex items-center justify-center md:justify-start gap-2 bg-blue-600 text-white px-4 py-2.5 md:py-2 rounded-lg hover:bg-blue-700 transition w-full md:w-auto"
          >
            <Plus className="w-5 h-5" />
            <span className="font-medium">Novo Usuário</span>
          </button>
        </div>

        {/* Info Box */}
        <div className="bg-blue-50 border border-blue-200 rounded-lg p-3 md:p-4">
          <div className="flex items-start gap-3">
            <Shield className="w-5 h-5 text-blue-600 mt-0.5 flex-shrink-0" />
            <div className="flex-1">
              <h3 className="text-sm md:text-base font-semibold text-blue-900 mb-1">
                Sobre os Usuários Administrativos
              </h3>
              <p className="text-xs md:text-sm text-blue-800">
                Esta página lista apenas <strong>usuários administrativos</strong> (sem CRECI).
                Para gerenciar <strong>corretores</strong> (com CRECI obrigatório), acesse a página{' '}
                <a href="/dashboard/corretores" className="underline hover:text-blue-900 font-semibold">
                  Corretores
                </a>.
              </p>
            </div>
          </div>
        </div>
      </div>

      {/* Filters */}
      <div className="mb-4 md:mb-6 flex flex-col sm:flex-row sm:items-center gap-3 sm:gap-4">
        <label className="flex items-center gap-2 cursor-pointer">
          <input
            type="checkbox"
            checked={showActiveOnly}
            onChange={(e) => setShowActiveOnly(e.target.checked)}
            className="w-4 h-4 text-blue-600 border-gray-300 rounded focus:ring-blue-500"
          />
          <span className="text-sm text-gray-700">Mostrar apenas ativos</span>
        </label>
        <div className="text-sm text-gray-600">
          Total: <span className="font-semibold">{users.length}</span> usuário(s)
        </div>
      </div>

      {/* Error Message */}
      {error && (
        <div className="mb-4 md:mb-6 bg-red-50 border border-red-200 text-red-700 px-3 md:px-4 py-3 rounded-lg text-sm md:text-base">
          {error}
        </div>
      )}

      {/* Users List */}
      {users.length === 0 ? (
        <div className="text-center py-12 bg-gray-50 rounded-lg">
          <UserCog className="w-12 h-12 md:w-16 md:h-16 text-gray-400 mx-auto mb-4" />
          <h3 className="text-base md:text-lg font-semibold text-gray-900 mb-2">
            Nenhum usuário encontrado
          </h3>
          <p className="text-sm md:text-base text-gray-600 mb-4 px-4">
            {showActiveOnly
              ? 'Não há usuários ativos no momento.'
              : 'Comece adicionando o primeiro usuário da equipe.'}
          </p>
          <button
            onClick={() => router.push('/dashboard/equipe/novo')}
            className="inline-flex items-center gap-2 bg-blue-600 text-white px-4 py-2 rounded-lg hover:bg-blue-700 transition"
          >
            <Plus className="w-5 h-5" />
            Adicionar Usuário
          </button>
        </div>
      ) : (
        <>
          {/* Mobile Card View */}
          <div className="md:hidden space-y-4">
            {users.map((user) => (
              <div key={user.id} className="bg-white rounded-lg shadow-sm border border-gray-200 p-4">
                {/* User Header */}
                <div className="flex items-start gap-3 mb-3">
                  <div className="flex-shrink-0">
                    {user.photo_url ? (
                      <img
                        className="h-12 w-12 rounded-full object-cover"
                        src={user.photo_url}
                        alt={user.name}
                      />
                    ) : (
                      <div className="h-12 w-12 rounded-full bg-gray-200 flex items-center justify-center">
                        <span className="text-gray-600 font-medium text-lg">
                          {user.name.charAt(0).toUpperCase()}
                        </span>
                      </div>
                    )}
                  </div>
                  <div className="flex-1 min-w-0">
                    <h3 className="text-base font-semibold text-gray-900 truncate">
                      {user.name}
                    </h3>
                    <div className="flex items-center gap-2 mt-1">
                      <span
                        className={`inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-xs font-medium ${
                          user.role === 'admin'
                            ? 'bg-purple-100 text-purple-800'
                            : 'bg-blue-100 text-blue-800'
                        }`}
                      >
                        {user.role === 'admin' ? (
                          <Shield className="w-3 h-3" />
                        ) : (
                          <UserCog className="w-3 h-3" />
                        )}
                        {getRoleDisplayName(user.role)}
                      </span>
                      <span
                        className={`inline-flex px-2 py-0.5 text-xs leading-5 font-semibold rounded-full ${
                          user.is_active
                            ? 'bg-green-100 text-green-800'
                            : 'bg-red-100 text-red-800'
                        }`}
                      >
                        {user.is_active ? 'Ativo' : 'Inativo'}
                      </span>
                    </div>
                  </div>
                </div>

                {/* User Details */}
                <div className="space-y-2 mb-3">
                  <div className="flex items-start gap-2 text-sm">
                    <Mail className="w-4 h-4 text-gray-400 mt-0.5 flex-shrink-0" />
                    <span className="text-gray-700 break-all">{user.email}</span>
                  </div>
                  {user.phone && (
                    <div className="flex items-center gap-2 text-sm">
                      <Phone className="w-4 h-4 text-gray-400 flex-shrink-0" />
                      <span className="text-gray-700">{user.phone}</span>
                    </div>
                  )}
                  {user.document && (
                    <div className="flex items-center gap-2 text-sm">
                      <FileText className="w-4 h-4 text-gray-400 flex-shrink-0" />
                      <span className="text-gray-700">{user.document}</span>
                    </div>
                  )}
                  <div className="flex items-center gap-2 text-sm pt-1 border-t border-gray-100">
                    <Shield className="w-4 h-4 text-gray-400 flex-shrink-0" />
                    {user.role === 'admin' ? (
                      <span className="text-purple-600 font-medium">Acesso total</span>
                    ) : (
                      <span className="text-gray-600">
                        {user.permissions?.length || 0} permissão(ões)
                      </span>
                    )}
                  </div>
                </div>

                {/* Actions */}
                <div className="flex gap-2 pt-3 border-t border-gray-100">
                  <button
                    onClick={() => router.push(`/dashboard/equipe/${user.id}`)}
                    className="flex-1 bg-blue-50 text-blue-700 px-4 py-2 rounded-lg hover:bg-blue-100 transition font-medium text-sm"
                  >
                    Editar
                  </button>
                  <button
                    onClick={() => handleDeleteUser(user.id)}
                    className="flex-1 bg-red-50 text-red-700 px-4 py-2 rounded-lg hover:bg-red-100 transition font-medium text-sm"
                  >
                    Excluir
                  </button>
                </div>
              </div>
            ))}
          </div>

          {/* Desktop Table View */}
          <div className="hidden md:block bg-white rounded-lg shadow overflow-hidden">
            <div className="overflow-x-auto">
              <table className="min-w-full divide-y divide-gray-200">
                <thead className="bg-gray-50">
                  <tr>
                    <th className="px-4 lg:px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Usuário
                    </th>
                    <th className="px-4 lg:px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Contato
                    </th>
                    <th className="px-4 lg:px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Perfil
                    </th>
                    <th className="px-4 lg:px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Permissões
                    </th>
                    <th className="px-4 lg:px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Status
                    </th>
                    <th className="px-4 lg:px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Ações
                    </th>
                  </tr>
                </thead>
                <tbody className="bg-white divide-y divide-gray-200">
                  {users.map((user) => (
                    <tr key={user.id} className="hover:bg-gray-50">
                      <td className="px-4 lg:px-6 py-4 whitespace-nowrap">
                        <div className="flex items-center">
                          <div className="flex-shrink-0 h-10 w-10">
                            {user.photo_url ? (
                              <img
                                className="h-10 w-10 rounded-full object-cover"
                                src={user.photo_url}
                                alt={user.name}
                              />
                            ) : (
                              <div className="h-10 w-10 rounded-full bg-gray-200 flex items-center justify-center">
                                <span className="text-gray-600 font-medium">
                                  {user.name.charAt(0).toUpperCase()}
                                </span>
                              </div>
                            )}
                          </div>
                          <div className="ml-4">
                            <div className="text-sm font-medium text-gray-900">
                              {user.name}
                            </div>
                            <div className="text-sm text-gray-500">
                              {user.document || 'Sem documento'}
                            </div>
                          </div>
                        </div>
                      </td>
                      <td className="px-4 lg:px-6 py-4 whitespace-nowrap">
                        <div className="text-sm text-gray-900">{user.email}</div>
                        <div className="text-sm text-gray-500">
                          {user.phone || 'Sem telefone'}
                        </div>
                      </td>
                      <td className="px-4 lg:px-6 py-4 whitespace-nowrap">
                        <span
                          className={`inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-xs font-medium ${
                            user.role === 'admin'
                              ? 'bg-purple-100 text-purple-800'
                              : 'bg-blue-100 text-blue-800'
                          }`}
                        >
                          {user.role === 'admin' ? (
                            <Shield className="w-3 h-3" />
                          ) : (
                            <UserCog className="w-3 h-3" />
                          )}
                          {getRoleDisplayName(user.role)}
                        </span>
                      </td>
                      <td className="px-4 lg:px-6 py-4">
                        <div className="text-sm text-gray-900">
                          {user.role === 'admin' ? (
                            <span className="text-purple-600 font-medium">
                              Acesso total
                            </span>
                          ) : (
                            <span className="text-gray-600">
                              {user.permissions?.length || 0} permissão(ões)
                            </span>
                          )}
                        </div>
                      </td>
                      <td className="px-4 lg:px-6 py-4 whitespace-nowrap">
                        <span
                          className={`inline-flex px-2 text-xs leading-5 font-semibold rounded-full ${
                            user.is_active
                              ? 'bg-green-100 text-green-800'
                              : 'bg-red-100 text-red-800'
                          }`}
                        >
                          {user.is_active ? 'Ativo' : 'Inativo'}
                        </span>
                      </td>
                      <td className="px-4 lg:px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                        <div className="flex items-center justify-end gap-2">
                          <button
                            onClick={() => router.push(`/dashboard/equipe/${user.id}`)}
                            className="text-blue-600 hover:text-blue-900 font-medium"
                          >
                            Editar
                          </button>
                          <button
                            onClick={() => handleDeleteUser(user.id)}
                            className="text-red-600 hover:text-red-900"
                          >
                            <Trash2 className="w-4 h-4" />
                          </button>
                        </div>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        </>
      )}
    </div>
  );
}
