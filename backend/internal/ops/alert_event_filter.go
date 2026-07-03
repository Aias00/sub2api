package ops

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

type AlertEventFilterInput struct {
	LimitRaw         string
	StatusRaw        string
	SeverityRaw      string
	EmailSentRaw     string
	BeforeFiredAtRaw string
	BeforeIDRaw      string
	PlatformRaw      string
	GroupIDRaw       string
	StartTime        time.Time
	EndTime          time.Time
	HasTimeRange     bool
}

type AlertEventFilter struct {
	Limit int

	BeforeFiredAt *time.Time
	BeforeID      *int64

	Status    string
	Severity  string
	EmailSent *bool

	StartTime *time.Time
	EndTime   *time.Time

	Platform string
	GroupID  *int64
}

type AlertSilenceInput struct {
	RuleID      int64
	PlatformRaw string
	GroupID     *int64
	Region      *string
	UntilRaw    string
	ReasonRaw   string
}

type AlertSilence struct {
	RuleID   int64
	Platform string
	GroupID  *int64
	Region   *string
	Until    time.Time
	Reason   string
}

const (
	AlertStatusResolved       = "resolved"
	AlertStatusManualResolved = "manual_resolved"
)

func ParseAlertEventStatus(raw string) (string, error) {
	status := strings.TrimSpace(raw)
	switch status {
	case AlertStatusResolved, AlertStatusManualResolved:
		return status, nil
	default:
		return "", fmt.Errorf("Invalid status")
	}
}

func AlertEventResolvedAt(status string, now time.Time) *time.Time {
	switch strings.TrimSpace(status) {
	case AlertStatusResolved, AlertStatusManualResolved:
		t := now.UTC()
		return &t
	default:
		return nil
	}
}

func ParseAlertEventFilter(input AlertEventFilterInput) (*AlertEventFilter, error) {
	limit := 20
	if raw := strings.TrimSpace(input.LimitRaw); raw != "" {
		n, err := strconv.Atoi(raw)
		if err != nil || n <= 0 {
			return nil, fmt.Errorf("Invalid limit")
		}
		limit = n
	}

	filter := &AlertEventFilter{
		Limit:    limit,
		Status:   strings.TrimSpace(input.StatusRaw),
		Severity: strings.TrimSpace(input.SeverityRaw),
		Platform: strings.TrimSpace(input.PlatformRaw),
	}

	if raw := strings.TrimSpace(input.EmailSentRaw); raw != "" {
		switch strings.ToLower(raw) {
		case "true", "1":
			b := true
			filter.EmailSent = &b
		case "false", "0":
			b := false
			filter.EmailSent = &b
		default:
			return nil, fmt.Errorf("Invalid email_sent")
		}
	}

	rawTS := strings.TrimSpace(input.BeforeFiredAtRaw)
	rawID := strings.TrimSpace(input.BeforeIDRaw)
	if (rawTS == "") != (rawID == "") {
		return nil, fmt.Errorf("before_fired_at and before_id must be provided together")
	}
	if rawTS != "" {
		ts, err := parseRequiredTimestamp(rawTS)
		if err != nil {
			return nil, fmt.Errorf("Invalid before_fired_at")
		}
		filter.BeforeFiredAt = &ts

		id, err := strconv.ParseInt(rawID, 10, 64)
		if err != nil || id <= 0 {
			return nil, fmt.Errorf("Invalid before_id")
		}
		filter.BeforeID = &id
	}

	if raw := strings.TrimSpace(input.GroupIDRaw); raw != "" {
		id, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || id <= 0 {
			return nil, fmt.Errorf("Invalid group_id")
		}
		filter.GroupID = &id
	}

	if input.HasTimeRange {
		start := input.StartTime
		end := input.EndTime
		filter.StartTime = &start
		filter.EndTime = &end
	}
	return filter, nil
}

func ParseAlertSilence(input AlertSilenceInput) (*AlertSilence, error) {
	until, err := time.Parse(time.RFC3339, strings.TrimSpace(input.UntilRaw))
	if err != nil {
		return nil, fmt.Errorf("Invalid until")
	}
	return &AlertSilence{
		RuleID:   input.RuleID,
		Platform: strings.TrimSpace(input.PlatformRaw),
		GroupID:  input.GroupID,
		Region:   input.Region,
		Until:    until,
		Reason:   strings.TrimSpace(input.ReasonRaw),
	}, nil
}

func parseRequiredTimestamp(raw string) (time.Time, error) {
	if t, err := time.Parse(time.RFC3339Nano, raw); err == nil {
		return t, nil
	}
	return time.Parse(time.RFC3339, raw)
}
