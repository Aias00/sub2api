<template>
  <AppLayout>
    <div class="space-y-5">
      <div class="overflow-hidden rounded-2xl border border-gray-200 bg-white shadow-sm dark:border-dark-700 dark:bg-dark-900">
        <div class="flex flex-col gap-4 p-6 lg:flex-row lg:items-center lg:justify-between">
          <div>
            <p class="text-sm font-semibold uppercase tracking-[0.2em] text-primary-600 dark:text-primary-300">
              {{ t('admin.settings.runtime.badge') }}
            </p>
            <h1 class="mt-2 text-3xl font-black tracking-tight text-gray-950 dark:text-white sm:text-4xl">
              {{ t('admin.settings.runtime.title') }}
            </h1>
            <p class="mt-2 max-w-3xl text-sm leading-6 text-gray-500 dark:text-dark-300">
              {{ t('admin.settings.runtime.description') }}
            </p>
          </div>
          <div class="flex flex-wrap items-center gap-2">
            <button class="btn btn-secondary" type="button" :disabled="loading" @click="loadSettings">
              <Icon name="refresh" size="md" :class="loading ? 'animate-spin' : ''" />
              {{ t('common.refresh') }}
            </button>
            <button
              type="button"
              class="btn btn-secondary"
              :disabled="loading || saving"
              @click="formatAllJson"
            >
              <Icon name="terminal" size="md" />
              {{ t('admin.settings.runtime.formatAllJson') }}
            </button>
            <button
              type="button"
              class="btn btn-primary"
              :disabled="loading || saving || hasValidationErrors"
              @click="saveSettings"
            >
              <Icon name="check" size="md" />
              {{ saving ? t('common.saving') : t('common.save') }}
            </button>
          </div>
        </div>
      </div>

      <div v-if="loading" class="rounded-2xl border border-gray-200 bg-white p-8 text-center text-sm text-gray-500 shadow-sm dark:border-dark-700 dark:bg-dark-900 dark:text-dark-300">
        {{ t('common.loading') }}
      </div>

      <section v-if="!loading" class="settings-card">
        <div class="flex flex-col gap-3 border-b border-gray-100 pb-4 dark:border-dark-700 sm:flex-row sm:items-center sm:justify-between">
          <div>
            <p class="text-sm font-bold text-gray-950 dark:text-white">Worker 运行状态</p>
            <p class="mt-1 text-xs leading-5 text-gray-500 dark:text-dark-300">
              微信导出、生图工作台、热点采集的运行时健康与队列状态。
            </p>
          </div>
          <button class="btn btn-secondary btn-sm" type="button" :disabled="loadingWorkers" @click="loadWorkerStatuses">
            <Icon name="refresh" size="sm" :class="loadingWorkers ? 'animate-spin' : ''" />
            刷新 Worker
          </button>
        </div>
        <div v-if="workerStatusError" class="mt-4 rounded-xl border border-amber-200 bg-amber-50 px-4 py-3 text-sm text-amber-800 dark:border-amber-900/60 dark:bg-amber-950/30 dark:text-amber-200">
          {{ workerStatusError }}
        </div>
        <div class="mt-4 grid gap-4 lg:grid-cols-3">
          <article
            v-for="worker in workerStatuses"
            :key="worker.id"
            class="rounded-2xl border border-gray-200 bg-gray-50 p-4 dark:border-dark-700 dark:bg-dark-800/70"
          >
            <div class="flex items-start justify-between gap-3">
              <div>
                <p class="text-sm font-bold text-gray-950 dark:text-white">{{ worker.name }}</p>
                <p class="mt-1 text-xs text-gray-500 dark:text-dark-300">{{ worker.message || workerStatusMessage(worker.health) }}</p>
              </div>
              <span :class="workerHealthClass(worker.health)">
                {{ workerHealthLabel(worker.health) }}
              </span>
            </div>
            <dl class="mt-4 grid grid-cols-3 gap-2 text-center">
              <div class="rounded-xl bg-white p-2 dark:bg-dark-900">
                <dt class="text-[11px] text-gray-500 dark:text-dark-300">队列</dt>
                <dd class="mt-1 text-base font-black text-gray-950 dark:text-white">{{ formatWorkerNumber(worker.queue) }}</dd>
              </div>
              <div class="rounded-xl bg-white p-2 dark:bg-dark-900">
                <dt class="text-[11px] text-gray-500 dark:text-dark-300">运行</dt>
                <dd class="mt-1 text-base font-black text-gray-950 dark:text-white">{{ formatWorkerNumber(worker.running) }}</dd>
              </div>
              <div class="rounded-xl bg-white p-2 dark:bg-dark-900">
                <dt class="text-[11px] text-gray-500 dark:text-dark-300">异常</dt>
                <dd class="mt-1 text-base font-black text-gray-950 dark:text-white">{{ formatWorkerNumber(worker.failed) }}</dd>
              </div>
            </dl>
            <div class="mt-3 space-y-1 text-xs text-gray-500 dark:text-dark-300">
              <p>总量：{{ formatWorkerNumber(worker.total) }} · 成功：{{ formatWorkerNumber(worker.succeeded) }} · 卡死：{{ formatWorkerNumber(worker.stale) }}</p>
              <p>最后更新：{{ formatWorkerTime(worker.last_updated_at) }}</p>
              <p v-if="worker.status_path" class="truncate" :title="worker.status_path">状态文件：{{ worker.status_path }}</p>
              <p v-if="worker.attention_reasons?.length" class="text-amber-700 dark:text-amber-300">
                注意：{{ worker.attention_reasons.join(', ') }}
              </p>
            </div>
          </article>
          <div v-if="!workerStatuses.length" class="rounded-2xl border border-dashed border-gray-200 p-6 text-center text-sm text-gray-500 dark:border-dark-700 dark:text-dark-300 lg:col-span-3">
            {{ loadingWorkers ? '正在加载 Worker 状态…' : '暂无 Worker 状态数据' }}
          </div>
        </div>
      </section>

      <div v-if="!loading" class="grid gap-5 xl:grid-cols-[1fr_1fr]">
        <section class="settings-card">
          <SectionHeader :title="t('admin.settings.runtime.brandSection')" />
          <div class="grid gap-4 md:grid-cols-2">
            <TextField v-model="form.app_url" :label="t('admin.settings.runtime.appUrl')" :placeholder="t('admin.settings.runtime.appUrlPlaceholder')" :hint="t('admin.settings.runtime.appUrlHint')" class="md:col-span-2" />
            <TextField v-model="form.app_name" :label="t('admin.settings.runtime.appName')" :placeholder="t('admin.settings.runtime.appNamePlaceholder')" />
            <TextField v-model="form.app_description" :label="t('admin.settings.runtime.appDescription')" :placeholder="t('admin.settings.runtime.appDescriptionPlaceholder')" />
            <TextField v-model="form.app_logo" :label="t('admin.settings.runtime.appLogo')" :placeholder="t('admin.settings.runtime.imageUrlPlaceholder')" />
            <TextField v-model="form.app_favicon" :label="t('admin.settings.runtime.appFavicon')" :placeholder="t('admin.settings.runtime.imageUrlPlaceholder')" />
            <TextField v-model="form.app_preview_image" :label="t('admin.settings.runtime.appPreviewImage')" :placeholder="t('admin.settings.runtime.imageUrlPlaceholder')" class="md:col-span-2" />
            <TextField v-model="form.theme" :label="t('admin.settings.runtime.theme')" :placeholder="t('admin.settings.runtime.themePlaceholder')" />
            <TextField v-model="form.appearance" :label="t('admin.settings.runtime.appearance')" :placeholder="t('admin.settings.runtime.appearancePlaceholder')" />
            <TextField v-model="form.default_locale" :label="t('admin.settings.runtime.defaultLocale')" :placeholder="t('admin.settings.runtime.defaultLocalePlaceholder')" />
            <SwitchField v-model="form.locale_detect_enabled" :label="t('admin.settings.runtime.localeDetectEnabled')" :hint="t('admin.settings.runtime.localeDetectEnabledHint')" />
          </div>
        </section>

        <section class="settings-card">
          <SectionHeader :title="t('admin.settings.runtime.authSection')" />
          <div class="grid gap-4 md:grid-cols-3">
            <SwitchField v-model="form.email_auth_visible" :label="t('admin.settings.runtime.emailAuthVisible')" />
            <SwitchField v-model="form.google_auth_visible" :label="t('admin.settings.runtime.googleAuthVisible')" />
            <SwitchField v-model="form.github_auth_visible" :label="t('admin.settings.runtime.githubAuthVisible')" />
            <LocaleEnvelopeEditor v-model="form.auth_shell_config" :label="t('admin.settings.runtime.authShellConfig')" :hint="t('admin.settings.runtime.authShellConfigHint')" class="md:col-span-3" />
          </div>
        </section>

        <section class="settings-card xl:col-span-2">
          <SectionHeader :title="t('admin.settings.runtime.modelPlazaSection')" />
          <div class="space-y-4">
            <LocaleEnvelopeEditor v-model="form.model_plaza_shell_config" :label="t('admin.settings.runtime.modelPlazaShellConfig')" :hint="t('admin.settings.runtime.modelPlazaShellConfigHint')" />
          </div>
        </section>

        <section class="settings-card xl:col-span-2">
          <SectionHeader :title="t('admin.settings.runtime.homeSection')" />
          <div class="space-y-4">
            <LocaleEnvelopeEditor v-model="form.home_shell_config" :label="t('admin.settings.runtime.homeShellConfig')" :hint="t('admin.settings.runtime.homeShellConfigHint')" />
            <LocaleEnvelopeEditor v-model="form.home_business_shell_config" :label="t('admin.settings.runtime.homeBusinessShellConfig')" :hint="t('admin.settings.runtime.homeBusinessShellConfigHint')" />
          </div>
        </section>

        <section class="settings-card xl:col-span-2">
          <SectionHeader :title="t('admin.settings.runtime.docsSection')" />
          <div class="space-y-4">
            <TextField v-model="form.docs_content_base_path" :label="t('admin.settings.runtime.docsContentBasePath')" :placeholder="t('admin.settings.runtime.docsContentBasePathPlaceholder')" :hint="t('admin.settings.runtime.docsContentBasePathHint')" />
            <LocaleEnvelopeEditor v-model="form.docs_shell_config" :label="t('admin.settings.runtime.docsShellConfig')" :hint="t('admin.settings.runtime.docsShellConfigHint')" />
            <LocaleEnvelopeEditor v-model="form.legal_document_shell_config" :label="t('admin.settings.runtime.legalDocumentShellConfig')" :hint="t('admin.settings.runtime.legalDocumentShellConfigHint')" />
            <LocaleEnvelopeEditor v-model="form.api_keys_shell_config" :label="t('admin.settings.runtime.apiKeysShellConfig')" :hint="t('admin.settings.runtime.apiKeysShellConfigHint')" />
            <LocaleEnvelopeEditor v-model="form.key_usage_shell_config" :label="t('admin.settings.runtime.keyUsageShellConfig')" :hint="t('admin.settings.runtime.keyUsageShellConfigHint')" />
            <LocaleEnvelopeEditor v-model="form.dashboard_shell_config" :label="t('admin.settings.runtime.dashboardShellConfig')" :hint="t('admin.settings.runtime.dashboardShellConfigHint')" />
            <LocaleEnvelopeEditor v-model="form.usage_shell_config" :label="t('admin.settings.runtime.usageShellConfig')" :hint="t('admin.settings.runtime.usageShellConfigHint')" />
            <LocaleEnvelopeEditor v-model="form.api_guide_shell_config" :label="t('admin.settings.runtime.apiGuideShellConfig')" :hint="t('admin.settings.runtime.apiGuideShellConfigHint')" />
            <LocaleEnvelopeEditor v-model="form.api_test_shell_config" :label="t('admin.settings.runtime.apiTestShellConfig')" :hint="t('admin.settings.runtime.apiTestShellConfigHint')" />
            <LocaleEnvelopeEditor v-model="form.available_groups_shell_config" :label="t('admin.settings.runtime.availableGroupsShellConfig')" :hint="t('admin.settings.runtime.availableGroupsShellConfigHint')" />
            <LocaleEnvelopeEditor v-model="form.redeem_shell_config" :label="t('admin.settings.runtime.redeemShellConfig')" :hint="t('admin.settings.runtime.redeemShellConfigHint')" />
            <LocaleEnvelopeEditor v-model="form.affiliate_shell_config" :label="t('admin.settings.runtime.affiliateShellConfig')" :hint="t('admin.settings.runtime.affiliateShellConfigHint')" />
            <LocaleEnvelopeEditor v-model="form.available_channels_shell_config" :label="t('admin.settings.runtime.availableChannelsShellConfig')" :hint="t('admin.settings.runtime.availableChannelsShellConfigHint')" />
            <LocaleEnvelopeEditor v-model="form.channel_status_shell_config" :label="t('admin.settings.runtime.channelStatusShellConfig')" :hint="t('admin.settings.runtime.channelStatusShellConfigHint')" />
            <LocaleEnvelopeEditor v-model="form.custom_page_shell_config" :label="t('admin.settings.runtime.customPageShellConfig')" :hint="t('admin.settings.runtime.customPageShellConfigHint')" />
            <LocaleEnvelopeEditor v-model="form.profile_shell_config" :label="t('admin.settings.runtime.profileShellConfig')" :hint="t('admin.settings.runtime.profileShellConfigHint')" />
          </div>
        </section>

        <section class="settings-card xl:col-span-2">
          <SectionHeader :title="t('admin.settings.runtime.promptSection')" />
          <div class="grid gap-4 md:grid-cols-2">
            <LocaleEnvelopeEditor v-model="form.prompt_catalog_shell_config" :label="t('admin.settings.runtime.promptCatalogShellConfig')" :hint="t('admin.settings.runtime.promptCatalogShellConfigHint')" class="md:col-span-2" />
            <LocaleEnvelopeEditor v-model="form.workspace_shell_config" :label="t('admin.settings.runtime.workspaceShellConfig')" :hint="t('admin.settings.runtime.workspaceShellConfigHint')" class="md:col-span-2" />
            <ImagePromptFilterConfigEditor v-model="form.image_prompt_filter_config" :label="t('admin.settings.runtime.imagePromptFilterConfig')" :hint="t('admin.settings.runtime.imagePromptFilterConfigHint')" class="md:col-span-2" />
          </div>
        </section>

        <section class="settings-card">
          <SectionHeader :title="t('admin.settings.runtime.pricingSection')" />
          <div class="space-y-4">
            <TextField v-model="form.pricing_currency_symbol" :label="t('admin.settings.runtime.pricingCurrencySymbol')" :placeholder="t('admin.settings.runtime.pricingCurrencySymbolPlaceholder')" :hint="t('admin.settings.runtime.pricingCurrencySymbolHint')" />
            <LocaleEnvelopeEditor v-model="form.pricing_shell_config" :label="t('admin.settings.runtime.pricingShellConfig')" :hint="t('admin.settings.runtime.pricingShellConfigHint')" />
            <LocaleEnvelopeEditor v-model="form.payment_shell_config" :label="t('admin.settings.runtime.paymentShellConfig')" :hint="t('admin.settings.runtime.paymentShellConfigHint')" />
          </div>
        </section>

        <section class="settings-card">
          <SectionHeader :title="t('admin.settings.runtime.creditsSection')" />
          <div class="space-y-4">
            <TextField v-model="form.credits_per_balance" :label="t('admin.settings.runtime.creditsPerBalance')" :placeholder="t('admin.settings.runtime.creditsPerBalancePlaceholder')" :hint="t('admin.settings.runtime.creditsPerBalanceHint')" />
            <LocaleEnvelopeEditor v-model="form.credits_shell_config" :label="t('admin.settings.runtime.creditsShellConfig')" :hint="t('admin.settings.runtime.creditsShellConfigHint')" />
          </div>
        </section>

        <section class="settings-card">
          <SectionHeader :title="t('admin.settings.runtime.analyticsSection')" />
          <div class="space-y-4">
            <TextField v-model="form.google_analytics_id" :label="t('admin.settings.runtime.googleAnalyticsId')" />
            <TextField v-model="form.clarity_id" :label="t('admin.settings.runtime.clarityId')" />
            <TextField v-model="form.plausible_domain" :label="t('admin.settings.runtime.plausibleDomain')" />
            <TextField v-model="form.plausible_src" :label="t('admin.settings.runtime.plausibleSrc')" />
            <TextField v-model="form.openpanel_client_id" :label="t('admin.settings.runtime.openpanelClientId')" />
            <SwitchField v-model="form.public_integrations_enabled" :label="t('admin.settings.runtime.publicIntegrationsEnabled')" :hint="t('admin.settings.runtime.publicIntegrationsEnabledHint')" />
            <SwitchField v-model="form.vercel_analytics_enabled" :label="t('admin.settings.runtime.vercelAnalyticsEnabled')" />
          </div>
        </section>

        <section class="settings-card">
          <SectionHeader :title="t('admin.settings.runtime.marketingSection')" />
          <div class="space-y-4">
            <TextAreaField v-model="form.adsense_code" :label="t('admin.settings.runtime.adsenseCode')" :placeholder="t('admin.settings.runtime.adsenseCodePlaceholder')" />
            <SwitchField v-model="form.affonso_enabled" :label="t('admin.settings.runtime.affonsoEnabled')" />
            <TextField v-model="form.affonso_id" :label="t('admin.settings.runtime.affonsoId')" />
            <TextField v-model="form.affonso_cookie_duration" :label="t('admin.settings.runtime.affonsoCookieDuration')" />
            <SwitchField v-model="form.promotekit_enabled" :label="t('admin.settings.runtime.promotekitEnabled')" />
            <TextField v-model="form.promotekit_id" :label="t('admin.settings.runtime.promotekitId')" />
          </div>
        </section>

        <section class="settings-card xl:col-span-2">
          <SectionHeader :title="t('admin.settings.runtime.customerSection')" />
          <div class="grid gap-4 md:grid-cols-2">
            <SwitchField v-model="form.crisp_enabled" :label="t('admin.settings.runtime.crispEnabled')" />
            <TextField v-model="form.crisp_website_id" :label="t('admin.settings.runtime.crispWebsiteId')" />
            <SwitchField v-model="form.tawk_enabled" :label="t('admin.settings.runtime.tawkEnabled')" />
            <div class="grid gap-4 md:grid-cols-2">
              <TextField v-model="form.tawk_property_id" :label="t('admin.settings.runtime.tawkPropertyId')" />
              <TextField v-model="form.tawk_widget_id" :label="t('admin.settings.runtime.tawkWidgetId')" />
            </div>
          </div>
        </section>
      </div>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, defineComponent, h, onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import Toggle from '@/components/common/Toggle.vue'
