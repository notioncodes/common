package common

import (
	"fmt"
)

// MergeObjects merges two objects using type assertion for safety.
//
// Arguments:
// - base: The base object.
// - override: The override object.
//
// Returns:
// - The merged object.
// - An error if the objects are not maps[string]any.
func MergeObjects(left any, right any) (map[string]any, error) {
	baseMap, ok := left.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("base must be map[string]any, got %T", left)
	}

	overrideMap, ok := right.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("override must be map[string]any, got %T", right)
	}

	// Create a deep copy of baseMap.
	result := make(map[string]any, len(baseMap))
	for k, v := range baseMap {
		result[k] = v
	}

	for k, v := range overrideMap {
		if existing, found := result[k]; found {
			// If both are maps, merge recursively.
			if existingMap, ok := existing.(map[string]any); ok {
				if overrideSubMap, ok := v.(map[string]any); ok {
					merged, err := MergeObjects(existingMap, overrideSubMap)
					if err != nil {
						return nil, fmt.Errorf("failed to merge nested object at key %q: %w", k, err)
					}
					result[k] = merged
					continue
				}
			}
		}
		result[k] = v
	}

	return result, nil
}
