package service

import (
	"fmt"
	"strings"
	"time"
)

func buildContentModerationViolationEmailBody(siteName string, log *ContentModerationLog, cfg *ContentModerationConfig) string {
	if log == nil {
		return ""
	}
	userName := strings.TrimSpace(log.UserEmail)
	if userName == "" && log.UserID != nil {
		userName = fmt.Sprintf("UID %d", *log.UserID)
	}
	threshold := cfg.BanThreshold
	if threshold <= 0 {
		threshold = defaultContentModerationBanThreshold
	}
	statusBlock := ""
	if log.AutoBanned {
		statusBlock = emailHeroValue("账户当前处于封禁状态，所有 API 请求将被拒绝", "#dc2626")
	}
	return renderProductEmail(emailTemplateData{
		SiteName: normalizeEmailSiteName(siteName),
		Title:    "账户触发内容审计规则",
		Intro:    fmt.Sprintf("尊敬的用户 %s，您的 API 请求在内容审计中触发平台风控策略。详情如下。", userName),
		SupportHTML: emailDetailsBlock("触发详情", []emailDetailRow{
			{Label: "触发时间", Value: time.Now().Format("2006-01-02 15:04:05")},
			{Label: "触发来源", Value: "内容审核"},
			{Label: "所属分组", Value: defaultContentModerationString(log.GroupName, "-")},
			{Label: "命中类别", Value: fmt.Sprintf("%s / %.3f", defaultContentModerationString(log.HighestCategory, "-"), log.HighestScore)},
			{Label: "累计触发次数", Value: fmt.Sprintf("%d 次（阈值 %d）", log.ViolationCount, threshold)},
		}) + statusBlock,
		FooterHTML: emailMutedParagraph(fmt.Sprintf("此邮件由 %s 自动发送，请勿回复。", normalizeEmailSiteName(siteName))),
	})
}

func buildContentModerationAccountDisabledEmailBody(siteName string, log *ContentModerationLog, cfg *ContentModerationConfig) string {
	if log == nil {
		return ""
	}
	userName := strings.TrimSpace(log.UserEmail)
	if userName == "" && log.UserID != nil {
		userName = fmt.Sprintf("UID %d", *log.UserID)
	}
	threshold := cfg.BanThreshold
	if threshold <= 0 {
		threshold = defaultContentModerationBanThreshold
	}
	return renderProductEmail(emailTemplateData{
		SiteName: normalizeEmailSiteName(siteName),
		Title:    "账户已被自动禁用",
		Intro:    fmt.Sprintf("尊敬的用户 %s，您的账户在计数周期内多次触发平台风控策略，系统已自动禁用该账户。详情如下。", userName),
		SupportHTML: emailDetailsBlock("封禁详情", []emailDetailRow{
			{Label: "封禁时间", Value: time.Now().Format("2006-01-02 15:04:05")},
			{Label: "触发来源", Value: "内容审核"},
			{Label: "所属分组", Value: defaultContentModerationString(log.GroupName, "-")},
			{Label: "命中类别", Value: fmt.Sprintf("%s / %.3f", defaultContentModerationString(log.HighestCategory, "-"), log.HighestScore)},
			{Label: "累计触发次数", Value: fmt.Sprintf("%d 次（阈值 %d）", log.ViolationCount, threshold)},
		}) + emailHeroValue("账户当前处于封禁状态，所有 API 请求将被拒绝", "#dc2626"),
		FooterHTML: emailMutedParagraph("如需申诉或恢复账号，请联系平台管理员处理。") +
			emailMutedParagraph(fmt.Sprintf("此邮件由 %s 自动发送，请勿回复。", normalizeEmailSiteName(siteName))),
	})
}

func defaultContentModerationString(value string, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return strings.TrimSpace(value)
}
