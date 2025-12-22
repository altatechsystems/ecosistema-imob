'use client';

import * as React from 'react';
import { PropertyFilters as IPropertyFilters, TransactionType, PropertyType } from '@/types/property';
import { Input } from '@/components/ui/input';
import { Select, SelectOption } from '@/components/ui/select';
import { Button } from '@/components/ui/button';
import { Card } from '@/components/ui/card';
import { Search, X } from 'lucide-react';

export interface PropertyFiltersProps {
  filters: IPropertyFilters;
  onFiltersChange: (filters: IPropertyFilters) => void;
  onClearFilters: () => void;
  variant?: 'horizontal' | 'sidebar';
}

export function PropertyFiltersComponent({
  filters,
  onFiltersChange,
  onClearFilters,
  variant = 'horizontal',
}: PropertyFiltersProps) {
  const transactionTypeOptions: SelectOption[] = [
    { value: TransactionType.SALE, label: 'Venda' },
    { value: TransactionType.RENT, label: 'Aluguel' },
    { value: TransactionType.BOTH, label: 'Venda/Aluguel' },
  ];

  const propertyTypeOptions: SelectOption[] = [
    { value: PropertyType.APARTMENT, label: 'Apartamento' },
    { value: PropertyType.HOUSE, label: 'Casa' },
    { value: PropertyType.CONDO, label: 'Condomínio' },
    { value: PropertyType.COMMERCIAL, label: 'Comercial' },
    { value: PropertyType.LAND, label: 'Terreno' },
    { value: PropertyType.FARM, label: 'Chácara/Sítio' },
    { value: PropertyType.STUDIO, label: 'Studio' },
    { value: PropertyType.PENTHOUSE, label: 'Cobertura' },
    { value: PropertyType.TOWNHOUSE, label: 'Sobrado' },
  ];

  const bedroomsOptions: SelectOption[] = [
    { value: '1', label: '1 quarto' },
    { value: '2', label: '2 quartos' },
    { value: '3', label: '3 quartos' },
    { value: '4', label: '4+ quartos' },
  ];

  const parkingOptions: SelectOption[] = [
    { value: '1', label: '1 vaga' },
    { value: '2', label: '2 vagas' },
    { value: '3', label: '3+ vagas' },
  ];

  const handleFilterChange = (key: keyof IPropertyFilters, value: any) => {
    onFiltersChange({
      ...filters,
      [key]: value || undefined,
    });
  };

  const hasActiveFilters = Object.values(filters).some(v => v !== undefined && v !== null && v !== '');

  if (variant === 'sidebar') {
    return (
      <Card variant="bordered" padding="lg" className="sticky top-4">
        <div className="flex items-center justify-between mb-6">
          <h3 className="text-lg font-semibold text-gray-900">Filtros</h3>
          {hasActiveFilters && (
            <Button
              variant="ghost"
              size="sm"
              onClick={onClearFilters}
              leftIcon={<X className="w-4 h-4" />}
            >
              Limpar
            </Button>
          )}
        </div>

        <div className="space-y-4">
          <Select
            label="Tipo de transação"
            options={transactionTypeOptions}
            value={filters.transaction_type || ''}
            onChange={(e) => handleFilterChange('transaction_type', e.target.value as TransactionType)}
            placeholder="Todos"
          />

          <Select
            label="Tipo de imóvel"
            options={propertyTypeOptions}
            value={filters.property_type || ''}
            onChange={(e) => handleFilterChange('property_type', e.target.value as PropertyType)}
            placeholder="Todos"
          />

          <Input
            label="Cidade"
            placeholder="Ex: São Paulo"
            value={filters.city || ''}
            onChange={(e) => handleFilterChange('city', e.target.value)}
          />

          <Input
            label="Bairro"
            placeholder="Ex: Jardins"
            value={filters.neighborhood || ''}
            onChange={(e) => handleFilterChange('neighborhood', e.target.value)}
          />

          <div>
            <p className="text-sm font-medium text-gray-700 mb-2">Preço</p>
            <div className="grid grid-cols-2 gap-2">
              <Input
                type="number"
                placeholder="Mín"
                value={filters.min_price || ''}
                onChange={(e) => handleFilterChange('min_price', e.target.value ? Number(e.target.value) : undefined)}
              />
              <Input
                type="number"
                placeholder="Máx"
                value={filters.max_price || ''}
                onChange={(e) => handleFilterChange('max_price', e.target.value ? Number(e.target.value) : undefined)}
              />
            </div>
          </div>

          <Select
            label="Quartos"
            options={bedroomsOptions}
            value={filters.bedrooms?.toString() || ''}
            onChange={(e) => handleFilterChange('bedrooms', e.target.value ? Number(e.target.value) : undefined)}
            placeholder="Qualquer"
          />

          <Select
            label="Vagas de garagem"
            options={parkingOptions}
            value={filters.parking_spaces?.toString() || ''}
            onChange={(e) => handleFilterChange('parking_spaces', e.target.value ? Number(e.target.value) : undefined)}
            placeholder="Qualquer"
          />

          <div>
            <p className="text-sm font-medium text-gray-700 mb-2">Área (m²)</p>
            <div className="grid grid-cols-2 gap-2">
              <Input
                type="number"
                placeholder="Mín"
                value={filters.min_area || ''}
                onChange={(e) => handleFilterChange('min_area', e.target.value ? Number(e.target.value) : undefined)}
              />
              <Input
                type="number"
                placeholder="Máx"
                value={filters.max_area || ''}
                onChange={(e) => handleFilterChange('max_area', e.target.value ? Number(e.target.value) : undefined)}
              />
            </div>
          </div>
        </div>
      </Card>
    );
  }

  // Horizontal variant
  return (
    <Card variant="elevated" padding="lg">
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 xl:grid-cols-6 gap-4">
        <Select
          label="Transação"
          options={transactionTypeOptions}
          value={filters.transaction_type || ''}
          onChange={(e) => handleFilterChange('transaction_type', e.target.value as TransactionType)}
          placeholder="Todos"
        />

        <Select
          label="Tipo"
          options={propertyTypeOptions}
          value={filters.property_type || ''}
          onChange={(e) => handleFilterChange('property_type', e.target.value as PropertyType)}
          placeholder="Todos"
        />

        <Input
          label="Cidade"
          placeholder="Ex: São Paulo"
          value={filters.city || ''}
          onChange={(e) => handleFilterChange('city', e.target.value)}
        />

        <Input
          label="Bairro"
          placeholder="Ex: Jardins"
          value={filters.neighborhood || ''}
          onChange={(e) => handleFilterChange('neighborhood', e.target.value)}
        />

        <Select
          label="Quartos"
          options={bedroomsOptions}
          value={filters.bedrooms?.toString() || ''}
          onChange={(e) => handleFilterChange('bedrooms', e.target.value ? Number(e.target.value) : undefined)}
          placeholder="Qualquer"
        />

        <div className="flex items-end gap-2">
          <Button
            variant="primary"
            size="md"
            className="flex-1"
            leftIcon={<Search className="w-4 h-4" />}
          >
            Buscar
          </Button>
          {hasActiveFilters && (
            <Button
              variant="outline"
              size="md"
              onClick={onClearFilters}
            >
              <X className="w-4 h-4" />
            </Button>
          )}
        </div>
      </div>
    </Card>
  );
}
