package objectstorage

import (
	"testing"
	"time"
)

func TestAppendOneObject(t *testing.T) {
	result := &ListObjectResult{}

	newObj := S3Object{
		Key:          "test/file.txt",
		LastModified: time.Now(),
		Size:         1024,
	}

	updated := result.AppendOneObject(newObj)

	if len(updated.Objects) != 1 {
		t.Errorf("Expected 1 object, got %d", len(updated.Objects))
	}

	got := updated.Objects[0]
	if got.Key != newObj.Key || got.Size != newObj.Size || !got.LastModified.Equal(newObj.LastModified) {
		t.Errorf("Object mismatch.\nExpected: %+v\nGot: %+v", newObj, got)
	}
}
