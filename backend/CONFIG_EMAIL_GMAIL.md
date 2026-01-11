# Configuração de Email com Gmail SMTP

Este guia mostra como configurar o envio de emails usando sua conta Gmail.

## Passo a Passo

### 1. Gerar Senha de App no Gmail

O Gmail não permite usar sua senha normal para SMTP. Você precisa gerar uma **Senha de App**.

1. **Acesse**: https://myaccount.google.com/apppasswords
   - Você precisará fazer login na sua conta Google
   - Se solicitado, confirme sua identidade (2FA)

2. **Crie uma nova Senha de App**:
   - No campo "Selecione o app", escolha: **Mail**
   - No campo "Selecione o dispositivo", escolha: **Outro (nome personalizado)**
   - Digite um nome como: `Ecosistema Imob Backend`
   - Clique em **Gerar**

3. **Copie a senha gerada**:
   - O Google mostrará uma senha de 16 caracteres (algo como: `abcd efgh ijkl mnop`)
   - **IMPORTANTE**: Copie essa senha agora! Você não poderá vê-la novamente

### 2. Configurar o Backend

Edite o arquivo `backend/.env` e adicione:

```env
# Email Configuration (SMTP)
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=seu.email@gmail.com
SMTP_PASSWORD=abcdefghijklmnop  # Cole a senha de app (sem espaços!)

# Email Settings
EMAIL_FROM_NAME=Ecosistema Imob
FRONTEND_URL=http://localhost:3002
```

**⚠️ ATENÇÃO**:
- Use a **Senha de App** gerada (16 caracteres), NÃO sua senha normal do Gmail
- Remova os **espaços** da senha de app (cole como: `abcdefghijklmnop`)
- O `SMTP_USER` deve ser seu email completo do Gmail

### 3. Reiniciar o Backend

Após salvar o `.env`, reinicie o backend:

```bash
# Pare o servidor (Ctrl+C se estiver rodando)

# Inicie novamente
cd backend
go run cmd/server/main.go
```

Você verá no console:
```
✅ Email service enabled with SMTP: smtp.gmail.com:587
```

### 4. Testar o Envio

1. Acesse: http://localhost:3002/dashboard/equipe/novo
2. Preencha o formulário de convite
3. Clique em "Enviar Convite"
4. Verifique:
   - No console do backend: `✅ Invitation email sent successfully to email@example.com`
   - Na caixa de entrada do email convidado

## Problemas Comuns

### Erro: "Username and Password not accepted"
- ✅ Certifique-se de usar a **Senha de App**, não sua senha normal
- ✅ Remova espaços da senha de app
- ✅ Verifique se a autenticação em 2 fatores está ativada (necessária para Senhas de App)

### Erro: "Less secure app access"
- ✅ Use **Senhas de App** em vez de habilitar apps menos seguros
- Gmail não permite mais apps menos seguros desde maio de 2022

### Email não chega
- ✅ Verifique a pasta de Spam/Lixo Eletrônico
- ✅ Aguarde alguns segundos (pode haver atraso)
- ✅ Verifique se o email do destinatário está correto

### Email service disabled
Se você vir no console:
```
⚠️ Email service disabled - SMTP credentials not configured
```

Significa que uma ou mais variáveis estão faltando:
- `SMTP_HOST`
- `SMTP_USER`
- `SMTP_PASSWORD`

Verifique se todas estão configuradas no `.env`

## Alternativas ao Gmail

### Outlook/Hotmail
```env
SMTP_HOST=smtp.office365.com
SMTP_PORT=587
SMTP_USER=seu.email@outlook.com
SMTP_PASSWORD=sua_senha_normal  # Outlook aceita senha normal
```

### Provedor Profissional (Para Produção)

Para produção, considere usar:
- **SendGrid**: 100 emails/dia grátis
- **AWS SES**: 62.000 emails/mês grátis (com EC2)
- **Resend**: 100 emails/dia grátis, API moderna

## Modo de Desenvolvimento (Sem Email Real)

Se você não quiser configurar SMTP agora:
1. Deixe as variáveis `SMTP_*` comentadas ou vazias no `.env`
2. O backend **apenas logará** o conteúdo dos emails no console
3. Você verá o HTML do email e o link de convite no terminal

Exemplo de log:
```
⚠️ Email service disabled - would send to: user@example.com
📧 EMAIL SUBJECT: Convite para Altatech Systems - Ecosistema Imob
📧 EMAIL CONTENT (HTML):
<!DOCTYPE html>
...
```

Você pode copiar o link de convite do log e acessar diretamente no navegador!

## Segurança

- ✅ **NUNCA** commite o arquivo `.env` no Git (já está no `.gitignore`)
- ✅ Para produção, use variáveis de ambiente do servidor (Vercel, Railway, etc.)
- ✅ Revogue Senhas de App antigas que não usa mais
- ✅ Use um email específico para o sistema (como `noreply@seudominio.com`)

## Referências

- [Senhas de App do Google](https://support.google.com/accounts/answer/185833)
- [Gmail SMTP Settings](https://support.google.com/mail/answer/7126229)
