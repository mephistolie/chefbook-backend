package amqp

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	api "github.com/mephistolie/chefbook-backend-mail/api/mq"
	"github.com/mephistolie/chefbook-backend-mail/internal/entity"
	"github.com/mephistolie/chefbook-backend-mail/internal/logging"
	"github.com/mephistolie/chefbook-backend-mail/internal/service"
	amqp "github.com/rabbitmq/amqp091-go"
	"io"
	"time"
)

type Handler interface {
	Deliver(context.Context, api.SendRequest) error
}

func Run(ctx context.Context, url string, handler Handler) error {
	conn, err := amqp.Dial(url)
	if err != nil {
		return err
	}
	defer conn.Close()
	ch, err := conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()
	if err = ch.ExchangeDeclare(api.Exchange, "direct", true, false, false, false, nil); err != nil {
		return err
	}
	if _, err = ch.QueueDeclare("mail.dead", true, false, false, false, nil); err != nil {
		return err
	}
	if _, err = ch.QueueDeclare("mail.send", true, false, false, false, amqp.Table{"x-dead-letter-exchange": "", "x-dead-letter-routing-key": "mail.dead"}); err != nil {
		return err
	}
	if err = ch.QueueBind("mail.send", api.RoutingKey, api.Exchange, false, nil); err != nil {
		return err
	}
	if err = ch.Qos(1, 0, false); err != nil {
		return err
	}
	deliveries, err := ch.Consume("mail.send", "mail-service", false, false, false, false, nil)
	if err != nil {
		return err
	}
	for {
		select {
		case <-ctx.Done():
			return nil
		case d, ok := <-deliveries:
			if !ok {
				return errors.New("mail consumer disconnected")
			}
			if ctx.Err() != nil {
				_ = d.Nack(false, true)
				return nil
			}
			err = process(ctx, d.Type, d.Body, handler)
			if ctx.Err() != nil {
				_ = d.Nack(false, true)
				return nil
			}
			if err != nil {
				logging.Events{}.Delivery(ctx, "dead_lettered")
				if err = d.Reject(false); err != nil {
					return err
				}
			} else {
				logging.Events{}.Delivery(ctx, "accepted")
				if err = d.Ack(false); err != nil {
					return err
				}
			}
		}
	}
}
func process(ctx context.Context, kind string, body []byte, h Handler) error {
	return processWithDelay(ctx, kind, body, h, 5*time.Second)
}
func processWithDelay(ctx context.Context, kind string, body []byte, h Handler, delay time.Duration) error {
	if kind != api.MessageType || len(body) > 16384 {
		return service.ErrInvalid
	}
	var request api.SendRequest
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&request); err != nil {
		return service.ErrInvalid
	}
	if decoder.Decode(new(any)) != io.EOF {
		return service.ErrInvalid
	}
	var err error
	for i := 0; i < 3; i++ {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		attempt, cancel := context.WithTimeout(ctx, 35*time.Second)
		err = h.Deliver(attempt, request)
		cancel()
		if err == nil || errors.Is(err, service.ErrInvalid) || errors.Is(err, entity.ErrPermanentDelivery) {
			return err
		}
		if i < 2 {
			timer := time.NewTimer(time.Duration(i+1) * delay)
			select {
			case <-ctx.Done():
				timer.Stop()
				return ctx.Err()
			case <-timer.C:
			}
		}
	}
	return err
}
