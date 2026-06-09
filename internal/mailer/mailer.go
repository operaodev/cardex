package mailer

import (
	"fmt"
	"net/smtp"
)

type Mailer interface {
	SendVerificationEmail(to, name, token string) error
}

type smtpMailer struct {
	host     string
	port     string
	username string
	password string
	from     string
	baseURL  string
}

func NewSMTPMailer(host, port, username, password, from, baseURL string) Mailer {
	return &smtpMailer{
		host:     host,
		port:     port,
		username: username,
		password: password,
		from:     from,
		baseURL:  baseURL,
	}
}

func (m *smtpMailer) SendVerificationEmail(to, name, token string) error {
	subject := "Verifica tu cuenta"
	verifyURL := fmt.Sprintf("%s/users/verify?token=%s", m.baseURL, token)

	body := fmt.Sprintf(`Hola %s,

Gracias por registrarte. Por favor verifica tu correo electrónico haciendo clic en el siguiente enlace:

%s

Si no creaste esta cuenta, puedes ignorar este mensaje.

Saludos,
Equipo Cardex`, name, verifyURL)

	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s",
		m.from, to, subject, body)

	auth := smtp.PlainAuth("", m.username, m.password, m.host)
	addr := fmt.Sprintf("%s:%s", m.host, m.port)

	return smtp.SendMail(addr, auth, m.from, []string{to}, []byte(msg))
}
