package services

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"net/smtp"
	"os"
	"strconv"
	"time"
)

// EmailService handles sending emails
type EmailService struct {
	fromEmail string
	fromName  string
	baseURL   string // Frontend URL for invitation links
	smtpHost  string
	smtpPort  int
	smtpUser  string
	smtpPass  string
	enabled   bool // Whether email sending is enabled
}

// NewEmailService creates a new email service
func NewEmailService() *EmailService {
	smtpPortStr := getEnv("SMTP_PORT", "587")
	smtpPort, err := strconv.Atoi(smtpPortStr)
	if err != nil {
		smtpPort = 587
	}

	smtpHost := getEnv("SMTP_HOST", "")
	smtpUser := getEnv("SMTP_USER", "")
	smtpPass := getEnv("SMTP_PASSWORD", "")

	// Email is enabled if SMTP credentials are provided
	enabled := smtpHost != "" && smtpUser != "" && smtpPass != ""

	if enabled {
		log.Printf("✅ Email service enabled with SMTP: %s:%d", smtpHost, smtpPort)
	} else {
		log.Printf("⚠️  Email service disabled - SMTP credentials not configured")
		log.Printf("   Set SMTP_HOST, SMTP_USER, and SMTP_PASSWORD to enable email sending")
	}

	return &EmailService{
		fromEmail: getEnv("EMAIL_FROM", getEnv("SMTP_USER", "noreply@ecosistema-imob.com")),
		fromName:  getEnv("EMAIL_FROM_NAME", "Ecosistema Imob"),
		baseURL:   getEnv("FRONTEND_URL", "http://localhost:3002"),
		smtpHost:  smtpHost,
		smtpPort:  smtpPort,
		smtpUser:  smtpUser,
		smtpPass:  smtpPass,
		enabled:   enabled,
	}
}

// getEnv gets environment variable with fallback
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

// InvitationEmailData contains data for the invitation email template
type InvitationEmailData struct {
	InviteeName   string
	TenantName    string
	InviterName   string
	Role          string
	AcceptURL     string
	ExpiresAt     time.Time
	ExpiresInDays int
}

// SendInvitation sends an invitation email to a user
func (s *EmailService) SendInvitation(email, name, token, tenantID, tenantName, inviterName, role string, expiresAt time.Time) error {
	// Build accept URL
	acceptURL := fmt.Sprintf("%s/auth/accept-invitation?token=%s", s.baseURL, token)

	// Calculate days until expiration
	expiresInDays := int(time.Until(expiresAt).Hours() / 24)

	// Prepare template data
	data := InvitationEmailData{
		InviteeName:   name,
		TenantName:    tenantName,
		InviterName:   inviterName,
		Role:          s.translateRole(role),
		AcceptURL:     acceptURL,
		ExpiresAt:     expiresAt,
		ExpiresInDays: expiresInDays,
	}

	// Generate email HTML
	htmlBody, err := s.renderInvitationTemplate(data)
	if err != nil {
		log.Printf("❌ Error rendering email template: %v", err)
		return err
	}

	// Generate plain text version
	textBody := s.generatePlainTextInvitation(data)

	log.Printf("📧 Sending invitation email to %s (%s)", email, name)
	log.Printf("   Accept URL: %s", acceptURL)

	subject := fmt.Sprintf("Convite para %s - Ecosistema Imob", tenantName)

	// If email is enabled, send via SMTP
	if s.enabled {
		err = s.sendEmail(email, name, subject, htmlBody, textBody)
		if err != nil {
			log.Printf("❌ Error sending email via SMTP: %v", err)
			return err
		}
		log.Printf("✅ Invitation email sent successfully to %s", email)
		return nil
	}

	// If email is disabled, just log the content
	log.Printf("⚠️  Email service disabled - would send to: %s", email)
	log.Printf("📧 EMAIL SUBJECT: %s", subject)
	log.Printf("📧 EMAIL CONTENT (HTML):\n%s", htmlBody)
	log.Printf("📧 EMAIL CONTENT (TEXT):\n%s", textBody)

	return nil
}

// translateRole translates role codes to friendly names
func (s *EmailService) translateRole(role string) string {
	roleNames := map[string]string{
		"admin":        "Administrador",
		"manager":      "Gerente",
		"broker":       "Corretor",
		"broker_admin": "Corretor Administrador",
	}

	if name, exists := roleNames[role]; exists {
		return name
	}
	return role
}

