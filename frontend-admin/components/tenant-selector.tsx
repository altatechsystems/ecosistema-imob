'use client';

import { useState, useEffect, useRef } from 'react';
import { Building2, ChevronDown, Check } from 'lucide-react';

interface Tenant {
  id: string;
  name: string;
  slug: string;
  is_active: boolean;
}

interface TenantSelectorProps {
  isPlatformAdmin: boolean;
}

export function TenantSelector({ isPlatformAdmin }: TenantSelectorProps) {
  const [isOpen, setIsOpen] = useState(false);
  const [tenants, setTenants] = useState<Tenant[]>([]);
  const [currentTenant, setCurrentTenant] = useState<Tenant | null>(null);
  const [loading, setLoading] = useState(false);
  const dropdownRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (isPlatformAdmin) {
      fetchTenants();
    }
  }, [isPlatformAdmin]);

  useEffect(() => {
    // Close dropdown when clicking outside
    function handleClickOutside(event: MouseEvent) {
      if (dropdownRef.current && !dropdownRef.current.contains(event.target as Node)) {
        setIsOpen(false);
      }
    }

    if (isOpen) {
      document.addEventListener('mousedown', handleClickOutside);
      return () => document.removeEventListener('mousedown', handleClickOutside);
    }
  }, [isOpen]);

  const fetchTenants = async () => {
    try {
      setLoading(true);
      const response = await fetch(`${process.env.NEXT_PUBLIC_API_URL?.replace('/api/v1', '')}/tenants`);

      if (!response.ok) {
        throw new Error('Erro ao buscar tenants');
      }

      const data = await response.json();
      const tenantsList = data.data || [];
      setTenants(tenantsList);

      // Set current tenant from localStorage
      const currentTenantId = localStorage.getItem('tenant_id');
      if (currentTenantId) {
        const current = tenantsList.find((t: Tenant) => t.id === currentTenantId);
        setCurrentTenant(current || null);
      }
    } catch (error) {
      // Error fetching tenants - silently fail
    } finally {
      setLoading(false);
    }
  };

  const handleTenantChange = (tenant: Tenant) => {
    // Update localStorage
    localStorage.setItem('tenant_id', tenant.id);
    localStorage.setItem('tenant_name', tenant.name);

    // Update state
    setCurrentTenant(tenant);
    setIsOpen(false);

    // Reload the page to refresh all data with new tenant context
    window.location.reload();
  };

  // Don't show selector if not platform admin
  if (!isPlatformAdmin) {
    return null;
  }

  return (
    <div className="relative" ref={dropdownRef}>
      <button
        onClick={() => setIsOpen(!isOpen)}
        className="flex items-center gap-2 px-3 py-2 bg-white border border-gray-300 rounded-lg hover:bg-gray-50 transition-colors"
      >
        <Building2 className="w-4 h-4 text-gray-600" />
        <div className="text-left">
          <p className="text-xs text-gray-500">Tenant</p>
          <p className="text-sm font-medium text-gray-900">
            {loading ? 'Carregando...' : currentTenant?.name || 'Selecione'}
          </p>
        </div>
        <ChevronDown className={`w-4 h-4 text-gray-600 transition-transform ${isOpen ? 'rotate-180' : ''}`} />
      </button>

      {isOpen && (
        <div className="absolute top-full right-0 mt-2 w-72 bg-white border border-gray-200 rounded-lg shadow-lg z-50">
          <div className="p-2">
            <div className="px-3 py-2 text-xs font-semibold text-gray-500 uppercase">
              Selecionar Tenant
            </div>
            <div className="max-h-96 overflow-y-auto">
              {tenants.length === 0 ? (
                <div className="px-3 py-4 text-sm text-gray-500 text-center">
                  Nenhum tenant disponível
                </div>
              ) : (
                tenants.map((tenant) => (
                  <button
                    key={tenant.id}
                    onClick={() => handleTenantChange(tenant)}
                    className={`w-full flex items-center justify-between px-3 py-2 text-left rounded-lg transition-colors ${
                      currentTenant?.id === tenant.id
                        ? 'bg-blue-50 text-blue-900'
                        : 'hover:bg-gray-50 text-gray-900'
                    }`}
                  >
                    <div className="flex-1">
                      <p className="text-sm font-medium">{tenant.name}</p>
                      <p className="text-xs text-gray-500">{tenant.slug}</p>
                    </div>
                    {currentTenant?.id === tenant.id && (
                      <Check className="w-4 h-4 text-blue-600" />
                    )}
                    {!tenant.is_active && (
                      <span className="ml-2 px-2 py-0.5 bg-red-100 text-red-600 text-xs rounded-full">
                        Inativo
                      </span>
                    )}
                  </button>
                ))
              )}
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
