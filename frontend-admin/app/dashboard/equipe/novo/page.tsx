'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import { UserRole, STANDARD_PERMISSIONS, getRoleDisplayName } from '@/types/user';
import { ArrowLeft, Send, Shield, Mail, UserPlus } from 'lucide-react';
import { logger } from '@/lib/logger';

export default function InviteUserPage() {
  const router = useRouter();
  const [sending, setSending] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState(false);

  const [formData, setFormData] = useState({
    name: '',
    email: '',
    phone: '',
    role: 'manager' as UserRole,
    creci: '',
    permissions: [] as string[],
  });

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    try {
      setSending(true);
      setError(null);

      const tenantId = localStorage.getItem('tenant_id');
      if (!tenantId) {
        throw new Error('Tenant ID não encontrado');
      }

      const { auth } = await import('@/lib/firebase');
      const currentUser = auth.currentUser;

      if (!currentUser) {
        router.push('/login');
        return
;
      }

      const token = await currentUser.getIdToken(true);
      const apiUrl = process.env.NEXT_PUBLIC_ADMIN_API_URL || `${process.env.NEXT_PUBLIC_API_URL}/admin`;
      const url = `${apiUrl}/${tenantId}/users/invite`;

      // Build payload
      const payload: any = {
        email: formData.email,
        name: formData.name,
        phone: formData.phone || undefined,
        role: formData.role,
      };

      // Add CRECI if role is broker or broker_admin
      if (formData.role === 'broker' || formData.role === 'broker_admin') {
        if (!formData.creci) {
          throw new Error('CRECI é obrigatório para corretores');
        }
        payload.creci = formData.creci;
      }

      // Add permissions for manager role
      if (formData.role === 'manager') {
        payload.permissions = formData.permissions;
      }

      logger.dev('Sending invitation:', payload);

      const response = await fetch(url, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(payload),
      });

      if (!response.ok) {
        const errorData = await response.json().catch(() => ({}));
        throw new Error(errorData.error || 'Erro ao enviar convite');
      }

      const data = await response.json();
      logger.dev('Invitation sent:', data);

      // Success
      setSuccess(true);
      setTimeout(() => {
        router.push('/dashboard/equipe');
      }, 2000);
    } catch (err: any) {
      logger.error('Error sending invitation:', err);
      setError(err.message || 'Erro ao enviar convite');
    } finally {
      setSending(false);
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
        { key: STANDARD_PERMISSIONS.LEAD_CREATE, label: 'Criar leads' },
        { key: STANDARD_PERMISSIONS.LEAD_UPDATE, label: 'Atualizar leads' },
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
        { key: STANDARD_PERMISSIONS.BROKER_CREATE, label: 'Criar corretores' },
        { key: STANDARD_PERMISSIONS.BROKER_UPDATE, label: 'Editar corretores' },
        { key: STANDARD_PERMISSIONS.BROKER_DELETE, label: 'Excluir corretores' },
      ],
    },
    {
      title: 'Anúncios',
      permissions: [
        { key: STANDARD_PERMISSIONS.LISTING_VIEW, label: 'Visualizar anúncios' },
        { key: STANDARD_PERMISSIONS.LISTING_CREATE, label: 'Criar anúncios' },
        { key: STANDARD_PERMISSIONS.LISTING_UPDATE, label: 'Editar anúncios' },
        { key: STANDARD_PERMISSIONS.LISTING_DELETE, label: 'Excluir anúncios' },
      ],
    },
    {
      title: 'Usuários',
      permissions: [
        { key: STANDARD_PERMISSIONS.USER_VIEW, label: 'Visualizar usuários' },
        { key: STANDARD_PERMISSIONS.USER_CREATE, label: 'Criar usuários' },
        { key: STANDARD_PERMISSIONS.USER_UPDATE, label: 'Editar usuários' },
        { key: STANDARD_PERMISSIONS.USER_DELETE, label: 'Excluir usuários' },
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
      title: 'Configurações',
      permissions: [
        { key: STANDARD_PERMISSIONS.SETTINGS_VIEW, label: 'Visualizar configurações' },
        { key: STANDARD_PERMISSIONS.SETTINGS_UPDATE, label: 'Editar configurações' },
      ],
    },
  ];

  if (success) {
    return (
      <div className="container mx-auto px-4 py-6 md:py-8">
        <div className="max-w-2xl mx-auto">
          <div className="bg-green-50 border border-green-200 rounded-lg p-6 text-center">
            <div className="w-16 h-16 bg-green-100 rounded-full flex items-center justify-center mx-auto mb-4">
              <svg
                className="w-8 h-8 text-green-600"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M5 13l4 4L19 7"
                />
              </svg>
            </div>
            <h2 className="text-2xl font-bold text-green-900 mb-2">
              Convite Enviado!
            </h2>
            <p className="text-green-800">
              Um email foi enviado para <strong>{formData.email}</strong> com instruções para aceitar o convite e criar a conta.
            </p>
            <p className="text-sm text-green-700 mt-4">
              Redirecionando para a lista de equipe...
            </p>
          </div>
        </div>
      </div>
    );
  }

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
        <div className="flex items-center gap-3 mb-2">
          <div className="w-12 h-12 bg-purple-100 rounded-lg flex items-center justify-center">
            <Mail className="w-6 h-6 text-purple-600" />
          </div>
          <div>
            <h1 className="text-2xl md:text-3xl lg:text-4xl font-bold text-gray-900">Convidar Membro</h1>
            <p className="text-sm md:text-base text-gray-600 mt-1">
              Envie um convite por email para adicionar um novo membro à equipe
            </p>
          </div>
        </div>
      </div>

      {/* Info Notice */}
      <div className="mb-4 md:mb-6 bg-purple-50 border border-purple-200 rounded-lg p-3 md:p-4">
        <div className="flex items-start gap-3">
          <Shield className="w-5 h-5 text-purple-600 mt-0.5 flex-shrink-0" />
          <div>
            <h3 className="text-sm md:text-base font-semibold text-purple-900 mb-1">
              Como Funciona
            </h3>
            <p className="text-xs md:text-sm text-purple-800">
              O novo membro receberá um email com um link para aceitar o convite e criar sua senha.
              O convite expira em <strong>7 dias</strong>.
            </p>
          </div>
        </div>
      </div>

      {/* Error Message */}
      {error && (
        <div className="mb-4 md:mb-6 bg-red-50 border border-red-200 text-red-700 px-3 md:px-4 py-3 rounded-lg text-sm md:text-base">
          {error}
        </div>
      )}

      {/* Form */}
      <form onSubmit={handleSubmit} className="space-y-4 md:space-y-6">
        {/* Basic Information */}
        <div className="bg-white rounded-lg shadow p-4 md:p-6">
          <h2 className="text-lg md:text-xl font-semibold text-gray-900 mb-4">
            Informações do Convidado
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
                className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-purple-500 focus:border-transparent"
                placeholder="Ex: João Silva"
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
                className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-purple-500 focus:border-transparent"
                placeholder="email@exemplo.com"
                required
              />
              <p className="mt-1 text-xs text-gray-500">
                O convite será enviado para este email
              </p>
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Telefone (opcional)
              </label>
              <input
                type="tel"
                value={formData.phone}
                onChange={(e) => setFormData({ ...formData, phone: e.target.value })}
                className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-purple-500 focus:border-transparent"
                placeholder="(00) 00000-0000"
              />
            </div>
          </div>
        </div>

        {/* Role and Permissions */}
        <div className="bg-white rounded-lg shadow p-4 md:p-6">
          <h2 className="text-lg md:text-xl font-semibold text-gray-900 mb-4">
            Perfil e Acesso
          </h2>
          <div className="mb-4">
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Função *
            </label>
            <select
              value={formData.role}
              onChange={(e) => setFormData({ ...formData, role: e.target.value as UserRole, creci: '' })}
              className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-purple-500 focus:border-transparent"
              required
            >
              <option value="admin">
                {getRoleDisplayName('admin')} - Acesso total ao sistema
              </option>
              <option value="manager">
                {getRoleDisplayName('manager')} - Permissões específicas
              </option>
              <option value="broker">
                {getRoleDisplayName('broker')} - Corretor (requer CRECI)
              </option>
              <option value="broker_admin">
                {getRoleDisplayName('broker_admin')} - Corretor Administrador (requer CRECI)
              </option>
            </select>
          </div>

          {/* CRECI field for broker roles */}
          {(formData.role === 'broker' || formData.role === 'broker_admin') && (
            <div className="mb-4">
              <label className="block text-sm font-medium text-gray-700 mb-1">
                CRECI *
              </label>
              <input
                type="text"
                value={formData.creci}
                onChange={(e) => setFormData({ ...formData, creci: e.target.value })}
                className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-purple-500 focus:border-transparent"
                placeholder="Ex: 12345-F/SP"
                required
              />
              <p className="mt-1 text-xs text-gray-500">
                Formato: XXXXX-F/UF (F para pessoa física)
              </p>
            </div>
          )}
        </div>

        {/* Permissions - Only for Manager */}
        {formData.role === 'manager' && (
          <div className="bg-white rounded-lg shadow p-4 md:p-6">
            <div className="flex items-center justify-between mb-2">
              <h2 className="text-lg md:text-xl font-semibold text-gray-900">
                Permissões
              </h2>
              <div className="flex gap-2">
                <button
                  type="button"
                  onClick={() => {
                    const allPermissions = permissionGroups.flatMap((group) =>
                      group.permissions.map((p) => p.key)
                    );
                    setFormData({ ...formData, permissions: allPermissions });
                  }}
                  className="text-xs md:text-sm px-3 py-1.5 text-purple-600 hover:text-purple-700 hover:bg-purple-50 rounded-lg transition font-medium"
                >
                  Selecionar Todas
                </button>
                <button
                  type="button"
                  onClick={() => setFormData({ ...formData, permissions: [] })}
                  className="text-xs md:text-sm px-3 py-1.5 text-gray-600 hover:text-gray-700 hover:bg-gray-50 rounded-lg transition font-medium"
                >
                  Limpar Seleção
                </button>
              </div>
            </div>
            <p className="text-xs md:text-sm text-gray-600 mb-4">
              Selecione as permissões específicas para este gerente
            </p>
            <div className="space-y-6">
              {permissionGroups.map((group) => {
                const groupPermissions = group.permissions.map((p) => p.key);
                const allSelected = groupPermissions.every((p) =>
                  formData.permissions.includes(p)
                );
                const noneSelected = groupPermissions.every(
                  (p) => !formData.permissions.includes(p)
                );

                return (
                  <div key={group.title}>
                    <div className="flex items-center justify-between mb-2">
                      <h3 className="text-sm font-semibold text-gray-900">
                        {group.title}
                      </h3>
                      <button
                        type="button"
                        onClick={() => {
                          if (allSelected) {
                            // Deselect all from this group
                            setFormData({
                              ...formData,
                              permissions: formData.permissions.filter(
                                (p) => !groupPermissions.includes(p as any)
                              ),
                            });
                          } else {
                            // Select all from this group
                            const newPermissions = [
                              ...formData.permissions.filter(
                                (p) => !groupPermissions.includes(p as any)
                              ),
                              ...groupPermissions,
                            ];
                            setFormData({ ...formData, permissions: newPermissions });
                          }
                        }}
                        className="text-xs px-2 py-1 text-purple-600 hover:text-purple-700 hover:bg-purple-50 rounded transition"
                      >
                        {allSelected ? 'Desmarcar Todas' : 'Marcar Todas'}
                      </button>
                    </div>
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
                            className="w-4 h-4 text-purple-600 border-gray-300 rounded focus:ring-purple-500"
                          />
                          <span className="text-sm text-gray-700">
                            {permission.label}
                          </span>
                        </label>
                      ))}
                    </div>
                  </div>
                );
              })}
            </div>
          </div>
        )}

        {/* Admin/Broker Notice */}
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
                </p>
              </div>
            </div>
          </div>
        )}

        {(formData.role === 'broker' || formData.role === 'broker_admin') && (
          <div className="bg-blue-50 border border-blue-200 rounded-lg p-3 md:p-4">
            <div className="flex items-start gap-3">
              <Shield className="w-5 h-5 text-blue-600 mt-0.5 flex-shrink-0" />
              <div>
                <h3 className="text-sm md:text-base font-semibold text-blue-900 mb-1">
                  Corretor - CRECI Obrigatório
                </h3>
                <p className="text-xs md:text-sm text-blue-800">
                  Corretores precisam de um CRECI válido para atuar na plataforma.
                  {formData.role === 'broker_admin' && ' Corretores administradores têm acesso total ao sistema.'}
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
            disabled={sending}
          >
            Cancelar
          </button>
          <button
            type="submit"
            disabled={sending}
            className="w-full sm:w-auto flex items-center justify-center gap-2 px-6 py-2.5 sm:py-2 bg-gradient-to-r from-purple-600 to-blue-600 text-white rounded-lg hover:from-purple-700 hover:to-blue-700 transition disabled:opacity-50 disabled:cursor-not-allowed font-medium"
          >
            {sending ? (
              <>
                <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-white"></div>
                Enviando convite...
              </>
            ) : (
              <>
                <Send className="w-4 h-4" />
                Enviar Convite
              </>
            )}
          </button>
        </div>
      </form>
    </div>
  );
}
