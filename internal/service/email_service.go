package service

import (
	"fmt"
	"net/smtp"

	"github.com/hscHeric/go-potential-api/internal/config"
	"github.com/jordan-wright/email"
)

type EmailService interface {
	SendInvitationEmail(to, token, role string) error
	SendPasswordResetEmail(to, token string) error
}

type emailService struct {
	cfg *config.Config
}

func NewEmailService(cfg *config.Config) EmailService {
	return &emailService{cfg: cfg}
}

/*
* Lebrar de extrair os emails para um arquivo a parte de forma que fique mais facil modificar quando quiser
 */
func (s *emailService) SendInvitationEmail(to, token, role string) error {
	invitationLink := fmt.Sprintf("%s/complete-registration?token=%s", s.cfg.Server.FrontendURL, token)

	e := email.NewEmail()
	e.From = s.cfg.Email.From
	e.To = []string{to}
	e.Subject = "Convite para completar cadastro - Escola"
	e.HTML = fmt.Appendf(nil, `
		<h2>Bem-vindo à Potential-Idiomas!</h2>
		<p>Você foi convidado para se cadastrar como <strong>%s</strong>.</p>
		<p>Para completar seu cadastro, clique no link abaixo:</p>
		<p><a href="%s">Completar Cadastro</a></p>
		<p>Este link expira em 72 horas.</p>
	`, role, invitationLink)

	// host para conexão smtp
	addr := fmt.Sprintf("%s:%s", s.cfg.Email.SMTPHost, s.cfg.Email.SMTPPort)

	// Se não tiver auth como MailHog, enviar sem autenticação
	if s.cfg.Email.SMTPUser == "" {
		return e.Send(addr, nil)
	}

	auth := smtp.PlainAuth("", s.cfg.Email.SMTPUser, s.cfg.Email.SMTPPass, s.cfg.Email.SMTPHost)
	return e.Send(addr, auth)
}

func (s *emailService) SendPasswordResetEmail(to, token string) error {
	resetLink := fmt.Sprintf("%s/reset-password?token=%s", s.cfg.Server.FrontendURL, token)

	e := email.NewEmail()
	e.From = s.cfg.Email.From
	e.To = []string{to}
	e.Subject = "Recuperação de senha - Potential-Idiomas"
	e.HTML = fmt.Appendf(nil, `
		<h2>Recuperação de Senha</h2>
		<p>Você solicitou a recuperação de senha.</p>
		<p>Para redefinir sua senha, clique no link abaixo:</p>
		<p><a href="%s">Redefinir Senha</a></p>
		<p>Este link expira em 2 horas.</p>
		<p>Se você não solicitou esta recuperação, ignore este email.</p>
	`, resetLink)

	addr := fmt.Sprintf("%s:%s", s.cfg.Email.SMTPHost, s.cfg.Email.SMTPPort)

	if s.cfg.Email.SMTPUser == "" {
		return e.Send(addr, nil)
	}

	auth := smtp.PlainAuth("", s.cfg.Email.SMTPUser, s.cfg.Email.SMTPPass, s.cfg.Email.SMTPHost)
	return e.Send(addr, auth)
}
