package flashdb

import (
	"fmt"
	"math"
	"runtime"
	"strconv"
	"testing"
	"time"

	"github.com/cshekharsharma/photon/utils/stdlib"
)

func newStore(t *testing.T, shards uint8, budget uint64, janitorEvery time.Duration) *Store {
	t.Helper()
	s, err := New(Options{
		Shards:         shards,
		MaxStorageSize: budget,
		JanitorEvery:   janitorEvery,
	})

	if err != nil {
		t.Fatalf("new store: %v", err)
	}

	return s
}

func TestNew_OptionsValidation(t *testing.T) {
	_, err := New(Options{Shards: 1, MaxStorageSize: 0})
	if err == nil {
		t.Fatalf("expected error for zero MaxStorageSize")
	}

	s, err := New(Options{MaxStorageSize: 1 << 20})
	if err != nil || s == nil {
		t.Fatalf("unexpected: %v %v", s, err)
	}
	s.Close()

	s, err = New(Options{MaxStorageSize: 0})
	if err == nil || s != nil {
		t.Fatalf("expected error, got: s=%v. err=%v", s, err)
	}

	s, err = New(Options{Shards: 0})
	if err == nil || s != nil {
		t.Fatalf("expected error, got: s=%v. err=%v", s, err)
	}
}

func TestSetGetExistsDelete_Basics(t *testing.T) {
	s := newStore(t, 4, 1<<20, 0)
	defer s.Close()

	// invalid key
	if _, err := s.Get(""); err == nil {
		t.Fatalf("expected invalid key")
	}

	if ok, err := s.Exists(""); err == nil || ok {
		t.Fatalf("expected invalid key on exists")
	}

	if ok, err := s.Delete(""); err == nil || ok {
		t.Fatalf("expected invalid key on delete")
	}

	// set & get
	ok, err := s.Set("u:1", map[string]any{"name": "neo", "n": int64(7)}, 0)
	if err != nil || !ok {
		t.Fatalf("set failed: %v %v", ok, err)
	}

	v, err := s.Get("u:1")
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}

	m, _ := v.(map[string]any)
	if m["name"].(string) != "neo" || m["n"].(int64) != 7 {
		t.Fatalf("got %+v", m)
	}

	// exists
	ex, err := s.Exists("u:1")
	if err != nil || !ex {
		t.Fatalf("exists failed")
	}

	// delete
	del, err := s.Delete("u:1")
	if err != nil || !del {
		t.Fatalf("delete failed")
	}

	// not found thereafter
	if _, err := s.Get("u:1"); err == nil {
		t.Fatalf("expected not found")
	}

	ex, _ = s.Exists("u:1")
	if ex {
		t.Fatalf("expected !exists")
	}
}

func TestSet_Update_Shrink_NoTrim(t *testing.T) {
	base := computeSize("x", valWithString(2000))
	st := newStore1(t, uint64(base*3), 0)

	defer st.Close()

	ok, err := st.Set("x", valWithString(2000), 5*time.Second)
	if err != nil || !ok {
		t.Fatalf("initial Set failed: ok=%v err=%v", ok, err)
	}

	// another key to ensure LRU has >1 item (to detect accidental evictions)
	if ok, _ := st.Set("y", valWithString(10), 0); !ok {
		t.Fatalf("preload y failed")
	}

	// shrink value for x
	ok, err = st.Set("x", valWithString(10), 0)
	if err != nil || !ok {
		t.Fatalf("shrink Set failed: ok=%v err=%v", ok, err)
	}

	// x should exist and be the smaller payload now
	if v, err := st.Get("x"); err != nil || len(v.(map[string]any)["s"].(string)) != 10 {
		t.Fatalf("Get(x) wrong after shrink, v=%v err=%v", v, err)
	}

	// y should still exist (no trimming)
	if _, err := st.Get("y"); err != nil {
		t.Fatalf("y unexpectedly missing after shrink trim: %v", err)
	}
}

func TestSet_Update_Grow_TrimLRU_NotSelf(t *testing.T) {
	sizeA := computeSize("A", valWithString(1200))
	sizeBsmall := computeSize("B", valWithString(50))
	sizeBbig := computeSize("B", valWithString(5000))

	capacity := uint64(sizeBbig + 64)
	if need := uint64(sizeA + sizeBsmall); capacity < need {
		capacity = need
	}

	if capacity >= uint64(sizeA+sizeBbig) {
		t.Fatalf("bad test math: capacity must be < sizeA+sizeBbig")
	}

	st := newStore1(t, capacity, 0)
	defer st.Close()

	if ok, _ := st.Set("A", valWithString(1200), 0); !ok {
		t.Fatalf("Set(A) failed")
	}

	if ok, _ := st.Set("B", valWithString(50), 0); !ok {
		t.Fatalf("Set(B small) failed")
	}

	// Make A the LRU by touching B.
	if _, err := st.Get("B"); err != nil {
		t.Fatalf("Get(B) recency bump failed: %v", err)
	}

	// Grow B -> should evict A, keep B
	if ok, err := st.Set("B", valWithString(5000), 0); err != nil || !ok {
		t.Fatalf("Set(B big) failed: ok=%v err=%v", ok, err)
	}

	if _, err := st.Get("A"); err == nil {
		t.Fatalf("A should have been evicted as LRU")
	}

	if v, err := st.Get("B"); err != nil || len(v.(map[string]any)["s"].(string)) != 5000 {
		t.Fatalf("B not present/updated after growth trim, v=%v err=%v", v, err)
	}
}

// Reject item that can never fit capacity
func TestSet_RejectTooLargeForShard(t *testing.T) {
	tooBig := computeSize("K", valWithString(1<<15))
	st := newStore1(t, uint64(tooBig-1), 0) // capacity 1 less than needed
	defer st.Close()

	ok, err := st.Set("K", valWithString(1<<15), 0)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}

	if ok {
		t.Fatalf("expected reject for too-large item")
	}

	if _, err := st.Get("K"); err == nil {
		t.Fatalf("too-large key should not be stored")
	}
}

