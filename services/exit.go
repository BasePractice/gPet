package services

import (
	"context"
	"os/signal"
	"syscall"
)

type ExitHandler func(context.Context)

func Handle(parent context.Context, handler ExitHandler) {
	ctx, cancel := signal.NotifyContext(parent, syscall.SIGINT, syscall.SIGTERM)
	defer cancel()
	<-ctx.Done()
	handler(ctx)
}
