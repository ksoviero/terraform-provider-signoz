// SPDX-License-Identifier: MPL-2.0

package client

import (
	"context"
	"encoding/json"
	"fmt"
)

const downtimeSchedulesPath = "/api/v1/downtime_schedules"

func (c *Client) ListDowntimeSchedules(ctx context.Context) ([]map[string]interface{}, error) {
	return c.ListMap(ctx, downtimeSchedulesPath)
}

func (c *Client) GetDowntimeSchedule(ctx context.Context, id string) (map[string]interface{}, error) {
	return c.GetMap(ctx, fmt.Sprintf("%s/%s", downtimeSchedulesPath, id))
}

func (c *Client) CreateDowntimeSchedule(ctx context.Context, body map[string]interface{}) (map[string]interface{}, error) {
	out, err := c.CreateMap(ctx, downtimeSchedulesPath, body)
	if err != nil {
		return nil, err
	}
	if id := MapString(out, "id"); id != "" && MapString(out, "name") == "" {
		return c.GetDowntimeSchedule(ctx, id)
	}
	return out, nil
}

func (c *Client) UpdateDowntimeSchedule(ctx context.Context, id string, body map[string]interface{}) error {
	return c.UpdateMap(ctx, fmt.Sprintf("%s/%s", downtimeSchedulesPath, id), body)
}

func (c *Client) DeleteDowntimeSchedule(ctx context.Context, id string) error {
	return c.DeleteMap(ctx, fmt.Sprintf("%s/%s", downtimeSchedulesPath, id))
}

// BuildDowntimeScheduleBody builds POST/PUT body from Terraform fields.
func BuildDowntimeScheduleBody(name, description, scheduleJSON string, alertIDs []string) (map[string]interface{}, error) {
	body := map[string]interface{}{
		"name": name,
	}
	if description != "" {
		body["description"] = description
	}
	if len(alertIDs) > 0 {
		body["alertIds"] = alertIDs
	}
	if scheduleJSON != "" {
		var sched map[string]interface{}
		if err := json.Unmarshal([]byte(scheduleJSON), &sched); err != nil {
			return nil, fmt.Errorf("schedule must be valid JSON: %w", err)
		}
		body["schedule"] = sched
	}
	return body, nil
}
