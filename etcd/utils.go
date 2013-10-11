// Utility functions

package etcd

import (
	"fmt"
	"net/url"
)

// Convert options to a string of HTML parameters
func optionsToString(options Options, validOptions map[string]bool) (string, error) {
	p := "?"
	v := url.Values{}
	for opKey, opVal := range options {
		if validGetOptions[opKey] {
			v.Set(opKey, fmt.Sprintf("%v", opVal))
		} else {
			return "", fmt.Errorf("Invalid option: %v", opKey)
		}
	}
	p += v.Encode()
	return p, nil
}
