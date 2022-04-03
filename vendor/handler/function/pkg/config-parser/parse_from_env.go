package config_parser

import (
	"fmt"
	"os"
	"strings"
)

func ParseEnv(keys []string) (map[string]string, error) {
	values := make(map[string]string)
	undefinedKeys := []string{}
	for _, key := range keys {
		value, isSet := os.LookupEnv(key)
		if !isSet {
			undefinedKeys = append(undefinedKeys, key)
		} else {
			values[key] = value
		}
	}
	if len(undefinedKeys) != 0 {
		return nil, fmt.Errorf("the following env keys are not defined : %s the function can't process without it", strings.Join(undefinedKeys, ","))
	}

	return values, nil
}
