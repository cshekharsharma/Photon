package types

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

func PopulateStructFromMap(dst interface{}, src map[string]interface{}) error {
	cleaned := normalizeInterfaceData(src).(map[string]interface{})
	// Ensure dst is a pointer to a struct
	val := reflect.ValueOf(dst)

	if val.Kind() != reflect.Pointer || val.IsNil() {
		return errors.New("destination must be a non-nil pointer to a struct")
	}

	val = val.Elem()
	if val.Kind() != reflect.Struct {
		return errors.New("destination must be a pointer to a struct")
	}

	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		structField := typ.Field(i)

		if !field.CanSet() {
			continue
		}

		// Use struct tag "json" to find map key

		tag := structField.Tag.Get("json")
		mapKey := structField.Name

		if tag != "" {
			if commaIdx := strings.Index(tag, ","); commaIdx != -1 {
				tag = tag[:commaIdx]
			}

			if tag != "" {
				mapKey = tag
			}
		}

		// Skip if key not in map
		rawVal, ok := cleaned[mapKey]
		if !ok {
			continue
		}

		err := setValueForStructToMap(field, reflect.ValueOf(rawVal))
		if err != nil {
			return fmt.Errorf("failed to set field '%s': %w", structField.Name, err)
		}
	}

	return nil
}

func StructToMap(data interface{}) (map[string]interface{}, error) {
	v := reflect.ValueOf(data)
	if v.Kind() == reflect.Pointer {
		v = v.Elem() // Dereference the pointer if it's a pointer
	}

	if v.Kind() != reflect.Struct {
		return nil, fmt.Errorf("expected a struct or a pointer to a struct, got %v", v.Kind())
	}

	result := make(map[string]interface{})
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		value := v.Field(i)

		// Only consider exported fields
		if field.PkgPath == "" {
			result[field.Name] = value.Interface()
		}
	}

	return result, nil
}

func setValueForStructToMap(field reflect.Value, value reflect.Value) error {
	if !value.IsValid() {
		field.Set(reflect.Zero(field.Type()))
		return nil
	}

	fieldType := field.Type()
	if fieldType.Kind() == reflect.Pointer {
		field, fieldType = ensurePointerField(field)
	}

	if handled, err := setMapField(field, fieldType, value); handled || err != nil {
		return err
	}

	if handled, err := setNestedStructField(field, fieldType, value); handled || err != nil {
		return err
	}

	if value.Kind() == reflect.String {
		return setStringField(field, fieldType, value.String())
	}

	if setSpecialNumericField(field, fieldType, value) {
		return nil
	}

	if value.Type().ConvertibleTo(fieldType) {
		field.Set(value.Convert(fieldType))
		return nil
	}

	return fmt.Errorf("type mismatch: cannot convert from %v to %v", value.Type(), fieldType)
}

func ensurePointerField(field reflect.Value) (reflect.Value, reflect.Type) {
	fieldType := field.Type().Elem()
	if field.IsNil() {
		field.Set(reflect.New(fieldType))
	}
	return field.Elem(), fieldType
}

func setMapField(field reflect.Value, fieldType reflect.Type, value reflect.Value) (bool, error) {
	if fieldType.Kind() != reflect.Map || fieldType.Key().Kind() != reflect.String || value.Kind() != reflect.Map {
		return false, nil
	}

	rawMap, ok := value.Interface().(map[string]interface{})
	if !ok {
		return false, nil
	}

	newMap, err := convertStringInterfaceMap(rawMap, fieldType)
	if err != nil {
		return true, err
	}
	field.Set(newMap)
	return true, nil
}

func convertStringInterfaceMap(rawMap map[string]interface{}, fieldType reflect.Type) (reflect.Value, error) {
	elemType := fieldType.Elem()
	newMap := reflect.MakeMap(fieldType)

	for k, v := range rawMap {
		elemVal, err := convertStructMapElement(k, v, elemType)
		if err != nil {
			return newMap, err
		}
		newMap.SetMapIndex(reflect.ValueOf(k), elemVal)
	}

	return newMap, nil
}

func convertStructMapElement(key string, rawValue interface{}, elemType reflect.Type) (reflect.Value, error) {
	elemVal := reflect.New(elemType).Elem()
	if elemType.Kind() == reflect.Struct {
		if subMap, ok := rawValue.(map[string]interface{}); ok {
			if err := PopulateStructFromMap(elemVal.Addr().Interface(), subMap); err != nil {
				return elemVal, fmt.Errorf("failed to populate map struct for key '%s': %w", key, err)
			}
		}
		return elemVal, nil
	}

	val := reflect.ValueOf(rawValue)
	if val.Type().ConvertibleTo(elemType) {
		elemVal.Set(val.Convert(elemType))
	}
	return elemVal, nil
}

