package main

import (
	"context"
	"errors"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/icegotcha/cinema-ticket-booking-system/backend/internal/database"
	"github.com/icegotcha/cinema-ticket-booking-system/backend/internal/queue"
	"github.com/icegotcha/cinema-ticket-booking-system/backend/internal/worker"
)

type Runner interface {
	Run(ctx context.Context) error
}

func main() {
	db, err := database.NewClient()
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	channel, err := queue.NewRabbitMQClient()
	if err != nil {
		log.Fatalf("failed to create RabbitMQ client: %v", err)
	}

	workers := []Runner{
		worker.NewBookingSuccessWorker(channel, db),
	}

	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	var wg sync.WaitGroup
	errCh := make(chan error, len(workers))

	for _, w := range workers {
		wg.Add(1)

		go func(worker Runner) {
			defer wg.Done()

			if err := worker.Run(ctx); err != nil &&
				!errors.Is(err, context.Canceled) {
				errCh <- err
			}
		}(w)
	}

	go func() {
		wg.Wait()
		close(errCh)
	}()

	for err := range errCh {
		log.Printf("worker stopped with error: %v", err)
		stop()
	}
}
