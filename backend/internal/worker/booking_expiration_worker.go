package worker

import (
	"context"
	"fmt"
	"log"
	"time"
)

type BookingExpirer interface {
	ExpireBookings(context.Context) error
}

type BookingExpirationWorker struct {
	service  BookingExpirer
	interval time.Duration
}

func NewBookingExpirationWorker(service BookingExpirer, interval time.Duration) *BookingExpirationWorker {
	return &BookingExpirationWorker{service: service, interval: interval}
}

func (w *BookingExpirationWorker) Run(ctx context.Context) error {
	if w.interval <= 0 {
		return fmt.Errorf("booking expiration interval must be positive")
	}
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()
	for {
		if ctx.Err() != nil {
			return nil
		}
		sweepCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		err := w.service.ExpireBookings(sweepCtx)
		cancel()
		if err != nil && ctx.Err() == nil {
			log.Printf("expire seat holds: %v", err)
		}
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
		}
	}
}
