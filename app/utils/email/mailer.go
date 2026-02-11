package email

type Mailer interface {
	SendEmail(to, subject, body string) error
}
