package service

import (
	"context"
	"fmt"

	"github.com/Aias00/cloudbase/internal/payment"
)

const paymentResultReturnPath = payment.PaymentResultReturnPath

const (
	PaymentSourceHostedRedirect    = payment.PaymentSourceHostedRedirect
	PaymentSourceWechatInAppResume = payment.PaymentSourceWechatInAppResume

	SettingPaymentVisibleMethodAlipaySource  = payment.SettingPaymentVisibleMethodAlipaySource
	SettingPaymentVisibleMethodWxpaySource   = payment.SettingPaymentVisibleMethodWxpaySource
	SettingPaymentVisibleMethodAlipayEnabled = payment.SettingPaymentVisibleMethodAlipayEnabled
	SettingPaymentVisibleMethodWxpayEnabled  = payment.SettingPaymentVisibleMethodWxpayEnabled

	VisibleMethodSourceOfficialAlipay = payment.VisibleMethodSourceOfficialAlipay
	VisibleMethodSourceEasyPayAlipay  = payment.VisibleMethodSourceEasyPayAlipay
	VisibleMethodSourceOfficialWechat = payment.VisibleMethodSourceOfficialWechat
	VisibleMethodSourceEasyPayWechat  = payment.VisibleMethodSourceEasyPayWechat

	wechatPaymentResumeTokenType = payment.WeChatPaymentResumeTokenType

	paymentResumeNotConfiguredCode    = payment.PaymentResumeNotConfiguredCode
	paymentResumeNotConfiguredMessage = payment.PaymentResumeNotConfiguredMessage

	paymentResumeTokenTTL       = payment.PaymentResumeTokenTTL
	wechatPaymentResumeTokenTTL = payment.WeChatPaymentResumeTokenTTL
)

type ResumeTokenClaims = payment.ResumeTokenClaims
type WeChatPaymentResumeClaims = payment.WeChatPaymentResumeClaims
type PaymentResumeService = payment.ResumeService

type visibleMethodLoadBalancer struct {
	inner         payment.LoadBalancer
	configService *PaymentConfigService
}

func NewPaymentResumeService(signingKey []byte, verifyFallbacks ...[]byte) *PaymentResumeService {
	return payment.NewResumeService(signingKey, verifyFallbacks...)
}

func NormalizeVisibleMethod(method string) string {
	return payment.NormalizeVisibleMethod(method)
}

func NormalizeVisibleMethods(methods []string) []string {
	return payment.NormalizeVisibleMethods(methods)
}

func NormalizePaymentSource(source string) string {
	return payment.NormalizePaymentSource(source)
}

func NormalizeVisibleMethodSource(method, source string) string {
	return payment.NormalizeVisibleMethodSource(method, source)
}

func VisibleMethodProviderKeyForSource(method, source string) (string, bool) {
	return payment.VisibleMethodProviderKeyForSource(method, source)
}

func newVisibleMethodLoadBalancer(inner payment.LoadBalancer, configService *PaymentConfigService) payment.LoadBalancer {
	if inner == nil || configService == nil || configService.entClient == nil {
		return inner
	}
	return &visibleMethodLoadBalancer{inner: inner, configService: configService}
}

func (lb *visibleMethodLoadBalancer) GetInstanceConfig(ctx context.Context, instanceID int64) (map[string]string, error) {
	return lb.inner.GetInstanceConfig(ctx, instanceID)
}

func (lb *visibleMethodLoadBalancer) SelectInstance(ctx context.Context, providerKey string, paymentType payment.PaymentType, strategy payment.Strategy, orderAmount float64) (*payment.InstanceSelection, error) {
	visibleMethod := NormalizeVisibleMethod(paymentType)
	if providerKey != "" || (visibleMethod != payment.TypeAlipay && visibleMethod != payment.TypeWxpay) {
		return lb.inner.SelectInstance(ctx, providerKey, paymentType, strategy, orderAmount)
	}

	inst, err := lb.configService.resolveEnabledVisibleMethodInstance(ctx, visibleMethod)
	if err != nil {
		return nil, err
	}
	if inst == nil {
		return nil, fmt.Errorf("visible payment method %s has no enabled provider instance", visibleMethod)
	}
	return lb.inner.SelectInstance(ctx, inst.ProviderKey, paymentType, strategy, orderAmount)
}

func visibleMethodEnabledSettingKey(method string) string {
	return payment.VisibleMethodEnabledSettingKey(method)
}

func visibleMethodSourceSettingKey(method string) string {
	return payment.VisibleMethodSourceSettingKey(method)
}

func CanonicalizeReturnURL(raw string, srcHost string, srcURL string) (string, error) {
	return payment.CanonicalizeReturnURL(raw, srcHost, srcURL)
}

func buildPaymentReturnURL(base string, orderID int64, outTradeNo string, resumeToken string) (string, error) {
	return payment.BuildPaymentReturnURL(base, orderID, outTradeNo, resumeToken)
}

func validatePaymentResumeExpiry(expiresAt int64, code, message string) error {
	return payment.ValidatePaymentResumeExpiry(expiresAt, code, message)
}

func signPaymentResumePayload(payload string, key []byte) string {
	return payment.SignPaymentResumePayload(payload, key)
}
