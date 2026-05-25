package alert

import (
	"fmt"
	"net/smtp"
	"strings"
	"time"
)

// EmailConfig holds SMTP configuration for the email handler.
type EmailConfig struct {
	SMTPHost string
	SMTPPort int
	Username string
	Password string
	From     string
	To       []string
}

// EmailHandler sends alert events via email over SMTP.
type EmailHandler struct {
	cfg  EmailConfig
	auth smtp.Auth
}

// NewEmailHandler creates an EmailHandler with the given SMTP config.
func NewEmailHandler(cfg EmailConfig) *EmailHandler {
	auth := smtp.PlainAuth("", cfg.Username, cfg.Password, cfg.SMTPHost)
	return &EmailHandler{cfg: cfg, auth: auth}
}

// Handle sends an email notification for the given alert event.
func (h *EmailHandler) Handle(e Event) error {
	addr := fmt.Sprintf("%s:%d", h.cfg.SMTPHost, h.cfg.SMTPPort)
	subject := fmt.Sprintf("[procwatch] Alert: %s", e.ProcessName)
	body := fmt.Sprintf(
		"Process : %s\nMetric  : %s\nValue   : %.2f\nThreshold: %.2f\nTime    : %s\n",
		e.ProcessName,
		e.Metric,
		e.Value,
		e.Threshold,
		e.Timestamp.Format(time.RFC3339),
	)
	msg := []byte("To: " + strings.Join(h.cfg.To, ",") + "\r\n" +
		"From: " + h.cfg.From + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"\r\n" +
		body)
	return smtp.SendMail(addr, h.auth, h.cfg.From, h.cfg.To, msg)
}
