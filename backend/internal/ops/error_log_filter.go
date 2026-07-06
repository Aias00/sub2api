package ops

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

type ErrorLogFilterInput struct {
	StartTime time.Time
	EndTime   time.Time
	Page      int
	PageSize  int

	ViewRaw            string
	PhaseRaw           string
	OwnerRaw           string
	SourceRaw          string
	QueryRaw           string
	UserQueryRaw       string
	ModelRaw           string
	RequestIDRaw       string
	ClientRequestIDRaw string
	PlatformRaw        string
	GroupIDRaw         string
	AccountIDRaw       string
	UserIDRaw          string
	APIKeyIDRaw        string
	ResolvedRaw        string
	StatusCodesRaw     string

	ClearUpstreamPhase bool
}

type ErrorLogFilter struct {
	StartTime *time.Time
	EndTime   *time.Time
	Page      int
	PageSize  int

	View            string
	Phase           string
	Owner           string
	Source          string
	Query           string
	UserQuery       string
	Model           string
	RequestID       string
	ClientRequestID string
	Platform        string
	GroupID         *int64
	AccountID       *int64
	UserID          *int64
	APIKeyID        *int64
	Resolved        *bool
	StatusCodes     []int
}

func ParseErrorLogFilter(input ErrorLogFilterInput) (*ErrorLogFilter, error) {
	filter := &ErrorLogFilter{
		Page:            input.Page,
		PageSize:        input.PageSize,
		View:            ParseListView(input.ViewRaw),
		Phase:           strings.TrimSpace(input.PhaseRaw),
		Owner:           strings.TrimSpace(input.OwnerRaw),
		Source:          strings.TrimSpace(input.SourceRaw),
		Query:           strings.TrimSpace(input.QueryRaw),
		UserQuery:       strings.TrimSpace(input.UserQueryRaw),
		Model:           strings.TrimSpace(input.ModelRaw),
		RequestID:       strings.TrimSpace(input.RequestIDRaw),
		ClientRequestID: strings.TrimSpace(input.ClientRequestIDRaw),
		Platform:        strings.TrimSpace(input.PlatformRaw),
	}
	if !input.StartTime.IsZero() {
		start := input.StartTime
		filter.StartTime = &start
	}
	if !input.EndTime.IsZero() {
		end := input.EndTime
		filter.EndTime = &end
	}
	if input.ClearUpstreamPhase && strings.EqualFold(strings.TrimSpace(filter.Phase), "upstream") {
		filter.Phase = ""
	}

	var err error
	if filter.GroupID, err = parseOptionalPositiveInt64(input.GroupIDRaw, "Invalid group_id"); err != nil {
		return nil, err
	}
	if filter.AccountID, err = parseOptionalPositiveInt64(input.AccountIDRaw, "Invalid account_id"); err != nil {
		return nil, err
	}
	if filter.UserID, err = parseOptionalPositiveInt64(input.UserIDRaw, "Invalid user_id"); err != nil {
		return nil, err
	}
	if filter.APIKeyID, err = parseOptionalPositiveInt64(input.APIKeyIDRaw, "Invalid api_key_id"); err != nil {
		return nil, err
	}
	if filter.Resolved, err = parseOptionalBool(input.ResolvedRaw, "Invalid resolved"); err != nil {
		return nil, err
	}
	if filter.StatusCodes, err = parseStatusCodes(input.StatusCodesRaw); err != nil {
		return nil, err
	}
	return filter, nil
}

func parseOptionalBool(raw string, errMessage string) (*bool, error) {
	v := strings.TrimSpace(raw)
	if v == "" {
		return nil, nil
	}
	switch strings.ToLower(v) {
	case "1", "true", "yes":
		b := true
		return &b, nil
	case "0", "false", "no":
		b := false
		return &b, nil
	default:
		return nil, fmt.Errorf("%s", errMessage)
	}
}

func parseStatusCodes(raw string) ([]int, error) {
	v := strings.TrimSpace(raw)
	if v == "" {
		return nil, nil
	}
	parts := strings.Split(v, ",")
	out := make([]int, 0, len(parts))
	for _, part := range parts {
		p := strings.TrimSpace(part)
		if p == "" {
			continue
		}
		n, err := strconv.Atoi(p)
		if err != nil || n < 0 {
			return nil, fmt.Errorf("invalid status_codes")
		}
		out = append(out, n)
	}
	return out, nil
}
