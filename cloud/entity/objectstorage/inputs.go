// Package objectstorage provides structures for interacting with an S3-compatible storage system.
package objectstorage

import (
	"io"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// PutObjectInput contains the parameters for uploading an object to a storage bucket.
// It includes the bucket's name, a reader interface to read the content from,
// the target path within the bucket, access control settings, and the MIME type of the object.
type PutObjectInput struct {
	Bucket       string                // Bucket specifies the name of the bucket where the object will be stored.
	Source       io.Reader             // Source is the reader interface that provides the object's content.
	TargetPath   string                // TargetPath specifies the full path where the object should be stored within the bucket.
	Acl          types.ObjectCannedACL // Acl defines the canned access control list setting for the uploaded object.
	MimeType     string                // MimeType specifies the MIME type of the object being uploaded.
	CacheControl string                // CacheControl specifies the HTTP Cache-Control header to persist with the object.
	DisableACL   bool                  // DisableACL omits the ACL field entirely for buckets that do not allow object ACLs.
}

// PresignedPutObjectInput contains the parameters for creating a time-bound
// direct-upload URL.
type PresignedPutObjectInput struct {
	Bucket               string        // Bucket specifies the target bucket.
	Key                  string        // Key specifies the object key to upload.
	ContentType          string        // ContentType is the required MIME type for the upload.
	TTL                  time.Duration // TTL controls how long the URL remains valid.
	ServerSideEncryption string        // ServerSideEncryption requests bucket-side encryption such as AES256.
}
