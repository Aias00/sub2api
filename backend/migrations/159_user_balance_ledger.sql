-- 用户统一余额流水表，记录所有余额变动（充值、扣费、兑换、调整等）
-- 参考 user_affiliate_ledger 的设计模式：纯 SQL migration，DECIMAL(20,8) 精度

CREATE TABLE IF NOT EXISTS user_balance_ledger (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    -- 流水类型（类型化枚举）
    entry_type VARCHAR(30) NOT NULL,

    -- 金额变动（正数入账，负数扣费）
    amount DECIMAL(20,8) NOT NULL,

    -- 余额快照（可选，历史数据为 NULL）
    balance_before DECIMAL(20,8) NULL,
    balance_after DECIMAL(20,8) NULL,

    -- 来源追溯
    source_type VARCHAR(30) NOT NULL,
    source_id BIGINT NULL,

    -- 描述和元数据
    description TEXT NULL,
    metadata_json JSONB NULL,

    -- 时间戳
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT user_balance_ledger_entry_type_check
        CHECK (entry_type IN (
            'recharge',           -- 充值订单
            'api_usage',          -- API调用扣费
            'image_workspace',    -- 图片工作台扣费（预留）
            'wechat_export',      -- 微信导出扣费（预留）
            'redeem',             -- 兑换码兑换
            'admin_adjustment',   -- 管理员调整
            'affiliate_transfer', -- 返利转余额（预留，实际不写入）
            'refund',             -- 退款扣减
            'promo_bonus',        -- 优惠码奖励
            'oauth_bind_bonus',   -- OAuth首次绑定奖励
            'expiry',             -- 过期清零（预留）
            'correction'          -- 系统纠正
        )),
    CONSTRAINT user_balance_ledger_source_type_check
        CHECK (source_type IN (
            'payment_order',
            'usage_log',
            'redeem_code',
            'admin_action',
            'affiliate_ledger',
            'refund',
            'promo_code_usage',
            'oauth_binding',
            'image_workspace_record',
            'wechat_export_task',
            'system_correction'
        ))
);

COMMENT ON TABLE user_balance_ledger IS '用户余额统一流水记录，回答"我的余额为什么从 A 变成 B"';
COMMENT ON COLUMN user_balance_ledger.entry_type IS '流水类型：recharge/api_usage/redeem/admin_adjustment/refund/promo_bonus/oauth_bind_bonus/expiry/correction';
COMMENT ON COLUMN user_balance_ledger.amount IS '金额变动：正数入账，负数扣费';
COMMENT ON COLUMN user_balance_ledger.balance_before IS '变动前余额快照（历史数据为 NULL）';
COMMENT ON COLUMN user_balance_ledger.balance_after IS '变动后余额快照';
COMMENT ON COLUMN user_balance_ledger.source_type IS '来源类型：payment_order/usage_log/redeem_code/admin_action/refund/promo_code_usage/oauth_binding/system_correction';
COMMENT ON COLUMN user_balance_ledger.source_id IS '来源记录 ID（如 payment_order.id、usage_log.id、redeem_code.id）';
COMMENT ON COLUMN user_balance_ledger.description IS '流水描述文案';
COMMENT ON COLUMN user_balance_ledger.metadata_json IS '扩展元数据（JSON格式，如订单号、模型名、操作人等）';

-- 用户查询优化：按用户ID + 时间倒序
CREATE INDEX idx_user_balance_ledger_user_time
    ON user_balance_ledger(user_id, created_at DESC);

-- 来源追溯优化：部分索引（仅索引有 source_id 的记录）
CREATE INDEX idx_user_balance_ledger_source
    ON user_balance_ledger(source_type, source_id)
    WHERE source_id IS NOT NULL;

-- 非API扣费流水优化：充值、兑换、调整等低频操作
CREATE INDEX idx_user_balance_ledger_non_usage
    ON user_balance_ledger(user_id, created_at DESC)
    WHERE entry_type != 'api_usage';

-- 尽力回填充值订单（仅 status=completed 且 order_type=balance 的订单）
-- balance_before/balance_after 无法可靠回填，置 NULL
INSERT INTO user_balance_ledger (
    user_id, entry_type, amount,
    balance_before, balance_after,
    source_type, source_id,
    description, created_at
)
SELECT
    po.user_id,
    'recharge',
    po.amount,
    NULL,
    NULL,
    'payment_order',
    po.id,
    format('充值订单 %s (%s)', po.out_trade_no, po.payment_type),
    po.created_at
FROM payment_orders po
WHERE po.status = 'completed'
  AND po.order_type = 'balance'
  AND NOT EXISTS (
      SELECT 1 FROM user_balance_ledger
      WHERE source_type = 'payment_order' AND source_id = po.id
  );

-- 尽力回填兑换码（仅 type=balance 且 status=used 的兑换码）
INSERT INTO user_balance_ledger (
    user_id, entry_type, amount,
    balance_before, balance_after,
    source_type, source_id,
    description, created_at
)
SELECT
    rc.used_by,
    CASE
        WHEN rc.type = 'admin_balance' THEN 'admin_adjustment'
        ELSE 'redeem'
    END,
    rc.value,
    NULL,
    NULL,
    'redeem_code',
    rc.id,
    COALESCE(rc.notes, format('兑换码 %s', substring(rc.code FROM 1 FOR 8))),
    COALESCE(rc.used_at, rc.created_at)
FROM redeem_codes rc
WHERE rc.status = 'used'
  AND (rc.type = 'balance' OR rc.type = 'admin_balance')
  AND rc.used_by IS NOT NULL
  AND NOT EXISTS (
      SELECT 1 FROM user_balance_ledger
      WHERE source_type = 'redeem_code' AND source_id = rc.id
  );

-- 不回填 usage_logs（API调用），原因：
-- 1. 数量庞大，回填耗时可能很长
-- 2. 无 balance_before/balance_after 快照
-- 3. 用户已有 /usage 页面查看 API 调用详情