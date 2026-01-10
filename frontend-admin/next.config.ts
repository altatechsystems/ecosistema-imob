import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  // Force rebuild with correct environment variables
  generateBuildId: async () => {
    return `build-${Date.now()}`;
  },

  // CRITICAL: Hardcoded URLs to bypass Vercel environment variable cache issues
  // These values override any cached builds that have incorrect placeholder values
  env: {
    NEXT_PUBLIC_API_URL: 'https://backend-api-333057134750.southamerica-east1.run.app/api/v1',
    NEXT_PUBLIC_ADMIN_API_URL: 'https://backend-api-333057134750.southamerica-east1.run.app/api/v1/admin',
  },

  images: {
    remotePatterns: [
      {
        protocol: 'https',
        hostname: 'storage.googleapis.com',
      },
      {
        protocol: 'https',
        hostname: 'firebasestorage.googleapis.com',
      },
    ],
    formats: ['image/webp'],
    deviceSizes: [640, 750, 828, 1080, 1200, 1920, 2048, 3840],
    imageSizes: [16, 32, 48, 64, 96, 128, 256, 384],
    minimumCacheTTL: 60,
    dangerouslyAllowSVG: false,
    contentDispositionType: 'attachment',
    contentSecurityPolicy: "default-src 'self'; script-src 'none'; sandbox;",
  },
};

export default nextConfig;
