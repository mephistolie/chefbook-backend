package repository

import (
	"context"
	"github.com/mephistolie/chefbook-backend-mail/internal/entity"
)

type Sender interface {
	Send(context.Context, entity.Mail) error
}
