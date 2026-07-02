package service

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"html"
	"net/mail"
	"net/url"
	"strconv"
	"strings"
	"time"

	dbent "github.com/Wei-Shaw/cloudbase/ent"
	dbuser "github.com/Wei-Shaw/cloudbase/ent/user"
	infraerrors "github.com/Wei-Shaw/cloudbase/internal/pkg/errors"
	"github.com/Wei-Shaw/cloudbase/internal/pkg/logger"
)

const emailPreferenceTokenVersion = "v1"

type emailPreferenceLinks struct {
	DashboardURL   string
	ManageURL      string
	UnsubscribeURL string
	SubscribeURL   string
}

type EmailPreferencePage struct {
	SiteName       string
	Title          string
	Message        string
	Email          string
	Unsubscribed   bool
	DashboardURL   string
	ManageURL      string
	UnsubscribeURL string
	SubscribeURL   string
}

func (s *AuthService) sendWelcomeEmailForNewUser(ctx context.Context, user *User, signupSource string) {
	if s == nil || user == nil || user.ID <= 0 || !isDeliverableWelcomeEmail(user.Email) {
		return
	}
	if s.emailQueueService == nil && s.emailService == nil {
		return
	}
	if !s.markWelcomeEmailQueued(ctx, user.ID) {
		return
	}

	siteName := s.welcomeEmailSiteName(ctx)
	links := s.emailPreferenceLinks(ctx, user)
	displayName := firstNonEmpty(user.Username, strings.Split(user.Email, "@")[0], siteName)
	email := strings.TrimSpace(user.Email)

	if s.emailQueueService != nil {
		if err := s.emailQueueService.EnqueueWelcomeEmail(email, siteName, displayName, links.DashboardURL, links.ManageURL, links.UnsubscribeURL); err != nil {
			logger.LegacyPrintf("service.auth", "[Auth] Failed to enqueue welcome email for user %d signup_source=%s: %v", user.ID, signupSource, err)
		}
		return
	}

	go func() {
		sendCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := s.emailService.SendWelcomeEmail(sendCtx, email, siteName, displayName, links.DashboardURL, links.ManageURL, links.UnsubscribeURL); err != nil {
			logger.LegacyPrintf("service.auth", "[Auth] Failed to send welcome email for user %d signup_source=%s: %v", user.ID, signupSource, err)
		}
	}()
}

func (s *AuthService) markWelcomeEmailQueued(ctx context.Context, userID int64) bool {
	if s == nil || s.entClient == nil || userID <= 0 {
		return false
	}
	client := s.entClient
	if tx := dbent.TxFromContext(ctx); tx != nil {
		client = tx.Client()
	}
	affected, err := client.User.Update().
		Where(
			dbuser.IDEQ(userID),
			dbuser.WelcomeEmailSentAtIsNil(),
			dbuser.MarketingEmailsUnsubscribedAtIsNil(),
		).
		SetWelcomeEmailSentAt(time.Now().UTC()).
		Save(ctx)
	if err != nil {
		logger.LegacyPrintf("service.auth", "[Auth] Failed to mark welcome email queued for user %d: %v", userID, err)
		return false
	}
	return affected > 0
}

func (s *AuthService) welcomeEmailSiteName(ctx context.Context) string {
	if s != nil && s.settingService != nil {
		return s.settingService.GetSiteName(ctx)
	}
	return "Cloudbase"
}

func (s *AuthService) emailPreferenceLinks(ctx context.Context, user *User) emailPreferenceLinks {
	if s == nil || user == nil {
		return emailPreferenceLinks{}
	}
	baseURL := s.publicFrontendURL(ctx)
	token := s.emailPreferenceToken(user.ID, user.Email)
	return emailPreferenceLinks{
		DashboardURL:   buildAbsoluteAppURL(baseURL, "/dashboard", ""),
		ManageURL:      buildAbsoluteAppURL(baseURL, "/api/v1/email-preferences/manage", token),
		UnsubscribeURL: buildAbsoluteAppURL(baseURL, "/api/v1/email-preferences/unsubscribe", token),
		SubscribeURL:   buildAbsoluteAppURL(baseURL, "/api/v1/email-preferences/subscribe", token),
	}
}

func (s *AuthService) publicFrontendURL(ctx context.Context) string {
	if s != nil && s.settingService != nil {
		if baseURL := strings.TrimSpace(s.settingService.GetFrontendURL(ctx)); baseURL != "" {
			return baseURL
		}
	}
	return ""
}

