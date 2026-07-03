package ops

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

type RequestDetailFilterInput struct {
	StartTime time.Time
	EndTime   time.Time
	Page      int
	PageSize  int

	KindRaw          string
	PlatformRaw      string
	ModelRaw         string
	RequestIDRaw     string
	QueryRaw         string
	SortRaw          string
	UserIDRaw        string
	APIKeyIDRaw      string
	AccountIDRaw     string
	GroupIDRaw       string
	MinDurationMsRaw string
	MaxDurationMsRaw string
}

type RequestDetailFilter struct {
	StartTime *time.Time
	EndTime   *time.Time
	Page      int
	PageSize  int

	Kind      string
	Platform  string
	Model     string
	RequestID string
	Query     string
	Sort      string

	UserID    *int64
	APIKeyID  *int64
	AccountID *int64
	GroupID   *int64

	MinDurationMs *int
	MaxDurationMs *int
}

func ParseRequestDetailFilter(input RequestDetailFilterInput) (*RequestDetailFilter, error) {
	filter := &RequestDetailFilter{
		Page:      input.Page,
		PageSize:  input.PageSize,
		Kind:      strings.TrimSpace(input.KindRaw),
		Platform:  strings.TrimSpace(input.PlatformRaw),
		Model:     strings.TrimSpace(input.ModelRaw),
		RequestID: strings.TrimSpace(input.RequestIDRaw),
		Query:     strings.TrimSpace(input.QueryRaw),
		Sort:      strings.TrimSpace(input.SortRaw),
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
	if filter.GroupID, err = parseOptionalPositiveInt64(input.GroupIDRaw, "Invalid group_id"); err != nil {
		return nil, err
	}
	if filter.MinDurationMs, err = parseOptionalNonNegativeInt(input.MinDurationMsRaw, "Invalid min_duration_ms"); err != nil {
		return nil, err
	}
	if filter.MaxDurationMs, err = parseOptionalNonNegativeInt(input.MaxDurationMsRaw, "Invalid max_duration_ms"); err != nil {
		return nil, err
	}
	return filter, nil
}

func parseOptionalNonNegativeInt(raw string, errMessage string) (*int, error) {
	v := strings.TrimSpace(raw)
	if v == "" {
		return nil, nil
	}
	n, err := strconv.Atoi(v)
	if err != nil || n < 0 {
		return nil, fmt.Errorf("%s", errMessage)
	}
	return &n, nil
}
