package handler

import (
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

type HotContentHandler struct {
	service *service.HotContentService
}

func NewHotContentHandler(service *service.HotContentService) *HotContentHandler {
	return &HotContentHandler{service: service}
}

func (h *HotContentHandler) ListSources(c *gin.Context) {
	items, err := h.service.ListSources(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"items": items})
}

func (h *HotContentHandler) ListItems(c *gin.Context) {
	page, pageSize := response.ParsePagination(c)
	items, result, err := h.service.ListItems(c.Request.Context(), pagination.PaginationParams{Page: page, PageSize: pageSize}, service.HotContentListFilters{
		SourceID: c.Query("source_id"),
		Query:    c.Query("q"),
		Status:   c.Query("status"),
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.PaginatedWithResult(c, items, &response.PaginationResult{
		Total:    result.Total,
		Page:     result.Page,
		PageSize: result.PageSize,
		Pages:    result.Pages,
	})
}

func (h *HotContentHandler) ListRunEvents(c *gin.Context) {
	page, pageSize := response.ParsePagination(c)
	items, result, err := h.service.ListRunEvents(c.Request.Context(), c.Query("run_id"), pagination.PaginationParams{Page: page, PageSize: pageSize})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.PaginatedWithResult(c, items, &response.PaginationResult{
		Total:    result.Total,
		Page:     result.Page,
		PageSize: result.PageSize,
		Pages:    result.Pages,
	})
}
