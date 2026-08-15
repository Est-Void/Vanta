package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/Est-Void/Vanta/internal/daemon"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := daemon.Run(ctx); err != nil {
		log.Fatal(err)
	}
}