// renderInvitationTemplate renders the HTML email template
func (s *EmailService) renderInvitationTemplate(data InvitationEmailData) (string, error) {
	tmpl := `
<!DOCTYPE html>
<html lang="pt-BR">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Convite para Ecosistema Imob</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
            line-height: 1.6;
            color: #333;
            background-color: #f4f4f4;
            margin: 0;
            padding: 0;
        }
        .container {
            max-width: 600px;
            margin: 40px auto;
            background-color: #ffffff;
            border-radius: 8px;
            overflow: hidden;
            box-shadow: 0 2px 8px rgba(0,0,0,0.1);
        }
        .header {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            padding: 40px 30px;
            text-align: center;
        }
        .header h1 {
            margin: 0;
            font-size: 28px;
            font-weight: 600;
        }
        .content {
            padding: 40px 30px;
        }
        .greeting {
            font-size: 18px;
            margin-bottom: 20px;
        }
        .message {
            font-size: 16px;
            margin-bottom: 30px;
            color: #555;
        }
        .details {
            background-color: #f8f9fa;
            border-left: 4px solid #667eea;
            padding: 20px;
            margin: 20px 0;
        }
        .details p {
            margin: 8px 0;
        }
        .details strong {
            color: #333;
        }
        .cta-button {
            display: inline-block;
            padding: 14px 32px;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            text-decoration: none;
            border-radius: 6px;
            font-weight: 600;
            font-size: 16px;
            text-align: center;
            margin: 20px 0;
        }
        .cta-button:hover {
            opacity: 0.9;
        }
        .expiration {
            font-size: 14px;
            color: #888;
            margin-top: 20px;
            padding: 15px;
            background-color: #fff3cd;
            border-left: 4px solid #ffc107;
            border-radius: 4px;
        }
        .footer {
            background-color: #f8f9fa;
            padding: 30px;
            text-align: center;
            font-size: 14px;
            color: #888;
            border-top: 1px solid #e9ecef;
        }
        .link-fallback {
            font-size: 12px;
            color: #888;
            margin-top: 20px;
            word-break: break-all;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🏢 Ecosistema Imob</h1>
        </div>

        <div class="content">
            <p class="greeting">Olá, <strong>{{.InviteeName}}</strong>!</p>

            <p class="message">
                Você foi convidado por <strong>{{.InviterName}}</strong> para fazer parte da equipe
                <strong>{{.TenantName}}</strong> no Ecosistema Imob.
            </p>

            <div class="details">
                <p><strong>📋 Detalhes do Convite:</strong></p>
                <p>🏢 Empresa: {{.TenantName}}</p>
                <p>👤 Função: {{.Role}}</p>
                <p>👥 Convidado por: {{.InviterName}}</p>
            </div>

            <p>Para aceitar o convite e criar sua conta, clique no botão abaixo:</p>

            <div style="text-align: center;">
                <a href="{{.AcceptURL}}" class="cta-button">Aceitar Convite</a>
            </div>

            <div class="expiration">
                <strong>⏰ Atenção:</strong> Este convite expira em {{.ExpiresInDays}} dias
                ({{.ExpiresAt.Format "02/01/2006 às 15:04"}}).
            </div>

            <p class="link-fallback">
                Se o botão não funcionar, copie e cole este link no seu navegador:<br>
                <a href="{{.AcceptURL}}">{{.AcceptURL}}</a>
            </p>
        </div>

        <div class="footer">
            <p>Este é um e-mail automático. Por favor, não responda.</p>
            <p>&copy; 2025 Ecosistema Imob. Todos os direitos reservados.</p>
        </div>
    </div>
</body>
</html>
`

	t, err := template.New("invitation").Parse(tmpl)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// generatePlainTextInvitation generates a plain text version of the invitation email
func (s *EmailService) generatePlainTextInvitation(data InvitationEmailData) string {
	return fmt.Sprintf(`
Olá, %s!

Você foi convidado por %s para fazer parte da equipe %s no Ecosistema Imob.

DETALHES DO CONVITE:
- Empresa: %s
- Função: %s
- Convidado por: %s

Para aceitar o convite e criar sua conta, acesse o link abaixo:

%s

ATENÇÃO: Este convite expira em %d dias (%s).

---
Este é um e-mail automático. Por favor, não responda.
© 2025 Ecosistema Imob. Todos os direitos reservados.
`,
		data.InviteeName,
		data.InviterName,
		data.TenantName,
		data.TenantName,
		data.Role,
		data.InviterName,
		data.AcceptURL,
		data.ExpiresInDays,
		data.ExpiresAt.Format("02/01/2006 às 15:04"),
	)
}

// sendEmail sends an email via SMTP
func (s *EmailService) sendEmail(toEmail, toName, subject, htmlBody, textBody string) error {
	// Build email message
	from := fmt.Sprintf("%s <%s>", s.fromName, s.fromEmail)
	to := fmt.Sprintf("%s <%s>", toName, toEmail)

	// Create MIME message with both HTML and plain text parts
	message := fmt.Sprintf(`From: %s
To: %s
Subject: %s
MIME-Version: 1.0
Content-Type: multipart/alternative; boundary="boundary123"

--boundary123
Content-Type: text/plain; charset="UTF-8"

%s

--boundary123
Content-Type: text/html; charset="UTF-8"

%s

--boundary123--
`, from, to, subject, textBody, htmlBody)

	// Setup authentication
	auth := smtp.PlainAuth("", s.smtpUser, s.smtpPass, s.smtpHost)

	// Connect to the server and send email
	addr := fmt.Sprintf("%s:%d", s.smtpHost, s.smtpPort)

	err := smtp.SendMail(
		addr,
		auth,
		s.fromEmail,
		[]string{toEmail},
		[]byte(message),
	)

	return err
}

// SendPasswordResetEmail sends a password reset email
func (s *EmailService) SendPasswordResetEmail(email, name, resetURL string) error {
	// TODO: Implement password reset email
	log.Printf("📧 Would send password reset email to %s", email)
	return nil
}

// SendWelcomeEmail sends a welcome email after user accepts invitation
func (s *EmailService) SendWelcomeEmail(email, name, tenantName string) error {
	// TODO: Implement welcome email
	log.Printf("📧 Would send welcome email to %s", email)
	return nil
}