import LocaleEnvelopeEditor from '@/components/common/LocaleEnvelopeEditor.vue'
import ImagePromptFilterConfigEditor from '@/components/admin/ImagePromptFilterConfigEditor.vue'
import Icon from '@/components/icons/Icon.vue'
import { useAppStore } from '@/stores/app'
import adminAPI from '@/api/admin'
import type { RuntimeWorkerStatus, SystemSettings, UpdateSettingsRequest } from '@/api/admin/settings'
import { extractApiErrorMessage } from '@/utils/apiError'

// Shell config fields that contain JSON and need validation
const jsonFields = [
  'model_plaza_shell_config',
  'home_shell_config',
  'home_business_shell_config',
  'docs_shell_config',
  'legal_document_shell_config',
  'api_keys_shell_config',
  'key_usage_shell_config',
  'dashboard_shell_config',
  'usage_shell_config',
  'api_guide_shell_config',
  'api_test_shell_config',
  'available_groups_shell_config',
  'redeem_shell_config',
  'affiliate_shell_config',
  'available_channels_shell_config',
  'channel_status_shell_config',
  'custom_page_shell_config',
  'profile_shell_config',
  'auth_shell_config',
  'prompt_catalog_shell_config',
  'workspace_shell_config',
  'image_prompt_filter_config',
  'pricing_shell_config',
  'payment_shell_config',
  'credits_shell_config',
] as const

