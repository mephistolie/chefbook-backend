package smtp

import (
	"errors"
	"fmt"
	"github.com/mephistolie/chefbook-backend-mail/internal/entity"
	"net/textproto"
	"testing"
)

func TestSMTPFailureClassification(t *testing.T) {
	for _, code := range []int{500, 535, 550, 554} {
		err := classify(fmt.Errorf("wrapped: %w", &textproto.Error{Code: code, Msg: "secret@example.org"}))
		if !errors.Is(err, entity.ErrPermanentDelivery) || err.Error() != "permanent mail delivery failure" {
			t.Fatal("permanent rejection not sanitized")
		}
	}
	transient := &textproto.Error{Code: 421, Msg: "retry later"}
	if classify(transient) != transient {
		t.Fatal("temporary SMTP failure not retryable")
	}
	if classify(nil) != nil {
		t.Fatal("success changed")
	}
}
