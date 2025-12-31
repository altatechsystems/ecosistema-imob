'use client';

import { useState, useEffect } from 'react';
import { Calendar, Send, CheckCircle, XCircle, Clock, AlertCircle, RefreshCw, TrendingUp, Users, MessageCircle, Timer } from 'lucide-react';
import { adminApi } from '@/lib/api';
import { MetricsCard } from '@/components/dashboard/metrics-card';

interface ScheduledConfirmation {
  id: string;
  property_id: string;
  owner_id: string;
  broker_id: string;
  token_id: string;
  confirmation_url: string;
  scheduled_for: string;
  sent_at?: string;
  status: 'pending' | 'sent' | 'failed' | 'cancelled' | 'responded';
  delivery_method: string;
  delivery_status?: string;
  responded_at?: string;
  response?: string;
  created_at: string;
}

interface ScheduleResponse {
  total_properties: number;
  scheduled_count: number;
  skipped_count: number;
  scheduled_for_date: string;
  skipped_reasons?: string[];
  scheduled_confirm_ids?: string[];
}

interface ConfirmationMetrics {
  total_confirmations: number;
  pending_count: number;
  sent_count: number;
  responded_count: number;
  failed_count: number;
  cancelled_count: number;
  response_rate: number;
  success_rate: number;
  inactive_owner_count: number;
  average_response_time: string;
}

