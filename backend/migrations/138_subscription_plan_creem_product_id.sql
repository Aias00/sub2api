ALTER TABLE subscription_plans
ADD COLUMN IF NOT EXISTS creem_product_id varchar(128) NOT NULL DEFAULT '';
