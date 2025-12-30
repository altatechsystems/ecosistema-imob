'use client';

import { useState, useEffect } from 'react';
import { useParams, useRouter } from 'next/navigation';
import {
  Building2,
  MapPin,
  Bed,
  Bath,
  Maximize,
  Car,
  ArrowLeft,
  Edit,
  Trash2,
  Eye,
  EyeOff,
  Calendar,
  DollarSign,
  Home,
  ChevronLeft,
  ChevronRight
} from 'lucide-react';

interface Owner {
  id: string;
  name?: string;
  email?: string;
  phone?: string;
  document?: string;
  document_type?: string;
  owner_status?: string;
}

interface Property {
  id: string;
  reference?: string;
  slug?: string;
  street?: string;
  city?: string;
  state?: string;
  neighborhood?: string;
  zip_code?: string;
  price_amount?: number;
  bedrooms?: number;
  bathrooms?: number;
  suites?: number;
  parking_spaces?: number;
  total_area?: number;
  built_area?: number;
  property_type?: string;
  transaction_type?: string;
  status?: string;
  visibility?: string;
  description?: string;
  featured?: boolean;
  created_at?: string;
  updated_at?: string;
  canonical_listing_id?: string;
  owner_id?: string;
  captador_name?: string; // Nome do corretor captador
  captador_id?: string;   // ID do broker quando cadastrado
}

interface Listing {
  id: string;
  title?: string;
  description?: string;
  photos?: Photo[];
}

interface Photo {
  id: string;
  url: string;
  thumb_url: string;
  medium_url: string;
  large_url: string;
  order: number;
  is_cover: boolean;
}

