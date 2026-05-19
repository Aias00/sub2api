package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/ent/pendingauthsession"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestLoginRequiresCurrentAgreementWhenEnabled(t *testing.T) {
	handler, client := newOAuthPendingFlowTestHandlerWithDependencies(t, oauthPendingFlowTestHandlerOptions{
		settingValues: loginAgreementTestSettingValues(t),
	})
	ctx := context.Background()
	agreementRevision := currentLoginAgreementTestRevision(t, handler)
	passwordHash, err := handler.authService.HashPassword("Aizazadi2024!")
	require.NoError(t, err)
	user, err := client.User.Create().
		SetEmail("admin@example.com").
		SetPasswordHash(passwordHash).
		SetRole(service.RoleAdmin).
		SetStatus(service.StatusActive).
		Save(ctx)
	require.NoError(t, err)

	recorder := httptest.NewRecorder()
	ginCtx, _ := gin.CreateTestContext(recorder)
	body := bytes.NewBufferString(`{"email":"admin@example.com","password":"Aizazadi2024!"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", body)
	req.Header.Set("Content-Type", "application/json")
	ginCtx.Request = req

	handler.Login(ginCtx)

	require.Equal(t, http.StatusForbidden, recorder.Code)
	require.Contains(t, recorder.Body.String(), "LOGIN_AGREEMENT_REQUIRED")

	recorder = httptest.NewRecorder()
	ginCtx, _ = gin.CreateTestContext(recorder)
	body = bytes.NewBufferString(`{"email":"admin@example.com","password":"Aizazadi2024!","agreement_accepted":true,"agreement_revision":"` + agreementRevision + `"}`)
	req = httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", body)
	req.Header.Set("Content-Type", "application/json")
	ginCtx.Request = req

	handler.Login(ginCtx)

	require.Equal(t, http.StatusOK, recorder.Code)
	updatedUser, err := client.User.Get(ctx, user.ID)
	require.NoError(t, err)
	require.Equal(t, agreementRevision, updatedUser.LoginAgreementAcceptedRevision)
	require.NotNil(t, updatedUser.LoginAgreementAcceptedAt)
}

func TestRegisterRequiresCurrentAgreementWhenEnabled(t *testing.T) {
	handler, client := newOAuthPendingFlowTestHandlerWithDependencies(t, oauthPendingFlowTestHandlerOptions{
		settingValues: loginAgreementTestSettingValues(t),
	})
	ctx := context.Background()
	agreementRevision := currentLoginAgreementTestRevision(t, handler)

	recorder := httptest.NewRecorder()
	ginCtx, _ := gin.CreateTestContext(recorder)
	body := bytes.NewBufferString(`{"email":"user@example.com","password":"Aizazadi2024!"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", body)
	req.Header.Set("Content-Type", "application/json")
	ginCtx.Request = req

	handler.Register(ginCtx)

	require.Equal(t, http.StatusForbidden, recorder.Code)
	require.Contains(t, recorder.Body.String(), "LOGIN_AGREEMENT_REQUIRED")

	recorder = httptest.NewRecorder()
	ginCtx, _ = gin.CreateTestContext(recorder)
	body = bytes.NewBufferString(`{"email":"user@example.com","password":"Aizazadi2024!","agreement_accepted":true,"agreement_revision":"` + agreementRevision + `"}`)
	req = httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", body)
	req.Header.Set("Content-Type", "application/json")
	ginCtx.Request = req

	handler.Register(ginCtx)

	require.Equal(t, http.StatusOK, recorder.Code)
	createdUser, err := client.User.Query().Only(ctx)
	require.NoError(t, err)
	require.Equal(t, agreementRevision, createdUser.LoginAgreementAcceptedRevision)
	require.NotNil(t, createdUser.LoginAgreementAcceptedAt)
}

func TestExchangePendingOAuthCompletionRequiresCurrentAgreementForTokenIssue(t *testing.T) {
	handler, client := newOAuthPendingFlowTestHandlerWithDependencies(t, oauthPendingFlowTestHandlerOptions{
		settingValues: loginAgreementTestSettingValues(t),
	})
	ctx := context.Background()
	agreementRevision := currentLoginAgreementTestRevision(t, handler)

	userEntity, err := client.User.Create().
		SetEmail("oauth@example.com").
		SetUsername("oauth-user").
		SetPasswordHash("hash").
		SetRole(service.RoleUser).
		SetStatus(service.StatusActive).
		Save(ctx)
	require.NoError(t, err)

	session, err := client.PendingAuthSession.Create().
		SetSessionToken("agreement-session-token").
		SetIntent("login").
		SetProviderType("linuxdo").
		SetProviderKey("linuxdo").
		SetProviderSubject("agreement-123").
		SetTargetUserID(userEntity.ID).
		SetResolvedEmail(userEntity.Email).
		SetBrowserSessionKey("agreement-browser-key").
		SetUpstreamIdentityClaims(map[string]any{"username": "oauth-user"}).
		SetLocalFlowState(map[string]any{
			oauthCompletionResponseKey: map[string]any{
				"access_token": "access-token",
				"redirect":     "/dashboard",
			},
		}).
		SetExpiresAt(time.Now().UTC().Add(10 * time.Minute)).
		Save(ctx)
	require.NoError(t, err)

	recorder := httptest.NewRecorder()
	ginCtx, _ := gin.CreateTestContext(recorder)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/oauth/pending/exchange", nil)
	req.AddCookie(&http.Cookie{Name: oauthPendingSessionCookieName, Value: encodeCookieValue(session.SessionToken)})
	req.AddCookie(&http.Cookie{Name: oauthPendingBrowserCookieName, Value: encodeCookieValue(session.BrowserSessionKey)})
	ginCtx.Request = req

	handler.ExchangePendingOAuthCompletion(ginCtx)

	require.Equal(t, http.StatusForbidden, recorder.Code)
	require.Contains(t, recorder.Body.String(), "LOGIN_AGREEMENT_REQUIRED")

	recorder = httptest.NewRecorder()
	ginCtx, _ = gin.CreateTestContext(recorder)
	body, err := json.Marshal(map[string]any{
		"agreement_accepted": true,
		"agreement_revision": agreementRevision,
	})
	require.NoError(t, err)
	req = httptest.NewRequest(http.MethodPost, "/api/v1/auth/oauth/pending/exchange", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: oauthPendingSessionCookieName, Value: encodeCookieValue(session.SessionToken)})
	req.AddCookie(&http.Cookie{Name: oauthPendingBrowserCookieName, Value: encodeCookieValue(session.BrowserSessionKey)})
	ginCtx.Request = req

	handler.ExchangePendingOAuthCompletion(ginCtx)

	require.Equal(t, http.StatusOK, recorder.Code)
	updatedUser, err := client.User.Get(ctx, userEntity.ID)
	require.NoError(t, err)
	require.Equal(t, agreementRevision, updatedUser.LoginAgreementAcceptedRevision)
	require.NotNil(t, updatedUser.LoginAgreementAcceptedAt)

	consumed, err := client.PendingAuthSession.Query().
		Where(pendingauthsession.IDEQ(session.ID)).
		Only(ctx)
	require.NoError(t, err)
	require.NotNil(t, consumed.ConsumedAt)
}

func loginAgreementTestSettingValues(t *testing.T) map[string]string {
	t.Helper()
	docs, err := json.Marshal([]service.LoginAgreementDocument{
		{ID: "terms", Title: "服务条款", ContentMD: "test"},
	})
	require.NoError(t, err)
	return map[string]string{
		service.SettingKeyLoginAgreementEnabled:   "true",
		service.SettingKeyLoginAgreementMode:      "modal",
		service.SettingKeyLoginAgreementUpdatedAt: "2026-05-19",
		service.SettingKeyLoginAgreementDocuments: string(docs),
	}
}

func currentLoginAgreementTestRevision(t *testing.T, handler *AuthHandler) string {
	t.Helper()
	settings, err := handler.settingSvc.GetPublicSettings(context.Background())
	require.NoError(t, err)
	require.NotEmpty(t, settings.LoginAgreementRevision)
	return settings.LoginAgreementRevision
}
