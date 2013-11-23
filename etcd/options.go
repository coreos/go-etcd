package etcd

import (
	"fmt"
	"net/url"
	"reflect"
)

type options map[string]interface{}

// An internally-used data structure that represents a mapping
// between valid options and their kinds
type validOptions map[string]reflect.Kind

// Convert options to a string of HTML parameters
func (ops *options) toParameters(validOps validOptions) (string, error) {
	p := "?"
	values := url.Values{}

	for k, v := range *ops {
		// Check if the given option is valid (that it exists)
		kind := validOps[k]
		if kind == reflect.Invalid {
			return "", fmt.Errorf("Invalid option: %v", k)
		}

		// Check if the given option is of the valid type
		t := reflect.TypeOf(v)
		if kind != t.Kind() {
			return "", fmt.Errorf("Option %s should be of %v kind, not of %v kind.",
				k, kind, t.Kind())
		}

		values.Set(k, fmt.Sprintf("%v", v))
	}

	p += values.Encode()
	return p, nil
}
