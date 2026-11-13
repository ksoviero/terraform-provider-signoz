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
