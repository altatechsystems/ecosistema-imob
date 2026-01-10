'use client';

import { useEffect, useState, useRef, useCallback } from 'react';
import { adminApi } from '@/lib/api';
import { useAuth } from '@/hooks/use-auth';
import { useTenant } from '@/contexts/tenant-context';
import { formatCurrency, formatNumber } from '@/lib/utils';
import { logger } from '@/lib/logger';
import {
  Building2,
  MessageSquare,
  Users,
  TrendingUp,
  Eye,
  CheckCircle,
  Upload,
} from 'lucide-react';

interface DashboardMetrics {
  total_properties: number;
  available_properties: number;
  sold_properties: number;
  rented_properties: number;
  total_leads: number;
  new_leads: number;
  converted_leads: number;
  total_owners: number;
  total_brokers: number;
  total_value?: number;
}

export default function DashboardPage() {
  const { effectiveTenantId, selectedTenantId } = useTenant();
  const [metrics, setMetrics] = useState<DashboardMetrics | null>(null);
  const [loading, setLoading] = useState(true);

  // Use ref to track loading state to prevent race conditions
  const isLoadingRef = useRef(false);
  const abortControllerRef = useRef<AbortController | null>(null);

  const loadMetrics = useCallback(async () => {
    if (!effectiveTenantId) {
      logger.error('Tenant ID não encontrado no contexto');
      setLoading(false);
      return;
    }

    // Se for tenant_master e não tiver tenant selecionado, não buscar métricas
    // (tenant_master não tem imóveis próprios)
    if (effectiveTenantId === 'tenant_master' && !selectedTenantId) {
      logger.dev('[Dashboard] tenant_master selected but no specific tenant chosen, skipping metrics load');
      setMetrics({
        total_properties: 0,
        available_properties: 0,
        sold_properties: 0,
        rented_properties: 0,
        total_leads: 0,
        new_leads: 0,
        converted_leads: 0,
        total_owners: 0,
        total_brokers: 0,
      });
      setLoading(false);
      return;
    }

    // Evitar múltiplas chamadas simultâneas usando ref
    if (isLoadingRef.current) {
      logger.dev('[Dashboard] Already loading metrics, skipping...');
      return;
    }

    // Cancelar requisição anterior se existir
    if (abortControllerRef.current) {
      abortControllerRef.current.abort();
    }

    // Criar novo AbortController
    const abortController = new AbortController();
    abortControllerRef.current = abortController;

    try {
      isLoadingRef.current = true;
      setLoading(true);
      logger.dev('[Dashboard] Using tenant_id:', effectiveTenantId);

      // Fazer todas as requests em PARALELO para evitar erro 429
      const [propertiesData, availableData, soldData, rentedData] = await Promise.all([
        fetch(`${process.env.NEXT_PUBLIC_API_URL}/${effectiveTenantId}/properties?limit=1`, {
          signal: abortController.signal
        }).then(r => r.ok ? r.json() : { total: 0 }),
        fetch(`${process.env.NEXT_PUBLIC_API_URL}/${effectiveTenantId}/properties?status=available&limit=1`, {
          signal: abortController.signal
        }).then(r => r.ok ? r.json() : { total: 0 }),
        fetch(`${process.env.NEXT_PUBLIC_API_URL}/${effectiveTenantId}/properties?status=sold&limit=1`, {
          signal: abortController.signal
        }).then(r => r.ok ? r.json() : { total: 0 }),
        fetch(`${process.env.NEXT_PUBLIC_API_URL}/${effectiveTenantId}/properties?status=rented&limit=1`, {
          signal: abortController.signal
        }).then(r => r.ok ? r.json() : { total: 0 }),
      ]);

      // Verificar se não foi abortado
      if (abortController.signal.aborted) {
        return;
      }

      logger.dev('[Dashboard] Received data:', {
        properties: propertiesData.total,
        available: availableData.total,
        sold: soldData.total,
        rented: rentedData.total
      });

      setMetrics({
        total_properties: propertiesData.total || 0,
        available_properties: availableData.total || 0,
        sold_properties: soldData.total || 0,
        rented_properties: rentedData.total || 0,
        total_leads: 0, // TODO: Implementar quando tiver endpoint de leads
        new_leads: 0,
        converted_leads: 0,
        total_owners: 0,
        total_brokers: 0,
      });
    } catch (error) {
      // Ignorar erros de abort
      if (error instanceof Error && error.name === 'AbortError') {
        logger.dev('[Dashboard] Request aborted');
        return;
      }
      logger.error('Failed to load metrics:', error);
    } finally {
      isLoadingRef.current = false;
      setLoading(false);
      abortControllerRef.current = null;
    }
  }, [effectiveTenantId]);

  useEffect(() => {
    // Debounce: aguardar um pequeno delay antes de carregar métricas
    // Isso evita múltiplas requisições quando o componente renderiza várias vezes rapidamente (React Strict Mode)
    const timeoutId = setTimeout(() => {
      if (effectiveTenantId) {
        loadMetrics();
      }
    }, 150); // 150ms de delay para evitar requisições duplicadas

    // Cleanup: cancelar timeout e requisições pendentes quando o componente for desmontado ou tenant mudar
    return () => {
      clearTimeout(timeoutId);
      if (abortControllerRef.current) {
        abortControllerRef.current.abort();
        abortControllerRef.current = null;
      }
      isLoadingRef.current = false;
    };
  }, [effectiveTenantId, loadMetrics]);

  if (loading) {
    return (
      <div className="flex items-center justify-center h-full">
        <div className="text-center">
          <div className="w-16 h-16 border-4 border-blue-600 border-t-transparent rounded-full animate-spin mx-auto mb-4"></div>
          <p className="text-gray-600">Carregando métricas...</p>
        </div>
      </div>
    );
  }

  const stats = [
    {
      title: 'Total de Imóveis',
      value: metrics?.total_properties || 0,
      icon: Building2,
      color: 'bg-blue-500',
      change: '+12%',
    },
    {
      title: 'Imóveis Disponíveis',
      value: metrics?.available_properties || 0,
      icon: Eye,
      color: 'bg-green-500',
      change: '+8%',
    },
    {
      title: 'Leads Total',
      value: metrics?.total_leads || 0,
      icon: MessageSquare,
      color: 'bg-purple-500',
      change: '+24%',
    },
    {
      title: 'Leads Novos',
      value: metrics?.new_leads || 0,
      icon: TrendingUp,
      color: 'bg-orange-500',
      change: '+15%',
    },
    {
      title: 'Proprietários',
      value: metrics?.total_owners || 0,
      icon: Users,
      color: 'bg-indigo-500',
      change: '+5%',
    },
    {
      title: 'Negócios Fechados',
      value: (metrics?.sold_properties || 0) + (metrics?.rented_properties || 0),
      icon: CheckCircle,
      color: 'bg-teal-500',
      change: '+18%',
    },
  ];

  return (
    <div className="space-y-6">
      {/* Page Header */}
      <div>
        <h1 className="text-3xl font-bold text-gray-900">Dashboard</h1>
        <p className="text-gray-600 mt-1">
          Visão geral do seu negócio imobiliário
        </p>
      </div>

      {/* Stats Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {stats.map((stat) => {
          const Icon = stat.icon;
          return (
            <div
              key={stat.title}
              className="bg-white rounded-lg border border-gray-200 p-6 hover:shadow-lg transition-shadow"
            >
              <div className="flex items-center justify-between mb-4">
                <div className={`${stat.color} p-3 rounded-lg`}>
                  <Icon className="w-6 h-6 text-white" />
                </div>
                <span className="text-sm font-medium text-green-600">
                  {stat.change}
                </span>
              </div>
              <h3 className="text-2xl font-bold text-gray-900 mb-1">
                {formatNumber(stat.value)}
              </h3>
              <p className="text-sm text-gray-600">{stat.title}</p>
            </div>
          );
        })}
      </div>

      {/* Recent Activity / Charts Section */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Recent Properties */}
        <div className="bg-white rounded-lg border border-gray-200 p-6">
          <h2 className="text-lg font-bold text-gray-900 mb-4">
            Imóveis Recentes
          </h2>
          <div className="space-y-3">
            <div className="text-center py-8 text-gray-500">
              <Building2 className="w-12 h-12 mx-auto mb-2 text-gray-300" />
              <p>Nenhum imóvel recente</p>
            </div>
          </div>
        </div>

        {/* Recent Leads */}
        <div className="bg-white rounded-lg border border-gray-200 p-6">
          <h2 className="text-lg font-bold text-gray-900 mb-4">
            Leads Recentes
          </h2>
          <div className="space-y-3">
            <div className="text-center py-8 text-gray-500">
              <MessageSquare className="w-12 h-12 mx-auto mb-2 text-gray-300" />
              <p>Nenhum lead recente</p>
            </div>
          </div>
        </div>
      </div>

      {/* Quick Actions */}
      <div className="bg-white rounded-lg border border-gray-200 p-6">
        <h2 className="text-lg font-bold text-gray-900 mb-4">
          Ações Rápidas
        </h2>
        <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
          <button className="p-4 border-2 border-dashed border-gray-300 rounded-lg hover:border-blue-500 hover:bg-blue-50 transition-colors text-center">
            <Building2 className="w-8 h-8 mx-auto mb-2 text-gray-400" />
            <p className="text-sm font-medium text-gray-700">Novo Imóvel</p>
          </button>
          <button className="p-4 border-2 border-dashed border-gray-300 rounded-lg hover:border-blue-500 hover:bg-blue-50 transition-colors text-center">
            <Upload className="w-8 h-8 mx-auto mb-2 text-gray-400" />
            <p className="text-sm font-medium text-gray-700">Importar XML</p>
          </button>
          <button className="p-4 border-2 border-dashed border-gray-300 rounded-lg hover:border-blue-500 hover:bg-blue-50 transition-colors text-center">
            <Users className="w-8 h-8 mx-auto mb-2 text-gray-400" />
            <p className="text-sm font-medium text-gray-700">Novo Proprietário</p>
          </button>
          <button className="p-4 border-2 border-dashed border-gray-300 rounded-lg hover:border-blue-500 hover:bg-blue-50 transition-colors text-center">
            <MessageSquare className="w-8 h-8 mx-auto mb-2 text-gray-400" />
            <p className="text-sm font-medium text-gray-700">Ver Leads</p>
          </button>
        </div>
      </div>
    </div>
  );
}
