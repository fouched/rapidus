package mailer

import (
	"errors"
	"testing"
)

var msg = Message{
	From:        "me@here.com",
	FromName:    "Joe",
	To:          "you@there.com",
	Subject:     "test",
	Template:    "test",
	Attachments: []string{"./testdata/mail/test.html.tmpl"},
}

func TestMail_SendSMTPMessage(t *testing.T) {
	err := mailer.SendSMTPMessage(msg)
	if err != nil {
		t.Error(err)
	}
}

func TestMail_SendUsingChan(t *testing.T) {
	mailer.Jobs <- msg
	res := <-mailer.Results
	if res.Error != nil {
		t.Error(errors.New("failed to send over channel"))
	}

	msg.To = "not_an_email"
	mailer.Jobs <- msg
	res = <-mailer.Results
	if res.Error == nil {
		t.Error(errors.New("no error received with invalid to address"))
	}

	msg.To = "you@there.com"
}

func TestMail_SendAPIMessage(t *testing.T) {
	mailer.API = "unknown"
	mailer.APIKey = "abc123"
	mailer.APIUrl = "https://www.fake.com"

	err := mailer.SendAPIMessage(msg)
	if err == nil {
		t.Error(errors.New("no error for invalid api"))
	}

	mailer.API = ""
	mailer.APIKey = ""
	mailer.APIUrl = ""
}

func TestMail_buildHTMLMessage(t *testing.T) {
	_, err := mailer.buildHTMLMessage(msg)
	if err != nil {
		t.Error(err)
	}
}

func TestMail_buildPlainTextMessage(t *testing.T) {
	_, err := mailer.buildPlainTextMessage(msg)
	if err != nil {
		t.Error(err)
	}
}

func TestMail_Send(t *testing.T) {
	err := mailer.Send(msg)
	if err != nil {
		t.Error(err)
	}

	mailer.API = "unknown"
	mailer.APIKey = "abc123"
	mailer.APIUrl = "https://www.fake.com"

	err = mailer.Send(msg)
	if err == nil {
		t.Error(errors.New("no error for invalid api"))
	}

	mailer.API = ""
	mailer.APIKey = ""
	mailer.APIUrl = ""
}