// Admission REJECT: incoming is colder than victim -> return false, keep victim
func TestSet_AdmissionReject_KeepHotVictim(t *testing.T) {
	// Fill with a hot victim; capacity allows only one large entry
	valVictim := valWithString(4000)
	valIncoming := valWithString(4000)

	capacity := uint64(computeSize("hot", valVictim))
	st := newStore1(t, capacity, 0)
	defer st.Close()

	if ok, _ := st.Set("hot", valVictim, 0); !ok {
		t.Fatalf("store hot failed")
	}

	// Make victim *very hot* in TinyLFU
	for i := 0; i < 1000; i++ {
		st.lfu.Record("hot")
	}

	// Incoming is cold (will get only 1 record when Set runs)
	ok, err := st.Set("cold", valIncoming, 0)
	if err != nil {
		t.Fatalf("Set(cold) unexpected err: %v", err)
	}

	if ok {
		t.Fatalf("expected admission reject for cold incoming")
	}

	// Ensure victim still present; incoming not present
	if _, err := st.Get("hot"); err != nil {
		t.Fatalf("hot victim should remain: %v", err)
	}

	if _, err := st.Get("cold"); err == nil {
		t.Fatalf("cold should not have been admitted")
	}
}

func TestSet_AdmissionAccept_EvictVictim_SameLenKeys(t *testing.T) {
	t.Helper()

	v1 := valWithString(3000)
	v2 := valWithString(3000)

	// Use 1-char keys so sizes match exactly.
	size := computeSize("a", v1)

	st := newStore1(t, uint64(size), 0)
	defer st.Close()

	if ok, _ := st.Set("a", v1, 0); !ok {
		t.Fatalf("seed failed")
	}

	st.lfu = nil // deterministic

	if ok, err := st.Set("b", v2, 0); !ok || err != nil {
		t.Fatalf("Set(b) failed: ok=%v err=%v", ok, err)
	}

	if _, err := st.Get("b"); err != nil {
		t.Fatalf("b should be present: %v", err)
	}

	if _, err := st.Get("a"); err == nil {
		t.Fatalf("a should have been evicted")
	}
}

// Default TTL path (ttl==0 -> use DefaultTTL seconds)
func TestSet_DefaultTTL_Applies(t *testing.T) {
	// 2 second default TTL; set with ttl=0 should apply default
	base := computeSize("k", valNum(1))
	st := newStore1(t, uint64(base*4), 2 /*seconds default*/)
	defer st.Close()

	if ok, err := st.Set("k", valNum(1), 0); err != nil || !ok {
		t.Fatalf("Set with default TTL failed: ok=%v err=%v", ok, err)
	}

	ttl, err := st.GetTTL("k")
	if err != nil {
		t.Fatalf("GetTTL err: %v", err)
	}

	// should be in (0, 2] seconds window – allow race slack
	if ttl <= 0 || ttl > 2 {
		t.Fatalf("default TTL not applied, ttl=%d", ttl)
	}
}

// Invalid key path
func TestSet_InvalidKey(t *testing.T) {
	st := newStore1(t, 1<<20, 0)
	defer st.Close()

	if ok, err := st.Set("", valNum(1), 0); err != ErrInvalidKey || ok {
		t.Fatalf("expected ErrInvalidKey, got ok=%v err=%v", ok, err)
	}
}

func TestTTL_Get_Exists_GetTTL_SetTTL(t *testing.T) {
	s := newStore(t, 1, 1<<20, 0)
	defer s.Close()

	ok, err := s.Set("k1", map[string]any{"f": 1}, 2*time.Second)
	if !ok || err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	ttl, err := s.GetTTL("k1")
	if err != nil {
		t.Fatalf("GetTTL failed: %v", err)
	}

	// Contract: >=0 when TTL exists (flooring allowed). Not -1.
	if ttl < 0 {
		t.Fatalf("expected ttl >= 0 (not -1), got %d", ttl)
	}

	// let it expire
	time.Sleep(2100 * time.Millisecond)

	if _, err := s.Get("k1"); err == nil {
		t.Fatalf("expected Get to fail after expiry")
	}

	if ok, _ := s.Exists("k1"); ok {
		t.Fatalf("Exists should be false after expiry")
	}

	if _, err := s.GetTTL("k1"); err == nil {
		t.Fatalf("GetTTL should error after expiry")
	}
}

func TestSetTTL(t *testing.T) {
	s := newStore(t, 1, 1<<20, 0)
	defer s.Close()

	// insert without ttl
	ok, err := s.Set("k1", map[string]any{"f": "v"}, 0)
	if !ok || err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// initially no ttl
	ttl, err := s.GetTTL("k1")
	if err != nil {
		t.Fatalf("GetTTL: %v", err)
	}

	if ttl != -1 {
		t.Fatalf("expected -1 for no ttl, got %d", ttl)
	}

	// apply a 1s ttl (flooring means GetTTL may return 0 or 1 right away)
	if err := s.SetTTL("k1", 1*time.Second); err != nil {
		t.Fatalf("SetTTL failed: %v", err)
	}

	ttl, err = s.GetTTL("k1")
	if err != nil {
		t.Fatalf("GetTTL after SetTTL: %v", err)
	}

	if ttl < 0 || ttl > 1 { // allow 0 or 1 due to truncation
		t.Fatalf("expected ttl in [0,1] right after setting 1s, got %d", ttl)
	}

	// clear ttl
	if err := s.SetTTL("k1", 0); err != nil {
		t.Fatalf("SetTTL clear failed: %v", err)
	}

	ttl, err = s.GetTTL("k1")
	if err != nil || ttl != -1 {
		t.Fatalf("expected -1 after clearing ttl, got %d, err=%v", ttl, err)
	}

	// negative case: SetTTL on missing key
	if err := s.SetTTL("does-not-exist", time.Second); err == nil {
		t.Fatalf("expected error when SetTTL on missing key")
	}
}

