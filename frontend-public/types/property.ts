// Property Types - Matching backend models

import { Broker } from './broker';

export enum TransactionType {
  SALE = 'sale',
  RENT = 'rent',
  BOTH = 'both',
}

export enum PropertyType {
  APARTMENT = 'apartment',
  HOUSE = 'house',
  CONDO = 'condo',
  COMMERCIAL = 'commercial',
  LAND = 'land',
  FARM = 'farm',
  STUDIO = 'studio',
  PENTHOUSE = 'penthouse',
  TOWNHOUSE = 'townhouse',
}

export enum PropertyStatus {
  AVAILABLE = 'available',
  SOLD = 'sold',
  RENTED = 'rented',
  RESERVED = 'reserved',
  UNAVAILABLE = 'unavailable',
}

export enum PropertyVisibility {
  PUBLIC = 'public',
  PRIVATE = 'private',
  EXCLUSIVE = 'exclusive',
}

export enum PropertyDataCompleteness {
  COMPLETE = 'complete',
  PARTIAL = 'partial',
  INCOMPLETE = 'incomplete',
}

export interface Property {
  id: string;
  tenant_id: string;
  owner_id: string;
  transaction_type: TransactionType;
  property_type: PropertyType;
  status: PropertyStatus;
  visibility?: PropertyVisibility;

  // Price
  sale_price?: number;
  rental_price?: number;

  // Location
  street: string;
  number: string;
  complement?: string;
  neighborhood: string;
  city: string;
  state: string;
  postal_code: string;
  country?: string;
  latitude?: number;
  longitude?: number;

  // Characteristics
  bedrooms?: number;
  bathrooms?: number;
  suites?: number;
  parking_spaces?: number;
  area_sqm?: number;
  total_area_sqm?: number;

  // Features
  furnished?: boolean;
  pet_friendly?: boolean;
  has_pool?: boolean;
  has_gym?: boolean;
  has_elevator?: boolean;
  has_security?: boolean;
  has_garden?: boolean;
  has_balcony?: boolean;

  // Details
  title?: string;
  description?: string;
  year_built?: number;
  floor?: number;
  total_floors?: number;

  // SEO and metadata
  slug?: string;
  featured?: boolean;
  views_count?: number;
  leads_count?: number;
  data_completeness?: PropertyDataCompleteness;

  // Images
  images?: PropertyImage[];
  cover_image_url?: string;

  // Broker information (captador/corretor responsável)
  captador_id?: string;
  captador_name?: string;
  captador?: Broker; // Populated broker data

  // Timestamps
  created_at?: Date | string;
  updated_at?: Date | string;
}

export interface PropertyImage {
  id: string;
  url: string; // Original URL from backend Photo model
  thumb_url: string; // 400x300 WebP
  medium_url: string; // 800x600 WebP
  large_url: string; // 1600x1200 WebP
  order: number;
  is_cover: boolean;
}

export interface PropertyFilters {
  transaction_type?: TransactionType;
  property_type?: PropertyType;
  status?: PropertyStatus;
  visibility?: PropertyVisibility;
  city?: string;
  neighborhood?: string;
  min_price?: number;
  max_price?: number;
  bedrooms?: number;
  bathrooms?: number;
  parking_spaces?: number;
  min_area?: number;
  max_area?: number;
  furnished?: boolean;
  pet_friendly?: boolean;
  featured?: boolean;
}

export interface PaginationOptions {
  limit?: number;
  order_by?: string;
  order_direction?: 'asc' | 'desc';
  start_after?: string;
}

export interface PropertyListResponse {
  success: boolean;
  data: Property[];
  count: number;
  has_more?: boolean;
}

export interface PropertyResponse {
  success: boolean;
  data: Property;
}
