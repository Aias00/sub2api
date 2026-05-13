package service

import (
	"context"
	"fmt"
	"html"
	"net/url"
	"strings"
)

type emailTemplateData struct {
	SiteName    string
	LogoURL     string
	Title       string
	Intro       string
	PrimaryHTML string
	SupportHTML string
	FooterHTML  string
}

type emailDetailRow struct {
	Label string
	Value string
}

func appendEmailHTML(b *strings.Builder, value string) {
	_, _ = b.WriteString(value)
}

func renderProductEmail(data emailTemplateData) string {
	siteName := normalizeEmailSiteName(data.SiteName)
	title := strings.TrimSpace(data.Title)
	if title == "" {
		title = siteName
	}

	var b strings.Builder
	appendEmailHTML(&b, `<!doctype html>
<html>
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>`)
	appendEmailHTML(&b, emailEscape(title))
	appendEmailHTML(&b, `</title>
</head>
<body style="margin:0;padding:0;background:#f8f8f7;color:#202020;font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,Helvetica,Arial,sans-serif;-webkit-font-smoothing:antialiased;">
  <div style="padding:40px 16px;">
    <div style="max-width:920px;margin:0 auto;background:#ffffff;border:1px solid #dedede;border-radius:22px;padding:56px 64px 50px;box-sizing:border-box;">
      <div style="width:58px;height:58px;border:1px solid #ececec;border-radius:14px;background:#ffffff;box-shadow:0 8px 20px rgba(0,0,0,.10);display:flex;align-items:center;justify-content:center;margin-bottom:36px;">`)
	appendEmailHTML(&b, emailLogoHTML(data.LogoURL, siteName))
	appendEmailHTML(&b, `</div>
      <h1 style="margin:0 0 26px;font-size:30px;line-height:1.18;font-weight:700;letter-spacing:-.035em;color:#202020;">`)
	appendEmailHTML(&b, emailEscape(title))
	appendEmailHTML(&b, `</h1>`)

	if strings.TrimSpace(data.Intro) != "" {
		appendEmailHTML(&b, `<p style="margin:0 0 30px;font-size:20px;line-height:1.55;color:#202020;">`)
		appendEmailHTML(&b, emailEscape(data.Intro))
		appendEmailHTML(&b, `</p>`)
	}
	if strings.TrimSpace(data.PrimaryHTML) != "" {
		appendEmailHTML(&b, data.PrimaryHTML)
	}
	if strings.TrimSpace(data.SupportHTML) != "" {
		appendEmailHTML(&b, `<div style="margin-top:28px;">`)
		appendEmailHTML(&b, data.SupportHTML)
		appendEmailHTML(&b, `</div>`)
	}

	appendEmailHTML(&b, `<div style="height:1px;background:#d8d8d8;margin:38px 0 30px;"></div>`)
	if strings.TrimSpace(data.FooterHTML) != "" {
		appendEmailHTML(&b, data.FooterHTML)
	} else {
		appendEmailHTML(&b, emailMutedParagraph("This is an automated message. You can safely ignore it if you did not request it."))
	}
	appendEmailHTML(&b, `
    </div>
  </div>
</body>
</html>`)
	return b.String()
}

func normalizeEmailSiteName(siteName string) string {
	siteName = strings.TrimSpace(siteName)
	if siteName == "" {
		return "Sub2API"
	}
	return siteName
}

func emailEscape(value string) string {
	return html.EscapeString(strings.TrimSpace(value))
}

func emailLogoHTML(logoURL, siteName string) string {
	if logoURL = normalizeEmailImageURL(logoURL, ""); logoURL != "" {
		return fmt.Sprintf(`<img src="%s" width="36" height="36" alt="%s" style="display:block;width:36px;height:36px;border:0;border-radius:10px;object-fit:contain;">`, emailEscape(logoURL), emailEscape(siteName))
	}
	return `<div style="width:30px;height:30px;background:#20201d;border-radius:5px;box-shadow:inset -10px -10px 0 #11110f,inset 10px -10px 0 #34342f;transform:rotate(30deg);"></div>`
}

func resolveEmailLogoURL(ctx context.Context, repo SettingRepository) string {
	if repo == nil {
		return ""
	}
	settings, err := repo.GetMultiple(ctx, []string{
		SettingKeySiteLogo,
		SettingKeyFrontendURL,
		SettingKeyAPIBaseURL,
	})
	if err != nil {
		return ""
	}

	baseURL := firstNonEmpty(settings[SettingKeyFrontendURL], settings[SettingKeyAPIBaseURL])
	siteLogo := settings[SettingKeySiteLogo]
	if logo, ok := parseSiteLogoDataURL(siteLogo); ok {
		if endpoint := emailSiteLogoEndpointURL(baseURL, logo.ETag); endpoint != "" {
			return endpoint
		}
	}
	if logo := normalizeEmailImageURL(siteLogo, baseURL); logo != "" {
		return logo
	}
	return emailDefaultLogoURL(baseURL)
}

