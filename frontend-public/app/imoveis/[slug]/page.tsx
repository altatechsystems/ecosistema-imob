'use client';

import * as React from 'react';
import Link from 'next/link';
import Image from 'next/image';
import { useParams } from 'next/navigation';
import { Property } from '@/types/property';
import { PropertyCard } from '@/components/property/property-card';
import { ContactForm } from '@/components/forms/contact-form';
import { Button } from '@/components/ui/button';
import { Card } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { api } from '@/lib/api';
import {
  formatCurrency,
  formatArea,
  getPropertyTypeLabel,
  getTransactionTypeLabel,
  buildWhatsAppUrl,
  getPropertyFeatures,
  getPropertyAmenities,
} from '@/lib/utils';
import {
  Home,
  MapPin,
  Bed,
  Bath,
  Car,
  Maximize2,
  MessageCircle,
  Share2,
  Heart,
  ChevronLeft,
  ChevronRight,
} from 'lucide-react';
import { LeadChannel } from '@/types/lead';

export default function PropertyDetailsPage() {
  const params = useParams();
  const slug = params?.slug as string;

  const [property, setProperty] = React.useState<Property | null>(null);
  const [similarProperties, setSimilarProperties] = React.useState<Property[]>([]);
  const [isLoading, setIsLoading] = React.useState(true);
  const [isCreatingLead, setIsCreatingLead] = React.useState(false);
  const [currentImageIndex, setCurrentImageIndex] = React.useState(0);

  React.useEffect(() => {
    if (slug) {
      loadProperty();
    }
  }, [slug]);

  const loadProperty = async () => {
    try {
      setIsLoading(true);
      const data = await api.getPropertyBySlug(slug);
      setProperty(data);

      // Load similar properties
      if (data.id) {
        const similar = await api.getSimilarProperties(data.id, 4);
        setSimilarProperties(similar);
      }
    } catch (error) {
      console.error('Failed to load property:', error);
    } finally {
      setIsLoading(false);
    }
  };

  const handleWhatsAppClick = async () => {
    if (!property || isCreatingLead) return;

    try {
      setIsCreatingLead(true);

      // 1. Criar Lead PRIMEIRO (conforme AI_DEV_DIRECTIVE Section 8)
      const leadResponse = await api.createLead({
        property_id: property.id!,
        name: 'Lead via WhatsApp',
        phone: '',
        channel: LeadChannel.WHATSAPP,
        consent_given: true,
        consent_text: 'Consentimento implícito ao clicar em "Falar no WhatsApp"',
        consent_date: new Date().toISOString(),
      });

      // 2. Construir mensagem com Lead ID
      const reference = property.reference || property.title || `${getPropertyTypeLabel(property.property_type)} - ${property.city}`;
      const message = `Olá! Tenho interesse no imóvel ${reference}.\n\nLead ID: #${leadResponse.data.id}`;
      const whatsappUrl = buildWhatsAppUrl(process.env.NEXT_PUBLIC_WHATSAPP || '', message);

      // 3. Redirecionar para WhatsApp
      window.open(whatsappUrl, '_blank');
    } catch (error) {
      console.error('Erro ao criar lead:', error);
      // Fallback: abrir WhatsApp mesmo sem Lead (não ideal)
      const message = `Olá! Tenho interesse no imóvel: ${property.title || getPropertyTypeLabel(property.property_type)} - ${property.city}`;
      const whatsappUrl = buildWhatsAppUrl(process.env.NEXT_PUBLIC_WHATSAPP || '', message);
      window.open(whatsappUrl, '_blank');
    } finally {
      setIsCreatingLead(false);
    }
  };

  const handleShare = async () => {
    if (navigator.share) {
      await navigator.share({
        title: property?.title || 'Imóvel',
        text: `Confira este imóvel: ${property?.title}`,
        url: window.location.href,
      });
    }
  };

  const nextImage = () => {
    if (!property?.images || property.images.length === 0) return;
    setCurrentImageIndex((prev) => (prev + 1) % property.images!.length);
  };

  const prevImage = () => {
    if (!property?.images || property.images.length === 0) return;
    setCurrentImageIndex((prev) => (prev - 1 + property.images!.length) % property.images!.length);
  };

  if (isLoading) {
    return (
      <div className="min-h-screen bg-gray-50">
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
            <div className="h-96 bg-gray-200 rounded-lg" />
            <div className="h-12 bg-gray-200 rounded w-2/3" />
            <div className="h-8 bg-gray-200 rounded w-1/2" />
          </div>
        </div>
      </div>
    );
  }

  if (!property) {
    return (
      <div className="min-h-screen bg-gray-50">
        <header className="bg-white border-b">
          <div className="container mx-auto px-4 py-4">
            <Link href="/" className="flex items-center gap-2">
              <Home className="w-8 h-8 text-blue-600" />
              <span className="text-2xl font-bold text-gray-900">Imobiliária</span>
            </Link>
          </div>
        </header>
        <div className="container mx-auto px-4 py-16 text-center">
          <h1 className="text-2xl font-bold text-gray-900 mb-4">Imóvel não encontrado</h1>
          <Link href="/imoveis">
            <Button variant="primary">Ver Todos os Imóveis</Button>
          </Link>
        </div>
      </div>
    );
  }

  const price = property.sale_price || property.rental_price;
  const features = getPropertyFeatures(property);
  const amenities = getPropertyAmenities(property);

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Header */}
      <header className="bg-white border-b sticky top-0 z-50">
        <div className="container mx-auto px-4 py-4">
          <div className="flex items-center justify-between">
            <Link href="/" className="flex items-center gap-2">
              <Home className="w-8 h-8 text-blue-600" />
              <span className="text-2xl font-bold text-gray-900">Imobiliária</span>
            </Link>

            <nav className="hidden md:flex items-center gap-6">
              <Link href="/imoveis" className="text-gray-700 hover:text-blue-600 font-medium">
                Voltar para Imóveis
              </Link>
            </nav>
          </div>
        </div>
      </header>

      <div className="container mx-auto px-4 py-8">
        {/* Image Gallery */}
        <div className="relative w-full h-96 md:h-[500px] mb-8 rounded-lg overflow-hidden">
          {property.images && property.images.length > 0 ? (
            <>
              <Image
                src={property.images[currentImageIndex]?.large_url || property.cover_image_url || '/placeholder-property.jpg'}
                alt={property.title || 'Imóvel'}
                fill
                className="object-cover"
                priority
              />
              {property.images.length > 1 && (
                <>
                  <button
                    onClick={prevImage}
                    className="absolute left-4 top-1/2 -translate-y-1/2 bg-white/80 hover:bg-white p-2 rounded-full"
                  >
                    <ChevronLeft className="w-6 h-6" />
                  </button>
                  <button
                    onClick={nextImage}
                    className="absolute right-4 top-1/2 -translate-y-1/2 bg-white/80 hover:bg-white p-2 rounded-full"
                  >
                    <ChevronRight className="w-6 h-6" />
                  </button>
                  <div className="absolute bottom-4 left-1/2 -translate-x-1/2 bg-black/50 text-white px-3 py-1 rounded-full text-sm">
                    {currentImageIndex + 1} / {property.images.length}
                  </div>
                </>
              )}
            </>
          ) : (
            <Image
              src="/placeholder-property.jpg"
              alt="Sem imagem"
              fill
              className="object-cover"
            />
          )}

          {property.featured && (
            <div className="absolute top-4 left-4">
              <Badge variant="featured">Destaque</Badge>
            </div>
          )}
        </div>

        <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
          {/* Main Content */}
          <div className="lg:col-span-2 space-y-6">
            {/* Title and Price */}
            <Card variant="bordered" padding="lg">
              <div className="flex items-start justify-between mb-4">
                <div>
                  <Badge variant="info" size="md" className="mb-3">
                    {getTransactionTypeLabel(property.transaction_type)}
                  </Badge>
                  <h1 className="text-3xl font-bold text-gray-900 mb-2">
                    {property.title || `${getPropertyTypeLabel(property.property_type)} em ${property.neighborhood}`}
                  </h1>
                  <div className="flex items-center text-gray-600">
                    <MapPin className="w-5 h-5 mr-2" />
                    <span>
                      {property.street}, {property.number} - {property.neighborhood}, {property.city} - {property.state}
                    </span>
                  </div>
                </div>
                <div className="flex gap-2">
                  <Button variant="ghost" size="sm" onClick={handleShare}>
                    <Share2 className="w-5 h-5" />
                  </Button>
                  <Button variant="ghost" size="sm">
                    <Heart className="w-5 h-5" />
                  </Button>
                </div>
              </div>

              <div className="border-t pt-4">
                <p className="text-sm text-gray-600 mb-1">
                  {property.transaction_type === 'rent' ? 'Aluguel' : 'Venda'}
                </p>
                <p className="text-4xl font-bold text-blue-600">
                  {formatCurrency(price)}
                </p>
              </div>
            </Card>

            {/* Features */}
            <Card variant="bordered" padding="lg">
              <h2 className="text-xl font-bold text-gray-900 mb-4">Características</h2>
              <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
                {property.bedrooms && (
                  <div className="flex items-center gap-3">
                    <Bed className="w-6 h-6 text-blue-600" />
                    <div>
                      <p className="text-2xl font-bold text-gray-900">{property.bedrooms}</p>
                      <p className="text-sm text-gray-600">Quartos</p>
                    </div>
                  </div>
                )}
                {property.bathrooms && (
                  <div className="flex items-center gap-3">
                    <Bath className="w-6 h-6 text-blue-600" />
                    <div>
                      <p className="text-2xl font-bold text-gray-900">{property.bathrooms}</p>
                      <p className="text-sm text-gray-600">Banheiros</p>
                    </div>
                  </div>
                )}
                {property.parking_spaces && (
                  <div className="flex items-center gap-3">
                    <Car className="w-6 h-6 text-blue-600" />
                    <div>
                      <p className="text-2xl font-bold text-gray-900">{property.parking_spaces}</p>
                      <p className="text-sm text-gray-600">Vagas</p>
                    </div>
                  </div>
                )}
                {property.area_sqm && (
                  <div className="flex items-center gap-3">
                    <Maximize2 className="w-6 h-6 text-blue-600" />
                    <div>
                      <p className="text-2xl font-bold text-gray-900">{property.area_sqm}</p>
                      <p className="text-sm text-gray-600">m²</p>
                    </div>
                  </div>
                )}
              </div>
            </Card>

            {/* Description */}
            {property.description && (
              <Card variant="bordered" padding="lg">
                <h2 className="text-xl font-bold text-gray-900 mb-4">Descrição</h2>
                <p className="text-gray-700 leading-relaxed whitespace-pre-line">
                  {property.description}
                </p>
              </Card>
            )}

            {/* Amenities */}
            {amenities.length > 0 && (
              <Card variant="bordered" padding="lg">
                <h2 className="text-xl font-bold text-gray-900 mb-4">Comodidades</h2>
                <div className="grid grid-cols-2 md:grid-cols-3 gap-3">
                  {amenities.map((amenity, index) => (
                    <div key={index} className="flex items-center gap-2 text-gray-700">
                      <div className="w-2 h-2 bg-blue-600 rounded-full" />
                      <span>{amenity}</span>
                    </div>
                  ))}
                </div>
              </Card>
            )}
          </div>

          {/* Sidebar */}
          <div className="space-y-6">
            {/* Quick Contact */}
            <Card variant="elevated" padding="lg">
              <h3 className="text-lg font-bold text-gray-900 mb-4">Entre em Contato</h3>
              <div className="space-y-3">
                <Button
                  variant="secondary"
                  size="lg"
                  className="w-full"
                  leftIcon={<MessageCircle className="w-5 h-5" />}
                  onClick={handleWhatsAppClick}
                >
                  WhatsApp
                </Button>
              </div>
            </Card>

            {/* Contact Form */}
            <ContactForm
              propertyId={property.id}
              propertyTitle={property.title || `${getPropertyTypeLabel(property.property_type)} em ${property.city}`}
            />
          </div>
        </div>

        {/* Similar Properties */}
        {similarProperties.length > 0 && (
          <div className="mt-16">
            <h2 className="text-3xl font-bold text-gray-900 mb-8">Imóveis Similares</h2>
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
              {similarProperties.map((similarProperty) => (
                <PropertyCard key={similarProperty.id} property={similarProperty} variant="grid" />
              ))}
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
