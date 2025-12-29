# Secrets Management

Este documento descreve como gerenciar secrets (credenciais, chaves de API, tokens) de forma segura no projeto.

## ⚠️ REGRA DE OURO

**NUNCA COMMITE SECRETS NO GIT**

Secrets incluem:
- Firebase Service Account keys (`.json`)
- Database passwords
- API keys
- JWT secrets
- OAuth client secrets
- Encryption keys
- Third-party credentials

## Estrutura de Secrets

### Backend (.env)

Arquivo: `backend/.env` (NUNCA commitado)

```env
# Server
PORT=8080
GIN_MODE=debug
ENVIRONMENT=development

# Firebase
FIREBASE_PROJECT_ID=ecosistema-imob-dev
FIREBASE_CREDENTIALS=./config/firebase-adminsdk.json

# Google Cloud Storage
GCS_BUCKET_NAME=ecosistema-imob-dev.firebasestorage.app

# CORS
ALLOWED_ORIGINS=http://localhost:3000,http://localhost:3001

# JWT (se implementado no futuro)
JWT_SECRET=your-super-secret-key-here-min-32-chars
JWT_EXPIRATION=24h

# Redis (se implementado no futuro)
REDIS_URL=localhost:6379
REDIS_PASSWORD=

# Webhooks (se implementado)
WEBHOOK_SECRET=your-webhook-secret
```

### Frontend Public (.env.local)

Arquivo: `frontend-public/.env.local` (NUNCA commitado)

```env
# API
NEXT_PUBLIC_API_URL=http://localhost:8080/api
NEXT_PUBLIC_TENANT_ID=your-tenant-id-here

# Firebase (Frontend - SÓ chaves públicas!)
NEXT_PUBLIC_FIREBASE_API_KEY=AIza...
NEXT_PUBLIC_FIREBASE_AUTH_DOMAIN=ecosistema-imob-dev.firebaseapp.com
NEXT_PUBLIC_FIREBASE_PROJECT_ID=ecosistema-imob-dev
NEXT_PUBLIC_FIREBASE_STORAGE_BUCKET=ecosistema-imob-dev.firebasestorage.app
NEXT_PUBLIC_FIREBASE_MESSAGING_SENDER_ID=123456789
NEXT_PUBLIC_FIREBASE_APP_ID=1:123456789:web:abc123

# WhatsApp
NEXT_PUBLIC_WHATSAPP=5511999999999
```

### Frontend Admin (.env.local)

Arquivo: `frontend-admin/.env.local` (NUNCA commitado)

```env
# Same as frontend-public
NEXT_PUBLIC_API_URL=http://localhost:8080/api
NEXT_PUBLIC_TENANT_ID=your-tenant-id-here

# Firebase (Frontend)
NEXT_PUBLIC_FIREBASE_API_KEY=AIza...
# ... (mesmos campos do public)
```

## Setup de Desenvolvimento

### 1. Criar .env Files

```bash
# Backend
cp backend/.env.example backend/.env
# Editar backend/.env com valores reais

# Frontend Public
cp frontend-public/.env.example frontend-public/.env.local
# Editar frontend-public/.env.local

# Frontend Admin
cp frontend-admin/.env.example frontend-admin/.env.local
# Editar frontend-admin/.env.local
```

### 2. Firebase Service Account

**Obter credencial:**
1. Acesse Firebase Console: https://console.firebase.google.com
2. Selecione projeto `ecosistema-imob-dev`
3. **Settings** > **Service accounts**
4. Clique em **Generate new private key**
5. Salve como `backend/config/firebase-adminsdk.json`

**IMPORTANTE:**
- Arquivo NÃO deve estar no git (.gitignore já configurado)
- Permissões: `chmod 600 backend/config/firebase-adminsdk.json` (Linux/Mac)
- Compartilhe via 1Password/Bitwarden com o time

### 3. Verificar .gitignore

