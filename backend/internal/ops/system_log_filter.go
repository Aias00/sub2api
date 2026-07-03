package ops

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

type SystemLogListFilterInput struct {
	StartTime time.Time
	EndTime   time.Time
	Page      int
	PageSize  int

	LevelRaw           string
	ComponentRaw       string
	RequestIDRaw       string
	ClientRequestIDRaw string
	UserIDRaw          string
	APIKeyIDRaw        string
	AccountIDRaw       string
	PlatformRaw        string
	ModelRaw           string
	QueryRaw           string
}

type SystemLogListFilter struct {
	StartTime *time.Time
	EndTime   *time.Time
	Page      int
	PageSize  int

	Level           string
	Component       string
	RequestID       string
	ClientRequestID string
	UserID          *int64
	APIKeyID        *int64
	AccountID       *int64
	Platform        string
	Model           string
	Query           string
}

type SystemLogCleanupFilterInput struct {
	StartTimeRaw string
	EndTimeRaw   string

	LevelRaw           string
	ComponentRaw       string
	RequestIDRaw       string
	ClientRequestIDRaw string
	UserID             *int64
	APIKeyID           *int64
	AccountID          *int64
	PlatformRaw        string
	ModelRaw           string
	QueryRaw           string
}

type SystemLogCleanupFilter struct {
	StartTime *time.Time
	EndTime   *time.Time

	Level           string
	Component       string
	RequestID       string
	ClientRequestID string
	UserID          *int64
	APIKeyID        *int64
	AccountID       *int64
	Platform        string
	Model           string
	Query           string
}

func ParseSystemLogListFilter(input SystemLogListFilterInput) (*SystemLogListFilter, error) {
	filter := &SystemLogListFilter{
		Page:            input.Page,
		PageSize:        input.PageSize,
		Level:           strings.TrimSpace(input.LevelRaw),
		Component:       strings.TrimSpace(input.ComponentRaw),
		RequestID:       strings.TrimSpace(input.RequestIDRaw),
		ClientRequestID: strings.TrimSpace(input.ClientRequestIDRaw),
		Platform:        strings.TrimSpace(input.PlatformRaw),
		Model:           strings.TrimSpace(input.ModelRaw),
		Query:           strings.TrimSpace(input.QueryRaw),
	}
	if !input.StartTime.IsZero() {
		start := input.StartTime
		filter.StartTime = &start
	}
	if !input.EndTime.IsZero() {
		end := input.EndTime
		filter.EndTime = &end
	}

	var err error
	if filter.UserID, err = parseOptionalPositiveInt64(input.UserIDRaw, "Invalid user_id"); err != nil {
		return nil, err
	}
	if filter.APIKeyID, err = parseOptionalPositiveInt64(input.APIKeyIDRaw, "Invalid api_key_id"); err != nil {
		return nil, err
	}
	if filter.AccountID, err = parseOptionalPositiveInt64(input.AccountIDRaw, "Invalid account_id"); err != nil {
		return nil, err
	}
	return filter, nil
}

func ParseSystemLogCleanupFilter(input SystemLogCleanupFilterInput) (*SystemLogCleanupFilter, error) {
	start, err := parseOptionalTimestamp(input.StartTimeRaw)
	if err != nil {
		return nil, fmt.Errorf("Invalid start_time")
	}
	end, err := parseOptionalTimestamp(input.EndTimeRaw)
	if err != nil {
		return nil, fmt.Errorf("Invalid end_time")
	}
	if input.APIKeyID != nil && *input.APIKeyID <= 0 {
		return nil, fmt.Errorf("Invalid api_key_id")
	}

	return &SystemLogCleanupFilter{
		StartTime:       start,
		EndTime:         end,
		Level:           strings.TrimSpace(input.LevelRaw),
		Component:       strings.TrimSpace(input.ComponentRaw),
		RequestID:       strings.TrimSpace(input.RequestIDRaw),
		ClientRequestID: strings.TrimSpace(input.ClientRequestIDRaw),
		UserID:          input.UserID,
		APIKeyID:        input.APIKeyID,
		AccountID:       input.AccountID,
		Platform:        strings.TrimSpace(input.PlatformRaw),
		Model:           strings.TrimSpace(input.ModelRaw),
		Query:           strings.TrimSpace(input.QueryRaw),
	}, nil
}

func parseOptionalPositiveInt64(raw string, errMessage string) (*int64, error) {
	v := strings.TrimSpace(raw)
	if v == "" {
		return nil, nil
	}
	id, err := strconv.ParseInt(v, 10, 64)
	if err != nil || id <= 0 {
		return nil, fmt.Errorf("%s", errMessage)
	}
	return &id, nil
}

func parseOptionalTimestamp(raw string) (*time.Time, error) {
	v := strings.TrimSpace(raw)
	if v == "" {
		return nil, nil
	}
	if t, err := time.Parse(time.RFC3339Nano, v); err == nil {
		return &t, nil
	}
	t, err := time.Parse(time.RFC3339, v)
	if err != nil {
		return nil, err
	}
	return &t, nil
}
