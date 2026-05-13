package service

import (
	"context"
	"fmt"
)

// SendWelcomeEmail sends the first-run onboarding email. This is intentionally
// separate from verification/reset emails so unsubscribe preferences can apply
// only to non-transactional onboarding content.
func (s *EmailService) SendWelcomeEmail(ctx context.Context, email, siteName, displayName, dashboardURL, manageURL, unsubscribeURL string) error {
	siteName = normalizeEmailSiteName(siteName)
	subject := fmt.Sprintf("Welcome to %s", siteName)
	body := BuildWelcomeEmailBodyWithLogo(siteName, resolveEmailLogoURL(ctx, s.settingRepo), displayName, dashboardURL, manageURL, unsubscribeURL)
	if err := s.SendEmail(ctx, email, subject, body); err != nil {
		return fmt.Errorf("send welcome email: %w", err)
	}
	return nil
}
