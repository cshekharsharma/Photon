package stdlib

import (
	"sync"
)

// Dictionary - the dictionary object with key of type string & vlaue of type string
type ConcurrentMap struct {
	Data map[string]string
	Id   string
	sync.RWMutex
}

// Add adds a new item to the dictionary
func (dict *ConcurrentMap) Add(key string, value string) {
	dict.Lock()
	defer dict.Unlock()

	if dict.Data == nil {
		dict.Data = make(map[string]string)
	}

	dict.Data[key] = value
}

// Remove removes a value from the dictionary, given its key
func (dict *ConcurrentMap) Remove(key string) bool {
	dict.Lock()
	defer dict.Unlock()

	_, ok := dict.Data[key]

	if ok {
		delete(dict.Data, key)
	}

	return ok
}

// Exist returns true if the key exists in the dictionary
func (dict *ConcurrentMap) Exist(key string) bool {
	dict.RLock()
	defer dict.RUnlock()
	_, ok := dict.Data[key]
	return ok
}

// Get returns the value associated with the key
func (dict *ConcurrentMap) Get(key string) string {
	dict.RLock()
	defer dict.RUnlock()
	return dict.Data[key]
}

// Clear removes all the Reports from the dictionary
func (dict *ConcurrentMap) Clear() {
	dict.Lock()
	defer dict.Unlock()
	dict.Data = make(map[string]string)
}

// Size returns the amount of elements in the dictionary
func (dict *ConcurrentMap) Size() int {
	dict.RLock()
	defer dict.RUnlock()
	return len(dict.Data)
}

// GetKeys returns a slice of all the keys present
func (dict *ConcurrentMap) GetKeys() []string {
	dict.RLock()
	defer dict.RUnlock()

	keys := []string{}

	for i := range dict.Data {
		keys = append(keys, i)
	}
	return keys
}

// GetValues returns a slice of all the values present
func (dict *ConcurrentMap) GetValues() []string {
	dict.RLock()
	defer dict.RUnlock()

	values := []string{}

	for i := range dict.Data {
		values = append(values, dict.Data[i])
	}

	return values
}
