'use client';

// Debug page to check environment variables
export default function DebugEnvPage() {
  const envVars = {
    NEXT_PUBLIC_API_URL: process.env.NEXT_PUBLIC_API_URL,
    NEXT_PUBLIC_ADMIN_API_URL: process.env.NEXT_PUBLIC_ADMIN_API_URL,
    NEXT_PUBLIC_FIREBASE_API_KEY: process.env.NEXT_PUBLIC_FIREBASE_API_KEY,
    NEXT_PUBLIC_FIREBASE_AUTH_DOMAIN: process.env.NEXT_PUBLIC_FIREBASE_AUTH_DOMAIN,
    NEXT_PUBLIC_FIREBASE_PROJECT_ID: process.env.NEXT_PUBLIC_FIREBASE_PROJECT_ID,
    NODE_ENV: process.env.NODE_ENV,
  };

  return (
    <div className="min-h-screen bg-gray-100 p-8">
      <div className="max-w-4xl mx-auto">
        <h1 className="text-3xl font-bold mb-8">Environment Variables Debug</h1>

        <div className="bg-white rounded-lg shadow p-6">
          <h2 className="text-xl font-semibold mb-4">Current Environment Variables:</h2>

          <div className="space-y-3">
            {Object.entries(envVars).map(([key, value]) => (
              <div key={key} className="border-b pb-3">
                <div className="font-mono text-sm text-gray-600">{key}</div>
                <div className="font-mono text-sm mt-1 break-all">
                  {value || <span className="text-red-500">undefined</span>}
                </div>
              </div>
            ))}
          </div>
        </div>

        <div className="mt-8 bg-yellow-50 border border-yellow-200 rounded-lg p-4">
          <h3 className="font-semibold text-yellow-800 mb-2">Note:</h3>
          <p className="text-sm text-yellow-700">
            If values show as "undefined" or contain placeholder text, the environment variables
            are not being injected during build time. Check Vercel environment variable configuration.
          </p>
        </div>
      </div>
    </div>
  );
}
