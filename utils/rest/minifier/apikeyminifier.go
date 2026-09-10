package minifier

import (
	"errors"
	"reflect"
	"sync"
)

var (
	configMu                sync.RWMutex
	isApiKeyMinifierEnabled bool          // Indicates whether the API key minifier is enabled.
	apiKeyMinifierMap       = &sync.Map{} // A map containing the original keys and their minified counterparts.
)

var (
	errMinifierDisabled = errors.New("MinifyApiKeys: API key minifier is not enabled")
	errInvalidInput     = errors.New("MinifyApiKeys: input must be a struct, map[string]interface{}, slice/array, pointer, or interface")
	errMapKeyNotString  = errors.New("MinifyApiKeys: map key is not a string")
	errCycleDetected    = errors.New("MinifyApiKeys: cycle detected in input")
)

type visitKey struct {
	kind reflect.Kind
	ptr  uintptr
}

type minifierEngine struct {
	keyMap   *sync.Map
	visiting map[visitKey]struct{}
}

// SetupApiKeyMinifierConfig configures the API key minifier.
//
// Parameters:
// - isEnabled: A boolean indicating whether the API key minifier should be enabled.
// - keymap: A pointer to a sync.Map containing the original keys and their minified counterparts.
func SetupApiKeyMinifierConfig(isEnabled bool, keymap *sync.Map) {
	configMu.Lock()
	defer configMu.Unlock()
	isApiKeyMinifierEnabled = isEnabled
	if keymap == nil {
		apiKeyMinifierMap = &sync.Map{}
		return
	}
	apiKeyMinifierMap = keymap
}

// IsApiKeyMinifierEnabled returns whether the API key minifier is enabled.
func IsApiKeyMinifierEnabled() bool {
	configMu.RLock()
	defer configMu.RUnlock()
	return isApiKeyMinifierEnabled
}

// MinifyApiKeys recursively minifies the keys of a given input based on the predefined key mapping.
//
// Parameters:
// - input: An interface{} that must be either a struct, a map[string]interface{}, or a nested combination.
//
// Returns:
// - map[string]interface{}: A map with minified keys at all levels.
// - error: An error if the input is not a valid structure or contains unsupported types.
func MinifyApiKeys(input interface{}) (interface{}, error) {
	config := getMinifierConfig()
	if !config.enabled {
		return nil, errMinifierDisabled
	}

	engine := minifierEngine{
		keyMap:   config.keyMap,
		visiting: make(map[visitKey]struct{}),
	}

	minified, err := engine.minifyValue(reflect.ValueOf(input), true)
	if err != nil {
		return nil, err
	}

	return minified, nil
}

// minifyRecursive handles recursion for structs and maps.
func minifyRecursive(input interface{}) (map[string]interface{}, error) {
	config := getMinifierConfig()
	engine := minifierEngine{
		keyMap:   config.keyMap,
		visiting: make(map[visitKey]struct{}),
	}
	out, err := engine.minifyRecursiveValue(reflect.ValueOf(input))
	if err != nil {
		return nil, err
	}

	return out, nil
}

// handleNestedValues checks if a value is a map, struct, or slice and applies recursion.
func handleNestedValues(value interface{}) (interface{}, error) {
	config := getMinifierConfig()
	engine := minifierEngine{
		keyMap:   config.keyMap,
		visiting: make(map[visitKey]struct{}),
	}
	return engine.minifyValue(reflect.ValueOf(value), false)
}

type minifierConfig struct {
	enabled bool
	keyMap  *sync.Map
}

func getMinifierConfig() minifierConfig {
	configMu.RLock()
	defer configMu.RUnlock()
	if apiKeyMinifierMap == nil {
		return minifierConfig{
			enabled: isApiKeyMinifierEnabled,
			keyMap:  &sync.Map{},
		}
	}

	return minifierConfig{
		enabled: isApiKeyMinifierEnabled,
		keyMap:  apiKeyMinifierMap,
	}
}

