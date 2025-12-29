'use client';

import * as React from 'react';
import Link from 'next/link';
import Image from 'next/image';
import { Property } from '@/types/property';
import { Card, CardContent, CardFooter } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import {
  formatCurrency,
  formatArea,
  getPropertyTypeLabel,
  getTransactionTypeLabel,
  getStatusLabel,
  buildWhatsAppUrl,
} from '@/lib/utils';
import { Bed, Bath, Car, MapPin, Maximize2, MessageCircle } from 'lucide-react';
import { api } from '@/lib/api';
import { LeadChannel } from '@/types/lead';

export interface PropertyCardProps {
  property: Property;
  variant?: 'grid' | 'list';
  onWhatsAppClick?: (property: Property) => void;
}

export const PropertyCard = React.memo(function PropertyCard({ property, variant = 'grid', onWhatsAppClick }: PropertyCardProps) {
  const [isCreatingLead, setIsCreatingLead] = React.useState(false);
  const price = property.sale_price || property.rental_price;
  const priceLabel = property.transaction_type === 'rent' ? 'Aluguel' : 'Venda';

  const handleWhatsAppClick = async (e: React.MouseEvent) => {
    e.preventDefault();

    if (isCreatingLead) return;

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

      // 2. Callback opcional
      if (onWhatsAppClick) {
        onWhatsAppClick(property);
      }

      // 3. Construir mensagem com Lead ID
      const reference = property.reference || property.title || `${getPropertyTypeLabel(property.property_type)} - ${property.city}`;
      const message = `Olá! Tenho interesse no imóvel ${reference}.\n\nLead ID: #${leadResponse.data.id}`;
      const whatsappUrl = buildWhatsAppUrl(process.env.NEXT_PUBLIC_WHATSAPP || '', message);

      // 4. Redirecionar para WhatsApp
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

  const features = [
    { icon: Bed, value: property.bedrooms, label: 'quartos' },
    { icon: Bath, value: property.bathrooms, label: 'banheiros' },
    { icon: Car, value: property.parking_spaces, label: 'vagas' },
    { icon: Maximize2, value: property.area_sqm ? formatArea(property.area_sqm) : null, label: '' },
  ].filter(f => f.value);

  if (variant === 'list') {
    return (
      <Card variant="bordered" padding="none" className="hover:shadow-md transition-shadow">
        <Link href={`/imoveis/${property.slug || property.id}`}>
          <div className="flex flex-col sm:flex-row">
            {/* Image */}
            <div className="relative w-full sm:w-64 md:w-80 h-56 sm:h-64 md:h-auto flex-shrink-0">
              <Image
                src={property.cover_image_url || '/placeholder-property.jpg'}
                alt={property.title || `${getPropertyTypeLabel(property.property_type)} em ${property.city}`}
                fill
                sizes="(max-width: 640px) 100vw, (max-width: 768px) 256px, 320px"
                className="object-cover rounded-t-lg sm:rounded-l-lg sm:rounded-tr-none"
                loading="lazy"
                quality={75}
              />
              <div className="absolute top-2 left-2 sm:top-3 sm:left-3 flex gap-1.5 sm:gap-2">
                {property.featured && (
                  <Badge variant="featured" size="sm">Destaque</Badge>
                )}
                <Badge variant="info" size="sm">
                  {getTransactionTypeLabel(property.transaction_type)}
                </Badge>
              </div>
            </div>

            {/* Content */}
            <div className="flex-1 p-3 sm:p-4">
              <div className="flex flex-col h-full">
                <div className="flex-1">
                  <h3 className="text-base sm:text-lg md:text-xl font-semibold text-gray-900 mb-2 line-clamp-2">
                    {property.title || `${getPropertyTypeLabel(property.property_type)} em ${property.neighborhood}`}
                  </h3>

                  <div className="flex items-center text-gray-600 text-xs sm:text-sm mb-2 sm:mb-3">
                    <MapPin className="w-3 h-3 sm:w-4 sm:h-4 mr-1 flex-shrink-0" />
                    <span className="line-clamp-1">{property.neighborhood}, {property.city} - {property.state}</span>
                  </div>

                  {property.description && (
                    <p className="text-gray-600 text-xs sm:text-sm mb-3 sm:mb-4 line-clamp-2 hidden sm:block">
                      {property.description}
                    </p>
                  )}

                  <div className="flex flex-wrap gap-2 sm:gap-3 md:gap-4 mb-3 sm:mb-4">
                    {features.map((feature, index) => (
                      <div key={index} className="flex items-center text-gray-700 text-sm">
                        <feature.icon className="w-4 h-4 mr-1.5 text-gray-500" />
                        <span>{typeof feature.value === 'number' ? feature.value : feature.value} {feature.label}</span>
                      </div>
                    ))}
                  </div>
                </div>

                <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between pt-3 sm:pt-4 border-t gap-3">
                  <div>
                    <p className="text-xs sm:text-sm text-gray-600">{priceLabel}</p>
                    <p className="text-xl sm:text-2xl font-bold text-blue-600">
                      {formatCurrency(price)}
                    </p>
                  </div>

                  <Button
                    variant="secondary"
                    size="sm"
                    leftIcon={<MessageCircle className="w-4 h-4" />}
                    onClick={handleWhatsAppClick}
                    className="w-full sm:w-auto"
                  >
                    <span className="sm:hidden">Contato</span>
                    <span className="hidden sm:inline">WhatsApp</span>
                  </Button>
                </div>
              </div>
            </div>
          </div>
        </Link>
      </Card>
    );
  }

  // Grid variant (default)
  return (
    <Card variant="bordered" padding="none" className="hover:shadow-md transition-shadow">
      <Link href={`/imoveis/${property.slug || property.id}`}>
        {/* Image */}
        <div className="relative w-full h-48 sm:h-56">
          <Image
            src={property.cover_image_url || '/placeholder-property.jpg'}
            alt={property.title || `${getPropertyTypeLabel(property.property_type)} em ${property.city}`}
            fill
            sizes="(max-width: 640px) 100vw, (max-width: 1024px) 50vw, 33vw"
            className="object-cover rounded-t-lg"
            loading="lazy"
            quality={75}
          />
          <div className="absolute top-2 left-2 sm:top-3 sm:left-3 flex gap-1.5 sm:gap-2">
            {property.featured && (
              <Badge variant="featured" size="sm">Destaque</Badge>
            )}
            <Badge variant="info" size="sm">
              {getTransactionTypeLabel(property.transaction_type)}
            </Badge>
          </div>
        </div>

        <CardContent className="p-3 sm:p-4">
          {/* Price */}
          <div className="mb-2 sm:mb-3">
            <p className="text-xs sm:text-sm text-gray-600">{priceLabel}</p>
            <p className="text-xl sm:text-2xl font-bold text-blue-600">
              {formatCurrency(price)}
            </p>
          </div>

          {/* Title */}
          <h3 className="text-base sm:text-lg font-semibold text-gray-900 mb-2 line-clamp-2">
            {property.title || `${getPropertyTypeLabel(property.property_type)} em ${property.neighborhood}`}
          </h3>

          {/* Location */}
          <div className="flex items-center text-gray-600 text-xs sm:text-sm mb-2 sm:mb-3">
            <MapPin className="w-3 h-3 sm:w-4 sm:h-4 mr-1 flex-shrink-0" />
            <span className="line-clamp-1">{property.neighborhood}, {property.city} - {property.state}</span>
          </div>

          {/* Features */}
          <div className="flex flex-wrap gap-2 sm:gap-3 mb-3 sm:mb-4">
            {features.map((feature, index) => (
              <div key={index} className="flex items-center text-gray-700 text-xs sm:text-sm">
                <feature.icon className="w-3 h-3 sm:w-4 sm:h-4 mr-1 text-gray-500 flex-shrink-0" />
                <span>{typeof feature.value === 'number' ? feature.value : feature.value} {feature.label}</span>
              </div>
            ))}
          </div>
        </CardContent>

        <CardFooter className="p-3 sm:p-4 pt-0">
          <Button
            variant="secondary"
            size="sm"
            className="w-full text-xs sm:text-sm"
            leftIcon={<MessageCircle className="w-3 h-3 sm:w-4 sm:h-4" />}
            onClick={handleWhatsAppClick}
          >
            <span className="hidden sm:inline">Entrar em Contato</span>
            <span className="sm:hidden">Contato</span>
          </Button>
        </CardFooter>
      </Link>
    </Card>
  );
});
