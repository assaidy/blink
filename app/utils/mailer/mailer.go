package email

// Mailer defines the interface for sending emails.
type Mailer interface {
	// SendEmail sends an email to the specified recipient with the given subject and body.
	SendEmail(to, subject, body string) error
}