type RuntimeSettingsForm = Required<Pick<SystemSettings,
  | 'app_url'
  | 'app_name'
  | 'app_description'
  | 'app_logo'
  | 'app_favicon'
  | 'app_preview_image'
  | 'theme'
  | 'appearance'
  | 'default_locale'
  | 'model_plaza_shell_config'
  | 'home_shell_config'
  | 'home_business_shell_config'
  | 'docs_content_base_path'
  | 'docs_shell_config'
  | 'legal_document_shell_config'
  | 'api_keys_shell_config'
  | 'key_usage_shell_config'
  | 'dashboard_shell_config'
  | 'usage_shell_config'
  | 'api_guide_shell_config'
  | 'api_test_shell_config'
  | 'available_groups_shell_config'
  | 'redeem_shell_config'
  | 'affiliate_shell_config'
  | 'available_channels_shell_config'
  | 'channel_status_shell_config'
  | 'custom_page_shell_config'
  | 'profile_shell_config'
  | 'auth_shell_config'
  | 'prompt_catalog_shell_config'
  | 'workspace_shell_config'
  | 'image_prompt_filter_config'
  | 'pricing_shell_config'
  | 'payment_shell_config'
  | 'pricing_currency_symbol'
  | 'credits_per_balance'
  | 'credits_shell_config'
  | 'locale_detect_enabled'
  | 'email_auth_visible'
  | 'google_auth_visible'
  | 'github_auth_visible'
  | 'google_analytics_id'
  | 'clarity_id'
  | 'plausible_domain'
  | 'plausible_src'
  | 'openpanel_client_id'
  | 'public_integrations_enabled'
  | 'vercel_analytics_enabled'
  | 'adsense_code'
  | 'affonso_enabled'
  | 'affonso_id'
  | 'affonso_cookie_duration'
  | 'promotekit_enabled'
  | 'promotekit_id'
  | 'crisp_enabled'
  | 'crisp_website_id'
  | 'tawk_enabled'
  | 'tawk_property_id'
  | 'tawk_widget_id'
