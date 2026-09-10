package aws

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/cshekharsharma/photon/cloud/entity/objectstorage"
	"github.com/cshekharsharma/photon/utils/types"

	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	awstypes "github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// awsS3ClientInterface defines an interface for interacting with AWS Simple Storage Service (S3).
// This interface abstracts operations for retrieving, storing, copying, and listing objects within an S3 bucket,
// allowing for more flexible integration and testing of S3 operations.
type awsS3ClientInterface interface {
	GetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error)
	HeadObject(ctx context.Context, params *s3.HeadObjectInput, optFns ...func(*s3.Options)) (*s3.HeadObjectOutput, error)
	PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error)
	CopyObject(ctx context.Context, params *s3.CopyObjectInput, optFns ...func(*s3.Options)) (*s3.CopyObjectOutput, error)
	ListObjectsV2(ctx context.Context, params *s3.ListObjectsV2Input, optFns ...func(*s3.Options)) (*s3.ListObjectsV2Output, error)
}

// awsS3PresignerInterface defines an interface for creating pre-signed URLs for AWS S3 objects.
// This interface allows generating URLs that clients can use to access S3 objects directly,
// with a specified expiration time and without further authorization.
type awsS3PresignerInterface interface {
	PresignGetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.PresignOptions)) (*v4.PresignedHTTPRequest, error)
	PresignPutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.PresignOptions)) (*v4.PresignedHTTPRequest, error)
}

// ObjectStorage provides methods for managing objects on AWS S3.
// It holds a reference to the AWS S3 client to interact with the S3 service.
type ObjectStorage struct {
	S3Client          awsS3ClientInterface
	S3PresignerClient awsS3PresignerInterface
}

var createChunks = types.CreateChunks

// Constants used for determining limits in S3 operations.
const MAX_OBJECT_PROCESSING_LIMIT int32 = 1000
const MAX_GET_OBJECT_BATCH_LIMIT int = 50

// GetObject retrieves a specific object from the provided AWS S3 bucket using the given key.
// It makes use of the AWS SDK's GetObject function and maps the returned AWS object attributes
// to a custom GetObjectResult structure.
//
// Parameters:
//   - bucket: The name of the S3 bucket where the object is stored.
//   - key: The unique identifier for the object within the bucket.
//
// Returns:
//   - A pointer to a GetObjectResult struct containing details of the retrieved object.
//   - An error object if any issues are encountered during the retrieval. Otherwise, nil.
func (os ObjectStorage) GetObject(ctx context.Context, bucket string, key string) (*objectstorage.GetObjectResult, error) {
	awsResult, err := os.S3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})

	if err != nil {
		return nil, err
	}

	result := new(objectstorage.GetObjectResult)

	result.ContentBody = awsResult.Body
	result.ContentLength = *awsResult.ContentLength
	result.ContentType = *awsResult.ContentType
	result.ETag = *awsResult.ETag
	result.LastModified = *awsResult.LastModified

	return result, nil
}

// GetMultipleObjects retrieves multiple objects from the specified AWS S3 bucket using a list of keys.
// The objects are fetched concurrently in batches, determined by the MAX_GET_OBJECT_BATCH_LIMIT constant.
//
// Parameters:
//   - bucket: The name of the S3 bucket from which the objects are to be retrieved.
//   - keys: A slice of string keys, each representing a unique identifier for an object within the bucket.
//
// Returns:
//   - A slice of pointers to GetObjectResult structures, each containing details of a retrieved object.
//   - An error object if issues are encountered during the retrieval or while batching the keys. Otherwise, nil.
func (os ObjectStorage) GetMultipleObjects(ctx context.Context, bucket string, keys []string) ([]*objectstorage.GetObjectResult, error) {
	var results = []*objectstorage.GetObjectResult{}
	batches, err := createChunks(types.ToInterfaceSlice(keys), MAX_GET_OBJECT_BATCH_LIMIT)

	if err != nil {
		return nil, err
	}

	var wg sync.WaitGroup
	channel := make(chan *objectstorage.GetObjectResult)

	for _, batch := range batches {
		for _, key := range batch {
			wg.Add(1)
			go func(bucket, key string) {
				defer wg.Done()
				result, err := os.GetObject(ctx, bucket, key)
				if err != nil {
					return
				}
				channel <- result
			}(bucket, fmt.Sprintf("%s", key))
		}
	}

	go func() {
		wg.Wait()
		close(channel)
	}()

	for item := range channel {
		results = append(results, item)
	}

	return results, nil
}

