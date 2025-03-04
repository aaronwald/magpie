package email

import (
	"crypto/tls"
	"fmt"
	"log/slog"
	"net/smtp"
)

func Send(to, subject, body string, gmail_username string, gmail_password string) error {
	from := gmail_username
	password := gmail_password

	// Gmail SMTP server address
	smtpHost := "smtp.gmail.com"
	smtpPort := "465"

	// Message
	message := []byte("Subject: " + subject + "\r\n" +
		"\r\n" + body + "\r\n")

	// Authentication
	auth := smtp.PlainAuth("", from, password, smtpHost)

	// TLS config
	tlsconfig := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         smtpHost,
	}

	// Connect to the SMTP server
	conn, err := tls.Dial("tcp", smtpHost+":"+smtpPort, tlsconfig)
	if err != nil {
		return fmt.Errorf("failed to connect to SMTP server: %w", err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, smtpHost)
	if err != nil {
		return fmt.Errorf("failed to create SMTP client: %w", err)
	}
	// Authenticate
	if err = client.Auth(auth); err != nil {
		return fmt.Errorf("failed to authenticate: %w", err)
	}

	// Set the sender and recipient
	if err = client.Mail(from); err != nil {
		return fmt.Errorf("failed to set sender: %w", err)
	}
	if err = client.Rcpt(to); err != nil {
		return fmt.Errorf("failed to set recipient: %w", err)
	}

	// Send the email body
	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("failed to send email body: %w", err)
	}
	_, err = w.Write(message)
	if err != nil {
		return fmt.Errorf("failed to write email body: %w", err)
	}
	err = w.Close()
	if err != nil {
		return fmt.Errorf("failed to close email body: %w", err)
	}

	// Close the connection
	client.Quit()

	slog.Info("Email sent successfully")
	return nil
}
