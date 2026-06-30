<template>
  <div class="home-business-page public-dark-page min-h-screen bg-[#101114] text-white">
    <PublicDarkHeader :account-label="isAuthenticated ? t('nav.dashboard') : t('common.login')" />

    <main class="px-6 py-10 sm:py-14">
      <div class="mx-auto max-w-6xl">
        <section class="mb-10 grid gap-6 lg:grid-cols-[minmax(0,1fr)_360px] lg:items-start">
          <div>
            <p class="text-sm font-semibold uppercase tracking-[0.22em] text-cyan-200/70">
              {{ t('wechatExport.pageTitle') }}
            </p>
            <h1 class="mt-4 max-w-4xl text-4xl font-black leading-tight text-white sm:text-5xl">
              {{ t('wechatExport.pageTitle') }}
            </h1>
            <p class="mt-4 max-w-3xl text-base leading-8 text-white/60">
              {{ t('wechatExport.pageHint') }}
            </p>
          </div>

          <aside class="rounded-2xl border border-cyan-300/20 bg-cyan-300/[0.055] p-5 shadow-[0_20px_60px_rgba(0,0,0,0.22)]">
            <div class="flex items-center justify-between gap-3">
              <div>
                <p class="text-xs font-black uppercase tracking-[0.18em] text-cyan-100/55">WeChat Session</p>
                <div class="mt-2 flex items-center gap-2">
                  <span class="text-sm font-bold text-white/80">扫码登录</span>
                  <span
                    class="rounded-full border border-white/10 bg-white/[0.045] px-2.5 py-1 text-xs font-semibold"
                    :class="session?.status === 'ready' ? 'text-emerald-100' : 'text-white/50'"
                  >
                    {{ sessionStatusLabel }}
                  </span>
                </div>
              </div>
            </div>

            <div v-if="sessionWarning" class="wechat-export-warning mt-3 rounded-xl border px-3 py-2 text-xs leading-5">
              {{ sessionWarning }}
            </div>

            <div class="mt-4 flex flex-wrap gap-2">
              <button
                v-if="session?.status !== 'ready'"
                type="button"
                class="rounded-xl bg-cyan-300 px-4 py-2 text-sm font-bold text-slate-950 transition hover:bg-cyan-200 disabled:opacity-45"
                :disabled="!isAuthenticated || sessionLoading"
                @click="handleCreateSession"
              >
                {{ sessionLoading ? '创建中' : '创建二维码' }}
              </button>
              <button
                v-if="session?.status === 'ready'"
                type="button"
                class="rounded-xl border border-emerald-200/20 px-3 py-1.5 text-sm font-semibold text-emerald-100 transition hover:bg-emerald-300/10 disabled:opacity-45"
                :disabled="sessionValidating"
                @click="handleValidateSession"
              >
                {{ sessionValidating ? '检查中' : '校验' }}
              </button>
              <button
                v-if="session?.id"
                type="button"
                class="rounded-xl border border-white/10 px-3 py-1.5 text-sm font-semibold text-white/70 transition hover:bg-white/[0.06]"
                :disabled="!isAuthenticated"
                @click="handleLogoutSession"
              >
                退出
              </button>
            </div>

            <div v-if="qrcodeUrl && session?.status !== 'ready'" class="mt-4 flex items-center gap-3 rounded-2xl border border-white/10 bg-black/20 p-3">
              <img v-if="isQRCodeImage" :src="qrcodeUrl" alt="二维码" class="h-24 w-24 rounded-xl bg-white p-1" />
              <div class="min-w-0">
                <p class="text-sm font-bold text-white/80">微信扫码确认</p>
                <p class="mt-1 text-xs leading-5 text-white/45">登录后可同步公众号文章并创建导出任务。</p>
              </div>
            </div>
          </aside>
        </section>

        <div
          v-if="!isAuthenticated"
          class="wechat-export-warning mb-6 rounded-2xl border p-4 text-sm leading-6"
        >
          {{ t('wechatExport.loginRequired') }}
        </div>
        <div
          v-else-if="requiresReadySession"
          class="wechat-export-warning mb-6 rounded-2xl border p-4 text-sm leading-6"
        >
          微信会话未登录或已失效，请先扫码登录并等待状态变为“已就绪”，再使用公众号搜索、同步、导入、导出和任务操作。
        </div>

        <!-- 公众号管理 -->
        <section class="mb-6 rounded-2xl border border-white/10 bg-white/[0.035] p-5">
          <div class="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
            <div>
              <div class="flex items-center gap-2">
                <span class="text-xl font-black text-white">公众号管理</span>
                <span class="rounded-full border border-white/10 bg-white/[0.045] px-2.5 py-1 text-xs font-semibold text-white/50">
                  {{ accounts.length }} 个已绑定
                </span>
              </div>
              <p class="mt-2 text-sm leading-6 text-white/45">搜索并绑定公众号，或对已绑定公众号执行文章同步。</p>
            </div>
            <div class="flex w-full flex-wrap gap-2 lg:w-[520px]">
              <input
                v-model="accountSearchQuery"
                type="text"
                placeholder="搜索公众号"
                class="min-w-0 flex-1 rounded-xl border border-white/10 bg-[#101114] px-3 py-2 text-sm text-white outline-none placeholder:text-white/30 focus:border-cyan-300/40 focus:bg-[#111318]"
                :disabled="wechatActionDisabled || accountSearching"
              />
              <button
                type="button"
                class="rounded-xl border border-cyan-300/30 px-4 py-2 text-sm font-bold text-cyan-100 hover:bg-cyan-300/10 disabled:opacity-45"
                :disabled="wechatActionDisabled || accountSearching || !accountSearchQuery.trim()"
                @click="handleSearchRemoteAccounts"
              >
                {{ accountSearching ? '搜索' : '查找' }}
              </button>
            </div>
          </div>
          <p v-if="accountSearchMessage" class="mt-4 rounded-xl border border-cyan-300/15 bg-cyan-300/[0.06] px-3 py-2 text-xs text-cyan-50">
            {{ accountSearchMessage }}
          </p>

          <!-- 搜索结果（如果有） -->
          <div v-if="accountSearchResults.length > 0" class="mt-3 border-t border-white/10 pt-3">
            <div class="text-sm font-bold text-white/75 mb-2">搜索结果</div>
            <div class="grid gap-2 sm:grid-cols-2">
              <div v-for="account in accountSearchResults" :key="`remote-${account.fakeid}`" class="flex items-center justify-between gap-2 rounded-xl border border-cyan-300/15 bg-cyan-300/[0.04] px-3 py-2">
                <div class="min-w-0">
                  <p class="truncate text-sm font-bold text-white">{{ account.nickname || account.fakeid }}</p>
                  <p class="truncate text-xs text-white/45">{{ account.fakeid }}</p>
                </div>
                <button
                  type="button"
                  class="rounded-xl bg-white px-3 py-1.5 text-sm font-bold text-slate-950 hover:bg-white/90"
                  :disabled="wechatActionDisabled || accountLoading"
                  @click="handleBindSearchResult(account)"
                >
                  绑定
                </button>
              </div>
            </div>
          </div>

          <!-- 已绑定公众号列表 -->
          <div v-if="accounts.length > 0" class="mt-3 border-t border-white/10 pt-3">
            <div class="text-sm font-bold text-white/75 mb-2">已绑定公众号</div>
            <div class="grid gap-2 sm:grid-cols-2">
              <div v-for="account in accounts" :key="account.id" class="flex items-center justify-between gap-2 rounded-xl border border-white/10 bg-[#101114] px-3 py-2">
                <div class="min-w-0">
                  <p class="truncate text-sm font-bold text-white">{{ account.nickname || account.fakeid }}</p>
                  <p class="truncate text-xs text-white/45">{{ account.fakeid }}</p>
                </div>
                <button type="button" class="rounded-xl border border-cyan-200/30 bg-cyan-200/10 px-2.5 py-1 text-sm font-semibold text-cyan-100 hover:bg-cyan-200/20 disabled:opacity-45" :disabled="wechatActionDisabled || syncingFakeid === account.fakeid" @click="handleSyncAccount(account.fakeid)">
                  {{ syncingFakeid === account.fakeid ? '同步中' : '同步' }}
                </button>
              </div>
            </div>
          </div>

          <!-- 同步进度和继续同步按钮 -->
          <div v-if="syncingProgress && syncingProgress.fakeid" class="mt-2 text-xs text-cyan-100/80">
            {{ syncProgressText(syncingProgress) }}
            <button
              v-if="syncingProgress.hasMore"
              type="button"
              class="ml-2 rounded bg-cyan-200/20 px-2 py-0.5 text-xs font-semibold text-cyan-100 hover:bg-cyan-200/30"
              :disabled="wechatActionDisabled"
              @click="handleContinueSync(syncingProgress.fakeid, syncingProgress.synced)"
            >
              继续同步
            </button>
          </div>
        </section>

        <!-- 主内容区：文章列表 -->
        <div class="mb-6">
          <section class="rounded-2xl border border-white/10 bg-white/[0.035] p-5">
            <div class="flex items-center justify-between mb-3">
              <h2 class="text-xl font-black text-white">文章列表</h2>
              <button
                type="button"
                class="rounded-xl border border-white/10 px-3 py-1.5 text-sm font-semibold text-white/70 hover:bg-white/[0.06] disabled:opacity-45"
                :disabled="!isAuthenticated || loading"
                @click="refreshAll"
              >
                刷新
              </button>
            </div>

            <!-- 导入链接 -->
            <div class="mb-3">
              <form class="flex gap-2" @submit.prevent="handleImport">
                <input
                  v-model="articleLink"
                  type="url"
                  placeholder="粘贴文章链接 mp.weixin.qq.com/s/..."
                  class="min-h-10 flex-1 rounded-xl border border-white/10 bg-[#101114] px-3 text-sm text-white outline-none placeholder:text-white/30 focus:border-cyan-300/40"
                  :disabled="wechatActionDisabled || importing"
                />
                <button
                  type="submit"
                  class="min-h-10 rounded-xl bg-cyan-300 px-4 text-sm font-bold text-slate-950 hover:bg-cyan-200 disabled:opacity-45"
                  :disabled="wechatActionDisabled || importing || !articleLink.trim()"
                >
                  {{ importing ? '导入' : '导入' }}
                </button>
              </form>
            </div>

            <!-- 筛选器 -->
            <div class="mb-3 grid gap-2 grid-cols-[1fr_1fr_1fr]">
              <input
                v-model="articleSearchQuery"
                type="search"
                placeholder="搜索文章"
                class="min-h-9 rounded-xl border border-white/10 bg-[#101114] px-3 text-sm text-white outline-none placeholder:text-white/30 focus:border-cyan-300/40"
              />
              <select
                v-model="articleAccountFilter"
                class="min-h-9 rounded-xl border border-white/10 bg-[#101114] px-3 text-sm text-white outline-none focus:border-cyan-300/40"
              >
                <option value="all">全部公众号</option>
                <option v-for="account in accounts" :key="account.fakeid" :value="account.fakeid">
                  {{ account.nickname || account.fakeid }}
                </option>
              </select>
              <select
                v-model="articleStatusFilter"
                class="min-h-9 rounded-xl border border-white/10 bg-[#101114] px-3 text-sm text-white outline-none focus:border-cyan-300/40"
              >
                <option value="all">全部状态</option>
                <option value="pending">待抓取</option>
                <option value="fetched">已抓取</option>
                <option value="normal">正常</option>
              </select>
            </div>

            <!-- 选择操作 -->
            <div class="mb-3 flex items-center gap-2">
              <button
                type="button"
                class="rounded-xl border border-cyan-300/30 bg-cyan-300/10 px-3 py-1.5 text-sm font-semibold text-cyan-100 hover:bg-cyan-300/20 disabled:opacity-45"
                :disabled="wechatActionDisabled"
                @click="selectFilteredArticles"
              >
                全选 ({{ filteredArticles.length }})
              </button>              <button
                type="button"
                class="rounded-xl border border-white/10 px-3 py-1.5 text-sm font-semibold text-white/65 hover:bg-white/[0.06] disabled:opacity-45"
                :disabled="wechatActionDisabled"
                @click="clearSelectedArticles"
              >
                清空
              </button>
              <span class="text-xs text-white/40">已选 {{ selectedArticleIds.length }} 篇</span>
            </div>

            <!-- 导出操作 -->
            <div class="mb-3 rounded-xl border border-white/10 bg-[#101114] px-3 py-3">
              <div class="flex flex-wrap items-center gap-2">
                <span class="mr-1 text-sm font-bold text-white/75">导出操作</span>
                <label
                  v-for="format in availableFormats"
                  :key="format"
                  class="inline-flex min-h-9 cursor-pointer items-center gap-2 rounded-xl border border-white/10 bg-white/[0.035] px-3 text-sm font-semibold uppercase text-white/75 hover:bg-white/[0.06]"
                >
                  <input v-model="formats" type="checkbox" class="h-4 w-4 accent-cyan-200" :value="format" :disabled="wechatActionDisabled" />
                  {{ format }}
                </label>
                <label class="inline-flex min-h-9 cursor-pointer items-center gap-2 rounded-xl border border-white/10 bg-white/[0.035] px-3 text-sm font-semibold text-white/75 hover:bg-white/[0.06]">
                  <input v-model="includeEngagement" type="checkbox" class="h-4 w-4 accent-cyan-200" :disabled="wechatActionDisabled" />
                  互动数据
                </label>
                <div v-if="isAuthenticated && estimatedCredits !== null" class="flex min-h-9 items-center gap-3 rounded-xl border border-white/10 bg-white/[0.035] px-3 text-xs text-white/55">
                  <span>预计 {{ estimatedCredits.toFixed(2) }} 余额</span>
                  <span :class="insufficientBalance ? 'text-red-200' : 'text-emerald-200'">余额 {{ userBalance.toFixed(2) }}</span>
                </div>
                <button
                  type="button"
                  class="min-h-9 rounded-xl bg-white px-4 text-sm font-black text-slate-950 hover:bg-white/90 disabled:opacity-45"
                  :disabled="wechatActionDisabled || creating || selectedArticleIds.length === 0 || formats.length === 0 || insufficientBalance"
                  @click="handleCreateTask"
                >
                  {{ creating ? '创建中' : `导出 ${selectedArticleIds.length} 篇` }}
                </button>
              </div>
              <p v-if="insufficientBalance" class="mt-2 text-xs font-semibold text-red-100/80">
                余额不足，请充值后再创建任务
              </p>
              <div v-if="message" class="mt-2 rounded-xl border border-emerald-200/20 bg-emerald-300/10 px-3 py-2 text-sm text-emerald-50">
                {{ message }}
              </div>
              <div v-if="errorMessage" class="mt-2 rounded-xl border border-red-300/20 bg-red-300/10 px-3 py-2 text-sm text-red-100">
                {{ errorMessage }}
              </div>
            </div>

            <!-- 文章列表 -->
            <div v-if="articles.length > 0">
              <div class="rounded-xl border border-white/10 bg-[#101114]">
                <div v-if="filteredArticles.length === 0" class="px-4 py-6 text-center text-xs text-white/45">
                  没有符合条件的文章
                </div>
                <div v-else>
                  <label
                    v-for="article in paginatedArticles"
                    :key="article.id"
                    class="grid cursor-pointer grid-cols-[40px_1fr] gap-2 border-b border-white/10 px-3 py-2 last:border-b-0 hover:bg-white/[0.03]"
                  >
                    <input
                      v-model="selectedArticleIds"
                      type="checkbox"
                      class="mt-0.5 h-4 w-4 accent-cyan-200"
                      :value="article.id"
                      :disabled="wechatActionDisabled"
                    />
                    <div class="min-w-0">
                      <p class="truncate text-sm font-semibold text-white">{{ article.title || article.link }}</p>
                      <p class="mt-0.5 truncate text-xs text-white/45">{{ article.link }}</p>
                      <p class="mt-0.5 text-xs text-white/30">
                        {{ article.content_status || 'pending' }}
                        <span v-if="article.account_fakeid"> · {{ article.account_fakeid }}</span>
                        <span v-if="article.publish_at"> · {{ article.publish_at }}</span>
                      </p>
                    </div>
                  </label>
                </div>
              </div>
              <!-- 分页 -->
              <div v-if="articleTotalPages > 1" class="mt-2 flex items-center justify-between text-xs text-white/50">
                <span>第 {{ articleCurrentPage }} / {{ articleTotalPages }} 页，共 {{ filteredArticles.length }} 篇</span>
                <div class="flex items-center gap-1">
                  <button
                    type="button"
                    class="rounded-lg border border-white/10 px-2 py-1 hover:bg-white/[0.06] disabled:opacity-30"
                    :disabled="articleCurrentPage <= 1"
                    @click="articleCurrentPage--"
                  >上一页</button>
                  <template v-for="p in articleVisiblePageRange" :key="p">
                    <span v-if="p < 0" class="px-1 text-white/30">…</span>
                    <button
                      v-else
                      type="button"
                      class="min-w-[28px] rounded-lg border px-2 py-1"
                      :class="p === articleCurrentPage ? 'border-cyan-300/40 bg-cyan-300/15 text-cyan-100' : 'border-white/10 hover:bg-white/[0.06]'"
                      @click="articleCurrentPage = p"
                    >{{ p }}</button>
                  </template>
                  <button
                    type="button"
                    class="rounded-lg border border-white/10 px-2 py-1 hover:bg-white/[0.06] disabled:opacity-30"
                    :disabled="articleCurrentPage >= articleTotalPages"
                    @click="articleCurrentPage++"
                  >下一页</button>
                </div>
              </div>
              <div v-if="hasMoreRemoteArticles" class="mt-3 flex justify-center">
                <button
                  type="button"
                  class="rounded-xl border border-cyan-300/30 bg-cyan-300/10 px-4 py-2 text-sm font-bold text-cyan-100 hover:bg-cyan-300/20 disabled:opacity-45"
                  :disabled="articleLoadingMore"
                  @click="loadMoreWeChatArticles"
                >
                  {{ articleLoadingMore ? '加载中' : `加载更多文章（${articles.length}/${articleRemoteTotal}）` }}
                </button>
              </div>
            </div>
          </section>
        </div>

        <!-- 任务监控（紧凑） -->
        <section class="rounded-2xl border border-white/10 bg-white/[0.035] p-5">
          <div class="flex items-center justify-between mb-3">
            <h2 class="text-xl font-black text-white">任务监控</h2>
            <div class="flex items-center gap-3">
              <div class="text-xs font-semibold" :class="workerStatusTone">{{ workerStatusLabel }}</div>
              <div class="flex gap-1">
                <span class="rounded-xl bg-white/10 px-2.5 py-1 text-xs font-semibold">{{ workerStatus?.queued_count ?? 0 }} 排队</span>
                <span class="rounded-xl bg-white/10 px-2.5 py-1 text-xs font-semibold">{{ workerStatus?.running_count ?? 0 }} 运行</span>
                <span class="rounded-xl bg-white/10 px-2.5 py-1 text-xs font-semibold">{{ workerStatus?.completed_count ?? 0 }} 完成</span>
              </div>
            </div>
          </div>
          <div v-if="workerStatus?.health === 'waiting' && workerStatus?.message" class="wechat-worker-status-message wechat-worker-status-waiting mb-3 rounded-xl border px-3 py-2 text-xs">
            ⚠️ {{ workerStatus.message }}
          </div>
          <div v-if="workerStatus?.health === 'attention' && workerStatus?.message" class="wechat-worker-status-message wechat-worker-status-attention mb-3 rounded-xl border px-3 py-2 text-xs">
            ⚠️ {{ workerStatus.message }}
          </div>

          <!-- 任务筛选和批量操作 -->
          <div class="flex flex-wrap items-center gap-2 mb-3">
            <select
              v-model="taskStatusFilter"
              class="min-h-9 rounded-xl border border-white/10 bg-[#101114] px-3 text-sm text-white outline-none focus:border-cyan-300/40"
            >
              <option value="all">全部任务</option>
              <option value="queued">排队中</option>
              <option value="running">运行中</option>
              <option value="completed">已完成</option>
              <option value="failed">失败</option>
            </select>
            <button
              type="button"
              class="wechat-task-action wechat-task-action-select rounded-xl border px-3 py-1.5 text-sm font-semibold transition"
              :disabled="wechatActionDisabled"
              @click="selectFilteredTasks"
            >
              全选 ({{ filteredTasks.length }})
            </button>
            <button
              type="button"
              class="wechat-task-action wechat-task-action-neutral rounded-xl border px-3 py-1.5 text-sm font-semibold transition"
              :disabled="wechatActionDisabled"
              @click="clearSelectedTasks"
            >
              清空
            </button>
            <button
              type="button"
              class="wechat-task-action wechat-task-action-danger rounded-xl border px-3 py-1.5 text-sm font-semibold transition"
              :disabled="wechatActionDisabled || Boolean(batchTaskAction) || selectedCancellableTaskIds.length === 0"
              @click="handleBatchCancelTasks"
            >
              {{ batchTaskAction === 'cancel' ? '取消中' : `取消 ${selectedCancellableTaskIds.length}` }}
            </button>
            <button
              type="button"
              class="wechat-task-action wechat-task-action-retry rounded-xl border px-3 py-1.5 text-sm font-semibold transition"
              :disabled="wechatActionDisabled || Boolean(batchTaskAction) || selectedRetryableTaskIds.length === 0"
              @click="handleBatchRetryTasks"
            >
              {{ batchTaskAction === 'retry' ? '重试中' : `重试 ${selectedRetryableTaskIds.length}` }}
            </button>
            <button
              type="button"
              class="wechat-task-action wechat-task-action-download rounded-xl border px-3 py-1.5 text-sm font-semibold transition"
              :disabled="wechatActionDisabled || Boolean(batchTaskAction) || selectedTaskIds.length === 0"
              @click="handleBatchDownloadArtifacts"
            >
              {{ batchTaskAction === 'download' ? '准备中' : '下载产物' }}
            </button>
            <span class="text-xs text-white/40">已选 {{ selectedTaskIds.length }} 个</span>
          </div>

          <!-- 任务列表 -->
          <div v-if="tasks.length > 0" class="max-h-[300px] overflow-auto">
            <div v-if="filteredTasks.length === 0" class="px-4 py-6 text-center text-xs text-white/45">
              没有符合条件的任务
            </div>
            <div v-else class="space-y-2">
              <article v-for="task in filteredTasks" :key="task.id" class="rounded-xl border border-white/10 bg-[#101114] p-4">
                <div class="flex items-start justify-between gap-2">
                  <div class="flex gap-2">
                    <input
                      v-model="selectedTaskIds"
                      type="checkbox"
                      class="mt-0.5 h-4 w-4 accent-cyan-200"
                      :value="task.id"
                      :disabled="wechatActionDisabled"
                    />
                    <div>
                      <div class="flex items-center gap-2">
                        <span class="text-xs font-black text-white">{{ getTaskTitle(task) }}</span>
                        <span class="rounded-xl bg-white/10 px-2 py-0.5 text-xs font-semibold uppercase text-white/60">{{ task.status }}</span>
                      </div>
	                      <div class="mt-1 flex items-center gap-2">
	                        <div class="h-1.5 w-32 overflow-hidden rounded-full bg-white/10">
	                          <div class="h-full rounded-full bg-cyan-200" :style="{ width: `${taskProgress(task)}%` }"></div>
	                        </div>
	                        <span class="text-xs text-white/45">{{ task.successful_article_count }}/{{ task.selected_article_count }}</span>
	                      </div>
	                      <p v-if="taskLeaseState(task)" class="mt-1 text-xs text-cyan-100/70">{{ taskLeaseState(task) }}</p>
	                      <details v-if="task.failed_article_count > 0 || taskFailureSummary(task)" class="mt-2 text-xs text-red-100">
	                        <summary class="cursor-pointer font-semibold">失败详情</summary>
	                        <pre class="mt-2 max-h-32 overflow-auto whitespace-pre-wrap rounded-xl bg-red-950/30 p-2">{{ taskFailureSummary(task) || task.error_message }}</pre>
	                      </details>
	                      <details v-if="taskEngagementSummary(task)" class="mt-2 text-xs text-amber-100">
	                        <summary class="cursor-pointer font-semibold">互动数据提示</summary>
	                        <pre class="mt-2 max-h-32 overflow-auto whitespace-pre-wrap rounded-xl bg-amber-950/20 p-2">{{ taskEngagementSummary(task) }}</pre>
	                      </details>
	                      <div class="mt-2 flex flex-wrap gap-1">
	                        <span
	                          v-for="event in taskTimeline(task).slice(0, 4)"
	                          :key="event.key"
	                          class="inline-flex items-center gap-1 rounded-full bg-white/5 px-2 py-0.5 text-xs text-white/45"
	                        >
	                          <span class="h-1.5 w-1.5 rounded-full" :class="event.tone"></span>
	                          {{ event.label }}
	                        </span>
	                      </div>
	                    </div>
	                  </div>
                  <div class="flex flex-wrap gap-1">
                    <span v-for="format in task.formats" :key="format" class="rounded-xl bg-white/10 px-2 py-0.5 text-xs font-semibold uppercase text-white/60">
                      {{ format }}
                    </span>
                  </div>
                </div>
                <div class="mt-2 flex flex-wrap gap-1">
                  <button
                    v-if="(artifactsByTask[task.id] || []).length > 0"
                    type="button"
                    class="rounded-xl border border-emerald-200/30 bg-emerald-200/10 px-3 py-1.5 text-sm font-semibold text-emerald-100 hover:bg-emerald-200/20 disabled:opacity-45"
                    :disabled="wechatActionDisabled || taskActionId === task.id"
                    @click="handleDownloadTaskZip(task)"
                  >
                    ZIP
                  </button>
                  <button
                    v-if="canCancelTask(task)"
                    type="button"
                    class="wechat-task-action wechat-task-action-danger rounded-xl border px-3 py-1.5 text-sm font-semibold transition"
                    :disabled="wechatActionDisabled || taskActionId === task.id"
                    @click="handleCancelTask(task.id)"
                  >
                    取消
                  </button>
                  <button
                    v-if="canRetryTask(task)"
                    type="button"
                    class="wechat-task-action wechat-task-action-retry rounded-xl border px-3 py-1.5 text-sm font-semibold transition"
                    :disabled="wechatActionDisabled || taskActionId === task.id"
                    @click="handleRetryTask(task.id)"
                  >
                    重试
                  </button>
                  <button
                    v-for="artifact in artifactsByTask[task.id] || []"
                    :key="artifact.id"
                    type="button"
                    class="rounded-xl border border-cyan-300/30 bg-cyan-300/10 px-3 py-1.5 text-sm font-semibold text-cyan-100 hover:bg-cyan-300/20 disabled:opacity-45"
                    :disabled="wechatActionDisabled || taskActionId === task.id"
                    @click="handleDownloadArtifact(artifact)"
                  >
                    {{ artifact.format }}
                  </button>
                </div>
              </article>
            </div>
          </div>
        </section>
      </div>
    </main>
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import PublicDarkHeader from '@/components/layout/PublicDarkHeader.vue'
import {
  bindWeChatAccount,
  cancelWeChatExportTask,
  createWeChatQRCodeSession,
  createWeChatExportTask,
  downloadWeChatExportTaskZip,
  downloadWeChatExportArtifact,
  getWeChatExportWorkerStatus,
  getWeChatSession,
  importWeChatArticleLink,
  listWeChatArticles,
  listWeChatExportArtifacts,
  listWeChatExportTaskLogs,
  listWeChatExportTasks,
  logoutWeChatSession,
  pollWeChatSession,
  quoteWeChatExportTask,
  retryWeChatExportTask,
  searchWeChatAccounts,
  syncWeChatAccount,
  validateWeChatSession,
  type WeChatAccount,
  type WeChatArticle,
  type WeChatExportArtifact,
  type WeChatExportFormat,
  type WeChatExportTask,
  type WeChatExportTaskLog,
  type WeChatExportWorkerStatus,
  type WeChatSession,
} from '@/api/wechat-export'
import { useAuthStore } from '@/stores'

