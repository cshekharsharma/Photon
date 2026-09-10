package aws

import (
	"context"
	"errors"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3/types"

	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/cshekharsharma/photon/cloud/entity/objectstorage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockS3Client struct {
	mock.Mock
}

type MockS3Presigner struct {
	mock.Mock
}

func (m *MockS3Client) GetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(*s3.GetObjectOutput), args.Error(1)
}

func (m *MockS3Client) HeadObject(ctx context.Context, params *s3.HeadObjectInput, optFns ...func(*s3.Options)) (*s3.HeadObjectOutput, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(*s3.HeadObjectOutput), args.Error(1)
}

func (m *MockS3Client) PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(*s3.PutObjectOutput), args.Error(1)
}

func (m *MockS3Client) CopyObject(ctx context.Context, params *s3.CopyObjectInput, optFns ...func(*s3.Options)) (*s3.CopyObjectOutput, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(*s3.CopyObjectOutput), args.Error(1)
}

func (m *MockS3Client) ListObjectsV2(ctx context.Context, params *s3.ListObjectsV2Input, optFns ...func(*s3.Options)) (*s3.ListObjectsV2Output, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(*s3.ListObjectsV2Output), args.Error(1)
}

func (m *MockS3Presigner) PresignGetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.PresignOptions)) (*v4.PresignedHTTPRequest, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(*v4.PresignedHTTPRequest), args.Error(1)
}

func (m *MockS3Presigner) PresignPutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.PresignOptions)) (*v4.PresignedHTTPRequest, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(*v4.PresignedHTTPRequest), args.Error(1)
}

func TestGetObject(t *testing.T) {
	ctx := context.TODO()
	mockS3 := new(MockS3Client)
	storage := ObjectStorage{S3Client: mockS3}

	mockS3.On("GetObject", ctx, &s3.GetObjectInput{
		Bucket: aws.String("bucket"),
		Key:    aws.String("key"),
	}).Return(&s3.GetObjectOutput{
		ContentLength: aws.Int64(123),
		ContentType:   aws.String("text/plain"),
		ETag:          aws.String("etag"),
		LastModified:  aws.Time(time.Now()),
	}, nil)

	result, err := storage.GetObject(ctx, "bucket", "key")
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, int64(123), result.ContentLength)
}

func TestGetObject_Error(t *testing.T) {
	ctx := context.TODO()
	mockS3 := new(MockS3Client)
	storage := ObjectStorage{S3Client: mockS3}

	mockS3.On("GetObject", ctx, mock.Anything).Return((*s3.GetObjectOutput)(nil), errors.New("get error"))
	_, err := storage.GetObject(ctx, "bucket", "key")
	assert.Error(t, err)
}

func TestGetMultipleObjects(t *testing.T) {
	ctx := context.TODO()

	mockS3 := new(MockS3Client)
	storage := ObjectStorage{S3Client: mockS3}

	bucket := "test-bucket"
	keys := []string{"key1", "key2", "key3"}
	expectedResults := make([]*objectstorage.GetObjectResult, len(keys))

	for i, key := range keys {
		mockS3.On("GetObject", ctx, &s3.GetObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(key),
		}).Return(&s3.GetObjectOutput{
			ContentLength: aws.Int64(123),
			ContentType:   aws.String("text/plain"),
			ETag:          aws.String("etag" + strconv.Itoa(i)),
			LastModified:  aws.Time(time.Now()),
		}, nil)

		expectedResults[i] = &objectstorage.GetObjectResult{
			ContentLength: 123,
			ContentType:   "text/plain",
			ETag:          "etag" + strconv.Itoa(i),
			LastModified:  time.Now(),
		}
	}

	results, err := storage.GetMultipleObjects(ctx, bucket, keys)

	assert.NoError(t, err)

	sort.Slice(results, func(i, j int) bool {
		return results[i].ETag < results[j].ETag
	})
	sort.Slice(expectedResults, func(i, j int) bool {
		return expectedResults[i].ETag < expectedResults[j].ETag
	})

	assert.Len(t, results, len(keys))
	for i, result := range results {
		assert.Equal(t, expectedResults[i].ETag, result.ETag)
	}

	mockS3.AssertExpectations(t)
}

