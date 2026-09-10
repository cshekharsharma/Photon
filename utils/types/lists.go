package types

import (
	"errors"
	"fmt"
	"math"
	"reflect"
)

// Create 2 dimensional chunked array from the provided one dimensional
// array/list of elements, as each chunk of size n provided as `size` parameter.
//
// Example:
//
//	Input   = [1,2,3,4,5,6]
//	Size    = 2
//	Output  = [[1,2], [3,4], [5,6]]
func CreateChunks(s []interface{}, size int) ([][]interface{}, error) {
	if size < 1 {
		return nil, errors.New("size cannot be less than 1")
	}

	length := len(s)
	chunks := int(math.Ceil(float64(length) / float64(size)))

	var n [][]interface{}
	for i, end := 0, 0; chunks > 0; chunks-- {
		end = (i + 1) * size

		if end > length {
			end = length
		}

		n = append(n, s[i*size:end])
		i++
	}

	return n, nil
}

// ExistsInList checks if a given value exists within a slice and returns
// a boolean indicating its existence and the index at which it was found.
//
// Parameters:
//   - val: The value to search for in the slice.
//   - array: The slice to search within.
//
// Returns:
//   - exists: A boolean value indicating whether 'val' exists in 'array'.
//   - index: An integer representing the index where 'val' was found in 'array'.
//     If 'val' is not found, the index is set to -1.
func ExistsInList(val interface{}, array interface{}) (bool, int) {
	exists := false
	index := -1

	switch reflect.TypeOf(array).Kind() {
	case reflect.Slice:
		s := reflect.ValueOf(array)

		for i := 0; i < s.Len(); i++ {
			if reflect.DeepEqual(val, s.Index(i).Interface()) {
				index = i
				return true, index
			}
		}
	}

	return exists, index
}

// CastInterfaceSlice converts a slice of interfaces to a slice of type T.
// It returns an error if any item cannot be converted to T.
func CastInterfaceSlice[T any](slice []interface{}) ([]T, error) {
	if slice == nil {
		return nil, fmt.Errorf("input slice cannot be nil")
	}
	var result []T
	for _, v := range slice {
		switch val := v.(type) {
		case T:
			result = append(result, val) // Append the value of type T
		default:
			return nil, fmt.Errorf("element %v cannot be converted to %T", v, new(T))
		}
	}
	return result, nil
}

func ConvertToInterfaceSlice[T any](ids []T) []interface{} {
	result := make([]interface{}, len(ids))
	for i, id := range ids {
		result[i] = id
	}
	return result
}
