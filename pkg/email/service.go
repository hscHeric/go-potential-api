// Package email contem o serviço de envio de emails
package email

import (
	"bytes"
	"fmt"
	"html/template"

	"gopkg.in/gomail.v2"
)

type Service struct {
	smtpHost     string
	smtpPort     int
	smtpUsername string
	smtpPassword string
	fromEmail    string
	frontendURL  string
}

// NewService cria uma nova instância do serviço de email
func NewService(host string, port int, username, password, fromEmail, frontendURL string) *Service {
	return &Service{
		smtpHost:     host,
		smtpPort:     port,
		smtpUsername: username,
		smtpPassword: password,
		fromEmail:    fromEmail,
		frontendURL:  frontendURL,
	}
}

// SendActivationEmail envia email de ativação de conta
func (s *Service) SendActivationEmail(to, name, token string) error {
	activationURL := fmt.Sprintf("%s/activate?token=%s", s.frontendURL, token)

	data := map[string]any{
		"Name":          name,
		"ActivationURL": activationURL,
	}

	subject := "Ative sua conta - Potential Idiomas"
	body, err := s.renderTemplate("activation", data)
	if err != nil {
		return err
	}

	return s.sendEmail(to, subject, body)
}

// SendPasswordResetEmail envia email de recuperação de senha
func (s *Service) SendPasswordResetEmail(to, name, token string) error {
	resetURL := fmt.Sprintf("%s/reset-password?token=%s", s.frontendURL, token)

	data := map[string]interface{}{
		"Name":     name,
		"ResetURL": resetURL,
	}

	subject := "Recuperação de Senha - Potential Idiomas"
	body, err := s.renderTemplate("password_reset", data)
	if err != nil {
		return err
	}

	return s.sendEmail(to, subject, body)
}

// SendWelcomeEmail envia email de boas-vindas após ativação
func (s *Service) SendWelcomeEmail(to, name string) error {
	data := map[string]interface{}{
		"Name": name,
	}

	subject := "Bem-vindo à Potential Idiomas!"
	body, err := s.renderTemplate("welcome", data)
	if err != nil {
		return err
	}

	return s.sendEmail(to, subject, body)
}

// sendEmail envia um email usando SMTP
func (s *Service) sendEmail(to, subject, body string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", s.fromEmail)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)

	d := gomail.NewDialer(s.smtpHost, s.smtpPort, s.smtpUsername, s.smtpPassword)

	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

// renderTemplate renderiza um template de email
func (s *Service) renderTemplate(templateName string, data map[string]any) (string, error) {
	templates := map[string]string{
		"activation": `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Ativação de Conta</title>
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <div style="max-width: 600px; margin: 0 auto; padding: 20px;">
        <h2 style="color: #2c3e50;">Olá, {{.Name}}!</h2>
        <p>Você foi convidado para fazer parte da Potential Idiomas.</p>
        <p>Para ativar sua conta e completar seu cadastro, clique no botão abaixo:</p>
        <div style="text-align: center; margin: 30px 0;">
            <a href="{{.ActivationURL}}" 
               style="background-color: #3498db; color: white; padding: 12px 30px; 
                      text-decoration: none; border-radius: 5px; display: inline-block;">
                Ativar Conta
            </a>
        </div>
        <p style="color: #7f8c8d; font-size: 14px;">
            Se o botão não funcionar, copie e cole o link abaixo no seu navegador:
        </p>
        <p style="word-break: break-all; color: #3498db; font-size: 12px;">
            {{.ActivationURL}}
        </p>
        <p style="margin-top: 30px; color: #7f8c8d; font-size: 12px;">
            Este link expira em 48 horas.
        </p>
    </div>
</body>
</html>
`,
		"password_reset": `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Recuperação de Senha</title>
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <div style="max-width: 600px; margin: 0 auto; padding: 20px;">
        <h2 style="color: #2c3e50;">Olá, {{.Name}}!</h2>
        <p>Recebemos uma solicitação para redefinir sua senha.</p>
        <p>Para criar uma nova senha, clique no botão abaixo:</p>
        <div style="text-align: center; margin: 30px 0;">
            <a href="{{.ResetURL}}" 
               style="background-color: #e74c3c; color: white; padding: 12px 30px; 
                      text-decoration: none; border-radius: 5px; display: inline-block;">
                Redefinir Senha
            </a>
        </div>
        <p style="color: #7f8c8d; font-size: 14px;">
            Se você não solicitou esta redefinição, ignore este email.
        </p>
        <p style="word-break: break-all; color: #e74c3c; font-size: 12px;">
            {{.ResetURL}}
        </p>
        <p style="margin-top: 30px; color: #7f8c8d; font-size: 12px;">
            Este link expira em 1 hora.
        </p>
    </div>
</body>
</html>
`,
		"welcome": `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Bem-vindo!</title>
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <div style="max-width: 600px; margin: 0 auto; padding: 20px;">
        <h2 style="color: #27ae60;">Bem-vindo à Potential Idiomas, {{.Name}}!</h2>
        <p>Sua conta foi ativada com sucesso.</p>
        <p>Agora você pode acessar a plataforma e começar sua jornada de aprendizado.</p>
        <p style="margin-top: 30px;">
            Se tiver alguma dúvida, entre em contato conosco.
        </p>
        <p>Atenciosamente,<br>Equipe Potential Idiomas</p>
    </div>
</body>
</html>
`,
	}

	tmplContent, ok := templates[templateName]
	if !ok {
		return "", fmt.Errorf("template não encontrado: %s", templateName)
	}

	tmpl, err := template.New(templateName).Parse(tmplContent)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}
