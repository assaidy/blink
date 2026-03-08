package email

import (
	"fmt"

	"gopkg.in/gomail.v2"
)

type PapercutMailer struct {
	dialer    *gomail.Dialer
	emailFrom string
}

func NewPapercutMailer(smtpHost, emailFrom string) PapercutMailer {
	dialer := &gomail.Dialer{
		Host: smtpHost,
		Port: 25,
	}
	return PapercutMailer{dialer: dialer, emailFrom: emailFrom}
}

func (me PapercutMailer) SendEmail(to, subject, body string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", fmt.Sprintf("Blink Chat App <%s>", me.emailFrom))
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/plain; charset=UTF-8", body)
	return me.dialer.DialAndSend(m)
}
