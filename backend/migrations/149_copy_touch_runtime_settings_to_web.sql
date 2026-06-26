WITH runtime_setting_key_map(new_key, old_key) AS (
    VALUES
        ('web_app_url', 'touch_app_url'),
        ('web_app_name', 'touch_app_name'),
        ('web_app_description', 'touch_app_description'),
        ('web_app_logo', 'touch_app_logo'),
        ('web_app_favicon', 'touch_app_favicon'),
        ('web_app_preview_image', 'touch_app_preview_image'),
        ('web_theme', 'touch_theme'),
        ('web_appearance', 'touch_appearance'),
        ('web_default_locale', 'touch_default_locale'),
        ('prompt_cases_title', 'touch_prompt_cases_title'),
        ('prompt_cases_description', 'touch_prompt_cases_description'),
        ('prompt_templates_title', 'touch_prompt_templates_title'),
        ('prompt_templates_description', 'touch_prompt_templates_description'),
        ('workspace_shell_config', 'touch_workspace_shell_config'),
        ('pricing_title', 'touch_pricing_title'),
        ('pricing_description', 'touch_pricing_description'),
        ('pricing_shell_config', 'touch_pricing_shell_config'),
        ('credits_title', 'touch_credits_title'),
        ('credits_description', 'touch_credits_description'),
        ('credits_purchase_label', 'touch_credits_purchase_label'),
        ('credits_balance_label', 'touch_credits_balance_label'),
        ('web_locale_detect_enabled', 'touch_locale_detect_enabled'),
        ('web_email_auth_visible', 'touch_email_auth_visible'),
        ('web_google_auth_visible', 'touch_google_auth_visible'),
        ('web_github_auth_visible', 'touch_github_auth_visible'),
        ('web_google_analytics_id', 'touch_google_analytics_id'),
        ('web_clarity_id', 'touch_clarity_id'),
        ('web_plausible_domain', 'touch_plausible_domain'),
        ('web_plausible_src', 'touch_plausible_src'),
        ('web_openpanel_client_id', 'touch_openpanel_client_id'),
        ('web_vercel_analytics_enabled', 'touch_vercel_analytics_enabled'),
        ('web_adsense_code', 'touch_adsense_code'),
        ('web_affonso_enabled', 'touch_affonso_enabled'),
        ('web_affonso_id', 'touch_affonso_id'),
        ('web_affonso_cookie_duration', 'touch_affonso_cookie_duration'),
        ('web_promotekit_enabled', 'touch_promotekit_enabled'),
        ('web_promotekit_id', 'touch_promotekit_id'),
        ('web_crisp_enabled', 'touch_crisp_enabled'),
        ('web_crisp_website_id', 'touch_crisp_website_id'),
        ('web_tawk_enabled', 'touch_tawk_enabled'),
        ('web_tawk_property_id', 'touch_tawk_property_id'),
        ('web_tawk_widget_id', 'touch_tawk_widget_id')
)
INSERT INTO settings (key, value, updated_at)
SELECT m.new_key, s.value, NOW()
FROM runtime_setting_key_map m
JOIN settings s ON s.key = m.old_key
WHERE s.value <> ''
ON CONFLICT (key) DO UPDATE
SET value = EXCLUDED.value,
    updated_at = NOW()
WHERE settings.value = ''
  AND EXCLUDED.value <> '';
