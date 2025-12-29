'use client';

import { useState, useEffect, useMemo, useCallback, useRef } from 'react';
import { useRouter } from 'next/navigation';
import Image from 'next/image';
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

type PropertyTypeFilter = 'all' | 'available' | 'apartment' | 'house' | 'chacara' | 'terreno' | 'fazenda' | 'sitio';

export default function ImoveisPage() {
  const router = useRouter();
  const [searchTerm, setSearchTerm] = useState('');
  const [properties, setProperties] = useState<Property[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [displayCount, setDisplayCount] = useState(12);
  const [typeFilter, setTypeFilter] = useState<PropertyTypeFilter>('all');
  const observerTarget = useRef<HTMLDivElement>(null);
  const itemsPerPage = 12;

  const fetchProperties = useCallback(async () => {
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
  }, []);

  useEffect(() => {
    fetchProperties();
  }, [fetchProperties]);

  const formatPrice = (price: number) => {
    return new Intl.NumberFormat('pt-BR', {
      style: 'currency',
      currency: 'BRL',
    }).format(price);
  };

  // Memoize expensive calculations
  const stats = useMemo(() => ({
    total: properties.length,
    available: properties.filter(p => p.status?.toLowerCase() === 'available').length,
    apartments: properties.filter(p => p.property_type?.toLowerCase() === 'apartment').length,
    houses: properties.filter(p => p.property_type?.toLowerCase() === 'house').length,
    chacaras: properties.filter(p => p.reference?.toUpperCase().startsWith('CH')).length,
    terrenos: properties.filter(p => p.reference?.toUpperCase().startsWith('TE')).length,
    fazendas: properties.filter(p => p.reference?.toUpperCase().startsWith('FA')).length,
    sitios: properties.filter(p => p.reference?.toUpperCase().startsWith('ST')).length,
  }), [properties]);

  const filteredProperties = useMemo(() => {
    let filtered = properties;

    // Apply type filter
    if (typeFilter !== 'all') {
      filtered = filtered.filter(property => {
        switch (typeFilter) {
          case 'available':
            return property.status?.toLowerCase() === 'available';
          case 'apartment':
            return property.property_type?.toLowerCase() === 'apartment';
          case 'house':
            return property.property_type?.toLowerCase() === 'house';
          case 'chacara':
            return property.reference?.toUpperCase().startsWith('CH');
          case 'terreno':
            return property.reference?.toUpperCase().startsWith('TE');
          case 'fazenda':
            return property.reference?.toUpperCase().startsWith('FA');
          case 'sitio':
            return property.reference?.toUpperCase().startsWith('ST');
          default:
            return true;
        }
      });
    }

    // Apply search term filter
    if (searchTerm) {
      filtered = filtered.filter(property =>
        property.reference?.toLowerCase().includes(searchTerm.toLowerCase()) ||
        property.street?.toLowerCase().includes(searchTerm.toLowerCase()) ||
        property.city?.toLowerCase().includes(searchTerm.toLowerCase()) ||
        property.neighborhood?.toLowerCase().includes(searchTerm.toLowerCase())
      );
    }

    return filtered;
  }, [properties, searchTerm, typeFilter]);

  // Infinite scroll - show only displayCount items
  const displayedProperties = useMemo(() =>
    filteredProperties.slice(0, displayCount),
    [filteredProperties, displayCount]
  );

  const hasMore = displayCount < filteredProperties.length;

  // Reset display count when search or filter changes
  useEffect(() => {
    setDisplayCount(12);
  }, [searchTerm, typeFilter]);

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

  return (
    <div className="p-3 sm:p-4 md:p-6">
      {/* Stats Cards - Compact */}
      <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-8 gap-2 sm:gap-3 mb-4 sm:mb-6">
        <button
          onClick={() => setTypeFilter('all')}
          className={`bg-white rounded-lg shadow-sm p-2 sm:p-3 text-left transition-all hover:shadow-md ${
            typeFilter === 'all' ? 'ring-2 ring-blue-500' : ''
          }`}
        >
          <div className="flex items-center gap-1.5 sm:gap-2">
            <div className="w-7 h-7 sm:w-8 sm:h-8 bg-blue-100 rounded flex items-center justify-center flex-shrink-0">
              <Building2 className="w-3.5 h-3.5 sm:w-4 sm:h-4 text-blue-600" />
            </div>
            <div className="min-w-0">
              <p className="text-[10px] sm:text-xs text-gray-600 truncate">Total</p>
              <p className="text-base sm:text-lg font-bold text-gray-900">{stats.total}</p>
            </div>
          </div>
        </button>

        <button
          onClick={() => setTypeFilter('available')}
          className={`bg-white rounded-lg shadow-sm p-2 sm:p-3 text-left transition-all hover:shadow-md ${
            typeFilter === 'available' ? 'ring-2 ring-green-500' : ''
          }`}
        >
          <div className="flex items-center gap-1.5 sm:gap-2">
            <div className="w-7 h-7 sm:w-8 sm:h-8 bg-green-100 rounded flex items-center justify-center flex-shrink-0">
              <Building2 className="w-3.5 h-3.5 sm:w-4 sm:h-4 text-green-600" />
            </div>
            <div className="min-w-0">
              <p className="text-[10px] sm:text-xs text-gray-600 truncate">Disponíveis</p>
              <p className="text-base sm:text-lg font-bold text-green-600">{stats.available}</p>
            </div>
          </div>
        </button>

        <button
          onClick={() => setTypeFilter('apartment')}
          className={`bg-white rounded-lg shadow-sm p-2 sm:p-3 text-left transition-all hover:shadow-md ${
            typeFilter === 'apartment' ? 'ring-2 ring-orange-500' : ''
          }`}
        >
          <div className="flex items-center gap-1.5 sm:gap-2">
            <div className="w-7 h-7 sm:w-8 sm:h-8 bg-orange-100 rounded flex items-center justify-center flex-shrink-0">
              <Building2 className="w-3.5 h-3.5 sm:w-4 sm:h-4 text-orange-600" />
            </div>
            <div className="min-w-0">
              <p className="text-[10px] sm:text-xs text-gray-600 truncate">Aptos</p>
              <p className="text-base sm:text-lg font-bold text-orange-600">{stats.apartments}</p>
            </div>
          </div>
        </button>

        <button
          onClick={() => setTypeFilter('house')}
          className={`bg-white rounded-lg shadow-sm p-2 sm:p-3 text-left transition-all hover:shadow-md ${
            typeFilter === 'house' ? 'ring-2 ring-purple-500' : ''
          }`}
        >
          <div className="flex items-center gap-1.5 sm:gap-2">
            <div className="w-7 h-7 sm:w-8 sm:h-8 bg-purple-100 rounded flex items-center justify-center flex-shrink-0">
              <Building2 className="w-3.5 h-3.5 sm:w-4 sm:h-4 text-purple-600" />
            </div>
            <div className="min-w-0">
              <p className="text-[10px] sm:text-xs text-gray-600 truncate">Casas</p>
              <p className="text-base sm:text-lg font-bold text-purple-600">{stats.houses}</p>
            </div>
          </div>
        </button>

        <button
          onClick={() => setTypeFilter('chacara')}
          className={`bg-white rounded-lg shadow-sm p-2 sm:p-3 text-left transition-all hover:shadow-md ${
            typeFilter === 'chacara' ? 'ring-2 ring-teal-500' : ''
          }`}
        >
          <div className="flex items-center gap-1.5 sm:gap-2">
            <div className="w-7 h-7 sm:w-8 sm:h-8 bg-teal-100 rounded flex items-center justify-center flex-shrink-0">
              <Building2 className="w-3.5 h-3.5 sm:w-4 sm:h-4 text-teal-600" />
            </div>
            <div className="min-w-0">
              <p className="text-[10px] sm:text-xs text-gray-600 truncate">Chácaras</p>
              <p className="text-base sm:text-lg font-bold text-teal-600">{stats.chacaras}</p>
            </div>
          </div>
        </button>

        <button
          onClick={() => setTypeFilter('terreno')}
          className={`bg-white rounded-lg shadow-sm p-2 sm:p-3 text-left transition-all hover:shadow-md ${
            typeFilter === 'terreno' ? 'ring-2 ring-amber-500' : ''
          }`}
        >
          <div className="flex items-center gap-1.5 sm:gap-2">
            <div className="w-7 h-7 sm:w-8 sm:h-8 bg-amber-100 rounded flex items-center justify-center flex-shrink-0">
              <Building2 className="w-3.5 h-3.5 sm:w-4 sm:h-4 text-amber-600" />
            </div>
            <div className="min-w-0">
              <p className="text-[10px] sm:text-xs text-gray-600 truncate">Terrenos</p>
              <p className="text-base sm:text-lg font-bold text-amber-600">{stats.terrenos}</p>
            </div>
          </div>
        </button>

        <button
          onClick={() => setTypeFilter('fazenda')}
          className={`bg-white rounded-lg shadow-sm p-2 sm:p-3 text-left transition-all hover:shadow-md ${
            typeFilter === 'fazenda' ? 'ring-2 ring-emerald-500' : ''
          }`}
        >
          <div className="flex items-center gap-1.5 sm:gap-2">
            <div className="w-7 h-7 sm:w-8 sm:h-8 bg-emerald-100 rounded flex items-center justify-center flex-shrink-0">
              <Building2 className="w-3.5 h-3.5 sm:w-4 sm:h-4 text-emerald-600" />
            </div>
            <div className="min-w-0">
              <p className="text-[10px] sm:text-xs text-gray-600 truncate">Fazendas</p>
              <p className="text-base sm:text-lg font-bold text-emerald-600">{stats.fazendas}</p>
            </div>
          </div>
        </button>

        <button
          onClick={() => setTypeFilter('sitio')}
          className={`bg-white rounded-lg shadow-sm p-2 sm:p-3 text-left transition-all hover:shadow-md ${
            typeFilter === 'sitio' ? 'ring-2 ring-lime-500' : ''
          }`}
        >
          <div className="flex items-center gap-1.5 sm:gap-2">
            <div className="w-7 h-7 sm:w-8 sm:h-8 bg-lime-100 rounded flex items-center justify-center flex-shrink-0">
              <Building2 className="w-3.5 h-3.5 sm:w-4 sm:h-4 text-lime-600" />
            </div>
            <div className="min-w-0">
              <p className="text-[10px] sm:text-xs text-gray-600 truncate">Sítios</p>
              <p className="text-base sm:text-lg font-bold text-lime-600">{stats.sitios}</p>
            </div>
          </div>
        </button>
      </div>

      {/* Actions Bar */}
      <div className="bg-white rounded-lg shadow-sm p-3 sm:p-4 mb-4 sm:mb-6">
        <div className="flex flex-col sm:flex-row gap-2 sm:gap-4 justify-between">
          {/* Search */}
          <div className="flex-1 relative">
            <Search className="absolute left-2.5 sm:left-3 top-1/2 transform -translate-y-1/2 text-gray-400 w-4 h-4 sm:w-5 sm:h-5" />
            <input
              type="text"
              placeholder="Buscar imóveis..."
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

          {/* Add Button */}
          <button
            onClick={() => router.push('/dashboard/imoveis/novo')}
            className="flex items-center justify-center gap-2 px-3 sm:px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors text-sm sm:text-base"
          >
            <Plus className="w-4 h-4 sm:w-5 sm:h-5" />
            <span className="hidden sm:inline">Novo Imóvel</span>
            <span className="sm:hidden">Novo</span>
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
        <>
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3 sm:gap-4 md:gap-6 mb-4 sm:mb-6">
          {displayedProperties.map((property) => (
            <div
              key={property.id}
              onClick={() => router.push(`/dashboard/imoveis/${property.id}`)}
              className="bg-white rounded-lg shadow-sm overflow-hidden hover:shadow-md transition-all cursor-pointer group"
            >
              {/* Property Image */}
              <div className="h-48 bg-gradient-to-br from-blue-400 to-blue-600 flex items-center justify-center relative overflow-hidden group-hover:brightness-95 transition-all">
                {property.image_url ? (
                  <Image
                    src={property.image_url}
                    alt={property.reference || 'Imóvel'}
                    fill
                    sizes="(max-width: 768px) 100vw, (max-width: 1024px) 50vw, 33vw"
                    className="object-cover"
                    loading="lazy"
                    quality={75}
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
                  <span className="px-4 py-2 text-sm text-blue-600 group-hover:bg-blue-50 rounded-lg transition-colors">
                    Ver detalhes
                  </span>
                </div>
              </div>
            </div>
          ))}
        </div>

        {/* Infinite Scroll Trigger & Loading Indicator */}
        <div ref={observerTarget} className="py-8">
          {hasMore && (
            <div className="text-center">
              <div className="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
              <p className="text-gray-600 mt-2">Carregando mais imóveis...</p>
            </div>
          )}
          {!hasMore && filteredProperties.length > 12 && (
            <div className="text-center">
              <p className="text-gray-600">
                Mostrando todos os {filteredProperties.length} imóveis
              </p>
            </div>
          )}
        </div>
        </>
      )}
    </div>
  );
}