func TestGetTTL_AllBranches(t *testing.T) {
	// invalid key
	{
		st := newStore1(t, 4*1024, 0)
		defer st.Close()
		if _, err := st.GetTTL(""); err != ErrInvalidKey {
			t.Fatalf("GetTTL invalid key: got %v want %v", err, ErrInvalidKey)
		}
	}

	st := newStore1(t, 8*1024, 0)
	defer st.Close()

	// missing
	if _, err := st.GetTTL("nope"); err != ErrNotFound {
		t.Fatalf("GetTTL missing: got %v want ErrNotFound", err)
	}

	// no TTL -> -1
	if ok, err := st.Set("noTTL", m("v", "x"), 0); err != nil || !ok {
		t.Fatalf("seed noTTL failed: ok=%v err=%v", ok, err)
	}

	if ttl, err := st.GetTTL("noTTL"); err != nil || ttl != -1 {
		t.Fatalf("GetTTL noTTL: ttl=%d err=%v, want -1,nil", ttl, err)
	}

	// positive TTL
	if ok, err := st.Set("hasTTL", m("v", "x"), 3*time.Second); err != nil || !ok {
		t.Fatalf("seed hasTTL failed: ok=%v err=%v", ok, err)
	}

	ttl, err := st.GetTTL("hasTTL")
	if err != nil {
		t.Fatalf("GetTTL pos ttl error: %v", err)
	}

	if ttl < 1 {
		t.Fatalf("GetTTL pos ttl should be >=1, got %d", ttl)
	}

	// expired at call time → removal + ErrNotFound
	if ok, err := st.Set("willExpire", m("v", "x"), 10*time.Millisecond); err != nil || !ok {
		t.Fatalf("seed willExpire failed: ok=%v err=%v", ok, err)
	}

	time.Sleep(20 * time.Millisecond)
	if _, err := st.GetTTL("willExpire"); err != ErrNotFound {
		t.Fatalf("GetTTL expired should ErrNotFound, got %v", err)
	}
}

func TestAppend_AllBranches(t *testing.T) {
	// 1) invalid key
	{
		st := newStore1(t, 4*1024, 0)
		defer st.Close()

		if err := st.Append("", map[string]string{"f": "x"}); err != ErrInvalidKey {
			t.Fatalf("Append invalid key: got %v want %v", err, ErrInvalidKey)
		}
	}

	// 2) expired key -> ErrNotFound (use a store with generous capacity)
	{
		st := newStore1(t, 8*1024, 0)
		defer st.Close()

		if ok, err := st.Set("exp", m("s", "x"), 5*time.Millisecond); err != nil || !ok {
			t.Fatalf("seed exp failed: ok=%v err=%v", ok, err)
		}

		time.Sleep(10 * time.Millisecond)
		if err := st.Append("exp", map[string]string{"s": "y"}); err != ErrNotFound {
			t.Fatalf("Append expired: got %v want ErrNotFound", err)
		}
	}

	// 3) illegal field (field absent)
	{
		st := newStore1(t, 8*1024, 0)
		defer st.Close()
		if ok, err := st.Set("a", m("s", "a"), 0); err != nil || !ok {
			t.Fatalf("seed a failed: ok=%v err=%v", ok, err)
		}

		if err := st.Append("a", map[string]string{"nope": "zzz"}); err != ErrIllegalField {
			t.Fatalf("Append illegal field: got %v", err)
		}
	}

	// 4) bad type (field exists but not a string)
	{
		st := newStore1(t, 8*1024, 0)
		defer st.Close()

		if ok, err := st.Set("num", m("n", 1), 0); err != nil || !ok {
			t.Fatalf("seed num failed: ok=%v err=%v", ok, err)
		}

		if err := st.Append("num", map[string]string{"n": "sfx"}); err != ErrBadType {
			t.Fatalf("Append bad type: got %v", err)
		}
	}

	// 5) success + size growth + budget enforcement trims LRU victim (not self)
	{
		valA := m("s", "a")
		valB := m("s", "bbb")

		// capacity large enough to hold A + B initially, then a big append on A will overflow
		capA := uint64(computeSize("a", valA))
		capB := uint64(computeSize("b", valB))

		st := newStore1(t, capA+capB+128, 0)
		defer st.Close()

		// seed LRU order: put B first, then A; then bump A again to make it MRU
		if ok, err := st.Set("b", valB, 0); err != nil || !ok {
			t.Fatalf("seed b failed: ok=%v err=%v", ok, err)
		}

		if ok, err := st.Set("a", valA, 0); err != nil || !ok {
			t.Fatalf("seed a failed: ok=%v err=%v", ok, err)
		}

		_, _ = st.Get("a") // ensure A is MRU, B becomes LRU

		// append huge suffix to A -> shard over budget -> evict LRU (B) in the loop
		huge := bigString(8 * 1024)
		if err := st.Append("a", map[string]string{"s": huge}); err != nil {
			t.Fatalf("Append success failed: %v", err)
		}

		if _, err := st.Get("b"); err == nil {
			t.Fatalf("b should have been evicted by budget enforcement")
		}
	}
}

