package minifier

import (
	"reflect"
	"sync"
	"testing"
	"time"
)

type profile struct {
	Bio string
}

type person struct {
	FirstName string
	LastName  string
	Profile   profile
	Tags      []string
	hidden    string
}

type cyclicNode struct {
	Name string
	Next *cyclicNode
}

type holderWithNilPointer struct {
	Ptr *person
}

func buildKeyMap() *sync.Map {
	m := &sync.Map{}
	m.Store("FirstName", "fn")
	m.Store("LastName", "ln")
	m.Store("Profile", "p")
	m.Store("Bio", "b")
	m.Store("Tags", "t")
	m.Store("Name", "n")
	m.Store("Next", "nx")
	m.Store("Address", "a")
	m.Store("City", "c")
	m.Store("Street", "s")
	m.Store("", "") // ensure empty mapping is ignored
	return m
}

func setupEnabled(t *testing.T) {
	t.Helper()
	SetupApiKeyMinifierConfig(true, buildKeyMap())
}

func TestSetupApiKeyMinifierConfigAndIsEnabled(t *testing.T) {
	SetupApiKeyMinifierConfig(false, nil)
	if IsApiKeyMinifierEnabled() {
		t.Fatalf("expected minifier disabled")
	}

	SetupApiKeyMinifierConfig(true, nil)
	if !IsApiKeyMinifierEnabled() {
		t.Fatalf("expected minifier enabled")
	}

	out, err := MinifyApiKeys(map[string]interface{}{"FirstName": "John"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := map[string]interface{}{"FirstName": "John"}
	if !reflect.DeepEqual(out, expected) {
		t.Fatalf("expected %v, got %v", expected, out)
	}
}

func TestMinifyApiKeys_SuccessCases(t *testing.T) {
	setupEnabled(t)

	nested := map[string]interface{}{
		"FirstName": "John",
		"Address": map[string]interface{}{
			"City":   "New York",
			"Street": "5th Avenue",
		},
		"Tags": []interface{}{
			"one",
			map[string]interface{}{"LastName": "Doe"},
			[]interface{}{map[string]interface{}{"FirstName": "Nested"}},
		},
	}

	out, err := MinifyApiKeys(nested)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := map[string]interface{}{
		"fn": "John",
		"a": map[string]interface{}{
			"c": "New York",
			"s": "5th Avenue",
		},
		"t": []interface{}{
			"one",
			map[string]interface{}{"ln": "Doe"},
			[]interface{}{map[string]interface{}{"fn": "Nested"}},
		},
	}
	if !reflect.DeepEqual(out, expected) {
		t.Fatalf("expected %v, got %v", expected, out)
	}
}

func TestMinifyApiKeys_TopLevelArrayAndPointerAndInterface(t *testing.T) {
	setupEnabled(t)

	arrOut, err := MinifyApiKeys([2]map[string]interface{}{
		{"FirstName": "A"},
		{"LastName": "B"},
	})
	if err != nil {
		t.Fatalf("unexpected array error: %v", err)
	}
	arrExpected := []interface{}{
		map[string]interface{}{"fn": "A"},
		map[string]interface{}{"ln": "B"},
	}
	if !reflect.DeepEqual(arrOut, arrExpected) {
		t.Fatalf("expected %v, got %v", arrExpected, arrOut)
	}

	p := &person{FirstName: "John", LastName: "Doe", Profile: profile{Bio: "hello"}, Tags: []string{"x"}, hidden: "skip"}
	ptrOut, err := MinifyApiKeys(p)
	if err != nil {
		t.Fatalf("unexpected pointer error: %v", err)
	}
	ptrExpected := map[string]interface{}{
		"fn": "John",
		"ln": "Doe",
		"p":  map[string]interface{}{"b": "hello"},
		"t":  []interface{}{"x"},
	}
	if !reflect.DeepEqual(ptrOut, ptrExpected) {
		t.Fatalf("expected %v, got %v", ptrExpected, ptrOut)
	}

	var iface interface{} = map[string]interface{}{"FirstName": "Iface"}
	ifaceOut, err := MinifyApiKeys(iface)
	if err != nil {
		t.Fatalf("unexpected interface error: %v", err)
	}
	ifaceExpected := map[string]interface{}{"fn": "Iface"}
	if !reflect.DeepEqual(ifaceOut, ifaceExpected) {
		t.Fatalf("expected %v, got %v", ifaceExpected, ifaceOut)
	}
}

func TestMinifyApiKeys_PrimitivesWithinSlicesArePreserved(t *testing.T) {
	setupEnabled(t)

	out, err := MinifyApiKeys([]interface{}{
		"hello",
		123,
		true,
		map[string]interface{}{"FirstName": "John"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := []interface{}{
		"hello",
		123,
		true,
		map[string]interface{}{"fn": "John"},
	}
	if !reflect.DeepEqual(out, expected) {
		t.Fatalf("expected %v, got %v", expected, out)
	}
}

func TestMinifyApiKeys_NestedNilPointerIsPreserved(t *testing.T) {
	setupEnabled(t)

	out, err := MinifyApiKeys(holderWithNilPointer{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := map[string]interface{}{"Ptr": nil}
	if !reflect.DeepEqual(out, expected) {
		t.Fatalf("expected %v, got %v", expected, out)
	}
}

func TestMinifyApiKeys_ErrorCases(t *testing.T) {
	SetupApiKeyMinifierConfig(false, buildKeyMap())
	if _, err := MinifyApiKeys(map[string]interface{}{"FirstName": "John"}); err == nil {
		t.Fatalf("expected disabled error")
	}

	setupEnabled(t)
	if _, err := MinifyApiKeys(nil); err == nil {
		t.Fatalf("expected nil root error")
	}
	if _, err := MinifyApiKeys("string"); err == nil {
		t.Fatalf("expected invalid root type error")
	}
	if _, err := MinifyApiKeys(map[interface{}]interface{}{"x": 1}); err == nil {
		t.Fatalf("expected non-string map key error")
	}
	if _, err := MinifyApiKeys([]interface{}{map[interface{}]interface{}{"x": 1}}); err == nil {
		t.Fatalf("expected nested non-string map key error")
	}

	var nilPtr *person
	if _, err := MinifyApiKeys(nilPtr); err == nil {
		t.Fatalf("expected nil pointer root error")
	}
	var nilIface interface{}
	if _, err := MinifyApiKeys(nilIface); err == nil {
		t.Fatalf("expected nil interface root error")
	}
}

func TestMinifyApiKeys_CycleDetection(t *testing.T) {
	setupEnabled(t)

	n := &cyclicNode{Name: "root"}
	n.Next = n

	if _, err := MinifyApiKeys(n); err == nil {
		t.Fatalf("expected cycle detection error for struct pointer cycle")
	}

	m := map[string]interface{}{}
	m["self"] = m
	if _, err := MinifyApiKeys(m); err == nil {
		t.Fatalf("expected cycle detection error for map cycle")
	}

	s := make([]interface{}, 1)
	s[0] = s
	if _, err := MinifyApiKeys(s); err == nil {
		t.Fatalf("expected cycle detection error for slice cycle")
	}
}

func TestMinifyRecursive_SuccessAndErrors(t *testing.T) {
	setupEnabled(t)

	out, err := minifyRecursive(map[string]interface{}{"FirstName": "John"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := map[string]interface{}{"fn": "John"}
	if !reflect.DeepEqual(out, expected) {
		t.Fatalf("expected %v, got %v", expected, out)
	}

	p := &person{FirstName: "Ptr"}
	out, err = minifyRecursive(p)
	if err != nil {
		t.Fatalf("unexpected pointer error: %v", err)
	}
	if out["fn"] != "Ptr" {
		t.Fatalf("expected fn=Ptr, got %v", out["fn"])
	}

	if _, err := minifyRecursive(123); err == nil {
		t.Fatalf("expected invalid input error")
	}
	if _, err := minifyRecursive((*person)(nil)); err == nil {
		t.Fatalf("expected nil pointer error")
	}
}

func TestHandleNestedValues_AllBranches(t *testing.T) {
	setupEnabled(t)

	out, err := handleNestedValues(map[string]interface{}{"FirstName": "John"})
	if err != nil {
		t.Fatalf("unexpected map error: %v", err)
	}
	if !reflect.DeepEqual(out, map[string]interface{}{"fn": "John"}) {
		t.Fatalf("unexpected map output: %v", out)
	}

	out, err = handleNestedValues([]interface{}{map[string]interface{}{"LastName": "Doe"}, 3})
	if err != nil {
		t.Fatalf("unexpected slice error: %v", err)
	}
	if !reflect.DeepEqual(out, []interface{}{map[string]interface{}{"ln": "Doe"}, 3}) {
		t.Fatalf("unexpected slice output: %v", out)
	}

	out, err = handleNestedValues("plain")
	if err != nil {
		t.Fatalf("unexpected primitive error: %v", err)
	}
	if out != "plain" {
		t.Fatalf("expected plain output, got %v", out)
	}

	if _, err := handleNestedValues(map[interface{}]interface{}{"x": 1}); err == nil {
		t.Fatalf("expected non-string map key error")
	}

	var nilIface interface{}
	out, err = handleNestedValues(nilIface)
	if err != nil {
		t.Fatalf("unexpected nil interface error: %v", err)
	}
	if out != nil {
		t.Fatalf("expected nil output, got %v", out)
	}
}

func TestEngineEnterAndMinifyKeyBranches(t *testing.T) {
	engine := &minifierEngine{keyMap: nil, visiting: map[visitKey]struct{}{}}

	if got := engine.minifyKey("FirstName"); got != "FirstName" {
		t.Fatalf("expected original key, got %s", got)
	}

	km := &sync.Map{}
	km.Store("FirstName", 123)
	engine.keyMap = km
	if got := engine.minifyKey("FirstName"); got != "FirstName" {
		t.Fatalf("expected original key for non-string mapping, got %s", got)
	}

	km.Store("FirstName", "")
	if got := engine.minifyKey("FirstName"); got != "FirstName" {
		t.Fatalf("expected original key for empty mapping, got %s", got)
	}

	km.Store("FirstName", "fn")
	if got := engine.minifyKey("FirstName"); got != "fn" {
		t.Fatalf("expected mapped key fn, got %s", got)
	}

	done, err := engine.enter(reflect.ValueOf(123))
	if err != nil {
		t.Fatalf("unexpected error for non-trackable kind: %v", err)
	}
	done()

	var nilMap map[string]interface{}
	done, err = engine.enter(reflect.ValueOf(nilMap))
	if err != nil {
		t.Fatalf("unexpected error for nil map: %v", err)
	}
	done()

	mp := map[string]interface{}{"k": "v"}
	done, err = engine.enter(reflect.ValueOf(mp))
	if err != nil {
		t.Fatalf("unexpected error entering map: %v", err)
	}
	if _, err = engine.enter(reflect.ValueOf(mp)); err == nil {
		t.Fatalf("expected cycle error for duplicate enter")
	}
	done()

	type rec struct{ Next *rec }
	var ptr *rec
	done, err = engine.enter(reflect.ValueOf(ptr))
	if err != nil {
		t.Fatalf("unexpected error for nil ptr: %v", err)
	}
	done()
}

func TestGetMinifierConfig_NilMapFallback(t *testing.T) {
	configMu.Lock()
	isApiKeyMinifierEnabled = true
	apiKeyMinifierMap = nil
	configMu.Unlock()

	cfg := getMinifierConfig()
	if !cfg.enabled {
		t.Fatalf("expected enabled true")
	}
	if cfg.keyMap == nil {
		t.Fatalf("expected non-nil fallback key map")
	}

	SetupApiKeyMinifierConfig(false, nil)
}

func TestEngineInternalBranches(t *testing.T) {
	engine := &minifierEngine{
		keyMap:   buildKeyMap(),
		visiting: map[visitKey]struct{}{},
	}

	out, err := engine.minifyValue(reflect.Value{}, false)
	if err != nil {
		t.Fatalf("unexpected error for invalid non-root value: %v", err)
	}
	if out != nil {
		t.Fatalf("expected nil output for invalid non-root value, got %v", out)
	}

	var nilMap map[string]interface{}
	outMap, err := engine.minifyMap(reflect.ValueOf(nilMap))
	if err != nil {
		t.Fatalf("unexpected error for nil map: %v", err)
	}
	if outMap != nil {
		t.Fatalf("expected nil map output, got %v", outMap)
	}

	noIfaceField := reflect.ValueOf(time.Time{}).Field(0)
	if noIfaceField.CanInterface() {
		t.Fatalf("expected non-interfaceable field")
	}
	if _, err := engine.minifyValue(noIfaceField, false); err == nil {
		t.Fatalf("expected error for non-interfaceable value")
	}

	if _, err := engine.minifyRecursiveValue(reflect.Value{}); err == nil {
		t.Fatalf("expected invalid recursive value error")
	}

	ptrInput := &person{FirstName: "x"}
	ptrValue := reflect.ValueOf(ptrInput)
	engine.visiting[visitKey{kind: reflect.Pointer, ptr: ptrValue.Pointer()}] = struct{}{}
	if _, err := engine.minifyRecursiveValue(ptrValue); err == nil {
		t.Fatalf("expected enter error for pre-visited pointer")
	}
}
