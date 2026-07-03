package ops

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

type OpenAITokenStatsFilterInput struct {
	TimeRangeRaw string
	PlatformRaw  string
	GroupIDRaw   string
	TopNRaw      string
	PageRaw      string
	PageSizeRaw  string
	Now          time.Time
}

type OpenAITokenStatsFilter struct {
	TimeRange string
	StartTime time.Time
	EndTime   time.Time
	Platform  string
	GroupID   *int64
	Page      int
	PageSize  int
	TopN      int
}

func ParseOpenAITokenStatsFilter(input OpenAITokenStatsFilterInput) (*OpenAITokenStatsFilter, error) {
	timeRange := strings.TrimSpace(input.TimeRangeRaw)
	if timeRange == "" {
		timeRange = "30d"
	}
	dur, ok := ParseOpenAITokenStatsDuration(timeRange)
	if !ok {
		return nil, fmt.Errorf("invalid time_range")
	}

	end := input.Now
	if end.IsZero() {
		end = time.Now()
	}
	end = end.UTC()
	start := end.Add(-dur)

	filter := &OpenAITokenStatsFilter{
		TimeRange: timeRange,
		StartTime: start,
		EndTime:   end,
		Platform:  strings.TrimSpace(input.PlatformRaw),
	}

	if v := strings.TrimSpace(input.GroupIDRaw); v != "" {
		id, err := strconv.ParseInt(v, 10, 64)
		if err != nil || id <= 0 {
			return nil, fmt.Errorf("invalid group_id")
		}
		filter.GroupID = &id
	}

	topNRaw := strings.TrimSpace(input.TopNRaw)
	pageRaw := strings.TrimSpace(input.PageRaw)
	pageSizeRaw := strings.TrimSpace(input.PageSizeRaw)
	if topNRaw != "" && (pageRaw != "" || pageSizeRaw != "") {
		return nil, fmt.Errorf("invalid query: top_n cannot be used with page/page_size")
	}

	if topNRaw != "" {
		topN, err := strconv.Atoi(topNRaw)
		if err != nil || topN < 1 || topN > 100 {
			return nil, fmt.Errorf("invalid top_n")
		}
		filter.TopN = topN
		return filter, nil
	}

	filter.Page = 1
	filter.PageSize = 20
	if pageRaw != "" {
		page, err := strconv.Atoi(pageRaw)
		if err != nil || page < 1 {
			return nil, fmt.Errorf("invalid page")
		}
		filter.Page = page
	}
	if pageSizeRaw != "" {
		pageSize, err := strconv.Atoi(pageSizeRaw)
		if err != nil || pageSize < 1 || pageSize > 100 {
			return nil, fmt.Errorf("invalid page_size")
		}
		filter.PageSize = pageSize
	}
	return filter, nil
}
