package democompare

import (
	"encoding/json"
	"reflect"
	"sort"
	"strings"
	"testing"
)

func ToJSON(t *testing.T, v any) string {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("JSON marshal failed: %v", err)
	}
	return string(b)
}

func ParseJSON(t *testing.T, s string) any {
	t.Helper()
	var v any
	if err := json.Unmarshal([]byte(s), &v); err != nil {
		t.Fatalf("JSON unmarshal failed: %v", err)
	}
	return v
}

func JSONEqual(t *testing.T, goJSON, jsJSON string) bool {
	t.Helper()
	return reflect.DeepEqual(ParseJSON(t, goJSON), ParseJSON(t, jsJSON))
}

func QueryStringEqual(t *testing.T, goQS, jsQS, delimiter string) bool {
	t.Helper()
	goParams := parseQueryStringWithDelimiter(goQS, delimiter)
	jsParams := parseQueryStringWithDelimiter(jsQS, delimiter)
	return reflect.DeepEqual(goParams, jsParams)
}

func parseQueryStringWithDelimiter(s, delimiter string) map[string][]string {
	result := make(map[string][]string)
	if s == "" {
		return result
	}
	s = strings.TrimPrefix(s, "?")
	parts := strings.Split(s, delimiter)
	for _, part := range parts {
		idx := strings.Index(part, "=")
		var key, val string
		if idx == -1 {
			key = part
			val = ""
		} else {
			key = part[:idx]
			val = part[idx+1:]
		}
		result[key] = append(result[key], val)
	}
	for k := range result {
		sort.Strings(result[k])
	}
	return result
}