>>

const stringFields = [
  'app_url',
  'app_name',
  'app_description',
  'app_logo',
  'app_favicon',
  'app_preview_image',
  'theme',
  'appearance',
  'default_locale',
  'model_plaza_shell_config',
  'home_shell_config',
  'home_business_shell_config',
  'docs_content_base_path',
  'docs_shell_config',
  'legal_document_shell_config',
  'api_keys_shell_config',
  'key_usage_shell_config',
  'dashboard_shell_config',
  'usage_shell_config',
  'api_guide_shell_config',
  'api_test_shell_config',
  'available_groups_shell_config',
  'redeem_shell_config',
  'affiliate_shell_config',
  'available_channels_shell_config',
  'channel_status_shell_config',
  'custom_page_shell_config',
  'profile_shell_config',
  'auth_shell_config',
  'prompt_catalog_shell_config',
  'workspace_shell_config',
  'image_prompt_filter_config',
  'pricing_shell_config',
  'payment_shell_config',
  'pricing_currency_symbol',
  'credits_per_balance',
  'credits_shell_config',
  'google_analytics_id',
  'clarity_id',
  'plausible_domain',
  'plausible_src',
  'openpanel_client_id',
  'adsense_code',
  'affonso_id',
  'affonso_cookie_duration',
  'promotekit_id',
  'crisp_website_id',
  'tawk_property_id',
  'tawk_widget_id',
] as const

