package contract

import (
	"context"
	"time"

	"github.com/cshekharsharma/photon/cloud/entity/objectstorage"
)

// ObjectStorageInterface interface is a common interface that has to be implemented
// by all cloud vendor specific implementions (ie AWS, Azure, GCP).
// This interface provides a cloud agnostic approach to the workflow, where consumers
// of object store cloud workflow do not have to know the exact cloud vendor, can work
// with the inteface signature itself.
type ObjectStorageInterface interface {

	// Get single object from cloud object store
	GetObject(ctx context.Context, bucket string, key string) (*objectstorage.GetObjectResult, error)

	// Get multiple objects (concurrently) from cloud object store
	GetMultipleObjects(ctx context.Context, bucket string, keys []string) ([]*objectstorage.GetObjectResult, error)

	// Put/upload single object to cloud object store
	PutObject(ctx context.Context, input *objectstorage.PutObjectInput) (*objectstorage.PutObjectResult, error)

	// Copy object from one location to another on cloud object store
	CopyObject(ctx context.Context, srcBucket string, srcObject string, destBucket string, destObject string) (*objectstorage.CopyObjectResult, error)

	// Get time-bound presigned url for any cloud object
	GetPresignedUrl(ctx context.Context, bucket string, key string, ttl time.Duration) (*objectstorage.GetPresignedUrlResult, error)

	// Get time-bound presigned url for direct object upload
	GetPresignedPutUrl(ctx context.Context, input *objectstorage.PresignedPutObjectInput) (*objectstorage.GetPresignedPutUrlResult, error)

	// Head object metadata without reading its body
	HeadObject(ctx context.Context, bucket string, key string) (*objectstorage.HeadObjectResult, error)

	// List all objects for provided prefix from cloud object store
	ListObject(ctx context.Context, bucket string, prefix string, maxObjects int32) (*objectstorage.ListObjectResult, error)
}
