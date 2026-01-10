/**
 * Safe logging utility that prevents sensitive data exposure in production
 *
 * Usage:
 * - logger.dev() - Only logs in development
 * - logger.prod() - Logs in both dev and production (use sparingly)
 * - logger.error() - Always logs errors (sanitized in production)
 */

const isDevelopment = process.env.NODE_ENV === 'development';

/**
 * Sanitize sensitive data from objects before logging in production
 */
function sanitizeData(data: any): any {
  if (!data || typeof data !== 'object') {
    return data;
  }

  const sensitiveKeys = [
    'password',
    'token',
    'firebase_token',
    'firebase_uid',
    'tenant_id',
    'broker_id',
    'user_id',
    'email',
    'phone',
    'document',
    'cpf',
    'cnpj',
    'creci',
    'api_key',
    'secret',
    'authorization',
  ];

  const sanitized = Array.isArray(data) ? [...data] : { ...data };

  for (const key in sanitized) {
    const lowerKey = key.toLowerCase();

    // Check if key contains sensitive information
    const isSensitive = sensitiveKeys.some(sensitiveKey =>
      lowerKey.includes(sensitiveKey)
    );

    if (isSensitive) {
      sanitized[key] = '[REDACTED]';
    } else if (typeof sanitized[key] === 'object' && sanitized[key] !== null) {
      // Recursively sanitize nested objects
      sanitized[key] = sanitizeData(sanitized[key]);
    }
  }

  return sanitized;
}

export const logger = {
  /**
   * Development-only logs
   * These will NOT appear in production builds
   */
  dev: (...args: any[]) => {
    if (isDevelopment) {
      console.log('[DEV]', ...args);
    }
  },

  /**
   * Production-safe logs
   * Sanitizes sensitive data in production
   */
  prod: (...args: any[]) => {
    if (isDevelopment) {
      console.log('[LOG]', ...args);
    } else {
      // Sanitize data in production
      const sanitized = args.map(arg => sanitizeData(arg));
      console.log(...sanitized);
    }
  },

  /**
   * Error logging
   * Sanitizes sensitive data in production but preserves error messages
   */
  error: (...args: any[]) => {
    if (isDevelopment) {
      console.error('[ERROR]', ...args);
    } else {
      // In production, log error message but sanitize data
      const sanitized = args.map(arg => {
        if (arg instanceof Error) {
          return {
            message: arg.message,
            name: arg.name,
            // Don't include stack trace in production
          };
        }
        return sanitizeData(arg);
      });
      console.error(...sanitized);
    }
  },

  /**
   * Warning logs
   * Sanitizes sensitive data in production
   */
  warn: (...args: any[]) => {
    if (isDevelopment) {
      console.warn('[WARN]', ...args);
    } else {
      const sanitized = args.map(arg => sanitizeData(arg));
      console.warn(...sanitized);
    }
  },

  /**
   * Info logs (always visible but sanitized in production)
   */
  info: (...args: any[]) => {
    if (isDevelopment) {
      console.info('[INFO]', ...args);
    } else {
      const sanitized = args.map(arg => sanitizeData(arg));
      console.info(...sanitized);
    }
  },
};

// Export individual functions for convenience
export const { dev, prod, error, warn, info } = logger;
