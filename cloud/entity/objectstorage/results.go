// Package objectstorage provides structures and methods for managing objects in a cloud storage system,
// such as AWS S3.
package objectstorage

import (
	"io"
	"time"
)

// GetObjectResult contains the result of a request to retrieve an object from storage.
// It provides access to the object's content and metadata such as content type, size, and modification time.
type GetObjectResult struct {
	ContentBody   io.ReadCloser // Stream for reading the object's content.
	ContentLength int64         // Length of the object content in bytes.
	ContentType   string        // MIME type of the object.
	ETag          string        // ETag is an identifier for a specific version of the object.
	LastModified  time.Time     // Timestamp of when the object was last modified.
}

// PutObjectResult encapsulates the outcome of an object upload operation.
// It includes status of the upload and identifiers such as ETag and object key.
type PutObjectResult struct {
	IsUploaded bool   // Indicates whether the upload was successful.
	ETag       string // ETag is an identifier for a specific version of the uploaded object.
	ObjectKey  string // ObjectKey is the unique identifier for the object in the storage.
}

// CopyObjectResult represents the result of a copy operation within the storage.
// It includes the success status and the key of the copied object.
type CopyObjectResult struct {
	IsCopied  bool   // Indicates whether the object was successfully copied.
	ObjectKey string // ObjectKey is the unique identifier for the copied object.
}

// GetPresignedUrlResult contains a presigned URL for accessing an object.
// It includes the URL, its validity duration, and the last modification time of the object.
type GetPresignedUrlResult struct {
	TTL          time.Duration // TTL indicates how long the presigned URL is valid.
	PresignedUrl string        // PresignedUrl is the URL for direct object access.
	LastModified time.Time     // Timestamp of when the object was last modified.
}

// GetPresignedPutUrlResult contains a presigned URL for direct object upload.
type GetPresignedPutUrlResult struct {
	TTL          time.Duration
	PresignedUrl string
	CreatedAt    time.Time
}

// HeadObjectResult contains metadata for an object without its body.
type HeadObjectResult struct {
	ContentLength int64
	ContentType   string
	ETag          string
	LastModified  time.Time
}

// S3Object represents a single object within S3 or similar storage systems.
// It includes essential details such as the key, size, and last modification time of the object.
type S3Object struct {
	Key          string    // Key is the unique identifier for the object within the storage.
	LastModified time.Time // Timestamp of when the object was last modified.
	Size         int64     // Size of the object in bytes.
}

// ListObjectResult contains the results of a listing operation in the storage.
// It holds a collection of objects and the total count of objects listed.
type ListObjectResult struct {
	Objects     []S3Object // Objects is a list of objects retrieved from the storage.
	ObjectCount int        // ObjectCount is the total number of objects in the list.
}

// AppendOneObject adds a single S3Object to the Objects slice of a ListObjectResult and returns the modified result.
// This method allows for incremental addition of objects to the result set.
func (r *ListObjectResult) AppendOneObject(contentBody S3Object) *ListObjectResult {
	r.Objects = append(r.Objects, contentBody)
	return r
}