func TestAppend_SizeGrowth_EvictionTrim(t *testing.T) {
	// 1 shard, very small budget to force trims.
	s := newStore(t, 1, 512, 0)
	defer s.Close()

	// Insert a value with string field.
	_, _ = s.Set("a", map[string]any{"s": "x"}, 0)
	_, _ = s.Set("b", map[string]any{"s": "y"}, 0)

	// Appending makes the record larger; after enough growth, shard may need trims.
	for i := 0; i < 10; i++ {
		if err := s.Append("a", map[string]string{"s": "0123456789"}); err != nil {
			t.Fatalf("append failed: %v", err)
		}
	}

	// Access "b" to make it MRU; "a" has become large and MRU too via append.
	// Insert "c" which should trigger LRU eviction (victim tail depends on recency).
	_, _ = s.Set("c", map[string]any{"s": "z"}, 0)

	// One of a/b may have been evicted based on LRU order & size. Ensure store remains consistent.
	// If "b" still exists, appending to it must work; if it was evicted, not found is OK.
	if err := s.Append("b", map[string]string{"s": "!"}); err != nil && err != ErrNotFound {
		t.Fatalf("append b unexpected err: %v", err)
	}
}

func TestAppend_Errors(t *testing.T) {
	s := newStore(t, 1, 1<<20, 0)
	defer s.Close()

	_, _ = s.Set("a", map[string]any{"s": "x", "n": int64(1)}, 0)

	// missing field
	if err := s.Append("a", map[string]string{"missing": "!"}); err != ErrIllegalField {
		t.Fatalf("expected ErrIllegalField, got %v", err)
	}

	// bad type
	if err := s.Append("a", map[string]string{"n": "!"}); err != ErrBadType {
		t.Fatalf("expected ErrBadType, got %v", err)
	}

	// missing key
	if err := s.Append("nope", map[string]string{"s": "!"}); err != ErrNotFound {
		t.Fatalf("expected ErrNotFound")
	}
}

func TestIncrement_AllBranches(t *testing.T) {
	// invalid key
	{
		st := newStore1(t, 4*1024, 0)
		defer st.Close()
		if err := st.Increment("", map[string]int64{"x": 1}); err != ErrInvalidKey {
			t.Fatalf("Increment invalid key: got %v want %v", err, ErrInvalidKey)
		}
	}

	st := newStore1(t, 16*1024, 0)
	defer st.Close()

	// missing
	if err := st.Increment("nope", map[string]int64{"n": 1}); err != ErrNotFound {
		t.Fatalf("Increment missing: got %v want ErrNotFound", err)
	}

	// expired
	if ok, err := st.Set("exp", m("n", int64(1)), 5*time.Millisecond); err != nil || !ok {
		t.Fatalf("seed exp failed: ok=%v err=%v", ok, err)
	}

	time.Sleep(10 * time.Millisecond)
	if err := st.Increment("exp", map[string]int64{"n": 1}); err != ErrNotFound {
		t.Fatalf("Increment expired: got %v want ErrNotFound", err)
	}

	// good: int64
	if ok, err := st.Set("i64", m("n", int64(10)), 0); err != nil || !ok {
		t.Fatalf("seed i64 failed: ok=%v err=%v", ok, err)
	}

	if err := st.Increment("i64", map[string]int64{"n": 5}); err != nil {
		t.Fatalf("Increment int64 failed: %v", err)
	}

	if got, _ := st.Get("i64"); got.(map[string]any)["n"].(int64) != 15 {
		t.Fatalf("int64 value not updated")
	}

	// good: float64
	if ok, err := st.Set("f64", m("n", float64(1.5)), 0); err != nil || !ok {
		t.Fatalf("seed f64 failed: ok=%v err=%v", ok, err)
	}

	if err := st.Increment("f64", map[string]int64{"n": 2}); err != nil {
		t.Fatalf("Increment float64 failed: %v", err)
	}

	if got, _ := st.Get("f64"); got.(map[string]any)["n"].(float64) != 3.5 {
		t.Fatalf("float64 value not updated")
	}

	// good: int
	if ok, err := st.Set("i", m("n", int(7)), 0); err != nil || !ok {
		t.Fatalf("seed int failed: ok=%v err=%v", ok, err)
	}

	if err := st.Increment("i", map[string]int64{"n": 3}); err != nil {
		t.Fatalf("Increment int failed: %v", err)
	}

	if got, _ := st.Get("i"); got.(map[string]any)["n"].(int) != 10 {
		t.Fatalf("int value not updated")
	}

	// good: int32
	if ok, err := st.Set("i32", m("n", int32(2)), 0); err != nil || !ok {
		t.Fatalf("seed int32 failed: ok=%v err=%v", ok, err)
	}
	if err := st.Increment("i32", map[string]int64{"n": 4}); err != nil {
		t.Fatalf("Increment int32 failed: %v", err)
	}
	if got, _ := st.Get("i32"); got.(map[string]any)["n"].(int32) != int32(6) {
		t.Fatalf("int32 value not updated")
	}

	// illegal field (missing)
	mInput := m("n", int64(1))
	if ok, err := st.Set("hasFields", mInput, 0); err != nil || !ok {
		t.Fatalf("seed hasFields failed: ok=%v err=%v", ok, err)
	}

	if err := st.Increment("hasFields", map[string]int64{"absent": 1}); err != ErrIllegalField {
		t.Fatalf("Increment illegal field: got %v want %v", err, ErrIllegalField)
	}

	// bad type (string field)
	if ok, err := st.Set("badType", m("s", "str"), 0); err != nil || !ok {
		t.Fatalf("seed badType failed: ok=%v err=%v", ok, err)
	}

	if err := st.Increment("badType", map[string]int64{"s": 1}); err != ErrBadType {
		t.Fatalf("Increment bad type: got %v want %v", err, ErrBadType)
	}

	// overflow guards (int & int32) – should map to ErrBadType via addInt=false
	if ok, err := st.Set("ovfi", m("n", int(^uint(0)>>1)), 0); err != nil || !ok { // MaxInt
		t.Fatalf("seed ovfi failed: ok=%v err=%v", ok, err)
	}

	if err := st.Increment("ovfi", map[string]int64{"n": 1}); err != ErrBadType {
		t.Fatalf("Increment int overflow: got %v want ErrBadType", err)
	}

	if ok, err := st.Set("ovfi32", m("n", int32(2147483647)), 0); err != nil || !ok { // MaxInt32
		t.Fatalf("seed ovfi32 failed: ok=%v err=%v", ok, err)
	}

	if err := st.Increment("ovfi32", map[string]int64{"n": 1}); err != ErrBadType {
		t.Fatalf("Increment int32 overflow: got %v want ErrBadType", err)
	}
}

