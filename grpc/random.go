package grpc

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	mathrand "math/rand"
	"strings"
)

// RandomizeSkeleton walks a grpcurl-generated message-template JSON and replaces
// zero-valued primitives with plausible-looking random values. Strings ending
// in `_UNSPECIFIED` are left alone — those are almost certainly enums, and we
// can't tell the other valid values from JSON alone. Existing non-zero values
// are preserved so users can pin specific fields and re-roll the rest.
func RandomizeSkeleton(skeleton string) (string, error) {
	if strings.TrimSpace(skeleton) == "" {
		return skeleton, nil
	}
	var v interface{}
	if err := json.Unmarshal([]byte(skeleton), &v); err != nil {
		return "", fmt.Errorf("parsing skeleton JSON: %w", err)
	}
	v = randomizeValue(v)
	out, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshaling randomized JSON: %w", err)
	}
	return string(out), nil
}

func randomizeValue(v interface{}) interface{} {
	switch x := v.(type) {
	case map[string]interface{}:
		for k, val := range x {
			x[k] = randomizeValue(val)
		}
		return x
	case []interface{}:
		for i, val := range x {
			x[i] = randomizeValue(val)
		}
		return x
	case string:
		if x == "" {
			return randomString()
		}
		// Likely an enum default — leave for the user to set manually.
		return x
	case float64:
		if x == 0 {
			return mathrand.Intn(1000) + 1
		}
		return x
	case bool:
		if !x {
			return mathrand.Intn(2) == 0
		}
		return x
	}
	return v
}

func randomString() string {
	buf := make([]byte, 4)
	if _, err := rand.Read(buf); err != nil {
		return "sample-value"
	}
	return "sample-" + hex.EncodeToString(buf)
}