func buildAbsoluteAppURL(baseURL, path, token string) string {
	parsed, err := url.Parse(strings.TrimSpace(baseURL))
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return ""
	}
	parsed.Path = path
	query := parsed.Query()
	if strings.TrimSpace(token) != "" {
		query.Set("token", token)
	}
	parsed.RawQuery = query.Encode()
	parsed.Fragment = ""
	return parsed.String()
}

func (s *AuthService) emailPreferenceToken(userID int64, email string) string {
	payload := fmt.Sprintf("%s:%d:%s", emailPreferenceTokenVersion, userID, emailPreferenceHash(email))
	mac := hmac.New(sha256.New, []byte(s.emailPreferenceSecret()))
	_, _ = mac.Write([]byte(payload))
	signed := payload + ":" + hex.EncodeToString(mac.Sum(nil))
	return base64.RawURLEncoding.EncodeToString([]byte(signed))
}

func (s *AuthService) emailPreferenceSecret() string {
	if s != nil && s.cfg != nil && strings.TrimSpace(s.cfg.JWT.Secret) != "" {
		return strings.TrimSpace(s.cfg.JWT.Secret)
	}
	return "cloudbase-email-preference-development-secret"
}

func emailPreferenceHash(email string) string {
	sum := sha256.Sum256([]byte(strings.ToLower(strings.TrimSpace(email))))
	return hex.EncodeToString(sum[:])
}

func (s *AuthService) userFromEmailPreferenceToken(ctx context.Context, token string) (*User, error) {
	token = strings.TrimSpace(token)
	if token == "" || len(token) > 2048 {
		return nil, infraerrors.BadRequest("EMAIL_PREFERENCE_TOKEN_INVALID", "invalid email preference token")
	}
	raw, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil {
		return nil, infraerrors.BadRequest("EMAIL_PREFERENCE_TOKEN_INVALID", "invalid email preference token")
	}
	parts := strings.Split(string(raw), ":")
	if len(parts) != 4 || parts[0] != emailPreferenceTokenVersion {
		return nil, infraerrors.BadRequest("EMAIL_PREFERENCE_TOKEN_INVALID", "invalid email preference token")
	}
	userID, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil || userID <= 0 || parts[2] == "" || parts[3] == "" {
		return nil, infraerrors.BadRequest("EMAIL_PREFERENCE_TOKEN_INVALID", "invalid email preference token")
	}

	payload := strings.Join(parts[:3], ":")
	mac := hmac.New(sha256.New, []byte(s.emailPreferenceSecret()))
	_, _ = mac.Write([]byte(payload))
	expected := hex.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(expected), []byte(parts[3])) {
		return nil, infraerrors.BadRequest("EMAIL_PREFERENCE_TOKEN_INVALID", "invalid email preference token")
	}
	if s == nil || s.userRepo == nil {
		return nil, ErrServiceUnavailable
	}
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, infraerrors.BadRequest("EMAIL_PREFERENCE_TOKEN_INVALID", "invalid email preference token")
	}
	if !hmac.Equal([]byte(emailPreferenceHash(user.Email)), []byte(parts[2])) {
		return nil, infraerrors.BadRequest("EMAIL_PREFERENCE_TOKEN_INVALID", "invalid email preference token")
	}
	return user, nil
}

func (s *AuthService) ManageEmailPreferencePage(ctx context.Context, token string) (*EmailPreferencePage, error) {
	user, err := s.userFromEmailPreferenceToken(ctx, token)
	if err != nil {
		return nil, err
	}
	links := s.emailPreferenceLinks(ctx, user)
	return &EmailPreferencePage{
		SiteName:       s.welcomeEmailSiteName(ctx),
		Title:          "Email preferences",
		Message:        "Manage onboarding and product update emails for this account.",
		Email:          user.Email,
		Unsubscribed:   user.MarketingEmailsUnsubscribedAt != nil,
		DashboardURL:   links.DashboardURL,
		ManageURL:      links.ManageURL,
		UnsubscribeURL: links.UnsubscribeURL,
		SubscribeURL:   links.SubscribeURL,
	}, nil
}

func (s *AuthService) UnsubscribeMarketingEmails(ctx context.Context, token string) (*EmailPreferencePage, error) {
	user, err := s.userFromEmailPreferenceToken(ctx, token)
	if err != nil {
		return nil, err
	}
	if s.entClient == nil {
		return nil, ErrServiceUnavailable
	}
	if _, err := s.entClient.User.UpdateOneID(user.ID).
		SetMarketingEmailsUnsubscribedAt(time.Now().UTC()).
		Save(ctx); err != nil {
		return nil, ErrServiceUnavailable
	}
	page, err := s.ManageEmailPreferencePage(ctx, token)
	if err != nil {
		return nil, err
	}
	page.Title = "Unsubscribed"
	page.Message = "You will no longer receive onboarding and product update emails."
	page.Unsubscribed = true
	return page, nil
}

