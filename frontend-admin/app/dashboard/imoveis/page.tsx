'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { Building2, Plus, Search, Filter, MapPin, Bed, Bath, Maximize } from 'lucide-react';

interface Property {
  id: string;
  reference?: string;
  slug?: string;
  street?: string;
  city?: string;
  state?: string;
  neighborhood?: string;
  price_amount?: number;
  bedrooms?: number;
  bathrooms?: number;
  total_area?: number;
  property_type?: string;
  status?: string;
  image_url?: string;
}

export default function ImoveisPage() {
  const router = useRouter();
  const [searchTerm, setSearchTerm] = useState('');
  const [properties, setProperties] = useState<Property[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  useEffect(() => {
    fetchProperties();
  }, []);

  const fetchProperties = async () => {
    try {
      setLoading(true);
      const tenantId = localStorage.getItem('tenant_id');
      console.log('🏢 Tenant ID:', tenantId);
      console.log('🌍 API URL:', process.env.NEXT_PUBLIC_API_URL);

      if (!tenantId) {
        setError('Tenant ID não encontrado');
        return;
      }

      // Add limit parameter to avoid timeout with large datasets
      const url = `${process.env.NEXT_PUBLIC_API_URL}/${tenantId}/properties?limit=1000`;
      console.log('📍 Fetching from:', url);

      const response = await fetch(url);

      if (!response.ok) {
        throw new Error('Erro ao buscar imóveis');
      }

      const data = await response.json();
      console.log('✅ API Response:', data);
      console.log('📊 Total properties:', data.data?.length || 0);

      // Set properties with cover_image_url from the API response
      // The backend already includes cover_image_url for each property
      const propertiesData = data.data || [];
      const propertiesWithImages = propertiesData.map((property: any) => ({
        ...property,
        image_url: property.cover_image_url || undefined,
      }));

      setProperties(propertiesWithImages);
      console.log('✅ Properties loaded with images');
    } catch (err: any) {
      console.error('Erro ao buscar imóveis:', err);
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  const formatPrice = (price: number) => {
    return new Intl.NumberFormat('pt-BR', {
      style: 'currency',
      currency: 'BRL',
    }).format(price);
  };

  const filteredProperties = properties.filter(property =>
    property.reference?.toLowerCase().includes(searchTerm.toLowerCase()) ||
    property.street?.toLowerCase().includes(searchTerm.toLowerCase()) ||
    property.city?.toLowerCase().includes(searchTerm.toLowerCase()) ||
    property.neighborhood?.toLowerCase().includes(searchTerm.toLowerCase())
  );

  return (
    <div className="p-6">
      {/* Header */}
      <div className="mb-6">
        <h1 className="text-2xl font-bold text-gray-900 mb-2">Imóveis</h1>
        <p className="text-gray-600">Gerencie todos os imóveis cadastrados</p>
      </div>

      {/* Stats Cards */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4 mb-6">
        <div className="bg-white rounded-lg shadow-sm p-4">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm text-gray-600">Total de Imóveis</p>
              <p className="text-2xl font-bold text-gray-900">{properties.length}</p>
            </div>
            <div className="w-12 h-12 bg-blue-100 rounded-lg flex items-center justify-center">
              <Building2 className="w-6 h-6 text-blue-600" />
            </div>
          </div>
        </div>

        <div className="bg-white rounded-lg shadow-sm p-4">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm text-gray-600">Disponíveis</p>
              <p className="text-2xl font-bold text-green-600">
                {properties.filter(p => p.status?.toLowerCase() === 'available').length}
              </p>
            </div>
            <div className="w-12 h-12 bg-green-100 rounded-lg flex items-center justify-center">
              <Building2 className="w-6 h-6 text-green-600" />
            </div>
          </div>
        </div>

        <div className="bg-white rounded-lg shadow-sm p-4">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm text-gray-600">Apartamentos</p>
              <p className="text-2xl font-bold text-orange-600">
                {properties.filter(p => p.property_type?.toLowerCase() === 'apartment').length}
              </p>
            </div>
            <div className="w-12 h-12 bg-orange-100 rounded-lg flex items-center justify-center">
              <Building2 className="w-6 h-6 text-orange-600" />
            </div>
          </div>
        </div>

        <div className="bg-white rounded-lg shadow-sm p-4">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm text-gray-600">Casas</p>
              <p className="text-2xl font-bold text-purple-600">
                {properties.filter(p => p.property_type?.toLowerCase() === 'house').length}
              </p>
            </div>
            <div className="w-12 h-12 bg-purple-100 rounded-lg flex items-center justify-center">
              <Building2 className="w-6 h-6 text-purple-600" />
            </div>
          </div>
        </div>
      </div>

      {/* Actions Bar */}
      <div className="bg-white rounded-lg shadow-sm p-4 mb-6">
        <div className="flex flex-col sm:flex-row gap-4 justify-between">
          {/* Search */}
          <div className="flex-1 relative">
            <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400 w-5 h-5" />
            <input
              type="text"
              placeholder="Buscar imóveis por endereço, código, proprietário..."
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              className="w-full pl-10 pr-4 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
          </div>

          {/* Filter Button */}
          <button className="flex items-center gap-2 px-4 py-2 border border-gray-300 rounded-lg hover:bg-gray-50 transition-colors">
            <Filter className="w-5 h-5" />
            Filtros
          </button>

          {/* Add Button */}
          <button className="flex items-center gap-2 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors">
            <Plus className="w-5 h-5" />
            Novo Imóvel
          </button>
        </div>
      </div>

      {/* Loading State */}
      {loading && (
        <div className="bg-white rounded-lg shadow-sm p-12">
          <div className="text-center">
            <div className="inline-block animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mb-4"></div>
            <p className="text-gray-600">Carregando imóveis...</p>
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
      {!loading && !error && filteredProperties.length === 0 && (
        <div className="bg-white rounded-lg shadow-sm p-12">
          <div className="text-center">
            <div className="inline-flex items-center justify-center w-16 h-16 bg-blue-100 rounded-full mb-4">
              <Building2 className="w-8 h-8 text-blue-600" />
            </div>
            <h3 className="text-lg font-semibold text-gray-900 mb-2">
              {searchTerm ? 'Nenhum imóvel encontrado' : 'Nenhum imóvel cadastrado'}
            </h3>
            <p className="text-gray-600 mb-6">
              {searchTerm
                ? 'Tente buscar com outros termos'
                : 'Comece adicionando seu primeiro imóvel para gerenciar seu portfólio'}
            </p>
            {!searchTerm && (
              <button className="inline-flex items-center gap-2 px-6 py-3 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors">
                <Plus className="w-5 h-5" />
                Cadastrar Primeiro Imóvel
              </button>
            )}
          </div>
        </div>
      )}

      {/* Properties Grid */}
      {!loading && !error && filteredProperties.length > 0 && (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6 mb-6">
          {filteredProperties.map((property) => (
            <div key={property.id} className="bg-white rounded-lg shadow-sm overflow-hidden hover:shadow-md transition-shadow">
              {/* Property Image */}
              <div className="h-48 bg-gradient-to-br from-blue-400 to-blue-600 flex items-center justify-center relative overflow-hidden">
                {property.image_url ? (
                  <img
                    src={property.image_url}
                    alt={property.reference || 'Imóvel'}
                    className="w-full h-full object-cover"
                  />
                ) : (
                  <Building2 className="w-16 h-16 text-white opacity-50" />
                )}
              </div>

              {/* Property Info */}
              <div className="p-4">
                <div className="flex items-start justify-between mb-2">
                  <h3 className="text-lg font-semibold text-gray-900 flex-1">
                    {property.reference || property.slug || 'Sem referência'}
                  </h3>
                  <span className="px-2 py-1 bg-blue-100 text-blue-600 text-xs rounded-full">
                    {property.status || 'Disponível'}
                  </span>
                </div>

                <div className="flex items-center text-gray-600 text-sm mb-3">
                  <MapPin className="w-4 h-4 mr-1" />
                  <span className="truncate">
                    {property.street && property.city
                      ? `${property.street}, ${property.city} - ${property.state}`
                      : property.city
                        ? `${property.city} - ${property.state}`
                        : property.neighborhood || 'Endereço não informado'
                    }
                  </span>
                </div>

                <div className="flex items-center gap-4 text-sm text-gray-600 mb-3">
                  {property.bedrooms && property.bedrooms > 0 && (
                    <div className="flex items-center gap-1">
                      <Bed className="w-4 h-4" />
                      <span>{property.bedrooms}</span>
                    </div>
                  )}
                  {property.bathrooms && property.bathrooms > 0 && (
                    <div className="flex items-center gap-1">
                      <Bath className="w-4 h-4" />
                      <span>{property.bathrooms}</span>
                    </div>
                  )}
                  {property.total_area && property.total_area > 0 && (
                    <div className="flex items-center gap-1">
                      <Maximize className="w-4 h-4" />
                      <span>{property.total_area}m²</span>
                    </div>
                  )}
                </div>

                <div className="flex items-center justify-between">
                  <span className="text-xl font-bold text-gray-900">
                    {property.price_amount ? formatPrice(property.price_amount) : 'Preço não informado'}
                  </span>
                  <button
                    onClick={() => router.push(`/dashboard/imoveis/${property.id}`)}
                    className="px-4 py-2 text-sm text-blue-600 hover:bg-blue-50 rounded-lg transition-colors"
                  >
                    Ver detalhes
                  </button>
                </div>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
