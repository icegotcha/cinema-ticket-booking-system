package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/icegotcha/cinema-ticket-booking-system/backend/internal/database"
	"github.com/icegotcha/cinema-ticket-booking-system/backend/internal/queue"
	"github.com/icegotcha/cinema-ticket-booking-system/backend/internal/worker"
)

func main() {
	db, err := database.NewClient()
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	auditChannel, err := queue.NewRabbitMQClient()
	if err != nil {
		log.Fatalf("failed to create RabbitMQ client: %v", err)
	}

	defer auditChannel.Close()

	expirationWorker, cleanupExpiration, err := worker.SetupExpirationWorker(db)
	if err != nil {
		log.Fatalf("initialize expiration worker: %v", err)
	}
	defer cleanupExpiration()

	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	auditWorker := worker.NewBookingAuditWorker(auditChannel.Channel, db)
	worker.RunAll(ctx, auditWorker, expirationWorker)
}
