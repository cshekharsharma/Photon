package stdlib

import "testing"

func TestNewLRUPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected panic on capacity < 0")
		}
	}()
	_ = NewLRU[string, int](-1, nil)
}

func TestLRUBasicOps(t *testing.T) {
	evictedKeys := []string{}
	ev := func(k string, v int) { evictedKeys = append(evictedKeys, k) }

	l := NewLRU(2, ev)

	// Set and Get
	if ev := l.Set("a", 1); ev {
		t.Errorf("expected no eviction")
	}
	if v, ok := l.Get("a"); !ok || v != 1 {
		t.Errorf("expected get=1,ok=true got %v,%v", v, ok)
	}

	// Update existing
	l.Set("a", 2)
	if v, _ := l.Get("a"); v != 2 {
		t.Errorf("expected updated value=2 got %v", v)
	}

	if v, _ := l.Get("a_nonexists"); v != 0 {
		t.Errorf("expected updated value=0 got %v", v)
	}

	// Insert more keys to trigger eviction
	l.Set("b", 3)
	if ev := l.Set("c", 4); !ev {
		t.Errorf("expected eviction")
	}
	if l.Len() != 2 {
		t.Errorf("expected len=2 got %d", l.Len())
	}
	if l.Capacity() != 2 {
		t.Errorf("expected capacity=2")
	}

	// Ensure oldest key was evicted ("a")
	if _, ok := l.Peek("a"); ok {
		t.Errorf("expected a to be evicted")
	}

	// Keys order MRU->LRU
	keys := l.Keys()
	if len(keys) != 2 {
		t.Fatalf("expected 2 keys")
	}
	if keys[0] != "c" || keys[1] != "b" {
		t.Errorf("unexpected order %v", keys)
	}

	// Delete existing
	if !l.Delete("b") {
		t.Errorf("expected delete true")
	}
	if l.Delete("missing") {
		t.Errorf("expected delete false")
	}

	// Purge
	l.Set("x", 10)
	l.Set("y", 11)
	l.Purge()
	if l.Len() != 0 {
		t.Errorf("expected empty after purge")
	}
	if len(evictedKeys) == 0 {
		t.Errorf("expected eviction callbacks on purge")
	}
}

func TestLRUTailKey(t *testing.T) {
	l := NewLRU[string, int](2, nil)
	if _, ok := l.TailKey(); ok {
		t.Errorf("expected TailKey false on empty cache")
	}

	l.Set("a", 1)
	l.Set("b", 2)
	if k, ok := l.TailKey(); !ok || k != "a" {
		t.Errorf("expected tail key a, got %v %v", k, ok)
	}
}

func TestEvictCallbackOnSetAndDelete(t *testing.T) {
	called := false
	ev := func(k string, v int) { called = true }
	l := NewLRU(1, ev)
	l.Set("a", 1)
	l.Set("b", 2) // triggers eviction
	if !called {
		t.Errorf("expected eviction callback on set overflow")
	}
	called = false
	l.Delete("b")
	if !called {
		t.Errorf("expected eviction callback on delete")
	}
}

func TestSafeLRUBasics(t *testing.T) {
	core := NewLRU[string, int](2, nil)
	safe := NewSafeLRU(core)

	safe.Set("a", 1)
	safe.Set("b", 2)
	if safe.Len() != 2 {
		t.Errorf("expected len=2")
	}
	if v, ok := safe.Get("a"); !ok || v != 1 {
		t.Errorf("expected get=1")
	}
	if v, ok := safe.Peek("b"); !ok || v != 2 {
		t.Errorf("expected peek=2")
	}
	if !safe.Delete("a") {
		t.Errorf("expected delete true")
	}
	if safe.Capacity() != 2 {
		t.Errorf("expected capacity=2")
	}
	safe.Keys()
	safe.Purge()
}

func TestNewSafeLRUPanicsOnNil(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected panic on nil core")
		}
	}()
	_ = NewSafeLRU[string, int](nil)
}

func TestSyncRWMutexImplNoopMethods(t *testing.T) {
	var m syncRWMutexImpl
	locked := false

	m.Lock()
	locked = true
	m.Unlock()
	if !locked {
		t.Fatal("expected write lock path to run")
	}

	readLocked := false
	m.RLock()
	readLocked = true
	m.RUnlock()
	if !readLocked {
		t.Fatal("expected read lock path to run")
	}
}
