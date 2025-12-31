import axios, { AxiosInstance } from 'axios';
import { auth } from './firebase';
import {
  Property,
  PropertyFilters,
  PropertyListResponse,
  PropertyResponse,
  PaginationOptions,
} from '@/types/property';
import { Lead, LeadListResponse, CreateLeadRequest, CreateLeadResponse } from '@/types/lead';
import type {
  ConfirmPropertyStatusPriceRequest,
  GenerateOwnerConfirmationLinkRequest,
  GenerateOwnerConfirmationLinkResponse,
} from '@/types/property';

class AdminApiClient {
  private client: AxiosInstance;
  private tenantId: string;

  constructor() {
    this.tenantId = process.env.NEXT_PUBLIC_TENANT_ID || '';

    // Add tenant_id to base URL for admin routes
    const baseURL = process.env.NEXT_PUBLIC_ADMIN_API_URL;
    const adminBaseURL = baseURL?.endsWith('/admin')
      ? `${baseURL}/${this.tenantId}`
      : baseURL;

    this.client = axios.create({
      baseURL: adminBaseURL,
      headers: {
        'Content-Type': 'application/json',
      },
      timeout: 30000, // 30s for admin operations
    });

    // Request interceptor - add auth token
    this.client.interceptors.request.use(
      async (config) => {
        if (auth?.currentUser) {
          const token = await auth.currentUser.getIdToken();
          config.headers.Authorization = `Bearer ${token}`;
        }
        return config;
      },
      (error) => Promise.reject(error)
    );

    // Response interceptor
    this.client.interceptors.response.use(
      (response) => response,
      (error) => {
        console.error('API Error:', error.response?.data || error.message);

        // Handle auth errors
        if (error.response?.status === 401) {
          // Redirect to login or refresh token
          if (typeof window !== 'undefined') {
            window.location.href = '/login';
          }
        }

        return Promise.reject(error);
      }
    );
  }

  // ========== PROPERTIES ==========

  async getProperties(
    filters?: PropertyFilters,
    pagination?: PaginationOptions
  ): Promise<PropertyListResponse> {
    const params = new URLSearchParams();

    if (filters) {
      Object.entries(filters).forEach(([key, value]) => {
        if (value !== undefined && value !== null) {
          params.append(key, String(value));
        }
      });
    }

    if (pagination) {
      if (pagination.limit) params.append('limit', String(pagination.limit));
      if (pagination.order_by) params.append('order_by', pagination.order_by);
      if (pagination.order_direction) params.append('order_direction', pagination.order_direction);
      if (pagination.start_after) params.append('start_after', pagination.start_after);
    }

    const response = await this.client.get<PropertyListResponse>(
      `/properties?${params.toString()}`
    );
    return response.data;
  }

  async getProperty(id: string): Promise<Property> {
    const response = await this.client.get<PropertyResponse>(
      `/properties/${id}`
    );
    return response.data.data;
  }

  async createProperty(data: Partial<Property>): Promise<Property> {
    const response = await this.client.post<PropertyResponse>(
      '/properties',
      data
    );
    return response.data.data;
  }

  async updateProperty(id: string, data: Partial<Property>): Promise<Property> {
    const response = await this.client.put<PropertyResponse>(
      `/properties/${id}`,
      data
    );
    return response.data.data;
  }

  async deleteProperty(id: string): Promise<void> {
    await this.client.delete(`/properties/${id}`);
  }

  async updatePropertyStatus(id: string, status: string): Promise<Property> {
    const response = await this.client.patch<PropertyResponse>(
      `/properties/${id}/status`,
      { status }
    );
    return response.data.data;
  }

  async updatePropertyVisibility(id: string, visibility: string): Promise<Property> {
    const response = await this.client.patch<PropertyResponse>(
      `/properties/${id}/visibility`,
      { visibility }
    );
    return response.data.data;
  }

  // ========== LEADS ==========

  async getLeads(
    filters?: {
      property_id?: string;
      status?: string;
      channel?: string;
    },
    pagination?: PaginationOptions
  ): Promise<LeadListResponse> {
    const params = new URLSearchParams();

    if (filters) {
      Object.entries(filters).forEach(([key, value]) => {
        if (value) params.append(key, value);
      });
    }

    if (pagination) {
      if (pagination.limit) params.append('limit', String(pagination.limit));
      if (pagination.order_by) params.append('order_by', pagination.order_by);
      if (pagination.order_direction) params.append('order_direction', pagination.order_direction);
    }

    const response = await this.client.get<LeadListResponse>(
      `/leads?${params.toString()}`
    );
    return response.data;
  }

  async getLead(id: string): Promise<Lead> {
    const response = await this.client.get(`/leads/${id}`);
    return response.data.data;
  }

  async updateLeadStatus(id: string, status: string): Promise<Lead> {
    const response = await this.client.patch(`/leads/${id}/status`, { status });
    return response.data.data;
  }

  async assignLeadToBroker(leadId: string, brokerId: string): Promise<Lead> {
    const response = await this.client.post(`/leads/${leadId}/assign`, {
      broker_id: brokerId,
    });
    return response.data.data;
  }

  // ========== OWNERS ==========