const { t } = useI18n()
const authStore = useAuthStore()

const availableFormats: WeChatExportFormat[] = ['html', 'markdown']

const articleLink = ref('')
const session = ref<WeChatSession | null>(null)
const qrcodeUrl = ref('')
const accounts = ref<WeChatAccount[]>([])
const accountSearchResults = ref<WeChatAccount[]>([])
const accountSearchQuery = ref('')
const accountSearchMessage = ref('')
const articles = ref<WeChatArticle[]>([])
const articleSearchQuery = ref('')
const articleStatusFilter = ref('all')
const articleAccountFilter = ref('all')
const articleCurrentPage = ref(1)
const articlePageSize = 20
const articleRemotePage = ref(1)
const articleRemotePages = ref(0)
const articleRemoteTotal = ref(0)
const articleRemotePageSize = 100
const articleLoadingMore = ref(false)
const tasks = ref<WeChatExportTask[]>([])
const workerStatus = ref<WeChatExportWorkerStatus | null>(null)
const taskStatusFilter = ref('all')
const selectedArticleIds = ref<number[]>([])
const selectedTaskIds = ref<number[]>([])
const formats = ref<WeChatExportFormat[]>(['html', 'markdown'])
const includeEngagement = ref(false)
const artifactsByTask = ref<Record<number, WeChatExportArtifact[]>>({})
const taskLogsByTask = ref<Record<number, WeChatExportTaskLog[]>>({})
const loading = ref(false)
const importing = ref(false)
const creating = ref(false)
const sessionLoading = ref(false)
const sessionValidating = ref(false)
const accountLoading = ref(false)
const accountSearching = ref(false)
const syncingFakeid = ref('')
type SyncingProgress = { fakeid: string; status: string; synced: number; total: number; hasMore: boolean }
const syncingProgress = ref<SyncingProgress | null>(null)
const taskActionId = ref<number | null>(null)
const batchTaskAction = ref<'cancel' | 'retry' | 'download' | ''>('')
const estimatedCredits = ref<number | null>(null)
const message = ref('')
const errorMessage = ref('')
let sessionPollTimer: number | undefined
let taskRefreshTimer: number | undefined
const syncConfirmationPollMs = 3000
const syncConfirmationMaxAttempts = 40

