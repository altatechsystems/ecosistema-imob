# Guia de Deployment - Ecosistema Imob

Este guia descreve como fazer o deploy do sistema completo em produção.

## 📋 Pré-requisitos

- Conta Vercel (para frontend)
- Conta Railway/Render/Google Cloud Run (para backend Go)
- Projeto Firebase configurado
- Conta Gmail ou provedor SMTP para envio de emails

---

## 🚀 Deploy do Frontend (Next.js)

### Opção 1: Vercel (Recomendado)

1. **Conecte seu repositório ao Vercel**
   ```bash
   # No diretório do projeto
   cd frontend-admin
   vercel
   ```

2. **Configure as variáveis de ambiente no Vercel**

   Acesse: `https://vercel.com/[seu-projeto]/settings/environment-variables`

   Adicione as seguintes variáveis:

   ```env
   # Firebase Configuration
   NEXT_PUBLIC_FIREBASE_API_KEY=sua_api_key
   NEXT_PUBLIC_FIREBASE_AUTH_DOMAIN=seu-projeto.firebaseapp.com
   NEXT_PUBLIC_FIREBASE_PROJECT_ID=seu-projeto-id
   NEXT_PUBLIC_FIREBASE_STORAGE_BUCKET=seu-projeto.appspot.com
   NEXT_PUBLIC_FIREBASE_MESSAGING_SENDER_ID=seu_sender_id
   NEXT_PUBLIC_FIREBASE_APP_ID=seu_app_id
   NEXT_PUBLIC_FIREBASE_DATABASE_URL=https://seu-projeto.firebaseio.com

   # API Configuration
   NEXT_PUBLIC_API_URL=https://seu-backend.railway.app/api/v1
   NEXT_PUBLIC_ADMIN_API_URL=https://seu-backend.railway.app/api/v1/admin

   # Tenant Configuration
   NEXT_PUBLIC_TENANT_SLUG=sua-empresa

   # Feature Flags
   NEXT_PUBLIC_ENABLE_ANALYTICS=false
   NEXT_PUBLIC_ENABLE_CHAT=false
   ```

3. **Configure o domínio**
   - Adicione seu domínio customizado em Settings > Domains
   - Configure DNS apontando para Vercel

4. **Deploy**
   ```bash
   vercel --prod
   ```

### Opção 2: Build Manual

```bash
cd frontend-admin
npm run build
npm start
```

---

## 🔧 Deploy do Backend (Go)

### Opção 1: Railway (Recomendado)

1. **Instale o Railway CLI**
   ```bash
   npm install -g @railway/cli
   ```

2. **Faça login**
   ```bash
   railway login
   ```

3. **Crie um novo projeto**
   ```bash
   cd backend
   railway init
   ```

4. **Configure as variáveis de ambiente**

   Acesse: `https://railway.app/project/[seu-projeto]/variables`

   ```env
   # Firebase Configuration
   FIREBASE_PROJECT_ID=seu-projeto-id
   GOOGLE_APPLICATION_CREDENTIALS_JSON={"type":"service_account",...}

   # Server Configuration
   PORT=8080
   GIN_MODE=release
   ENVIRONMENT=production

   # CORS Configuration
   ALLOWED_ORIGINS=https://seu-dominio.com

   # Cloud Storage
   GCS_BUCKET_NAME=seu-bucket.appspot.com

   # Email Configuration (SMTP)
   SMTP_HOST=smtp.gmail.com
   SMTP_PORT=587
   SMTP_USER=seu.email@gmail.com
   SMTP_PASSWORD=sua_senha_app
   EMAIL_FROM_NAME=Sua Empresa
   FRONTEND_URL=https://seu-dominio.com

   # Logging
   LOG_LEVEL=info
   ```

5. **Deploy**
   ```bash
   railway up
   ```

### Opção 2: Google Cloud Run

1. **Configure o gcloud CLI**
   ```bash
   gcloud auth login
   gcloud config set project SEU_PROJECT_ID
   ```

2. **Crie um Dockerfile** (já existe no projeto)

3. **Build e deploy**
   ```bash
   cd backend
   gcloud run deploy ecosistema-imob-backend \
     --source . \
     --region us-central1 \
     --allow-unauthenticated \
     --set-env-vars "ENVIRONMENT=production,GIN_MODE=release"
   ```

