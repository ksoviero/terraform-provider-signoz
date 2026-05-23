// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/ksoviero/terraform-provider-signoz/internal/client"
)

func readAlertRuleByID(ctx context.Context, c *client.Client, id string) (alertRuleModel, error) {
	full, err := c.GetRuleMap(ctx, id)
	if err != nil {
		return alertRuleModel{}, err
	}
	spec, err := client.RuleSpecFromMap(full)
	if err != nil {
		return alertRuleModel{}, err
	}
	return modelFromRuleMap(full, spec), nil
}

func readAlertRuleByTitle(ctx context.Context, c *client.Client, title string) (alertRuleModel, error) {
	rules, err := c.ListRules(ctx)
	if err != nil {
		return alertRuleModel{}, err
	}
	var matches []client.Rule
	for _, r := range rules {
		if r.Alert == title {
			matches = append(matches, r)
		}
	}
	switch len(matches) {
	case 0:
		return alertRuleModel{}, fmt.Errorf("no alert rule with alert title %q", title)
	case 1:
		return readAlertRuleByID(ctx, c, matches[0].ID)
	default:
		return alertRuleModel{}, fmt.Errorf("multiple alert rules with alert title %q (%d matches); use id instead", title, len(matches))
	}
}

func readAllAlertRules(ctx context.Context, c *client.Client) ([]alertRuleModel, error) {
	rules, err := c.ListRules(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]alertRuleModel, 0, len(rules))
	for _, r := range rules {
		m, err := readAlertRuleByID(ctx, c, r.ID)
		if err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, nil
}
