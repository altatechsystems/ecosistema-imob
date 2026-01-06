'use client';

import { useState, useEffect, useMemo, useCallback, useRef } from 'react';
import { useRouter } from 'next/navigation';
import { Users, Search, Filter, MessageSquare, Phone, Mail, UserCheck, ChevronDown } from 'lucide-react';
import { adminApi } from '@/lib/api';
import { Lead, LeadStatus, LeadChannel } from '@/types/lead';

type LeadStatusFilter = 'all' | LeadStatus;

export default function LeadsPage() {
  const router = useRouter();
  const [searchTerm, setSearchTerm] = useState('');
  const [leads, setLeads] = useState<Lead[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [displayCount, setDisplayCount] = useState(20);
  const [statusFilter, setStatusFilter] = useState<LeadStatusFilter>('all');
  const [updatingLeadId, setUpdatingLeadId] = useState<string | null>(null);
  const [openDropdownId, setOpenDropdownId] = useState<string | null>(null);
  const observerTarget = useRef<HTMLDivElement>(null);
  const itemsPerPage = 20;

  const fetchLeads = useCallback(async () => {
    try {
      setLoading(true);

      const response = await adminApi.getLeads({}, { limit: 1000 });

      setLeads(response.data || []);
    } catch (err: any) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  }, []);

  const handleStatusChange = async (leadId: string, newStatus: LeadStatus) => {
    try {
      setUpdatingLeadId(leadId);
      setOpenDropdownId(null);

      await adminApi.updateLeadStatus(leadId, newStatus);

      // Update local state
      setLeads(prevLeads =>
        prevLeads.map(lead =>
          lead.id === leadId ? { ...lead, status: newStatus } : lead
        )
      );
    } catch (err: any) {
      alert('Erro ao atualizar status do lead. Tente novamente.');
    } finally {
      setUpdatingLeadId(null);
    }
  };

  // Close dropdown when clicking outside
  useEffect(() => {
    const handleClickOutside = () => setOpenDropdownId(null);
    if (openDropdownId) {
      document.addEventListener('click', handleClickOutside);
      return () => document.removeEventListener('click', handleClickOutside);
    }
  }, [openDropdownId]);

  useEffect(() => {
    fetchLeads();
  }, [fetchLeads]);

  const formatDate = (date: Date | string | undefined) => {
    if (!date) return 'Data não informada';
    const d = typeof date === 'string' ? new Date(date) : date;
    return new Intl.DateTimeFormat('pt-BR', {
      day: '2-digit',
      month: '2-digit',
      year: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    }).format(d);
  };

  // Calculate stats
  const stats = useMemo(() => {
    const result = {
      total: leads.length,
      new: 0,
      contacted: 0,
      qualified: 0,
      negotiating: 0,
      converted: 0,
      lost: 0,
      whatsapp: 0,
      form: 0,
      phone: 0,
      email: 0,
    };

    leads.forEach(lead => {
      // Status counts
      if (lead.status === LeadStatus.NEW) result.new++;
      else if (lead.status === LeadStatus.CONTACTED) result.contacted++;
      else if (lead.status === LeadStatus.QUALIFIED) result.qualified++;
      else if (lead.status === LeadStatus.NEGOTIATING) result.negotiating++;
      else if (lead.status === LeadStatus.CONVERTED) result.converted++;
      else if (lead.status === LeadStatus.LOST) result.lost++;

      // Channel counts
      if (lead.channel === LeadChannel.WHATSAPP) result.whatsapp++;
      else if (lead.channel === LeadChannel.FORM) result.form++;
      else if (lead.channel === LeadChannel.PHONE) result.phone++;
      else if (lead.channel === LeadChannel.EMAIL) result.email++;
    });

    return result;
  }, [leads]);

  const filteredLeads = useMemo(() => {
    let filtered = leads;

    // Filter only WhatsApp leads
    filtered = filtered.filter(lead => lead.channel === LeadChannel.WHATSAPP);

    // Apply status filter
    if (statusFilter !== 'all') {
      filtered = filtered.filter(lead => lead.status === statusFilter);
    }

    // Apply search term filter
    if (searchTerm) {
      filtered = filtered.filter(lead =>
        lead.name?.toLowerCase().includes(searchTerm.toLowerCase()) ||
        lead.email?.toLowerCase().includes(searchTerm.toLowerCase()) ||
        lead.phone?.toLowerCase().includes(searchTerm.toLowerCase()) ||
        lead.id?.toLowerCase().includes(searchTerm.toLowerCase())
      );
    }

    return filtered;
  }, [leads, searchTerm, statusFilter]);

  // Infinite scroll - show only displayCount items
  const displayedLeads = useMemo(() =>
    filteredLeads.slice(0, displayCount),
    [filteredLeads, displayCount]
  );

  const hasMore = displayCount < filteredLeads.length;

  // Reset display count when search or filter changes
  useEffect(() => {
    setDisplayCount(20);
  }, [searchTerm, statusFilter]);

  // Infinite scroll observer
  useEffect(() => {
    const observer = new IntersectionObserver(
      (entries) => {
        if (entries[0].isIntersecting && hasMore && !loading) {
          setDisplayCount(prev => prev + itemsPerPage);
        }
      },
      { threshold: 0.1 }
    );

    const currentTarget = observerTarget.current;
    if (currentTarget) {
      observer.observe(currentTarget);
    }

    return () => {
      if (currentTarget) {
        observer.unobserve(currentTarget);
      }
    };
  }, [hasMore, loading, itemsPerPage]);

  const getStatusColor = (status?: LeadStatus) => {
    switch (status) {
      case LeadStatus.NEW: return 'bg-blue-100 text-blue-700';
      case LeadStatus.CONTACTED: return 'bg-yellow-100 text-yellow-700';
      case LeadStatus.QUALIFIED: return 'bg-purple-100 text-purple-700';
      case LeadStatus.NEGOTIATING: return 'bg-orange-100 text-orange-700';
      case LeadStatus.CONVERTED: return 'bg-green-100 text-green-700';
      case LeadStatus.LOST: return 'bg-red-100 text-red-700';
      default: return 'bg-gray-100 text-gray-700';
    }
  };

  const getStatusLabel = (status?: LeadStatus) => {
    switch (status) {
      case LeadStatus.NEW: return 'Novo';
      case LeadStatus.CONTACTED: return 'Contatado';
      case LeadStatus.QUALIFIED: return 'Qualificado';
      case LeadStatus.NEGOTIATING: return 'Negociando';
      case LeadStatus.CONVERTED: return 'Convertido';
      case LeadStatus.LOST: return 'Perdido';
      default: return 'Novo';
    }
  };

  const getChannelIcon = (channel: LeadChannel) => {
    switch (channel) {
      case LeadChannel.WHATSAPP: return <MessageSquare className="w-4 h-4" />;
      case LeadChannel.FORM: return <Mail className="w-4 h-4" />;
      case LeadChannel.PHONE: return <Phone className="w-4 h-4" />;
      case LeadChannel.EMAIL: return <Mail className="w-4 h-4" />;
      default: return <UserCheck className="w-4 h-4" />;
    }
  };

  const getChannelLabel = (channel: LeadChannel) => {
    switch (channel) {
      case LeadChannel.WHATSAPP: return 'WhatsApp';
      case LeadChannel.FORM: return 'Formulário';
      case LeadChannel.PHONE: return 'Telefone';
      case LeadChannel.EMAIL: return 'Email';
      case LeadChannel.CHAT: return 'Chat';
      case LeadChannel.REFERRAL: return 'Indicação';
      default: return channel;
    }
  };

  return (
    <div className="p-3 sm:p-4 md:p-6">
      {/* Header */}
      <div className="mb-6">
        <h1 className="text-2xl sm:text-3xl font-bold text-gray-900 mb-2">Leads WhatsApp</h1>
        <p className="text-gray-600">Acompanhe o funil de conversão dos seus leads</p>
      </div>

      {/* Funnel Visualization - Modern Sales Pipeline */}
      <div className="bg-white rounded-lg shadow-sm p-4 sm:p-6 mb-6">
        <h2 className="text-lg font-semibold text-gray-900 mb-4">Funil de Vendas</h2>

        <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-6 gap-3 sm:gap-4">
          {/* Step 1: Novo */}
          <button
            onClick={() => setStatusFilter(LeadStatus.NEW)}
            className={`relative group transition-all hover:scale-105 ${
              statusFilter === LeadStatus.NEW ? 'scale-105' : ''
            }`}
          >
            <div className={`bg-gradient-to-br from-blue-50 to-blue-100 rounded-lg p-4 border-2 transition-all ${
              statusFilter === LeadStatus.NEW ? 'border-blue-500 shadow-lg' : 'border-blue-200'
            }`}>
              <div className="flex items-center justify-between mb-2">
                <div className="w-10 h-10 bg-blue-500 rounded-lg flex items-center justify-center">
                  <MessageSquare className="w-5 h-5 text-white" />
                </div>
                <span className="text-xs font-medium text-blue-600 bg-blue-200 px-2 py-1 rounded-full">1</span>
              </div>
              <p className="text-2xl font-bold text-blue-900 mb-1">{stats.new}</p>
              <p className="text-xs font-medium text-blue-700">Novos</p>
              <div className="absolute -right-2 top-1/2 -translate-y-1/2 text-blue-300 hidden lg:block">
                →
              </div>
            </div>
          </button>

          {/* Step 2: Contatado */}
          <button
            onClick={() => setStatusFilter(LeadStatus.CONTACTED)}
            className={`relative group transition-all hover:scale-105 ${
              statusFilter === LeadStatus.CONTACTED ? 'scale-105' : ''
            }`}
          >
            <div className={`bg-gradient-to-br from-yellow-50 to-yellow-100 rounded-lg p-4 border-2 transition-all ${
              statusFilter === LeadStatus.CONTACTED ? 'border-yellow-500 shadow-lg' : 'border-yellow-200'
            }`}>
              <div className="flex items-center justify-between mb-2">
                <div className="w-10 h-10 bg-yellow-500 rounded-lg flex items-center justify-center">
                  <UserCheck className="w-5 h-5 text-white" />
                </div>
                <span className="text-xs font-medium text-yellow-600 bg-yellow-200 px-2 py-1 rounded-full">2</span>
              </div>
              <p className="text-2xl font-bold text-yellow-900 mb-1">{stats.contacted}</p>
              <p className="text-xs font-medium text-yellow-700">Contatados</p>
              <div className="absolute -right-2 top-1/2 -translate-y-1/2 text-yellow-300 hidden lg:block">
                →
              </div>
            </div>
          </button>

          {/* Step 3: Qualificado */}
          <button
            onClick={() => setStatusFilter(LeadStatus.QUALIFIED)}
            className={`relative group transition-all hover:scale-105 ${
              statusFilter === LeadStatus.QUALIFIED ? 'scale-105' : ''
            }`}
          >
            <div className={`bg-gradient-to-br from-purple-50 to-purple-100 rounded-lg p-4 border-2 transition-all ${
              statusFilter === LeadStatus.QUALIFIED ? 'border-purple-500 shadow-lg' : 'border-purple-200'
            }`}>
              <div className="flex items-center justify-between mb-2">
                <div className="w-10 h-10 bg-purple-500 rounded-lg flex items-center justify-center">
                  <UserCheck className="w-5 h-5 text-white" />
                </div>
                <span className="text-xs font-medium text-purple-600 bg-purple-200 px-2 py-1 rounded-full">3</span>
              </div>
              <p className="text-2xl font-bold text-purple-900 mb-1">{stats.qualified}</p>
              <p className="text-xs font-medium text-purple-700">Qualificados</p>
              <div className="absolute -right-2 top-1/2 -translate-y-1/2 text-purple-300 hidden lg:block">
                →
              </div>
            </div>
          </button>

          {/* Step 4: Negociando */}
          <button
            onClick={() => setStatusFilter(LeadStatus.NEGOTIATING)}
            className={`relative group transition-all hover:scale-105 ${
              statusFilter === LeadStatus.NEGOTIATING ? 'scale-105' : ''
            }`}
          >
            <div className={`bg-gradient-to-br from-orange-50 to-orange-100 rounded-lg p-4 border-2 transition-all ${
              statusFilter === LeadStatus.NEGOTIATING ? 'border-orange-500 shadow-lg' : 'border-orange-200'
            }`}>
              <div className="flex items-center justify-between mb-2">
                <div className="w-10 h-10 bg-orange-500 rounded-lg flex items-center justify-center">
                  <MessageSquare className="w-5 h-5 text-white" />
                </div>
                <span className="text-xs font-medium text-orange-600 bg-orange-200 px-2 py-1 rounded-full">4</span>
              </div>
              <p className="text-2xl font-bold text-orange-900 mb-1">{stats.negotiating}</p>
              <p className="text-xs font-medium text-orange-700">Negociando</p>
              <div className="absolute -right-2 top-1/2 -translate-y-1/2 text-orange-300 hidden lg:block">
                →
              </div>
            </div>
          </button>

          {/* Step 5: Convertido */}
          <button
            onClick={() => setStatusFilter(LeadStatus.CONVERTED)}
            className={`relative group transition-all hover:scale-105 ${
              statusFilter === LeadStatus.CONVERTED ? 'scale-105' : ''
            }`}
          >
            <div className={`bg-gradient-to-br from-green-50 to-green-100 rounded-lg p-4 border-2 transition-all ${
              statusFilter === LeadStatus.CONVERTED ? 'border-green-500 shadow-lg' : 'border-green-200'
            }`}>
              <div className="flex items-center justify-between mb-2">
                <div className="w-10 h-10 bg-green-500 rounded-lg flex items-center justify-center">
                  <UserCheck className="w-5 h-5 text-white" />
                </div>
                <span className="text-xs font-medium text-green-600 bg-green-200 px-2 py-1 rounded-full">✓</span>
              </div>
              <p className="text-2xl font-bold text-green-900 mb-1">{stats.converted}</p>
              <p className="text-xs font-medium text-green-700">Convertidos</p>
            </div>
          </button>

          {/* Step 6: Perdidos */}
          <button
            onClick={() => setStatusFilter(LeadStatus.LOST)}
            className={`relative group transition-all hover:scale-105 ${
              statusFilter === LeadStatus.LOST ? 'scale-105' : ''
            }`}
          >
            <div className={`bg-gradient-to-br from-red-50 to-red-100 rounded-lg p-4 border-2 transition-all ${
              statusFilter === LeadStatus.LOST ? 'border-red-500 shadow-lg' : 'border-red-200'
            }`}>
              <div className="flex items-center justify-between mb-2">
                <div className="w-10 h-10 bg-red-500 rounded-lg flex items-center justify-center">
                  <Users className="w-5 h-5 text-white" />
                </div>
                <span className="text-xs font-medium text-red-600 bg-red-200 px-2 py-1 rounded-full">✕</span>
              </div>
              <p className="text-2xl font-bold text-red-900 mb-1">{stats.lost}</p>
              <p className="text-xs font-medium text-red-700">Perdidos</p>
            </div>
          </button>
        </div>

        {/* Funnel Metrics */}
        <div className="mt-6 grid grid-cols-1 sm:grid-cols-3 gap-4">
          <div className="bg-gray-50 rounded-lg p-4">
            <p className="text-sm text-gray-600 mb-1">Total de Leads</p>
            <p className="text-2xl font-bold text-gray-900">{stats.whatsapp}</p>
            <p className="text-xs text-gray-500 mt-1">via WhatsApp</p>
          </div>
          <div className="bg-gray-50 rounded-lg p-4">
            <p className="text-sm text-gray-600 mb-1">Taxa de Conversão</p>
            <p className="text-2xl font-bold text-green-600">
              {stats.whatsapp > 0 ? ((stats.converted / stats.whatsapp) * 100).toFixed(1) : 0}%
            </p>
            <p className="text-xs text-gray-500 mt-1">{stats.converted} convertidos de {stats.whatsapp}</p>
          </div>
          <div className="bg-gray-50 rounded-lg p-4">
            <p className="text-sm text-gray-600 mb-1">Em Atendimento</p>
            <p className="text-2xl font-bold text-orange-600">
              {stats.contacted + stats.qualified + stats.negotiating}
            </p>
            <p className="text-xs text-gray-500 mt-1">aguardando fechamento</p>
          </div>
        </div>
      </div>

      {/* Actions Bar */}
      <div className="bg-white rounded-lg shadow-sm p-3 sm:p-4 mb-4 sm:mb-6">
        <div className="flex flex-col sm:flex-row gap-2 sm:gap-4">
          {/* Search */}
          <div className="flex-1 relative">
            <Search className="absolute left-2.5 sm:left-3 top-1/2 transform -translate-y-1/2 text-gray-400 w-4 h-4 sm:w-5 sm:h-5" />
            <input
              type="text"
              placeholder="Buscar por nome, email, telefone ou ID..."
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              className="w-full pl-8 sm:pl-10 pr-3 sm:pr-4 py-2 text-sm sm:text-base border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
          </div>

          {/* Filter Button */}
          <button className="flex items-center justify-center gap-2 px-3 sm:px-4 py-2 border border-gray-300 rounded-lg hover:bg-gray-50 transition-colors text-sm sm:text-base">
            <Filter className="w-4 h-4 sm:w-5 sm:h-5" />
            <span className="hidden sm:inline">Filtros</span>
          </button>
        </div>
      </div>

      {/* Loading State */}
      {loading && (
        <div className="bg-white rounded-lg shadow-sm p-8">
          <div className="animate-pulse space-y-4">
            {[...Array(5)].map((_, i) => (
              <div key={i} className="flex items-center gap-4">
                <div className="w-12 h-12 bg-gray-200 rounded-full"></div>
                <div className="flex-1">
                  <div className="h-4 bg-gray-200 rounded w-1/4 mb-2"></div>
                  <div className="h-3 bg-gray-200 rounded w-1/2"></div>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Error State */}
      {error && !loading && (
        <div className="bg-red-50 border border-red-200 rounded-lg p-4 mb-6">
          <p className="text-red-600">{error}</p>
        </div>
      )}

      {/* Empty State */}
      {!loading && !error && filteredLeads.length === 0 && (
        <div className="bg-white rounded-lg shadow-sm p-12">
          <div className="text-center">
            <div className="inline-flex items-center justify-center w-16 h-16 bg-blue-100 rounded-full mb-4">
              <Users className="w-8 h-8 text-blue-600" />
            </div>
            <h3 className="text-lg font-semibold text-gray-900 mb-2">
              {searchTerm ? 'Nenhum lead encontrado' : 'Nenhum lead cadastrado'}
            </h3>
            <p className="text-gray-600">
              {searchTerm
                ? 'Tente buscar com outros termos'
                : 'Os leads começarão a aparecer quando visitantes entrarem em contato através do site'}
            </p>
          </div>
        </div>
      )}

      {/* Leads Table */}
      {!loading && !error && filteredLeads.length > 0 && (
        <>
          <div className="bg-white rounded-lg shadow-sm mb-6" style={{ overflow: 'visible' }}>
            <div className="overflow-x-auto" style={{ overflow: 'visible' }}>
              <table className="w-full">
                <thead className="bg-gray-50 border-b border-gray-200">
                  <tr>
                    <th className="px-4 py-3 text-left text-xs font-medium text-gray-600 uppercase tracking-wider">
                      Lead
                    </th>
                    <th className="px-4 py-3 text-left text-xs font-medium text-gray-600 uppercase tracking-wider">
                      Contato
                    </th>
                    <th className="px-4 py-3 text-left text-xs font-medium text-gray-600 uppercase tracking-wider">
                      Canal
                    </th>
                    <th className="px-4 py-3 text-left text-xs font-medium text-gray-600 uppercase tracking-wider">
                      Status
                    </th>
                    <th className="px-4 py-3 text-left text-xs font-medium text-gray-600 uppercase tracking-wider">
                      Data
                    </th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-gray-200">
                  {displayedLeads.map((lead) => (
                    <tr
                      key={lead.id}
                      className="hover:bg-gray-50 transition-colors"
                    >
                      <td
                        className="px-4 py-4 cursor-pointer"
                        onClick={() => router.push(`/dashboard/leads/${lead.id}`)}
                      >
                        <div>
                          <p className="font-medium text-gray-900">{lead.name || 'Nome não informado'}</p>
                          <p className="text-sm text-gray-500">ID: {lead.id?.slice(0, 8)}...</p>
                        </div>
                      </td>
                      <td
                        className="px-4 py-4 cursor-pointer"
                        onClick={() => router.push(`/dashboard/leads/${lead.id}`)}
                      >
                        <div className="text-sm">
                          {lead.email && (
                            <p className="text-gray-900 flex items-center gap-1">
                              <Mail className="w-3 h-3 text-gray-400" />
                              {lead.email}
                            </p>
                          )}
                          {lead.phone && (
                            <p className="text-gray-600 flex items-center gap-1">
                              <Phone className="w-3 h-3 text-gray-400" />
                              {lead.phone}
                            </p>
                          )}
                        </div>
                      </td>
                      <td
                        className="px-4 py-4 cursor-pointer"
                        onClick={() => router.push(`/dashboard/leads/${lead.id}`)}
                      >
                        <div className="flex items-center gap-2">
                          {getChannelIcon(lead.channel)}
                          <span className="text-sm text-gray-700">{getChannelLabel(lead.channel)}</span>
                        </div>
                      </td>
                      <td className="px-4 py-4">
                        <div className="relative inline-block" onClick={(e) => e.stopPropagation()}>
                          <button
                            onClick={() => setOpenDropdownId(openDropdownId === lead.id ? null : (lead.id || null))}
                            disabled={updatingLeadId === lead.id}
                            className={`flex items-center gap-2 px-3 py-1 rounded-full text-xs font-medium transition-all ${getStatusColor(lead.status)} ${
                              updatingLeadId === lead.id ? 'opacity-50 cursor-not-allowed' : 'hover:shadow-md cursor-pointer'
                            }`}
                          >
                            {updatingLeadId === lead.id ? (
                              <>
                                <div className="w-3 h-3 border-2 border-current border-t-transparent rounded-full animate-spin"></div>
                                <span>Atualizando...</span>
                              </>
                            ) : (
                              <>
                                <span>{getStatusLabel(lead.status)}</span>
                                <ChevronDown className="w-3 h-3" />
                              </>
                            )}
                          </button>

                          {/* Dropdown Menu - Absolute positioning */}
                          {openDropdownId === lead.id && (
                            <div className="absolute top-full left-0 mt-1 bg-white rounded-lg shadow-xl border border-gray-200 py-1 z-[9999] min-w-[200px]">
                              {Object.values(LeadStatus).map((status) => (
                                <button
                                  key={status}
                                  onClick={() => handleStatusChange(lead.id!, status)}
                                  className={`w-full px-4 py-2.5 text-left text-sm hover:bg-gray-50 transition-colors flex items-center gap-3 ${
                                    lead.status === status ? 'bg-gray-50 font-medium' : ''
                                  }`}
                                >
                                  <span className={`w-2.5 h-2.5 rounded-full flex-shrink-0 ${
                                    status === LeadStatus.NEW ? 'bg-blue-500' :
                                    status === LeadStatus.CONTACTED ? 'bg-yellow-500' :
                                    status === LeadStatus.QUALIFIED ? 'bg-purple-500' :
                                    status === LeadStatus.NEGOTIATING ? 'bg-orange-500' :
                                    status === LeadStatus.CONVERTED ? 'bg-green-500' :
                                    'bg-red-500'
                                  }`}></span>
                                  <span className="flex-1">{getStatusLabel(status)}</span>
                                  {lead.status === status && (
                                    <span className="text-blue-600 font-bold">✓</span>
                                  )}
                                </button>
                              ))}
                            </div>
                          )}
                        </div>
                      </td>
                      <td
                        className="px-4 py-4 cursor-pointer"
                        onClick={() => router.push(`/dashboard/leads/${lead.id}`)}
                      >
                        <p className="text-sm text-gray-600">{formatDate(lead.created_at)}</p>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>

          {/* Infinite Scroll Trigger & Loading Indicator */}
          <div ref={observerTarget} className="py-8">
            {hasMore && (
              <div className="text-center">
                <div className="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
                <p className="text-gray-600 mt-2">Carregando mais leads...</p>
              </div>
            )}
            {!hasMore && filteredLeads.length > 20 && (
              <div className="text-center">
                <p className="text-gray-600">
                  Mostrando todos os {filteredLeads.length} leads
                </p>
              </div>
            )}
          </div>
        </>
      )}
    </div>
  );
}
