package app

import (
	"context"
	"github.com/mephistolie/chefbook-backend-mail/internal/config"
	"github.com/mephistolie/chefbook-backend-mail/internal/repository/smtp"
	"github.com/mephistolie/chefbook-backend-mail/internal/service"
	"github.com/mephistolie/chefbook-backend-mail/internal/transport/amqp"
)

func Run(ctx context.Context) error {
	c, err := config.Load()
	if err != nil {
		return err
	}
	sender := smtp.Sender{Host: c.Host, Port: c.Port, Username: c.Username, Password: c.Password, From: c.From, Timeout: c.Timeout}
	return amqp.Run(ctx, c.AMQPURL, service.Service{Sender: sender, Development: c.Development})
}
