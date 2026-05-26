package handler

import (
	"context"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/service"
)

type agreementAcceptanceInput struct {
	Accepted bool
	Revision string
}

func (h *AuthHandler) loginAgreementRequirement(ctx context.Context) (bool, string, error) {
	if h == nil {
		return false, "", nil
	}

	settingSvc := h.settingSvc
	if settingSvc == nil && h.authService != nil {
		settingSvc = h.authService.SettingService()
	}
	if settingSvc == nil {
		return false, "", nil
	}

	return settingSvc.GetCurrentLoginAgreementRequirement(ctx)
}

func (h *AuthHandler) ensureAnonymousLoginAgreementAccepted(ctx context.Context, input agreementAcceptanceInput) (string, error) {
	enabled, currentRevision, err := h.loginAgreementRequirement(ctx)
	if err != nil || !enabled {
		return currentRevision, err
	}
	if !input.Accepted || strings.TrimSpace(input.Revision) != currentRevision {
		return currentRevision, service.ErrLoginAgreementRequired.WithMetadata(map[string]string{
			"agreement_revision": currentRevision,
		})
	}
	return currentRevision, nil
}

func (h *AuthHandler) ensureUserAcceptedCurrentLoginAgreement(ctx context.Context, user *service.User, input agreementAcceptanceInput) error {
	enabled, currentRevision, err := h.loginAgreementRequirement(ctx)
	if err != nil || !enabled {
		return err
	}
	if user == nil {
		return service.ErrLoginAgreementRequired.WithMetadata(map[string]string{
			"agreement_revision": currentRevision,
		})
	}
	if !input.Accepted || strings.TrimSpace(input.Revision) != currentRevision {
		return service.ErrLoginAgreementRequired.WithMetadata(map[string]string{
			"agreement_revision": currentRevision,
		})
	}
	return nil
}

func (h *AuthHandler) recordLoginAgreementAcceptance(ctx context.Context, user *service.User, revision string) error {
	_ = ctx
	_ = user
	_ = revision
	return nil
}