const booleanFields = [
  'locale_detect_enabled',
  'email_auth_visible',
  'google_auth_visible',
  'github_auth_visible',
  'public_integrations_enabled',
  'vercel_analytics_enabled',
  'affonso_enabled',
  'promotekit_enabled',
  'crisp_enabled',
  'tawk_enabled',
] as const

const form = reactive<RuntimeSettingsForm>({
  app_url: '',
  app_name: '',
  app_description: '',
  app_logo: '',
  app_favicon: '',
  app_preview_image: '',
  theme: '',
  appearance: '',
  default_locale: '',
  model_plaza_shell_config: '',
  home_shell_config: '',
  home_business_shell_config: '',
  docs_content_base_path: '',
  docs_shell_config: '',
  legal_document_shell_config: '',
  api_keys_shell_config: '',
  key_usage_shell_config: '',
  dashboard_shell_config: '',
  usage_shell_config: '',
  api_guide_shell_config: '',
  api_test_shell_config: '',
  available_groups_shell_config: '',
  redeem_shell_config: '',
  affiliate_shell_config: '',
  available_channels_shell_config: '',
  channel_status_shell_config: '',
  custom_page_shell_config: '',
  profile_shell_config: '',
  auth_shell_config: '',
  prompt_catalog_shell_config: '',
  workspace_shell_config: '',
  image_prompt_filter_config: '',
  pricing_shell_config: '',
  payment_shell_config: '',
  pricing_currency_symbol: '',
  credits_per_balance: '',
  credits_shell_config: '',
  locale_detect_enabled: false,
  email_auth_visible: true,
  google_auth_visible: true,
  github_auth_visible: true,
  google_analytics_id: '',
  clarity_id: '',
  plausible_domain: '',
  plausible_src: '',
  openpanel_client_id: '',
  public_integrations_enabled: true,
  vercel_analytics_enabled: false,
  adsense_code: '',
  affonso_enabled: false,
  affonso_id: '',
  affonso_cookie_duration: '',
  promotekit_enabled: false,
  promotekit_id: '',
  crisp_enabled: false,
  crisp_website_id: '',
  tawk_enabled: false,
  tawk_property_id: '',
  tawk_widget_id: '',
})