// PutObject uploads an object to the specified AWS S3 bucket.
//
// Parameters:
//   - input: A pointer to PutObjectInput structure containing details of the object to be uploaded such as
//     the bucket name, object key, access control list (ACL), content type, and the object's data.
//
// Behavior:
//   - If the provided ACL is empty, the function sets the default ACL to "private".
//
// Returns:
//   - A pointer to PutObjectResult structure, containing details of the uploaded object, like its ETag and object key.
//   - An error object if issues are encountered during the upload. Otherwise, nil.
func (os ObjectStorage) PutObject(ctx context.Context, input *objectstorage.PutObjectInput) (*objectstorage.PutObjectResult, error) {
	putInput := &s3.PutObjectInput{
		Bucket:      aws.String(input.Bucket),
		Key:         aws.String(input.TargetPath),
		ContentType: aws.String(input.MimeType),
		Body:        input.Source,
	}
	if input.CacheControl != "" {
		putInput.CacheControl = aws.String(input.CacheControl)
	}
	if !input.DisableACL {
		acl := input.Acl
		if acl == "" {
			acl = awstypes.ObjectCannedACLPrivate
		}
		putInput.ACL = acl
	}

	awsResult, err := os.S3Client.PutObject(ctx, putInput)

	if err != nil {
		return nil, err
	}

	result := new(objectstorage.PutObjectResult)
	result.IsUploaded = true
	result.ETag = *awsResult.ETag
	result.ObjectKey = input.TargetPath

	return result, nil
}

// CopyObject copies an object from a source bucket to a destination bucket on AWS S3.
//
// Parameters:
//   - srcBucket: The name of the source bucket where the object currently resides.
//   - srcObject: The key (path) of the object within the source bucket.
//   - destBucket: The name of the destination bucket where the object should be copied to.
//   - destObject: The desired key (path) of the object within the destination bucket.
//
// Returns:
//   - A pointer to CopyObjectResult structure, which contains information about the copied object.
//   - An error object if issues are encountered during the copy operation. Otherwise, nil.
func (os ObjectStorage) CopyObject(ctx context.Context, srcBucket string, srcObject string, destBucket string, destObject string) (*objectstorage.CopyObjectResult, error) {
	_, err := os.S3Client.CopyObject(ctx, &s3.CopyObjectInput{
		Bucket:     aws.String(destBucket),
		CopySource: aws.String(srcBucket + "/" + srcObject),
		Key:        aws.String(destBucket + "/" + destObject),
	})

	if err != nil {
		return nil, err
	}

	result := new(objectstorage.CopyObjectResult)
	result.IsCopied = true
	result.ObjectKey = destObject

	return result, nil
}

// GetPresignedUrl generates a presigned URL for an S3 object, allowing temporary access to the object.
// This method creates a URL that can be used to retrieve the content of the specified S3 object.
//
// Parameters:
//   - bucket: The name of the S3 bucket in which the object is stored.
//   - key: The object's key (its unique identifier within the bucket).
//   - ttl: Duration after which the presigned URL will expire.
//
// Returns:
//   - A pointer to GetPresignedUrlResult structure, which contains:
//   - The generated presigned URL.
//   - The time-to-live (TTL) duration of the URL.
//   - The time at which the URL was generated.
//   - An error object if there were issues generating the presigned URL; nil otherwise.
func (os ObjectStorage) GetPresignedUrl(ctx context.Context, bucket string, key string, ttl time.Duration) (*objectstorage.GetPresignedUrlResult, error) {
	currentTime := time.Now()

	presignedUrl, err := os.S3PresignerClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(ttl))

	if err != nil {
		return nil, err
	}

	result := new(objectstorage.GetPresignedUrlResult)
	result.PresignedUrl = presignedUrl.URL
	result.TTL = ttl
	result.LastModified = currentTime

	return result, nil
}

