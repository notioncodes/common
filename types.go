package common

// ToInterfaceSlice converts a slice of any type to a slice of interfaces.
//
// Arguments:
// - input: The input to convert.
//
// Returns:
// - The slice of interfaces.
func ToInterfaceSlice[T any](input []T) []interface{} {
	out := make([]interface{}, len(input))
	for i, v := range input {
		out[i] = v
	}
	return out
}

// ToInterfaceMap converts a map of any type to a map of interfaces.
//
// Arguments:
// - input: The input to convert.
//
// Returns:
// - The map of interfaces.
func ToInterfaceMap(input any) interface{} {
	return input.(interface{})
}

// AppendToMap appends a map of interfaces to a base map of interfaces.
//
// Arguments:
// - base: The base map to append to.
// - append: The map to append.
func AppendToMap(base map[string]interface{}, append map[string]interface{}) map[string]interface{} {
	for k, v := range append {
		base[k] = v
	}
	return base
}