const { t: translate, tm } = useI18n()
const appStore = useAppStore()
const loading = ref(false)
const saving = ref(false)
const loadingWorkers = ref(false)
const workerStatuses = ref<RuntimeWorkerStatus[]>([])
const workerStatusError = ref('')
const jsonValidationErrors = reactive<Record<string, string | null>>({})

function t(key: string): string {
  if (key.startsWith('admin.settings.runtime.')) {
    const raw = tm(key)
    if (typeof raw === 'string') {
      return raw
    }
  }
  return translate(key)
}

const hasValidationErrors = computed(() => {
  return Object.values(jsonValidationErrors).some(v => v !== null)
})

function applySettings(settings: Partial<SystemSettings>) {
  for (const field of stringFields) {
    const value = settings[field]
    form[field] = typeof value === 'string' ? value : ''
  }
  for (const field of booleanFields) {
    const value = settings[field]
    form[field] = typeof value === 'boolean' ? value : false
  }
}

function buildPayload(): UpdateSettingsRequest {
  const payload: UpdateSettingsRequest = {}
  for (const field of stringFields) {
    payload[field] = form[field]
  }
  for (const field of booleanFields) {
    payload[field] = form[field]
  }
  return payload
}

/** Validate all JSON fields and return true if all are valid */
function validateJsonFields(): boolean {
  let allValid = true
  for (const field of jsonFields) {
    const value = form[field]
    if (!value || !value.trim()) {
      jsonValidationErrors[field] = null
      continue
    }
    try {
      JSON.parse(value)
      jsonValidationErrors[field] = null
    } catch {
      jsonValidationErrors[field] = `Invalid JSON in ${field}`
      allValid = false
    }
  }
  return allValid
}

