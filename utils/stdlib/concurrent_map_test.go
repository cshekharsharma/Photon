package stdlib

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConcurrentMapAddGet(t *testing.T) {
	dict := ConcurrentMap{}
	dict.Add("key1", "value1")

	value := dict.Get("key1")
	assert.Equal(t, "value1", value, "Add should insert the correct value")
}

func TestConcurrentMapRemove(t *testing.T) {
	dict := ConcurrentMap{}
	dict.Add("key1", "value1")

	removed := dict.Remove("key1")
	assert.True(t, removed, "Remove should return true when an existing key is removed")

	removed = dict.Remove("key2")
	assert.False(t, removed, "Remove should return false when a non-existing key is removed")
}

func TestConcurrentMapExist(t *testing.T) {
	dict := ConcurrentMap{}
	dict.Add("key1", "value1")

	assert.True(t, dict.Exist("key1"), "Exist should return true for an existing key")
	assert.False(t, dict.Exist("key2"), "Exist should return false for a non-existing key")
}

func TestConcurrentMapSize(t *testing.T) {
	dict := ConcurrentMap{}
	dict.Add("key1", "value1")
	dict.Add("key2", "value2")

	size := dict.Size()
	assert.Equal(t, 2, size, "Size should return the correct number of elements")
}

func TestConcurrentMapClear(t *testing.T) {
	dict := ConcurrentMap{}
	dict.Add("key1", "value1")
	dict.Add("key2", "value2")

	dict.Clear()
	assert.Equal(t, 0, dict.Size(), "Clear should remove all elements")
}

func TestConcurrentMapGetKeys(t *testing.T) {
	dict := ConcurrentMap{}
	dict.Add("key1", "value1")
	dict.Add("key2", "value2")

	keys := dict.GetKeys()
	assert.ElementsMatch(t, []string{"key1", "key2"}, keys, "GetKeys should return all keys")
}

func TestConcurrentMapGetValues(t *testing.T) {
	dict := ConcurrentMap{}
	dict.Add("key1", "value1")
	dict.Add("key2", "value2")

	values := dict.GetValues()
	assert.ElementsMatch(t, []string{"value1", "value2"}, values, "GetValues should return all values")
}
