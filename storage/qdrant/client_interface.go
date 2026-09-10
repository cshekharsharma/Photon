package qdrant

import (
	"context"
)

// QdrantClientInterface is a stable interface consumed by applications.
// It intentionally hides Qdrant SDK types and raw HTTP details.
type QdrantClientInterface interface {
	Health(ctx context.Context) error

	UpsertPoints(ctx context.Context, req UpsertPointsRequest) error
	DeletePoints(ctx context.Context, req DeletePointsRequest) error
	Search(ctx context.Context, req SearchRequest) ([]ScoredPoint, error)

	// Collection management (minimal baseline)
	CreateCollection(ctx context.Context, req CreateCollectionRequest) error
	DeleteCollection(ctx context.Context, collection string) error
	GetCollectionInfo(ctx context.Context, collection string) (CollectionInfo, error)
}
