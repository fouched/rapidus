package mailer

import (
	"bytes"
	"fmt"
	"github.com/ainsleyclark/go-mail/drivers"
	"github.com/ainsleyclark/go-mail/mail"
	"github.com/vanng822/go-premailer/premailer"
	smtpmail "github.com/xhit/go-simple-mail/v2"
	"html/template"
	"os"
	"path/filepath"
	"time"
)

type Mail struct {
	Domain      string
	Templates   string
	Host        string
	Port        int
	Username    string
	Password    string
	Encryption  string
	FromAddress string
	FromName    string
	Jobs        chan Message
	Results     chan Result
	API         string
	APIKey      string
	APIUrl      string
}

type Message struct {
	From        string
	FromName    string
	To          string
	Subject     string
	Template    string
	Attachments []string
	Data        interface{}
}

type Result struct {
	Success bool
	Error   error
}

// ListenForMail listens to the mail channel and sends email with it receives a payload.
// It runs continually in the background, and sends error/success messages back on the
// Results channel.
// Note: that if api and api key are set, it will prefer using an api to send mail iso SMTP
func (m *Mail) ListenForMail() {
	for {
		msg := <-m.Jobs
		err := m.Send(msg)
		if err != nil {
			m.Results <- Result{false, err}
		} else {
			m.Results <- Result{true, nil}
		}
	}
}

// Send allows sending of mail directly
// Note: that if api and api key are set, it will prefer using an api to send mail iso SMTP
func (m *Mail) Send(msg Message) error {
	if len(m.API) > 0 && len(m.APIKey)&len(m.APIUrl) > 0 && m.API != "smtp" {
		return m.SendAPIMessage(msg)
	}
	return m.SendSMTPMessage(msg)
}

// SendAPIMessage allows sending of mail directly via API
// mailgun, sparkpost or sendgrid are supported
func (m *Mail) SendAPIMessage(msg Message) error {
	msg = m.sanitizeMessage(msg)

	cfg := mail.Config{
		URL:         m.APIUrl,
		APIKey:      m.APIKey,
		Domain:      m.Domain,
		FromAddress: msg.From,
		FromName:    msg.FromName,
	}

	switch m.API {
	case "mailgun":
		return m.sendUsingMailGun(msg, cfg)
	case "sparkpost":
		return m.sendUsingSparkPost(msg, cfg)
	case "sendgrid":
		return m.sendUsingSendGrid(msg, cfg)
	default:
		return fmt.Errorf("unknown email api %s, only mailgun, sparkpost or sendgrid accepted", m.API)
	}
}

func (m *Mail) sendUsingMailGun(msg Message, cfg mail.Config) error {
	mailer, err := drivers.NewMailgun(cfg)
	err = m.sendMailerMessage(msg, mailer)
	if err != nil {
		return err
	}
	return nil

}

func (m *Mail) sendUsingSparkPost(msg Message, cfg mail.Config) error {
	mailer, err := drivers.NewSparkPost(cfg)
	err = m.sendMailerMessage(msg, mailer)
	if err != nil {
		return err
	}
	return nil
}

func (m *Mail) sendUsingSendGrid(msg Message, cfg mail.Config) error {
	mailer, err := drivers.NewSendGrid(cfg)
	err = m.sendMailerMessage(msg, mailer)
	if err != nil {
		return err
	}
	return nil
}

func (m *Mail) sendMailerMessage(msg Message, mailer mail.Mailer) error {
	formattedMessage, err := m.buildHTMLMessage(msg)
	if err != nil {
		return err
	}

	textMessage, err := m.buildPlainTextMessage(msg)
	if err != nil {
		return err
	}

	tx := &mail.Transmission{
		Recipients: []string{msg.To},
		Subject:    msg.Subject,
		HTML:       formattedMessage,
		PlainText:  textMessage,
	}

	err = m.addAPIAttachments(msg, tx)
	if err != nil {
		return err
	}

	_, err = mailer.Send(tx)
	if err != nil {
		return err
	}

	return nil
}