func TestGetMultipleObjects_CreateChunksErrorAndGetObjectError(t *testing.T) {
	mockS3 := new(MockS3Client)
	storage := ObjectStorage{S3Client: mockS3}

	oldCreate := createChunks
	createChunks = func(s []interface{}, size int) ([][]interface{}, error) {
		return nil, errors.New("chunk error")
	}
	_, err := storage.GetMultipleObjects(context.Background(), "bucket", []string{"a"})
	assert.Error(t, err)
	createChunks = oldCreate

	mockS3.On("GetObject", mock.Anything, mock.Anything).Return((*s3.GetObjectOutput)(nil), errors.New("get error"))
	results, err := storage.GetMultipleObjects(context.Background(), "bucket", []string{"a"})
	assert.NoError(t, err)
	assert.Len(t, results, 0)
}

func TestPutObject(t *testing.T) {
	ctx := context.TODO()
	mockS3 := new(MockS3Client)
	storage := ObjectStorage{S3Client: mockS3}

	input := &objectstorage.PutObjectInput{
		Bucket:       "bucket",
		Source:       strings.NewReader("Hello, World!"),
		TargetPath:   "path/to/object",
		Acl:          "private",
		MimeType:     "text/plain",
		CacheControl: "public,max-age=60",
	}

	mockS3.On("PutObject", ctx, &s3.PutObjectInput{
		Bucket:       aws.String(input.Bucket),
		Key:          aws.String(input.TargetPath),
		ContentType:  aws.String(input.MimeType),
		CacheControl: aws.String(input.CacheControl),
		Body:         input.Source,
		ACL:          input.Acl,
	}).Return(&s3.PutObjectOutput{
		ETag: aws.String("etag"),
	}, nil)

	result, err := storage.PutObject(ctx, input)
	assert.NoError(t, err)
	assert.NotNil(t, result)
}

func TestPutObject_DefaultACL_AndError(t *testing.T) {
	ctx := context.TODO()
	mockS3 := new(MockS3Client)
	storage := ObjectStorage{S3Client: mockS3}

	input := &objectstorage.PutObjectInput{
		Bucket:     "bucket",
		Source:     strings.NewReader("data"),
		TargetPath: "path",
		Acl:        "",
		MimeType:   "text/plain",
	}

	mockS3.On("PutObject", ctx, mock.MatchedBy(func(in *s3.PutObjectInput) bool {
		return in.ACL == types.ObjectCannedACLPrivate
	})).Return((*s3.PutObjectOutput)(nil), errors.New("put error"))

	_, err := storage.PutObject(ctx, input)
	assert.Error(t, err)
}

func TestPutObject_DisableACL(t *testing.T) {
	ctx := context.TODO()
	mockS3 := new(MockS3Client)
	storage := ObjectStorage{S3Client: mockS3}

	input := &objectstorage.PutObjectInput{
		Bucket:       "bucket",
		Source:       strings.NewReader("data"),
		TargetPath:   "path",
		MimeType:     "text/plain",
		CacheControl: "public,max-age=2592000,stale-while-revalidate=86400",
		DisableACL:   true,
	}

	mockS3.On("PutObject", ctx, mock.MatchedBy(func(in *s3.PutObjectInput) bool {
		return in != nil &&
			in.ACL == "" &&
			in.CacheControl != nil &&
			*in.CacheControl == input.CacheControl
	})).Return(&s3.PutObjectOutput{
		ETag: aws.String("etag"),
	}, nil)

	result, err := storage.PutObject(ctx, input)
	assert.NoError(t, err)
	assert.NotNil(t, result)
}

func TestCopyObject(t *testing.T) {
	ctx := context.TODO()
	mockS3 := new(MockS3Client)
	storage := ObjectStorage{S3Client: mockS3}

	srcBucket, srcObject := "src-bucket", "src-object"
	destBucket, destObject := "dest-bucket", "dest-object"

	mockS3.On("CopyObject", ctx, &s3.CopyObjectInput{
		Bucket:     aws.String(destBucket),
		CopySource: aws.String(srcBucket + "/" + srcObject),
		Key:        aws.String(destBucket + "/" + destObject),
	}).Return(&s3.CopyObjectOutput{}, nil)

	result, err := storage.CopyObject(ctx, srcBucket, srcObject, destBucket, destObject)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.IsCopied)
}

func TestCopyObject_Error(t *testing.T) {
	ctx := context.TODO()
	mockS3 := new(MockS3Client)
	storage := ObjectStorage{S3Client: mockS3}

	mockS3.On("CopyObject", ctx, mock.Anything).Return((*s3.CopyObjectOutput)(nil), errors.New("copy error"))
	_, err := storage.CopyObject(ctx, "b", "s", "d", "o")
	assert.Error(t, err)
}