4. **Configure variáveis de ambiente**
   ```bash
   gcloud run services update ecosistema-imob-backend \
     --update-env-vars FIREBASE_PROJECT_ID=seu-projeto-id
   ```

### Opção 3: Render

1. Acesse [render.com](https://render.com)
2. New > Web Service
3. Conecte seu repositório
4. Configure:
   - **Build Command**: `go build -o bin/server ./cmd/server`
   - **Start Command**: `./bin/server`
   - **Environment**: Go
5. Adicione variáveis de ambiente (mesmas do Railway)

---

## 📧 Configuração de Email em Produção

### Opção 1: Gmail (Desenvolvimento/Pequena Escala)

- Limite: ~500 emails/dia
- Configuração: Use senha de app conforme `backend/CONFIG_EMAIL_GMAIL.md`

### Opção 2: SendGrid (Recomendado para Produção)

1. **Crie uma conta**: https://sendgrid.com/
2. **Obtenha API Key**
3. **Atualize o código** (opcional):
   ```go
   // Em backend/internal/services/email_service.go
   // Substituir smtp.SendMail por SendGrid API
   ```
4. **Configure variáveis**:
   ```env
   SENDGRID_API_KEY=SG.xxx
   ```

### Opção 3: AWS SES

1. **Configure AWS SES** no console AWS
2. **Verifique domínio/email**
3. **Configure**:
   ```env
   AWS_REGION=us-east-1
   AWS_ACCESS_KEY_ID=sua_key
   AWS_SECRET_ACCESS_KEY=sua_secret
   ```

### Opção 4: Resend (Moderna e Simples)

1. **Crie conta**: https://resend.com/
2. **Obtenha API Key**
3. **Configure**:
   ```env
   RESEND_API_KEY=re_xxx
   ```

---

## 🔐 Configuração do Firebase

### 1. Service Account (Backend)

1. Acesse: [Console Firebase](https://console.firebase.google.com/)
2. Settings > Service Accounts
3. Clique em "Generate new private key"
4. Salve o arquivo JSON

**Para Railway/Render**:
```bash
# Converta o JSON para string
cat firebase-adminsdk.json | jq -c . | pbcopy
# Cole como variável GOOGLE_APPLICATION_CREDENTIALS_JSON
```

**Para Google Cloud Run**:
```bash
# Upload do arquivo
gcloud secrets create firebase-key --data-file=firebase-adminsdk.json
```

### 2. Firebase Authentication

1. **Ative métodos de login**:
   - Authentication > Sign-in method
   - Habilite: Email/Password

2. **Configure domínios autorizados**:
   - Adicione seu domínio de produção

### 3. Firestore Security Rules

```javascript
rules_version = '2';
service cloud.firestore {
  match /databases/{database}/documents {
    // Tenants collection
    match /tenants/{tenantId} {
      // Allow read if user belongs to tenant
      allow read: if request.auth != null &&
                     request.auth.token.tenant_id == tenantId;

      // Users subcollection
      match /users/{userId} {
        allow read: if request.auth != null &&
                       request.auth.token.tenant_id == tenantId;
        allow write: if request.auth != null &&
                        request.auth.token.tenant_id == tenantId &&
                        (request.auth.token.role == 'admin' ||
                         request.auth.token.role == 'manager');
      }

      // User invitations - only accessible via backend
      match /user_invitations/{invitationId} {
        allow read: if true; // Public read for verification
        allow write: if false; // Only backend can write
      }

      // Other collections...
      match /{document=**} {
        allow read, write: if request.auth != null &&
                              request.auth.token.tenant_id == tenantId;
      }
    }
  }
}
```

---

## 🗄️ Banco de Dados (Firestore)

### Índices Necessários

Crie os seguintes índices compostos:

1. **Properties**:
   - `tenant_id` (Ascending) + `status` (Ascending) + `created_at` (Descending)
   - `tenant_id` (Ascending) + `type` (Ascending) + `created_at` (Descending)

2. **Leads**:
   - `tenant_id` (Ascending) + `status` (Ascending) + `created_at` (Descending)

3. **User Invitations**:
   - `tenant_id` (Ascending) + `status` (Ascending) + `created_at` (Descending)
   - `token` (Ascending) - para lookup rápido

**Como criar**:
- Firestore Console > Indexes
- Ou deixe o Firestore criar automaticamente quando houver erro

---

## 🔍 Checklist de Deploy

### Backend

- [ ] Variáveis de ambiente configuradas
- [ ] Service Account do Firebase configurado
- [ ] CORS configurado com domínio correto
- [ ] SMTP/Email provider configurado
- [ ] Logs habilitados
- [ ] Health check endpoint funcionando (`/health`)
- [ ] SSL/HTTPS habilitado

### Frontend

- [ ] Variáveis de ambiente configuradas
- [ ] Firebase config atualizado para produção
- [ ] API URL apontando para backend em produção
- [ ] Build otimizado (`npm run build`)
- [ ] Domínio configurado
- [ ] SSL/HTTPS habilitado
- [ ] Analytics configurado (se habilitado)

### Firebase

- [ ] Domínios autorizados configurados
- [ ] Security Rules atualizadas
- [ ] Índices criados
- [ ] Authentication habilitado
- [ ] Service Account gerado

### Email

- [ ] Provedor SMTP configurado
- [ ] Template de email testado
- [ ] Sender email verificado
- [ ] Limites de envio verificados

---

## 📊 Monitoramento

### Logs

**Backend (Railway)**:
```bash
railway logs
```

**Backend (Google Cloud Run)**:
```bash
gcloud run services logs read ecosistema-imob-backend
```

**Frontend (Vercel)**:
- Acesse: Dashboard > Deployments > [deployment] > Logs

### Métricas

**Firebase**:
- Authentication > Usage
- Firestore > Usage

**Vercel**:
- Analytics (se habilitado)

**Railway/Cloud Run**:
- Métricas de CPU/Memory no dashboard

---

## 🆘 Troubleshooting

### Email não está sendo enviado

1. Verifique logs do backend:
   ```bash
   railway logs | grep "Email"
   ```

2. Confirme variáveis SMTP:
   ```bash
   railway variables
   ```

3. Teste SMTP manualmente:
   ```go
   // Use o código em backend/internal/services/email_service.go
   ```

### 401 Unauthorized

1. Verifique se Firebase Service Account está configurado
2. Confirme que token está sendo enviado no header
3. Verifique CORS configuration

### Firestore Permission Denied

1. Atualize Security Rules
2. Verifique se custom claims estão sendo setados
3. Confira tenant_id no token

---

## 🔄 Atualizações Futuras

### Deploy de Novas Versões

**Frontend**:
```bash
cd frontend-admin
git pull
vercel --prod
```

**Backend**:
```bash
cd backend
git pull
railway up
# ou
gcloud run deploy
```

### Rollback

**Vercel**:
- Dashboard > Deployments > [previous deployment] > Promote to Production

**Railway**:
```bash
railway rollback
```

**Cloud Run**:
```bash
gcloud run services update-traffic ecosistema-imob-backend --to-revisions=PREVIOUS_REVISION=100
```

---

## 📚 Recursos Adicionais

- [Documentação Next.js Deploy](https://nextjs.org/docs/deployment)
- [Documentação Vercel](https://vercel.com/docs)
- [Documentação Railway](https://docs.railway.app/)
- [Documentação Google Cloud Run](https://cloud.google.com/run/docs)
- [Documentação Firebase](https://firebase.google.com/docs)
- [SendGrid Docs](https://docs.sendgrid.com/)

---

## 🎯 Próximos Passos

Após o deployment:

1. **Teste o fluxo completo**:
   - Cadastro de usuário
   - Login
   - Envio de convite
   - Aceite de convite
   - CRUD de imóveis

2. **Configure monitoramento**:
   - Sentry para error tracking
   - Google Analytics
   - Uptime monitoring

3. **Backup**:
   - Configure backup automático do Firestore
   - Backup de Service Accounts

4. **Segurança**:
   - Revise Security Rules
   - Configure rate limiting
   - Habilite 2FA para admins

---

**Desenvolvido por**: Altatech Systems
**Data**: Janeiro 2025
**Versão**: 1.0.0
