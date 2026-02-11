package email

import (
	"fmt"

	"github.com/assaidy/blink/app/env"
	"gopkg.in/gomail.v2"
)

type PapercutMailer struct {
	dialer *gomail.Dialer
}

func NewPapercutMailer() PapercutMailer {
	dialer := &gomail.Dialer{
		Host: env.PapercutSmtHost,
		Port: 25,
	}
	return PapercutMailer{dialer: dialer}
}

func (me PapercutMailer) SendEmail(to, subject, body string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", fmt.Sprintf("Blink Chat App <%s>", env.EmailFrom))
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/plain; charset=UTF-8", body)
	return me.dialer.DialAndSend(m)
}
