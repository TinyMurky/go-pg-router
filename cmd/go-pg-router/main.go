// Package main will start go-pg-router
package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/TinyMurky/go-pg-router/internal/listener"
	"github.com/TinyMurky/go-pg-router/internal/pgproto"
)

func main() {
	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)

	defer func() {
		stop()
		if r := recover(); r != nil {
			slog.Error("go-pg-router panic", "panic", r)
		}
	}()

	err := realMain(ctx)

	stop()

	if err != nil {
		slog.Error("go-pg-router:", "error", err.Error())
	}

	slog.Info("go-pg-router: successful shutdown")
}

func realMain(ctx context.Context) error {

	address := "localhost:3000"

	l, err := net.Listen("tcp", address)

	if err != nil {
		return fmt.Errorf("listen to tcp at %q: %w", address, err)
	}

	// Close will unblock l.Accept
	defer l.Close()

	pgHandler := pgproto.NewPGHandler()
	tcpListener := listener.New(pgHandler)

	// tcpListener work in separated goroutine
	go tcpListener.Start(l)

	slog.Info("go-pg-router start listening at:", "address", address)
	// If context Closed, this will unblock, then defer listener will be triggered
	<-ctx.Done()
	cause := context.Cause(ctx)

	if cause == nil {
		cause = ctx.Err()
	}

	slog.Info("go-pg-router terminated", "reason", cause.Error())
	return nil
}
