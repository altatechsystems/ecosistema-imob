'use client';

import { useState, useEffect } from 'react';
import { useRouter, useParams } from 'next/navigation';
import { User, UserRole, STANDARD_PERMISSIONS, getRoleDisplayName } from '@/types/user';
import { ArrowLeft, Save, Shield, UserCog } from 'lucide-react';

export default function EditUserPage() {
  const router = useRouter();
  const params = useParams();
  const userId = params.id as string;

  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const [formData, setFormData] = useState({
    name: '',
    email: '',
    phone: '',
    document: '',
    document_type: 'cpf' as 'cpf' | 'cnpj',
    role: 'manager' as UserRole,
    is_active: true,
    permissions: [] as string[],
  });

  useEffect(() => {
    loadUser();
  }, [userId]);

  const loadUser = async () => {
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
      const currentUser = auth.currentUser;

      if (!currentUser) {
        router.push('/login');
        return;
      }

      const token = await currentUser.getIdToken(true);
      const url = `${process.env.NEXT_PUBLIC_API_URL}/admin/${tenantId}/users/${userId}`;

      const response = await fetch(url, {
        headers: {
          'Authorization': `Bearer ${token}`,
        },
      });

      if (!response.ok) {
        const errorData = await response.json().catch(() => ({}));
        throw new Error(errorData.error || 'Erro ao buscar usuário');
      }

      const userData: User = await response.json();
      setUser(userData);
      setFormData({
        name: userData.name,
        email: userData.email,
        phone: userData.phone || '',
        document: userData.document || '',
        document_type: userData.document_type || 'cpf',
        role: userData.role,
        is_active: userData.is_active,
        permissions: userData.permissions || [],
      });
    } catch (err: any) {
      console.error('Error loading user:', err);
      setError(err.message || 'Erro ao carregar usuário');
    } finally {
      setLoading(false);
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    try {
      setSaving(true);
      setError(null);

      const tenantId = localStorage.getItem('tenant_id');
      if (!tenantId) {
        throw new Error('Tenant ID não encontrado');
      }

      const { auth } = await import('@/lib/firebase');
      const currentUser = auth.currentUser;

      if (!currentUser) {
        router.push('/login');
        return;
      }

      const token = await currentUser.getIdToken(true);
      const url = `${process.env.NEXT_PUBLIC_API_URL}/admin/${tenantId}/users/${userId}`;

      const response = await fetch(url, {
        method: 'PUT',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(formData),
      });

      if (!response.ok) {
        const errorData = await response.json().catch(() => ({}));
        throw new Error(errorData.error || 'Erro ao atualizar usuário');
      }

      router.push('/dashboard/equipe');
    } catch (err: any) {
      console.error('Error updating user:', err);
      setError(err.message || 'Erro ao atualizar usuário');
    } finally {
      setSaving(false);
    }
  };

  const handlePermissionToggle = (permission: string) => {
    setFormData((prev) => {
      const permissions = prev.permissions.includes(permission)
        ? prev.permissions.filter((p) => p !== permission)
        : [...prev.permissions, permission];
      return { ...prev, permissions };
    });
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto"></div>
          <p className="mt-4 text-gray-600">Carregando usuário...</p>
        </div>
      </div>
    );
  }

  if (error && !user) {
    return (
      <div className="container mx-auto px-4 py-8">
        <div className="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-lg">
          {error}
        </div>
        <button
          onClick={() => router.push('/dashboard/equipe')}
          className="mt-4 flex items-center gap-2 text-blue-600 hover:text-blue-800"
        >
          <ArrowLeft className="w-4 h-4" />
          Voltar para Equipe
        </button>
      </div>
    );
  }

  const permissionGroups = [
    {
      title: 'Propriedades',
      permissions: [
        { key: STANDARD_PERMISSIONS.PROPERTY_VIEW, label: 'Visualizar imóveis' },
        { key: STANDARD_PERMISSIONS.PROPERTY_CREATE, label: 'Criar imóveis' },
        { key: STANDARD_PERMISSIONS.PROPERTY_UPDATE, label: 'Editar imóveis' },
        { key: STANDARD_PERMISSIONS.PROPERTY_DELETE, label: 'Excluir imóveis' },
      ],
    },
    {
      title: 'Leads',
      permissions: [
        { key: STANDARD_PERMISSIONS.LEAD_VIEW, label: 'Visualizar leads' },
        { key: STANDARD_PERMISSIONS.LEAD_MANAGE, label: 'Gerenciar leads' },
        { key: STANDARD_PERMISSIONS.LEAD_ASSIGN, label: 'Atribuir leads' },
        { key: STANDARD_PERMISSIONS.LEAD_DELETE, label: 'Excluir leads' },
      ],
    },
    {
      title: 'Proprietários',
      permissions: [
        { key: STANDARD_PERMISSIONS.OWNER_VIEW, label: 'Visualizar proprietários' },
        { key: STANDARD_PERMISSIONS.OWNER_CREATE, label: 'Criar proprietários' },
        { key: STANDARD_PERMISSIONS.OWNER_UPDATE, label: 'Editar proprietários' },
        { key: STANDARD_PERMISSIONS.OWNER_DELETE, label: 'Excluir proprietários' },
      ],
    },
    {
      title: 'Corretores',
      permissions: [
        { key: STANDARD_PERMISSIONS.BROKER_VIEW, label: 'Visualizar corretores' },
        { key: STANDARD_PERMISSIONS.BROKER_MANAGE, label: 'Gerenciar corretores' },
      ],
    },
    {
      title: 'Relatórios',
      permissions: [
        { key: STANDARD_PERMISSIONS.REPORT_VIEW, label: 'Visualizar relatórios' },
        { key: STANDARD_PERMISSIONS.REPORT_EXPORT, label: 'Exportar relatórios' },
      ],
    },
    {
      title: 'Sistema',
      permissions: [
        { key: STANDARD_PERMISSIONS.SETTINGS_VIEW, label: 'Visualizar configurações' },
        { key: STANDARD_PERMISSIONS.SETTINGS_EDIT, label: 'Editar configurações' },
        { key: STANDARD_PERMISSIONS.USER_MANAGE, label: 'Gerenciar usuários' },
        { key: STANDARD_PERMISSIONS.IMPORT_DATA, label: 'Importar dados' },
      ],
    },
  ];

  return (
    <div className="container mx-auto px-4 py-6 md:py-8">
      {/* Header */}
      <div className="mb-6 md:mb-8">
        <button
          onClick={() => router.push('/dashboard/equipe')}
          className="flex items-center gap-2 text-gray-600 hover:text-gray-900 mb-4 text-sm md:text-base"
        >
          <ArrowLeft className="w-4 h-4" />
          Voltar para Equipe
        </button>
        <h1 className="text-2xl md:text-3xl lg:text-4xl font-bold text-gray-900">Editar Usuário</h1>
        <p className="text-sm md:text-base text-gray-600 mt-1">
          Atualize as informações e permissões do usuário administrativo
        </p>
      </div>

      {/* Error Message */}
      {error && (
        <div className="mb-4 md:mb-6 bg-red-50 border border-red-200 text-red-700 px-3 md:px-4 py-3 rounded-lg text-sm md:text-base">
          {error}
        </div>
      )}

      {/* Form */}
      <form onSubmit={handleSubmit} className="space-y-4 md:space-y-6">
        {/* Basic Information Card */}
        <div className="bg-white rounded-lg shadow p-4 md:p-6">
          <h2 className="text-lg md:text-xl font-semibold text-gray-900 mb-4">
            Informações Básicas
          </h2>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Nome Completo *
              </label>
              <input
                type="text"
                value={formData.name}
                onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                required
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Email *
              </label>
              <input
                type="email"
                value={formData.email}
                onChange={(e) => setFormData({ ...formData, email: e.target.value })}
                className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                required
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Telefone
              </label>
              <input
                type="tel"
                value={formData.phone}
                onChange={(e) => setFormData({ ...formData, phone: e.target.value })}
                className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Documento
              </label>
              <input
                type="text"
                value={formData.document}
                onChange={(e) => setFormData({ ...formData, document: e.target.value })}
                className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
              />
            </div>
          </div>
        </div>

        {/* Role and Status Card */}
        <div className="bg-white rounded-lg shadow p-4 md:p-6">
          <h2 className="text-lg md:text-xl font-semibold text-gray-900 mb-4">
            Perfil e Status
          </h2>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Perfil *
              </label>
              <select
                value={formData.role}
                onChange={(e) => setFormData({ ...formData, role: e.target.value as UserRole })}
                className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                required
              >
                <option value="admin">
                  {getRoleDisplayName('admin')} - Acesso total
                </option>
                <option value="manager">
                  {getRoleDisplayName('manager')} - Permissões específicas
                </option>
              </select>
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Status
              </label>
              <div className="flex items-center gap-4 pt-2">
                <label className="flex items-center gap-2 cursor-pointer">
                  <input
                    type="radio"
                    checked={formData.is_active}
                    onChange={() => setFormData({ ...formData, is_active: true })}
                    className="w-4 h-4 text-blue-600"
                  />
                  <span className="text-sm text-gray-700">Ativo</span>
                </label>
                <label className="flex items-center gap-2 cursor-pointer">
                  <input
                    type="radio"
                    checked={!formData.is_active}
                    onChange={() => setFormData({ ...formData, is_active: false })}
                    className="w-4 h-4 text-blue-600"
                  />
                  <span className="text-sm text-gray-700">Inativo</span>
                </label>
              </div>
            </div>
          </div>
        </div>

        {/* Permissions Card - Only for Manager role */}
        {formData.role === 'manager' && (
          <div className="bg-white rounded-lg shadow p-4 md:p-6">
            <h2 className="text-lg md:text-xl font-semibold text-gray-900 mb-2">
              Permissões
            </h2>
            <p className="text-xs md:text-sm text-gray-600 mb-4">
              Selecione as permissões específicas para este gerente
            </p>
            <div className="space-y-6">
              {permissionGroups.map((group) => (
                <div key={group.title}>
                  <h3 className="text-sm font-semibold text-gray-900 mb-2">
                    {group.title}
                  </h3>
                  <div className="grid grid-cols-1 md:grid-cols-2 gap-2">
                    {group.permissions.map((permission) => (
                      <label
                        key={permission.key}
                        className="flex items-center gap-2 cursor-pointer hover:bg-gray-50 p-2 rounded"
                      >
                        <input
                          type="checkbox"
                          checked={formData.permissions.includes(permission.key)}
                          onChange={() => handlePermissionToggle(permission.key)}
                          className="w-4 h-4 text-blue-600 border-gray-300 rounded focus:ring-blue-500"
                        />
                        <span className="text-sm text-gray-700">
                          {permission.label}
                        </span>
                      </label>
                    ))}
                  </div>
                </div>
              ))}
            </div>
          </div>
        )}

        {/* Admin Notice */}
        {formData.role === 'admin' && (
          <div className="bg-purple-50 border border-purple-200 rounded-lg p-3 md:p-4">
            <div className="flex items-start gap-3">
              <Shield className="w-5 h-5 text-purple-600 mt-0.5 flex-shrink-0" />
              <div>
                <h3 className="text-sm md:text-base font-semibold text-purple-900 mb-1">
                  Administrador - Acesso Total
                </h3>
                <p className="text-xs md:text-sm text-purple-800">
                  Administradores têm acesso irrestrito a todas as funcionalidades do sistema.
                  Não é necessário configurar permissões individuais.
                </p>
              </div>
            </div>
          </div>
        )}

        {/* Action Buttons */}
        <div className="flex flex-col sm:flex-row sm:items-center sm:justify-end gap-3 sm:gap-4 pt-4 md:pt-6">
          <button
            type="button"
            onClick={() => router.push('/dashboard/equipe')}
            className="w-full sm:w-auto px-6 py-2.5 sm:py-2 border border-gray-300 text-gray-700 rounded-lg hover:bg-gray-50 transition font-medium"
            disabled={saving}
          >
            Cancelar
          </button>
          <button
            type="submit"
            disabled={saving}
            className="w-full sm:w-auto flex items-center justify-center gap-2 px-6 py-2.5 sm:py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition disabled:opacity-50 disabled:cursor-not-allowed font-medium"
          >
            {saving ? (
              <>
                <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-white"></div>
                Salvando...
              </>
            ) : (
              <>
                <Save className="w-4 h-4" />
                Salvar Alterações
              </>
            )}
          </button>
        </div>
      </form>
    </div>
  );
}
