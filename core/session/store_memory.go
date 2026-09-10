package session

import (
	"github.com/alexedwards/scs/v2/memstore"
)

func newMemoryStore() storeWithCloser {
	return storeWithCloser{store: memstore.New()}
}
