package amqp

import (
	"context"
	"errors"
	api "github.com/mephistolie/chefbook-backend-mail/api/mq"
	"github.com/mephistolie/chefbook-backend-mail/internal/entity"
	"testing"
)

type handler struct{ calls int }

func (h *handler) Deliver(context.Context, api.SendRequest) error { h.calls++; return nil }
func TestMessageValidation(t *testing.T) {
	for _, body := range []string{`{`, `{"to":"a@example.org","unknown":true}`, `{} {}`, `{}null`} {
		h := &handler{}
		if process(context.Background(), api.MessageType, []byte(body), h) == nil || h.calls != 0 {
			t.Fatal("invalid message accepted")
		}
	}
}
func TestValidMessage(t *testing.T) {
	h := &handler{}
	if err := process(context.Background(), api.MessageType, []byte(`{"to":"a@example.org","template":"password_changed"}`), h); err != nil || h.calls != 1 {
		t.Fatal(err)
	}
}
func TestUnknownType(t *testing.T) {
	h := &handler{}
	if process(context.Background(), "other", []byte(`{}`), h) == nil || h.calls != 0 {
		t.Fatal("unknown type accepted")
	}
}

type failingHandler struct {
	calls int
	err   error
}

func (h *failingHandler) Deliver(context.Context, api.SendRequest) error { h.calls++; return h.err }
func TestBoundedTransientRetries(t *testing.T) {
	h := &failingHandler{err: errors.New("temporary")}
	err := processWithDelay(context.Background(), api.MessageType, []byte(`{}`), h, 0)
	if err == nil || h.calls != 3 {
		t.Fatalf("expected 3 attempts, got %d", h.calls)
	}
}
func TestPermanentFailureIsNotRetried(t *testing.T) {
	h := &failingHandler{err: entity.ErrPermanentDelivery}
	err := processWithDelay(context.Background(), api.MessageType, []byte(`{}`), h, 0)
	if err == nil || h.calls != 1 {
		t.Fatalf("permanent failure retried %d times", h.calls)
	}
}
func TestCancelledDeliveryNeverStarts(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	h := &failingHandler{err: errors.New("temporary")}
	err := processWithDelay(ctx, api.MessageType, []byte(`{}`), h, 0)
	if !errors.Is(err, context.Canceled) || h.calls != 0 {
		t.Fatal("cancelled message was sent")
	}
}

func TestOversizedPayload(t *testing.T) {
	h := &handler{}
	if processWithDelay(context.Background(), api.MessageType, make([]byte, 16385), h, 0) == nil || h.calls != 0 {
		t.Fatal("oversized message accepted")
	}
}
