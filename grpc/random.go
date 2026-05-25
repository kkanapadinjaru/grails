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
	v = randomizeField("", v)
	out, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshaling randomized JSON: %w", err)
	}
	return string(out), nil
}

// aipHint classifies a field name against the small set of well-known AIP
// (https://google.aip.dev) conventions. We treat these names as a contract:
// services that use them are opting into the convention, so producing a
// "smart" default is more useful than a random value. Unknown names fall
// through to the generic randomization.
type aipHint int

const (
	hintNone        aipHint = iota
	hintPageSize            // ints: 10..50 (AIP-158 page_size, limit)
	hintZeroInt             // ints: 0 (skip/offset/page_number)
	hintEmptyString         // strings: "" (filter, order_by, page_token, masks)
)

// aipFieldHints maps both snake_case and camelCase variants since proto3
// JSON uses camelCase by default but explicit json_name (as in the Solvas
// protos) keeps the snake form.
var aipFieldHints = map[string]aipHint{
	"page_size": hintPageSize, "pageSize": hintPageSize,
	"limit":       hintPageSize,
	"max_results": hintPageSize, "maxResults": hintPageSize,

	"skip":   hintZeroInt,
	"offset": hintZeroInt,
	"page_number": hintZeroInt, "pageNumber": hintZeroInt,

	"page_token": hintEmptyString, "pageToken": hintEmptyString,
	"next_page_token": hintEmptyString, "nextPageToken": hintEmptyString,
	"cursor": hintEmptyString,

	"filter":   hintEmptyString,
	"order_by": hintEmptyString, "orderBy": hintEmptyString,
	"sort":   hintEmptyString,
	"query":  hintEmptyString,
	"search": hintEmptyString,

	"field_mask": hintEmptyString, "fieldMask": hintEmptyString,
	"update_mask": hintEmptyString, "updateMask": hintEmptyString,
	"read_mask": hintEmptyString, "readMask": hintEmptyString,
}

func randomizeField(name string, v interface{}) interface{} {
	hint := aipFieldHints[name]
	switch x := v.(type) {
	case map[string]interface{}:
		for k, val := range x {
			x[k] = randomizeField(k, val)
		}
		return x
	case []interface{}:
		// Repeated fields inherit the parent name so a `repeated string tags`
		// under a `filter` field still gets the empty-string treatment.
		for i, val := range x {
			x[i] = randomizeField(name, val)
		}
		return x
	case string:
		if x == "" {
			if hint == hintEmptyString {
				return ""
			}
			return randomString()
		}
		// Likely an enum default — leave for the user to set manually.
		return x
	case float64:
		if x == 0 {
			switch hint {
			case hintPageSize:
				return mathrand.Intn(41) + 10 // 10..50
			case hintZeroInt:
				return 0
			}
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

// ApplyParentPattern fills the field named by parentField in jsonBody with a
// realized version of parentPattern (e.g. "o/*/items/*"), substituting JWT
// claim values for wildcards whose preceding literal segment is mapped in
// claimMap. Unmapped wildcards get a random sample value. parentField may use
// dot notation for nested fields ("metadata.parent"). If the field doesn't
// exist in the JSON object, ApplyParentPattern creates it.
func ApplyParentPattern(jsonBody, parentField, parentPattern string, claims map[string]any, claimMap map[string]string) (string, error) {
	if parentField == "" || parentPattern == "" {
		return jsonBody, nil
	}
	if strings.TrimSpace(jsonBody) == "" {
		jsonBody = "{}"
	}

	var doc map[string]any
	if err := json.Unmarshal([]byte(jsonBody), &doc); err != nil {
		return "", fmt.Errorf("parsing request JSON: %w", err)
	}

	value := realizePattern(parentPattern, claims, claimMap)
	setNested(doc, strings.Split(parentField, "."), value)

	out, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshaling request JSON: %w", err)
	}
	return string(out), nil
}

// realizePattern walks segments of an HTTP path pattern, replacing every "*"
// with the JWT claim value mapped from the immediately preceding literal
// segment, or a random sample value when no mapping is configured or the
// claim is absent.
func realizePattern(pattern string, claims map[string]any, claimMap map[string]string) string {
	parts := strings.Split(pattern, "/")
	var prevLiteral string
	for i, p := range parts {
		if p != "*" {
			prevLiteral = p
			continue
		}
		if claimName, ok := claimMap[prevLiteral]; ok && claimName != "" {
			if raw, present := claims[claimName]; present {
				if s := stringify(raw); s != "" {
					parts[i] = s
					continue
				}
			}
		}
		parts[i] = randomString()
	}
	return strings.Join(parts, "/")
}

func stringify(v any) string {
	switch x := v.(type) {
	case string:
		return x
	case float64:
		return fmt.Sprintf("%v", x)
	case bool:
		return fmt.Sprintf("%v", x)
	case nil:
		return ""
	default:
		b, err := json.Marshal(x)
		if err != nil {
			return ""
		}
		return string(b)
	}
}

func setNested(doc map[string]any, path []string, value any) {
	cur := doc
	for i, key := range path {
		if i == len(path)-1 {
			cur[key] = value
			return
		}
		next, ok := cur[key].(map[string]any)
		if !ok {
			next = map[string]any{}
			cur[key] = next
		}
		cur = next
	}
}
