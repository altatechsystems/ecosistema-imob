# Security Audit - Secrets Management

**Data:** 28 de Dezembro de 2025
**Auditor:** Claude Code
**Status:** ⚠️ AÇÃO RECOMENDADA

## Resumo Executivo

Foi detectado que o arquivo `backend/config/firebase-adminsdk.json` (Firebase Service Account Key) esteve presente no histórico do Git, embora tenha sido removido do commit atual.

**Status Atual:**
- ✅ Arquivo removido do commit atual (commit a31ad87)
- ✅ Arquivo está no .gitignore
- ⚠️ Arquivo ainda existe no histórico do Git (commits 4550ea5 e anteriores)

## Detalhes da Descoberta

### Histórico do Arquivo

```bash
# Commits que modificaram o arquivo
4550ea5 - feat(storage): implement Firebase Storage integration for image uploads
a31ad87 - chore: update .gitignore to exclude sensitive files and binaries (REMOVIDO)
```

### Verificação Realizada

```bash
# 1. Arquivo NÃO está no commit atual
$ git ls-tree HEAD backend/config/firebase-adminsdk.json
(vazio - arquivo não encontrado)

# 2. Arquivo ESTÁ no .gitignore
$ git check-ignore backend/config/firebase-adminsdk.json
backend/config/firebase-adminsdk.json

# 3. Arquivo EXISTE no histórico
$ git log --all --full-history -- backend/config/firebase-adminsdk.json
a31ad87 chore: update .gitignore to exclude sensitive files and binaries
4550ea5 feat(storage): implement Firebase Storage integration for image uploads
```

## Nível de Risco

**MÉDIO** 🟡

**Justificativa:**
- Secret foi exposto no Git, mas repositório parece ser privado
- Secret foi removido do commit atual (boa prática)
- Secret ainda existe no histórico (pode ser acessado)
- Se o repositório for público ou se alguém clonou antes da remoção, o secret está comprometido

## Ações Recomendadas

### 1. IMEDIATO - Rotacionar Firebase Service Account (CRÍTICO)

Mesmo que o repositório seja privado, rotacione o service account por precaução:

