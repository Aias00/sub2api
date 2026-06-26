<template>
  <AppLayout>
    <div class="space-y-5">
      <div class="overflow-hidden rounded-3xl border border-gray-200 bg-white shadow-sm dark:border-dark-700 dark:bg-dark-900">
        <div class="flex flex-col gap-4 p-5 lg:flex-row lg:items-center lg:justify-between">
          <div>
            <p class="text-xs font-semibold uppercase tracking-[0.24em] text-primary-500">
              {{ t('admin.settings.runtime.badge') }}
            </p>
            <h1 class="mt-2 text-2xl font-black tracking-tight text-gray-950 dark:text-white">
              {{ t('admin.settings.runtime.title') }}
            </h1>
            <p class="mt-2 max-w-3xl text-sm leading-6 text-gray-500 dark:text-gray-400">
              {{ t('admin.settings.runtime.description') }}
            </p>
          </div>
          <div class="flex flex-wrap items-center gap-2">
            <button class="btn btn-secondary" type="button" :disabled="loading" @click="loadSettings">
              <Icon name="refresh" size="md" :class="loading ? 'animate-spin' : ''" />
              {{ t('common.refresh') }}
            </button>
            <button class="btn btn-primary" type="button" :disabled="loading || saving" @click="saveSettings">
              <Icon name="check" size="md" />
              {{ saving ? t('common.saving') : t('common.save') }}
            </button>
          </div>
        </div>
      </div>

      <div v-if="loading" class="rounded-2xl border border-gray-200 bg-white p-8 text-center text-sm text-gray-500 shadow-sm dark:border-dark-700 dark:bg-dark-900 dark:text-gray-400">
        {{ t('common.loading') }}
      </div>

      <div v-else class="grid gap-5 xl:grid-cols-[1fr_1fr]">
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
            <TextAreaField v-model="form.auth_shell_config" :label="t('admin.settings.runtime.authShellConfig')" :placeholder="t('admin.settings.runtime.authShellConfigPlaceholder')" :hint="t('admin.settings.runtime.authShellConfigHint')" class="md:col-span-3" />
          </div>
        </section>

        <section class="settings-card xl:col-span-2">
          <SectionHeader :title="t('admin.settings.runtime.modelPlazaSection')" />
          <div class="space-y-4">
            <TextAreaField v-model="form.model_plaza_shell_config" :label="t('admin.settings.runtime.modelPlazaShellConfig')" :placeholder="t('admin.settings.runtime.modelPlazaShellConfigPlaceholder')" :hint="t('admin.settings.runtime.modelPlazaShellConfigHint')" />
          </div>
        </section>

        <section class="settings-card xl:col-span-2">
          <SectionHeader :title="t('admin.settings.runtime.homeSection')" />
          <div class="space-y-4">
            <TextAreaField v-model="form.home_shell_config" :label="t('admin.settings.runtime.homeShellConfig')" :placeholder="t('admin.settings.runtime.homeShellConfigPlaceholder')" :hint="t('admin.settings.runtime.homeShellConfigHint')" />
            <TextAreaField v-model="form.home_business_shell_config" :label="t('admin.settings.runtime.homeBusinessShellConfig')" :placeholder="t('admin.settings.runtime.homeBusinessShellConfigPlaceholder')" :hint="t('admin.settings.runtime.homeBusinessShellConfigHint')" />
          </div>
        </section>

        <section class="settings-card xl:col-span-2">
          <SectionHeader :title="t('admin.settings.runtime.docsSection')" />
          <div class="space-y-4">
            <TextAreaField v-model="form.docs_content_base_path" :label="t('admin.settings.runtime.docsContentBasePath')" :placeholder="t('admin.settings.runtime.docsContentBasePathPlaceholder')" :hint="t('admin.settings.runtime.docsContentBasePathHint')" />
            <TextAreaField v-model="form.docs_shell_config" :label="t('admin.settings.runtime.docsShellConfig')" :placeholder="t('admin.settings.runtime.docsShellConfigPlaceholder')" :hint="t('admin.settings.runtime.docsShellConfigHint')" />
            <TextAreaField v-model="form.legal_document_shell_config" :label="t('admin.settings.runtime.legalDocumentShellConfig')" :placeholder="t('admin.settings.runtime.legalDocumentShellConfigPlaceholder')" :hint="t('admin.settings.runtime.legalDocumentShellConfigHint')" />
            <TextAreaField v-model="form.api_keys_shell_config" :label="t('admin.settings.runtime.apiKeysShellConfig')" :placeholder="t('admin.settings.runtime.apiKeysShellConfigPlaceholder')" :hint="t('admin.settings.runtime.apiKeysShellConfigHint')" />
            <TextAreaField v-model="form.key_usage_shell_config" :label="t('admin.settings.runtime.keyUsageShellConfig')" :placeholder="t('admin.settings.runtime.keyUsageShellConfigPlaceholder')" :hint="t('admin.settings.runtime.keyUsageShellConfigHint')" />
            <TextAreaField v-model="form.dashboard_shell_config" :label="t('admin.settings.runtime.dashboardShellConfig')" :placeholder="t('admin.settings.runtime.dashboardShellConfigPlaceholder')" :hint="t('admin.settings.runtime.dashboardShellConfigHint')" />
            <TextAreaField v-model="form.usage_shell_config" :label="t('admin.settings.runtime.usageShellConfig')" :placeholder="t('admin.settings.runtime.usageShellConfigPlaceholder')" :hint="t('admin.settings.runtime.usageShellConfigHint')" />
            <TextAreaField v-model="form.api_guide_shell_config" :label="t('admin.settings.runtime.apiGuideShellConfig')" :placeholder="t('admin.settings.runtime.apiGuideShellConfigPlaceholder')" :hint="t('admin.settings.runtime.apiGuideShellConfigHint')" />
            <TextAreaField v-model="form.api_test_shell_config" :label="t('admin.settings.runtime.apiTestShellConfig')" :placeholder="t('admin.settings.runtime.apiTestShellConfigPlaceholder')" :hint="t('admin.settings.runtime.apiTestShellConfigHint')" />
            <TextAreaField v-model="form.available_groups_shell_config" :label="t('admin.settings.runtime.availableGroupsShellConfig')" :placeholder="t('admin.settings.runtime.availableGroupsShellConfigPlaceholder')" :hint="t('admin.settings.runtime.availableGroupsShellConfigHint')" />
            <TextAreaField v-model="form.redeem_shell_config" :label="t('admin.settings.runtime.redeemShellConfig')" :placeholder="t('admin.settings.runtime.redeemShellConfigPlaceholder')" :hint="t('admin.settings.runtime.redeemShellConfigHint')" />
            <TextAreaField v-model="form.affiliate_shell_config" :label="t('admin.settings.runtime.affiliateShellConfig')" :placeholder="t('admin.settings.runtime.affiliateShellConfigPlaceholder')" :hint="t('admin.settings.runtime.affiliateShellConfigHint')" />
            <TextAreaField v-model="form.available_channels_shell_config" :label="t('admin.settings.runtime.availableChannelsShellConfig')" :placeholder="t('admin.settings.runtime.availableChannelsShellConfigPlaceholder')" :hint="t('admin.settings.runtime.availableChannelsShellConfigHint')" />
            <TextAreaField v-model="form.channel_status_shell_config" :label="t('admin.settings.runtime.channelStatusShellConfig')" :placeholder="t('admin.settings.runtime.channelStatusShellConfigPlaceholder')" :hint="t('admin.settings.runtime.channelStatusShellConfigHint')" />
            <TextAreaField v-model="form.custom_page_shell_config" :label="t('admin.settings.runtime.customPageShellConfig')" :placeholder="t('admin.settings.runtime.customPageShellConfigPlaceholder')" :hint="t('admin.settings.runtime.customPageShellConfigHint')" />
            <TextAreaField v-model="form.profile_shell_config" :label="t('admin.settings.runtime.profileShellConfig')" :placeholder="t('admin.settings.runtime.profileShellConfigPlaceholder')" :hint="t('admin.settings.runtime.profileShellConfigHint')" />
          </div>
        </section>

        <section class="settings-card xl:col-span-2">
          <SectionHeader :title="t('admin.settings.runtime.promptSection')" />
          <div class="grid gap-4 md:grid-cols-2">
            <TextAreaField v-model="form.prompt_catalog_shell_config" :label="t('admin.settings.runtime.promptCatalogShellConfig')" :placeholder="t('admin.settings.runtime.promptCatalogShellConfigPlaceholder')" :hint="t('admin.settings.runtime.promptCatalogShellConfigHint')" class="md:col-span-2" />
            <TextAreaField v-model="form.workspace_shell_config" :label="t('admin.settings.runtime.workspaceShellConfig')" :placeholder="t('admin.settings.runtime.workspaceShellConfigPlaceholder')" :hint="t('admin.settings.runtime.workspaceShellConfigHint')" class="md:col-span-2" />
          </div>
        </section>

        <section class="settings-card">
          <SectionHeader :title="t('admin.settings.runtime.pricingSection')" />
          <div class="space-y-4">
            <TextField v-model="form.pricing_currency_symbol" :label="t('admin.settings.runtime.pricingCurrencySymbol')" :placeholder="t('admin.settings.runtime.pricingCurrencySymbolPlaceholder')" :hint="t('admin.settings.runtime.pricingCurrencySymbolHint')" />
            <TextAreaField v-model="form.pricing_shell_config" :label="t('admin.settings.runtime.pricingShellConfig')" :placeholder="t('admin.settings.runtime.pricingShellConfigPlaceholder')" :hint="t('admin.settings.runtime.pricingShellConfigHint')" />
            <TextAreaField v-model="form.payment_shell_config" :label="t('admin.settings.runtime.paymentShellConfig')" :placeholder="t('admin.settings.runtime.paymentShellConfigPlaceholder')" :hint="t('admin.settings.runtime.paymentShellConfigHint')" />
          </div>
        </section>

        <section class="settings-card">
          <SectionHeader :title="t('admin.settings.runtime.creditsSection')" />
          <div class="space-y-4">
            <TextField v-model="form.credits_per_balance" :label="t('admin.settings.runtime.creditsPerBalance')" :placeholder="t('admin.settings.runtime.creditsPerBalancePlaceholder')" :hint="t('admin.settings.runtime.creditsPerBalanceHint')" />
            <TextAreaField v-model="form.credits_shell_config" :label="t('admin.settings.runtime.creditsShellConfig')" :placeholder="t('admin.settings.runtime.creditsShellConfigPlaceholder')" :hint="t('admin.settings.runtime.creditsShellConfigHint')" />
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
import Icon from '@/components/icons/Icon.vue'
import { useAppStore } from '@/stores/app'
import adminAPI from '@/api/admin'
import type { SystemSettings, UpdateSettingsRequest } from '@/api/admin/settings'
import { extractApiErrorMessage } from '@/utils/apiError'

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