const isAuthenticated = computed(() => authStore.isAuthenticated)
const userBalance = computed(() => authStore.user?.balance ?? 0)
const hasReadySession = computed(() => session.value?.status === 'ready')
const requiresReadySession = computed(() => isAuthenticated.value && !hasReadySession.value)
const wechatActionDisabled = computed(() => !isAuthenticated.value || !hasReadySession.value)
const insufficientBalance = computed(() => {
  if (!isAuthenticated.value) return false
  if (estimatedCredits.value === null) return false
  return userBalance.value < estimatedCredits.value
})
const sessionStatusLabel = computed(() => {
  if (!session.value) return t('wechatExport.session.statusNotConnected')
  if (session.value.status === 'ready') {
    return session.value.login_account_name ? `${t('wechatExport.session.statusReady')} · ${session.value.login_account_name}` : t('wechatExport.session.statusReady')
  }
  if (session.value.status === 'scan_confirmed') return t('wechatExport.session.statusScanConfirmed')
  return session.value.status || t('wechatExport.session.statusNotConnected')
})
const sessionWarning = computed(() => {
  if (!isAuthenticated.value) return ''
  if (!session.value || session.value.status === 'not_connected') {
    return t('wechatExport.session.warningNoSession')
  }
  if (session.value.status === 'expired') {
    return t('wechatExport.session.warningExpired')
  }
  if (session.value.status === 'pending' || session.value.status === 'scan_confirmed') {
    return t('wechatExport.session.warningNotReady')
  }
  return ''
})
const workerStatusLabel = computed(() => {
  if (!workerStatus.value) return 'unknown'
  return workerStatus.value.health
})
const workerStatusTone = computed(() => {
  switch (workerStatus.value?.health) {
    case 'attention':
      return 'text-red-700'
    case 'waiting':
      return 'text-amber-700'
    case 'active':
      return 'text-sky-700'
    case 'idle':
      return 'text-emerald-700'
    default:
      return 'text-slate-500'
  }
})
const isQRCodeImage = computed(() => qrcodeUrl.value.startsWith('data:image/'))
const filteredArticles = computed(() => {
  const query = articleSearchQuery.value.trim().toLowerCase()
  const status = articleStatusFilter.value
  const accountFakeid = articleAccountFilter.value
  return articles.value.filter((article) => {
    if (status !== 'all' && article.content_status !== status) return false
    if (accountFakeid !== 'all' && article.account_fakeid !== accountFakeid) return false
    if (!query) return true
    return [
      article.title,
      article.link,
      article.author || '',
      article.account_fakeid || '',
      article.source_type,
      article.content_status,
    ].some((value) => String(value || '').toLowerCase().includes(query))
  })
})
const articleTotalPages = computed(() => Math.ceil(filteredArticles.value.length / articlePageSize))
const hasMoreRemoteArticles = computed(() => articleRemotePages.value > 0 && articleRemotePage.value < articleRemotePages.value)
const paginatedArticles = computed(() => {
  const start = (articleCurrentPage.value - 1) * articlePageSize
  return filteredArticles.value.slice(start, start + articlePageSize)
})
watch([articleSearchQuery, articleStatusFilter, articleAccountFilter], () => {
  articleCurrentPage.value = 1
})
const articleVisiblePageRange = computed(() => {
  const total = articleTotalPages.value
  const current = articleCurrentPage.value
  if (total <= 7) return Array.from({ length: total }, (_, i) => i + 1)
  const pages: number[] = [1]
  let start = Math.max(2, current - 1)
  let end = Math.min(total - 1, current + 1)
  if (current <= 3) { start = 2; end = 4 }
  if (current >= total - 2) { start = total - 3; end = total - 1 }
  if (start > 2) pages.push(-1) // ellipsis sentinel
  for (let i = start; i <= end; i++) pages.push(i)
  if (end < total - 1) pages.push(-2) // ellipsis sentinel
  pages.push(total)
  return pages
})
const filteredTasks = computed(() => {
  if (taskStatusFilter.value === 'all') return tasks.value
  return tasks.value.filter((task) => task.status === taskStatusFilter.value)
})
const selectedTasks = computed(() => {
  const selected = new Set(selectedTaskIds.value)
  return tasks.value.filter((task) => selected.has(task.id))
})
const selectedCancellableTaskIds = computed(() => selectedTasks.value.filter(canCancelTask).map((task) => task.id))
const selectedRetryableTaskIds = computed(() => selectedTasks.value.filter(canRetryTask).map((task) => task.id))