func (s *AuthService) SubscribeMarketingEmails(ctx context.Context, token string) (*EmailPreferencePage, error) {
	user, err := s.userFromEmailPreferenceToken(ctx, token)
	if err != nil {
		return nil, err
	}
	if s.entClient == nil {
		return nil, ErrServiceUnavailable
	}
	if _, err := s.entClient.User.UpdateOneID(user.ID).
		ClearMarketingEmailsUnsubscribedAt().
		Save(ctx); err != nil {
		return nil, ErrServiceUnavailable
	}
	page, err := s.ManageEmailPreferencePage(ctx, token)
	if err != nil {
		return nil, err
	}
	page.Title = "Subscribed"
	page.Message = "You are subscribed to onboarding and product update emails."
	page.Unsubscribed = false
	return page, nil
}

func RenderEmailPreferencePageHTML(page *EmailPreferencePage) string {
	if page == nil {
		page = &EmailPreferencePage{SiteName: "Cloudbase", Title: "Email preferences", Message: "The preference link is invalid or expired."}
	}
	siteName := normalizeEmailSiteName(page.SiteName)
	status := "Subscribed"
	actionHTML := ""
	if page.Unsubscribed {
		status = "Unsubscribed"
		if page.SubscribeURL != "" {
			actionHTML = fmt.Sprintf(`<a href="%s" style="display:inline-block;background:#090b12;color:#fff;text-decoration:none;border-radius:12px;padding:13px 22px;font-weight:750;">Subscribe again</a>`, html.EscapeString(page.SubscribeURL))
		}
	} else if page.UnsubscribeURL != "" {
		actionHTML = fmt.Sprintf(`<a href="%s" style="display:inline-block;background:#090b12;color:#fff;text-decoration:none;border-radius:12px;padding:13px 22px;font-weight:750;">Unsubscribe</a>`, html.EscapeString(page.UnsubscribeURL))
	}
	dashboardHTML := ""
	if page.DashboardURL != "" {
		dashboardHTML = fmt.Sprintf(`<a href="%s" style="display:inline-block;margin-left:10px;color:#202020;text-decoration:underline;font-weight:700;">Open dashboard</a>`, html.EscapeString(page.DashboardURL))
	}
	return fmt.Sprintf(`<!doctype html>
<html><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><title>%s</title></head>
<body style="margin:0;background:#f8f8f7;color:#202020;font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,Helvetica,Arial,sans-serif;">
<main style="min-height:100vh;display:flex;align-items:center;justify-content:center;padding:32px 16px;box-sizing:border-box;">
<section style="width:100%%;max-width:720px;background:#fff;border:1px solid #dedede;border-radius:22px;padding:42px 46px;box-sizing:border-box;">
<p style="margin:0 0 8px;letter-spacing:.22em;text-transform:uppercase;color:#6b7280;font-size:12px;font-weight:800;">%s</p>
<h1 style="margin:0 0 18px;font-size:34px;line-height:1.15;letter-spacing:-.035em;">%s</h1>
<p style="margin:0 0 22px;font-size:18px;line-height:1.6;color:#4b5563;">%s</p>
<div style="margin:0 0 28px;padding:18px 20px;border:1px solid #e5e7eb;border-radius:14px;background:#fafafa;">
<div style="font-size:13px;color:#6b7280;margin-bottom:6px;">Account</div>
<div style="font-size:18px;font-weight:700;">%s</div>
<div style="margin-top:10px;font-size:15px;color:#4b5563;">Status: <strong style="color:#202020;">%s</strong></div>
</div>
<div>%s%s</div>
</section></main></body></html>`,
		html.EscapeString(page.Title),
		html.EscapeString(siteName),
		html.EscapeString(page.Title),
		html.EscapeString(page.Message),
		html.EscapeString(page.Email),
		status,
		actionHTML,
		dashboardHTML,
	)
}

func isDeliverableWelcomeEmail(email string) bool {
	normalized := strings.ToLower(strings.TrimSpace(email))
	if normalized == "" || isReservedEmail(normalized) {
		return false
	}
	_, err := mail.ParseAddress(normalized)
	return err == nil
}
