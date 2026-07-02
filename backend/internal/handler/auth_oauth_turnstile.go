package handler

import (
	"strings"

	"github.com/Wei-Shaw/cloudbase/internal/pkg/ip"
	"github.com/gin-gonic/gin"
)

func (h *AuthHandler) verifyTurnstileForOAuthStart(c *gin.Context) error {
	if h == nil || h.authService == nil {
		return nil
	}
	token := strings.TrimSpace(c.Query("turnstile_token"))
	return h.authService.VerifyTurnstile(c.Request.Context(), token, ip.GetClientIP(c))
}
