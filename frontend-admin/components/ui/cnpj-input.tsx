'use client';

import { useState } from 'react';

interface CnpjInputProps {
  label: string;
  value: string;
  onChange: (value: string) => void;
  helperText?: string;
  error?: string;
  required?: boolean;
  disabled?: boolean;
}

export function CnpjInput({
  label,
  value,
  onChange,
  helperText,
  error,
  required = false,
  disabled = false,
}: CnpjInputProps) {
  const [touched, setTouched] = useState(false);

  const formatCNPJ = (val: string) => {
    // Remove tudo que não é dígito
    const digitsOnly = val.replace(/\D/g, '');

    // Aplica máscara XX.XXX.XXX/XXXX-XX
    if (digitsOnly.length <= 2) {
      return digitsOnly;
    } else if (digitsOnly.length <= 5) {
      return `${digitsOnly.slice(0, 2)}.${digitsOnly.slice(2)}`;
    } else if (digitsOnly.length <= 8) {
      return `${digitsOnly.slice(0, 2)}.${digitsOnly.slice(2, 5)}.${digitsOnly.slice(5)}`;
    } else if (digitsOnly.length <= 12) {
      return `${digitsOnly.slice(0, 2)}.${digitsOnly.slice(2, 5)}.${digitsOnly.slice(5, 8)}/${digitsOnly.slice(8)}`;
    } else {
      return `${digitsOnly.slice(0, 2)}.${digitsOnly.slice(2, 5)}.${digitsOnly.slice(5, 8)}/${digitsOnly.slice(8, 12)}-${digitsOnly.slice(12, 14)}`;
    }
  };

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const formatted = formatCNPJ(e.target.value);
    onChange(formatted);
  };

  const validateCNPJ = (cnpj: string) => {
    if (!cnpj) return true; // Empty is valid if not required
    const digits = cnpj.replace(/\D/g, '');
    return digits.length === 14;
  };

  const isValid = !value || validateCNPJ(value);
  const showError = touched && !isValid;

  return (
    <div className="space-y-1">
      <label className="block text-sm font-medium text-gray-700">
        {label} {required && <span className="text-red-500">*</span>}
      </label>

      <input
        type="text"
        value={value}
        onChange={handleChange}
        onBlur={() => setTouched(true)}
        disabled={disabled}
        placeholder="00.000.000/0000-00"
        className={`w-full px-4 py-2 border rounded-lg focus:outline-none focus:ring-2 transition-colors disabled:bg-gray-100 disabled:cursor-not-allowed ${
          showError || error
            ? 'border-red-500 focus:ring-red-500'
            : 'border-gray-300 focus:ring-blue-500 focus:border-transparent'
        }`}
        maxLength={18}
      />

      {helperText && !showError && !error && (
        <p className="text-xs text-gray-500">{helperText}</p>
      )}

      {error && (
        <p className="text-xs text-red-600">{error}</p>
      )}

      {showError && !error && value && (
        <p className="text-xs text-orange-600">
          CNPJ deve ter 14 dígitos
        </p>
      )}
    </div>
  );
}
