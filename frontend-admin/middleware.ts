import { NextResponse } from 'next/server';
import type { NextRequest } from 'next/server';

export function middleware(request: NextRequest) {
  const response = NextResponse.next();

  // Security Headers
  // X-Frame-Options: Prevent clickjacking attacks
  response.headers.set('X-Frame-Options', 'DENY');

  // X-Content-Type-Options: Prevent MIME type sniffing
  response.headers.set('X-Content-Type-Options', 'nosniff');

  // Referrer-Policy: Control referrer information
  response.headers.set('Referrer-Policy', 'strict-origin-when-cross-origin');

  // X-XSS-Protection: Enable browser XSS protection
  response.headers.set('X-XSS-Protection', '1; mode=block');

  // Strict-Transport-Security: Enforce HTTPS (only in production)
  if (process.env.NODE_ENV === 'production') {
    response.headers.set(
      'Strict-Transport-Security',
      'max-age=31536000; includeSubDomains; preload'
    );
  }

  // Content-Security-Policy: Prevent XSS and other injection attacks
  const apiUrl = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';
  const cspHeader = `
    default-src 'self';
    script-src 'self' 'unsafe-eval' 'unsafe-inline' https://www.gstatic.com;
    style-src 'self' 'unsafe-inline';
    img-src 'self' data: https: blob:;
    font-src 'self' data:;
    connect-src 'self' http://localhost:8080 http://localhost:8080/api/v1 https://*.googleapis.com https://*.firebaseio.com wss://*.firebaseio.com https://firebaseio.com https://identitytoolkit.googleapis.com ${apiUrl};
    frame-ancestors 'none';
    base-uri 'self';
    form-action 'self';
  `.replace(/\s{2,}/g, ' ').trim();

  response.headers.set('Content-Security-Policy', cspHeader);

  // Permissions-Policy: Control browser features
  response.headers.set(
    'Permissions-Policy',
    'camera=(), microphone=(), geolocation=()'
  );

  return response;
}

// Apply middleware to all routes except static files and API routes
export const config = {
  matcher: [
    /*
     * Match all request paths except for the ones starting with:
     * - api (API routes)
     * - _next/static (static files)
     * - _next/image (image optimization files)
     * - favicon.ico (favicon file)
     * - public folder files
     */
    '/((?!api|_next/static|_next/image|favicon.ico|.*\\.(?:svg|png|jpg|jpeg|gif|webp)$).*)',
  ],
};