const { t } = useI18n()
const appStore = useAppStore()
const loading = ref(false)
const saving = ref(false)

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

async function saveSettings() {
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
})

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
      h('span', { class: 'block text-sm font-medium text-gray-700 dark:text-gray-300' }, props.label),
      h('input', {
        value: props.modelValue,
        type: 'text',
        placeholder: props.placeholder,
        class: 'input',
        onInput: (event: Event) => emit('update:modelValue', (event.target as HTMLInputElement).value),
      }),
      props.hint ? h('p', { class: 'text-xs leading-5 text-gray-500 dark:text-gray-400' }, props.hint) : null,
    ])
  },
})

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
      h('span', { class: 'block text-sm font-medium text-gray-700 dark:text-gray-300' }, props.label),
      h('textarea', {
        value: props.modelValue,
        rows: 4,
        placeholder: props.placeholder,
        class: 'input min-h-28 font-mono text-xs leading-5',
        onInput: (event: Event) => emit('update:modelValue', (event.target as HTMLTextAreaElement).value),
      }),
      props.hint ? h('p', { class: 'text-xs leading-5 text-gray-500 dark:text-gray-400' }, props.hint) : null,
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
        h('p', { class: 'text-sm font-medium text-gray-800 dark:text-gray-200' }, props.label),
        props.hint ? h('p', { class: 'mt-1 text-xs leading-5 text-gray-500 dark:text-gray-400' }, props.hint) : null,
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
  @apply rounded-3xl border border-gray-200 bg-white p-5 shadow-sm dark:border-dark-700 dark:bg-dark-900;
}
</style>
