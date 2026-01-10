'use client';

import { useState } from 'react';

interface CpfInputProps {
  label: string;
  value: string;
  onChange: (value: string) => void;
  helperText?: string;
  error?: string;
  required?: boolean;
  disabled?: boolean;
}

export function CpfInput({
  label,
  value,
  onChange,
  helperText,
  error,
  required = false,
  disabled = false,
}: CpfInputProps) {
  const [touched, setTouched] = useState(false);

  const formatCPF = (val: string) => {
    // Remove tudo que não é dígito
    const digitsOnly = val.replace(/\D/g, '');

    // Aplica máscara XXX.XXX.XXX-XX
    if (digitsOnly.length <= 3) {
      return digitsOnly;
    } else if (digitsOnly.length <= 6) {
      return `${digitsOnly.slice(0, 3)}.${digitsOnly.slice(3)}`;
    } else if (digitsOnly.length <= 9) {
      return `${digitsOnly.slice(0, 3)}.${digitsOnly.slice(3, 6)}.${digitsOnly.slice(6)}`;
    } else {
      return `${digitsOnly.slice(0, 3)}.${digitsOnly.slice(3, 6)}.${digitsOnly.slice(6, 9)}-${digitsOnly.slice(9, 11)}`;
    }
  };

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const formatted = formatCPF(e.target.value);
    onChange(formatted);
  };

  const validateCPF = (cpf: string) => {
    if (!cpf) return true; // Empty is valid if not required
    const digits = cpf.replace(/\D/g, '');
    return digits.length === 11;
  };

  const isValid = !value || validateCPF(value);
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
        placeholder="000.000.000-00"
        className={`w-full px-4 py-2 border rounded-lg focus:outline-none focus:ring-2 transition-colors disabled:bg-gray-100 disabled:cursor-not-allowed ${
          showError || error
            ? 'border-red-500 focus:ring-red-500'
            : 'border-gray-300 focus:ring-blue-500 focus:border-transparent'
        }`}
        maxLength={14}
      />

      {helperText && !showError && !error && (
        <p className="text-xs text-gray-500">{helperText}</p>
      )}

      {error && (
        <p className="text-xs text-red-600">{error}</p>
      )}

      {showError && !error && value && (
        <p className="text-xs text-orange-600">
          CPF deve ter 11 dígitos
        </p>
      )}
    </div>
  );
}
