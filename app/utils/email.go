package utils

import (
	"fmt"

	"github.com/assaidy/blink/app/env"
	"gopkg.in/gomail.v2"
)

func SendEmail(to, subject, body string) error {
	dialer := &gomail.Dialer{
		Host:     env.SmtpHost,
		Port:     25,
		Username: env.SmtpUsername,
		Password: env.SmtpPassword,
	}

	m := gomail.NewMessage()
	m.SetHeader("From", fmt.Sprintf("Blink Chat App <%s>", env.EmailFrom))
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/plain; charset=UTF-8", body)

	return dialer.DialAndSend(m)
}
