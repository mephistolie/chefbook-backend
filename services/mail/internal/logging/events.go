package logging

import (
	"context"
	"log/slog"
)

type Events struct{}

func (Events) Delivery(ctx context.Context, outcome string) {
	slog.InfoContext(ctx, "mail.delivery", "outcome", outcome)
}
func (Events) Stopped(ctx context.Context) { slog.ErrorContext(ctx, "mail.runtime.stopped") }
