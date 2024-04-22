package main

import (
	//"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"

	//"strings"
	"mime"
	//"mime/quotedprintable"
	"net/mail"
	"net/smtp"

	"gopkg.in/mailgun/mailgun-go.v1"
)

// https://stackoverflow.com/a/59355954
type loginAuth struct {
	username, password string
}

// LoginAuth is used for smtp login auth
func LoginAuth(username, password string) smtp.Auth {
	return &loginAuth{username, password}
}

func (a *loginAuth) Start(server *smtp.ServerInfo) (string, []byte, error) {
	return "LOGIN", []byte(a.username), nil
}

func (a *loginAuth) Next(fromServer []byte, more bool) ([]byte, error) {
	if more {
		switch string(fromServer) {
		case "Username:":
			return []byte(a.username), nil
		case "Password:":
			return []byte(a.password), nil
		default:
			return nil, errors.New("Unknown from server")
		}
	}
	return nil, nil
}

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
	AuthType string `json:"type"`
	Identity string `json:"identity"`
	Username string `json:"username"`
	Password string `json:"password"`
	Host     string `json:"host"`
}
type SmtpTransport struct {
	Host string   `json:"host"`
	Auth SmtpAuth `json:"auth"`
}

func (t SmtpTransport) Send(email Mail) error {
	header := make(map[string]string)
	from, err := mail.ParseAddress(email.Sender)
	if err != nil {
		return err
	}
	header["From"] = from.String()
	if email.Recipient != "" {
		to, err := mail.ParseAddress(email.Recipient)
		if err != nil {
			return err
		}
		header["To"] = to.String()
	}
	header["Subject"] = mime.QEncoding.Encode("utf-8", email.Subject)
	header["MIME-Version"] = "1.0"
	header["Content-Type"] = "text/plain; charset=\"utf-8\""
	//header["Content-Transfer-Encoding"] = "base64"
	//header["Content-Transfer-Encoding"] = "quoted-printable"
	message := ""
	for k, v := range header {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}

	//message += "\r\n" + base64.StdEncoding.EncodeToString([]byte(email.Body))
	//builder := new(strings.Builder)
	//w := quotedprintable.NewWriter(builder)
	//w.Write([]byte(email.Body))
	//w.Close()
	//message += "\r\n" + builder.String()
	message += "\r\n" + email.Body

	recpt := make([]string, 0)
	if email.Recipient != "" {
		recpt = append(recpt, email.Recipient)
	}
	recpt = append(recpt, email.BCC[:]...)

	var auth smtp.Auth
	if t.Auth.AuthType == "plain" {
		auth = smtp.PlainAuth(t.Auth.Identity, t.Auth.Username, t.Auth.Password, t.Auth.Host)
	} else if t.Auth.AuthType == "login" {
		auth = LoginAuth(t.Auth.Username, t.Auth.Password)
	} else {
		return errors.New("Invalid auth type: " + t.Auth.AuthType)
	}
	err = smtp.SendMail(
		t.Host,
		auth,
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
