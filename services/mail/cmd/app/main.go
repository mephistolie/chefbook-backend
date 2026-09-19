package main

import (
	"context"
	"github.com/mephistolie/chefbook-backend-mail/internal/app"
	"github.com/mephistolie/chefbook-backend-mail/internal/logging"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()
	if app.Run(ctx) != nil {
		logging.Events{}.Stopped(ctx)
		os.Exit(1)
	}
}
