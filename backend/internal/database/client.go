package database

import (
	"context"
	"fmt"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func NewClient() (*mongo.Database, error) {
	mongoURI := os.Getenv("MONGODB_URI")
	databaseName := os.Getenv("MONGODB_DATABASE")
	clientOptions := options.Client().ApplyURI(mongoURI)

	client, err := mongo.Connect(clientOptions)
	if err != nil {
		return nil, fmt.Errorf("connect to mongodb: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := client.Ping(ctx, nil); err != nil {
		_ = client.Disconnect(context.Background())

		return nil, fmt.Errorf("ping mongodb: %w", err)
	}

	return client.Database(databaseName), nil

}
