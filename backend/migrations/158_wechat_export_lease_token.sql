-- Migration 158: 添加worker_lease_token字段用于Worker信任边界验证
-- 基于2026-06架构Review的安全改进方案（P0）

-- 添加lease_token字段用于worker身份验证
-- Token在ClaimNextTask时生成，Complete/Fail/AddTaskLog时验证
ALTER TABLE wechat_export_tasks
    ADD COLUMN IF NOT EXISTS worker_lease_token TEXT NOT NULL DEFAULT '';

-- 添加worker_run_id字段用于唯一标识一次运行（可选）
-- 用于审计追踪，区分同一任务的多次运行
ALTER TABLE wechat_export_tasks
    ADD COLUMN IF NOT EXISTS worker_run_id TEXT NOT NULL DEFAULT '';

-- 索引优化查询性能
-- 用于CompleteTask/FailTask/AddTaskLog的token验证查询
CREATE INDEX IF NOT EXISTS idx_wechat_export_tasks_worker_lease_token
    ON wechat_export_tasks(worker_lease_token);

-- 注意事项：
-- 1. worker_lease_token在Claim时生成（64字符hex，256-bit）
-- 2. Complete/Fail时验证token匹配 + lease未过期
-- 3. Cancel时清空token（防止worker继续操作）
-- 4. Token使用json:"-"完全隐藏，只在Claim响应中单独返回
-- 5. AddTaskLog强制验证token（防止伪造日志）