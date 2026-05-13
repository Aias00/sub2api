package service

import (
	"fmt"
	"strings"
)

func BuildWelcomeEmailBodyWithLogo(siteName, logoURL, displayName, dashboardURL, manageURL, unsubscribeURL string) string {
	siteName = normalizeEmailSiteName(siteName)
	displayName = strings.TrimSpace(displayName)
	if displayName == "" {
		displayName = siteName
	}

	var primary strings.Builder
	primary.WriteString(`<div style="text-align:center;margin:6px 0 34px;">`)
	for _, item := range []struct {
		Emoji string
		Bg    string
	}{
		{"😁", "#b8dcff"},
		{"👋", "#bceda9"},
		{"🎉", "#f7b2a9"},
	} {
		primary.WriteString(fmt.Sprintf(`<span style="display:inline-block;width:58px;height:58px;margin:0 6px;border-radius:9px;background:%s;font-size:32px;line-height:58px;text-align:center;">%s</span>`, item.Bg, item.Emoji))
	}
	primary.WriteString(`</div>`)

	primary.WriteString(fmt.Sprintf(`<h2 style="margin:0 0 20px;text-align:center;font-size:34px;line-height:1.18;font-weight:800;letter-spacing:-.04em;color:#202020;">Welcome to %s.</h2>`, emailEscape(siteName)))
	primary.WriteString(fmt.Sprintf(`<p style="margin:0 auto 30px;max-width:660px;text-align:center;font-size:20px;line-height:1.55;color:#202020;">Hi %s, your account is ready. Use one API key to route models, track usage, and keep cost under control.</p>`, emailEscape(displayName)))

	if strings.TrimSpace(dashboardURL) != "" {
		primary.WriteString(`<div style="text-align:center;margin:0 0 36px;">`)
		primary.WriteString(emailActionButton("Open dashboard", dashboardURL))
		primary.WriteString(`</div>`)
	}

	primary.WriteString(`<table role="presentation" cellpadding="0" cellspacing="0" style="width:100%;border-collapse:collapse;margin:0 0 10px;"><tr>`)
	primary.WriteString(`<td style="width:42%;vertical-align:top;padding:0 24px 0 0;"><div style="background:#e2f1ff;border-radius:12px;padding:26px 22px;color:#0f172a;"><div style="font-size:44px;line-height:1;margin-bottom:18px;">🔑</div><div style="font-size:18px;line-height:1.4;font-weight:750;">Create your first API key</div><p style="margin:12px 0 0;font-size:14px;line-height:1.6;color:#475569;">Bind a key to a group or platform, then copy the generated gateway endpoint and authorization header.</p></div></td>`)
	primary.WriteString(`<td style="vertical-align:top;padding:0 0 0 24px;"><h3 style="margin:0 0 14px;font-size:22px;line-height:1.35;color:#202020;">Start with a short path</h3><p style="margin:0 0 16px;font-size:18px;line-height:1.6;color:#202020;">Open the dashboard, create or select an API key, and run the first request from the gateway guide.</p>`)
	if strings.TrimSpace(dashboardURL) != "" {
		primary.WriteString(fmt.Sprintf(`<a href="%s" style="font-size:18px;line-height:1.5;color:#2f6f2f;text-decoration:underline;font-weight:700;">Go to dashboard →</a>`, emailEscape(dashboardURL)))
	}
	primary.WriteString(`</td></tr></table>`)

	footer := emailMutedParagraph(fmt.Sprintf("You're receiving this email because you created a %s account.", siteName))
	links := make([]string, 0, 2)
	if strings.TrimSpace(manageURL) != "" {
		links = append(links, fmt.Sprintf(`<a href="%s" style="color:#202020;text-decoration:underline;font-weight:650;">Manage subscriptions</a>`, emailEscape(manageURL)))
	}
	if strings.TrimSpace(unsubscribeURL) != "" {
		links = append(links, fmt.Sprintf(`<a href="%s" style="color:#202020;text-decoration:underline;font-weight:650;">Unsubscribe</a>`, emailEscape(unsubscribeURL)))
	}
	if len(links) > 0 {
		footer += fmt.Sprintf(`<p style="margin:8px 0 0;font-size:15px;line-height:1.65;color:#666666;">%s</p>`, strings.Join(links, ` <span style="color:#a3a3a3;">|</span> `))
	}

	return renderProductEmail(emailTemplateData{
		SiteName:    siteName,
		LogoURL:     logoURL,
		Title:       fmt.Sprintf("Welcome to %s", siteName),
		PrimaryHTML: primary.String(),
		FooterHTML:  footer,
	})
}
