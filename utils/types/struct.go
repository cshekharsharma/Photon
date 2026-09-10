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
		fieldType = fieldType.Elem()
		if field.IsNil() {
			field.Set(reflect.New(fieldType))
		}
		field = field.Elem()
	}

	//handle the map[string]V to map[K]V conversion
	if fieldType.Kind() == reflect.Map &&
		fieldType.Key().Kind() == reflect.String &&
		value.Kind() == reflect.Map {

		rawMap, ok := value.Interface().(map[string]interface{})
		if ok {
			elemType := fieldType.Elem()
			newMap := reflect.MakeMap(fieldType)

			for k, v := range rawMap {
				elemVal := reflect.New(elemType).Elem()
				switch elemType.Kind() {
				case reflect.Struct:
					if subMap, ok := v.(map[string]interface{}); ok {
						// recursively populate nested struct
						if err := PopulateStructFromMap(elemVal.Addr().Interface(), subMap); err != nil {
							return fmt.Errorf("failed to populate map struct for key '%s': %w", k, err)
						}
					}
				default:
					val := reflect.ValueOf(v)
					if val.Type().ConvertibleTo(elemType) {
						elemVal.Set(val.Convert(elemType))
					}
				}
				newMap.SetMapIndex(reflect.ValueOf(k), elemVal)
			}

			field.Set(newMap)
			return nil
		}
	}

	// Handle nested struct
	if fieldType.Kind() == reflect.Struct && value.Kind() == reflect.Map {
		mapVal, ok := value.Interface().(map[string]interface{})
		if ok {
			return PopulateStructFromMap(field.Addr().Interface(), mapVal)
		}
	}

	// Handle string input
	if value.Kind() == reflect.String {
		strValue := value.String()

		switch fieldType.Kind() {
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
			if field.Type() == reflect.TypeOf(time.Time{}) {
				parsedTime, err := time.Parse(time.DateTime, strValue)
				if err != nil {
					return fmt.Errorf("could not parse '%v' as time: %v", strValue, err)
				}
				field.Set(reflect.ValueOf(parsedTime))
				return nil
			}
		}
	}

	// float64 to int
	if value.Kind() == reflect.Float64 && fieldType.Kind() == reflect.Int {
		field.SetInt(int64(value.Float()))
		return nil
	}

	// int to bool
	if fieldType.Kind() == reflect.Bool && value.Kind() == reflect.Int {
		field.SetBool(value.Int() != 0)
		return nil
	}

	if value.Type().ConvertibleTo(fieldType) {
		field.Set(value.Convert(fieldType))
	} else {
		return fmt.Errorf("type mismatch: cannot convert from %v to %v", value.Type(), fieldType)
	}

	return nil
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
