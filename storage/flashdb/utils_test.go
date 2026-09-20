package flashdb

import (
	"errors"
	"math"
	"reflect"
	"testing"
)

func TestClone_Nil(t *testing.T) {
	if out := clone(nil); out != nil {
		t.Fatalf("clone(nil) = %#v, want nil", out)
	}
}

func TestClone_ShallowCopy(t *testing.T) {
	inner := map[string]any{"x": 1}
	src := map[string]any{
		"a": 42,
		"b": "str",
		"c": inner,
	}
	cp := clone(src)

	if len(cp) != len(src) {
		t.Fatalf("len(clone)=%d, want %d", len(cp), len(src))
	}
	if cp["a"].(int) != 42 || cp["b"].(string) != "str" {
		t.Fatalf("clone values mismatch: %#v", cp)
	}

	src["a"] = 99
	if cp["a"].(int) != 42 {
		t.Fatalf("expected shallow copy (top-level map independent)")
	}

	srcPtr := reflect.ValueOf(src["c"]).Pointer()
	cpPtr := reflect.ValueOf(cp["c"]).Pointer()
	if srcPtr != cpPtr {
		t.Fatalf("expected inner map identity to be shared (shallow copy)")
	}

	inner["x"] = 2
	if cp["c"].(map[string]any)["x"].(int) != 2 {
		t.Fatalf("inner mutation should reflect in shallow copy")
	}
}

func TestAddInt_Int_OK(t *testing.T) {
	v, ok := addInt(int(10), 5)
	if !ok || v.(int) != 15 {
		t.Fatalf("addInt(int, +5) got (%v,%v) want (15,true)", v, ok)
	}
}

func TestAddInt_Int_Overflow_Positive(t *testing.T) {
	v, ok := addInt(int(math.MaxInt), 10000000)
	if ok {
		t.Fatalf("expected overflow failure, got v=%v ok=%v", v, ok)
	}
}

func TestAddInt_Int_Overflow_Negative(t *testing.T) {
	v, ok := addInt(int(math.MinInt), -1)
	if ok {
		t.Fatalf("expected underflow failure, got v=%v ok=%v", v, ok)
	}
}

func TestAddInt_Int32_OK(t *testing.T) {
	v, ok := addInt(int32(100), -50)
	if !ok || v.(int32) != 50 {
		t.Fatalf("addInt(int32, -50) got (%v,%v) want (50,true)", v, ok)
	}
}

func TestAddInt_Int32_Overflow(t *testing.T) {
	if _, ok := addInt(int32(math.MaxInt32), 1); ok {
		t.Fatalf("expected int32 overflow on +1")
	}

	if _, ok := addInt(int32(math.MinInt32), -1); ok {
		t.Fatalf("expected int32 underflow on -1")
	}
}

func TestNumericConversionHelpers(t *testing.T) {
	if got := nonNegativeInt64ToUint64(-1); got != 0 {
		t.Fatalf("negative size should clamp to 0, got %d", got)
	}
	if got := nonNegativeInt64ToUint64(7); got != 7 {
		t.Fatalf("positive size = %d, want 7", got)
	}
	if _, ok := int64ToInt32(int64(math.MaxInt32) + 1); ok {
		t.Fatalf("expected int32 conversion failure")
	}
	if got, ok := int64ToInt32(12); !ok || got != 12 {
		t.Fatalf("int32 conversion = %d,%v; want 12,true", got, ok)
	}
}

func TestAddInt_Int64_OK(t *testing.T) {
	v, ok := addInt(int64(7), 3)
	if !ok || v.(int64) != 10 {
		t.Fatalf("addInt(int64, +3) got (%v,%v) want (10,true)", v, ok)
	}
}

func TestAddInt_Float64_OK(t *testing.T) {
	v, ok := addInt(float64(1.5), 2)

	if !ok {
		t.Fatalf("float64 add failed")
	}

	if got := v.(float64); got != 3.5 {
		t.Fatalf("addInt(float64, +2)=%v want 3.5", got)
	}
}

func TestAddInt_UnsupportedType(t *testing.T) {
	if v, ok := addInt("nope", 1); ok || v != nil {
		t.Fatalf("expected (nil,false) for unsupported type, got (%v,%v)", v, ok)
	}
}

func TestSafeDeepSize_Success(t *testing.T) {
	v := map[string]any{"a": 1, "b": "x"}
	sz := safeDeepSize(v)

	if sz < 0 {
		t.Fatalf("safeDeepSize returned negative: %d", sz)
	}
}

func TestSafeDeepSize_ErrorPath(t *testing.T) {
	// Give it a type likely unsupported by a reflect-based size estimator (e.g. channel),
	// which should lead to (size,err) -> err != nil and safeDeepSize -> 0.
	ch := make(chan int)
	sz := safeDeepSize(ch)

	if sz != 0 {
		t.Fatalf("safeDeepSize on unsupported should be 0, got %d", sz)
	}
}

func TestSafeDeepSize_PanicRecovery(t *testing.T) {
	orig := getInMemorySizeInBytesHook
	defer func() { getInMemorySizeInBytesHook = orig }()
	getInMemorySizeInBytesHook = func(v interface{}) (int64, error) {
		panic("boom")
	}

	if sz := safeDeepSize(map[string]any{"a": 1}); sz != 0 {
		t.Fatalf("expected 0 on panic recovery, got %d", sz)
	}
}

func TestSafeDeepSize_NegativeSize(t *testing.T) {
	orig := getInMemorySizeInBytesHook
	defer func() { getInMemorySizeInBytesHook = orig }()
	getInMemorySizeInBytesHook = func(v interface{}) (int64, error) {
		return -1, nil
	}

	if sz := safeDeepSize(map[string]any{"a": 1}); sz != 0 {
		t.Fatalf("expected 0 on negative size, got %d", sz)
	}
}

func TestSafeDeepSize_HookError(t *testing.T) {
	orig := getInMemorySizeInBytesHook
	defer func() { getInMemorySizeInBytesHook = orig }()
	getInMemorySizeInBytesHook = func(v interface{}) (int64, error) {
		return 0, errors.New("size error")
	}

	if sz := safeDeepSize(map[string]any{"a": 1}); sz != 0 {
		t.Fatalf("expected 0 on hook error, got %d", sz)
	}
}

func TestComputeSize_WithValueAndKey(t *testing.T) {
	v := map[string]any{"k": "v"}
	s1 := computeSize("a", v)
	s2 := computeSize("abc", v)

	if s2 <= s1 {
		t.Fatalf("expected size to grow with key length: s1=%d s2=%d", s1, s2)
	}

	if int(s2-s1) != len("abc")-len("a") {
		t.Fatalf("expected delta == key len diff, got %d", s2-s1)
	}
}

func TestComputeSize_NilValue(t *testing.T) {
	k1 := "k"
	k2 := "longer-key"
	s1 := computeSize(k1, nil)
	s2 := computeSize(k2, nil)

	if s1 <= 0 {
		t.Fatalf("computeSize(nil) should be positive, got %d", s1)
	}

	wantDelta := int64(len(k2) - len(k1))
	if s2-s1 != wantDelta {
		t.Fatalf("expected delta %d from key length, got %d (s1=%d s2=%d)",
			wantDelta, s2-s1, s1, s2)
	}
}