  async getOwners(pagination?: PaginationOptions): Promise<any> {
    const params = new URLSearchParams();
    if (pagination?.limit) params.append('limit', String(pagination.limit));

    const response = await this.client.get(`/owners?${params.toString()}`);
    return response.data;
  }

  async getOwner(id: string): Promise<any> {
    const response = await this.client.get(`/owners/${id}`);
    return response.data.data;
  }

  async createOwner(data: any): Promise<any> {
    const response = await this.client.post('/owners', data);
    return response.data.data;
  }

  async updateOwner(id: string, data: any): Promise<any> {
    const response = await this.client.put(`/owners/${id}`, data);
    return response.data.data;
  }

  // ========== BROKERS ==========

  async getBrokers(pagination?: PaginationOptions): Promise<any> {
    const params = new URLSearchParams();
    if (pagination?.limit) params.append('limit', String(pagination.limit));

    const response = await this.client.get(`/brokers?${params.toString()}`);
    return response.data;
  }

  async getBroker(id: string): Promise<any> {
    const response = await this.client.get(`/brokers/${id}`);
    return response.data.data;
  }

  async createBroker(data: any): Promise<any> {
    const response = await this.client.post('/brokers', data);
    return response.data.data;
  }

  async updateBroker(id: string, data: any): Promise<any> {
    const response = await this.client.put(`/brokers/${id}`, data);
    return response.data.data;
  }

  // ========== IMPORTS ==========

  async uploadImport(files: { xml: File; xls?: File }): Promise<any> {
    const formData = new FormData();
    formData.append('xml', files.xml);
    if (files.xls) {
      formData.append('xls', files.xls);
    }

    const response = await this.client.post(
      `/tenants/${this.tenantId}/import`,
      formData,
      {
        headers: {
          'Content-Type': 'multipart/form-data',
        },
      }
    );
    return response.data;
  }

  async getImportBatch(batchId: string): Promise<any> {
    const response = await this.client.get(`/import-batches/${batchId}`);
    return response.data.data;
  }

  async getImportBatches(pagination?: PaginationOptions): Promise<any> {
    const params = new URLSearchParams();
    if (pagination?.limit) params.append('limit', String(pagination.limit));

    const response = await this.client.get(`/import-batches?${params.toString()}`);
    return response.data;
  }

  // ========== STORAGE (Photos) ==========

  async uploadPropertyPhoto(propertyId: string, file: File): Promise<any> {
    const formData = new FormData();
    formData.append('file', file);

    const response = await this.client.post(
      `/properties/${propertyId}/images`,
      formData,
      {
        headers: {
          'Content-Type': 'multipart/form-data',
        },
      }
    );
    return response.data;
  }

  async deletePropertyPhoto(propertyId: string, imageId: string): Promise<void> {
    await this.client.delete(`/properties/${propertyId}/images/${imageId}`);
  }

  async getPropertyPhotos(propertyId: string): Promise<any> {
    const response = await this.client.get(`/properties/${propertyId}/images`);
    return response.data.data;
  }

  // ========== DASHBOARD / METRICS ==========

  async getDashboardMetrics(): Promise<any> {
    const response = await this.client.get('/dashboard/metrics');
    return response.data.data;
  }

  async getPropertyStats(period?: string): Promise<any> {
    const params = period ? `?period=${period}` : '';
    const response = await this.client.get(`/dashboard/property-stats${params}`);
    return response.data.data;
  }

  async getLeadStats(period?: string): Promise<any> {
    const params = period ? `?period=${period}` : '';
    const response = await this.client.get(`/dashboard/lead-stats${params}`);
    return response.data.data;
  }

  // ========== PROMPT 08: PROPERTY STATUS CONFIRMATION ==========

  async confirmPropertyStatusPrice(
    propertyId: string,
    data: ConfirmPropertyStatusPriceRequest
  ): Promise<Property> {
    const response = await this.client.patch<PropertyResponse>(
      `/properties/${propertyId}/confirmations`,
      data
    );
    return response.data.data;
  }

  async generateOwnerConfirmationLink(
    propertyId: string,
    data: GenerateOwnerConfirmationLinkRequest
  ): Promise<GenerateOwnerConfirmationLinkResponse> {
    const response = await this.client.post<{ success: boolean; data: GenerateOwnerConfirmationLinkResponse }>(
      `/properties/${propertyId}/owner-confirmation-link`,
      data
    );
    return response.data.data;
  }

  // Scheduled Confirmations
  async scheduleMonthlyConfirmations(data: { scheduled_for?: string; dry_run?: boolean }): Promise<any> {
    const response = await this.client.post<{ success: boolean; data: any }>(
      '/scheduled-confirmations/schedule',
      data
    );
    return response.data.data;
  }

  async processPendingConfirmations(): Promise<void> {
    await this.client.post('/scheduled-confirmations/process');
  }

  async getScheduledConfirmations(status?: string): Promise<any[]> {
    const response = await this.client.get<{ success: boolean; data: any[] }>(
      `/scheduled-confirmations${status ? `?status=${status}` : ''}`
    );
    return response.data.data || [];
  }

  async getConfirmationMetrics(): Promise<any> {
    const response = await this.client.get<{ success: boolean; data: any }>(
      '/scheduled-confirmations/metrics'
    );
    return response.data.data;
  }
}

export const adminApi = new AdminApiClient();
export default adminApi;
