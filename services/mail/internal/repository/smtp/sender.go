package smtp

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/mephistolie/chefbook-backend-mail/internal/entity"
	"net"
	"net/mail"
	"net/smtp"
	"net/textproto"
	"strings"
	"time"
)

type Sender struct {
	Host, Username, Password, From string
	Port                           int
	Timeout                        time.Duration
}

func (s Sender) Send(ctx context.Context, m entity.Mail) (result error) {
	defer func() { result = classify(result) }()
	ctx, cancel := context.WithTimeout(ctx, s.Timeout)
	defer cancel()
	address := net.JoinHostPort(s.Host, fmt.Sprint(s.Port))
	d := net.Dialer{}
	conn, err := d.DialContext(ctx, "tcp", address)
	if err != nil {
		return err
	}
	defer conn.Close()
	rawConnection := conn
	stop := context.AfterFunc(ctx, func() { _ = rawConnection.Close() })
	defer stop()
	if deadline, ok := ctx.Deadline(); ok {
		_ = conn.SetDeadline(deadline)
	}
	cfg := &tls.Config{ServerName: s.Host, MinVersion: tls.VersionTLS12}
	if s.Port == 465 {
		tc := tls.Client(conn, cfg)
		if err = tc.HandshakeContext(ctx); err != nil {
			return err
		}
		conn = tc
	}
	c, err := smtp.NewClient(conn, s.Host)
	if err != nil {
		return err
	}
	defer c.Close()
	if s.Port != 465 {
		if err = c.StartTLS(cfg); err != nil {
			return err
		}
	}
	if err = c.Auth(smtp.PlainAuth("", s.Username, s.Password, s.Host)); err != nil {
		return err
	}
	from, err := mail.ParseAddress(s.From)
	if err != nil {
		return err
	}
	if strings.ContainsAny(m.Subject+m.To+s.From, "\r\n") {
		return fmt.Errorf("invalid mail headers")
	}
	if err = c.Mail(from.Address); err != nil {
		return err
	}
	if err = c.Rcpt(m.To); err != nil {
		return err
	}
	w, err := c.Data()
	if err != nil {
		return err
	}
	encoded := base64.StdEncoding.EncodeToString([]byte(m.Body))
	var lines strings.Builder
	for len(encoded) > 76 {
		lines.WriteString(encoded[:76] + "\r\n")
		encoded = encoded[76:]
	}
	lines.WriteString(encoded + "\r\n")
	_, err = fmt.Fprintf(w, "From: %s\r\nTo: %s\r\nSubject: %s\r\nDate: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/html; charset=UTF-8\r\nContent-Transfer-Encoding: base64\r\n\r\n%s", from.String(), m.To, m.Subject, time.Now().UTC().Format(time.RFC1123Z), lines.String())
	if err != nil {
		_ = w.Close()
		return err
	}
	if err = w.Close(); err != nil {
		return err
	}
	// SMTP accepted the message; a subsequent QUIT failure must not trigger a duplicate send.
	_ = c.Quit()
	return nil
}

// Permanent SMTP rejection must not repeatedly contact the provider. Never expose
// server response text: it may include a recipient, credential or message body.
func classify(err error) error {
	var response *textproto.Error
	if errors.As(err, &response) && response.Code >= 500 && response.Code < 600 {
		return entity.ErrPermanentDelivery
	}
	return err
}
