package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/mail"
	"net/smtp"
	"strings"

	"gopkg.in/mailgun/mailgun-go.v1"
)

type TransportType string

const (
	TransportSmtp    TransportType = "smtp"
	TransportMailgun               = "mailgun"
	TransportNull                  = "null"
)

type ConfigMailtransport struct {
	Type      TransportType `json:"type" validate:"required"`
	Transport interface{}   `json:"transport"`
}

func (m *ConfigMailtransport) GetTransport() MailTransport {
	if transport, ok := m.Transport.(MailTransport); ok {
		return transport
	}
	return NullTransport{}
}

type configMailtransport ConfigMailtransport

func (m *ConfigMailtransport) UnmarshalJSON(b []byte) error {
	var msg json.RawMessage
	var mt configMailtransport
	mt.Transport = &msg
	if err := json.Unmarshal(b, &mt); err != nil {
		return err
	}

	*m = ConfigMailtransport(mt)
	switch mt.Type {
	case TransportSmtp:
		var t SmtpTransport
		if err := json.Unmarshal(msg, &t); err != nil {
			return err
		}
		(*m).Transport = t
		return nil
	case TransportMailgun:
		var t MailgunTransport
		if err := json.Unmarshal(msg, &t); err != nil {
			return err
		}
		(*m).Transport = t
		return nil
	}

	var t NullTransport
	(*m).Type = "null"
	(*m).Transport = t
	return nil
}

type MailTransport interface {
	Send(mail Mail) error
}

type NullTransport struct {
}

func (t NullTransport) Send(mail Mail) error {
	return nil
}

type SmtpAuth struct {
	Identity string `json:"identity"`
	Username string `json:"username"`
	Password string `json:"password"`
	Host     string `json:"host"`
}
type SmtpTransport struct {
	Host string   `json:"host"`
	Auth SmtpAuth `json:"auth"`
}

func encodeRFC2047(String string) string {
	// use mail's rfc2047 to encode any string
	addr := mail.Address{String, ""}
	return strings.Trim(addr.String(), " <>")
}

func (t SmtpTransport) Send(email Mail) error {
	from := mail.Address{email.Recipient, email.Sender}
	to := mail.Address{email.Recipient, email.Recipient}

	header := make(map[string]string)
	header["From"] = from.String()
	header["To"] = to.String()
	header["Subject"] = encodeRFC2047(email.Subject)
	header["MIME-Version"] = "1.0"
	header["Content-Type"] = "text/plain; charset=\"utf-8\""
	header["Content-Transfer-Encoding"] = "base64"
	message := ""
	for k, v := range header {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}

	message += "\r\n" + base64.StdEncoding.EncodeToString([]byte(email.Body))

	recpt := make([]string, 0)
	if email.Recipient != "" {
		recpt = append(recpt, email.Recipient)
	}
	recpt = append(recpt, email.BCC[:]...)

	err := smtp.SendMail(
		t.Host,
		smtp.PlainAuth(t.Auth.Identity, t.Auth.Username, t.Auth.Password, t.Auth.Host),
		email.Sender,
		recpt,
		[]byte(message),
	)
	if err != nil {
		return err
	}
	return nil
}

type MailgunTransport struct {
	Domain    string `json:"domain"`
	SecretKey string `json:"secretkey"`
	PublicKey string `json:"publickey"`
}

func (t MailgunTransport) Send(mail Mail) error {
	m := mailgun.NewMailgun(
		t.Domain,
		t.SecretKey,
		t.PublicKey,
	)
	message := m.NewMessage(mail.Sender, mail.Subject, mail.Body, mail.Recipient)
	_, _, err := m.Send(message)
	if err != nil {
		return err
	}

	return nil
}