func TestIncrement_Decrement_TypesAndOverflow(t *testing.T) {
	s := newStore(t, 1, 1<<20, 0)
	defer s.Close()

	_, _ = s.Set("k", map[string]any{
		"i":   int(1),
		"i32": int32(2),
		"i64": int64(3),
		"f64": float64(1.5),
		"s":   "notnum",
	}, 0)

	// ok increments
	if err := s.Increment("k", map[string]int64{
		"i":   2,
		"i32": 3,
		"i64": 4,
		"f64": 5,
	}); err != nil {
		t.Fatalf("increment failed: %v", err)
	}

	v, _ := s.Get("k")
	m := v.(map[string]any)
	if m["i"].(int) != 3 || m["i32"].(int32) != 5 || m["i64"].(int64) != 7 || m["f64"].(float64) != 6.5 {
		t.Fatalf("bad math: %+v", m)
	}

	// overflow on int32
	_, _ = s.Set("k2", map[string]any{"i32": int32(math.MaxInt32)}, 0)
	if err := s.Increment("k2", map[string]int64{"i32": 1}); err != ErrBadType {
		t.Fatalf("expected ErrBadType on overflow, got %v", err)
	}

	// missing field
	if err := s.Increment("k", map[string]int64{"no": 1}); err != ErrIllegalField {
		t.Fatalf("expected ErrIllegalField")
	}

	// bad type
	if err := s.Increment("k", map[string]int64{"s": 1}); err != ErrBadType {
		t.Fatalf("expected ErrBadType on non-number")
	}

	// decrement via wrapper
	if err := s.Decrement("k", map[string]int64{"i": 1}); err != nil {
		t.Fatalf("decrement failed: %v", err)
	}

	v, _ = s.Get("k")
	if v.(map[string]any)["i"].(int) != 2 {
		t.Fatalf("decrement wrong: %+v", v)
	}
}

func TestMultiOps(t *testing.T) {
	s := newStore(t, 1, 1<<20, 0)
	defer s.Close()

	// MultiSet empty/noop
	if out, err := s.MultiSet(nil, 0); err != nil || len(out) != 0 {
		t.Fatalf("multiset empty")
	}

	// MultiDelete empty/noop
	if out, err := s.MultiDelete(nil); err != nil || len(out) != 0 {
		t.Fatalf("multidel empty")
	}

	// MultiGet empty/noop
	if out, err := s.MultiGet(nil); err != nil || len(out) != 0 {
		t.Fatalf("multiget empty")
	}

	items := map[string]map[string]any{
		"a": {"x": 1},
		"b": {"x": 2},
		"":  {"x": 3}, // invalid
	}

	setRes, _ := s.MultiSet(items, 0)
	if !setRes["a"] || !setRes["b"] || setRes[""] {
		t.Fatalf("multiset result unexpected: %+v", setRes)
	}

	got, _ := s.MultiGet([]string{"a", "b", "c", ""})
	if len(got) != 2 || got["a"] == nil || got["b"] == nil {
		t.Fatalf("multiget wrong: %+v", got)
	}

	delRes, _ := s.MultiDelete([]string{"a", "b", "c", ""})
	if !delRes["a"] || !delRes["b"] || delRes["c"] || delRes[""] {
		t.Fatalf("multidel wrong: %+v", delRes)
	}
}

func TestAdmission_TinyLFU_vs_LRU(t *testing.T) {
	// budget just enough to store one larger entry.
	// We’ll create hot victim and cold incoming to exercise ShouldAdmit().
	s := newStore(t, 1, 512, 0)
	defer s.Close()

	// hot key "hot" — record it frequently in LFU
	for i := 0; i < 32; i++ {
		// record via Store API paths: Set(Get bumps LFU) or call LFU directly is internal.
		_, _ = s.Set("hot", map[string]any{"s": "H"}, 0)
		_, _ = s.Get("hot")
	}

	// insert a second key to fill capacity
	_, _ = s.Set("warm", map[string]any{"s": "W"}, 0)

	// Now attempt to insert cold key "cold" that's competing for space.
	// TinyLFU.ShouldAdmit should usually reject it against hot tail candidate.
	ok, err := s.Set("cold", map[string]any{"s": "C C C C C C"}, 0)
	if err != nil {
		t.Fatalf("set cold: %v", err)
	}

	// Either admitted (and evicted warm) or rejected. But "hot" must remain.
	_, err = s.Get("hot")
	if err != nil {
		t.Fatalf("hot should stay (tinyLFU bias)")
	}

	// And if cold was admitted, the store is still consistent.
	if ok {
		if _, err := s.Get("cold"); err != nil {
			t.Fatalf("cold admitted but not retrievable")
		}
	}
}

func TestJanitor_Expires(t *testing.T) {
	// fast janitor sweep
	s, err := New(Options{
		Shards:         2,
		MaxStorageSize: 1 << 20,
		JanitorEvery:   20 * time.Millisecond,
	})

	if err != nil {
		t.Fatalf("new: %v", err)
	}
	defer s.Close()

	_, _ = s.Set("g1", map[string]any{"x": 1}, 10*time.Millisecond)
	_, _ = s.Set("g2", map[string]any{"x": 2}, 10*time.Millisecond)

	time.Sleep(60 * time.Millisecond) // let janitor tick a few times

	if _, err := s.Get("g1"); err == nil {
		t.Fatalf("expected janitor to evict expired g1")
	}

	if _, err := s.Get("g2"); err == nil {
		t.Fatalf("expected janitor to evict expired g2")
	}
}