func TestGetPresignedUrl(t *testing.T) {
	mockS3Client := new(MockS3Client)
	mockPresigner := new(MockS3Presigner)
	storage := ObjectStorage{
		S3Client:          mockS3Client,
		S3PresignerClient: mockPresigner,
	}

	bucket := "test-bucket"
	key := "test-key"
	ttl := 15 * time.Minute

	expectedPresignedUrl := "http://example.com/test"
	mockPresigner.On("PresignGetObject", mock.Anything, mock.AnythingOfType("*s3.GetObjectInput"),
		mock.Anything).Return(&v4.PresignedHTTPRequest{
		URL: expectedPresignedUrl,
	}, nil)

	result, err := storage.GetPresignedUrl(context.Background(), bucket, key, ttl)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expectedPresignedUrl, result.PresignedUrl)
	assert.LessOrEqual(t, time.Now().Add(-ttl), result.LastModified)

	mockPresigner.AssertExpectations(t)
}

func TestGetPresignedUrl_Error(t *testing.T) {
	mockPresigner := new(MockS3Presigner)
	storage := ObjectStorage{
		S3PresignerClient: mockPresigner,
	}

	mockPresigner.On("PresignGetObject", mock.Anything, mock.AnythingOfType("*s3.GetObjectInput"), mock.Anything).
		Return((*v4.PresignedHTTPRequest)(nil), errors.New("presign error"))

	_, err := storage.GetPresignedUrl(context.Background(), "bucket", "key", time.Minute)
	assert.Error(t, err)
}

func TestGetPresignedPutUrl(t *testing.T) {
	mockPresigner := new(MockS3Presigner)
	storage := ObjectStorage{S3PresignerClient: mockPresigner}
	input := &objectstorage.PresignedPutObjectInput{
		Bucket:               "bucket",
		Key:                  "key",
		ContentType:          "image/jpeg",
		TTL:                  2 * time.Minute,
		ServerSideEncryption: "AES256",
	}
	mockPresigner.On("PresignPutObject", mock.Anything, mock.MatchedBy(func(in *s3.PutObjectInput) bool {
		return in != nil &&
			aws.ToString(in.Bucket) == input.Bucket &&
			aws.ToString(in.Key) == input.Key &&
			aws.ToString(in.ContentType) == input.ContentType &&
			string(in.ServerSideEncryption) == input.ServerSideEncryption
	}), mock.Anything).Return(&v4.PresignedHTTPRequest{URL: "http://upload"}, nil)

	result, err := storage.GetPresignedPutUrl(context.Background(), input)
	assert.NoError(t, err)
	assert.Equal(t, "http://upload", result.PresignedUrl)
	assert.Equal(t, input.TTL, result.TTL)
}

func TestGetPresignedPutUrl_DefaultTTLNilAndError(t *testing.T) {
	mockPresigner := new(MockS3Presigner)
	storage := ObjectStorage{S3PresignerClient: mockPresigner}

	_, err := storage.GetPresignedPutUrl(context.Background(), nil)
	assert.Error(t, err)

	mockPresigner.On("PresignPutObject", mock.Anything, mock.AnythingOfType("*s3.PutObjectInput"), mock.Anything).
		Return((*v4.PresignedHTTPRequest)(nil), errors.New("put presign error"))
	_, err = storage.GetPresignedPutUrl(context.Background(), &objectstorage.PresignedPutObjectInput{Bucket: "b", Key: "k"})
	assert.Error(t, err)
}

func TestHeadObject(t *testing.T) {
	mockS3 := new(MockS3Client)
	storage := ObjectStorage{S3Client: mockS3}
	now := time.Now()
	mockS3.On("HeadObject", mock.Anything, &s3.HeadObjectInput{
		Bucket: aws.String("bucket"),
		Key:    aws.String("key"),
	}).Return(&s3.HeadObjectOutput{
		ContentLength: aws.Int64(42),
		ContentType:   aws.String("image/jpeg"),
		ETag:          aws.String("etag"),
		LastModified:  aws.Time(now),
	}, nil)

	result, err := storage.HeadObject(context.Background(), "bucket", "key")
	assert.NoError(t, err)
	assert.Equal(t, int64(42), result.ContentLength)
	assert.Equal(t, "image/jpeg", result.ContentType)
	assert.Equal(t, "etag", result.ETag)
	assert.Equal(t, now, result.LastModified)
}

