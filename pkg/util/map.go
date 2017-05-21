package util

import (
	"fmt"
	"sort"
	"strings"
)

func StringToMap (data string) map[string]string {
	var runtimeConfigOptions = make(map[string]string)

	split := strings.Split(data, ",")

	if len(data) > 0 && len(split) > 0 {
		for _, pair := range split {
			z := strings.Split(pair, "=")
			runtimeConfigOptions[z[0]] = z[1]
		}
	}
	return runtimeConfigOptions;
}

func MapToSortedList(input map[string]string) []string {
	keys := sortedKeyValues(input);

	output := make([]string, len(input))
	for i, key := range keys {
		output[i] = fmt.Sprintf("%s=%s", key, input[key])
	}
	return output;
}

func MergeMaps(newValues map[string]string, defaults map[string]string) map[string]string {
	mergeOutput := make(map[string]string);

	for key, val := range (newValues) {
		mergeOutput[key] = val
	}

	for key, val := range (defaults) {
		defaultIfNotSet(mergeOutput, key, val);
	}

	return mergeOutput;
}

func sortedKeyValues(input map[string]string) []string {
	keys := make([]string, len(input))
	i := 0
	for key := range input {
		keys[i] = key
		i++
	}
	sort.Strings(keys)
	return keys
}

func defaultIfNotSet(m map[string]string, key string, defaultValue string) {
	_, ok := m[key]
	if !ok {
		m[key] = defaultValue
	}
}
