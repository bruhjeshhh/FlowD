package worker

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/smtp"
	"os"
	"strings"
)

type EmailPayload struct {
	To      string            `json:"to"`
	From    string            `json:"from,omitempty"`
	Subject string            `json:"subject"`
	Body    string            `json:"body"`
	HTML    string            `json:"html,omitempty"`
	CC      []string          `json:"cc,omitempty"`
	BCC     []string          `json:"bcc,omitempty"`
	Headers map[string]string `json:"headers,omitempty"`
}

func handleemails(log *slog.Logger, payload []byte) (bool, error) {
	var email EmailPayload
	if err := json.Unmarshal(payload, &email); err != nil {
		log.Error("failed to parse email payload", "error", err)
		return false, fmt.Errorf("parse payload: %w", err)
	}

	if email.To == "" {
		log.Error("email 'to' field is required")
		return false, fmt.Errorf("missing 'to' field")
	}

	if email.From != "" && !strings.Contains(email.From, "@") {
		log.Error("email 'from' field is invalid")
		return false, fmt.Errorf("invalid 'from' field")
	}

	if email.Subject == "" {
		email.Subject = "No Subject"
	}

	if email.Body == "" && email.HTML == "" {
		log.Error("email body or html is required")
		return false, fmt.Errorf("missing body and html")
	}

	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")
	smtpUsername := os.Getenv("SMTP_USERNAME")
	smtpPassword := os.Getenv("SMTP_PASSWORD")
	smtpFrom := os.Getenv("SMTP_FROM")

	if smtpHost == "" {
		log.Error("SMTP_HOST not configured")
		return false, fmt.Errorf("SMTP_HOST not configured")
	}

	if smtpFrom == "" {
		smtpFrom = "FlowD <noreply@flowd.local>"
	}

	from := email.From
	if from == "" {
		from = smtpFrom
	}

	if err := sendEmail(smtpHost, smtpPort, smtpUsername, smtpPassword, from, &email); err != nil {
		log.Error("failed to send email", "error", err)
		return false, fmt.Errorf("send email: %w", err)
	}

	log.Info("email sent successfully", "to", email.To, "subject", email.Subject)
	return true, nil
}

func sendEmail(host, port, username, password, from string, email *EmailPayload) error {
	var msg strings.Builder

	msg.WriteString(fmt.Sprintf("From: %s\r\n", from))
	msg.WriteString(fmt.Sprintf("To: %s\r\n", email.To))
	msg.WriteString(fmt.Sprintf("Subject: %s\r\n", email.Subject))

	to := []string{email.To}
	if len(email.BCC) > 0 {
		msg.WriteString(fmt.Sprintf("Bcc: %s\r\n", strings.Join(email.BCC, ",")))
		to = append(to, email.BCC...)
	}

	if len(email.CC) > 0 {
		msg.WriteString(fmt.Sprintf("Cc: %s\r\n", strings.Join(email.CC, ",")))
		to = append(to, email.CC...)
	}

	if email.Headers["Reply-To"] == "" && from != "" {
		msg.WriteString(fmt.Sprintf("Reply-To: %s\r\n", from))
	}

	for k, v := range email.Headers {
		msg.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
	}

	msg.WriteString("MIME-Version: 1.0\r\n")

	if email.HTML != "" {
		msg.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
	} else {
		msg.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
	}

	msg.WriteString("\r\n")

	if email.HTML != "" {
		msg.WriteString(email.HTML)
	} else {
		msg.WriteString(email.Body)
	}

	addr := fmt.Sprintf("%s:%s", host, port)

	var auth smtp.Auth
	if username != "" && password != "" {
		auth = smtp.PlainAuth("", username, password, host)
	}

	return smtp.SendMail(addr, auth, from, to, []byte(msg.String()))
}
