// Lead Types - Matching backend models

export enum LeadStatus {
  NEW = 'new',
  CONTACTED = 'contacted',
  QUALIFIED = 'qualified',
  NEGOTIATING = 'negotiating',
  CONVERTED = 'converted',
  LOST = 'lost',
}

export enum LeadChannel {
  WHATSAPP = 'whatsapp',
  FORM = 'form',
  PHONE = 'phone',
  EMAIL = 'email',
  CHAT = 'chat',
  REFERRAL = 'referral',
}

export interface Lead {
  id?: string;
  tenant_id: string;
  property_id: string;
  broker_id?: string;

  // Contact info
  name?: string;
  email?: string;
  phone?: string;

  // Lead details
  message?: string;
  channel: LeadChannel;
  status?: LeadStatus;

  // PROMPT 07: Tracking (UTM parameters)
  utm_source?: string;
  utm_campaign?: string;
  utm_medium?: string;
  referrer?: string;

  // LGPD
  consent_given: boolean;
  consent_text?: string;
  consent_date?: Date | string;
  consent_ip?: string;
  consent_revoked?: boolean;
  revoked_at?: Date | string;
  is_anonymized?: boolean;
  anonymized_at?: Date | string;
  anonymization_reason?: string;

  // Timestamps
  created_at?: Date | string;
  updated_at?: Date | string;
}

export interface CreateLeadRequest {
  property_id: string;
  name: string;
  email?: string;
  phone: string;
  message?: string;
  channel: LeadChannel;
  consent_text: string;
}

export interface CreateLeadResponse {
  success: boolean;
  data: Lead;
}

export interface LeadListResponse {
  success: boolean;
  data: Lead[];
  count: number;
  has_more?: boolean;
}