func (m *Mail) addAPIAttachments(msg Message, tx *mail.Transmission) error {
	if len(msg.Attachments) > 0 {
		var attachments []mail.Attachment

		for _, x := range msg.Attachments {
			var attach mail.Attachment
			content, err := os.ReadFile(x)
			if err != nil {
				return err
			}

			fileName := filepath.Base(x)
			attach.Bytes = content
			attach.Filename = fileName
			attachments = append(attachments, attach)
		}

		tx.Attachments = attachments
	}

	return nil
}

// SendSMTPMessage builds and sends an email using SMTP. It is called by ListenForMail
// and can also be called directly
func (m *Mail) SendSMTPMessage(msg Message) error {
	msg = m.sanitizeMessage(msg)
	formattedMessage, err := m.buildHTMLMessage(msg)
	if err != nil {
		return err
	}

	textMessage, err := m.buildPlainTextMessage(msg)
	if err != nil {
		return err
	}

	server := smtpmail.NewSMTPClient()
	server.Host = m.Host
	server.Port = m.Port
	server.Username = m.Username
	server.Password = m.Password
	server.Encryption = m.getEncryption(m.Encryption)
	server.KeepAlive = false
	server.ConnectTimeout = 10 * time.Second
	server.SendTimeout = 10 * time.Second

	smtpClient, err := server.Connect()
	if err != nil {
		return err
	}

	email := smtpmail.NewMSG()
	email.SetFrom(msg.From).
		AddTo(msg.To).
		SetSubject(msg.Subject)

	email.SetBody(smtpmail.TextHTML, formattedMessage)
	email.AddAlternative(smtpmail.TextPlain, textMessage)

	if len(msg.Attachments) > 0 {
		for _, x := range msg.Attachments {
			email.AddAttachment(x)
		}
	}

	err = email.Send(smtpClient)
	if err != nil {
		return err
	}

	return nil
}

func (m *Mail) sanitizeMessage(msg Message) Message {

	if msg.From == "" {
		msg.From = m.FromAddress
	}

	if msg.FromName == "" {
		msg.FromName = m.FromName
	}

	return msg
}

func (m *Mail) getEncryption(e string) smtpmail.Encryption {
	switch e {
	case "tls":
		return smtpmail.EncryptionSTARTTLS
	case "ssl":
		return smtpmail.EncryptionSSL
	case "none":
		return smtpmail.EncryptionNone
	default:
		return smtpmail.EncryptionSTARTTLS
	}
}

func (m *Mail) buildHTMLMessage(msg Message) (string, error) {
	templateToRender := fmt.Sprintf("%s/%s.html.tmpl", m.Templates, msg.Template)

	t, err := template.New("email-html").ParseFiles(templateToRender)
	if err != nil {
		return "", err
	}

	var tpl bytes.Buffer
	if err = t.ExecuteTemplate(&tpl, "body", msg.Data); err != nil {
		return "", err
	}

	formattedMessage := tpl.String()
	formattedMessage, err = m.inlineCSS(formattedMessage)
	if err != nil {
		return "", err
	}

	return formattedMessage, nil
}

func (m *Mail) buildPlainTextMessage(msg Message) (string, error) {
	templateToRender := fmt.Sprintf("%s/%s.text.tmpl", m.Templates, msg.Template)

	t, err := template.New("email-text").ParseFiles(templateToRender)
	if err != nil {
		return "", err
	}

	var tpl bytes.Buffer
	if err = t.ExecuteTemplate(&tpl, "body", msg.Data); err != nil {
		return "", err
	}

	textMessage := tpl.String()

	return textMessage, nil
}

func (m *Mail) inlineCSS(s string) (string, error) {
	options := premailer.Options{
		RemoveClasses:     false,
		CssToAttributes:   false,
		KeepBangImportant: true,
	}

	prem, err := premailer.NewPremailerFromString(s, &options)
	if err != nil {
		return "", err
	}

	html, err := prem.Transform()
	if err != nil {
		return "", err
	}

	return html, nil
}
