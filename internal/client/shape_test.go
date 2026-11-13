// SPDX-License-Identifier: MPL-2.0

package client_test

import (
	"testing"

	"github.com/ksoviero/terraform-provider-signoz/internal/client"
)

func TestShapeLike(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name, want, got, expect string
		eq                      client.Equivalent
	}{
		{
			name:   "fields the API adds are ignored, and want is kept as written",
			want:   `{"a": 1,  "b": {"c": "x"}}`,
			got:    `{"a": 1, "b": {"c": "x", "d": null, "e": 0}, "f": "new"}`,
			expect: `{"a": 1,  "b": {"c": "x"}}`,
		},
		{
			name:   "a changed field shows with the API's value, the rest with want's",
			want:   `{"a": 1, "b": "x"}`,
			got:    `{"a": 2, "b": "x", "c": true}`,
			expect: `{"a":2,"b":"x"}`,
		},
		{
			name:   "a zero value the API leaves out is the same",
			want:   `{"a": false, "b": 0, "c": "", "d": null, "e": [], "f": {}, "g": 1}`,
			got:    `{"g": 1}`,
			expect: `{"a": false, "b": 0, "c": "", "d": null, "e": [], "f": {}, "g": 1}`,
		},
		{
			name:   "a set value the API leaves out is drift",
			want:   `{"a": true, "b": 1}`,
			got:    `{"b": 1}`,
			expect: `{"b":1}`,
		},
		{
			name:   "a null from the API for a zero value is the same",
			want:   `{"a": [], "b": ""}`,
			got:    `{"a": null, "b": null}`,
			expect: `{"a": [], "b": ""}`,
		},
		{
			name:   "list items are compared one by one",
			want:   `{"l": [{"n": "A"}, {"n": "B"}]}`,
			got:    `{"l": [{"n": "A", "x": 1}, {"n": "C", "x": 1}]}`,
			expect: `{"l":[{"n":"A"},{"n":"C"}]}`,
		},
		{
			name:   "an extra list item is drift, kept whole",
			want:   `{"l": [{"n": "A"}]}`,
			got:    `{"l": [{"n": "A", "x": 1}, {"n": "B", "x": 1}]}`,
			expect: `{"l":[{"n":"A"},{"n":"B","x":1}]}`,
		},
		{
			name:   "a missing list item is drift",
			want:   `{"l": [{"n": "A"}, {"n": "B"}]}`,
			got:    `{"l": [{"n": "A"}]}`,
			expect: `{"l":[{"n":"A"}]}`,
		},
		{
			name:   "a value of another type is drift",
			want:   `{"a": {"b": 1}}`,
			got:    `{"a": "x"}`,
			expect: `{"a":"x"}`,
		},
		{
			name:   "durations written two ways are the same",
			want:   `{"a": "60s", "b": "24h", "c": "2m"}`,
			got:    `{"a": "1m0s", "b": "24h0m0s", "c": "3m0s"}`,
			expect: `{"a":"60s","b":"24h","c":"3m0s"}`,
			eq:     client.SameDuration,
		},
		{
			name:   "without SameDuration, a duration-like string is just text",
			want:   `{"label": "1h"}`,
			got:    `{"label": "60m"}`,
			expect: `{"label":"60m"}`,
		},
		{
			name:   "outside headers a key's case matters",
			want:   `{"Title": "x"}`,
			got:    `{"title": "x"}`,
			expect: `{}`,
		},
		{
			name:   "a key in two other cases matches neither",
			want:   `{"h": {"subject": "x"}}`,
			got:    `{"h": {"Subject": "x", "SUBJECT": "x"}}`,
			expect: `{"h":{}}`,
		},
		{
			name:   "a key the API returns in another case matches",
			want:   `{"headers": {"subject": "x"}}`,
			got:    `{"headers": {"Subject": "x"}}`,
			expect: `{"headers": {"subject": "x"}}`,
		},
		{
			name:   "nothing in state keeps all of the API's",
			want:   ``,
			got:    `{"a": 1, "b": null}`,
			expect: `{"a": 1, "b": null}`,
		},
		{
			name:   "rule rewrites: function names in lower case, respaced filter lists",
			want:   `{"q": {"functions": [{"name": "cumulativeSum"}], "filter": {"expression": "a IN [\"x\",\"y\"]"}, "name": "A"}}`,
			got:    `{"q": {"functions": [{"name": "cumulativesum"}], "filter": {"expression": "a IN ['x', 'y']"}, "name": "A"}}`,
			expect: `{"q": {"functions": [{"name": "cumulativeSum"}], "filter": {"expression": "a IN [\"x\",\"y\"]"}, "name": "A"}}`,
			eq:     client.RuleValuesEquivalent,
		},
		{
			name:   "rule rewrites don't hide another name changing case",
			want:   `{"q": {"name": "A"}}`,
			got:    `{"q": {"name": "a"}}`,
			expect: `{"q":{"name":"a"}}`,
			eq:     client.RuleValuesEquivalent,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			out, err := client.ShapeLike(tc.want, tc.got, tc.eq)
			if err != nil {
				t.Fatal(err)
			}
			if out != tc.expect {
				t.Fatalf("got  %s\nwant %s", out, tc.expect)
			}
		})
	}
}

// Tidying a rule's spec (RuleSpecFromMap) mustn't change the raw spec read from the same rule.
func TestRawRuleSpecFromMap_untouchedByTidying(t *testing.T) {
	t.Parallel()

	full := map[string]interface{}{
		"id": "r",
		"condition": map[string]interface{}{
			"absentFor": 0.0, "seasonality": "",
			"compositeQuery": map[string]interface{}{"queries": []interface{}{
				map[string]interface{}{"spec": map[string]interface{}{"functions": []interface{}{map[string]interface{}{"name": "cumulativeSum"}}}},
			}},
		},
		"notificationSettings": map[string]interface{}{"newGroupEvalDelay": "0s"},
	}
	if _, err := client.RuleSpecFromMap(full); err != nil {
		t.Fatal(err)
	}
	raw, err := client.RawRuleSpecFromMap(full)
	if err != nil {
		t.Fatal(err)
	}
	want := `{"condition":{"absentFor":0,"compositeQuery":{"queries":[{"spec":{"functions":[{"name":"cumulativeSum"}]}}]},"seasonality":""},"notificationSettings":{"newGroupEvalDelay":"0s"}}`
	if raw != want {
		t.Fatalf("got  %s\nwant %s", raw, want)
	}
}

func TestWithoutRuleEchoes(t *testing.T) {
	t.Parallel()

	got, err := client.WithoutRuleEchoes(`{"evalWindow":"5m","frequency":"1m","evaluation":{"spec":{"evalWindow":"5m"}}}`)
	if err != nil {
		t.Fatal(err)
	}
	if want := `{"evaluation":{"spec":{"evalWindow":"5m"}}}`; got != want {
		t.Fatalf("got %s, want %s", got, want)
	}
	unchanged := `{"evaluation": {}}`
	if got, _ := client.WithoutRuleEchoes(unchanged); got != unchanged {
		t.Fatalf("a spec without echoes changed: %s", got)
	}
}
