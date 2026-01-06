'use client';

import { useEffect, useState } from 'react';
import { useParams, useRouter } from 'next/navigation';
import { ArrowLeft, User, Phone, Mail, MapPin, Home, AlertCircle, Edit2, Save, X } from 'lucide-react';

interface Owner {
  id: string;
  name: string;
  phone?: string;
  email?: string;
  address?: string;
  city?: string;
  state?: string;
  notes?: string;
  data_source?: string;
  data_quality?: string;
  created_at?: string;
  updated_at?: string;
}

interface Property {
  id: string;
  reference: string;
  property_type: string;
  price_amount: number;
  price_currency: string;
  street?: string;
  number?: string;
  neighborhood?: string;
  city: string;
  state: string;
  cover_image_url?: string;
  images?: Array<{ url: string; order: number }>;
  bedrooms?: number;
  bathrooms?: number;
  total_area?: number;
}

export default function OwnerDetailsPage() {
  const params = useParams();
  const router = useRouter();
  const ownerId = params.id as string;

  const [owner, setOwner] = useState<Owner | null>(null);
  const [properties, setProperties] = useState<Property[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [isEditing, setIsEditing] = useState(false);
  const [isSaving, setIsSaving] = useState(false);
  const [editForm, setEditForm] = useState<Partial<Owner>>({});

  useEffect(() => {
    const fetchOwnerData = async () => {
      try {
        setLoading(true);
        setError(null);

        const tenantId = localStorage.getItem('tenant_id');
        if (!tenantId) {
          throw new Error('Tenant ID não encontrado');
        }

        // Get Firebase auth token
        const { auth } = await import('@/lib/firebase');
        const user = auth.currentUser;
        if (!user) {
          throw new Error('Usuário não autenticado');
        }
        const token = await user.getIdToken();

        // Fetch owner details
        const ownerResponse = await fetch(
          `${process.env.NEXT_PUBLIC_API_URL}/admin/${tenantId}/owners/${ownerId}`,
          {
            headers: {
              'Authorization': `Bearer ${token}`,
            },
          }
        );

        if (!ownerResponse.ok) {
          if (ownerResponse.status === 404) {
            throw new Error('Proprietário não encontrado');
          }
          throw new Error('Erro ao carregar dados do proprietário');
        }

        const ownerData = await ownerResponse.json();
        // API returns {success: true, data: owner}
        setOwner(ownerData.data || ownerData);

        // Fetch properties owned by this owner
        const propertiesResponse = await fetch(
          `${process.env.NEXT_PUBLIC_API_URL}/admin/${tenantId}/properties?owner_id=${ownerId}`,
          {
            headers: {
              'Authorization': `Bearer ${token}`,
            },
          }
        );

        if (propertiesResponse.ok) {
          const propertiesData = await propertiesResponse.json();
          // API returns {success: true, data: properties}
          setProperties(propertiesData.data || propertiesData.properties || []);
        }
      } catch (err: any) {
        setError(err.message || 'Erro ao carregar dados');
      } finally{
        setLoading(false);
      }
    };

    if (ownerId) {
      fetchOwnerData();
    }
  }, [ownerId]);

  const handleStartEdit = () => {
    setEditForm({
      name: owner?.name || '',
      phone: owner?.phone || '',
      email: owner?.email || '',
      address: owner?.address || '',
      city: owner?.city || '',
      state: owner?.state || '',
      notes: owner?.notes || '',
    });
    setIsEditing(true);
  };

  const handleCancelEdit = () => {
    setEditForm({});
    setIsEditing(false);
  };

  const handleSaveEdit = async () => {
    try {
      setIsSaving(true);
      setError(null);

      const tenantId = localStorage.getItem('tenant_id');
      if (!tenantId) {
        throw new Error('Tenant ID não encontrado');
      }

      // Get Firebase auth token
      const { auth } = await import('@/lib/firebase');
      const user = auth.currentUser;
      if (!user) {
        throw new Error('Usuário não autenticado');
      }
      const token = await user.getIdToken();

      const response = await fetch(
        `${process.env.NEXT_PUBLIC_API_URL}/admin/${tenantId}/owners/${ownerId}`,
        {
          method: 'PUT',
          headers: {
            'Authorization': `Bearer ${token}`,
            'Content-Type': 'application/json',
          },
          body: JSON.stringify(editForm),
        }
      );

      if (!response.ok) {
        throw new Error('Erro ao atualizar proprietário');
      }

      // Update local state
      setOwner({ ...owner!, ...editForm });
      setIsEditing(false);
      setEditForm({});
    } catch (err: any) {
      setError(err.message || 'Erro ao atualizar dados');
    } finally {
      setIsSaving(false);
    }
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"></div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="p-6">
        <button
          onClick={() => router.back()}
          className="mb-6 flex items-center gap-2 text-gray-600 hover:text-gray-900"
        >
          <ArrowLeft className="w-4 h-4" />
          Voltar
        </button>
        <div className="bg-red-50 border border-red-200 rounded-lg p-6">
          <div className="flex items-center gap-3">
            <AlertCircle className="w-6 h-6 text-red-600 flex-shrink-0" />
            <div>
              <h3 className="font-semibold text-red-900">Erro ao carregar dados</h3>
              <p className="text-sm text-red-700 mt-1">{error}</p>
            </div>
          </div>
        </div>
      </div>
    );
  }

  if (!owner) {
    return (
      <div className="p-6">
        <button
          onClick={() => router.back()}
          className="mb-6 flex items-center gap-2 text-gray-600 hover:text-gray-900"
        >
          <ArrowLeft className="w-4 h-4" />
          Voltar
        </button>
        <div className="bg-yellow-50 border border-yellow-200 rounded-lg p-6">
          <p className="text-yellow-900">Proprietário não encontrado</p>
        </div>
      </div>
    );
  }

  const getDataQualityBadge = (quality?: string) => {
    switch (quality?.toLowerCase()) {
      case 'completo':
        return <span className="px-2 py-1 text-xs font-medium bg-green-100 text-green-800 rounded">Completo</span>;
      case 'parcial':
        return <span className="px-2 py-1 text-xs font-medium bg-yellow-100 text-yellow-800 rounded">Parcial</span>;
      case 'placeholder':
        return <span className="px-2 py-1 text-xs font-medium bg-gray-100 text-gray-800 rounded">Placeholder</span>;
      default:
        return <span className="px-2 py-1 text-xs font-medium bg-gray-100 text-gray-600 rounded">Desconhecido</span>;
    }
  };

  return (
    <div className="p-6 max-w-6xl mx-auto">
      {/* Header */}
      <button
        onClick={() => router.back()}
        className="mb-6 flex items-center gap-2 text-gray-600 hover:text-gray-900"
      >
        <ArrowLeft className="w-4 h-4" />
        Voltar
      </button>

      <div className="mb-6 flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900 mb-2">Detalhes do Proprietário</h1>
          <p className="text-gray-600">Informações completas e imóveis associados</p>
        </div>
        {!isEditing ? (
          <button
            onClick={handleStartEdit}
            className="flex items-center gap-2 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors"
          >
            <Edit2 className="w-4 h-4" />
            Editar
          </button>
        ) : (
          <div className="flex items-center gap-2">
            <button
              onClick={handleCancelEdit}
              disabled={isSaving}
              className="flex items-center gap-2 px-4 py-2 bg-gray-200 text-gray-700 rounded-lg hover:bg-gray-300 transition-colors disabled:opacity-50"
            >
              <X className="w-4 h-4" />
              Cancelar
            </button>
            <button
              onClick={handleSaveEdit}
              disabled={isSaving}
              className="flex items-center gap-2 px-4 py-2 bg-green-600 text-white rounded-lg hover:bg-green-700 transition-colors disabled:opacity-50"
            >
              <Save className="w-4 h-4" />
              {isSaving ? 'Salvando...' : 'Salvar'}
            </button>
          </div>
        )}
      </div>

      {/* Owner Information */}
      <div className="bg-white rounded-lg shadow-sm p-6 mb-6">
        <div className="flex items-start justify-between mb-6">
          <div className="flex items-center gap-4">
            <div className="w-16 h-16 bg-blue-100 rounded-full flex items-center justify-center">
              <User className="w-8 h-8 text-blue-600" />
            </div>
            <div className="flex-1">
              {isEditing ? (
                <input
                  type="text"
                  value={editForm.name || ''}
                  onChange={(e) => setEditForm({ ...editForm, name: e.target.value })}
                  className="text-xl font-bold text-gray-900 border-b-2 border-blue-500 focus:outline-none w-full"
                  placeholder="Nome do proprietário"
                />
              ) : (
                <h2 className="text-xl font-bold text-gray-900">{owner.name}</h2>
              )}
              <div className="flex items-center gap-2 mt-1">
                {getDataQualityBadge(owner.data_quality)}
                {owner.data_source && (
                  <span className="text-xs text-gray-500">
                    Fonte: {owner.data_source}
                  </span>
                )}
              </div>
            </div>
          </div>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          {/* Contact Information */}
          <div>
            <h3 className="text-sm font-semibold text-gray-700 mb-3">Contato</h3>
            <div className="space-y-3">
              <div className="flex items-center gap-3">
                <Phone className="w-4 h-4 text-gray-400 flex-shrink-0" />
                {isEditing ? (
                  <input
                    type="tel"
                    value={editForm.phone || ''}
                    onChange={(e) => setEditForm({ ...editForm, phone: e.target.value })}
                    className="text-sm text-gray-900 border-b border-gray-300 focus:border-blue-500 focus:outline-none flex-1"
                    placeholder="Telefone"
                  />
                ) : (
                  <span className={`text-sm ${owner.phone ? 'text-gray-900' : 'text-gray-400'}`}>
                    {owner.phone || 'Telefone não informado'}
                  </span>
                )}
              </div>

              <div className="flex items-center gap-3">
                <Mail className="w-4 h-4 text-gray-400 flex-shrink-0" />
                {isEditing ? (
                  <input
                    type="email"
                    value={editForm.email || ''}
                    onChange={(e) => setEditForm({ ...editForm, email: e.target.value })}
                    className="text-sm text-gray-900 border-b border-gray-300 focus:border-blue-500 focus:outline-none flex-1"
                    placeholder="Email"
                  />
                ) : (
                  <span className={`text-sm ${owner.email ? 'text-gray-900' : 'text-gray-400'}`}>
                    {owner.email || 'Email não informado'}
                  </span>
                )}
              </div>
            </div>
          </div>

          {/* Address Information */}
          <div>
            <h3 className="text-sm font-semibold text-gray-700 mb-3">Endereço</h3>
            <div className="space-y-3">
              {isEditing ? (
                <div className="space-y-2">
                  <div className="flex items-center gap-3">
                    <MapPin className="w-4 h-4 text-gray-400 flex-shrink-0" />
                    <input
                      type="text"
                      value={editForm.address || ''}
                      onChange={(e) => setEditForm({ ...editForm, address: e.target.value })}
                      className="text-sm text-gray-900 border-b border-gray-300 focus:border-blue-500 focus:outline-none flex-1"
                      placeholder="Endereço"
                    />
                  </div>
                  <div className="flex items-center gap-3 ml-7">
                    <input
                      type="text"
                      value={editForm.city || ''}
                      onChange={(e) => setEditForm({ ...editForm, city: e.target.value })}
                      className="text-sm text-gray-900 border-b border-gray-300 focus:border-blue-500 focus:outline-none flex-1"
                      placeholder="Cidade"
                    />
                    <input
                      type="text"
                      value={editForm.state || ''}
                      onChange={(e) => setEditForm({ ...editForm, state: e.target.value })}
                      className="text-sm text-gray-900 border-b border-gray-300 focus:border-blue-500 focus:outline-none w-20"
                      placeholder="UF"
                      maxLength={2}
                    />
                  </div>
                </div>
              ) : (
                <>
                  {owner.address || owner.city || owner.state ? (
                    <div className="flex items-start gap-3">
                      <MapPin className="w-4 h-4 text-gray-400 mt-0.5" />
                      <div className="text-sm text-gray-900">
                        {owner.address && <div>{owner.address}</div>}
                        {(owner.city || owner.state) && (
                          <div className="text-gray-600">
                            {[owner.city, owner.state].filter(Boolean).join(' - ')}
                          </div>
                        )}
                      </div>
                    </div>
                  ) : (
                    <div className="flex items-center gap-3">
                      <MapPin className="w-4 h-4 text-gray-300" />
                      <span className="text-sm text-gray-400">Endereço não informado</span>
                    </div>
                  )}
                </>
              )}
            </div>
          </div>
        </div>

        {/* Notes */}
        {(owner.notes || isEditing) && (
          <div className="mt-6 pt-6 border-t border-gray-200">
            <h3 className="text-sm font-semibold text-gray-700 mb-2">Observações</h3>
            {isEditing ? (
              <textarea
                value={editForm.notes || ''}
                onChange={(e) => setEditForm({ ...editForm, notes: e.target.value })}
                className="w-full text-sm text-gray-600 border border-gray-300 rounded-lg p-2 focus:border-blue-500 focus:outline-none min-h-[100px]"
                placeholder="Adicione observações sobre o proprietário..."
              />
            ) : (
              <p className="text-sm text-gray-600 whitespace-pre-wrap">{owner.notes}</p>
            )}
          </div>
        )}

        {/* Metadata */}
        <div className="mt-6 pt-6 border-t border-gray-200">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4 text-xs text-gray-500">
            {owner.created_at && (
              <div>
                <span className="font-medium">Cadastrado em:</span>{' '}
                {new Date(owner.created_at).toLocaleString('pt-BR')}
              </div>
            )}
            {owner.updated_at && (
              <div>
                <span className="font-medium">Atualizado em:</span>{' '}
                {new Date(owner.updated_at).toLocaleString('pt-BR')}
              </div>
            )}
          </div>
        </div>
      </div>

      {/* Properties */}
      <div className="bg-white rounded-lg shadow-sm p-6">
        <div className="flex items-center gap-2 mb-4">
          <Home className="w-5 h-5 text-gray-700" />
          <h3 className="text-lg font-semibold text-gray-900">
            Imóveis ({properties.length})
          </h3>
        </div>

        {properties.length === 0 ? (
          <div className="text-center py-8 text-gray-500">
            <Home className="w-12 h-12 mx-auto mb-3 text-gray-300" />
            <p>Nenhum imóvel associado a este proprietário</p>
          </div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            {properties.map((property) => {
              const imageUrl = property.cover_image_url || (property.images && property.images.length > 0 ? property.images[0].url : null);
              const address = [property.street, property.number].filter(Boolean).join(', ') || property.neighborhood;
              const propertyTypeLabels: Record<string, string> = {
                'apartment': 'Apartamento',
                'house': 'Casa',
                'land': 'Terreno',
                'commercial': 'Comercial',
              };
              const typeLabel = propertyTypeLabels[property.property_type] || property.property_type;

              return (
                <div
                  key={property.id}
                  onClick={() => router.push(`/dashboard/imoveis/${property.id}`)}
                  className="border border-gray-200 rounded-lg overflow-hidden hover:shadow-md transition-shadow cursor-pointer"
                >
                  {imageUrl ? (
                    <img
                      src={imageUrl}
                      alt={property.reference}
                      className="w-full h-40 object-cover"
                    />
                  ) : (
                    <div className="w-full h-40 bg-gray-100 flex items-center justify-center">
                      <Home className="w-12 h-12 text-gray-300" />
                    </div>
                  )}
                  <div className="p-4">
                    <div className="flex items-center justify-between mb-2">
                      <span className="text-xs font-medium text-gray-500">{property.reference}</span>
                      <span className="text-xs px-2 py-1 bg-blue-100 text-blue-800 rounded">
                        {typeLabel}
                      </span>
                    </div>
                    <div className="flex items-center gap-2 mb-2 text-sm text-gray-600">
                      {property.bedrooms && <span>{property.bedrooms} quartos</span>}
                      {property.bathrooms && <span>• {property.bathrooms} banheiros</span>}
                      {property.total_area && <span>• {property.total_area}m²</span>}
                    </div>
                    <p className="text-sm text-gray-600 mb-2 line-clamp-1">
                      {address}, {property.city} - {property.state}
                    </p>
                    <p className="text-lg font-bold text-blue-600">
                      {new Intl.NumberFormat('pt-BR', {
                        style: 'currency',
                        currency: property.price_currency || 'BRL',
                      }).format(property.price_amount)}
                    </p>
                  </div>
                </div>
              );
            })}
          </div>
        )}
      </div>
    </div>
  );
}