function getErrorMessage(error: unknown) {
  if (error instanceof Error) return error.message
  if (error && typeof error === 'object') {
    const message = (error as { message?: unknown }).message
    if (typeof message === 'string' && message.trim()) return message
  }
  return '请求失败'
}

function setError(error: unknown) {
  errorMessage.value = getErrorMessage(error)
}

function isTransientSyncRequestError(error: unknown) {
  if (!error || typeof error !== 'object') return false
  const status = (error as { status?: unknown }).status
  const code = String((error as { code?: unknown }).code || '').toUpperCase()
  const message = getErrorMessage(error).toLowerCase()
  return status === 0 || code === 'ECONNABORTED' || message.includes('timeout') || message.includes('network error')
}

function requireReadySession() {
  if (!isAuthenticated.value) return false
  if (hasReadySession.value) return true
  errorMessage.value = '微信会话未登录或已失效，请先扫码登录。'
  return false
}

function syncProgressText(progress: SyncingProgress) {
  const synced = Math.max(0, progress.synced)
  const total = Math.max(0, progress.total)
  if (total > 0) {
    return `${progress.status}: ${Math.min(synced, total)}/${total}`
  }
  return `${progress.status}: 已完成 ${synced} 篇`
}

function delay(ms: number) {
  return new Promise(resolve => window.setTimeout(resolve, ms))
}

