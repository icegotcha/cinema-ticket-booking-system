package worker

import (
	"context"
	"errors"
	"log"
	"sync"
)

type Runner interface {
	Run(ctx context.Context) error
}

func RunAll(ctx context.Context, runners ...Runner) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var wg sync.WaitGroup
	workerErrors := make(chan error, len(runners))

	for _, runner := range runners {
		wg.Add(1)

		go func(runner Runner) {
			defer wg.Done()

			err := runner.Run(ctx)
			if err == nil || errors.Is(err, context.Canceled) {
				return
			}

			workerErrors <- err
		}(runner)
	}

	go func() {
		wg.Wait()
		close(workerErrors)
	}()

	for err := range workerErrors {
		log.Printf("worker stopped with error: %v", err)
		cancel()
	}
}
