package database

import (
	"context"
	"errors"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type Repository[T any] struct {
	collection *mongo.Collection
}

func NewRepository[T any](
	db *mongo.Database,
	collectionName string,
) *Repository[T] {
	return &Repository[T]{
		collection: db.Collection(collectionName),
	}
}

func (r *Repository[T]) Collection() *mongo.Collection {
	return r.collection
}

func (r *Repository[T]) Create(
	ctx context.Context,
	document *T,
) (bson.ObjectID, error) {
	result, err := r.collection.InsertOne(ctx, document)
	if err != nil {
		return bson.NilObjectID, fmt.Errorf("insert document: %w", err)
	}

	id, ok := result.InsertedID.(bson.ObjectID)
	if !ok {
		return bson.NilObjectID, errors.New(
			"inserted id is not bson.ObjectID",
		)
	}

	return id, nil
}

func (r *Repository[T]) FindByID(
	ctx context.Context,
	id bson.ObjectID,
) (*T, error) {
	return r.FindOne(ctx, bson.M{
		"_id": id,
	})
}

func (r *Repository[T]) FindOne(
	ctx context.Context,
	filter any,
	opts ...options.Lister[options.FindOneOptions],
) (*T, error) {
	var document T

	err := r.collection.
		FindOne(ctx, filter, opts...).
		Decode(&document)

	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}

		return nil, fmt.Errorf("find one document: %w", err)
	}

	return &document, nil
}

func (r *Repository[T]) FindMany(
	ctx context.Context,
	filter any,
	opts ...options.Lister[options.FindOptions],
) ([]T, error) {
	cursor, err := r.collection.Find(ctx, filter, opts...)
	if err != nil {
		return nil, fmt.Errorf("find documents: %w", err)
	}
	defer cursor.Close(ctx)

	documents := make([]T, 0)

	if err := cursor.All(ctx, &documents); err != nil {
		return nil, fmt.Errorf("decode documents: %w", err)
	}

	return documents, nil
}

func (r *Repository[T]) FindAll(
	ctx context.Context,
	opts ...options.Lister[options.FindOptions],
) ([]T, error) {
	return r.FindMany(ctx, bson.M{}, opts...)
}

func (r *Repository[T]) Count(
	ctx context.Context,
	filter any,
	opts ...options.Lister[options.CountOptions],
) (int64, error) {
	count, err := r.collection.CountDocuments(
		ctx,
		filter,
		opts...,
	)
	if err != nil {
		return 0, fmt.Errorf("count documents: %w", err)
	}

	return count, nil
}

func (r *Repository[T]) Exists(
	ctx context.Context,
	filter any,
) (bool, error) {
	count, err := r.collection.CountDocuments(
		ctx,
		filter,
		options.Count().SetLimit(1),
	)
	if err != nil {
		return false, fmt.Errorf(
			"check document existence: %w",
			err,
		)
	}

	return count > 0, nil
}

func (r *Repository[T]) UpdateByID(
	ctx context.Context,
	id bson.ObjectID,
	update any,
	opts ...options.Lister[options.UpdateOneOptions],
) error {
	result, err := r.collection.UpdateByID(
		ctx,
		id,
		update,
		opts...,
	)
	if err != nil {
		return fmt.Errorf("update document: %w", err)
	}

	if result.MatchedCount == 0 {
		return mongo.ErrNoDocuments
	}

	return nil
}

func (r *Repository[T]) UpdateOne(
	ctx context.Context,
	filter any,
	update any,
	opts ...options.Lister[options.UpdateOneOptions],
) error {
	result, err := r.collection.UpdateOne(
		ctx,
		filter,
		update,
		opts...,
	)
	if err != nil {
		return fmt.Errorf("update document: %w", err)
	}

	if result.MatchedCount == 0 {
		return mongo.ErrNoDocuments
	}

	return nil
}

func (r *Repository[T]) ReplaceByID(
	ctx context.Context,
	id bson.ObjectID,
	document *T,
	opts ...options.Lister[options.ReplaceOptions],
) error {
	result, err := r.collection.ReplaceOne(
		ctx,
		bson.M{"_id": id},
		document,
		opts...,
	)
	if err != nil {
		return fmt.Errorf("replace document: %w", err)
	}

	if result.MatchedCount == 0 {
		return mongo.ErrNoDocuments
	}

	return nil
}

func (r *Repository[T]) DeleteByID(
	ctx context.Context,
	id bson.ObjectID,
) error {
	return r.DeleteOne(ctx, bson.M{
		"_id": id,
	})
}

func (r *Repository[T]) DeleteOne(
	ctx context.Context,
	filter any,
	opts ...options.Lister[options.DeleteOneOptions],
) error {
	result, err := r.collection.DeleteOne(
		ctx,
		filter,
		opts...,
	)
	if err != nil {
		return fmt.Errorf("delete document: %w", err)
	}

	if result.DeletedCount == 0 {
		return mongo.ErrNoDocuments
	}

	return nil
}
