# 邀请中心 / 邀请管理模块设计

日期：2026-05-26

## 背景

当前仓库已经具备邀请码注册和邀请返利能力，但入口、规则配置、用户视图和管理员视图分散在多个页面中：

- 用户侧已有 `/affiliate` 页面，但只覆盖概览、邀请码分享、邀请用户列表和返利转余额，不包含返利流水和转余额流水。
- 管理员侧已有 `/admin/affiliates/invites`、`/admin/affiliates/rebates`、`/admin/affiliates/transfers` 三个记录页，但缺少模块首页、邀请码管理页和独立规则配置页。
- 邀请码注册开关与返利规则仍埋在系统设置页中，不属于独立模块。

现状导致两个问题：

1. 用户无法把“邀请码、邀请关系、返利产生、返利提取”当成一个完整功能理解。
2. 管理员需要在“系统设置”和“返利记录页”之间来回切换，缺少完整的邀请管理工作台。

## 目标

本轮目标是把现有邀请码注册与邀请返利能力重组为独立模块：

- 用户侧重构为单独的“邀请中心”。
- 管理员侧重构为单独的“邀请管理”模块。
- 复用现有邀请码绑定、返利计算、冻结释放、转余额逻辑。
- 只做模块边界、入口、信息架构、查询接口收拢，不改返利算法。

## 非目标

本轮不做以下内容：

- 不调整返利计算公式。
- 不引入按商品/套餐差异返利、首充返利、阶梯返利等新商业规则。
- 不迁移现有邀请返利数据结构。
- 不重做注册流程的邀请码校验逻辑。
- 不删除旧路由，仅做兼容性重定向或复用。

## 当前系统事实

### 用户侧

- `/affiliate` 已存在，当前页面在 [frontend/src/views/user/AffiliateView.vue](/Users/aias/Work/github/cloudbase/frontend/src/views/user/AffiliateView.vue)。
- 当前页面包含：
  - 实际返利比例
  - 邀请人数
  - 可转返利额度
  - 历史返利额度
  - 我的邀请码
  - 邀请链接
  - 邀请用户列表
  - 一键转余额
- 当前用户 API 仅有：
  - `getAffiliateDetail`
  - `transferAffiliateQuota`
  - 见 [frontend/src/api/user.ts](/Users/aias/Work/github/cloudbase/frontend/src/api/user.ts)

### 管理员侧

- 管理员已有邀请返利模块命名空间 `/admin/affiliates/*`，见 [frontend/src/router/index.ts](/Users/aias/Work/github/cloudbase/frontend/src/router/index.ts)。
- 当前仅有三张记录页：
  - `/admin/affiliates/invites`
  - `/admin/affiliates/rebates`
  - `/admin/affiliates/transfers`
- 当前侧边栏也只暴露三张记录页，见 [frontend/src/components/layout/AppSidebar.vue](/Users/aias/Work/github/cloudbase/frontend/src/components/layout/AppSidebar.vue)。

### 规则配置

- 邀请码注册开关仍位于系统设置页注册区域，见 [frontend/src/views/admin/SettingsView.vue](/Users/aias/Work/github/cloudbase/frontend/src/views/admin/SettingsView.vue)。
- 邀请返利总开关与返利规则仍位于系统设置页功能区，包含：
  - `affiliate_enabled`
  - `invitation_code_enabled`
  - `affiliate_rebate_rate`
  - `affiliate_rebate_freeze_hours`
  - `affiliate_rebate_duration_days`
  - `affiliate_rebate_per_invitee_cap`

### 后端能力

- 核心服务集中在 [backend/internal/service/affiliate_service.go](/Users/aias/Work/github/cloudbase/backend/internal/service/affiliate_service.go)。
- 已有能力包括：
  - 确保用户 affiliate profile
  - 通过邀请码查邀请人
  - 绑定 inviter
  - 返利累计
  - 冻结期处理
  - 返利转余额
  - 邀请用户列表
  - 管理员侧用户专属邀请码/专属返利比例管理
  - 管理员侧邀请/返利/转余额记录查询

## 目标信息架构

### 用户侧：邀请中心

保留现有路径 `/affiliate`，但升级为完整“邀请中心”，包含四个主区块：

1. 概览
   - 我的邀请码
   - 邀请链接
   - 邀请人数
   - 可转返利额度
   - 冻结返利额度
   - 历史返利总额
   - 当前实际返利比例
   - 操作：复制邀请码、复制邀请链接、转余额

2. 邀请记录
   - 我邀请的用户
   - 注册时间
   - 累计给我带来的返利

3. 返利记录
   - 哪一笔订单给我产生了多少返利
   - 包括返利时间、被邀请用户、订单金额、实付金额、返利金额、返利状态

4. 转余额记录
   - 返利转入余额历史
   - 包括转入时间、转入金额、余额快照、剩余返利快照

### 管理员侧：邀请管理

保留 `/admin/affiliates/*` 命名空间，补齐成完整模块：

- `/admin/affiliates/overview`
- `/admin/affiliates/rules`
- `/admin/affiliates/codes`
- `/admin/affiliates/invites`
- `/admin/affiliates/rebates`
- `/admin/affiliates/transfers`

各页职责：

1. `overview`
   - 模块首页
   - 展示启用状态、当前规则和经营指标

2. `rules`
   - 邀请码注册开关
   - 邀请返利开关
   - 全局返利比例
   - 冻结期
   - 有效期
   - 单人返利上限

3. `codes`
   - 用户邀请码管理
   - 专属邀请码
   - 专属返利比例
   - 操作：修改邀请码、重置邀请码、设置专属返利比例