func normalizeEmailImageURL(raw, baseURL string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	parsed, err := url.Parse(raw)
	if err != nil {
		return ""
	}
	if isHTTPImageURL(parsed) {
		return parsed.String()
	}
	if strings.HasPrefix(raw, "/") {
		base := emailOriginURL(baseURL)
		if base == nil {
			return ""
		}
		base.Path = raw
		return base.String()
	}
	return ""
}

func emailDefaultLogoURL(baseURL string) string {
	base := emailOriginURL(baseURL)
	if base == nil {
		return ""
	}
	base.Path = "/logo.png"
	return base.String()
}

func emailSiteLogoEndpointURL(baseURL, etag string) string {
	base := emailOriginURL(baseURL)
	if base == nil {
		return ""
	}
	base.Path = "/api/v1/settings/site-logo"
	query := base.Query()
	if etag != "" {
		query.Set("v", strings.Trim(etag, `"`))
	}
	base.RawQuery = query.Encode()
	return base.String()
}

func emailOriginURL(raw string) *url.URL {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || !isHTTPImageURL(parsed) {
		return nil
	}
	return &url.URL{Scheme: parsed.Scheme, Host: parsed.Host}
}

func isHTTPImageURL(parsed *url.URL) bool {
	if parsed == nil || parsed.Host == "" {
		return false
	}
	return parsed.Scheme == "http" || parsed.Scheme == "https"
}

func emailCodeBlock(code string) string {
	return fmt.Sprintf(`<div style="margin:10px 0 14px;font-size:40px;line-height:1.18;font-weight:500;letter-spacing:.075em;color:#202020;font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,Helvetica,Arial,sans-serif;font-variant-numeric:tabular-nums;">%s</div>`, emailEscape(code))
}

func emailHeroValue(value string, color string) string {
	color = strings.TrimSpace(color)
	if color == "" {
		color = "#202020"
	}
	return fmt.Sprintf(`<div style="margin:10px 0 14px;font-size:42px;line-height:1.15;font-weight:700;letter-spacing:-.035em;color:%s;">%s</div>`, emailEscape(color), emailEscape(value))
}

func emailActionButton(label, href string) string {
	return fmt.Sprintf(`<a href="%s" style="display:inline-block;background:#090b12;color:#ffffff;text-decoration:none;border-radius:12px;padding:14px 28px;font-size:16px;line-height:1.4;font-weight:700;">%s</a>`, emailEscape(href), emailEscape(label))
}

func emailFallbackLink(label, href string) string {
	return fmt.Sprintf(`<p style="margin:18px 0 0;font-size:13px;line-height:1.7;color:#6b7280;">%s<br><span style="word-break:break-all;color:#374151;">%s</span></p>`, emailEscape(label), emailEscape(href))
}

func emailMutedParagraph(text string) string {
	return fmt.Sprintf(`<p style="margin:0 0 16px;font-size:15px;line-height:1.65;color:#666666;">%s</p>`, emailEscape(text))
}

func emailDetailsBlock(title string, rows []emailDetailRow) string {
	if len(rows) == 0 {
		return ""
	}
	var b strings.Builder
	if strings.TrimSpace(title) != "" {
		appendEmailHTML(&b, `<h2 style="margin:0 0 16px;font-size:18px;line-height:1.4;color:#202020;">`)
		appendEmailHTML(&b, emailEscape(title))
		appendEmailHTML(&b, `</h2>`)
	}
	appendEmailHTML(&b, `<table role="presentation" cellpadding="0" cellspacing="0" style="width:100%;border-collapse:collapse;margin:0;">`)
	for _, row := range rows {
		appendEmailHTML(&b, `<tr><td style="padding:13px 0;border-bottom:1px solid #eeeeee;color:#666666;font-size:15px;line-height:1.5;">`)
		appendEmailHTML(&b, emailEscape(row.Label))
		appendEmailHTML(&b, `</td><td style="padding:13px 0;border-bottom:1px solid #eeeeee;color:#202020;font-size:15px;line-height:1.5;font-weight:650;text-align:right;">`)
		appendEmailHTML(&b, emailEscape(row.Value))
		appendEmailHTML(&b, `</td></tr>`)
	}
	appendEmailHTML(&b, `</table>`)
	return b.String()
}

func BuildTestEmailBody(siteName string) string {
	return BuildTestEmailBodyWithLogo(siteName, "")
}

func BuildTestEmailBodyWithLogo(siteName, logoURL string) string {
	siteName = normalizeEmailSiteName(siteName)
	return renderProductEmail(emailTemplateData{
		SiteName:    siteName,
		LogoURL:     logoURL,
		Title:       fmt.Sprintf("Test email from %s", siteName),
		Intro:       "Your SMTP configuration is working correctly.",
		PrimaryHTML: emailHeroValue("Sent successfully", "#16a34a"),
		FooterHTML:  emailMutedParagraph("This is an automated test message. No action is required."),
	})
}