func TestExists_ExpiredPathAndValidPath(t *testing.T) {
	s := newStore(t, 1, 1<<20, 0)
	defer s.Close()

	_, _ = s.Set("k", map[string]any{"x": 1}, 0)
	ok, _ := s.Exists("k")
	if !ok {
		t.Fatalf("exists should be true")
	}

	_, _ = s.Set("k2", map[string]any{"x": 2}, 20*time.Millisecond)
	time.Sleep(30 * time.Millisecond)

	ok, _ = s.Exists("k2")
	if ok {
		t.Fatalf("exists should be false post-expiry")
	}

	if _, err := s.Get("k2"); err == nil {
		t.Fatalf("k2 should be gone after Exists cleaned it")
	}
}

// --------------------- Benchmarks ---------------------

func BenchmarkSet(b *testing.B) {
	s := newStoreB(b, 8, 8<<20, 0)
	defer s.Close()

	val := map[string]any{"s": "payload", "n": int64(42)}
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		key := "k:" + itoa(i)
		if _, err := s.Set(key, val, 0); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGet(b *testing.B) {
	s := newStoreB(b, 8, 8<<20, 0)
	defer s.Close()

	val := map[string]any{"s": "payload", "n": int64(42)}
	for i := 0; i < 100_000; i++ {
		if _, err := s.Set("k:"+itoa(i), val, 0); err != nil {
			b.Fatal(err)
		}
	}

	runtime.GC()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = s.Get("k:" + itoa(i%100_000))
	}
}

func BenchmarkExists(b *testing.B) {
	s := newStoreB(b, 8, 8<<20, 0)
	defer s.Close()

	val := map[string]any{"x": 1}
	for i := 0; i < 100_000; i++ {
		if _, err := s.Set("k:"+itoa(i), val, 0); err != nil {
			b.Fatal(err)
		}
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if _, err := s.Exists("k:" + itoa(i%100_000)); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDelete(b *testing.B) {
	s := newStoreB(b, 8, 8<<20, 0)
	defer s.Close()

	val := map[string]any{"x": 1}
	keys := make([]string, 100_000)

	for i := 0; i < len(keys); i++ {
		keys[i] = "k:" + itoa(i)
		if _, err := s.Set(keys[i], val, 0); err != nil {
			b.Fatal(err)
		}
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		k := keys[i%len(keys)]
		if _, err := s.Delete(k); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkAppend(b *testing.B) {
	st := newStoreB(b, 8, 8<<20, 0)
	const key = "append:key:0"

	_, _ = st.Set(key, map[string]any{"s": ""}, 0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = st.Append(key, map[string]string{"s": "x"})
	}
}

func BenchmarkAppend_Parallel(b *testing.B) {
	st := newStoreB(b, 8, 8<<20, 0)

	const nkeys = 256
	for i := 0; i < nkeys; i++ {
		_, _ = st.Set("append:key:"+strconv.Itoa(i), map[string]any{"s": ""}, 0)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			k := "append:key:" + strconv.Itoa(i&(nkeys-1))
			_ = st.Append(k, map[string]string{"s": "x"})
			i++
		}
	})
}

// ------------------------------- Increment -------------------------------

func BenchmarkIncrement(b *testing.B) {
	st := newStoreB(b, 8, 8<<20, 0)
	const key = "incr:key:0"
	_, _ = st.Set(key, map[string]any{"n": int64(0)}, 0)

	offs := map[string]int64{"n": 1}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = st.Increment(key, offs)
	}
}

func BenchmarkIncrement_Parallel(b *testing.B) {
	st := newStoreB(b, 8, 8<<20, 0)
	const nkeys = 256

	for i := 0; i < nkeys; i++ {
		_, _ = st.Set("incr:key:"+strconv.Itoa(i), map[string]any{"n": int64(0)}, 0)
	}

	offs := map[string]int64{"n": 1}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			k := "incr:key:" + strconv.Itoa(i&(nkeys-1))
			_ = st.Increment(k, offs)
			i++
		}
	})
}

// ------------------------------- Decrement -------------------------------

func BenchmarkDecrement(b *testing.B) {
	st := newStoreB(b, 8, 8<<20, 0)
	const key = "decr:key:0"
	_, _ = st.Set(key, map[string]any{"n": int64(0)}, 0)

	offs := map[string]int64{"n": 1}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = st.Decrement(key, offs)
	}
}

func BenchmarkDecrement_Parallel(b *testing.B) {
	st := newStoreB(b, 8, 8<<20, 0)
	const nkeys = 256

	for i := 0; i < nkeys; i++ {
		_, _ = st.Set("decr:key:"+strconv.Itoa(i), map[string]any{"n": int64(0)}, 0)
	}

	offs := map[string]int64{"n": 1}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			k := "decr:key:" + strconv.Itoa(i&(nkeys-1))
			_ = st.Decrement(k, offs)
			i++
		}
	})
}

// ------------------------------- GetTTL -------------------------------

func BenchmarkGetTTL(b *testing.B) {
	st := newStoreB(b, 8, 8<<20, 0)

	const key = "ttl:key:0"
	_, _ = st.Set(key, map[string]any{"n": 1}, 5*time.Minute)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = st.GetTTL(key)
	}
}

func BenchmarkGetTTL_Parallel(b *testing.B) {
	st := newStoreB(b, 8, 8<<20, 0)
	const nkeys = 256

	for i := 0; i < nkeys; i++ {
		_, _ = st.Set("ttl:key:"+strconv.Itoa(i), map[string]any{"n": 1}, 5*time.Minute)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			k := "ttl:key:" + strconv.Itoa(i&(nkeys-1))
			_, _ = st.GetTTL(k)
			i++
		}
	})
}

