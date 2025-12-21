# 🔐 Config Directory

Esta pasta contém arquivos de configuração sensíveis.

## ⚠️ IMPORTANTE

**NUNCA commite arquivos com credenciais no Git!**

## Arquivo Necessário

Coloque aqui o arquivo baixado do Firebase Admin SDK:

```
backend/config/firebase-adminsdk.json
```

## Como Obter

1. Acesse: https://console.firebase.google.com/project/ecosistema-imob-dev/settings/serviceaccounts/adminsdk
2. Clique em "Gerar nova chave privada"
3. Salve o arquivo JSON baixado como `firebase-adminsdk.json` nesta pasta

## Estrutura Esperada

```json
{
  "type": "service_account",
  "project_id": "ecosistema-imob-dev",
  "private_key_id": "...",
  "private_key": "-----BEGIN PRIVATE KEY-----\n...\n-----END PRIVATE KEY-----\n",
  "client_email": "firebase-adminsdk-xxxxx@ecosistema-imob-dev.iam.gserviceaccount.com",
  ...
}
```

## Verificação

Para verificar se o arquivo está correto:

```bash
# Windows PowerShell
Test-Path backend\config\firebase-adminsdk.json

# Deve retornar: True
```

---

**Status:** ❌ Arquivo ainda não configurado
**Ação:** Siga o [FIREBASE_SETUP_GUIDE.md](../../FIREBASE_SETUP_GUIDE.md)
