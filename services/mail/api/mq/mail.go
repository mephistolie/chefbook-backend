package mq

import "time"

const Exchange = "mail"
const RoutingKey = "mail.send"
const MessageType = "mail.send.v1"
const (
	RegistrationCode         = "registration_code"
	PasswordReset            = "password_reset"
	EmailChangePrevious      = "email_change_previous"
	EmailChangeNew           = "email_change_new"
	PasswordChanged          = "password_changed"
	UsernameChanged          = "username_changed"
	AccountDeletionRequested = "account_deletion_requested"
	AccountDeleted           = "account_deleted"
	NewLogin                 = "new_login"
)

// SendRequest is private broker data. Never log its body or include it in errors.
// ExpirationTimestamp prevents delivering stale verification material after queue delays.
type SendRequest struct {
	To                  string     `json:"to"`
	Template            string     `json:"template"`
	Code                string     `json:"code,omitempty"`
	URL                 string     `json:"url,omitempty"`
	Username            string     `json:"username,omitempty"`
	Timestamp           *time.Time `json:"timestamp,omitempty"`
	DeleteSharedData    bool       `json:"deleteSharedData,omitempty"`
	Client              string     `json:"client,omitempty"`
	ExpirationTimestamp *time.Time `json:"expirationTimestamp,omitempty"`
}
