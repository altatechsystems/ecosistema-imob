# Configuração de Produção

## URLs de Produção

### Frontend Admin
- **URL**: https://frontend-admin-blond.vercel.app
- **Platform**: Vercel
- **Branch**: `main` (auto-deploy habilitado)

### Backend API
- **URL**: https://backend-api-333057134750.southamerica-east1.run.app
- **Platform**: Google Cloud Run
- **Region**: southamerica-east1 (São Paulo)
- **Branch**: `main`

## Firebase

### Projeto de Produção
- **Project ID**: ecosistema-imob-prod
- **Database**: imob-prod
- **Auth Domain**: ecosistema-imob-prod.firebaseapp.com
- **Storage Bucket**: ecosistema-imob-prod.firebasestorage.app

## Variáveis de Ambiente

### Backend (Cloud Run)
```
ENVIRONMENT=production
FIREBASE_PROJECT_ID=ecosistema-imob-prod
FIRESTORE_DATABASE=imob-prod
GCS_BUCKET_NAME=ecosistema-imob-prod.firebasestorage.app
ALLOWED_ORIGINS=http://localhost:3000,http://localhost:3001,http://localhost:3002,https://frontend-admin-blond.vercel.app,https://*.vercel.app
GIN_MODE=release
```

### Frontend Admin (Vercel)
```
NEXT_PUBLIC_API_URL=https://backend-api-333057134750.southamerica-east1.run.app/api/v1
NEXT_PUBLIC_ADMIN_API_URL=https://backend-api-333057134750.southamerica-east1.run.app/api/v1/admin
NEXT_PUBLIC_FIREBASE_API_KEY=[configurado no Vercel]
NEXT_PUBLIC_FIREBASE_AUTH_DOMAIN=ecosistema-imob-prod.firebaseapp.com
NEXT_PUBLIC_FIREBASE_PROJECT_ID=ecosistema-imob-prod
NEXT_PUBLIC_FIREBASE_STORAGE_BUCKET=ecosistema-imob-prod.firebasestorage.app
NEXT_PUBLIC_FIREBASE_MESSAGING_SENDER_ID=[configurado no Vercel]
NEXT_PUBLIC_FIREBASE_APP_ID=[configurado no Vercel]
NEXT_PUBLIC_FIREBASE_DATABASE_URL=https://ecosistema-imob-prod.firebaseio.com
```

## Deployment

### Backend
1. Push para branch `main` trigger deploy automático no Cloud Run
2. Pode fazer deploy manual via Cloud Console
3. Verificar logs: `gcloud run logs read backend-api --region=southamerica-east1`

### Frontend Admin
1. Push para branch `main` trigger deploy automático no Vercel
2. Deploy manual via Vercel Dashboard
3. Logs disponíveis no Vercel Dashboard

## Multi-Tenancy

### Tenant Admin
- **ID**: tenant_master
- **Nome**: ALTATECH Systems
- **Tipo**: Admin (acessa todos os tenants)

### Estrutura
- Cada cliente é um tenant separado
- Tenant master pode alternar entre tenants via seletor no header
- Dados isolados por tenant no Firestore

## Segurança

### CORS
Backend configurado para aceitar requisições de:
- localhost (desenvolvimento)
- frontend-admin-blond.vercel.app
- *.vercel.app (preview deployments)

### Firebase Auth
- Custom claims incluem: tenant_id, role, user_id
- Token refresh automático
- Logout limpa todos os tokens

## Monitoramento

### Health Check
- Backend: https://backend-api-333057134750.southamerica-east1.run.app/health
- Retorna: `{"service":"ecosistema-imob-api","status":"healthy","success":true}`

### Logs
- **Backend**: Google Cloud Logging
- **Frontend**: Vercel Logs
- **Firebase**: Firebase Console

## Troubleshooting

### Cache Issues
Se o Vercel continuar servindo builds antigos:
1. Limpar cache: Vercel Settings → Data Cache → Clear
2. Force rebuild: commit vazio ou alterar `next.config.ts`
3. Build ID é único por timestamp (configurado em `next.config.ts`)

### CORS Errors
Verificar se domínio está em `ALLOWED_ORIGINS` no Cloud Run

### Auth Errors
Verificar se Firebase credentials estão corretas no Vercel

## Backups

### Firestore
- Backups automáticos configurados
- Retenção: 30 dias
- Localização: southamerica-east1

### Storage
- Redundância regional
- Lifecycle rules configuradas
