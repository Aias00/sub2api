package service

import (
	"fmt"
	"html"
	"strings"
)

type emailTemplateData struct {
	SiteName    string
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

func renderProductEmail(data emailTemplateData) string {
	siteName := normalizeEmailSiteName(data.SiteName)
	title := strings.TrimSpace(data.Title)
	if title == "" {
		title = siteName
	}

	var b strings.Builder
	b.WriteString(`<!doctype html>
<html>
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>`)
	b.WriteString(emailEscape(title))
	b.WriteString(`</title>
</head>
<body style="margin:0;padding:0;background:#f7f7f5;color:#202020;font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,Helvetica,Arial,sans-serif;">
  <div style="padding:32px 16px;">
    <div style="max-width:720px;margin:0 auto;background:#ffffff;border:1px solid #dedede;border-radius:24px;padding:56px 64px;box-sizing:border-box;">
      <div style="width:64px;height:64px;border:1px solid #ececec;border-radius:16px;box-shadow:0 8px 22px rgba(0,0,0,.08);display:flex;align-items:center;justify-content:center;margin-bottom:42px;">`)
	b.WriteString(emailLogoHTML())
	b.WriteString(`</div>
      <h1 style="margin:0 0 28px;font-size:32px;line-height:1.2;font-weight:700;letter-spacing:-.03em;color:#202020;">`)
	b.WriteString(emailEscape(title))
	b.WriteString(`</h1>`)

	if strings.TrimSpace(data.Intro) != "" {
		b.WriteString(`<p style="margin:0 0 30px;font-size:20px;line-height:1.55;color:#202020;">`)
		b.WriteString(emailEscape(data.Intro))
		b.WriteString(`</p>`)
	}
	if strings.TrimSpace(data.PrimaryHTML) != "" {
		b.WriteString(data.PrimaryHTML)
	}
	if strings.TrimSpace(data.SupportHTML) != "" {
		b.WriteString(`<div style="margin-top:30px;">`)
		b.WriteString(data.SupportHTML)
		b.WriteString(`</div>`)
	}

	b.WriteString(`<div style="height:1px;background:#dedede;margin:42px 0 32px;"></div>`)
	if strings.TrimSpace(data.FooterHTML) != "" {
		b.WriteString(data.FooterHTML)
	} else {
		b.WriteString(emailMutedParagraph("This is an automated message. You can safely ignore it if you did not request it."))
	}
	b.WriteString(`
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

func emailLogoHTML() string {
	return `<svg width="34" height="34" viewBox="0 0 64 64" role="img" aria-label="logo" xmlns="http://www.w3.org/2000/svg">
  <path d="M32 4 56 18v28L32 60 8 46V18L32 4Z" fill="#20201d"/>
  <path d="M32 4 56 18 32 32 8 18 32 4Z" fill="#2b2b27"/>
  <path d="M32 32 56 18v28L32 60V32Z" fill="#11110f"/>
  <path d="M32 32 8 18v28l24 14V32Z" fill="#34342f"/>
  <path d="M16 18h32L32 28 16 18Z" fill="#ffffff"/>
</svg>`
}

func emailCodeBlock(code string) string {
	return fmt.Sprintf(`<div style="margin:12px 0 18px;font-size:42px;line-height:1.15;font-weight:500;letter-spacing:.14em;color:#202020;font-family:'SFMono-Regular','Roboto Mono','Courier New',monospace;">%s</div>`, emailEscape(code))
}

func emailHeroValue(value string, color string) string {
	color = strings.TrimSpace(color)
	if color == "" {
		color = "#202020"
	}
	return fmt.Sprintf(`<div style="margin:12px 0 18px;font-size:42px;line-height:1.15;font-weight:650;letter-spacing:-.03em;color:%s;">%s</div>`, emailEscape(color), emailEscape(value))
}

func emailActionButton(label, href string) string {
	return fmt.Sprintf(`<a href="%s" style="display:inline-block;background:#090b12;color:#ffffff;text-decoration:none;border-radius:12px;padding:14px 28px;font-size:16px;line-height:1.4;font-weight:700;">%s</a>`, emailEscape(href), emailEscape(label))
}

func emailFallbackLink(label, href string) string {
	return fmt.Sprintf(`<p style="margin:18px 0 0;font-size:13px;line-height:1.7;color:#6b7280;">%s<br><span style="word-break:break-all;color:#374151;">%s</span></p>`, emailEscape(label), emailEscape(href))
}

func emailMutedParagraph(text string) string {
	return fmt.Sprintf(`<p style="margin:0 0 20px;font-size:16px;line-height:1.65;color:#666666;">%s</p>`, emailEscape(text))
}

func emailDetailsBlock(title string, rows []emailDetailRow) string {
	if len(rows) == 0 {
		return ""
	}
	var b strings.Builder
	if strings.TrimSpace(title) != "" {
		b.WriteString(`<h2 style="margin:0 0 16px;font-size:18px;line-height:1.4;color:#202020;">`)
		b.WriteString(emailEscape(title))
		b.WriteString(`</h2>`)
	}
	b.WriteString(`<table role="presentation" cellpadding="0" cellspacing="0" style="width:100%;border-collapse:collapse;margin:0;">`)
	for _, row := range rows {
		b.WriteString(`<tr><td style="padding:13px 0;border-bottom:1px solid #eeeeee;color:#666666;font-size:15px;line-height:1.5;">`)
		b.WriteString(emailEscape(row.Label))
		b.WriteString(`</td><td style="padding:13px 0;border-bottom:1px solid #eeeeee;color:#202020;font-size:15px;line-height:1.5;font-weight:650;text-align:right;">`)
		b.WriteString(emailEscape(row.Value))
		b.WriteString(`</td></tr>`)
	}
	b.WriteString(`</table>`)
	return b.String()
}

func BuildTestEmailBody(siteName string) string {
	siteName = normalizeEmailSiteName(siteName)
	return renderProductEmail(emailTemplateData{
		SiteName:    siteName,
		Title:       fmt.Sprintf("Test email from %s", siteName),
		Intro:       "Your SMTP configuration is working correctly.",
		PrimaryHTML: emailHeroValue("Sent successfully", "#16a34a"),
		FooterHTML:  emailMutedParagraph("This is an automated test message. No action is required."),
	})
}
