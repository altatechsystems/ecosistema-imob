'use client';

import * as React from 'react';
import Link from 'next/link';
import Image from 'next/image';
import { useParams } from 'next/navigation';
import { Broker } from '@/types/broker';
import { Property } from '@/types/property';
import { PropertyCard } from '@/components/property/property-card';
import { Button } from '@/components/ui/button';
import { Card } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { api } from '@/lib/api';
import {
  Home,
  Mail,
  Phone,
  Award,
  Star,
  Building2,
  MessageCircle,
  User,
  MapPin,
} from 'lucide-react';
import { BreadcrumbStructuredData } from '@/components/seo/breadcrumb-structured-data';

export default function BrokerProfilePage() {
  const params = useParams();
  const brokerId = params?.id as string;

  const [broker, setBroker] = React.useState<Broker | null>(null);
  const [properties, setProperties] = React.useState<Property[]>([]);
  const [isLoading, setIsLoading] = React.useState(true);

  React.useEffect(() => {
    if (brokerId) {
      loadBrokerProfile();
    }
  }, [brokerId]);

  const loadBrokerProfile = async () => {
    try {
      setIsLoading(true);
      const brokerData = await api.getBrokerPublicProfile(brokerId);
      setBroker(brokerData);

      // Load broker's properties
      const propertiesData = await api.getBrokerProperties(brokerId, 50);
      setProperties(propertiesData);
    } catch (error) {
      console.error('Failed to load broker profile:', error);
    } finally {
      setIsLoading(false);
    }
  };

  const handleWhatsAppClick = () => {
    if (broker?.phone) {
      const phoneNumber = broker.phone.replace(/\D/g, '');
      const message = encodeURIComponent('Olá, vi seu perfil no site e gostaria de mais informações.');
      window.open(`https://wa.me/55${phoneNumber}?text=${message}`, '_blank');
    }
  };

  const handleEmailClick = () => {
    if (broker?.email) {
      window.open(`mailto:${broker.email}`, '_self');
    }
  };

  // Calculate statistics
  const totalProperties = properties.length;
  const soldProperties = 0; // TODO: Get from total_sales
  const averagePrice = properties.length > 0
    ? properties.reduce((sum, p) => sum + (p.sale_price || p.rental_price || 0), 0) / properties.length
    : 0;

  const formatPrice = (price: number) => {
    if (price >= 1000000) {
      return `R$${(price / 1000000).toFixed(1)}M`;
    } else if (price >= 1000) {
      return `R$${(price / 1000).toFixed(0)}K`;
    }
    return `R$${price.toLocaleString('pt-BR')}`;
  };

  const formatPriceRange = () => {
    if (properties.length === 0) return 'N/A';
    const prices = properties.map(p => p.sale_price || p.rental_price || 0).filter(p => p > 0);
    if (prices.length === 0) return 'N/A';
    const min = Math.min(...prices);
    const max = Math.max(...prices);
    return `${formatPrice(min)}-${formatPrice(max)}`;
  };

  if (isLoading) {
    return (
      <div className="min-h-screen bg-white">
        <header className="bg-white border-b">
          <div className="container mx-auto px-4 py-4">
            <Link href="/" className="flex items-center gap-2">
              <Home className="w-8 h-8 text-blue-600" />
              <span className="text-2xl font-bold text-gray-900">Imobiliária</span>
            </Link>
          </div>
        </header>
        <div className="container mx-auto px-4 py-8">
          <div className="animate-pulse space-y-4">
            <div className="h-32 bg-gray-200 rounded-full w-32" />
            <div className="h-8 bg-gray-200 rounded w-1/3" />
            <div className="h-6 bg-gray-200 rounded w-1/4" />
          </div>
        </div>
      </div>
    );
  }

  if (!broker) {
    return (
      <div className="min-h-screen bg-white">
        <header className="bg-white border-b">
          <div className="container mx-auto px-4 py-4">
            <Link href="/" className="flex items-center gap-2">
              <Home className="w-8 h-8 text-blue-600" />
              <span className="text-2xl font-bold text-gray-900">Imobiliária</span>
            </Link>
          </div>
        </header>
        <div className="container mx-auto px-4 py-16 text-center">
          <h1 className="text-2xl font-bold text-gray-900 mb-4">Corretor não encontrado</h1>
          <Link href="/imoveis">
            <Button variant="primary">Ver Todos os Imóveis</Button>
          </Link>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-white">
      {/* SEO Structured Data */}
      <BreadcrumbStructuredData
        items={[
          { name: 'Home', url: '/' },
          { name: 'Imóveis', url: '/imoveis' },
          { name: broker.name, url: `/corretores/${brokerId}` },
        ]}
      />

      {/* Header */}
      <header className="bg-white border-b sticky top-0 z-50">
        <div className="container mx-auto px-4 py-3">
          <div className="flex items-center justify-between">
            <Link href="/" className="flex items-center gap-2">
              <Home className="w-7 h-7 text-blue-600" />
              <span className="text-xl font-bold text-gray-900">Imobiliária</span>
            </Link>

            <nav className="hidden md:flex items-center gap-6">
              <Link href="/imoveis" className="text-gray-700 hover:text-blue-600 font-medium text-sm">
                Imóveis
              </Link>
            </nav>
          </div>
        </div>
      </header>

      {/* Breadcrumb */}
      <div className="border-b bg-gray-50">
        <div className="container mx-auto px-4 py-3">
          <div className="flex items-center gap-2 text-sm text-gray-600">
            <Link href="/" className="hover:text-blue-600">Home</Link>
            <span>›</span>
            <Link href="/imoveis" className="hover:text-blue-600">Imóveis</Link>
            <span>›</span>
            <span className="text-gray-900">{broker.name}</span>
          </div>
        </div>
      </div>

      <div className="container mx-auto px-4 py-6">
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
          {/* Main Content - Left Side */}
          <div className="lg:col-span-2">
            {/* Broker Header */}
            <div className="flex items-start gap-6 mb-8">
              {/* Broker Photo */}
              <div className="flex-shrink-0">
                {broker.photo_url ? (
                  <Image
                    src={broker.photo_url}
                    alt={broker.name}
                    width={120}
                    height={120}
                    className="rounded-full object-cover"
                  />
                ) : (
                  <div className="w-28 h-28 rounded-full bg-gray-200 flex items-center justify-center">
                    <User className="w-14 h-14 text-gray-400" />
                  </div>
                )}
              </div>

              {/* Broker Info */}
              <div className="flex-1">
                <div className="flex items-center gap-3 mb-2">
                  {broker.rating !== undefined && broker.rating > 0 && (
                    <Badge variant="default" size="sm">
                      <Star className="w-3 h-3 mr-1 text-yellow-500 fill-yellow-500" />
                      <span className="font-semibold">{broker.rating.toFixed(1)}</span>
                    </Badge>
                  )}
                  <span className="text-sm text-blue-600 font-semibold">Top Corretor</span>
                </div>

                <h1 className="text-3xl font-bold text-gray-900 mb-2">{broker.name}</h1>

                <div className="text-sm text-gray-600 space-y-1 mb-3">
                  {broker.company && (
                    <div className="flex items-center gap-2">
                      <Building2 className="w-4 h-4" />
                      <span>{broker.company}</span>
                    </div>
                  )}
                  <div className="flex items-center gap-2">
                    <Award className="w-4 h-4" />
                    <span>CRECI {broker.creci}</span>
                  </div>
                  {broker.experience !== undefined && broker.experience > 0 && (
                    <div className="text-gray-600">
                      {broker.experience} Anos de experiência
                    </div>
                  )}
                </div>

                {broker.rating !== undefined && broker.rating > 0 && (
                  <div className="flex items-center gap-2 mb-4">
                    <div className="flex items-center">
                      <Star className="w-4 h-4 text-yellow-500 fill-yellow-500" />
                      <span className="ml-1 font-semibold text-gray-900">{broker.rating.toFixed(1)}</span>
                    </div>
                    {broker.review_count !== undefined && broker.review_count > 0 && (
                      <span className="text-sm text-gray-600">{broker.review_count} avaliações</span>
                    )}
                  </div>
                )}
              </div>
            </div>

            {/* Statistics Cards - Horizontal */}
            <div className="mb-8 pb-6 border-b">
              <div className="grid grid-cols-3 gap-4">
                <div className="text-center p-4 bg-gray-50 rounded">
                  <p className="text-2xl font-bold text-gray-900 mb-1">{totalProperties}</p>
                  <p className="text-xs text-gray-600">Imóveis ativos</p>
                </div>
                <div className="text-center p-4 bg-gray-50 rounded">
                  <p className="text-2xl font-bold text-gray-900 mb-1">{formatPriceRange()}</p>
                  <p className="text-xs text-gray-600">Faixa de preço</p>
                </div>
                {averagePrice > 0 && (
                  <div className="text-center p-4 bg-gray-50 rounded">
                    <p className="text-2xl font-bold text-gray-900 mb-1">{formatPrice(averagePrice)}</p>
                    <p className="text-xs text-gray-600">Preço médio</p>
                  </div>
                )}
              </div>
            </div>

            {/* Get to Know Section */}
            {broker.bio && (
              <div className="mb-8 pb-8 border-b">
                <h2 className="text-xl font-bold text-gray-900 mb-3">Conheça {broker.name.split(' ')[0]}</h2>
                <p className="text-gray-700 leading-relaxed whitespace-pre-line">{broker.bio}</p>
              </div>
            )}

            {/* Properties Section */}
            <div>
              <h2 className="text-2xl font-bold text-gray-900 mb-1">
                Imóveis de {broker.name.split(' ')[0]} ({totalProperties})
              </h2>
              <p className="text-sm text-gray-600 mb-6">
                Confira todos os imóveis disponíveis com este corretor
              </p>

              {properties.length > 0 ? (
                <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                  {properties.map((property) => (
                    <PropertyCard key={property.id} property={property} variant="grid" />
                  ))}
                </div>
              ) : (
                <Card variant="bordered" padding="lg">
                  <p className="text-center text-gray-600">
                    Nenhum imóvel disponível no momento
                  </p>
                </Card>
              )}
            </div>
          </div>

          {/* Sidebar - Right Side */}
          <div className="lg:col-span-1">
            <div className="sticky top-24 space-y-6">
              {/* Specialties Card */}
              {(broker.specialties || broker.languages) && (
                <Card variant="elevated" padding="lg">
                  <h3 className="text-lg font-bold text-gray-900 mb-4">Especialidades</h3>

                  {broker.specialties && (
                    <div className="mb-4">
                      <div className="flex flex-wrap gap-2">
                        {broker.specialties.split(',').map((specialty, index) => (
                          <Badge key={index} variant="default" size="sm">
                            {specialty.trim()}
                          </Badge>
                        ))}
                      </div>
                    </div>
                  )}

                  {broker.languages && (
                    <div className="pt-4 border-t">
                      <p className="text-xs font-semibold text-gray-500 uppercase tracking-wide mb-2">
                        Idiomas
                      </p>
                      <p className="text-sm text-gray-700">{broker.languages}</p>
                    </div>
                  )}
                </Card>
              )}

              {/* Reviews Card - Mock */}
              <Card variant="elevated" padding="lg">
                <div className="flex items-center justify-between mb-4">
                  <h3 className="text-lg font-bold text-gray-900">Avaliações</h3>
                  {broker.rating !== undefined && broker.rating > 0 && (
                    <div className="flex items-center gap-1">
                      <Star className="w-5 h-5 text-yellow-500 fill-yellow-500" />
                      <span className="font-bold text-gray-900">{broker.rating.toFixed(1)}</span>
                    </div>
                  )}
                </div>

                {broker.review_count !== undefined && broker.review_count > 0 ? (
                  <>
                    <p className="text-sm text-gray-600 mb-4">
                      {broker.review_count} {broker.review_count === 1 ? 'avaliação' : 'avaliações'}
                    </p>

                    {/* Mock Review */}
                    <div className="border-t pt-4">
                      <div className="flex items-center gap-1 mb-2">
                        <Star className="w-4 h-4 text-yellow-500 fill-yellow-500" />
                        <Star className="w-4 h-4 text-yellow-500 fill-yellow-500" />
                        <Star className="w-4 h-4 text-yellow-500 fill-yellow-500" />
                        <Star className="w-4 h-4 text-yellow-500 fill-yellow-500" />
                        <Star className="w-4 h-4 text-yellow-500 fill-yellow-500" />
                        <span className="text-sm font-semibold ml-2">5.0</span>
                      </div>
                      <p className="text-xs text-gray-500 mb-2">15/12/2024 • Cliente Satisfeito</p>
                      <p className="text-sm font-semibold text-gray-900 mb-2">
                        Excelente atendimento e profissionalismo
                      </p>
                      <p className="text-sm text-gray-700 line-clamp-3">
                        {broker.name.split(' ')[0]} foi essencial para encontrar o imóvel perfeito.
                        Muito atencioso e conhecedor do mercado local...
                      </p>
                      <button className="text-sm text-blue-600 hover:text-blue-700 font-medium mt-2">
                        Ver mais
                      </button>
                    </div>

                    <div className="mt-4 pt-4 border-t flex gap-2">
                      <Button variant="ghost" size="sm" className="flex-1 text-sm">
                        Adicionar avaliação
                      </Button>
                      <Button variant="ghost" size="sm" className="flex-1 text-sm">
                        Ver todas
                      </Button>
                    </div>
                  </>
                ) : (
                  <div className="text-center py-6">
                    <Star className="w-12 h-12 text-gray-300 mx-auto mb-2" />
                    <p className="text-sm text-gray-600 mb-4">Ainda não há avaliações</p>
                    <Button variant="outline" size="sm">
                      Seja o primeiro a avaliar
                    </Button>
                  </div>
                )}
              </Card>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