4. `invites` / `rebates` / `transfers`
   - 继续复用现有记录页能力

## API 设计

### 用户侧新增接口

在不改返利逻辑的前提下，补齐用户中心所需查询接口：

1. `GET /api/v1/affiliate/summary`
   - 返回邀请码、邀请链接、返利概览和邀请列表。
   - 现有 `getAffiliateDetail` 可以演进为该接口的后端实现。

2. `GET /api/v1/affiliate/rebates`
   - 返回当前登录用户作为 inviter 的返利流水。

3. `GET /api/v1/affiliate/transfers`
   - 返回当前登录用户返利转余额流水。

说明：

- 本轮不强制拆出 `GET /api/v1/affiliate/invites`。
- 当前 `invitees` 可以先继续作为 `summary` 的一部分返回。
- 如果后续邀请人数过大，再单独做分页接口。

### 管理员侧新增接口

1. `GET /api/v1/admin/affiliates/overview`
   - 返回模块首页所需聚合指标：
     - `affiliate_enabled`
     - `invitation_code_enabled`
     - `affiliate_rebate_rate`
     - `affiliate_rebate_freeze_hours`
     - `affiliate_rebate_duration_days`
     - `affiliate_rebate_per_invitee_cap`
     - invited/rebated/available/frozen/history 等统计

2. `GET /api/v1/admin/affiliates/rules`
   - 单独读取邀请规则。

3. `PUT /api/v1/admin/affiliates/rules`
   - 单独更新邀请规则。

说明：

- 规则接口从 `PUT /admin/settings` 中拆出，是为了降低设置项部分更新误伤其他字段的风险。
- 现有 `/admin/affiliates/invites|rebates|transfers` 接口保持不变，优先复用。
- 现有用户专属邀请码/专属返利比例接口继续复用，并集中展示在 `codes` 页。

## 路由与导航调整

### 用户侧

- 继续保留侧边栏 `邀请返利` 入口。
- 入口指向 `/affiliate`。
- 页面标题和文案调整为“邀请中心”。

### 管理员侧

现有：

- `/admin/affiliates` 仅作为 redirect
- 侧边栏仅展示 3 张记录页

调整为：

- `/admin/affiliates` 默认跳转到 `/admin/affiliates/overview`
- 侧边栏子项扩展为：
  - 概览
  - 规则配置
  - 邀请码管理
  - 邀请记录
  - 返利记录
  - 转余额记录

### 系统设置页迁移策略

本轮不立即删除设置页中的旧入口。

策略：

- 设置页里的邀请模块配置区缩减为“说明 + 跳转到邀请管理”
- 避免管理员在过渡期完全找不到原功能
- 下一轮再考虑移除旧区块

## 数据与兼容性策略

- 不修改现有 `user_affiliates`、`user_affiliate_ledger` 等表结构。
- 不迁移现有邀请码。
- 不改变现有邀请关系绑定。
- 不改变返利产生、冻结释放、转余额的账务语义。
- 保留现有旧路径，避免外链和书签失效。

## 风险

1. 用户侧新增返利/转余额流水接口时，需要确保只能读取当前登录用户自己的数据。
2. 规则从系统设置中拆分后，需要保证前后端字段口径一致，避免出现“双写两套配置源”。
3. 管理员侧邀请码管理页如果直接复用“专属用户列表”，需要避免语义混乱：
   - 不是只有“自定义过”的用户才应可见
   - 应支持搜索任意用户并进行邀请码/返利配置
4. 设置页过渡期内会出现“老入口 + 新模块入口”并存，需要明确文案，避免管理员误以为是两套功能。

## 实施里程碑

### Milestone 1：后端补接口

- 新增用户侧：
  - `summary`
  - `rebates`
  - `transfers`
- 新增管理员侧：
  - `overview`
  - `rules`

### Milestone 2：管理员模块成型

- 新增：
  - `/admin/affiliates/overview`
  - `/admin/affiliates/rules`
  - `/admin/affiliates/codes`
- 统一管理员邀请管理导航
- 设置页改成跳转式入口

### Milestone 3：用户邀请中心重构

- 重构 `/affiliate`
- 增加返利流水和转余额流水展示
- 保留现有复制邀请码、复制链接、转余额动作

### Milestone 4：回归与收尾

- 文案、标题、导航统一
- 回归注册邀请码校验
- 回归邀请返利产生
- 回归返利转余额
- 回归管理员专属邀请码和专属返利比例配置

## 验收标准

### 用户侧

- 用户可以在一个页面内完成：
  - 查看邀请码
  - 复制邀请链接
  - 查看邀请记录
  - 查看返利流水
  - 查看转余额流水
  - 执行返利转余额

### 管理员侧

- 管理员可以在一个独立模块内完成：
  - 配置邀请规则
  - 管理邀请码
  - 查看邀请记录
  - 查看返利记录
  - 查看转余额记录

### 兼容性

- 已有邀请关系不丢失
- 已有返利规则不改变
- 已有记录查询仍可用
- 旧链接不报错

## 验证策略

### 后端

- 用户 `summary/rebates/transfers` 查询测试
- 管理员 `overview/rules` 查询与更新测试
- 当前用户隔离测试
- 邀请关系、返利、转余额回归测试

### 前端

- 用户邀请中心空态/有数据态
- 管理员模块导航与子页切换
- 规则更新提交流程
- 邀请码管理动作

### 集成

- 邀请链接进入注册页
- 注册绑定 inviter
- 充值后返利产生
- 返利冻结/释放
- 转余额后用户余额和流水一致