export default function ConfirmacoesAgendadasPage() {
  const [confirmations, setConfirmations] = useState<ScheduledConfirmation[]>([]);
  const [metrics, setMetrics] = useState<ConfirmationMetrics | null>(null);
  const [loading, setLoading] = useState(true);
  const [scheduling, setScheduling] = useState(false);
  const [processing, setProcessing] = useState(false);
  const [filterStatus, setFilterStatus] = useState<string>('all');
  const [scheduleResult, setScheduleResult] = useState<ScheduleResponse | null>(null);

  useEffect(() => {
    loadConfirmations();
    loadMetrics();
  }, [filterStatus]);

  const loadConfirmations = async () => {
    try {
      setLoading(true);
      const status = filterStatus === 'all' ? undefined : filterStatus;
      const response = await adminApi.getScheduledConfirmations(status);
      setConfirmations(response);
    } catch (err: any) {
      console.error('Erro ao carregar confirmações:', err);
      alert('Erro ao carregar confirmações agendadas');
    } finally {
      setLoading(false);
    }
  };

  const loadMetrics = async () => {
    try {
      const response = await adminApi.getConfirmationMetrics();
      setMetrics(response);
    } catch (err: any) {
      console.error('Erro ao carregar métricas:', err);
    }
  };

  const handleScheduleMonthly = async (dryRun: boolean = false) => {
    if (!dryRun && !confirm('Deseja agendar confirmações mensais para todos os imóveis ativos?')) {
      return;
    }

    try {
      setScheduling(true);
      const response = await adminApi.scheduleMonthlyConfirmations({ dry_run: dryRun });
      setScheduleResult(response);

      if (!dryRun) {
        alert(`Confirmações agendadas com sucesso!\n\n${response.scheduled_count} imóveis agendados\n${response.skipped_count} imóveis pulados`);
        loadConfirmations();
      }
    } catch (err: any) {
      console.error('Erro ao agendar confirmações:', err);
      alert('Erro ao agendar confirmações. Tente novamente.');
    } finally {
      setScheduling(false);
    }
  };

  const handleProcessPending = async () => {
    if (!confirm('Deseja processar todas as confirmações pendentes de hoje?')) {
      return;
    }

    try {
      setProcessing(true);
      await adminApi.processPendingConfirmations();
      alert('Confirmações pendentes processadas com sucesso!');
      loadConfirmations();
    } catch (err: any) {
      console.error('Erro ao processar confirmações:', err);
      alert('Erro ao processar confirmações. Tente novamente.');
    } finally {
      setProcessing(false);
    }
  };

  const getStatusBadge = (status: string) => {
    const styles = {
      pending: 'bg-yellow-100 text-yellow-800',
      sent: 'bg-blue-100 text-blue-800',
      failed: 'bg-red-100 text-red-800',
      cancelled: 'bg-gray-100 text-gray-800',
      responded: 'bg-green-100 text-green-800',
    };

    const labels = {
      pending: 'Pendente',
      sent: 'Enviado',
      failed: 'Falhou',
      cancelled: 'Cancelado',
      responded: 'Respondido',
    };

    return (
      <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${styles[status as keyof typeof styles] || 'bg-gray-100 text-gray-800'}`}>
        {labels[status as keyof typeof labels] || status}
      </span>
    );
  };

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'pending':
        return <Clock className="w-4 h-4 text-yellow-600" />;
      case 'sent':
        return <Send className="w-4 h-4 text-blue-600" />;
      case 'failed':
        return <XCircle className="w-4 h-4 text-red-600" />;
      case 'responded':
        return <CheckCircle className="w-4 h-4 text-green-600" />;
      default:
        return <AlertCircle className="w-4 h-4 text-gray-600" />;
    }
  };

  return (
    <div className="p-6 max-w-7xl mx-auto">
      {/* Header */}
      <div className="mb-6">
        <h1 className="text-2xl font-bold text-gray-900 mb-2">Confirmações Agendadas</h1>
        <p className="text-gray-600">Gerencie as confirmações mensais enviadas aos proprietários</p>
      </div>

      {/* Metrics */}
      {metrics && (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4 mb-6">
          <MetricsCard
            title="Taxa de Resposta"
            value={`${metrics.response_rate.toFixed(1)}%`}
            subtitle={`${metrics.responded_count} de ${metrics.responded_count + metrics.sent_count} enviados`}
            icon={TrendingUp}
            iconColor="text-green-600"
          />
          <MetricsCard
            title="Tempo Médio de Resposta"
            value={metrics.average_response_time}
            subtitle="Desde o envio até a resposta"
            icon={Timer}
            iconColor="text-blue-600"
          />
          <MetricsCard
            title="Proprietários Inativos"
            value={metrics.inactive_owner_count}
            subtitle="Sem resposta há 3+ meses"
            icon={Users}
            iconColor="text-orange-600"
          />
          <MetricsCard
            title="Taxa de Sucesso"
            value={`${metrics.success_rate.toFixed(1)}%`}
            subtitle={`${metrics.failed_count} falharam`}
            icon={MessageCircle}
            iconColor="text-purple-600"
          />
        </div>
      )}

      {/* Actions */}
      <div className="bg-white rounded-lg shadow-sm p-4 mb-6">
        <div className="flex flex-wrap gap-3">
          <button
            onClick={() => handleScheduleMonthly(true)}
            disabled={scheduling}
            className="flex items-center gap-2 px-4 py-2 bg-indigo-50 text-indigo-700 rounded-lg hover:bg-indigo-100 disabled:bg-gray-100 disabled:text-gray-400 disabled:cursor-not-allowed font-medium"
          >
            <Calendar className="w-4 h-4" />
            {scheduling ? 'Simulando...' : 'Simular Agendamento'}
          </button>

          <button
            onClick={() => handleScheduleMonthly(false)}
            disabled={scheduling}
            className="flex items-center gap-2 px-4 py-2 bg-green-600 text-white rounded-lg hover:bg-green-700 disabled:bg-gray-300 disabled:cursor-not-allowed font-medium"
          >
            <Calendar className="w-4 h-4" />
            {scheduling ? 'Agendando...' : 'Agendar Confirmações Mensais'}
          </button>

          <button
            onClick={handleProcessPending}
            disabled={processing}
            className="flex items-center gap-2 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:bg-gray-300 disabled:cursor-not-allowed font-medium"
          >
            <Send className="w-4 h-4" />
            {processing ? 'Processando...' : 'Processar Pendentes de Hoje'}
          </button>

          <button
            onClick={loadConfirmations}
            disabled={loading}
            className="flex items-center gap-2 px-4 py-2 bg-gray-100 text-gray-700 rounded-lg hover:bg-gray-200 disabled:bg-gray-50 disabled:cursor-not-allowed font-medium"
          >
            <RefreshCw className={`w-4 h-4 ${loading ? 'animate-spin' : ''}`} />
            Atualizar
          </button>
        </div>
      </div>

      {/* Schedule Result */}
      {scheduleResult && (
        <div className="bg-blue-50 border border-blue-200 rounded-lg p-4 mb-6">
          <h3 className="font-semibold text-blue-900 mb-2">Resultado do Agendamento</h3>
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mb-3">
            <div>
              <p className="text-sm text-blue-700">Total de Imóveis</p>
              <p className="text-2xl font-bold text-blue-900">{scheduleResult.total_properties}</p>
            </div>
            <div>
              <p className="text-sm text-green-700">Agendados</p>
              <p className="text-2xl font-bold text-green-900">{scheduleResult.scheduled_count}</p>
            </div>
            <div>
              <p className="text-sm text-yellow-700">Pulados</p>
              <p className="text-2xl font-bold text-yellow-900">{scheduleResult.skipped_count}</p>
            </div>
          </div>
          <p className="text-sm text-blue-700">
            Agendado para: <span className="font-semibold">{scheduleResult.scheduled_for_date}</span>
          </p>
          {scheduleResult.skipped_reasons && scheduleResult.skipped_reasons.length > 0 && (
            <details className="mt-3">
              <summary className="cursor-pointer text-sm font-medium text-blue-900">
                Ver razões ({scheduleResult.skipped_reasons.length})
              </summary>
              <ul className="mt-2 text-sm text-blue-800 list-disc list-inside">
                {scheduleResult.skipped_reasons.slice(0, 10).map((reason, idx) => (
                  <li key={idx}>{reason}</li>
                ))}
                {scheduleResult.skipped_reasons.length > 10 && (
                  <li className="text-blue-600">... e mais {scheduleResult.skipped_reasons.length - 10}</li>
                )}
              </ul>
            </details>
          )}
        </div>
      )}

      {/* Filters */}
      <div className="bg-white rounded-lg shadow-sm p-4 mb-6">
        <div className="flex items-center gap-2">
          <label className="text-sm font-medium text-gray-700">Filtrar por status:</label>
          <select
            value={filterStatus}
            onChange={(e) => setFilterStatus(e.target.value)}
            className="px-3 py-1.5 border border-gray-300 rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
          >
            <option value="all">Todos</option>
            <option value="pending">Pendentes</option>
            <option value="sent">Enviados</option>
            <option value="responded">Respondidos</option>
            <option value="failed">Falhados</option>
          </select>
          <span className="text-sm text-gray-500 ml-auto">
            {confirmations.length} {confirmations.length === 1 ? 'confirmação' : 'confirmações'}
          </span>
        </div>
      </div>

      {/* Confirmations List */}
      <div className="bg-white rounded-lg shadow-sm overflow-hidden">
        {loading ? (
          <div className="p-8 text-center">
            <RefreshCw className="w-8 h-8 text-gray-400 animate-spin mx-auto mb-2" />
            <p className="text-gray-600">Carregando confirmações...</p>
          </div>
        ) : confirmations.length === 0 ? (
          <div className="p-8 text-center">
            <Calendar className="w-12 h-12 text-gray-400 mx-auto mb-3" />
            <h3 className="text-lg font-medium text-gray-900 mb-1">Nenhuma confirmação encontrada</h3>
            <p className="text-gray-600">
              {filterStatus === 'all'
                ? 'Clique em "Agendar Confirmações Mensais" para criar novos agendamentos'
                : 'Tente outro filtro ou agende novas confirmações'}
            </p>
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="min-w-full divide-y divide-gray-200">
              <thead className="bg-gray-50">
                <tr>
                  <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Status
                  </th>
                  <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Imóvel
                  </th>
                  <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Agendado para
                  </th>
                  <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Enviado em
                  </th>
                  <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Método
                  </th>
                  <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Resposta
                  </th>
                  <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Link
                  </th>
                </tr>
              </thead>
              <tbody className="bg-white divide-y divide-gray-200">
                {confirmations.map((confirmation) => (
                  <tr key={confirmation.id} className="hover:bg-gray-50">
                    <td className="px-4 py-3 whitespace-nowrap">
                      <div className="flex items-center gap-2">
                        {getStatusIcon(confirmation.status)}
                        {getStatusBadge(confirmation.status)}
                      </div>
                    </td>
                    <td className="px-4 py-3 whitespace-nowrap text-sm text-gray-900">
                      {confirmation.property_id}
                    </td>
                    <td className="px-4 py-3 whitespace-nowrap text-sm text-gray-600">
                      {new Date(confirmation.scheduled_for).toLocaleDateString('pt-BR', {
                        day: '2-digit',
                        month: '2-digit',
                        year: 'numeric',
                        hour: '2-digit',
                        minute: '2-digit',
                      })}
                    </td>
                    <td className="px-4 py-3 whitespace-nowrap text-sm text-gray-600">
                      {confirmation.sent_at
                        ? new Date(confirmation.sent_at).toLocaleDateString('pt-BR', {
                            day: '2-digit',
                            month: '2-digit',
                            year: 'numeric',
                          })
                        : '-'}
                    </td>
                    <td className="px-4 py-3 whitespace-nowrap text-sm text-gray-600">
                      {confirmation.delivery_method === 'manual' ? 'Manual' : confirmation.delivery_method}
                    </td>
                    <td className="px-4 py-3 whitespace-nowrap text-sm">
                      {confirmation.response ? (
                        <span className="text-green-700 font-medium">
                          {confirmation.response === 'available' ? 'Disponível' :
                           confirmation.response === 'unavailable' ? 'Indisponível' :
                           confirmation.response === 'price_updated' ? 'Preço Atualizado' :
                           confirmation.response}
                        </span>
                      ) : (
                        <span className="text-gray-400">Sem resposta</span>
                      )}
                    </td>
                    <td className="px-4 py-3 whitespace-nowrap text-sm">
                      {confirmation.confirmation_url && (
                        <a
                          href={confirmation.confirmation_url}
                          target="_blank"
                          rel="noopener noreferrer"
                          className="text-indigo-600 hover:text-indigo-800 hover:underline"
                        >
                          Abrir link
                        </a>
                      )}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </div>
  );
}
