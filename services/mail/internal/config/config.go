package config

import (
	"errors"
	"net/mail"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	AMQPURL, Host, Username, Password, From string
	Port                                    int
	Development                             bool
	Timeout                                 time.Duration
}

func Load() (Config, error) {
	c := Config{AMQPURL: os.Getenv("AMQP_URL"), Host: os.Getenv("SMTP_HOST"), Username: os.Getenv("SMTP_USERNAME"), Password: os.Getenv("SMTP_PASSWORD"), From: os.Getenv("SMTP_EMAIL"), Port: 465, Timeout: 30 * time.Second}
	env := os.Getenv("ENVIRONMENT")
	if env != "production" && env != "develop" {
		return c, errors.New("ENVIRONMENT must be production or develop")
	}
	c.Development = env == "develop"
	if c.From == "" {
		c.From = "noreply@chefbook.io"
	}
	if p := os.Getenv("SMTP_PORT"); p != "" {
		n, e := strconv.Atoi(p)
		if e != nil {
			return c, errors.New("invalid SMTP_PORT")
		}
		c.Port = n
	}
	_, err := mail.ParseAddress(c.From)
	if err != nil || strings.ContainsAny(c.From, "\r\n") || c.AMQPURL == "" || c.Host == "" || c.Username == "" || c.Password == "" || c.Port < 1 || c.Port > 65535 {
		return c, errors.New("incomplete mail configuration")
	}
	return c, nil
}
