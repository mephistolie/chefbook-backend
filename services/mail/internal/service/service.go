package service

import (
	"context"
	"errors"
	api "github.com/mephistolie/chefbook-backend-mail/api/mq"
	"github.com/mephistolie/chefbook-backend-mail/internal/entity"
	"github.com/mephistolie/chefbook-backend-mail/internal/service/dependencies/repository"
	"html"
	"net/mail"
	"net/url"
	"strings"
	"time"
)

var ErrInvalid = errors.New("invalid mail request")

type Service struct {
	Sender      repository.Sender
	Development bool
	Now         func() time.Time
}

func (s Service) Deliver(ctx context.Context, r api.SendRequest) error {
	now := time.Now()
	if s.Now != nil {
		now = s.Now()
	}
	m, err := render(r, s.Development)
	if err != nil {
		return err
	}
	if r.ExpirationTimestamp != nil && !now.Before(*r.ExpirationTimestamp) {
		return nil
	}
	return s.Sender.Send(ctx, m)
}
func render(r api.SendRequest, dev bool) (entity.Mail, error) {
	a, err := mail.ParseAddress(r.To)
	if err != nil || a.Address != r.To || strings.ContainsAny(r.To, "\r\n") {
		return entity.Mail{}, ErrInvalid
	}
	title, body := "", ""
	switch r.Template {
	case api.RegistrationCode:
		if r.Code == "" || r.ExpirationTimestamp == nil {
			return entity.Mail{}, ErrInvalid
		}
		title = "Confirm your email"
		body = "Your ChefBook confirmation code: <strong>" + html.EscapeString(r.Code) + "</strong>. Enter it in the app where you started registration."
	case api.PasswordReset, api.EmailChangePrevious, api.EmailChangeNew:
		u, e := url.Parse(r.URL)
		if e != nil || u.Scheme != "https" || u.Host == "" || u.User != nil || r.ExpirationTimestamp == nil {
			return entity.Mail{}, ErrInvalid
		}
		title = map[string]string{api.PasswordReset: "Reset your password", api.EmailChangePrevious: "Approve email change", api.EmailChangeNew: "Confirm your new email"}[r.Template]
		body = "<a href=\"" + html.EscapeString(r.URL) + "\">" + title + "</a>"
	case api.PasswordChanged:
		title = "Password changed"
		body = "Your ChefBook password was changed."
	case api.UsernameChanged:
		title = "Username changed"
		body = "Your ChefBook username is now " + html.EscapeString(r.Username) + "."
	case api.AccountDeletionRequested:
		if r.Timestamp == nil {
			return entity.Mail{}, ErrInvalid
		}
		title = "Account deletion requested"
		body = "Your account is scheduled for deletion at " + r.Timestamp.UTC().Format(time.RFC3339) + "."
		if r.DeleteSharedData {
			body += " Shared recipes will also be deleted."
		}
	case api.AccountDeleted:
		title = "Account deleted"
		body = "Your ChefBook account was deleted."
	case api.NewLogin:
		title = "New login"
		body = "A new session was created for your ChefBook account. " + html.EscapeString(r.Client)
	default:
		return entity.Mail{}, ErrInvalid
	}
	title = "ChefBook: " + title
	if dev {
		title = "[DEV] " + title
	}
	return entity.Mail{To: r.To, Subject: title, Body: "<!doctype html><html><body><p>" + body + "</p></body></html>"}, nil
}
