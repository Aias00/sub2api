package service

import (
	"context"
	"strings"
)

type registrationInvitationLoader func(context.Context, string) (*RedeemCode, error)

// resolveRegistrationInvitationGate keeps one-time invitation codes and
// reusable affiliate invite codes compatible when registration is invite-only.
func (s *AuthService) resolveRegistrationInvitationGate(
	ctx context.Context,
	invitationCode string,
	affiliateCode string,
	loadInvitation registrationInvitationLoader,
) (*RedeemCode, string, error) {
	effectiveAffiliateCode := strings.TrimSpace(affiliateCode)
	if s == nil || s.settingService == nil || !s.settingService.IsInvitationCodeEnabled(ctx) {
		return nil, effectiveAffiliateCode, nil
	}

	invitationCode = strings.TrimSpace(invitationCode)
	if invitationCode == "" {
		if effectiveAffiliateCode != "" && s.CanUseAffiliateCodeAsRegistrationInvite(ctx, effectiveAffiliateCode) {
			return nil, effectiveAffiliateCode, nil
		}
		return nil, effectiveAffiliateCode, ErrInvitationCodeRequired
	}

	if loadInvitation != nil {
		if redeemCode, err := loadInvitation(ctx, invitationCode); err == nil && redeemCode != nil {
			if redeemCode.Type == RedeemTypeInvitation && redeemCode.Status == StatusUnused {
				return redeemCode, effectiveAffiliateCode, nil
			}
		}
	}

	if s.CanUseAffiliateCodeAsRegistrationInvite(ctx, invitationCode) {
		if effectiveAffiliateCode == "" {
			effectiveAffiliateCode = invitationCode
		}
		return nil, effectiveAffiliateCode, nil
	}
	if effectiveAffiliateCode != "" && s.CanUseAffiliateCodeAsRegistrationInvite(ctx, effectiveAffiliateCode) {
		return nil, effectiveAffiliateCode, nil
	}

	return nil, effectiveAffiliateCode, ErrInvitationCodeInvalid
}

func (s *AuthService) loadRedeemRegistrationInvitation(ctx context.Context, code string) (*RedeemCode, error) {
	if s == nil || s.redeemRepo == nil {
		return nil, ErrRedeemCodeNotFound
	}
	return s.redeemRepo.GetByCode(ctx, code)
}

// CanUseAffiliateCodeAsRegistrationInvite reports whether a reusable friend
// invite code can satisfy the invite-only registration gate.
func (s *AuthService) CanUseAffiliateCodeAsRegistrationInvite(ctx context.Context, code string) bool {
	if s == nil || s.affiliateService == nil {
		return false
	}
	return s.affiliateService.ValidateInviterCode(ctx, code) == nil
}
