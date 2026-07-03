package ops

import "strings"

type PlatformGroupFilterInput struct {
	PlatformRaw string
	GroupIDRaw  string
}

type PlatformGroupFilter struct {
	Platform string
	GroupID  *int64
}

func ParsePlatformGroupFilter(input PlatformGroupFilterInput) (*PlatformGroupFilter, error) {
	groupID, err := parseOptionalPositiveInt64(input.GroupIDRaw, "Invalid group_id")
	if err != nil {
		return nil, err
	}
	return &PlatformGroupFilter{
		Platform: strings.TrimSpace(input.PlatformRaw),
		GroupID:  groupID,
	}, nil
}