// GetPresignedPutUrl generates a presigned URL for direct S3 upload.
func (os ObjectStorage) GetPresignedPutUrl(ctx context.Context, input *objectstorage.PresignedPutObjectInput) (*objectstorage.GetPresignedPutUrlResult, error) {
	currentTime := time.Now()
	if input == nil {
		return nil, fmt.Errorf("presigned put input is required")
	}
	ttl := input.TTL
	if ttl <= 0 {
		ttl = 5 * time.Minute
	}
	putInput := &s3.PutObjectInput{
		Bucket:      aws.String(input.Bucket),
		Key:         aws.String(input.Key),
		ContentType: aws.String(input.ContentType),
	}
	if input.ServerSideEncryption != "" {
		putInput.ServerSideEncryption = awstypes.ServerSideEncryption(input.ServerSideEncryption)
	}
	presignedUrl, err := os.S3PresignerClient.PresignPutObject(ctx, putInput, s3.WithPresignExpires(ttl))
	if err != nil {
		return nil, err
	}
	return &objectstorage.GetPresignedPutUrlResult{
		TTL:          ttl,
		PresignedUrl: presignedUrl.URL,
		CreatedAt:    currentTime,
	}, nil
}

// HeadObject retrieves object metadata without reading the body.
func (os ObjectStorage) HeadObject(ctx context.Context, bucket string, key string) (*objectstorage.HeadObjectResult, error) {
	awsResult, err := os.S3Client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, err
	}
	result := &objectstorage.HeadObjectResult{}
	if awsResult.ContentLength != nil {
		result.ContentLength = *awsResult.ContentLength
	}
	if awsResult.ContentType != nil {
		result.ContentType = *awsResult.ContentType
	}
	if awsResult.ETag != nil {
		result.ETag = *awsResult.ETag
	}
	if awsResult.LastModified != nil {
		result.LastModified = *awsResult.LastModified
	}
	return result, nil
}

// ListObject retrieves a list of objects from a specified AWS S3 bucket
// that matches a given prefix. The method fetches objects in batches to
// handle large object lists efficiently. It uses AWS S3's continuation
// tokens for paginated results, ensuring that all relevant objects (up to
// the specified maxObjects limit) are retrieved.
//
// Parameters:
//   - bucket: The name of the S3 bucket to list objects from.
//   - prefix: The prefix string which the returned object keys must start with.
//   - maxObjects: The maximum number of objects to return. If there are more
//     objects in the bucket that match the given prefix, only
//     the first 'maxObjects' will be returned.
//
// Returns:
//   - A pointer to a ListObjectResult struct containing the retrieved objects.
//   - An error object if any errors are encountered during the fetch. Otherwise, nil.
func (os ObjectStorage) ListObject(ctx context.Context, bucket string, prefix string, maxObjects int32) (*objectstorage.ListObjectResult, error) {
	var nextContinuationToken *string
	var maxKeys int32

	objects := new(objectstorage.ListObjectResult)

	objectCount := 0
	batchSize := maxObjects

	params := &s3.ListObjectsV2Input{
		Bucket: aws.String(bucket),
		Prefix: aws.String(prefix),
	}

	for batchSize > 0 {
		if batchSize >= MAX_OBJECT_PROCESSING_LIMIT {
			maxKeys = MAX_OBJECT_PROCESSING_LIMIT
			batchSize -= maxKeys
		} else {
			maxKeys = batchSize
			batchSize = 0
		}

		params.MaxKeys = aws.Int32(maxKeys)
		awsResult, err := os.S3Client.ListObjectsV2(ctx, params)

		if err != nil {
			return nil, err
		}

		if len(awsResult.Contents) > 0 {
			for _, item := range awsResult.Contents {
				object := new(objectstorage.S3Object)
				object.Key = *item.Key
				object.LastModified = *item.LastModified
				object.Size = *item.Size

				objects.AppendOneObject(*object)
				objectCount++
			}
		} else {
			break
		}

		if *awsResult.IsTruncated {
			nextContinuationToken = awsResult.NextContinuationToken
			params.ContinuationToken = nextContinuationToken
		}

	}

	objects.ObjectCount = objectCount
	return objects, nil
}
