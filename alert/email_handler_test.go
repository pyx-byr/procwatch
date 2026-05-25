package alert

import (
	"net"
	"net/smtp"
	"testing"
	"time"
)

// startFakeSMTP starts a minimal TCP listener that accepts one connection and
// speaks just enough SMTP for smtp.SendMail to succeed.
func startFakeSMTP(t *testing.T) (addr string, done <-chan struct{}) {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	ch := make(chan struct{})
	go func() {
		defer close(ch)
		defer ln.Close()
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		conn.Write([]byte("220 localhost SMTP\r\n"))
		buf := make([]byte, 4096)
		for {
			n, _ := conn.Read(buf)
			if n == 0 {
				return
			}
			cmd := string(buf[:n])
			switch {
			case len(cmd) >= 4 && cmd[:4] == "EHLO":
				conn.Write([]byte("250-localhost\r\n250 AUTH PLAIN\r\n"))
			case len(cmd) >= 4 && cmd[:4] == "AUTH":
				conn.Write([]byte("235 OK\r\n"))
			case len(cmd) >= 4 && cmd[:4] == "MAIL":
				conn.Write([]byte("250 OK\r\n"))
			case len(cmd) >= 4 && cmd[:4] == "RCPT":
				conn.Write([]byte("250 OK\r\n"))
			case len(cmd) >= 4 && cmd[:4] == "DATA":
				conn.Write([]byte("354 Start\r\n"))
			case len(cmd) >= 1 && cmd[len(cmd)-1] == '\n' && len(cmd) >= 3 && cmd[len(cmd)-3:] == ".\r\n":
				conn.Write([]byte("250 OK\r\n"))
			case len(cmd) >= 4 && cmd[:4] == "QUIT":
				conn.Write([]byte("221 Bye\r\n"))
				return
			default:
				conn.Write([]byte("250 OK\r\n"))
			}
		}
	}()
	return ln.Addr().String(), ch
}

func TestNewEmailHandler_FieldsSet(t *testing.T) {
	cfg := EmailConfig{
		SMTPHost: "localhost",
		SMTPPort: 2525,
		Username: "user",
		Password: "pass",
		From:     "alert@example.com",
		To:       []string{"ops@example.com"},
	}
	h := NewEmailHandler(cfg)
	if h == nil {
		t.Fatal("expected non-nil handler")
	}
	if _, ok := interface{}(h.auth).(smtp.Auth); !ok {
		t.Fatal("expected smtp.Auth to be set")
	}
}

func TestEmailHandler_Handle_ConnectionRefused(t *testing.T) {
	cfg := EmailConfig{
		SMTPHost: "127.0.0.1",
		SMTPPort: 19999, // nothing listening
		From:     "a@b.com",
		To:       []string{"c@d.com"},
	}
	h := NewEmailHandler(cfg)
	e := Event{
		ProcessName: "myapp",
		Metric:      "cpu",
		Value:       95.0,
		Threshold:   80.0,
		Timestamp:   time.Now(),
	}
	if err := h.Handle(e); err == nil {
		t.Fatal("expected error for refused connection, got nil")
	}
}