// ------------------------------- SetTTL -------------------------------

func BenchmarkSetTTL(b *testing.B) {
	st := newStoreB(b, 8, 8<<20, 0)
	const key = "setttl:key:0"
	_, _ = st.Set(key, map[string]any{"n": 1}, 0)

	ttl := 10 * time.Second
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = st.SetTTL(key, ttl)
	}
}

func BenchmarkSetTTL_Parallel(b *testing.B) {
	st := newStoreB(b, 8, 8<<20, 0)
	const nkeys = 256

	for i := 0; i < nkeys; i++ {
		_, _ = st.Set("setttl:key:"+strconv.Itoa(i), map[string]any{"n": 1}, 0)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			k := "setttl:key:" + strconv.Itoa(i&(nkeys-1))
			_ = st.SetTTL(k, 10*time.Second)
			i++
		}
	})
}

// -- helpers

func newStoreB(b *testing.B, shards uint8, budget uint64, janitorEvery time.Duration) *Store {
	b.Helper()

	s, err := New(Options{
		Shards:         shards,
		MaxStorageSize: budget,
		JanitorEvery:   janitorEvery,
	})

	if err != nil {
		b.Fatalf("new store: %v", err)
	}

	return s
}

func itoa(i int) string {
	// tiny, non-allocating-ish fast path for small ints in tests/bench
	if i == 0 {
		return "0"
	}

	buf := [20]byte{}
	pos := len(buf)

	neg := i < 0
	u := uint64(i)

	if neg {
		u = uint64(-i)
	}

	for u > 0 {
		pos--
		// '0' + digit
		buf[pos] = byte('0' + u%10)
		u /= 10
	}

	if neg {
		pos--
		buf[pos] = '-'
	}

	return string(buf[pos:])
}

func newStore1(t *testing.T, capBytes uint64, defaultTTL uint64) *Store {
	t.Helper()
	st, err := New(Options{
		Shards:         1,
		MaxStorageSize: capBytes,
		DefaultTTL:     time.Duration(defaultTTL * uint64(time.Second)),
		JanitorEvery:   0, // disable janitor for deterministic tests
	})

	if err != nil {
		t.Fatalf("New failed: %v", err)
	}
	return st
}

func valWithString(n int) map[string]any {
	return map[string]any{"s": string(make([]byte, n))}
}

func valNum(n int64) map[string]any {
	return map[string]any{"n": n}
}

func m(k string, v any) map[string]any {
	return map[string]any{k: v}
}

func bigString(n int) string {
	b := make([]byte, n)

	for i := range b {
		b[i] = 'x'
	}

	return string(b)
}

//----------- More tests for thorough branching --------------

func TestAppend_Errors_Extra(t *testing.T) {
	s := newStore(t, 2, 1<<20, 0)
	defer s.Close()

	if _, err := s.Set("k1", map[string]any{"name": "neo", "n": 1}, 0); err != nil {
		t.Fatalf("set failed: %v", err)
	}

	if err := s.Append("k1", map[string]string{"missing": "x"}); err != ErrIllegalField {
		t.Fatalf("expected ErrIllegalField, got %v", err)
	}

	if err := s.Append("k1", map[string]string{"n": "x"}); err != ErrBadType {
		t.Fatalf("expected ErrBadType, got %v", err)
	}
}

