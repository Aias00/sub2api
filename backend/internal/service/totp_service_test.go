//go:build unit

package service

import (
	"context"
	"testing"

	"github.com/Aias00/cloudbase/internal/config"
	"github.com/stretchr/testify/require"
)

type totpSettingRepoStub struct {
	values map[string]string
}

func (s *totpSettingRepoStub) Get(context.Context, string) (*Setting, error) {
	panic("unexpected Get call")
}

func (s *totpSettingRepoStub) GetValue(_ context.Context, key string) (string, error) {
	return s.values[key], nil
}

func (s *totpSettingRepoStub) Set(context.Context, string, string) error {
	panic("unexpected Set call")
}

func (s *totpSettingRepoStub) GetMultiple(context.Context, []string) (map[string]string, error) {
	panic("unexpected GetMultiple call")
}

func (s *totpSettingRepoStub) SetMultiple(context.Context, map[string]string) error {
	panic("unexpected SetMultiple call")
}

func (s *totpSettingRepoStub) GetAll(context.Context) (map[string]string, error) {
	panic("unexpected GetAll call")
}

func (s *totpSettingRepoStub) Delete(context.Context, string) error {
	panic("unexpected Delete call")
}

func TestTotpServiceIssuerPrefersFrontendURLHost(t *testing.T) {
	t.Parallel()

	settingSvc := NewSettingService(&totpSettingRepoStub{
		values: map[string]string{
			SettingKeyFrontendURL: "https://cloudbase.eu.org/login",
			SettingKeySiteName:    "cloudbase",
		},
	}, &config.Config{})

	svc := &TotpService{settingService: settingSvc}

	require.Equal(t, "cloudbase.eu.org", svc.issuer(context.Background()))
}

func TestTotpServiceIssuerFallsBackToSiteName(t *testing.T) {
	t.Parallel()

	settingSvc := NewSettingService(&totpSettingRepoStub{
		values: map[string]string{
			SettingKeySiteName: "cloudbase",
		},
	}, &config.Config{})

	svc := &TotpService{settingService: settingSvc}

	require.Equal(t, "cloudbase", svc.issuer(context.Background()))
}

func TestTotpServiceIssuerFallsBackToDefault(t *testing.T) {
	t.Parallel()

	svc := &TotpService{settingService: NewSettingService(&totpSettingRepoStub{values: map[string]string{}}, &config.Config{})}

	require.Equal(t, defaultTotpIssuer, svc.issuer(context.Background()))
}
