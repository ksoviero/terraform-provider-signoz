// SPDX-License-Identifier: MPL-2.0

package client

// PruneJSONDefaults removes SigNoz API defaults that should not force Terraform drift:
// null, empty strings, empty objects, and empty arrays (nested maps are pruned recursively).
func PruneJSONDefaults(v interface{}) interface{} {
	switch x := v.(type) {
	case map[string]interface{}:
		PruneJSONDefaultsMap(x)
		return x
	case []interface{}:
		out := make([]interface{}, 0, len(x))
		for _, elem := range x {
			pruned := PruneJSONDefaults(elem)
			if pruned == nil {
				continue
			}
			out = append(out, pruned)
		}
		return out
	default:
		return v
	}
}

func PruneJSONDefaultsMap(m map[string]interface{}) {
	for k, v := range m {
		switch val := v.(type) {
		case nil:
			delete(m, k)
		case string:
			if val == "" {
				delete(m, k)
			}
		case map[string]interface{}:
			if len(val) == 0 {
				delete(m, k)
				continue
			}
			PruneJSONDefaultsMap(val)
			if len(val) == 0 {
				delete(m, k)
			}
		case []interface{}:
			if len(val) == 0 {
				delete(m, k)
				continue
			}
			pruned := PruneJSONDefaults(val)
			if arr, ok := pruned.([]interface{}); ok {
				if len(arr) == 0 {
					delete(m, k)
				} else {
					m[k] = arr
				}
			}
		default:
			m[k] = PruneJSONDefaults(val)
		}
	}
}

// pruneRuleSpecDefaults removes alert-rule spec fields the API adds with zero/empty defaults.
func pruneRuleSpecDefaults(m map[string]interface{}) {
	PruneJSONDefaultsMap(m)
	pruneRuleSpecValue(m)
}

func pruneRuleSpecValue(v interface{}) {
	switch x := v.(type) {
	case map[string]interface{}:
		for _, key := range []string{"step", "stepInterval"} {
			if val, ok := x[key]; ok && isNumericZero(val) {
				delete(x, key)
			}
		}
		for _, child := range x {
			pruneRuleSpecValue(child)
		}
	case []interface{}:
		for _, elem := range x {
			pruneRuleSpecValue(elem)
		}
	}
}

func isNumericZero(v interface{}) bool {
	switch n := v.(type) {
	case float64:
		return n == 0
	case int:
		return n == 0
	case int64:
		return n == 0
	default:
		return false
	}
}