func TestGetTTL_SetTTL(t *testing.T) {
	s := newStore(t, 2, 1<<20, 0)
	defer s.Close()

	if _, err := s.Set("k1", map[string]any{"name": "neo"}, 0); err != nil {
		t.Fatalf("set failed: %v", err)
	}

	ttl, err := s.GetTTL("k1")
	if err != nil || ttl != -1 {
		t.Fatalf("expected ttl=-1, got ttl=%d err=%v", ttl, err)
	}

	if err := s.SetTTL("k1", 50*time.Millisecond); err != nil {
		t.Fatalf("setttl failed: %v", err)
	}

	if _, err := s.GetTTL("k1"); err != nil {
		t.Fatalf("getttl after set failed: %v", err)
	}

	if err := s.SetTTL("k1", 0); err != nil {
		t.Fatalf("clear ttl failed: %v", err)
	}

	ttl, err = s.GetTTL("k1")
	if err != nil || ttl != -1 {
		t.Fatalf("expected ttl=-1 after clear, got ttl=%d err=%v", ttl, err)
	}

	if err := s.SetTTL("missing", time.Second); err != ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestGetTTL_Expired(t *testing.T) {
	s := newStore(t, 2, 1<<20, 0)
	defer s.Close()

	if _, err := s.Set("k1", map[string]any{"name": "neo"}, 10*time.Millisecond); err != nil {
		t.Fatalf("set failed: %v", err)
	}

	time.Sleep(20 * time.Millisecond)
	if _, err := s.GetTTL("k1"); err != ErrNotFound {
		t.Fatalf("expected ErrNotFound after expiry, got %v", err)
	}
}

func TestJanitor_RemovesExpired(t *testing.T) {
	s := newStore(t, 2, 1<<20, 5*time.Millisecond)
	defer s.Close()

	if _, err := s.Set("k1", map[string]any{"name": "neo"}, 5*time.Millisecond); err != nil {
		t.Fatalf("set failed: %v", err)
	}

	time.Sleep(20 * time.Millisecond)
	ok, err := s.Exists("k1")
	if err != nil {
		t.Fatalf("exists failed: %v", err)
	}
	if ok {
		t.Fatalf("expected expired key to be removed by janitor")
	}
}

func TestNew_TinyLFUError(t *testing.T) {
	_, err := New(Options{
		MaxStorageSize: 1,
		LFUDepth:       -1,
	})
	if err == nil {
		t.Fatalf("expected TinyLFU constructor error")
	}
}

func TestExists_RefreshBetweenReadAndWriteLock(t *testing.T) {
	st := newStore1(t, 1<<20, 0)
	defer st.Close()

	if ok, err := st.Set("k", m("v", "x"), 5*time.Millisecond); err != nil || !ok {
		t.Fatalf("seed failed: ok=%v err=%v", ok, err)
	}

	time.Sleep(10 * time.Millisecond)

	orig := existsBeforeWriteLockHook
	defer func() { existsBeforeWriteLockHook = orig }()
	existsBeforeWriteLockHook = func() {
		shard := st.shardFor("k")
		shard.mu.Lock()
		if rec, ok := shard.data["k"]; ok {
			rec.expiry = time.Now().Add(2 * time.Second)
		}
		shard.mu.Unlock()
	}

	ok, err := st.Exists("k")
	if err != nil || !ok {
		t.Fatalf("expected true after TTL refresh, got ok=%v err=%v", ok, err)
	}
}

func TestSet_UpdateTrimming_NoVictim(t *testing.T) {
	st := newStore1(t, 1<<20, 0)
	defer st.Close()

	if ok, err := st.Set("k", m("v", "x"), 0); err != nil || !ok {
		t.Fatalf("seed failed: ok=%v err=%v", ok, err)
	}

	shard := st.shardFor("k")
	shard.mu.Lock()
	shard.currentSize = shard.capacity + 1
	shard.lru = stdlib.NewLRU[string, struct{}](0, nil)
	shard.mu.Unlock()

	if ok, err := st.Set("k", m("v", "y"), 0); err != nil || !ok {
		t.Fatalf("update failed: ok=%v err=%v", ok, err)
	}
}

func TestSet_InsertRejectWhenFullWithoutVictim(t *testing.T) {
	st := newStore1(t, 1<<20, 0)
	defer st.Close()

	shard := st.shardFor("k")
	shard.mu.Lock()
	shard.currentSize = shard.capacity
	shard.lru = stdlib.NewLRU[string, struct{}](0, nil)
	shard.mu.Unlock()

	ok, err := st.Set("k", m("v", "x"), 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok {
		t.Fatalf("expected reject when shard is full and no victim exists")
	}
}

func TestAppend_RecomputeShrinkAndFullNoVictim(t *testing.T) {
	st := newStore1(t, 1<<20, 0)
	defer st.Close()

	if ok, err := st.Set("k", m("s", "value"), 0); err != nil || !ok {
		t.Fatalf("seed failed: ok=%v err=%v", ok, err)
	}

	shard := st.shardFor("k")
	shard.mu.Lock()
	shard.currentSize = shard.capacity + 100
	shard.lru = stdlib.NewLRU[string, struct{}](0, nil)
	shard.mu.Unlock()

	origSizeHook := getInMemorySizeInBytesHook
	defer func() { getInMemorySizeInBytesHook = origSizeHook }()
	getInMemorySizeInBytesHook = func(v interface{}) (int64, error) {
		return 1, nil
	}

	if err := st.Append("k", map[string]string{"s": ""}); err != nil {
		t.Fatalf("append failed: %v", err)
	}
}

func TestGetTTL_NegativeSecondsPath(t *testing.T) {
	st := newStore1(t, 1<<20, 0)
	defer st.Close()

	if ok, err := st.Set("k", m("v", "x"), time.Hour); err != nil || !ok {
		t.Fatalf("seed failed: ok=%v err=%v", ok, err)
	}

	orig := ttlRemainingSecondsHook
	defer func() { ttlRemainingSecondsHook = orig }()
	ttlRemainingSecondsHook = func(expiry time.Time) int64 { return -1 }

	if _, err := st.GetTTL("k"); err != ErrNotFound {
		t.Fatalf("expected ErrNotFound for negative TTL path, got %v", err)
	}
}

func TestSetTTL_InvalidKeyAndExpiredRemoval(t *testing.T) {
	st := newStore1(t, 1<<20, 0)
	defer st.Close()

	if err := st.SetTTL("", time.Second); err != ErrInvalidKey {
		t.Fatalf("expected ErrInvalidKey, got %v", err)
	}

	if ok, err := st.Set("k", m("v", "x"), 5*time.Millisecond); err != nil || !ok {
		t.Fatalf("seed failed: ok=%v err=%v", ok, err)
	}
	time.Sleep(10 * time.Millisecond)

	if err := st.SetTTL("k", time.Second); err != ErrNotFound {
		t.Fatalf("expected ErrNotFound for expired key, got %v", err)
	}
}

func TestJanitor_MaxDeletePerShardBreak(t *testing.T) {
	orig := maxDeletePerShardHook
	defer func() { maxDeletePerShardHook = orig }()
	maxDeletePerShardHook = func() uint16 { return 1 }

	st, err := New(Options{
		Shards:         1,
		MaxStorageSize: 1 << 24,
		JanitorEvery:   2 * time.Millisecond,
	})
	if err != nil {
		t.Fatalf("new store failed: %v", err)
	}
	defer st.Close()

	total := 300
	for i := 0; i < total; i++ {
		key := fmt.Sprintf("k-%d", i)
		if ok, err := st.Set(key, m("v", i), 1*time.Millisecond); err != nil || !ok {
			t.Fatalf("set %s failed: ok=%v err=%v", key, ok, err)
		}
	}

	time.Sleep(12 * time.Millisecond)

	sh := st.shards[0]
	sh.mu.RLock()
	remaining := len(sh.data)
	sh.mu.RUnlock()

	if remaining == 0 || remaining >= total {
		t.Fatalf("expected partial janitor cleanup due per-shard cap, remaining=%d total=%d", remaining, total)
	}
}
