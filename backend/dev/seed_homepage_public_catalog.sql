BEGIN;

INSERT INTO settings (key, value)
VALUES
  ('payment_enabled', 'true'),
  ('BALANCE_RECHARGE_MULTIPLIER', '1.2'),
  (
    'PAYMENT_RECHARGE_PRODUCTS',
    $json$
    [
      {
        "id": "starter",
        "name": "体验包",
        "description": "适合个人开发者快速上手",
        "amount": 29,
        "badge": "入门",
        "recommended": true,
        "features": ["适合初次体验", "按量计费余额", "可直接用于 API 调用"],
        "sort_order": 10
      },
      {
        "id": "growth",
        "name": "开发包",
        "description": "适合高频编码与日常迭代",
        "amount": 99,
        "badge": "热门",
        "recommended": true,
        "features": ["更高可用余额", "适合日常编码", "预算控制更轻松"],
        "sort_order": 20
      },
      {
        "id": "studio",
        "name": "团队包",
        "description": "适合小团队共享预算与联调",
        "amount": 299,
        "badge": "团队",
        "recommended": false,
        "features": ["适合团队联调", "统一充值入口", "适合阶段性冲刺"],
        "sort_order": 30
      }
    ]
    $json$
  ),
  (
    'model_plaza_items',
    $json$
    [
      {
        "id": "claude-haiku-4-5-20251001",
        "provider": "anthropic",
        "title": "claude-haiku-4-5-20251001",
        "badge": "Haiku",
        "description": "适合低成本高频调用、简单补全与轻量总结。",
        "capability_tags": ["轻量补全", "快速响应", "成本敏感"],
        "model_ids": ["claude-haiku-4-5-20251001"],
        "input_price": "¥0.4000 / 1M Tokens",
        "output_price": "¥2.0000 / 1M Tokens",
        "cache_read_price": "¥0.0400 / 1M Tokens",
        "cache_write_price": "¥0.5000 / 1M Tokens",
        "billing_badge": "按量计费",
        "visible": true,
        "sort_order": 10
      },
      {
        "id": "claude-opus-4-6",
        "provider": "anthropic",
        "title": "claude-opus-4-6",
        "badge": "Opus",
        "description": "适合复杂推理、深度重构和关键代码审查。",
        "capability_tags": ["复杂推理", "深度重构", "代码审查"],
        "model_ids": ["claude-opus-4-6"],
        "input_price": "¥2.0000 / 1M Tokens",
        "output_price": "¥10.0000 / 1M Tokens",
        "cache_read_price": "¥0.2000 / 1M Tokens",
        "cache_write_price": "¥2.5000 / 1M Tokens",
        "billing_badge": "按量计费",
        "visible": true,
        "sort_order": 20
      },
      {
        "id": "claude-opus-4-7",
        "provider": "anthropic",
        "title": "claude-opus-4-7",
        "badge": "Opus",
        "description": "适合长链路方案推演、复杂实现与高难度推理。",
        "capability_tags": ["方案推演", "高难实现", "复杂推理"],
        "model_ids": ["claude-opus-4-7"],
        "input_price": "¥2.0000 / 1M Tokens",
        "output_price": "¥10.0000 / 1M Tokens",
        "cache_read_price": "¥0.2000 / 1M Tokens",
        "cache_write_price": "¥2.5000 / 1M Tokens",
        "billing_badge": "按量计费",
        "visible": true,
        "sort_order": 30
      },
      {
        "id": "claude-sonnet-4-6",
        "provider": "anthropic",
        "title": "claude-sonnet-4-6",
        "badge": "Sonnet",
        "description": "适合日常编码、功能开发与 Agent 工作流。",
        "capability_tags": ["代码生成", "功能开发", "Agent 调用"],
        "model_ids": ["claude-sonnet-4-6"],
        "input_price": "¥1.2000 / 1M Tokens",
        "output_price": "¥6.0000 / 1M Tokens",
        "cache_read_price": "¥0.1200 / 1M Tokens",
        "cache_write_price": "¥1.5000 / 1M Tokens",
        "billing_badge": "按量计费",
        "visible": true,
        "sort_order": 40
      }
    ]
    $json$
  )
