//go:build unit

package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestEmailService_GetSMTPConfigs_IncludesPrimaryAndFallbackChannels(t *testing.T) {
	repo := &settingRepoStub{values: map[string]string{
		SettingKeySMTPHost:       "smtp.primary.test",
		SettingKeySMTPPort:       "465",
		SettingKeySMTPUsername:   "primary-user",
		SettingKeySMTPPassword:   "primary-pass",
		SettingKeySMTPFrom:       "primary@example.com",
		SettingKeySMTPFromName:   "Primary",
		SettingKeySMTPUseTLS:     "true",
		SettingKeySMTPDailyLimit: "100",
		SettingKeySMTPChannels: `[{
			"id":"backup-b",
			"name":"Backup B",
			"enabled":true,
			"host":"smtp.backup-b.test",
			"port":587,
			"username":"backup-b-user",
			"password":"backup-b-pass",
			"from_email":"backup-b@example.com",
			"from_name":"Backup B",
			"use_tls":true,
			"daily_limit":50,
			"sort_order":2
		},{
			"id":"disabled",
			"enabled":false,
			"host":"smtp.disabled.test",
			"port":587
		},{
			"id":"backup-a",
			"name":"Backup A",
			"enabled":true,
			"host":"smtp.backup-a.test",
			"port":2525,
			"username":"backup-a-user",
			"password":"backup-a-pass",
			"from_email":"backup-a@example.com",
			"from_name":"Backup A",
			"use_tls":false,
			"daily_limit":25,
			"sort_order":1
		}]`,
	}}
	svc := NewEmailService(repo, nil)

	configs, err := svc.GetSMTPConfigs(context.Background())
	require.NoError(t, err)
	require.Len(t, configs, 3)
	require.Equal(t, legacySMTPChannelID, configs[0].ID)
	require.Equal(t, "smtp.primary.test", configs[0].Host)
	require.Equal(t, 100, configs[0].DailyLimit)
	require.Equal(t, "backup-a", configs[1].ID)
	require.Equal(t, "smtp.backup-a.test", configs[1].Host)
	require.Equal(t, 25, configs[1].DailyLimit)
	require.Equal(t, "backup-b", configs[2].ID)
}

func TestEmailService_SendEmailWithFallback_SkipsLimitedChannel(t *testing.T) {
	cache := &smtpChannelCacheStub{counts: map[string]int64{
		"smtp_channel:primary": 1,
	}}
	svc := NewEmailService(nil, cache)
	var sentVia []string

	err := svc.sendEmailWithFallback(
		context.Background(),
		[]*SMTPConfig{
			{ID: legacySMTPChannelID, Host: "smtp.primary.test", DailyLimit: 1},
			{ID: "backup", Host: "smtp.backup.test", DailyLimit: 0},
		},
		"user@example.com",
		"subject",
		"body",
		func(config *SMTPConfig, to, subject, body string) error {
			sentVia = append(sentVia, config.ID)
			return nil
		},
	)

	require.NoError(t, err)
	require.Equal(t, []string{"backup"}, sentVia)
}

func TestEmailService_SendEmailWithFallback_TriesNextChannelAfterSendFailure(t *testing.T) {
	svc := NewEmailService(nil, nil)
	var sentVia []string

	err := svc.sendEmailWithFallback(
		context.Background(),
		[]*SMTPConfig{
			{ID: "primary", Host: "smtp.primary.test"},
			{ID: "backup", Host: "smtp.backup.test"},
		},
		"user@example.com",
		"subject",
		"body",
		func(config *SMTPConfig, to, subject, body string) error {
			sentVia = append(sentVia, config.ID)
			if config.ID == "primary" {
				return errors.New("temporary smtp failure")
			}
			return nil
		},
	)

	require.NoError(t, err)
	require.Equal(t, []string{"primary", "backup"}, sentVia)
}

func TestEmailService_SendEmailWithFallback_ReturnsLimitWhenAllChannelsLimited(t *testing.T) {
	cache := &smtpChannelCacheStub{counts: map[string]int64{
		"smtp_channel:primary": 1,
		"smtp_channel:backup":  1,
	}}
	svc := NewEmailService(nil, cache)

	err := svc.sendEmailWithFallback(
		context.Background(),
		[]*SMTPConfig{
			{ID: legacySMTPChannelID, Host: "smtp.primary.test", DailyLimit: 1},
			{ID: "backup", Host: "smtp.backup.test", DailyLimit: 1},
		},
		"user@example.com",
		"subject",
		"body",
		func(config *SMTPConfig, to, subject, body string) error {
			t.Fatalf("sender should not be called when every channel is limited")
			return nil
		},
	)

	require.ErrorIs(t, err, ErrEmailChannelDailyLimit)
}

type smtpChannelCacheStub struct {
	counts map[string]int64
}

func (s *smtpChannelCacheStub) GetVerificationCode(context.Context, string) (*VerificationCodeData, error) {
	return nil, nil
}
func (s *smtpChannelCacheStub) SetVerificationCode(context.Context, string, *VerificationCodeData, time.Duration) error {
	return nil
}
func (s *smtpChannelCacheStub) DeleteVerificationCode(context.Context, string) error { return nil }
func (s *smtpChannelCacheStub) GetNotifyVerifyCode(context.Context, string) (*VerificationCodeData, error) {
	return nil, nil
}
func (s *smtpChannelCacheStub) SetNotifyVerifyCode(context.Context, string, *VerificationCodeData, time.Duration) error {
	return nil
}
func (s *smtpChannelCacheStub) DeleteNotifyVerifyCode(context.Context, string) error { return nil }
func (s *smtpChannelCacheStub) GetPasswordResetToken(context.Context, string) (*PasswordResetTokenData, error) {
	return nil, nil
}
func (s *smtpChannelCacheStub) SetPasswordResetToken(context.Context, string, *PasswordResetTokenData, time.Duration) error {
	return nil
}
func (s *smtpChannelCacheStub) DeletePasswordResetToken(context.Context, string) error { return nil }
func (s *smtpChannelCacheStub) IsPasswordResetEmailInCooldown(context.Context, string) bool {
	return false
}
func (s *smtpChannelCacheStub) SetPasswordResetEmailCooldown(context.Context, string, time.Duration) error {
	return nil
}
func (s *smtpChannelCacheStub) IncrNotifyCodeUserRate(context.Context, int64, time.Duration) (int64, error) {
	return 0, nil
}
func (s *smtpChannelCacheStub) GetNotifyCodeUserRate(context.Context, int64) (int64, error) {
	return 0, nil
}
func (s *smtpChannelCacheStub) IncrActiveEmailDailyRate(_ context.Context, scope string, _ time.Duration) (int64, error) {
	if s.counts == nil {
		s.counts = map[string]int64{}
	}
	s.counts[scope]++
	return s.counts[scope], nil
}
func (s *smtpChannelCacheStub) GetActiveEmailDailyRate(_ context.Context, scope string) (int64, error) {
	return s.counts[scope], nil
}
