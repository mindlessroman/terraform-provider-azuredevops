package validate

import (
	"fmt"
)

// ListOfMapsHasNoDuplicateKeys produces a validation function that will
// assert that each element in a list of maps has unique values for a specified key
func ListOfMapsHasNoDuplicateKeys(key string) func(interface{}, string) ([]string, []error) {
	return func(i interface{}, k string) ([]string, []error) {
		v, ok := i.([]map[string]interface{})
		if !ok {
			return nil, []error{fmt.Errorf("expected type of %q to be []map[string]interface{}", k)}
		}

		values := map[interface{}]bool{}
		for _, mp := range v {
			valueForKey := mp[key]
			if values[valueForKey] {
				return nil, []error{fmt.Errorf("Found a duplicate value for key %q", key)}
			}
			values[valueForKey] = true
		}
		return nil, nil
	}
}