export default function PropertyDetailPage() {
  const params = useParams();
  const router = useRouter();
  const propertyId = params?.id as string;

  const [property, setProperty] = useState<Property | null>(null);
  const [owner, setOwner] = useState<Owner | null>(null);
  const [listing, setListing] = useState<Listing | null>(null);
  const [photos, setPhotos] = useState<Photo[]>([]);
  const [currentPhotoIndex, setCurrentPhotoIndex] = useState(0);
  const [loading, setLoading] = useState(true);
  const [loadingOwner, setLoadingOwner] = useState(false);
  const [error, setError] = useState('');

  useEffect(() => {
    if (propertyId) {
      fetchPropertyDetails();
    }
  }, [propertyId]);

  const fetchPropertyDetails = async () => {
    try {
      setLoading(true);
      const tenantId = localStorage.getItem('tenant_id');

      if (!tenantId) {
        setError('Tenant ID não encontrado');
        return;
      }

      // Buscar detalhes do imóvel
      const response = await fetch(`${process.env.NEXT_PUBLIC_API_URL}/${tenantId}/properties/${propertyId}`);

      if (!response.ok) {
        throw new Error('Erro ao buscar detalhes do imóvel');
      }

      const data = await response.json();
      console.log('Property details:', data);

      const propertyData = data.data;
      setProperty(propertyData);

      // Buscar fotos do listing canonical
      if (propertyData.canonical_listing_id) {
        await fetchPhotos(tenantId, propertyData.canonical_listing_id);
      }

      // Buscar dados do proprietário se existir owner_id
      if (propertyData.owner_id) {
        fetchOwnerDetails(tenantId, propertyData.owner_id);
      }
    } catch (err: any) {
      console.error('Erro ao buscar detalhes:', err);
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  const fetchOwnerDetails = async (tenantId: string, ownerId: string) => {
    try {
      setLoadingOwner(true);

      const { auth } = await import('@/lib/firebase');
      const user = auth.currentUser;

      if (!user) {
        console.error('Usuário não autenticado');
        return;
      }

      const token = await user.getIdToken(true);

      const response = await fetch(
        `${process.env.NEXT_PUBLIC_API_URL}/admin/${tenantId}/owners/${ownerId}`,
        {
          headers: {
            'Authorization': `Bearer ${token}`,
          },
        }
      );

      if (!response.ok) {
        throw new Error('Erro ao buscar dados do proprietário');
      }

      const data = await response.json();
      setOwner(data.data);
    } catch (err: any) {
      console.error('Erro ao buscar proprietário:', err);
    } finally {
      setLoadingOwner(false);
    }
  };

  const fetchPhotos = async (tenantId: string, listingId: string) => {
    try {
      const { auth } = await import('@/lib/firebase');
      const user = auth.currentUser;

      if (!user) {
        console.warn('No user authenticated');
        return;
      }

      const token = await user.getIdToken(true);
      const response = await fetch(
        `${process.env.NEXT_PUBLIC_API_URL}/admin/${tenantId}/listings/${listingId}`,
        {
          headers: {
            'Authorization': `Bearer ${token}`,
          },
        }
      );

      if (response.ok) {
        const listingData = await response.json();
        console.log('Listing data:', listingData);
        if (listingData.success && listingData.data) {
          // Store the full listing data
          setListing(listingData.data);
          // Set photos if available
          if (listingData.data.photos) {
            setPhotos(listingData.data.photos);
          }
        }
      }
    } catch (error) {
      console.error('Error fetching listing:', error);
    }
  };

  const formatPrice = (price: number) => {
    return new Intl.NumberFormat('pt-BR', {
      style: 'currency',
      currency: 'BRL',
    }).format(price);
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString('pt-BR');
  };

  const getPropertyTypeLabel = (type: string) => {
    const types: Record<string, string> = {
      'apartment': 'Apartamento',
      'house': 'Casa',
      'condo': 'Condomínio',
      'commercial': 'Comercial',
      'land': 'Terreno',
    };
    return types[type] || type;
  };

  const getTransactionTypeLabel = (type: string) => {
    const types: Record<string, string> = {
      'sale': 'Venda',
      'rent': 'Aluguel',
      'both': 'Venda/Aluguel',
    };
    return types[type] || type;
  };

  const getStatusLabel = (status: string) => {
    const statuses: Record<string, string> = {
      'available': 'Disponível',
      'rented': 'Alugado',
      'sold': 'Vendido',
      'reserved': 'Reservado',
    };
    return statuses[status] || status;
  };

  const getStatusColor = (status: string) => {
    const colors: Record<string, string> = {
      'available': 'bg-green-100 text-green-600',
      'rented': 'bg-blue-100 text-blue-600',
      'sold': 'bg-gray-100 text-gray-600',
      'reserved': 'bg-yellow-100 text-yellow-600',
    };
    return colors[status] || 'bg-gray-100 text-gray-600';
  };

  const nextPhoto = () => {
    setCurrentPhotoIndex((prev) => (prev + 1) % photos.length);
  };

  const prevPhoto = () => {
    setCurrentPhotoIndex((prev) => (prev - 1 + photos.length) % photos.length);
  };

  if (loading) {
    return (
      <div className="p-6">
        <div className="bg-white rounded-lg shadow-sm p-12">
          <div className="text-center">
            <div className="inline-block animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mb-4"></div>
            <p className="text-gray-600">Carregando detalhes do imóvel...</p>
          </div>
        </div>
      </div>
    );
  }

  if (error || !property) {
    return (
      <div className="p-6">
        <div className="bg-red-50 border border-red-200 rounded-lg p-4 mb-6">
          <p className="text-red-600">{error || 'Imóvel não encontrado'}</p>
        </div>
        <button
          onClick={() => router.push('/dashboard/imoveis')}
          className="flex items-center gap-2 text-blue-600 hover:text-blue-700"
        >
          <ArrowLeft className="w-5 h-5" />
          Voltar para lista
        </button>
      </div>
    );
  }

  return (
    <div className="p-4 md:p-6">
      {/* Header */}
      <div className="mb-6">
        <button
          onClick={() => router.push('/dashboard/imoveis')}
          className="flex items-center gap-2 text-gray-600 hover:text-gray-900 mb-4"
        >
          <ArrowLeft className="w-5 h-5" />
          <span className="hidden sm:inline">Voltar para lista</span>
          <span className="sm:hidden">Voltar</span>
        </button>

        <div className="flex flex-col md:flex-row md:items-start md:justify-between gap-4">
          <div className="flex-1">
            <h1 className="text-xl md:text-2xl font-bold text-gray-900 mb-3">
              {property.reference || property.slug || 'Imóvel'}
            </h1>
            <div className="flex flex-wrap items-center gap-2">
              <span className={`px-3 py-1 rounded-full text-xs md:text-sm font-medium ${getStatusColor(property.status || 'available')}`}>
                {getStatusLabel(property.status || 'available')}
              </span>
              <span className="px-3 py-1 bg-gray-100 text-gray-600 rounded-full text-xs md:text-sm font-medium">
                {getPropertyTypeLabel(property.property_type || '')}
              </span>
              <span className="px-3 py-1 bg-gray-100 text-gray-600 rounded-full text-xs md:text-sm font-medium">
                {getTransactionTypeLabel(property.transaction_type || '')}
              </span>
              {property.featured && (
                <span className="px-3 py-1 bg-yellow-100 text-yellow-600 rounded-full text-xs md:text-sm font-medium">
                  Destaque
                </span>
              )}
            </div>
          </div>

          <div className="flex gap-2">
            <button
              onClick={() => router.push(`/dashboard/imoveis/${propertyId}/editar`)}
              className="flex items-center gap-2 px-3 md:px-4 py-2 border border-gray-300 rounded-lg hover:bg-gray-50 transition-colors text-sm md:text-base"
            >
              <Edit className="w-4 h-4 md:w-5 md:h-5" />
              <span className="hidden sm:inline">Editar</span>
            </button>
            <button
              onClick={() => {
                if (confirm('Tem certeza que deseja excluir este imóvel?')) {
                  // TODO: Implementar exclusão
                  console.log('Excluir imóvel:', propertyId);
                }
              }}
              className="flex items-center gap-2 px-3 md:px-4 py-2 border border-red-300 text-red-600 rounded-lg hover:bg-red-50 transition-colors text-sm md:text-base"
            >
              <Trash2 className="w-4 h-4 md:w-5 md:h-5" />
              <span className="hidden sm:inline">Excluir</span>
            </button>
          </div>
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Main Content */}
        <div className="lg:col-span-2 space-y-6">
          {/* Photo Gallery */}
          <div className="bg-white rounded-lg shadow-sm overflow-hidden">
            {photos.length > 0 ? (
              <div className="relative">
                <div className="aspect-video bg-gray-100">
                  <img
                    src={photos[currentPhotoIndex].large_url}
                    alt={`Foto ${currentPhotoIndex + 1}`}
                    className="w-full h-full object-cover"
                  />
                </div>

                {/* Navigation */}
                {photos.length > 1 && (
                  <>
                    <button
                      onClick={prevPhoto}
                      className="absolute left-4 top-1/2 -translate-y-1/2 bg-white/90 hover:bg-white p-2 rounded-full shadow-lg transition-colors"
                    >
                      <ChevronLeft className="w-6 h-6" />
                    </button>
                    <button
                      onClick={nextPhoto}
                      className="absolute right-4 top-1/2 -translate-y-1/2 bg-white/90 hover:bg-white p-2 rounded-full shadow-lg transition-colors"
                    >
                      <ChevronRight className="w-6 h-6" />
                    </button>
                    <div className="absolute bottom-4 left-1/2 -translate-x-1/2 bg-black/60 text-white px-3 py-1 rounded-full text-sm">
                      {currentPhotoIndex + 1} / {photos.length}
                    </div>
                  </>
                )}

                {/* Thumbnails */}
                {photos.length > 1 && (
                  <div className="p-4 bg-gray-50 flex gap-2 overflow-x-auto">
                    {photos.map((photo, index) => (
                      <button
                        key={photo.id}
                        onClick={() => setCurrentPhotoIndex(index)}
                        className={`flex-shrink-0 w-20 h-20 rounded-lg overflow-hidden border-2 transition-all ${
                          index === currentPhotoIndex
                            ? 'border-blue-600 ring-2 ring-blue-200'
                            : 'border-gray-200 hover:border-gray-300'
                        }`}
                      >
                        <img
                          src={photo.thumb_url}
                          alt={`Miniatura ${index + 1}`}
                          className="w-full h-full object-cover"
                        />
                      </button>
                    ))}
                  </div>
                )}
              </div>
            ) : (
              <div className="aspect-video bg-gradient-to-br from-blue-400 to-blue-600 flex items-center justify-center">
                <Building2 className="w-24 h-24 text-white opacity-50" />
              </div>
            )}
          </div>

          {/* Property Details */}
          <div className="bg-white rounded-lg shadow-sm p-4 md:p-6">
            <h2 className="text-lg md:text-xl font-bold text-gray-900 mb-4">Características</h2>
            <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 gap-3 md:gap-4">
              {property.bedrooms && property.bedrooms > 0 && (
                <div className="flex items-center gap-2 md:gap-3">
                  <div className="w-10 h-10 md:w-12 md:h-12 bg-blue-100 rounded-lg flex items-center justify-center flex-shrink-0">
                    <Bed className="w-5 h-5 md:w-6 md:h-6 text-blue-600" />
                  </div>
                  <div>
                    <p className="text-xl md:text-2xl font-bold text-gray-900">{property.bedrooms}</p>
                    <p className="text-xs md:text-sm text-gray-600">Quartos</p>
                  </div>
                </div>
              )}

              {property.bathrooms && property.bathrooms > 0 && (
                <div className="flex items-center gap-2 md:gap-3">
                  <div className="w-10 h-10 md:w-12 md:h-12 bg-blue-100 rounded-lg flex items-center justify-center flex-shrink-0">
                    <Bath className="w-5 h-5 md:w-6 md:h-6 text-blue-600" />
                  </div>
                  <div>
                    <p className="text-xl md:text-2xl font-bold text-gray-900">{property.bathrooms}</p>
                    <p className="text-xs md:text-sm text-gray-600">Banheiros</p>
                  </div>
                </div>
              )}

              {property.suites && property.suites > 0 && (
                <div className="flex items-center gap-2 md:gap-3">
                  <div className="w-10 h-10 md:w-12 md:h-12 bg-purple-100 rounded-lg flex items-center justify-center flex-shrink-0">
                    <Bed className="w-5 h-5 md:w-6 md:h-6 text-purple-600" />
                  </div>
                  <div>
                    <p className="text-xl md:text-2xl font-bold text-gray-900">{property.suites}</p>
                    <p className="text-xs md:text-sm text-gray-600">Suítes</p>
                  </div>
                </div>
              )}

              {property.parking_spaces && property.parking_spaces > 0 && (
                <div className="flex items-center gap-2 md:gap-3">
                  <div className="w-10 h-10 md:w-12 md:h-12 bg-green-100 rounded-lg flex items-center justify-center flex-shrink-0">
                    <Car className="w-5 h-5 md:w-6 md:h-6 text-green-600" />
                  </div>
                  <div>
                    <p className="text-xl md:text-2xl font-bold text-gray-900">{property.parking_spaces}</p>
                    <p className="text-xs md:text-sm text-gray-600">Vagas</p>
                  </div>
                </div>
              )}

              {property.total_area && property.total_area > 0 && (
                <div className="flex items-center gap-2 md:gap-3">
                  <div className="w-10 h-10 md:w-12 md:h-12 bg-orange-100 rounded-lg flex items-center justify-center flex-shrink-0">
                    <Maximize className="w-5 h-5 md:w-6 md:h-6 text-orange-600" />
                  </div>
                  <div>
                    <p className="text-xl md:text-2xl font-bold text-gray-900">{property.total_area}</p>
                    <p className="text-xs md:text-sm text-gray-600">m² Total</p>
                  </div>
                </div>
              )}

              {property.built_area && property.built_area > 0 && (
                <div className="flex items-center gap-2 md:gap-3">
                  <div className="w-10 h-10 md:w-12 md:h-12 bg-yellow-100 rounded-lg flex items-center justify-center flex-shrink-0">
                    <Home className="w-5 h-5 md:w-6 md:h-6 text-yellow-600" />
                  </div>
                  <div>
                    <p className="text-xl md:text-2xl font-bold text-gray-900">{property.built_area}</p>
                    <p className="text-xs md:text-sm text-gray-600">m² Construído</p>
                  </div>
                </div>
              )}
            </div>
          </div>

          {/* Description - Show listing description if available, fallback to property description */}
          {(listing?.description || property.description) && (
            <div className="bg-white rounded-lg shadow-sm p-4 md:p-6">
              <h2 className="text-lg md:text-xl font-bold text-gray-900 mb-4">Descrição</h2>
              {listing?.title && (
                <h3 className="text-base md:text-lg font-semibold text-gray-800 mb-3">{listing.title}</h3>
              )}
              <p className="text-sm md:text-base text-gray-600 whitespace-pre-line leading-relaxed">
                {listing?.description || property.description}
              </p>
            </div>
          )}

          {/* Location */}
          <div className="bg-white rounded-lg shadow-sm p-4 md:p-6">
            <h2 className="text-lg md:text-xl font-bold text-gray-900 mb-4">Localização</h2>
            <div className="space-y-3">
              {property.street && (
                <div className="flex items-start gap-2">
                  <MapPin className="w-5 h-5 text-gray-400 mt-0.5" />
                  <div>
                    <p className="font-medium text-gray-900">{property.street}</p>
                    <p className="text-sm text-gray-600">
                      {property.neighborhood && `${property.neighborhood}, `}
                      {property.city} - {property.state}
                      {property.zip_code && ` - CEP: ${property.zip_code}`}
                    </p>
                  </div>
                </div>
              )}
            </div>
          </div>
        </div>

        {/* Sidebar */}
        <div className="space-y-4 md:space-y-6">
          {/* Price Card */}
          <div className="bg-white rounded-lg shadow-sm p-4 md:p-6">
            <div className="flex items-center gap-2 mb-2">
              <DollarSign className="w-4 h-4 md:w-5 md:h-5 text-gray-400" />
              <span className="text-xs md:text-sm text-gray-600">
                {getTransactionTypeLabel(property.transaction_type || '')}
              </span>
            </div>
            <p className="text-2xl md:text-3xl font-bold text-gray-900">
              {property.price_amount ? formatPrice(property.price_amount) : 'Sob consulta'}
            </p>
          </div>

          {/* Info Card */}
          <div className="bg-white rounded-lg shadow-sm p-4 md:p-6">
            <h3 className="text-base md:text-lg font-bold text-gray-900 mb-4">Informações</h3>
            <div className="space-y-3 text-xs md:text-sm">
              <div className="flex items-center justify-between">
                <span className="text-gray-600">Código</span>
                <span className="font-medium text-gray-900">{property.reference || '-'}</span>
              </div>
              <div className="flex items-center justify-between">
                <span className="text-gray-600">Tipo</span>
                <span className="font-medium text-gray-900">
                  {getPropertyTypeLabel(property.property_type || '')}
                </span>
              </div>
              <div className="flex items-center justify-between">
                <span className="text-gray-600">Status</span>
                <span className="font-medium text-gray-900">
                  {getStatusLabel(property.status || '')}
                </span>
              </div>
              <div className="flex items-center justify-between">
                <span className="text-gray-600">Visibilidade</span>
                <span className="font-medium text-gray-900">
                  {property.visibility === 'public' ? 'Público' : 'Privado'}
                </span>
              </div>
              {property.created_at && (
                <div className="flex items-center justify-between">
                  <span className="text-gray-600">Cadastrado em</span>
                  <span className="font-medium text-gray-900">
                    {formatDate(property.created_at)}
                  </span>
                </div>
              )}
              {property.updated_at && (
                <div className="flex items-center justify-between">
                  <span className="text-gray-600">Atualizado em</span>
                  <span className="font-medium text-gray-900">
                    {formatDate(property.updated_at)}
                  </span>
                </div>
              )}
            </div>
          </div>

          {/* Owner Card */}
          <div className="bg-white rounded-lg shadow-sm p-4 md:p-6">
            <h3 className="text-base md:text-lg font-bold text-gray-900 mb-4">Proprietário</h3>

            {loadingOwner ? (
              <div className="flex items-center justify-center py-8">
                <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
              </div>
            ) : owner ? (
              <div className="space-y-3 text-xs md:text-sm">
                <div>
                  <label className="block text-gray-600 mb-1">Nome</label>
                  <p className="font-medium text-gray-900">{owner.name || 'Não informado'}</p>
                </div>

                {owner.email && (
                  <div>
                    <label className="block text-gray-600 mb-1">E-mail</label>
                    <p className="font-medium text-gray-900">{owner.email}</p>
                  </div>
                )}

                {owner.phone && (
                  <div>
                    <label className="block text-gray-600 mb-1">Telefone</label>
                    <p className="font-medium text-gray-900">{owner.phone}</p>
                  </div>
                )}

                {owner.document && (
                  <div>
                    <label className="block text-gray-600 mb-1">
                      {owner.document_type === 'cnpj' ? 'CNPJ' : 'CPF'}
                    </label>
                    <p className="font-medium text-gray-900">{owner.document}</p>
                  </div>
                )}

                {owner.owner_status && (
                  <div>
                    <label className="block text-gray-600 mb-1">Status</label>
                    <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${
                      owner.owner_status === 'verified' ? 'bg-green-100 text-green-800' :
                      owner.owner_status === 'partial' ? 'bg-yellow-100 text-yellow-800' :
                      'bg-gray-100 text-gray-800'
                    }`}>
                      {owner.owner_status === 'verified' ? 'Verificado' :
                       owner.owner_status === 'partial' ? 'Parcial' :
                       'Incompleto'}
                    </span>
                  </div>
                )}

                <div className="pt-2">
                  <button
                    type="button"
                    onClick={() => router.push(`/dashboard/proprietarios/${property?.owner_id}`)}
                    className="text-xs md:text-sm text-blue-600 hover:text-blue-700 hover:underline"
                  >
                    Ver detalhes do proprietário →
                  </button>
                </div>
              </div>
            ) : property?.owner_id ? (
              <div className="text-gray-500 text-xs md:text-sm py-4">
                <p>Erro ao carregar dados do proprietário</p>
              </div>
            ) : (
              <div className="text-gray-500 text-xs md:text-sm py-4">
                <p>Nenhum proprietário vinculado</p>
              </div>
            )}
          </div>

          {/* Captador Card */}
          {property?.captador_name && (
            <div className="bg-white rounded-lg shadow-sm p-4 md:p-6">
              <h3 className="text-base md:text-lg font-bold text-gray-900 mb-4">Captador</h3>

              <div className="space-y-3 text-xs md:text-sm">
                <div>
                  <label className="block text-gray-600 mb-1">Nome do Corretor</label>
                  <p className="font-medium text-gray-900">{property.captador_name}</p>
                </div>

                {property.captador_id ? (
                  <div className="pt-2">
                    <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-green-100 text-green-800">
                      ✓ Cadastro completo
                    </span>
                    <button
                      type="button"
                      onClick={() => router.push(`/dashboard/corretores/${property.captador_id}`)}
                      className="ml-3 text-xs md:text-sm text-blue-600 hover:text-blue-700 hover:underline"
                    >
                      Ver detalhes do corretor →
                    </button>
                  </div>
                ) : (
                  <div className="pt-2">
                    <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-yellow-100 text-yellow-800">
                      ⚠ Cadastro pendente
                    </span>
                    <p className="text-xs text-gray-600 mt-2">
                      O captador precisa completar o cadastro com CPF e CRECI
                    </p>
                  </div>
                )}
              </div>
            </div>
          )}

          {/* Actions Card */}
          <div className="bg-white rounded-lg shadow-sm p-4 md:p-6">
            <h3 className="text-base md:text-lg font-bold text-gray-900 mb-4">Ações Rápidas</h3>
            <div className="space-y-2">
              <button
                onClick={() => {
                  if (property.slug) {
                    window.open(`${process.env.NEXT_PUBLIC_SITE_URL || 'http://localhost:3000'}/imoveis/${property.slug}`, '_blank');
                  }
                }}
                className="w-full flex items-center justify-center gap-2 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors text-sm md:text-base"
              >
                <Eye className="w-4 h-4 md:w-5 md:h-5" />
                Ver no Site
              </button>
              <button
                onClick={() => router.push(`/dashboard/imoveis/${propertyId}/editar`)}
                className="w-full flex items-center justify-center gap-2 px-4 py-2 border border-gray-300 rounded-lg hover:bg-gray-50 transition-colors text-sm md:text-base"
              >
                <Edit className="w-4 h-4 md:w-5 md:h-5" />
                Editar Imóvel
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