func setNestedStructField(field reflect.Value, fieldType reflect.Type, value reflect.Value) (bool, error) {
	if fieldType.Kind() != reflect.Struct || value.Kind() != reflect.Map {
		return false, nil
	}
	mapVal, ok := value.Interface().(map[string]interface{})
	if !ok {
		return false, nil
	}
	return true, PopulateStructFromMap(field.Addr().Interface(), mapVal)
}

func setStringField(field reflect.Value, fieldType reflect.Type, strValue string) error {
	switch fieldType.Kind() {
	case reflect.String:
		field.SetString(strValue)
		return nil
	case reflect.Bool:
		boolVal, err := strconv.ParseBool(strValue)
		if err != nil {
			return fmt.Errorf("could not parse '%v' as bool: %v", strValue, err)
		}
		field.SetBool(boolVal)
		return nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		intVal, err := strconv.ParseInt(strValue, 10, 64)
		if err != nil {
			return fmt.Errorf("could not convert '%v' to int: %v", strValue, err)
		}
		field.SetInt(intVal)
		return nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		uintVal, err := strconv.ParseUint(strValue, 10, 64)
		if err != nil {
			return fmt.Errorf("could not convert '%v' to uint: %v", strValue, err)
		}
		field.SetUint(uintVal)
		return nil
	case reflect.Float32, reflect.Float64:
		floatVal, err := strconv.ParseFloat(strValue, 64)
		if err != nil {
			return fmt.Errorf("could not convert '%v' to float: %v", strValue, err)
		}
		field.SetFloat(floatVal)
		return nil
	case reflect.Struct:
		return setStringStructField(field, strValue)
	default:
		return fmt.Errorf("type mismatch: cannot convert from string to %v", fieldType)
	}
}

func setStringStructField(field reflect.Value, strValue string) error {
	if field.Type() != reflect.TypeOf(time.Time{}) {
		return fmt.Errorf("type mismatch: cannot convert from string to %v", field.Type())
	}
	parsedTime, err := time.Parse(time.DateTime, strValue)
	if err != nil {
		return fmt.Errorf("could not parse '%v' as time: %v", strValue, err)
	}
	field.Set(reflect.ValueOf(parsedTime))
	return nil
}

func setSpecialNumericField(field reflect.Value, fieldType reflect.Type, value reflect.Value) bool {
	if value.Kind() == reflect.Float64 && fieldType.Kind() == reflect.Int {
		field.SetInt(int64(value.Float()))
		return true
	}
	if fieldType.Kind() == reflect.Bool && value.Kind() == reflect.Int {
		field.SetBool(value.Int() != 0)
		return true
	}
	return false
}

// convertMapToTypedMap converts a map[string]V or map[string]interface{} to map[K]V
func ConvertMapToTypedMap[K ~string, V any](raw interface{}) map[K]V {
	result := make(map[K]V)
	if raw == nil {
		return result
	}

	switch m := raw.(type) {
	case map[K]V:
		return m
	case map[string]V:
		for k, v := range m {
			result[K(k)] = v
		}
	case map[string]interface{}:
		expectedType := reflect.TypeOf((*V)(nil)).Elem()
		for k, v := range m {
			if reflect.TypeOf(v) == expectedType {
				result[K(k)] = v.(V)
			}
		}
	}

	return result
}

// normalizeInterfaceData recursively normalizes the input interface, removing nil values and converting
// slices to maps if necessary.
func normalizeInterfaceData(i interface{}) interface{} {
	switch v := i.(type) {
	case map[interface{}]interface{}:
		m2 := map[string]interface{}{}
		for key, val := range v {
			keyStr, ok := key.(string)
			if !ok {
				keyStr = fmt.Sprintf("%v", key)
			}
			m2[keyStr] = normalizeInterfaceData(val)
		}
		return m2
	case map[string]interface{}:
		m2 := map[string]interface{}{}
		for key, val := range v {
			m2[key] = normalizeInterfaceData(val)
		}
		return m2
	case []interface{}:
		for i, elem := range v {
			v[i] = normalizeInterfaceData(elem)
		}
		return v
	default:
		return v
	}
}
