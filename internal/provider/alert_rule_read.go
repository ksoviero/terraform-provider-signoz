// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/ksoviero/terraform-provider-signoz/internal/client"
)

func readAlertRuleByID(ctx context.Context, c *client.Client, id string) (alertRuleReadResult, error) {
	full, err := c.GetRuleMap(ctx, id)
	if err != nil {
		return alertRuleReadResult{}, err
	}
	spec, err := client.RuleSpecFromMap(full)
	if err != nil {
		return alertRuleReadResult{}, err
	}
	return modelFromRuleMap(full, spec), nil
}

func readAlertRuleByTitle(ctx context.Context, c *client.Client, title string) (alertRuleReadResult, error) {
	rules, err := c.ListRules(ctx)
	if err != nil {
		return alertRuleReadResult{}, err
	}
	var matches []client.Rule
	for _, r := range rules {
		if r.Alert == title {
			matches = append(matches, r)
		}
	}
	switch len(matches) {
	case 0:
		return alertRuleReadResult{}, fmt.Errorf("no alert rule with alert title %q", title)
	case 1:
		return readAlertRuleByID(ctx, c, matches[0].ID)
	default:
		return alertRuleReadResult{}, fmt.Errorf("multiple alert rules with alert title %q (%d matches); use id instead", title, len(matches))
	}
}

func readAllAlertRules(ctx context.Context, c *client.Client) ([]alertRuleReadResult, error) {
	rules, err := c.ListRules(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]alertRuleReadResult, 0, len(rules))
	for _, r := range rules {
		m, err := readAlertRuleByID(ctx, c, r.ID)
		if err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, nil
}
