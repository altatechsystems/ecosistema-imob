'use client';

import { useState } from 'react';

interface CreciInputProps {
  type: 'F' | 'J';
  label: string;
  value: string;
  onChange: (value: string) => void;
  helperText?: string;
  error?: string;
  required?: boolean;
  disabled?: boolean;
}

export function CreciInput({
  type,
  label,
  value,
  onChange,
  helperText,
  error,
  required = false,
  disabled = false,
}: CreciInputProps) {
  const [touched, setTouched] = useState(false);

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    let val = e.target.value.toUpperCase();

    // Remove caracteres inválidos (mantém apenas números, letras e hífen)
    val = val.replace(/[^0-9A-Z\-]/g, '');

    // Auto-formata: adiciona hífen e tipo automaticamente
    const numbers = val.replace(/\D/g, '');
    if (numbers.length > 0) {
      if (numbers.length <= 5) {
        val = numbers;
      } else {
        val = `${numbers.slice(0, 5)}-${type}`;
      }
    }

    onChange(val);
  };

  const validateFormat = (val: string) => {
    if (!val) return true; // Empty is valid if not required
    const regex = type === 'F'
      ? /^\d{5}-F$/
      : /^\d{5}-J$/;
    return regex.test(val);
  };

  const isValid = !value || validateFormat(value);
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
        placeholder={type === 'F' ? '12345-F' : '67890-J'}
        className={`w-full px-4 py-2 border rounded-lg focus:outline-none focus:ring-2 transition-colors disabled:bg-gray-100 disabled:cursor-not-allowed ${
          showError || error
            ? 'border-red-500 focus:ring-red-500'
            : 'border-gray-300 focus:ring-blue-500 focus:border-transparent'
        }`}
        maxLength={7}
      />

      {helperText && !showError && !error && (
        <p className="text-xs text-gray-500">{helperText}</p>
      )}

      {error && (
        <p className="text-xs text-red-600">{error}</p>
      )}

      {showError && !error && value && (
        <p className="text-xs text-orange-600">
          Formato esperado: {type === 'F' ? '12345-F' : '67890-J'}
        </p>
      )}
    </div>
  );
}
