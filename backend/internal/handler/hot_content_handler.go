package handler

import (
	"github.com/Aias00/cloudbase/internal/hot"
	"github.com/Aias00/cloudbase/internal/pkg/pagination"
	"github.com/Aias00/cloudbase/internal/pkg/response"
	"github.com/gin-gonic/gin"
)

type HotContentHandler struct {
	service *hot.Service
}

func NewHotContentHandler(service *hot.Service) *HotContentHandler {
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
	items, result, err := h.service.ListItems(c.Request.Context(), pagination.PaginationParams{Page: page, PageSize: pageSize}, hot.ListFilters{
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
