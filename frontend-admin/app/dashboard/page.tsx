'use client';

import { useEffect, useState } from 'react';
import { adminApi } from '@/lib/api';
import { formatCurrency, formatNumber } from '@/lib/utils';
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
  const [metrics, setMetrics] = useState<DashboardMetrics | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    loadMetrics();
  }, []);

  const loadMetrics = async () => {
    try {
      setLoading(true);

      // Buscar dados diretamente das APIs existentes
      const tenantId = localStorage.getItem('tenant_id');

      if (!tenantId) {
        console.error('Tenant ID não encontrado');
        setLoading(false);
        return;
      }

      // Buscar total de imóveis (usando o novo campo total da API)
      const propertiesResponse = await fetch(`${process.env.NEXT_PUBLIC_API_URL}/${tenantId}/properties?limit=1`);
      if (!propertiesResponse.ok) {
        throw new Error(`Failed to fetch properties: ${propertiesResponse.status}`);
      }
      const propertiesData = await propertiesResponse.json();
      const totalProperties = propertiesData.total || 0;

      // Delay to avoid rate limiting
      await new Promise(resolve => setTimeout(resolve, 300));

      // Buscar imóveis disponíveis
      const availableResponse = await fetch(`${process.env.NEXT_PUBLIC_API_URL}/${tenantId}/properties?status=available&limit=1`);
      if (!availableResponse.ok) {
        throw new Error(`Failed to fetch available properties: ${availableResponse.status}`);
      }
      const availableData = await availableResponse.json();
      const availableProperties = availableData.total || 0;

      // Delay to avoid rate limiting
      await new Promise(resolve => setTimeout(resolve, 300));

      // Buscar imóveis vendidos
      const soldResponse = await fetch(`${process.env.NEXT_PUBLIC_API_URL}/${tenantId}/properties?status=sold&limit=1`);
      if (!soldResponse.ok) {
        throw new Error(`Failed to fetch sold properties: ${soldResponse.status}`);
      }
      const soldData = await soldResponse.json();
      const soldProperties = soldData.total || 0;

      // Delay to avoid rate limiting
      await new Promise(resolve => setTimeout(resolve, 300));

      // Buscar imóveis alugados
      const rentedResponse = await fetch(`${process.env.NEXT_PUBLIC_API_URL}/${tenantId}/properties?status=rented&limit=1`);
      if (!rentedResponse.ok) {
        throw new Error(`Failed to fetch rented properties: ${rentedResponse.status}`);
      }
      const rentedData = await rentedResponse.json();
      const rentedProperties = rentedData.total || 0;

      setMetrics({
        total_properties: totalProperties,
        available_properties: availableProperties,
        sold_properties: soldProperties,
        rented_properties: rentedProperties,
        total_leads: 0, // TODO: Implementar quando tiver endpoint de leads
        new_leads: 0,
        converted_leads: 0,
        total_owners: 0,
        total_brokers: 0,
      });
    } catch (error) {
      console.error('Failed to load metrics:', error);
    } finally {
      setLoading(false);
    }
  };

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
