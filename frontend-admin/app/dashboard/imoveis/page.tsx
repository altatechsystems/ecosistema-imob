'use client';

import { useState, useEffect, useMemo, useCallback, useRef } from 'react';
import { useRouter } from 'next/navigation';
import Image from 'next/image';
import { Building2, Plus, Search, Filter, MapPin, Bed, Bath, Maximize, AlertCircle } from 'lucide-react';

interface Property {
  id: string;
  reference?: string;
  slug?: string;
  title?: string;         // Computed from canonical listing
  description?: string;   // Computed from canonical listing
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
  status_confirmed_at?: string;  // PROMPT 08
  pending_reason?: string;       // PROMPT 08
  image_url?: string;
}

type PropertyTypeFilter = 'all' | 'available' | 'pending_confirmation' | 'apartment' | 'house' | 'chacara' | 'terreno' | 'fazenda' | 'sitio';

export default function ImoveisPage() {
  const router = useRouter();
  const [searchTerm, setSearchTerm] = useState('');
  const [properties, setProperties] = useState<Property[]>([]);
  const [totalCount, setTotalCount] = useState(0); // Total properties in database
  const [loading, setLoading] = useState(true);
  const [loadingMore, setLoadingMore] = useState(false);
  const [error, setError] = useState('');
  const [displayCount, setDisplayCount] = useState(12);
  const [typeFilter, setTypeFilter] = useState<PropertyTypeFilter>('all');
  const [hasMoreToFetch, setHasMoreToFetch] = useState(true);
  const [currentPage, setCurrentPage] = useState(0);
  const [apiStats, setApiStats] = useState<any>(null); // Stats from API
  const observerTarget = useRef<HTMLDivElement>(null);
  const itemsPerPage = 12;
  const pageSize = 200; // Fetch 200 properties per API call

  const fetchProperties = useCallback(async (page: number = 0, append: boolean = false) => {
    try {
      if (page === 0) setLoading(true);
      else setLoadingMore(true);

      const tenantId = localStorage.getItem('tenant_id');

      if (!tenantId) {
        setError('Tenant ID não encontrado');
        return;
      }

      const startTime = performance.now();
      const offset = page * pageSize;
      const url = `${process.env.NEXT_PUBLIC_API_URL}/${tenantId}/properties?limit=${pageSize}&offset=${offset}`;

      const response = await fetch(url);

      if (!response.ok) {
        throw new Error('Erro ao buscar imóveis');
      }

      const data = await response.json();

      // Store total count and stats from API (only on first page)
      if (page === 0) {
        if (data.total) {
          setTotalCount(data.total);
        }
        if (data.stats) {
          setApiStats(data.stats);
        }
      }

      // Optimize: Only process essential fields
      const propertiesData = data.data || [];
      const optimizedProperties = propertiesData.map((property: any) => ({
        id: property.id,
        reference: property.reference,
        slug: property.slug,
        title: property.title,           // Computed from canonical listing
        description: property.description, // Computed from canonical listing
        street: property.street,
        city: property.city,
        state: property.state,
        neighborhood: property.neighborhood,
        price_amount: property.price_amount,
        bedrooms: property.bedrooms,
        bathrooms: property.bathrooms,
        total_area: property.total_area,
        property_type: property.property_type,
        status: property.status,
        status_confirmed_at: property.status_confirmed_at, // PROMPT 08
        pending_reason: property.pending_reason,           // PROMPT 08
        image_url: property.cover_image_url,
      }));

      if (append) {
        setProperties(prev => [...prev, ...optimizedProperties]);
      } else {
        setProperties(optimizedProperties);
      }

      // Check if there are more properties to fetch
      setHasMoreToFetch(optimizedProperties.length === pageSize);
      setCurrentPage(page);
    } catch (err: any) {
      setError(err.message);
    } finally {
      setLoading(false);
      setLoadingMore(false);
    }
  }, [pageSize]);

  useEffect(() => {
    fetchProperties(0, false);
  }, []);

  // Auto-fetch next pages in background
  useEffect(() => {
    if (!loading && !loadingMore && hasMoreToFetch && properties.length > 0) {
      const timer = setTimeout(() => {
        fetchProperties(currentPage + 1, true);
      }, 500); // Wait 500ms before fetching next page
      return () => clearTimeout(timer);
    }
  }, [loading, loadingMore, hasMoreToFetch, properties.length, currentPage, fetchProperties]);

  const formatPrice = (price: number) => {
    return new Intl.NumberFormat('pt-BR', {
      style: 'currency',
      currency: 'BRL',
    }).format(price);
  };

  const getPropertyTypeLabel = (type: string) => {
    const types: Record<string, string> = {
      'apartment': 'Apartamento',
      'house': 'Casa',
      'condo': 'Condomínio',
      'commercial': 'Comercial',
      'land': 'Terreno',
      'chacara': 'Chácara',
      'fazenda': 'Fazenda',
      'sitio': 'Sítio',
    };
    return types[type] || type;
  };

  // Check if all data is loaded
  const allDataLoaded = !hasMoreToFetch && !loadingMore;

  // Memoize expensive calculations - optimize by doing single pass
  const stats = useMemo(() => {
    const result = {
      total: properties.length,
      available: 0,
      pending_confirmation: 0, // PROMPT 08
      apartments: 0,
      houses: 0,
      chacaras: 0,
      terrenos: 0,
      fazendas: 0,
      sitios: 0,
    };

    // Single pass through properties
    properties.forEach(p => {
      if (p.status?.toLowerCase() === 'available') result.available++;
      if (p.status?.toLowerCase() === 'pending_confirmation') result.pending_confirmation++; // PROMPT 08
      if (p.property_type?.toLowerCase() === 'apartment') result.apartments++;
      if (p.property_type?.toLowerCase() === 'house') result.houses++;

      const ref = p.reference?.toUpperCase();
      if (ref?.startsWith('CH')) result.chacaras++;
      else if (ref?.startsWith('TE')) result.terrenos++;
      else if (ref?.startsWith('FA')) result.fazendas++;
      else if (ref?.startsWith('ST')) result.sitios++;
    });

    return result;
  }, [properties]);

  const filteredProperties = useMemo(() => {
    let filtered = properties;

    // Apply type filter
    if (typeFilter !== 'all') {
      filtered = filtered.filter(property => {
        switch (typeFilter) {
          case 'available':
            return property.status?.toLowerCase() === 'available';
          case 'pending_confirmation': // PROMPT 08
            return property.status?.toLowerCase() === 'pending_confirmation';
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

  // Auto-increase display count when more properties are loaded and we're near the end
  useEffect(() => {
    if (displayCount >= filteredProperties.length && filteredProperties.length > 0) {
      // If we're showing all available, increase display to accommodate new data
      const shouldExpand = properties.length > filteredProperties.length;
      if (shouldExpand) {
        setDisplayCount(prev => prev + itemsPerPage);
      }
    }
  }, [properties.length, filteredProperties.length, displayCount, itemsPerPage]);

  // Infinite scroll observer
  useEffect(() => {
    const observer = new IntersectionObserver(
      (entries) => {
        if (entries[0].isIntersecting && hasMore && !loading && !loadingMore) {
          setDisplayCount(prev => Math.min(prev + itemsPerPage, filteredProperties.length));
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
  }, [hasMore, loading, loadingMore, itemsPerPage, displayCount, filteredProperties.length]);

  return (
    <div className="p-3 sm:p-4 md:p-6">
      {/* Status Cards - First Row */}
      <div className="grid grid-cols-2 sm:grid-cols-3 gap-3 sm:gap-4 mb-3 sm:mb-4">
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
              <p className="text-[10px] sm:text-xs text-gray-600 truncate">
                Total {loadingMore && <span className="text-blue-600">⏳</span>}
              </p>
              <p className="text-base sm:text-lg font-bold text-gray-900">
                {totalCount > 0 ? totalCount : stats.total}
              </p>
              {totalCount > properties.length && (
                <p className="text-[9px] sm:text-[10px] text-gray-500">
                  {properties.length} carregados
                </p>
              )}
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
              <p className="text-[10px] sm:text-xs text-gray-600 truncate">
                Disponíveis {!allDataLoaded && <span className="text-blue-600">⏳</span>}
              </p>
              <p className="text-base sm:text-lg font-bold text-green-600">{apiStats?.available ?? stats.available}</p>
            </div>
          </div>
        </button>

        {/* PROMPT 08: Pending Confirmation Filter */}
        <button
          onClick={() => setTypeFilter('pending_confirmation')}
          className={`bg-white rounded-lg shadow-sm p-2 sm:p-3 text-left transition-all hover:shadow-md ${
            typeFilter === 'pending_confirmation' ? 'ring-2 ring-amber-500' : ''
          }`}
        >
          <div className="flex items-center gap-1.5 sm:gap-2">
            <div className="w-7 h-7 sm:w-8 sm:h-8 bg-amber-100 rounded flex items-center justify-center flex-shrink-0">
              <AlertCircle className="w-3.5 h-3.5 sm:w-4 sm:h-4 text-amber-600" />
            </div>
            <div className="min-w-0">
              <p className="text-[10px] sm:text-xs text-gray-600 truncate">
                Pend. Confirm. {!allDataLoaded && <span className="text-blue-600">⏳</span>}
              </p>
              <p className="text-base sm:text-lg font-bold text-amber-600">{apiStats?.pending_confirmation ?? stats.pending_confirmation}</p>
            </div>
          </div>
        </button>
      </div>

      {/* Property Type Cards - Second Row */}
      <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-6 gap-3 sm:gap-4 mb-4 sm:mb-6">
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
              <p className="text-[10px] sm:text-xs text-gray-600 truncate">
                Aptos {!allDataLoaded && <span className="text-blue-600">⏳</span>}
              </p>
              <p className="text-base sm:text-lg font-bold text-orange-600">{apiStats?.apartments ?? stats.apartments}</p>
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
              <p className="text-[10px] sm:text-xs text-gray-600 truncate">
                Casas {!allDataLoaded && <span className="text-blue-600">⏳</span>}
              </p>
              <p className="text-base sm:text-lg font-bold text-purple-600">{apiStats?.houses ?? stats.houses}</p>
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
              <p className="text-[10px] sm:text-xs text-gray-600 truncate">
                Chácaras {!allDataLoaded && <span className="text-blue-600">⏳</span>}
              </p>
              <p className="text-base sm:text-lg font-bold text-teal-600">{apiStats?.chacaras ?? stats.chacaras}</p>
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
              <p className="text-[10px] sm:text-xs text-gray-600 truncate">
                Terrenos {!allDataLoaded && <span className="text-blue-600">⏳</span>}
              </p>
              <p className="text-base sm:text-lg font-bold text-amber-600">{apiStats?.terrenos ?? stats.terrenos}</p>
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
              <p className="text-[10px] sm:text-xs text-gray-600 truncate">
                Fazendas {!allDataLoaded && <span className="text-blue-600">⏳</span>}
              </p>
              <p className="text-base sm:text-lg font-bold text-emerald-600">{apiStats?.fazendas ?? stats.fazendas}</p>
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
              <p className="text-[10px] sm:text-xs text-gray-600 truncate">
                Sítios {!allDataLoaded && <span className="text-blue-600">⏳</span>}
              </p>
              <p className="text-base sm:text-lg font-bold text-lime-600">{apiStats?.sitios ?? stats.sitios}</p>
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

      {/* Loading State - Skeleton */}
      {loading && (
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3 sm:gap-4 md:gap-6">
          {[...Array(12)].map((_, i) => (
            <div key={i} className="bg-white rounded-lg shadow-sm overflow-hidden animate-pulse">
              <div className="h-48 bg-gray-200"></div>
              <div className="p-4">
                <div className="h-6 bg-gray-200 rounded mb-3"></div>
                <div className="h-4 bg-gray-200 rounded w-3/4 mb-3"></div>
                <div className="flex gap-4 mb-3">
                  <div className="h-4 bg-gray-200 rounded w-12"></div>
                  <div className="h-4 bg-gray-200 rounded w-12"></div>
                  <div className="h-4 bg-gray-200 rounded w-16"></div>
                </div>
                <div className="h-8 bg-gray-200 rounded"></div>
              </div>
            </div>
          ))}
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
                    placeholder="blur"
                    blurDataURL="data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iNDAwIiBoZWlnaHQ9IjMwMCIgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIj48cmVjdCB3aWR0aD0iNDAwIiBoZWlnaHQ9IjMwMCIgZmlsbD0iI2RkZCIvPjwvc3ZnPg=="
                  />
                ) : (
                  <Building2 className="w-16 h-16 text-white opacity-50" />
                )}
              </div>

              {/* Property Info */}
              <div className="p-4">
                <div className="flex items-start justify-between mb-2">
                  <h3 className="text-lg font-semibold text-gray-900 flex-1 line-clamp-2">
                    {property.title || property.description || `${getPropertyTypeLabel(property.property_type || '')} em ${property.neighborhood}`}
                  </h3>
                  <span className={`px-2 py-1 text-xs rounded-full ml-2 flex-shrink-0 ${
                    property.status?.toLowerCase() === 'available' ? 'bg-green-100 text-green-700' :
                    property.status?.toLowerCase() === 'pending_confirmation' ? 'bg-amber-100 text-amber-700' :
                    property.status?.toLowerCase() === 'unavailable' ? 'bg-red-100 text-red-700' :
                    property.status?.toLowerCase() === 'sold' ? 'bg-purple-100 text-purple-700' :
                    property.status?.toLowerCase() === 'rented' ? 'bg-indigo-100 text-indigo-700' :
                    'bg-blue-100 text-blue-600'
                  }`}>
                    {property.status?.toLowerCase() === 'available' ? 'Disponível' :
                     property.status?.toLowerCase() === 'pending_confirmation' ? 'Pend. Confirm.' :
                     property.status?.toLowerCase() === 'unavailable' ? 'Indisponível' :
                     property.status?.toLowerCase() === 'sold' ? 'Vendido' :
                     property.status?.toLowerCase() === 'rented' ? 'Alugado' :
                     property.status || 'Disponível'}
                  </span>
                </div>
                {property.reference && (
                  <p className="text-xs text-gray-600 mb-2">Código: {property.reference}</p>
                )}

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
          {(loadingMore || hasMore) && (
            <div className="text-center">
              <div className="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
              <p className="text-gray-600 mt-2">
                {loadingMore
                  ? 'Carregando do servidor...'
                  : `Mostrando ${displayCount} de ${filteredProperties.length} imóveis...`}
              </p>
            </div>
          )}
          {!hasMore && !loadingMore && filteredProperties.length > 0 && (
            <div className="text-center">
              <p className="text-gray-600">
                ✓ Mostrando todos os {filteredProperties.length} imóveis
                {properties.length < totalCount && ` (${properties.length} de ${totalCount} carregados do servidor)`}
              </p>
            </div>
          )}
        </div>
        </>
      )}
    </div>
  );
}