ON CONFLICT (key) DO UPDATE
SET value = EXCLUDED.value,
    updated_at = NOW();

INSERT INTO groups (
  name,
  description,
  platform,
  subscription_type,
  rate_multiplier,
  status,
  default_validity_days,
  supported_model_scopes,
  sort_order
)
VALUES
  (
    'Claude',
    '复杂推理、系统设计与代码审查',
    'anthropic',
    'standard',
    1.0000,
    'active',
    30,
    '["Claude Opus 4.6", "Claude Sonnet 4.5", "Claude 3.7 Sonnet"]'::jsonb,
    10
  ),
  (
    'GPT',
    '代码生成、功能开发与高频实现',
    'openai',
    'standard',
    1.0000,
    'active',
    30,
    '["GPT-5.4", "GPT-5.3 Codex", "o3"]'::jsonb,
    20
  ),
  (
    'Gemini',
    '视觉理解、多模态与截图分析',
    'gemini',
    'standard',
    1.0000,
    'active',
    30,
    '["Gemini 2.5 Pro", "Gemini 2.5 Flash", "Gemini 2.5 Flash Image"]'::jsonb,
    30
  )
ON CONFLICT (name) WHERE deleted_at IS NULL DO UPDATE
SET description = EXCLUDED.description,
    platform = EXCLUDED.platform,
    subscription_type = EXCLUDED.subscription_type,
    rate_multiplier = EXCLUDED.rate_multiplier,
    status = EXCLUDED.status,
    default_validity_days = EXCLUDED.default_validity_days,
    supported_model_scopes = EXCLUDED.supported_model_scopes,
    sort_order = EXCLUDED.sort_order,
    updated_at = NOW();

DELETE FROM subscription_plans
WHERE name IN (
  'Claude Pro 月付',
  'GPT Builder 月付',
  'Gemini Vision 月付',
  'GPT Team 月付'
);

INSERT INTO subscription_plans (
  group_id,
  name,
  description,
  price,
  original_price,
  validity_days,
  validity_unit,
  features,
  product_name,
  for_sale,
  sort_order,
  creem_product_id
)
VALUES
  (
    (SELECT id FROM groups WHERE name = 'Claude' AND deleted_at IS NULL LIMIT 1),
    'Claude Pro 月付',
    '面向复杂推理、重构与代码审查的主力套餐',
    69,
    89,
    30,
    'day',
    E'Claude Opus / Sonnet 可用\n适合高复杂度编码\n推荐给重度开发者',
    'Claude Pro Monthly',
    TRUE,
    10,
    ''
  ),
  (
    (SELECT id FROM groups WHERE name = 'GPT' AND deleted_at IS NULL LIMIT 1),
    'GPT Builder 月付',
    '适合日常功能开发、Agent 调用与高频实现',
    49,
    69,
    30,
    'day',
    E'适合高频编码\nGPT 系列优先接入\n更轻量的开发成本',
    'GPT Builder Monthly',
    TRUE,
    20,
    ''
  ),
  (
    (SELECT id FROM groups WHERE name = 'Gemini' AND deleted_at IS NULL LIMIT 1),
    'Gemini Vision 月付',
    '适合截图分析、视觉理解和多模态辅助开发',
    39,
    59,
    30,
    'day',
    E'多模态能力优先\n适合截图与视觉任务\n适合设计协作场景',
    'Gemini Vision Monthly',
    TRUE,
    30,
    ''
  ),
  (
    (SELECT id FROM groups WHERE name = 'GPT' AND deleted_at IS NULL LIMIT 1),
    'GPT Team 月付',
    '适合双人到五人的小团队统一接入与预算控制',
    129,
    169,
    30,
    'day',
    E'适合团队共享入口\n更适合阶段性协作\n公开价格更易决策',
    'GPT Team Monthly',
    TRUE,
    40,
    ''
  );

COMMIT;
