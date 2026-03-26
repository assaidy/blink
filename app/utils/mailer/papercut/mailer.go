package papercut

import (
	"fmt"

	"github.com/assaidy/blink/app/utils/mailer"
	"gopkg.in/gomail.v2"
)

type Mailer struct {
	dialer    *gomail.Dialer
	emailFrom string
}

func New(smtpHost, emailFrom string) Mailer {
	m := Mailer{
		dialer: &gomail.Dialer{
			Host: smtpHost,
			Port: 25,
		},
		emailFrom: emailFrom,
	}
	_ = mailer.Mailer(m)
	return m
}

func (me Mailer) SendEmail(to, subject, body string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", fmt.Sprintf("Blink Chat App <%s>", me.emailFrom))
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/plain; charset=UTF-8", body)
	return me.dialer.DialAndSend(m)
}
