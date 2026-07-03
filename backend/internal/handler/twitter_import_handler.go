package handler

import (
	imagectx "github.com/Aias00/cloudbase/internal/image"
	"github.com/Aias00/cloudbase/internal/pkg/response"
	"github.com/gin-gonic/gin"
)

type TwitterImportHandler struct {
	service *imagectx.TwitterImportService
}

func NewTwitterImportHandler(service *imagectx.TwitterImportService) *TwitterImportHandler {
	return &TwitterImportHandler{service: service}
}

type twitterImportRequest struct {
	URL       string   `json:"url" binding:"required"`
	Prompt    string   `json:"prompt"`
	Title     string   `json:"title"`
	Category  string   `json:"category"`
	ImageURLs []string `json:"image_urls"`
	XAuto     *bool    `json:"x_auto"`
}

type twitterImportResponse struct {
	Item         promptCatalogCaseDTO `json:"item"`
	ImageURLs    []string             `json:"image_urls"`
	UploadedURLs []string             `json:"uploaded_urls"`
	Warnings     []string             `json:"warnings"`
}

func (h *TwitterImportHandler) Import(c *gin.Context) {
	var req twitterImportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequestWithError(c, err)
		return
	}

	result, err := h.service.Import(c.Request.Context(), imagectx.TwitterImportInput{
		URL:       req.URL,
		Prompt:    req.Prompt,
		Title:     req.Title,
		Category:  req.Category,
		ImageURLs: req.ImageURLs,
		XAuto:     req.XAuto,
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, twitterImportResponse{
		Item:         promptCatalogCaseDTOFromService(result.Item),
		ImageURLs:    nonNilStrings(result.ImageURLs),
		UploadedURLs: nonNilStrings(result.UploadedURLs),
		Warnings:     nonNilStrings(result.Warnings),
	})
}