function accountLastSyncedAt(fakeid: string) {
  return accounts.value.find(account => account.fakeid === fakeid)?.last_synced_at || ''
}

async function pollSyncConfirmation(fakeid: string, startFrom: number, baselineTotal: number, baselineLastSyncedAt: string) {
  for (let attempt = 1; attempt <= syncConfirmationMaxAttempts; attempt++) {
    await delay(syncConfirmationPollMs)
    if (syncingFakeid.value !== fakeid) return false

    syncingProgress.value = {
      fakeid,
      status: `同步请求仍在处理中，正在确认结果(${attempt}/${syncConfirmationMaxAttempts})...`,
      synced: startFrom,
      total: syncingProgress.value?.total ?? 0,
      hasMore: false,
    }

    const [accountResult, articleResult] = await Promise.all([
      searchWeChatAccounts(),
      listWeChatArticles({ page: 1, page_size: articleRemotePageSize }),
    ])

    accounts.value = accountResult
    articles.value = articleResult.items
    articleRemotePage.value = articleResult.page || 1
    articleRemotePages.value = articleResult.pages || 0
    articleRemoteTotal.value = articleResult.total || articleResult.items.length
    articleCurrentPage.value = 1

    const nextLastSyncedAt = accountLastSyncedAt(fakeid)
    const hasAccountSyncUpdate = Boolean(nextLastSyncedAt && nextLastSyncedAt !== baselineLastSyncedAt)
    const hasArticleGrowth = articleRemoteTotal.value > baselineTotal
    const hasVisibleSyncedArticle = articleResult.items.some(article => article.account_fakeid === fakeid)

    if (hasAccountSyncUpdate || hasArticleGrowth || hasVisibleSyncedArticle) {
      const syncedCount = Math.max(0, articleRemoteTotal.value - baselineTotal)
      syncingProgress.value = {
        fakeid,
        status: '已确认完成',
        synced: startFrom + syncedCount,
        total: articleRemoteTotal.value,
        hasMore: false,
      }
      message.value = syncedCount > 0
        ? `✅ 同步完成！新增 ${syncedCount} 篇文章。`
        : '✅ 同步完成！文章列表已更新。'
      await refreshTasks()
      window.setTimeout(() => {
        if (syncingProgress.value?.fakeid === fakeid) {
          syncingProgress.value = null
        }
      }, 3000)
      return true
    }
  }

  message.value = '同步请求仍在后台处理中，已继续轮询一段时间；请稍后刷新文章列表查看结果。'
  return false
}

async function refreshAll() {
  if (!isAuthenticated.value) return
  loading.value = true
  errorMessage.value = ''
  try {
    const [sessionResult, accountResult, articleResult, taskResult, workerStatusResult] = await Promise.all([
      getWeChatSession().catch(() => ({ status: 'not_connected' })),
      searchWeChatAccounts(),
      listWeChatArticles({ page: 1, page_size: articleRemotePageSize }),
      listWeChatExportTasks(),
      getWeChatExportWorkerStatus().catch(() => null),
    ])
    session.value = sessionResult
    accounts.value = accountResult
    articles.value = articleResult.items
    articleRemotePage.value = articleResult.page || 1
    articleRemotePages.value = articleResult.pages || 0
    articleRemoteTotal.value = articleResult.total || articleResult.items.length
    articleCurrentPage.value = 1
    tasks.value = taskResult.items
    workerStatus.value = workerStatusResult
    void refreshVisibleTaskLogs(taskResult.items)
  } catch (error) {
    setError(error)
  } finally {
    loading.value = false
  }
}

async function loadMoreWeChatArticles() {
  if (!isAuthenticated.value || articleLoadingMore.value || !hasMoreRemoteArticles.value) return
  articleLoadingMore.value = true
  errorMessage.value = ''
  try {
    const nextPage = articleRemotePage.value + 1
    const result = await listWeChatArticles({ page: nextPage, page_size: articleRemotePageSize })
    const existingIds = new Set(articles.value.map((article) => article.id))
    articles.value = [
      ...articles.value,
      ...result.items.filter((article) => !existingIds.has(article.id)),
    ]
    articleRemotePage.value = result.page || nextPage
    articleRemotePages.value = result.pages || articleRemotePages.value
    articleRemoteTotal.value = result.total || articleRemoteTotal.value
  } catch (error) {
    setError(error)
  } finally {
    articleLoadingMore.value = false
  }
}

function startSessionPolling() {
  if (sessionPollTimer) {
    window.clearInterval(sessionPollTimer)
  }
  sessionPollTimer = window.setInterval(() => {
    if (session.value?.id && ['pending', 'scan_confirmed'].includes(session.value.status)) {
      void handlePollSession()
    }
  }, 3000)
}

function startTaskRefresh() {
  if (taskRefreshTimer) {
    window.clearInterval(taskRefreshTimer)
  }
  taskRefreshTimer = window.setInterval(() => {
    if (!isAuthenticated.value) return
    if (tasks.value.some((task) => ['queued', 'running', 'uploading'].includes(task.status))) {
      void refreshTasks()
    }
  }, 5000)
}

async function handleCreateSession() {
  sessionLoading.value = true
  message.value = ''
  errorMessage.value = ''
  try {
    const result = await createWeChatQRCodeSession()
    session.value = result.session
    qrcodeUrl.value = result.qrcode_url
    message.value = '二维码会话已创建，已开始轮询。'
    startSessionPolling()
  } catch (error) {
    setError(error)
  } finally {
    sessionLoading.value = false
  }
}

async function handlePollSession() {
  if (!session.value?.id) return
  try {
    session.value = await pollWeChatSession(session.value.id)
  } catch (error) {
    setError(error)
  }
}

async function handleValidateSession() {
  sessionValidating.value = true
  message.value = ''
  errorMessage.value = ''
  try {
    session.value = await validateWeChatSession()
    message.value = '微信会话仍然有效。'
  } catch (error) {
    setError(error)
    await refreshAll()
  } finally {
    sessionValidating.value = false
  }
}

async function handleLogoutSession() {
  errorMessage.value = ''
  try {
    await logoutWeChatSession()
    session.value = { status: 'not_connected' }
    qrcodeUrl.value = ''
    message.value = '微信导出会话已退出。'
  } catch (error) {
    setError(error)
  }
}

