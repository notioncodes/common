package common

import (
	"fmt"
	"time"

	"github.com/notioncodes/types"
)

// AddMetadata adds metadata to an object.
//
// Arguments:
// - objectType: The type of object.
// - object: The object to add metadata to.
//
// Returns:
// - The object with metadata.
// - An error if the object is not a map[string]any.
func AddMetadata(objectType types.ObjectType, object any) (map[string]any, error) {
	objMap, ok := object.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("expected map[string]any, got %T", object)
	}

	objMap["metadata"] = map[string]any{
		"object_type": string(objectType),
		"exported_at": time.Now(),
	}

	return objMap, nil
}

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
