'use client';

import * as React from 'react';
import Link from 'next/link';
import Image from 'next/image';
import { Broker } from '@/types/broker';
import { Card } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import {
  User,
  Award,
  Star,
  Building2,
  MessageCircle,
} from 'lucide-react';

interface BrokerCardProps {
  broker: Broker;
  variant?: 'compact' | 'full';
  showContact?: boolean;
  onContactClick?: () => void;
}

export function BrokerCard({
  broker,
  variant = 'full',
  showContact = true,
  onContactClick
}: BrokerCardProps) {
  const handleWhatsAppClick = () => {
    if (onContactClick) {
      onContactClick();
    } else if (broker.phone) {
      // Format phone number for WhatsApp (remove formatting, keep only numbers)
      const phoneNumber = broker.phone.replace(/\D/g, '');
      const message = encodeURIComponent('Olá, vi seu perfil no site e gostaria de mais informações.');
      window.open(`https://wa.me/55${phoneNumber}?text=${message}`, '_blank');
    }
  };

  if (variant === 'compact') {
    return (
      <Card variant="bordered" padding="md">
        <div className="flex items-start gap-4">
          {/* Broker Photo */}
          <Link href={`/corretores/${broker.id}`} className="flex-shrink-0">
            {broker.photo_url ? (
              <Image
                src={broker.photo_url}
                alt={broker.name}
                width={80}
                height={80}
                className="rounded-full object-cover"
              />
            ) : (
              <div className="w-20 h-20 rounded-full bg-blue-100 flex items-center justify-center">
                <User className="w-10 h-10 text-blue-600" />
              </div>
            )}
          </Link>

          {/* Broker Info */}
          <div className="flex-1 min-w-0">
            <Link
              href={`/corretores/${broker.id}`}
              className="text-lg font-bold text-gray-900 hover:text-blue-600 transition-colors"
            >
              {broker.name}
            </Link>

            <div className="flex items-center gap-2 mt-1">
              <Badge variant="secondary" size="sm">
                <Award className="w-3 h-3 mr-1" />
                CRECI {broker.creci}
              </Badge>
            </div>

            {broker.company && (
              <div className="flex items-center gap-1 text-sm text-gray-600 mt-1">
                <Building2 className="w-4 h-4" />
                <span>{broker.company}</span>
              </div>
            )}

            {showContact && broker.phone && (
              <div className="mt-3">
                <Button
                  variant="primary"
                  size="sm"
                  onClick={handleWhatsAppClick}
                  leftIcon={<MessageCircle className="w-4 h-4" />}
                  className="w-full"
                >
                  WhatsApp
                </Button>
              </div>
            )}
          </div>
        </div>
      </Card>
    );
  }

  // Full variant (Zillow-inspired)
  return (
    <Card variant="elevated" padding="lg">
      <h3 className="text-lg font-bold text-gray-900 mb-4">Corretor Responsável</h3>

      <div className="flex items-start gap-4 mb-4">
        {/* Broker Photo */}
        <Link href={`/corretores/${broker.id}`} className="flex-shrink-0">
          {broker.photo_url ? (
            <Image
              src={broker.photo_url}
              alt={broker.name}
              width={100}
              height={100}
              className="rounded-full object-cover ring-2 ring-blue-100"
            />
          ) : (
            <div className="w-24 h-24 rounded-full bg-blue-100 flex items-center justify-center ring-2 ring-blue-200">
              <User className="w-12 h-12 text-blue-600" />
            </div>
          )}
        </Link>

        {/* Broker Basic Info */}
        <div className="flex-1 min-w-0">
          <Link
            href={`/corretores/${broker.id}`}
            className="text-xl font-bold text-gray-900 hover:text-blue-600 transition-colors block"
          >
            {broker.name}
          </Link>

          <div className="flex items-center gap-2 mt-2">
            <Badge variant="secondary">
              <Award className="w-3 h-3 mr-1" />
              CRECI {broker.creci}
            </Badge>
          </div>

          {broker.company && (
            <div className="flex items-center gap-1 text-sm text-gray-600 mt-2">
              <Building2 className="w-4 h-4" />
              <span className="font-medium">{broker.company}</span>
            </div>
          )}

          {/* Statistics */}
          {(broker.total_listings || broker.experience || broker.rating) && (
            <div className="grid grid-cols-2 sm:grid-cols-3 gap-3 mt-3 pt-3 border-t">
              {broker.total_listings !== undefined && broker.total_listings > 0 && (
                <div>
                  <p className="text-lg font-bold text-gray-900">{broker.total_listings}</p>
                  <p className="text-xs text-gray-600">Imóveis</p>
                </div>
              )}
              {broker.experience !== undefined && broker.experience > 0 && (
                <div>
                  <p className="text-lg font-bold text-gray-900">{broker.experience}</p>
                  <p className="text-xs text-gray-600">Anos exp.</p>
                </div>
              )}
              {broker.rating !== undefined && broker.rating > 0 && (
                <div className="flex items-center gap-1">
                  <Star className="w-4 h-4 text-yellow-500 fill-yellow-500" />
                  <p className="text-lg font-bold text-gray-900">{broker.rating.toFixed(1)}</p>
                  {broker.review_count !== undefined && broker.review_count > 0 && (
                    <p className="text-xs text-gray-600">({broker.review_count})</p>
                  )}
                </div>
              )}
            </div>
          )}
        </div>
      </div>

      {/* Profile Link */}
      <div className="mb-4">
        <Link
          href={`/corretores/${broker.id}`}
          className="text-sm text-blue-600 hover:text-blue-700 font-medium inline-block"
        >
          Ver perfil completo →
        </Link>
      </div>

      {/* Specialties */}
      {broker.specialties && (
        <div className="mb-4">
          <p className="text-xs font-semibold text-gray-500 uppercase tracking-wide mb-1">
            Especialidades
          </p>
          <p className="text-sm text-gray-700">{broker.specialties}</p>
        </div>
      )}

      {/* Languages */}
      {broker.languages && (
        <div className="mb-4">
          <p className="text-xs font-semibold text-gray-500 uppercase tracking-wide mb-1">
            Idiomas
          </p>
          <p className="text-sm text-gray-700">{broker.languages}</p>
        </div>
      )}

      {/* Contact Button */}
      {showContact && broker.phone && (
        <div className="pt-4 border-t">
          <Button
            variant="primary"
            size="lg"
            className="w-full"
            onClick={handleWhatsAppClick}
            leftIcon={<MessageCircle className="w-5 h-5" />}
          >
            Enviar WhatsApp
          </Button>
        </div>
      )}
    </Card>
  );
}