async function handleSearchRemoteAccounts() {
  if (!requireReadySession()) return
  accountSearching.value = true
  accountSearchMessage.value = ''
  errorMessage.value = ''
  try {
    accountSearchResults.value = await searchWeChatAccounts(accountSearchQuery.value, true)
    accountSearchMessage.value = `找到 ${accountSearchResults.value.length} 个公众号。`
  } catch (error) {
    setError(error)
  } finally {
    accountSearching.value = false
  }
}

async function handleBindSearchResult(account: WeChatAccount) {
  if (!requireReadySession()) return
  accountLoading.value = true
  accountSearchMessage.value = ''
  message.value = ''
  errorMessage.value = ''
  try {
    const result = await bindWeChatAccount({
      fakeid: account.fakeid,
      nickname: account.nickname,
      alias: account.alias,
      avatar: account.avatar,
      description: account.description,
    })
    accounts.value = await searchWeChatAccounts()
    // Auto-sync if backend indicates sync_required (API contract)
    if (result.sync_required) {
      message.value = `已绑定 ${account.nickname || account.fakeid}，正在自动同步文章...`
      void handleSyncAccount(account.fakeid, 0, true)
    } else {
      message.value = `已绑定 ${account.nickname || account.fakeid}。`
    }
  } catch (error) {
    setError(error)
  } finally {
    accountLoading.value = false
  }
}

async function handleSyncAccount(fakeid: string, beginFrom?: number, autoMode?: boolean) {
  if (!requireReadySession()) return
  errorMessage.value = ''
  syncingFakeid.value = fakeid

  const startFrom = beginFrom ?? 0
  const previousProgress = syncingProgress.value?.fakeid === fakeid ? syncingProgress.value : null
  const knownTotal = previousProgress && previousProgress.total > 0 ? previousProgress.total : 0
  const baselineArticleTotal = articleRemoteTotal.value
  const baselineLastSyncedAt = accountLastSyncedAt(fakeid)
  syncingProgress.value = {
    fakeid,
    status: startFrom > 0 ? '继续同步...' : '正在连接...',
    synced: startFrom,
    total: knownTotal,
    hasMore: false,
  }

  try {
    syncingProgress.value = {
      fakeid,
      status: '正在同步文章...',
      synced: startFrom,
      total: knownTotal,
      hasMore: false,
    }

    const response = await syncWeChatAccount(fakeid, startFrom)
    const result = response.result
    const totalSynced = startFrom + (result?.synced_count ?? 0)

    syncingProgress.value = {
      fakeid,
      status: '已完成',
      synced: totalSynced,
      total: result?.total_count ?? totalSynced,
      hasMore: result?.has_more ?? false,
    }

    // 如果有更多文章，自动继续同步
    // Batch sync strategy: Backend syncs up to 1000 articles per request
    // Frontend auto-continues with 2-second delay to avoid API overload
    // This allows unlimited total sync while respecting WeChat API rate limits
    if (result?.has_more) {
      if (autoMode ?? true) {
        // 自动模式：显示进度后延迟继续
        syncingProgress.value.status = '同步进度，准备继续...'
        const totalText = result.total_count > 0 ? `/${result.total_count}` : ''
        message.value = `已同步 ${totalSynced}${totalText} 篇文章，自动继续同步中...`

        // 刷新文章列表
        await refreshAll()

        // 延迟 2 秒后继续同步
        setTimeout(() => {
          void handleSyncAccount(fakeid, totalSynced, true)
        }, 2000)
      } else {
        // 手动模式：保留继续同步按钮
        message.value = t('wechatExport.account.syncResult', {
          synced: result?.synced_count ?? 0,
          pages: result?.page_count ?? 0,
          more: result?.has_more ? t('wechatExport.account.syncHasMore') : '',
        })
        await refreshAll()
      }
    } else {
      // 所有文章同步完成
      message.value = `✅ 同步完成！共同步 ${totalSynced} 篇文章。`
      await refreshAll()

      // 3 秒后清除进度显示
      setTimeout(() => {
        if (syncingProgress.value?.fakeid === fakeid) {
          syncingProgress.value = null
        }
      }, 3000)
    }
  } catch (error) {
    if (isTransientSyncRequestError(error)) {
      message.value = '同步请求已提交，前端连接超时，正在轮询确认结果...'
      errorMessage.value = ''
      const confirmed = await pollSyncConfirmation(fakeid, startFrom, baselineArticleTotal, baselineLastSyncedAt)
      if (!confirmed) {
        syncingProgress.value = null
      }
    } else {
      setError(error)
      syncingProgress.value = null
    }
  } finally {
    // 只有在非自动模式或同步完成时才清空 syncingFakeid
    if (!syncingProgress.value?.hasMore) {
      syncingFakeid.value = ''
    }
  }
}

async function handleContinueSync(fakeid: string, beginFrom: number) {
  if (!requireReadySession()) return
  // 手动触发继续同步（如果自动同步失败）
  await handleSyncAccount(fakeid, beginFrom, false)
}

async function handleImport() {
  if (!requireReadySession()) return
  importing.value = true
  message.value = ''
  errorMessage.value = ''
  try {
    const article = await importWeChatArticleLink(articleLink.value)
    articleLink.value = ''
    message.value = '文章链接已导入。'
    await refreshAll()
    if (!selectedArticleIds.value.includes(article.id)) {
      selectedArticleIds.value.push(article.id)
    }
  } catch (error) {
    setError(error)
  } finally {
    importing.value = false
  }
}

async function handleCreateTask() {
  if (!requireReadySession()) return
  creating.value = true
  message.value = ''
  errorMessage.value = ''
  try {
    const task = await createWeChatExportTask({
      article_ids: selectedArticleIds.value,
      formats: formats.value,
      include_engagement: includeEngagement.value,
    })
    message.value = `任务 #${task.id} 已排队，worker 会生成导出产物。`
    selectedArticleIds.value = []
    await refreshAll()
    startTaskRefresh()
  } catch (error) {
    setError(error)
  } finally {
    creating.value = false
  }
}

function selectFilteredArticles() {
  if (!requireReadySession()) return
  const next = new Set(selectedArticleIds.value)
  for (const article of filteredArticles.value) {
    next.add(article.id)
  }
  selectedArticleIds.value = Array.from(next)
}

function clearSelectedArticles() {
  if (!requireReadySession()) return
  selectedArticleIds.value = []
}

async function fetchQuote() {
  if (!isAuthenticated.value || !hasReadySession.value || selectedArticleIds.value.length === 0 || formats.value.length === 0) {
    estimatedCredits.value = null
    return
  }
  try {
    const quote = await quoteWeChatExportTask({
      article_ids: selectedArticleIds.value,
      formats: formats.value,
      include_engagement: includeEngagement.value,
    })
    estimatedCredits.value = quote.estimated_credits
  } catch (error) {
    estimatedCredits.value = null
    console.error('Failed to fetch quote:', error)
  }
}

function selectFilteredTasks() {
  if (!requireReadySession()) return
  const next = new Set(selectedTaskIds.value)
  for (const task of filteredTasks.value) {
    next.add(task.id)
  }
  selectedTaskIds.value = Array.from(next)
}

function clearSelectedTasks() {
  if (!requireReadySession()) return
  selectedTaskIds.value = []
}

async function refreshTasks() {
  const [result, status] = await Promise.all([
    listWeChatExportTasks(),
    getWeChatExportWorkerStatus().catch(() => null),
  ])
  tasks.value = result.items
  workerStatus.value = status
  void refreshVisibleTaskLogs(result.items)
}

async function refreshVisibleTaskLogs(items = tasks.value) {
  const entries = await Promise.all(items.map(async (task) => {
    try {
      // 同时加载日志和产物
      const logs = await listWeChatExportTaskLogs(task.id)
      // 自动加载产物（如果任务已完成）
      if (['completed', 'completed_with_errors'].includes(task.status)) {
        try {
          const artifacts = await listWeChatExportArtifacts(task.id)
          artifactsByTask.value = {
            ...artifactsByTask.value,
            [task.id]: artifacts,
          }
        } catch {
          // 产物加载失败不影响其他任务
        }
      }
      return [task.id, logs] as const
    } catch {
      return [task.id, taskLogsByTask.value[task.id] || []] as const
    }
  }))
  taskLogsByTask.value = {
    ...taskLogsByTask.value,
    ...Object.fromEntries(entries),
  }
}

