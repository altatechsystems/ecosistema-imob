'use client';

import { useState, useEffect } from 'react';
import { useParams, useRouter } from 'next/navigation';
import {
  ArrowLeft,
  User,
  Mail,
  Phone,
  MessageSquare,
  Calendar,
  MapPin,
  ExternalLink,
  Shield,
  Globe,
  TrendingUp,
  Building2,
  Edit,
} from 'lucide-react';
import { adminApi } from '@/lib/api';
import { Lead, LeadStatus, LeadChannel } from '@/types/lead';
import { Property } from '@/types/property';

export default function LeadDetailPage() {
  const params = useParams();
  const router = useRouter();
  const leadId = params?.id as string;

  const [lead, setLead] = useState<Lead | null>(null);
  const [property, setProperty] = useState<Property | null>(null);
  const [loading, setLoading] = useState(true);
  const [loadingProperty, setLoadingProperty] = useState(false);
  const [error, setError] = useState('');
  const [updatingStatus, setUpdatingStatus] = useState(false);

  useEffect(() => {
    if (leadId) {
      fetchLeadDetails();
    }
  }, [leadId]);

  const fetchLeadDetails = async () => {
    try {
      setLoading(true);

      const leadData = await adminApi.getLead(leadId);
      setLead(leadData);

      // Fetch property details if property_id exists
      if (leadData.property_id) {
        fetchPropertyDetails(leadData.property_id);
      }
    } catch (err: any) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  const fetchPropertyDetails = async (propertyId: string) => {
    try {
      setLoadingProperty(true);
      const propertyData = await adminApi.getProperty(propertyId);
      setProperty(propertyData);
    } catch (err: any) {
      // Error fetching property - silently fail
    } finally {
      setLoadingProperty(false);
    }
  };

  const handleStatusUpdate = async (newStatus: LeadStatus) => {
    if (!lead || !lead.id) return;

    try {
      setUpdatingStatus(true);
      const updatedLead = await adminApi.updateLeadStatus(lead.id, newStatus);
      setLead(updatedLead);
    } catch (err: any) {
      alert('Erro ao atualizar status do lead');
    } finally {
      setUpdatingStatus(false);
    }
  };

  const formatDate = (date: Date | string | undefined) => {
    if (!date) return 'Data não informada';
    const d = typeof date === 'string' ? new Date(date) : date;
    return new Intl.DateTimeFormat('pt-BR', {
      day: '2-digit',
      month: 'long',
      year: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    }).format(d);
  };

  const formatPrice = (price?: number) => {
    if (!price) return 'Preço não informado';
    return new Intl.NumberFormat('pt-BR', {
      style: 'currency',
      currency: 'BRL',
    }).format(price);
  };

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

  const openWhatsApp = () => {
    if (!lead?.phone) return;
    const phone = lead.phone.replace(/\D/g, '');
    const message = lead.name
      ? `Olá ${lead.name}, tudo bem?`
      : 'Olá, tudo bem?';
    window.open(`https://wa.me/55${phone}?text=${encodeURIComponent(message)}`, '_blank');
  };

  if (loading) {
    return (
      <div className="p-6">
        <div className="animate-pulse">
          <div className="h-8 bg-gray-200 rounded w-1/4 mb-6"></div>
          <div className="bg-white rounded-lg shadow-sm p-6">
            <div className="space-y-4">
              <div className="h-6 bg-gray-200 rounded w-3/4"></div>
              <div className="h-4 bg-gray-200 rounded w-1/2"></div>
              <div className="h-4 bg-gray-200 rounded w-2/3"></div>
            </div>
          </div>
        </div>
      </div>
    );
  }

  if (error || !lead) {
    return (
      <div className="p-6">
        <button
          onClick={() => router.push('/dashboard/leads')}
          className="flex items-center gap-2 text-gray-600 hover:text-gray-900 mb-6"
        >
          <ArrowLeft className="w-5 h-5" />
          Voltar para Leads
        </button>
        <div className="bg-red-50 border border-red-200 rounded-lg p-6">
          <p className="text-red-600">{error || 'Lead não encontrado'}</p>
        </div>
      </div>
    );
  }

  return (
    <div className="p-3 sm:p-4 md:p-6">
      {/* Header */}
      <div className="mb-6">
        <button
          onClick={() => router.push('/dashboard/leads')}
          className="flex items-center gap-2 text-gray-600 hover:text-gray-900 mb-4 transition-colors"
        >
          <ArrowLeft className="w-5 h-5" />
          Voltar para Leads
        </button>

        <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
          <div>
            <h1 className="text-2xl sm:text-3xl font-bold text-gray-900 mb-2">
              {lead.name || 'Lead sem nome'}
            </h1>
            <p className="text-gray-600">ID: {lead.id}</p>
          </div>
          <div className="flex items-center gap-3">
            <span className={`px-4 py-2 rounded-full text-sm font-medium ${getStatusColor(lead.status)}`}>
              {getStatusLabel(lead.status)}
            </span>
          </div>
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Main Content */}
        <div className="lg:col-span-2 space-y-6">
          {/* Contact Information */}
          <div className="bg-white rounded-lg shadow-sm p-6">
            <h2 className="text-xl font-semibold text-gray-900 mb-4 flex items-center gap-2">
              <User className="w-5 h-5" />
              Informações de Contato
            </h2>
            <div className="space-y-3">
              {lead.email && (
                <div className="flex items-center gap-3">
                  <Mail className="w-5 h-5 text-gray-400" />
                  <div>
                    <p className="text-sm text-gray-600">Email</p>
                    <a
                      href={`mailto:${lead.email}`}
                      className="text-blue-600 hover:underline"
                    >
                      {lead.email}
                    </a>
                  </div>
                </div>
              )}
              {lead.phone && (
                <div className="flex items-center gap-3">
                  <Phone className="w-5 h-5 text-gray-400" />
                  <div>
                    <p className="text-sm text-gray-600">Telefone</p>
                    <div className="flex items-center gap-2">
                      <a
                        href={`tel:${lead.phone}`}
                        className="text-gray-900"
                      >
                        {lead.phone}
                      </a>
                      <button
                        onClick={openWhatsApp}
                        className="px-3 py-1 bg-green-500 text-white text-sm rounded hover:bg-green-600 transition-colors flex items-center gap-1"
                      >
                        <MessageSquare className="w-3 h-3" />
                        WhatsApp
                      </button>
                    </div>
                  </div>
                </div>
              )}
              <div className="flex items-center gap-3">
                <Calendar className="w-5 h-5 text-gray-400" />
                <div>
                  <p className="text-sm text-gray-600">Data de Criação</p>
                  <p className="text-gray-900">{formatDate(lead.created_at)}</p>
                </div>
              </div>
            </div>
          </div>

          {/* Message */}
          {lead.message && (
            <div className="bg-white rounded-lg shadow-sm p-6">
              <h2 className="text-xl font-semibold text-gray-900 mb-4 flex items-center gap-2">
                <MessageSquare className="w-5 h-5" />
                Mensagem
              </h2>
              <p className="text-gray-700 whitespace-pre-wrap">{lead.message}</p>
            </div>
          )}

          {/* Property Information */}
          {property && (
            <div className="bg-white rounded-lg shadow-sm p-6">
              <h2 className="text-xl font-semibold text-gray-900 mb-4 flex items-center gap-2">
                <Building2 className="w-5 h-5" />
                Imóvel de Interesse
              </h2>
              <div className="border border-gray-200 rounded-lg p-4 hover:border-blue-300 transition-colors">
                <div className="flex items-start justify-between mb-3">
                  <div>
                    <h3 className="font-semibold text-gray-900 text-lg">
                      {property.title || property.slug || 'Imóvel'}
                    </h3>
                    <p className="text-sm text-gray-600 flex items-center gap-1 mt-1">
                      <MapPin className="w-4 h-4" />
                      {property.street}, {property.neighborhood} - {property.city}/{property.state}
                    </p>
                  </div>
                  <span className="text-xl font-bold text-blue-600">
                    {formatPrice(property.sale_price || property.rental_price)}
                  </span>
                </div>
                <div className="flex gap-4 text-sm text-gray-600 mb-3">
                  {property.bedrooms && property.bedrooms > 0 && (
                    <span>{property.bedrooms} quartos</span>
                  )}
                  {property.bathrooms && property.bathrooms > 0 && (
                    <span>{property.bathrooms} banheiros</span>
                  )}
                  {property.area_sqm && property.area_sqm > 0 && (
                    <span>{property.area_sqm}m²</span>
                  )}
                </div>
                <button
                  onClick={() => router.push(`/dashboard/imoveis/${property.id}`)}
                  className="text-blue-600 hover:text-blue-700 text-sm font-medium flex items-center gap-1"
                >
                  Ver detalhes do imóvel
                  <ExternalLink className="w-4 h-4" />
                </button>
              </div>
            </div>
          )}

          {/* UTM Tracking */}
          {(lead.utm_source || lead.utm_campaign || lead.utm_medium || lead.referrer) && (
            <div className="bg-white rounded-lg shadow-sm p-6">
              <h2 className="text-xl font-semibold text-gray-900 mb-4 flex items-center gap-2">
                <TrendingUp className="w-5 h-5" />
                Rastreamento de Campanha
              </h2>
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                {lead.utm_source && (
                  <div>
                    <p className="text-sm text-gray-600">Origem (Source)</p>
                    <p className="font-medium text-gray-900">{lead.utm_source}</p>
                  </div>
                )}
                {lead.utm_campaign && (
                  <div>
                    <p className="text-sm text-gray-600">Campanha</p>
                    <p className="font-medium text-gray-900">{lead.utm_campaign}</p>
                  </div>
                )}
                {lead.utm_medium && (
                  <div>
                    <p className="text-sm text-gray-600">Meio (Medium)</p>
                    <p className="font-medium text-gray-900">{lead.utm_medium}</p>
                  </div>
                )}
                {lead.referrer && (
                  <div className="sm:col-span-2">
                    <p className="text-sm text-gray-600 flex items-center gap-1">
                      <Globe className="w-4 h-4" />
                      Referrer
                    </p>
                    <p className="font-mono text-sm text-gray-700 break-all">{lead.referrer}</p>
                  </div>
                )}
              </div>
            </div>
          )}

          {/* LGPD Compliance */}
          <div className="bg-white rounded-lg shadow-sm p-6">
            <h2 className="text-xl font-semibold text-gray-900 mb-4 flex items-center gap-2">
              <Shield className="w-5 h-5" />
              Conformidade LGPD
            </h2>
            <div className="space-y-3">
              <div className="flex items-start gap-3">
                <div className={`w-2 h-2 rounded-full mt-2 ${lead.consent_given ? 'bg-green-500' : 'bg-red-500'}`}></div>
                <div>
                  <p className="text-sm text-gray-600">Consentimento</p>
                  <p className="font-medium text-gray-900">
                    {lead.consent_given ? 'Consentimento concedido' : 'Sem consentimento'}
                  </p>
                </div>
              </div>
              {lead.consent_text && (
                <div>
                  <p className="text-sm text-gray-600">Texto do Consentimento</p>
                  <p className="text-sm text-gray-700 italic bg-gray-50 p-3 rounded">
                    "{lead.consent_text}"
                  </p>
                </div>
              )}
              {lead.consent_date && (
                <div>
                  <p className="text-sm text-gray-600">Data do Consentimento</p>
                  <p className="text-gray-900">{formatDate(lead.consent_date)}</p>
                </div>
              )}
              {lead.consent_ip && (
                <div>
                  <p className="text-sm text-gray-600">IP do Consentimento</p>
                  <p className="font-mono text-sm text-gray-700">{lead.consent_ip}</p>
                </div>
              )}
            </div>
          </div>
        </div>

        {/* Sidebar */}
        <div className="space-y-6">
          {/* Status Management */}
          <div className="bg-white rounded-lg shadow-sm p-6">
            <h2 className="text-lg font-semibold text-gray-900 mb-4 flex items-center gap-2">
              <Edit className="w-5 h-5" />
              Gerenciar Status
            </h2>
            <div className="space-y-2">
              {Object.values(LeadStatus).map((status) => (
                <button
                  key={status}
                  onClick={() => handleStatusUpdate(status)}
                  disabled={updatingStatus || lead.status === status}
                  className={`w-full px-4 py-2 rounded-lg text-sm font-medium transition-colors ${
                    lead.status === status
                      ? getStatusColor(status)
                      : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
                  } ${updatingStatus ? 'opacity-50 cursor-not-allowed' : ''}`}
                >
                  {getStatusLabel(status)}
                </button>
              ))}
            </div>
          </div>

          {/* Lead Details */}
          <div className="bg-white rounded-lg shadow-sm p-6">
            <h2 className="text-lg font-semibold text-gray-900 mb-4">Detalhes</h2>
            <div className="space-y-3">
              <div>
                <p className="text-sm text-gray-600">Canal</p>
                <p className="font-medium text-gray-900">{getChannelLabel(lead.channel)}</p>
              </div>
              <div>
                <p className="text-sm text-gray-600">Criado em</p>
                <p className="font-medium text-gray-900">{formatDate(lead.created_at)}</p>
              </div>
              {lead.updated_at && (
                <div>
                  <p className="text-sm text-gray-600">Atualizado em</p>
                  <p className="font-medium text-gray-900">{formatDate(lead.updated_at)}</p>
                </div>
              )}
              {lead.broker_id && (
                <div>
                  <p className="text-sm text-gray-600">Corretor Atribuído</p>
                  <p className="font-medium text-gray-900">{lead.broker_id}</p>
                </div>
              )}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
