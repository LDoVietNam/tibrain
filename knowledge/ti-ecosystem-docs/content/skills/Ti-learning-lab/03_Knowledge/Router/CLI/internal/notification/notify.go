package notification

import (
	"fmt"
	"net/smtp"
	"os/exec"
)

func SendToast(title, message string) error {
	if title == "" {
		title = "Ti CLI"
	}

	cmd := exec.Command("powershell", "-Command",
		fmt.Sprintf("Add-Type -AssemblyName System.Windows.Forms; [System.Windows.Forms.MessageBox]::Show('%s', '%s')", message, title))
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Run()
}

func SendEmail(to, subject, body, smtpHost, smtpPort, from, password string) error {
	addr := smtpHost + ":" + smtpPort

	auth := smtp.PlainAuth("", from, password, smtpHost)

	msg := []byte("From: " + from + "\r\n" +
		"To: " + to + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"\r\n" +
		body + "\r\n")

	err := smtp.SendMail(addr, auth, from, []string{to}, msg)
	if err != nil {
		return fmt.Errorf("send email: %w", err)
	}

	return nil
}

type Notify struct {
	EmailTo  string
	SMTPHost string
	SMTPPort string
	From     string
	Password string
}

func NewNotify(emailTo, smtpHost, smtpPort, from, password string) *Notify {
	return &Notify{
		EmailTo:  emailTo,
		SMTPHost: smtpHost,
		SMTPPort: smtpPort,
		From:     from,
		Password: password,
	}
}

func (n *Notify) Toast(title, message string) error {
	return SendToast(title, message)
}

func (n *Notify) Email(subject, body string) error {
	if n.EmailTo == "" || n.From == "" || n.Password == "" {
		return fmt.Errorf("email not configured")
	}
	return SendEmail(n.EmailTo, subject, body, n.SMTPHost, n.SMTPPort, n.From, n.Password)
}
