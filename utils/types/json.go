package types

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

// A helper function to convert a field value to the corresponding struct
func convertToStruct(fieldValue interface{}, target interface{}) error {
	data, err := json.Marshal(fieldValue)
	if err != nil {
		return fmt.Errorf("error marshalling field value: %v", err)
	}

	if err := UnmarshalCustom(data, target); err != nil {
		return fmt.Errorf("error unmarshalling field value into target struct: %v", err)
	}
	return nil
}

// Generalized Unmarshal Function
func UnmarshalCustom(data []byte, out interface{}) error {
	var tempMap map[string]interface{}
	if err := json.Unmarshal(data, &tempMap); err != nil {
		return fmt.Errorf("error unmarshalling into map: %v", err)
	}

	v := reflect.ValueOf(out).Elem() // Get the struct as a reflect.Value
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)
		fieldName := fieldType.Name

		jsonTag := fieldType.Tag.Get("json")
		if jsonTag == "-" {
			continue
		}

		tagParts := strings.Split(jsonTag, ",")
		jsonFieldName := tagParts[0]
		omitempty := len(tagParts) > 1 && tagParts[1] == "omitempty"

		fieldValue, exists := tempMap[jsonFieldName]
		if !exists {
			if omitempty {
				continue
			}
			return fmt.Errorf("missing required field '%s' in JSON", jsonFieldName)
		}

		if err := convertField(field, fieldType, fieldValue); err != nil {
			return fmt.Errorf("error converting field %s: %v", fieldName, err)
		}
	}

	return nil
}

// Convert a field's value based on its type
func convertField(field reflect.Value, fieldType reflect.StructField, fieldValue interface{}) error {
	fieldKind := field.Kind()

	if fieldKind == reflect.Pointer {
		if field.IsNil() {
			field.Set(reflect.New(fieldType.Type.Elem()))
		}
		return convertField(field.Elem(), fieldType, fieldValue)
	}

	if fieldKind == reflect.Map {
		mapType := field.Type()
		mapKeyType := mapType.Key()
		mapElemType := mapType.Elem()

		rawData, ok := fieldValue.(map[string]interface{})
		if !ok {
			return fmt.Errorf("expected map for field, but got %T", fieldValue)
		}

		convertedMap := reflect.MakeMap(mapType)
		for key, value := range rawData {
			mapKey := reflect.ValueOf(key).Convert(mapKeyType)
			mapValue := reflect.New(mapElemType).Elem()

			// Handle cases where the map element type is interface{}
			if mapElemType.Kind() == reflect.Interface {
				mapValue.Set(reflect.ValueOf(value))
			} else if isPrimitive(mapElemType.Kind()) {
				if reflect.TypeOf(value).Kind() == mapElemType.Kind() {
					mapValue.Set(reflect.ValueOf(value).Convert(mapElemType))
				} else if mapElemType.Kind() == reflect.Int64 && reflect.TypeOf(value).Kind() == reflect.Float64 {
					// Special handling for float64 to int64 conversion if necessary
					mapValue.SetInt(int64(value.(float64)))
				} else {
					return fmt.Errorf("type mismatch for key %v in map, expected %v but got %T", key, mapElemType.Kind(), value)
				}
			} else if mapElemType.Kind() == reflect.Struct || (mapElemType.Kind() == reflect.Pointer && mapElemType.Elem().Kind() == reflect.Struct) {
				if err := convertToStruct(value, mapValue.Addr().Interface()); err != nil {
					return fmt.Errorf("error converting value for key %v: %v", key, err)
				}
			} else {
				return fmt.Errorf("unsupported map element type: %v", mapElemType.Kind())
			}

			convertedMap.SetMapIndex(mapKey, mapValue)
		}

		field.Set(convertedMap)
		return nil
	}

	if fieldKind == reflect.Struct {
		return convertToStruct(fieldValue, field.Addr().Interface())
	}

	if fieldKind == reflect.Slice {
		rawData, ok := fieldValue.([]interface{})
		if !ok {
			return fmt.Errorf("expected slice for field, but got %T", fieldValue)
		}

		convertedSlice := reflect.MakeSlice(field.Type(), 0, len(rawData))
		for _, value := range rawData {
			sliceElem := reflect.New(field.Type().Elem()).Elem()
			if err := convertField(sliceElem, fieldType, value); err != nil {
				return fmt.Errorf("error converting slice element: %v", err)
			}
			convertedSlice = reflect.Append(convertedSlice, sliceElem)
		}
		field.Set(convertedSlice)
		return nil
	}

	if fieldKind == reflect.String {
		strValue := ToString(fieldValue)
		field.SetString(strValue)
		return nil
	}

	if fieldKind == reflect.Int64 {
		if floatValue, ok := fieldValue.(float64); ok {
			field.SetInt(int64(floatValue))
		} else {
			parsed, err := ToInt64(fieldValue)
			if err != nil {
				return fmt.Errorf("not able to parse to int64")
			}
			field.SetInt(parsed)
		}
		return nil
	}

	if fieldKind == reflect.Int {
		if floatValue, ok := fieldValue.(float64); ok {
			field.SetInt(int64(floatValue))
		} else {
			parsed, err := ToInt(fieldValue)
			if err != nil {
				return fmt.Errorf("not able to parse to int")
			}
			field.SetInt(int64(parsed))
		}
		return nil
	}

	if fieldKind == reflect.Float64 {
		parsed, err := ToFloat64(fieldValue)
		if err != nil {
			return fmt.Errorf("not able to parse to float64")
		}
		field.SetFloat(parsed)
		return nil
	}

	if fieldKind == reflect.Float32 {
		parsed, err := ToFloat32(fieldValue)
		if err != nil {
			return fmt.Errorf("not able to parse to float32")
		}
		field.SetFloat(float64(parsed))
		return nil
	}

	if fieldKind == reflect.Bool {
		parsed, err := ToBool(fieldValue)
		if err != nil {
			return fmt.Errorf("not able to parse to bool")
		}
		field.SetBool(parsed)
		return nil
	}

	return fmt.Errorf("unsupported field type: %v", fieldKind)
}

// Helper to determine if a field is primitive
func isPrimitive(kind reflect.Kind) bool {
	return kind == reflect.String ||
		kind == reflect.Int || kind == reflect.Int64 ||
		kind == reflect.Float32 || kind == reflect.Float64 ||
		kind == reflect.Bool
}
