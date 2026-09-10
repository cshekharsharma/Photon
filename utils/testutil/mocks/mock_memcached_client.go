package mocks

import (
	"github.com/bradfitz/gomemcache/memcache"
	"github.com/stretchr/testify/mock"
)

type MockMemcachedClient struct {
	mock.Mock
}

func (m *MockMemcachedClient) Get(key string) (*memcache.Item, error) {
	args := m.Called(key)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*memcache.Item), args.Error(1)
}

func (m *MockMemcachedClient) GetMulti(keys []string) (map[string]*memcache.Item, error) {
	args := m.Called(keys)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]*memcache.Item), args.Error(1)
}

func (m *MockMemcachedClient) Set(item *memcache.Item) error {
	args := m.Called(item)
	return args.Error(0)
}

func (m *MockMemcachedClient) Add(item *memcache.Item) error {
	args := m.Called(item)
	return args.Error(0)
}

func (m *MockMemcachedClient) Replace(item *memcache.Item) error {
	args := m.Called(item)
	return args.Error(0)
}

func (m *MockMemcachedClient) Append(item *memcache.Item) error {
	args := m.Called(item)
	return args.Error(0)
}

func (m *MockMemcachedClient) Prepend(item *memcache.Item) error {
	args := m.Called(item)
	return args.Error(0)
}

func (m *MockMemcachedClient) CompareAndSwap(item *memcache.Item) error {
	args := m.Called(item)
	return args.Error(0)
}

func (m *MockMemcachedClient) Delete(key string) error {
	args := m.Called(key)
	return args.Error(0)
}

func (m *MockMemcachedClient) DeleteAll() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockMemcachedClient) GetAndTouch(key string, expiration int32) (*memcache.Item, error) {
	args := m.Called(key, expiration)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*memcache.Item), args.Error(1)
}

func (m *MockMemcachedClient) Touch(key string, seconds int32) error {
	args := m.Called(key, seconds)
	return args.Error(0)
}

func (m *MockMemcachedClient) Increment(key string, delta uint64) (uint64, error) {
	args := m.Called(key, delta)
	return args.Get(0).(uint64), args.Error(1)
}

func (m *MockMemcachedClient) Decrement(key string, delta uint64) (uint64, error) {
	args := m.Called(key, delta)
	return args.Get(0).(uint64), args.Error(1)
}

func (m *MockMemcachedClient) Ping() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockMemcachedClient) FlushAll() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockMemcachedClient) Close() error {
	args := m.Called()
	return args.Error(0)
}