async function handleBatchCancelTasks() {
  if (!requireReadySession()) return
  const taskIds = selectedCancellableTaskIds.value
  if (taskIds.length === 0) return
  batchTaskAction.value = 'cancel'
  errorMessage.value = ''
  try {
    let successCount = 0
    const failures: string[] = []
    for (const taskId of taskIds) {
      try {
        await cancelWeChatExportTask(taskId)
        successCount++
      } catch (error) {
        failures.push(`#${taskId}: ${error instanceof Error ? error.message : '请求失败'}`)
      }
    }
    message.value = `已取消 ${successCount} 个任务。`
    if (failures.length > 0) {
      errorMessage.value = `部分任务取消失败：${failures.join('; ')}`
    }
    selectedTaskIds.value = selectedTaskIds.value.filter((id) => !taskIds.includes(id))
    await refreshTasks()
  } finally {
    batchTaskAction.value = ''
  }
}

async function handleBatchRetryTasks() {
  if (!requireReadySession()) return
  const taskIds = selectedRetryableTaskIds.value
  if (taskIds.length === 0) return
  batchTaskAction.value = 'retry'
  errorMessage.value = ''
  try {
    let successCount = 0
    const failures: string[] = []
    for (const taskId of taskIds) {
      try {
        await retryWeChatExportTask(taskId)
        artifactsByTask.value = {
          ...artifactsByTask.value,
          [taskId]: [],
        }
        successCount++
      } catch (error) {
        failures.push(`#${taskId}: ${error instanceof Error ? error.message : '请求失败'}`)
      }
    }
    message.value = `已重新排队 ${successCount} 个任务。`
    if (failures.length > 0) {
      errorMessage.value = `部分任务重试失败：${failures.join('; ')}`
    }
    selectedTaskIds.value = selectedTaskIds.value.filter((id) => !taskIds.includes(id))
    await refreshTasks()
    if (successCount > 0) {
      startTaskRefresh()
      // 检查 worker 状态
      const status = await getWeChatExportWorkerStatus().catch(() => null)
      if (status?.health === 'waiting') {
        message.value += ' 但当前没有 worker 处理任务，请联系管理员启动 worker 进程。'
      }
    }
  } finally {
    batchTaskAction.value = ''
  }
}

async function handleBatchDownloadArtifacts() {
  if (!requireReadySession()) return
  const taskIds = selectedTaskIds.value
  if (taskIds.length === 0) return
  batchTaskAction.value = 'download'
  errorMessage.value = ''
  try {
    const downloadableTaskIds = taskIds.filter((taskId) => {
      const task = tasks.value.find((item) => item.id === taskId)
      return task && ['completed', 'completed_with_errors'].includes(task.status)
    })
    if (downloadableTaskIds.length === 0) {
      message.value = '没有选择可下载 ZIP 的已完成任务。'
      return
    }

    let successCount = 0
    const failures: string[] = []
    for (const taskId of downloadableTaskIds) {
      try {
        const task = tasks.value.find((item) => item.id === taskId)
        if (!task) continue
        const taskTitle = getTaskTitle(task)
        const blob = await downloadWeChatExportTaskZip(taskId)
        const fileName = `${taskTitle.replace(/[^a-zA-Z0-9一-龥]/g, '_')}.zip`
        triggerBrowserDownload(blob, fileName)
        successCount++
      } catch (error) {
        failures.push(`#${taskId}: ${error instanceof Error ? error.message : '下载失败'}`)
      }
    }
    if (successCount > 0) {
      message.value = `已下载 ${successCount}/${downloadableTaskIds.length} 个任务 ZIP。`
    } else {
      message.value = ''
    }
    if (failures.length > 0) {
      errorMessage.value = `部分任务 ZIP 下载失败：${failures.join('; ')}`
    }
  } catch (error) {
    setError(error)
  } finally {
    batchTaskAction.value = ''
  }
}

async function handleCancelTask(taskId: number) {
  if (!requireReadySession()) return
  taskActionId.value = taskId
  errorMessage.value = ''
  try {
    await cancelWeChatExportTask(taskId)
    message.value = `任务 #${taskId} 已取消。`
    await refreshTasks()
  } catch (error) {
    setError(error)
  } finally {
    taskActionId.value = null
  }
}

async function handleRetryTask(taskId: number) {
  if (!requireReadySession()) return
  taskActionId.value = taskId
  errorMessage.value = ''
  try {
    await retryWeChatExportTask(taskId)
    artifactsByTask.value = {
      ...artifactsByTask.value,
      [taskId]: [],
    }
    message.value = `任务 #${taskId} 已重新排队。`
    await refreshTasks()
    startTaskRefresh()
    // 检查 worker 状态，如果没有 worker 运行，给出提示
    const status = await getWeChatExportWorkerStatus().catch(() => null)
    if (status?.health === 'waiting') {
      message.value += ' 但当前没有 worker 处理任务，请联系管理员启动 worker 进程。'
    }
  } catch (error) {
    setError(error)
  } finally {
    taskActionId.value = null
  }
}

async function handleDownloadTaskZip(task: WeChatExportTask) {
  if (!requireReadySession()) return
  taskActionId.value = task.id
  errorMessage.value = ''
  try {
    const taskTitle = getTaskTitle(task)
    const blob = await downloadWeChatExportTaskZip(task.id)
    const fileName = `${taskTitle.replace(/[^a-zA-Z0-9一-龥]/g, '_')}.zip`
    triggerBrowserDownload(blob, fileName)
  } catch (error) {
    setError(error)
  } finally {
    taskActionId.value = null
  }
}

async function handleDownloadArtifact(artifact: WeChatExportArtifact) {
  if (!requireReadySession()) return
  taskActionId.value = artifact.task_id
  errorMessage.value = ''
  try {
    const blob = await downloadWeChatExportArtifact(artifact.id)
    triggerBrowserDownload(blob, artifact.file_name)
  } catch (error) {
    setError(error)
  } finally {
    taskActionId.value = null
  }
}

function triggerBrowserDownload(blob: Blob, fileName: string) {
  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = fileName
  document.body.appendChild(link)
  link.click()
  link.remove()
  URL.revokeObjectURL(url)
}

function taskProgress(task: WeChatExportTask) {
  if (task.selected_article_count <= 0) return task.status === 'completed' ? 100 : 0
  if (task.status === 'completed' || task.status === 'completed_with_errors') return 100
  if (task.status === 'failed' || task.status === 'cancelled') {
    return Math.round(((task.successful_article_count + task.failed_article_count) / task.selected_article_count) * 100)
  }
  return Math.max(5, Math.round(((task.successful_article_count + task.failed_article_count) / task.selected_article_count) * 100))
}

function canCancelTask(task: WeChatExportTask) {
  return ['queued', 'running'].includes(task.status)
}

function canRetryTask(task: WeChatExportTask) {
  return ['failed', 'completed_with_errors', 'cancelled'].includes(task.status)
}

function taskFailureSummary(task: WeChatExportTask) {
  const raw = task.result_manifest_json
  if (!raw) return ''
  try {
    const parsed = JSON.parse(raw)
    if (Array.isArray(parsed.failed_articles) && parsed.failed_articles.length > 0) {
      return JSON.stringify(parsed.failed_articles, null, 2)
    }
  } catch {
    return ''
  }
  return ''
}

function getTaskTitle(task: WeChatExportTask): string {
  // 从task.article_ids中获取第一篇文章的标题
  const firstArticleId = task.article_ids[0]
  if (!firstArticleId) return `任务 #${task.id}`
  const firstArticle = articles.value.find(a => a.id === firstArticleId)
  if (!firstArticle || !firstArticle.title) return `任务 #${task.id}`
  // 如果有多篇文章，显示标题 + 数量
  if (task.selected_article_count > 1) {
    return `${firstArticle.title.slice(0, 30)}... 等${task.selected_article_count}篇`
  }
  return firstArticle.title
}

function taskEngagementSummary(task: WeChatExportTask) {
  const raw = task.result_manifest_json
  if (!raw) return ''
  try {
    const parsed = JSON.parse(raw)
    const notes = collectEngagementNotes(parsed)
    return notes.length > 0 ? JSON.stringify(notes, null, 2) : ''
  } catch {
    return ''
  }
}

function collectEngagementNotes(value: unknown): Array<Record<string, unknown>> {
  const notes: Array<Record<string, unknown>> = []
  const visit = (item: unknown, path: string) => {
    if (!item || typeof item !== 'object') return
    if (Array.isArray(item)) {
      item.forEach((child, index) => visit(child, `${path}[${index}]`))
      return
    }
    const record = item as Record<string, unknown>
    const status = record.engagementFetchStatus || record.engagement_fetch_status
    if (status && status !== 'skipped' && status !== 'fetched') {
      notes.push({
        path,
        status,
        message: record.engagementFetchMessage || record.engagement_fetch_message || '互动数据不可用。',
      })
    }
    for (const [key, child] of Object.entries(record)) {
      if (child && typeof child === 'object') {
        visit(child, path ? `${path}.${key}` : key)
      }
    }
  }
  visit(value, 'manifest')
  return notes
}