func (e *minifierEngine) minifyValue(value reflect.Value, isRoot bool) (interface{}, error) {
	if !value.IsValid() {
		if isRoot {
			return nil, errInvalidInput
		}
		return nil, nil
	}

	if value.Kind() == reflect.Interface || value.Kind() == reflect.Pointer {
		if value.IsNil() {
			if isRoot {
				return nil, errInvalidInput
			}
			return nil, nil
		}
		done, err := e.enter(value)
		if err != nil {
			return nil, err
		}

		elem := value.Elem()
		minified, minifyErr := e.minifyValue(elem, isRoot)
		done()

		return minified, minifyErr
	}

	switch value.Kind() {
	case reflect.Struct:
		return e.minifyStruct(value)
	case reflect.Map:
		return e.minifyMap(value)
	case reflect.Slice, reflect.Array:
		return e.minifySlice(value)
	default:
		if isRoot {
			return nil, errInvalidInput
		}
		if !value.CanInterface() {
			return nil, errInvalidInput
		}
		return value.Interface(), nil
	}
}

func (e *minifierEngine) minifyRecursiveValue(value reflect.Value) (map[string]interface{}, error) {
	if !value.IsValid() {
		return nil, errInvalidInput
	}
	for value.Kind() == reflect.Interface || value.Kind() == reflect.Pointer {
		if value.IsNil() {
			return nil, errInvalidInput
		}
		done, err := e.enter(value)
		if err != nil {
			return nil, err
		}
		value = value.Elem()
		done()
	}

	switch value.Kind() {
	case reflect.Struct:
		return e.minifyStruct(value)
	case reflect.Map:
		return e.minifyMap(value)
	default:
		return nil, errInvalidInput
	}
}

func (e *minifierEngine) minifyStruct(value reflect.Value) (map[string]interface{}, error) {
	result := make(map[string]interface{}, value.NumField())
	valueType := value.Type()
	for i := 0; i < value.NumField(); i++ {
		field := valueType.Field(i)
		if field.PkgPath != "" {
			continue
		}

		fieldValue := value.Field(i)
		minifiedFieldValue, err := e.minifyValue(fieldValue, false)
		if err != nil {
			return nil, err
		}

		result[e.minifyKey(field.Name)] = minifiedFieldValue
	}
	return result, nil
}

func (e *minifierEngine) minifyMap(value reflect.Value) (map[string]interface{}, error) {
	if value.Type().Key().Kind() != reflect.String {
		return nil, errMapKeyNotString
	}
	if value.Kind() == reflect.Map && value.IsNil() {
		return nil, nil
	}

	done, err := e.enter(value)
	if err != nil {
		return nil, err
	}
	defer done()

	result := make(map[string]interface{}, value.Len())
	for _, key := range value.MapKeys() {
		fieldName := key.String()
		fieldValue := value.MapIndex(key)
		minifiedFieldValue, minifyErr := e.minifyValue(fieldValue, false)
		if minifyErr != nil {
			return nil, minifyErr
		}
		result[e.minifyKey(fieldName)] = minifiedFieldValue
	}
	return result, nil
}

func (e *minifierEngine) minifySlice(value reflect.Value) ([]interface{}, error) {
	if value.Kind() == reflect.Slice && value.IsNil() {
		return nil, nil
	}

	done, err := e.enter(value)
	if err != nil {
		return nil, err
	}
	defer done()

	out := make([]interface{}, value.Len())
	for i := 0; i < value.Len(); i++ {
		minified, minifyErr := e.minifyValue(value.Index(i), false)
		if minifyErr != nil {
			return nil, minifyErr
		}
		out[i] = minified
	}
	return out, nil
}

func (e *minifierEngine) minifyKey(originalKey string) string {
	if e.keyMap == nil {
		return originalKey
	}

	value, ok := e.keyMap.Load(originalKey)
	if !ok {
		return originalKey
	}

	mappedKey, ok := value.(string)
	if !ok || mappedKey == "" {
		return originalKey
	}

	return mappedKey
}

func (e *minifierEngine) enter(value reflect.Value) (func(), error) {
	switch value.Kind() {
	case reflect.Map, reflect.Slice, reflect.Pointer:
	default:
		return func() {}, nil
	}

	if value.IsNil() {
		return func() {}, nil
	}

	ptr := value.Pointer()
	key := visitKey{
		kind: value.Kind(),
		ptr:  ptr,
	}
	if _, exists := e.visiting[key]; exists {
		return nil, errCycleDetected
	}
	e.visiting[key] = struct{}{}

	return func() {
		delete(e.visiting, key)
	}, nil
}
