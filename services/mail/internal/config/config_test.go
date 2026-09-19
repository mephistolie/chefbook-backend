package config

import "testing"

func TestRequiredConfig(t *testing.T) {
	t.Setenv("ENVIRONMENT", "develop")
	t.Setenv("AMQP_URL", "amqp://localhost/dev")
	t.Setenv("SMTP_HOST", "smtp.example.org")
	t.Setenv("SMTP_PORT", "465")
	t.Setenv("SMTP_USERNAME", "smtp-user")
	t.Setenv("SMTP_PASSWORD", "placeholder")
	t.Setenv("SMTP_EMAIL", "")
	c, err := Load()
	if err != nil || !c.Development || c.From != "noreply@chefbook.io" {
		t.Fatal("valid configuration rejected")
	}
	t.Setenv("ENVIRONMENT", "")
	if _, err = Load(); err == nil {
		t.Fatal("missing environment silently defaulted")
	}
	t.Setenv("ENVIRONMENT", "production")
	t.Setenv("SMTP_PASSWORD", "")
	if _, err = Load(); err == nil {
		t.Fatal("missing credentials accepted")
	}
}