func TestHeadObject_Error(t *testing.T) {
	mockS3 := new(MockS3Client)
	storage := ObjectStorage{S3Client: mockS3}
	mockS3.On("HeadObject", mock.Anything, mock.Anything).Return((*s3.HeadObjectOutput)(nil), errors.New("head error"))
	_, err := storage.HeadObject(context.Background(), "bucket", "key")
	assert.Error(t, err)
}

func TestListObjectV2(t *testing.T) {
	mockS3 := new(MockS3Client)
	os := ObjectStorage{S3Client: mockS3}
	bucket := "test-bucket"
	prefix := "test/"
	maxObjects := int32(100)

	expectedObjects := []objectstorage.S3Object{
		{Key: "test/file1.jpg", LastModified: time.Now(), Size: 1024},
		{Key: "test/file2.jpg", LastModified: time.Now().Add(-time.Hour), Size: 2048},
	}

	mockS3.On("ListObjectsV2", mock.Anything, mock.AnythingOfType("*s3.ListObjectsV2Input")).Return(&s3.ListObjectsV2Output{
		Contents: []types.Object{
			{Key: aws.String("test/file1.jpg"), LastModified: aws.Time(expectedObjects[0].LastModified), Size: aws.Int64(expectedObjects[0].Size)},
			{Key: aws.String("test/file2.jpg"), LastModified: aws.Time(expectedObjects[1].LastModified), Size: aws.Int64(expectedObjects[1].Size)},
		},
		IsTruncated: aws.Bool(false),
	}, nil)

	result, err := os.ListObject(context.Background(), bucket, prefix, maxObjects)
	assert.NoError(t, err)
	assert.Len(t, result.Objects, len(expectedObjects))
	for i, obj := range result.Objects {
		assert.Equal(t, expectedObjects[i].Key, obj.Key)
		assert.Equal(t, expectedObjects[i].Size, obj.Size)
		assert.Equal(t, expectedObjects[i].LastModified, obj.LastModified)
	}

	mockS3.AssertExpectations(t)
}

func TestListObject_ErrorAndEmpty(t *testing.T) {
	mockS3 := new(MockS3Client)
	os := ObjectStorage{S3Client: mockS3}

	mockS3.On("ListObjectsV2", mock.Anything, mock.AnythingOfType("*s3.ListObjectsV2Input")).
		Return((*s3.ListObjectsV2Output)(nil), errors.New("list error")).Once()

	_, err := os.ListObject(context.Background(), "bucket", "pref", 10)
	assert.Error(t, err)

	mockS3.On("ListObjectsV2", mock.Anything, mock.AnythingOfType("*s3.ListObjectsV2Input")).
		Return(&s3.ListObjectsV2Output{
			Contents:    []types.Object{},
			IsTruncated: aws.Bool(false),
		}, nil).Once()

	result, err := os.ListObject(context.Background(), "bucket", "pref", 10)
	assert.NoError(t, err)
	assert.Equal(t, 0, result.ObjectCount)
}

func TestListObject_TruncatedAndLargeBatch(t *testing.T) {
	mockS3 := new(MockS3Client)
	os := ObjectStorage{S3Client: mockS3}

	first := &s3.ListObjectsV2Output{
		Contents: []types.Object{
			{Key: aws.String("k1"), LastModified: aws.Time(time.Now()), Size: aws.Int64(1)},
		},
		IsTruncated:           aws.Bool(true),
		NextContinuationToken: aws.String("next"),
	}
	second := &s3.ListObjectsV2Output{
		Contents: []types.Object{
			{Key: aws.String("k2"), LastModified: aws.Time(time.Now()), Size: aws.Int64(2)},
		},
		IsTruncated: aws.Bool(false),
	}

	mockS3.On("ListObjectsV2", mock.Anything, mock.AnythingOfType("*s3.ListObjectsV2Input")).
		Return(first, nil).Once()
	mockS3.On("ListObjectsV2", mock.Anything, mock.AnythingOfType("*s3.ListObjectsV2Input")).
		Return(second, nil).Once()

	result, err := os.ListObject(context.Background(), "bucket", "pref", MAX_OBJECT_PROCESSING_LIMIT+1)
	assert.NoError(t, err)
	assert.Equal(t, 2, result.ObjectCount)
}