/** Pretty-print all valid JSON shell config fields */
function formatAllJson() {
  for (const field of jsonFields) {
    const value = form[field]
    if (!value || !value.trim()) continue
    try {
      const parsed = JSON.parse(value)
      const formatted = JSON.stringify(parsed, null, 2)
      if (formatted !== value) {
        form[field] = formatted
      }
    } catch {
      // Skip invalid JSON
    }
  }
  appStore.showSuccess(t('admin.settings.runtime.formatAllJsonSuccess'))
}

async function loadSettings() {
  loading.value = true
  try {
    const settings = await adminAPI.settings.getSettings()
    applySettings(settings)
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('admin.settings.runtime.loadFailed')))
  } finally {
    loading.value = false
  }
}

async function loadWorkerStatuses() {
  loadingWorkers.value = true
  workerStatusError.value = ''
  try {
    const result = await adminAPI.settings.getRuntimeWorkers()
    workerStatuses.value = Array.isArray(result.workers) ? result.workers : []
  } catch (error) {
    workerStatuses.value = []
    workerStatusError.value = extractApiErrorMessage(error, 'Worker 状态加载失败')
  } finally {
    loadingWorkers.value = false
  }
}

async function saveSettings() {
  if (!validateJsonFields()) {
    appStore.showError(t('admin.settings.runtime.jsonValidationError'))
    return
  }
  saving.value = true
  try {
    const updated = await adminAPI.settings.updateSettings(buildPayload())
    applySettings(updated)
    await appStore.fetchPublicSettings(true)
    appStore.showSuccess(t('admin.settings.runtime.saveSuccess'))
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('admin.settings.runtime.saveFailed')))
  } finally {
    saving.value = false
  }
}

onMounted(() => {
  void loadSettings()
  void loadWorkerStatuses()
})

function workerHealthLabel(health: string): string {
  switch ((health || '').toLowerCase()) {
    case 'active':
      return '运行中'
    case 'idle':
      return '空闲'
    case 'waiting':
      return '等待'
    case 'attention':
      return '需处理'
    case 'not_configured':
      return '未配置'
    default:
      return '未知'
  }
}

function workerStatusMessage(health: string): string {
  switch ((health || '').toLowerCase()) {
    case 'active':
      return 'Worker 正在处理或最近有任务活动。'
    case 'idle':
      return 'Worker 当前无待处理任务。'
    case 'waiting':
      return '队列有任务等待 worker 消费。'
    case 'attention':
      return '存在失败、卡死或配置异常，需要检查。'
    case 'not_configured':
      return '运行时配置不完整。'
    default:
      return '状态暂不可判断。'
  }
}

