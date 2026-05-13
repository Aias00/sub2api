package handler

import (
	"net/http"

	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

func (h *AuthHandler) ManageEmailPreferences(c *gin.Context) {
	h.renderEmailPreferencePage(c, func(token string) (*service.EmailPreferencePage, error) {
		return h.authService.ManageEmailPreferencePage(c.Request.Context(), token)
	})
}

func (h *AuthHandler) UnsubscribeMarketingEmails(c *gin.Context) {
	h.renderEmailPreferencePage(c, func(token string) (*service.EmailPreferencePage, error) {
		return h.authService.UnsubscribeMarketingEmails(c.Request.Context(), token)
	})
}

func (h *AuthHandler) SubscribeMarketingEmails(c *gin.Context) {
	h.renderEmailPreferencePage(c, func(token string) (*service.EmailPreferencePage, error) {
		return h.authService.SubscribeMarketingEmails(c.Request.Context(), token)
	})
}

func (h *AuthHandler) renderEmailPreferencePage(c *gin.Context, load func(token string) (*service.EmailPreferencePage, error)) {
	status := http.StatusOK
	page, err := load(c.Query("token"))
	if err != nil {
		status = http.StatusBadRequest
		siteName := "Sub2API"
		if h != nil && h.settingSvc != nil {
			siteName = h.settingSvc.GetSiteName(c.Request.Context())
		}
		page = &service.EmailPreferencePage{
			SiteName: siteName,
			Title:    "Email preferences",
			Message:  "This email preference link is invalid or expired.",
		}
	}
	c.Data(status, "text/html; charset=utf-8", []byte(service.RenderEmailPreferencePageHTML(page)))
}
