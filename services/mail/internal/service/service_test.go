package service

import (
	"context"
	"errors"
	api "github.com/mephistolie/chefbook-backend-mail/api/mq"
	"github.com/mephistolie/chefbook-backend-mail/internal/entity"
	"strings"
	"testing"
	"time"
)

type sender struct {
	messages []entity.Mail
	err      error
}

func (s *sender) Send(ctx context.Context, m entity.Mail) error {
	s.messages = append(s.messages, m)
	return s.err
}
func TestCodeDeliveryAndExpiration(t *testing.T) {
	now := time.Now()
	expiry := now.Add(time.Minute)
	out := &sender{}
	svc := Service{Sender: out, Development: true, Now: func() time.Time { return now }}
	r := api.SendRequest{To: "a.b@example.org", Template: api.RegistrationCode, Code: "<123456>", ExpirationTimestamp: &expiry}
	if err := svc.Deliver(context.Background(), r); err != nil {
		t.Fatal(err)
	}
	if len(out.messages) != 1 || !strings.HasPrefix(out.messages[0].Subject, "[DEV] ") || !strings.Contains(out.messages[0].Body, "&lt;123456&gt;") {
		t.Fatal("mail content/prefix incorrect")
	}
	expiry = now
	if err := svc.Deliver(context.Background(), r); err != nil || len(out.messages) != 1 {
		t.Fatal("expired verification was sent")
	}
}
func TestInvalidRequestsNeverSend(t *testing.T) {
	expiry := time.Now().Add(time.Minute)
	for _, r := range []api.SendRequest{
		{To: "a@example.org\r\nBcc: b@example.org", Template: api.PasswordChanged},
		{To: "a@example.org", Template: "unknown"},
		{To: "a@example.org", Template: api.RegistrationCode, Code: "123456"},
		{To: "a@example.org", Template: api.PasswordReset, URL: "javascript:alert(1)", ExpirationTimestamp: &expiry},
	} {
		s := &sender{}
		err := (Service{Sender: s}).Deliver(context.Background(), r)
		if !errors.Is(err, ErrInvalid) || len(s.messages) != 0 {
			t.Fatal("unsafe mail accepted")
		}
	}
}
func TestFailurePropagates(t *testing.T) {
	expected := errors.New("smtp failure")
	s := &sender{err: expected}
	err := (Service{Sender: s}).Deliver(context.Background(), api.SendRequest{To: "a@example.org", Template: api.PasswordChanged})
	if !errors.Is(err, expected) {
		t.Fatal(err)
	}
}