```bash
# Verificar se secrets estão no .gitignore
git check-ignore backend/.env
git check-ignore backend/config/firebase-adminsdk.json
git check-ignore frontend-public/.env.local

# Deve retornar os caminhos (se estiverem ignorados corretamente)
```

### 4. Verificar Git History

```bash
# Verificar se algum secret foi commitado
git log --all --full-history -- "**/.env"
git log --all --full-history -- "**/firebase-adminsdk*.json"

# Se encontrar algo, use git-filter-repo para remover:
# https://github.com/newren/git-filter-repo
```

## Produção

### Opção 1: Environment Variables (Recomendado)

Não use arquivo `.json` em produção. Use variáveis de ambiente.

**Docker:**
```dockerfile
# Dockerfile
FROM golang:1.21-alpine

# ... build steps ...

# Secrets via environment variables
ENV FIREBASE_CREDENTIALS=/app/config/firebase-creds.json
```

**docker-compose.yml:**
```yaml
services:
  backend:
    environment:
      - FIREBASE_PROJECT_ID=${FIREBASE_PROJECT_ID}
      - FIREBASE_CREDENTIALS=/run/secrets/firebase_creds
    secrets:
      - firebase_creds

secrets:
  firebase_creds:
    file: ./secrets/firebase-adminsdk.json
```

**Kubernetes:**
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: firebase-credentials
type: Opaque
data:
  credentials.json: <base64-encoded-json>
---
apiVersion: apps/v1
kind: Deployment
spec:
  template:
    spec:
      containers:
      - name: backend
        env:
        - name: FIREBASE_CREDENTIALS
          value: /secrets/firebase/credentials.json
        volumeMounts:
        - name: firebase-creds
          mountPath: /secrets/firebase
          readOnly: true
      volumes:
      - name: firebase-creds
        secret:
          secretName: firebase-credentials
```

### Opção 2: Google Secret Manager (Melhor para GCP)

```go
// backend/internal/config/secrets.go
package config

import (
    "context"
    "fmt"

    secretmanager "cloud.google.com/go/secretmanager/apiv1"
    secretmanagerpb "cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
)

func GetSecret(ctx context.Context, projectID, secretID string) (string, error) {
    client, err := secretmanager.NewClient(ctx)
    if err != nil {
        return "", fmt.Errorf("failed to create secretmanager client: %v", err)
    }
    defer client.Close()

    name := fmt.Sprintf("projects/%s/secrets/%s/versions/latest", projectID, secretID)
    req := &secretmanagerpb.AccessSecretVersionRequest{
        Name: name,
    }

    result, err := client.AccessSecretVersion(ctx, req)
    if err != nil {
        return "", fmt.Errorf("failed to access secret version: %v", err)
    }

    return string(result.Payload.Data), nil
}

// Uso:
// firebaseJSON, err := GetSecret(ctx, "ecosistema-imob-dev", "firebase-adminsdk")
```

**Criar secret:**
```bash
# Criar secret no Secret Manager
gcloud secrets create firebase-adminsdk \
    --data-file=backend/config/firebase-adminsdk.json \
    --project=ecosistema-imob-dev

# Dar permissão ao service account da aplicação
gcloud secrets add-iam-policy-binding firebase-adminsdk \
    --member="serviceAccount:app@ecosistema-imob-dev.iam.gserviceaccount.com" \
    --role="roles/secretmanager.secretAccessor" \
    --project=ecosistema-imob-dev
```

### Opção 3: AWS Secrets Manager

```go
// Para deploy na AWS
import (
    "github.com/aws/aws-sdk-go/aws"
    "github.com/aws/aws-sdk-go/aws/session"
    "github.com/aws/aws-sdk-go/service/secretsmanager"
)