function taskTimeline(task: WeChatExportTask) {
  const persistedLogs = taskLogsByTask.value[task.id] || []
  if (persistedLogs.length > 0) {
    return persistedLogs.map((log) => ({
      key: `log-${log.id}`,
      label: taskLogLabel(log),
      time: log.created_at,
      description: log.message || log.status,
      tone: taskLogTone(log),
    }))
  }
  const events: Array<{ key: string; label: string; time?: string; description?: string; tone: string }> = [
    {
      key: 'created',
      label: '任务已创建',
      time: task.created_at,
      description: `${task.selected_article_count} 篇文章，格式：${task.formats.join(', ')}`,
      tone: 'bg-white/40',
    },
  ]

  if (['queued', 'running', 'uploading', 'completed', 'completed_with_errors', 'failed', 'cancelled'].includes(task.status)) {
    events.push({
      key: 'queued',
      label: '等待 worker 领取',
      description: task.worker_lease_until ? `Worker 租约到期 ${task.worker_lease_until}` : '正在等待 worker 领取任务。',
      tone: 'bg-cyan-200/70',
    })
  }
  if (['running', 'uploading', 'completed', 'completed_with_errors', 'failed'].includes(task.status)) {
    events.push({
      key: 'running',
      label: 'Worker 处理中',
      description: `已处理 ${task.successful_article_count + task.failed_article_count}/${task.selected_article_count} 篇文章。`,
      tone: 'bg-blue-200/70',
    })
  }
  if (task.failed_article_count > 0) {
    events.push({
      key: 'failed_articles',
      label: '文章失败已记录',
      description: `${task.failed_article_count} 篇文章失败，可展开失败详情查看 manifest 输出。`,
      tone: 'bg-amber-200/80',
    })
  }
  if (task.status === 'completed' || task.status === 'completed_with_errors') {
    events.push({
      key: 'completed',
      label: task.status === 'completed' ? '已完成' : '部分完成',
      time: task.updated_at,
      description: `已成功导出 ${task.successful_article_count} 篇文章。`,
      tone: task.status === 'completed' ? 'bg-emerald-200/80' : 'bg-amber-200/80',
    })
  }
  if (task.status === 'failed') {
    events.push({
      key: 'failed',
      label: '任务失败',
      time: task.updated_at,
      description: task.error_message || 'Worker 上报任务失败。',
      tone: 'bg-red-200/80',
    })
  }
  if (task.status === 'cancelled') {
    events.push({
      key: 'cancelled',
      label: '任务已取消',
      time: task.updated_at,
      description: '任务在完成前被取消。',
      tone: 'bg-white/30',
    })
  }
  if (task.expires_at) {
    events.push({
      key: 'expires',
      label: '产物过期时间',
      time: task.expires_at,
      description: '请在保留期结束前下载或保存产物。',
      tone: 'bg-purple-200/70',
    })
  }
  return events
}

function taskLogLabel(log: WeChatExportTaskLog) {
  const labels: Record<string, string> = {
    task_created: '任务已创建',
    task_claimed: 'Worker 已领取任务',
    task_completed: log.status === 'completed_with_errors' ? '部分完成' : '已完成',
    task_failed: '任务失败',
    task_cancelled: '任务已取消',
    task_retried: '任务已重试',
    article_fetch_started: '开始抓取文章',
    article_fetched: '文章已抓取',
    article_engagement_checked: '互动数据已检查',
    article_enriched: '文章元数据已保存',
    article_failed: '文章失败',
    artifacts_generated: '产物已生成',
  }
  return labels[log.event] || log.event || '任务事件'
}

function taskLogTone(log: WeChatExportTaskLog) {
  if (log.status === 'failed') return 'bg-red-200/80'
  if (log.status === 'completed_with_errors') return 'bg-amber-200/80'
  if (log.status === 'completed') return 'bg-emerald-200/80'
  if (log.status === 'running') return 'bg-blue-200/70'
  if (log.status === 'queued') return 'bg-cyan-200/70'
  return 'bg-white/40'
}

function taskLeaseState(task: WeChatExportTask) {
  if (task.status !== 'running' || !task.worker_lease_until) return ''
  const leaseTime = new Date(task.worker_lease_until).getTime()
  if (!Number.isFinite(leaseTime)) return ''
  const diffSeconds = Math.abs(Math.round((leaseTime - Date.now()) / 1000))
  return leaseTime < Date.now()
    ? `租约已过期 ${formatDurationSeconds(diffSeconds)}，将被回收`
    : `Worker 租约剩余 ${formatDurationSeconds(diffSeconds)}`
}

function formatDurationSeconds(value: number) {
  const seconds = Math.max(0, Math.floor(value))
  if (seconds < 60) return `${seconds}s`
  const minutes = Math.floor(seconds / 60)
  if (minutes < 60) return `${minutes}m`
  const hours = Math.floor(minutes / 60)
  if (hours < 24) return `${hours}h ${minutes % 60}m`
  const days = Math.floor(hours / 24)
  return `${days}d ${hours % 24}h`
}

// 监听选择变化，自动获取报价
watch([selectedArticleIds, formats, includeEngagement], () => {
  void fetchQuote()
}, { immediate: true })

onMounted(() => {
  void refreshAll()
  startSessionPolling()
  startTaskRefresh()
})

onBeforeUnmount(() => {
  if (sessionPollTimer) {
    window.clearInterval(sessionPollTimer)
  }
  if (taskRefreshTimer) {
    window.clearInterval(taskRefreshTimer)
  }
})
</script>

<style scoped>
.wechat-export-warning {
  border-color: rgba(217, 119, 6, 0.24) !important;
  background-color: rgba(255, 251, 235, 0.96) !important;
  color: rgb(120, 53, 15) !important;
}

.dark .wechat-export-warning {
  border-color: rgba(251, 191, 36, 0.3) !important;
  background-color: rgba(120, 53, 15, 0.26) !important;
  color: rgb(253, 230, 138) !important;
}

.wechat-task-action {
  opacity: 1 !important;
}

.wechat-task-action-select {
  border-color: rgba(15, 23, 42, 0.16) !important;
  background-color: rgb(63, 63, 70) !important;
  color: rgb(255, 255, 255) !important;
}

.wechat-task-action-neutral {
  border-color: rgb(226, 232, 240) !important;
  background-color: rgb(255, 255, 255) !important;
  color: rgb(71, 85, 105) !important;
}

.wechat-task-action-danger {
  border-color: rgb(253, 186, 116) !important;
  background-color: rgb(255, 247, 237) !important;
  color: rgb(154, 52, 18) !important;
}

.wechat-task-action-retry {
  border-color: rgba(15, 23, 42, 0.16) !important;
  background-color: rgb(82, 82, 91) !important;
  color: rgb(255, 255, 255) !important;
}

.wechat-task-action-download {
  border-color: rgb(191, 219, 254) !important;
  background-color: rgb(239, 246, 255) !important;
  color: rgb(29, 78, 216) !important;
}

.wechat-task-action:hover:not(:disabled) {
  filter: brightness(0.96);
}

.wechat-task-action:disabled {
  cursor: not-allowed;
  border-color: rgb(226, 232, 240) !important;
  background-color: rgb(248, 250, 252) !important;
  color: rgb(100, 116, 139) !important;
  filter: none;
}

.dark .wechat-task-action:disabled {
  border-color: rgba(148, 163, 184, 0.24) !important;
  background-color: rgba(15, 23, 42, 0.72) !important;
  color: rgb(203, 213, 225) !important;
}

.wechat-worker-status-message {
  line-height: 1.65;
  font-weight: 600;
}

.wechat-worker-status-waiting {
  border-color: rgb(253, 186, 116) !important;
  background-color: rgb(255, 247, 237) !important;
  color: rgb(154, 52, 18) !important;
}

.wechat-worker-status-attention {
  border-color: rgb(254, 202, 202) !important;
  background-color: rgb(254, 242, 242) !important;
  color: rgb(185, 28, 28) !important;
}

.dark .wechat-worker-status-waiting {
  border-color: rgba(251, 191, 36, 0.28) !important;
  background-color: rgba(120, 53, 15, 0.26) !important;
  color: rgb(253, 230, 138) !important;
}

.dark .wechat-worker-status-attention {
  border-color: rgba(248, 113, 113, 0.3) !important;
  background-color: rgba(127, 29, 29, 0.28) !important;
  color: rgb(254, 202, 202) !important;
}
</style>
