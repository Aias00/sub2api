package handler

import (
	"encoding/json"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	middleware "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

// UserBalanceLedgerHandler 用户余额流水 Handler
type UserBalanceLedgerHandler struct {
	ledgerService *service.UserBalanceLedgerService
}

// NewUserBalanceLedgerHandler 创建余额流水 Handler
func NewUserBalanceLedgerHandler(ledgerService *service.UserBalanceLedgerService) *UserBalanceLedgerHandler {
	return &UserBalanceLedgerHandler{
		ledgerService: ledgerService,
	}
}

// BalanceLedgerListRequest 查询请求参数
type BalanceLedgerListRequest struct {
	Page       int    `form:"page"`
	PageSize   int    `form:"page_size"`
	EntryTypes string `form:"entry_types"` // 逗号分隔的 entry_type 列表
	StartAt    string `form:"start_at"`    // ISO 8601 格式
	EndAt      string `form:"end_at"`      // ISO 8601 格式
}

// BalanceLedgerEntryResponse 流水记录响应
type BalanceLedgerEntryResponse struct {
	ID            int64   `json:"id"`
	EntryType     string  `json:"entry_type"`
	Amount        float64 `json:"amount"`
	BalanceBefore *float64 `json:"balance_before"`
	BalanceAfter  *float64 `json:"balance_after"`
	SourceType    string  `json:"source_type"`
	SourceID      *int64  `json:"source_id"`
	Description   string  `json:"description"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt     string  `json:"created_at"`
}

// BalanceLedgerListResponse 流水列表响应
type BalanceLedgerListResponse struct {
	Entries  []BalanceLedgerEntryResponse `json:"entries"`
	Total    int64                        `json:"total"`
	Page     int                          `json:"page"`
PageSize int                          `json:"page_size"`
}

// GetUserBalanceLedger GET /api/v1/user/balance-ledger
// 获取当前用户的余额流水列表
func (h *UserBalanceLedgerHandler) GetUserBalanceLedger(c *gin.Context) {
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	var req BalanceLedgerListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "Invalid query parameters")
		return
	}

	// 构建过滤器
	filter := service.BalanceLedgerFilter{
		Page:     req.Page,
		PageSize: req.PageSize,
	}

	// 解析 entry_types（逗号分隔）
	if req.EntryTypes != "" {
		types := parseEntryTypes(req.EntryTypes)
		if len(types) > 0 {
			filter.EntryTypes = types
		}
	}

	// 解析时间范围
	if req.StartAt != "" {
		t, err := time.Parse(time.RFC3339, req.StartAt)
		if err == nil {
			filter.StartAt = &t
		}
	}
	if req.EndAt != "" {
		t, err := time.Parse(time.RFC3339, req.EndAt)
		if err == nil {
			filter.EndAt = &t
		}
	}

	// 查询流水
	entries, total, err := h.ledgerService.ListUserLedger(c.Request.Context(), subject.UserID, filter)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	// 转换响应格式
	resp := BalanceLedgerListResponse{
		Entries: make([]BalanceLedgerEntryResponse, len(entries)),
		Total:   total,
		Page:    filter.Page,
		PageSize: filter.PageSize,
	}

	for i, entry := range entries {
		resp.Entries[i] = convertEntryToResponse(entry)
	}

	response.Success(c, resp)
}

// parseEntryTypes 解析逗号分隔的 entry_type 列表
func parseEntryTypes(s string) []service.BalanceLedgerEntryType {
	parts := splitString(s, ",")
	types := make([]service.BalanceLedgerEntryType, 0, len(parts))
	for _, p := range parts {
		p = trimSpace(p)
		if p != "" {
			types = append(types, service.BalanceLedgerEntryType(p))
		}
	}
	return types
}

// convertEntryToResponse 转换流水记录为响应格式
func convertEntryToResponse(entry service.UserBalanceLedgerEntry) BalanceLedgerEntryResponse {
	resp := BalanceLedgerEntryResponse{
		ID:            entry.ID,
		EntryType:     string(entry.EntryType),
		Amount:        entry.Amount,
		BalanceBefore: entry.BalanceBefore,
		BalanceAfter:  entry.BalanceAfter,
		SourceType:    string(entry.SourceType),
		SourceID:      entry.SourceID,
		Description:   entry.Description,
		CreatedAt:     entry.CreatedAt.Format(time.RFC3339),
	}

	// 解析 metadata JSON
	if entry.MetadataJSON != nil && len(entry.MetadataJSON) > 0 {
		var metadata map[string]interface{}
		if err := parseJSON(entry.MetadataJSON, &metadata); err == nil {
			resp.Metadata = metadata
		}
	}

	return resp
}

// 辅助函数（避免导入 strings 包）
func splitString(s, sep string) []string {
	if s == "" {
		return nil
	}
	result := []string{}
	start := 0
	for i := 0; i <= len(s)-len(sep); i++ {
		if s[i:i+len(sep)] == sep {
			result = append(result, s[start:i])
			start = i + len(sep)
		}
	}
	result = append(result, s[start:])
	return result
}

func trimSpace(s string) string {
	start := 0
	end := len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t' || s[start] == '\n' || s[start] == '\r') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\n' || s[end-1] == '\r') {
		end--
	}
	return s[start:end]
}

func parseJSON(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}