func GetAWSSecret(secretName string) (string, error) {
    sess := session.Must(session.NewSession())
    svc := secretsmanager.New(sess, aws.NewConfig().WithRegion("us-east-1"))

    input := &secretsmanager.GetSecretValueInput{
        SecretId: aws.String(secretName),
    }

    result, err := svc.GetSecretValue(input)
    if err != nil {
        return "", err
    }

    return *result.SecretString, nil
}
```

## CI/CD

### GitHub Actions

```yaml
# .github/workflows/deploy.yml
name: Deploy

on:
  push:
    branches: [main]

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Setup secrets
        env:
          FIREBASE_CREDENTIALS: ${{ secrets.FIREBASE_CREDENTIALS }}
        run: |
          mkdir -p backend/config
          echo "$FIREBASE_CREDENTIALS" > backend/config/firebase-adminsdk.json

      - name: Build
        env:
          FIREBASE_PROJECT_ID: ${{ secrets.FIREBASE_PROJECT_ID }}
        run: |
          cd backend
          go build -o bin/server cmd/server/main.go

      - name: Deploy
        run: |
          # Deploy steps...
```

**Configurar secrets no GitHub:**
1. Repositório > Settings > Secrets and variables > Actions
2. New repository secret:
   - Name: `FIREBASE_CREDENTIALS`
   - Value: (copie todo o conteúdo do `firebase-adminsdk.json`)

### GitLab CI

```yaml
# .gitlab-ci.yml
variables:
  FIREBASE_PROJECT_ID: ecosistema-imob-dev

deploy:
  stage: deploy
  script:
    - mkdir -p backend/config
    - echo "$FIREBASE_CREDENTIALS" > backend/config/firebase-adminsdk.json
    - cd backend && go build -o bin/server cmd/server/main.go
  only:
    - main
```

**Configurar secret no GitLab:**
1. Project > Settings > CI/CD > Variables
2. Add variable:
   - Key: `FIREBASE_CREDENTIALS`
   - Value: (conteúdo do JSON)
   - Type: File
   - Protected: ✅
   - Masked: ✅

## Rotação de Secrets

### Firebase Service Account

**Quando rotacionar:**
- A cada 90 dias (boa prática)
- Se houver suspeita de vazamento
- Ao remover membro do time que tinha acesso

**Como rotacionar:**
1. Gerar novo service account no Firebase Console
2. Atualizar `backend/config/firebase-adminsdk.json`
3. Testar aplicação
4. Atualizar secret no CI/CD
5. Fazer deploy
6. Deletar service account antigo no Firebase Console

### API Keys

```bash
# Firebase API Key (frontend)
# 1. Firebase Console > Project Settings > General
# 2. Web apps > Click app > Regenerate API Key
# 3. Atualizar .env.local e variáveis de CI/CD
```

## Detecção de Vazamentos

### git-secrets (Preventivo)

```bash
# Instalar
brew install git-secrets  # macOS
# ou
git clone https://github.com/awslabs/git-secrets

# Setup no projeto
cd ecosistema-imob
git secrets --install
git secrets --register-aws

# Adicionar patterns personalizados
git secrets --add 'firebase-adminsdk.*\.json'
git secrets --add 'private_key.*-----BEGIN'
git secrets --add 'AIza[0-9A-Za-z-_]{35}'

# Escanear histórico
git secrets --scan-history
```

### gitleaks (Scan completo)

```bash
# Instalar
brew install gitleaks

# Escanear repositório
gitleaks detect --source . --verbose

# Escanear antes de commit (pre-commit hook)
gitleaks protect --staged
```

### Configurar pre-commit hook

```bash
# .git/hooks/pre-commit
#!/bin/bash

# Verificar se há secrets
gitleaks protect --staged --verbose --redact

if [ $? -ne 0 ]; then
    echo "⚠️  SECRETS DETECTED! Commit blocked."
    echo "Remove secrets before committing."
    exit 1
fi

# Verificar .env files
if git diff --cached --name-only | grep -E "\.env$|\.env\..*"; then
    echo "⚠️  .env file in commit! Are you sure? (This is unusual)"
    echo "Press ENTER to continue or Ctrl+C to cancel"
    read
