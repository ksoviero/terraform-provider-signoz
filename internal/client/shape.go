// SPDX-License-Identifier: MPL-2.0

package client

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// Equivalent reports whether the API's value got means the same as the configured value want, at
// path (the keys and list indexes leading to it). It's for values the API rewrites, such as query
// function names it stores in lower case; plain equality is checked first, so it only needs to
// say yes to the rewrites.
type Equivalent func(path []string, want, got interface{}) bool

// ShapeLike compares what the API returned (got) with what Terraform last set (want), looking only
// at the fields want sets, and returns what to keep in state:
//
//   - want itself, exactly as given, when every field it sets has the same value in got. Fields
//     the API adds (its defaults, nulls, echoes) are ignored, so a new one never shows as drift.
//   - otherwise got reduced to want's fields, with want's own values wherever they still match:
//     the plan then shows only the fields that really changed.
//
// A field want sets to its zero value (false, 0, "", null, [] or {}) that got leaves out, or has
// as null, counts as the same: the API often omits its defaults. Lists are compared item by
// item; one of another length is drift, and got's extra items are kept whole. An empty want (an
// import, nothing in state yet) keeps all of got.
func ShapeLike(want, got string, eq Equivalent) (string, error) {
	if strings.TrimSpace(want) == "" {
		return got, nil
	}
	var w, g interface{}
	if err := json.Unmarshal([]byte(want), &w); err != nil {
		return "", fmt.Errorf("state value isn't valid JSON: %w", err)
	}
	if err := json.Unmarshal([]byte(got), &g); err != nil {
		return "", fmt.Errorf("API value isn't valid JSON: %w", err)
	}
	shaped, same := shapeLike(nil, w, g, eq)
	if same {
		return want, nil
	}
	b, err := json.Marshal(shaped)
	if err != nil {
		return "", err
	}
	return CanonicalJSON(string(b))
}

// shapeLike returns got reduced to want's shape, and whether it's the same as want.
func shapeLike(path []string, want, got interface{}, eq Equivalent) (interface{}, bool) {
	if got == nil && isZeroJSON(want) {
		return want, true
	}
	switch w := want.(type) {
	case map[string]interface{}:
		g, ok := got.(map[string]interface{})
		if !ok {
			return got, false
		}
		out := make(map[string]interface{}, len(w))
		same := true
		for k, wv := range w {
			gv, present := g[k]
			if !present {
				if isZeroJSON(wv) {
					out[k] = wv
				} else {
					same = false
				}
				continue
			}
			v, s := shapeLike(append(path[:len(path):len(path)], k), wv, gv, eq)
			out[k] = v
			same = same && s
		}
		return out, same
	case []interface{}:
		g, ok := got.([]interface{})
		if !ok {
			return got, false
		}
		out := make([]interface{}, len(g))
		same := len(w) == len(g)
		for i, gv := range g {
			if i >= len(w) {
				out[i] = gv
				continue
			}
			v, s := shapeLike(append(path[:len(path):len(path)], strconv.Itoa(i)), w[i], gv, eq)
			out[i] = v
			same = same && s
		}
		return out, same
	default:
		if reflect.DeepEqual(want, got) || (eq != nil && eq(path, want, got)) {
			return want, true
		}
		return got, false
	}
}

// isZeroJSON reports whether v is JSON's zero value of its kind.
func isZeroJSON(v interface{}) bool {
	switch t := v.(type) {
	case nil:
		return true
	case bool:
		return !t
	case float64:
		return t == 0
	case string:
		return t == ""
	case []interface{}:
		return len(t) == 0
	case map[string]interface{}:
		return len(t) == 0
	}
	return false
}
