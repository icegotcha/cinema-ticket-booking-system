package firebase

import (
	"context"
	"fmt"
	"sync"

	firebase "firebase.google.com/go/v4"
	firebaseauth "firebase.google.com/go/v4/auth"
)

var authClient struct {
	sync.Mutex
	client *firebaseauth.Client
}

// NewAuthClient reuses a successfully initialized client and its token key cache.
// Initialization failures are retried on subsequent requests.
func NewAuthClient(ctx context.Context) (*firebaseauth.Client, error) {
	authClient.Lock()
	defer authClient.Unlock()
	if authClient.client != nil {
		return authClient.client, nil
	}

	app, err := firebase.NewApp(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("initialize firebase app: %w", err)
	}

	client, err := app.Auth(ctx)
	if err != nil {
		return nil, fmt.Errorf("initialize firebase auth client: %w", err)
	}

	authClient.client = client
	return client, nil
}