fi

exit 0
```

```bash
chmod +x .git/hooks/pre-commit
```

## O Que Fazer Se Commitou um Secret

### 1. NÃO APENAS DELETE O ARQUIVO

```bash
# ❌ ERRADO (secret ainda está no git history)
git rm backend/config/firebase-adminsdk.json
git commit -m "Remove credentials"
```

### 2. USE git-filter-repo

```bash
# ✅ CORRETO - Remove de TODO o histórico

# Instalar git-filter-repo
pip3 install git-filter-repo

# Remover arquivo de todo histórico
git filter-repo --path backend/config/firebase-adminsdk.json --invert-paths

# Force push (CUIDADO - coordene com o time!)
git push origin --force --all
```

### 3. ROTACIONE O SECRET IMEDIATAMENTE

Mesmo após remover do git, o secret foi exposto. Você DEVE:

1. **Firebase Service Account:**
   - Firebase Console > Service accounts
   - Delete o service account comprometido
   - Gere um novo
   - Atualize aplicação

2. **API Keys:**
   - Regenere a key no provedor
   - Atualize aplicação e CI/CD

3. **Monitore:**
   - Verifique logs de acesso
   - Procure atividade suspeita
   - Configure alertas de uso anormal

## Boas Práticas

### ✅ DO

- Use `.env` files (nunca commitados)
- Use secret managers em produção (Secret Manager, AWS Secrets, Vault)
- Rotacione secrets regularmente
- Use least privilege (mínimo privilégio necessário)
- Documente onde secrets são usados
- Use git-secrets ou gitleaks
- Separe secrets por ambiente (dev/staging/prod)
- Compartilhe secrets via 1Password/Bitwarden

### ❌ DON'T

- Nunca commite secrets no git
- Nunca compartilhe secrets via Slack/Email/Discord
- Nunca use mesmos secrets em dev e prod
- Nunca hardcode secrets no código
- Nunca deixe secrets em logs
- Nunca use secrets em URLs (query params)
- Nunca commite `.env` files

## Checklist de Segurança

Antes de fazer commit:

- [ ] `.env` files estão no .gitignore?
- [ ] `firebase-adminsdk.json` está no .gitignore?
- [ ] Rodou `gitleaks detect`?
- [ ] Verificou `git status` (nenhum secret staged)?
- [ ] Verificou `git diff --cached` antes de commit?
- [ ] Secrets de prod são diferentes de dev?
- [ ] API keys de frontend são públicas (apenas as do frontend)?
- [ ] Documentou novos secrets em `.env.example`?

## Ferramentas Recomendadas

- **git-secrets**: https://github.com/awslabs/git-secrets
- **gitleaks**: https://github.com/gitleaks/gitleaks
- **detect-secrets**: https://github.com/Yelp/detect-secrets
- **truffleHog**: https://github.com/trufflesecurity/truffleHog
- **1Password**: https://1password.com (compartilhamento seguro)
- **Google Secret Manager**: https://cloud.google.com/secret-manager
- **AWS Secrets Manager**: https://aws.amazon.com/secrets-manager/
- **HashiCorp Vault**: https://www.vaultproject.io/

## Suporte

Se você acidentalmente commitou um secret:

1. **IMEDIATO:** Rotacione o secret comprometido
2. Siga passos em "O Que Fazer Se Commitou um Secret"
3. Notifique o time
4. Documente o incidente
5. Revise processos para prevenir recorrência

## Referências

- [OWASP Secrets Management](https://cheatsheetseries.owasp.org/cheatsheets/Secrets_Management_Cheat_Sheet.html)
- [GitHub Security Best Practices](https://docs.github.com/en/code-security/getting-started/best-practices-for-preventing-data-leaks-in-your-organization)
- [Google Cloud Secret Management](https://cloud.google.com/secret-manager/docs/best-practices)