function workerHealthClass(health: string): string {
  const base = 'shrink-0 rounded-full px-2.5 py-1 text-xs font-bold'
  switch ((health || '').toLowerCase()) {
    case 'active':
      return `${base} bg-emerald-100 text-emerald-700 dark:bg-emerald-500/15 dark:text-emerald-300`
    case 'idle':
      return `${base} bg-sky-100 text-sky-700 dark:bg-sky-500/15 dark:text-sky-300`
    case 'waiting':
      return `${base} bg-amber-100 text-amber-700 dark:bg-amber-500/15 dark:text-amber-300`
    case 'attention':
      return `${base} bg-red-100 text-red-700 dark:bg-red-500/15 dark:text-red-300`
    case 'not_configured':
      return `${base} bg-gray-200 text-gray-700 dark:bg-dark-700 dark:text-dark-200`
    default:
      return `${base} bg-gray-100 text-gray-600 dark:bg-dark-700 dark:text-dark-300`
  }
}

function formatWorkerNumber(value: number | null | undefined): string {
  return typeof value === 'number' && Number.isFinite(value) ? String(value) : '-'
}

function formatWorkerTime(value: string | null | undefined): string {
  if (!value) return '-'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value
  return date.toLocaleString()
}

const SectionHeader = defineComponent({
  props: {
    title: { type: String, required: true },
  },
  setup(props) {
    return () => h('div', { class: 'mb-4 border-b border-gray-100 pb-3 dark:border-dark-700' }, [
      h('h2', { class: 'text-base font-bold text-gray-950 dark:text-white' }, props.title),
    ])
  },
})

const TextField = defineComponent({
  props: {
    modelValue: { type: String, required: true },
    label: { type: String, required: true },
    placeholder: { type: String, default: '' },
    hint: { type: String, default: '' },
  },
  emits: ['update:modelValue'],
  setup(props, { emit, attrs }) {
    const inputClass = computed(() => ['space-y-1.5', attrs.class])
    return () => h('label', { class: inputClass.value }, [
      h('span', { class: 'block text-sm font-medium text-gray-700 dark:text-dark-200' }, props.label),
      h('input', {
        value: props.modelValue,
        type: 'text',
        placeholder: props.placeholder,
        class: 'input',
        onInput: (event: Event) => emit('update:modelValue', (event.target as HTMLInputElement).value),
      }),
      props.hint ? h('p', { class: 'text-xs leading-5 text-gray-500 dark:text-dark-300' }, props.hint) : null,
    ])
  },
})

// TextAreaField is still used for adsense_code (non-JSON content)
const TextAreaField = defineComponent({
  props: {
    modelValue: { type: String, required: true },
    label: { type: String, required: true },
    placeholder: { type: String, default: '' },
    hint: { type: String, default: '' },
  },
  emits: ['update:modelValue'],
  setup(props, { emit, attrs }) {
    const inputClass = computed(() => ['space-y-1.5', attrs.class])
    return () => h('label', { class: inputClass.value }, [
      h('span', { class: 'block text-sm font-medium text-gray-700 dark:text-dark-200' }, props.label),
      h('textarea', {
        value: props.modelValue,
        rows: 4,
        placeholder: props.placeholder,
        class: 'input min-h-28 font-mono text-xs leading-5',
        onInput: (event: Event) => emit('update:modelValue', (event.target as HTMLTextAreaElement).value),
      }),
      props.hint ? h('p', { class: 'text-xs leading-5 text-gray-500 dark:text-dark-300' }, props.hint) : null,
    ])
  },
})

const SwitchField = defineComponent({
  props: {
    modelValue: { type: Boolean, required: true },
    label: { type: String, required: true },
    hint: { type: String, default: '' },
  },
  emits: ['update:modelValue'],
  setup(props, { emit }) {
    return () => h('div', { class: 'flex items-start justify-between gap-3 rounded-2xl bg-gray-50 p-3 dark:bg-dark-800/70' }, [
      h('div', [
        h('p', { class: 'text-sm font-medium text-gray-800 dark:text-dark-200' }, props.label),
        props.hint ? h('p', { class: 'mt-1 text-xs leading-5 text-gray-500 dark:text-dark-300' }, props.hint) : null,
      ]),
      h(Toggle, {
        modelValue: props.modelValue,
        'onUpdate:modelValue': (value: boolean) => emit('update:modelValue', value),
      }),
    ])
  },
})
</script>

<style scoped>
.settings-card {
  @apply rounded-2xl border border-gray-200 bg-white p-6 shadow-sm dark:border-dark-700 dark:bg-dark-900;
}
</style>
