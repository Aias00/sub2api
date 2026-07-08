package admin

import (
	"github.com/Aias00/cloudbase/internal/gateway/vendorpreset"
	"github.com/Aias00/cloudbase/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

// ListUpstreamPresets returns the curated catalog of well-known upstream
// vendors (base URL + default models + API style) so the admin UI can offer a
// "pick a vendor -> auto-fill" dropdown when creating an upstream/apikey account.
//
// GET /api/v1/admin/upstream/presets
func (h *AccountHandler) ListUpstreamPresets(c *gin.Context) {
	response.Success(c, gin.H{
		"presets": vendorpreset.All(),
	})
}