**Passos:**
1. Acesse [Firebase Console](https://console.firebase.google.com)
2. Selecione projeto `ecosistema-imob-dev`
3. **Settings** > **Service accounts**
4. Na seção "Firebase Admin SDK":
   - Anote o email do service account atual
   - Clique em **Manage service account permissions** (abre Google Cloud Console)
5. No Google Cloud Console:
   - IAM & Admin > Service Accounts
   - Localize o service account: `firebase-adminsdk-xxxxx@ecosistema-imob-dev.iam.gserviceaccount.com`
   - Clique nos 3 pontos > **Manage keys**
   - **Delete** todas as keys antigas
6. Volte ao Firebase Console:
   - **Generate new private key**
   - Salve como `backend/config/firebase-adminsdk.json`
7. Teste a aplicação localmente
8. Atualize secrets no CI/CD (se houver)

**Por que rotacionar?**
- Qualquer pessoa que clonou o repositório antes do commit a31ad87 tem acesso ao secret
- Colaboradores removidos podem ter acesso
- Bots de varredura podem ter detectado o secret

### 2. OPCIONAL - Limpar Histórico do Git (AVANÇADO)

⚠️ **ATENÇÃO:** Esta operação reescreve o histórico do Git e requer coordenação com todo o time.

**Quando fazer:**
- Se o repositório é público ou foi público no passado
- Se há confirmação de que o secret foi comprometido
- Se há políticas de compliance que exigem remoção completa

**Como fazer:**

```bash
# 1. Backup do repositório
cp -r .git .git.backup

# 2. Instalar git-filter-repo (se não tiver)
pip3 install git-filter-repo

# 3. Remover arquivo de TODO o histórico
git filter-repo --path backend/config/firebase-adminsdk.json --invert-paths

# 4. Force push (COORDENE COM O TIME!)
git push origin --force --all
git push origin --force --tags

# 5. Time deve re-clonar repositório
# Todos os desenvolvedores:
cd ..
rm -rf ecosistema-imob
git clone <url>
```

**Alternativa com BFG Repo-Cleaner:**

```bash
# 1. Backup
cp -r .git .git.backup

# 2. Instalar BFG
brew install bfg  # macOS
# ou baixar de: https://rtyley.github.io/bfg-repo-cleaner/

# 3. Limpar arquivo
bfg --delete-files firebase-adminsdk.json

# 4. Cleanup
git reflog expire --expire=now --all
git gc --prune=now --aggressive

# 5. Force push
git push origin --force --all
```

**Impactos:**
- ❌ Todos os desenvolvedores precisam re-clonar o repositório
- ❌ Histórico de commits será reescrito (hashes diferentes)
- ❌ PRs abertos podem quebrar
- ❌ CI/CD pode precisar de ajustes
- ❌ Forks e clones antigos ficarão inconsistentes

### 3. RECOMENDADO - Implementar Detecção de Secrets

**Instalar git-secrets:**

```bash
# macOS
brew install git-secrets

# Linux
git clone https://github.com/awslabs/git-secrets
cd git-secrets
sudo make install

# Setup no projeto
cd /path/to/ecosistema-imob
git secrets --install
git secrets --register-aws

# Adicionar patterns personalizados
git secrets --add 'firebase-adminsdk.*\.json'
git secrets --add '"private_key":\s*".*BEGIN'
git secrets --add 'AIza[0-9A-Za-z-_]{35}'
git secrets --add '"type":\s*"service_account"'

# Escanear histórico
git secrets --scan-history
```

**Instalar gitleaks (alternativa melhor):**

```bash
# macOS
brew install gitleaks

# Linux
wget https://github.com/gitleaks/gitleaks/releases/download/v8.18.1/gitleaks_8.18.1_linux_x64.tar.gz
tar -xzf gitleaks_8.18.1_linux_x64.tar.gz
sudo mv gitleaks /usr/local/bin/

# Escanear repositório
gitleaks detect --source . --verbose

# Escanear histórico completo
gitleaks detect --source . --log-opts="--all" --verbose

# Adicionar pre-commit hook
echo '#!/bin/bash
gitleaks protect --staged --verbose
' > .git/hooks/pre-commit
chmod +x .git/hooks/pre-commit
```

### 4. RECOMENDADO - Adicionar GitHub/GitLab Secret Scanning

**GitHub:**
1. Repositório > Settings > Code security and analysis
2. Ativar:
   - ✅ Secret scanning
   - ✅ Push protection
   - ✅ Dependency graph
   - ✅ Dependabot alerts

**GitLab:**
1. Project > Settings > CI/CD
2. Secret Detection:
   - Adicionar job no `.gitlab-ci.yml`:

```yaml
include:
  - template: Security/Secret-Detection.gitlab-ci.yml
```

## Status dos Secrets Atuais

### ✅ Protegidos Corretamente

- `backend/.env` - No .gitignore ✅
- `frontend-public/.env.local` - No .gitignore ✅
- `frontend-admin/.env.local` - No .gitignore ✅
- `backend/config/firebase-adminsdk.json` - No .gitignore ✅

### ⚠️ Atenção Necessária

- `backend/config/firebase-adminsdk.json` - Existe no histórico do Git

## Checklist de Remediação

- [ ] **CRÍTICO:** Rotacionar Firebase Service Account
- [ ] Atualizar `backend/config/firebase-adminsdk.json` com nova key
- [ ] Testar aplicação com nova key
- [ ] Atualizar secrets no CI/CD (GitHub Actions, GitLab CI, etc.)
- [ ] **OPCIONAL:** Limpar histórico do Git com git-filter-repo
- [ ] **OPCIONAL:** Force push após limpeza (coordenar com time)
- [ ] **OPCIONAL:** Time re-clonar repositório
- [ ] Instalar gitleaks ou git-secrets
- [ ] Adicionar pre-commit hook para detecção
- [ ] Ativar Secret Scanning no GitHub/GitLab
- [ ] Documentar incident report (se aplicável)
- [ ] Revisar processo de onboarding para prevenir recorrência

## Monitoramento Contínuo

### Firebase Usage Logs

Monitore uso anormal do service account:

1. [Google Cloud Console](https://console.cloud.google.com)
2. Logging > Logs Explorer
3. Query:
```
resource.type="service_account"
protoPayload.authenticationInfo.principalEmail="firebase-adminsdk-xxxxx@ecosistema-imob-dev.iam.gserviceaccount.com"
```

**Alertas a configurar:**
- Logins de IPs desconhecidos
- Uso fora do horário comercial
- Operações de exclusão em massa
- Tentativas de acesso negadas

### GitHub Secret Scanning Alerts

Se ativou Secret Scanning, monitore:
1. Repositório > Security > Secret scanning alerts

## Lições Aprendidas

### O Que Funcionou Bem
- ✅ Arquivo foi removido do commit atual
- ✅ .gitignore está configurado corretamente
- ✅ Documentação de secrets management criada

### O Que Precisa Melhorar
- ❌ Secret foi commitado inicialmente (falta de pre-commit hook)
- ❌ Histórico não foi limpo após detecção
- ❌ Sem detecção automática de secrets (gitleaks, git-secrets)

### Prevenção Futura
1. **Pre-commit hooks**: Impedir commits de secrets
2. **Treinamento**: Educar time sobre secrets management
3. **Code review**: Revisar PRs para secrets antes de merge
4. **CI/CD scanning**: Escanear em cada build
5. **Secret rotation**: Política de rotação trimestral

## Referências

- [OWASP Secrets Management](https://cheatsheetseries.owasp.org/cheatsheets/Secrets_Management_Cheat_Sheet.html)
- [GitHub Secret Scanning](https://docs.github.com/en/code-security/secret-scanning/about-secret-scanning)
- [git-filter-repo](https://github.com/newren/git-filter-repo)
- [gitleaks](https://github.com/gitleaks/gitleaks)
- [BFG Repo-Cleaner](https://rtyley.github.io/bfg-repo-cleaner/)

## Suporte

Para dúvidas sobre esta auditoria:
- Consulte `SECRETS_MANAGEMENT.md`
- Revise `.gitignore`
- Execute `gitleaks detect` regularmente

---

**Próxima Auditoria:** 28 de Março de 2026 (90 dias)
