package watcher

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
)

// Watcher interface defines the methods for watching content changes.
// It is used to abstract the implementation details of different watchers.
type Watcher interface {
	// Watch starts watching a content source for changes.
	Watch(ctx context.Context)
}

// generateUniqueConfigID returns a hash representing this watcher content.
func (a *AwsAppConfigWatcher) generateUniqueID() string {
	raw := fmt.Sprintf("%s-%s-%s-%s",
		a.watcherOptions.Application,
		a.watcherOptions.Environment,
		a.watcherOptions.ContentScope,
		a.watcherOptions.ClientID)

	hash := sha1.Sum([]byte(raw))
	return hex.EncodeToString(hash[:])
}
