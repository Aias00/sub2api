<template>
  <section class="space-y-5">
    <div class="overflow-hidden rounded-3xl border border-gray-200 bg-white shadow-sm dark:border-dark-700 dark:bg-dark-900">
      <div class="flex flex-col gap-4 border-b border-gray-100 p-5 dark:border-dark-700 md:flex-row md:items-center md:justify-between">
        <div>
          <p class="text-xs font-semibold uppercase tracking-[0.24em] text-amber-500">{{ localText('余额充值', 'Balance top-up') }}</p>
          <h2 class="mt-2 text-xl font-bold text-gray-950 dark:text-white">{{ localText('充值商品', 'Recharge products') }}</h2>
          <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
            {{ localText('配置一个商品，用户端充值页就展示一个卡片。实际到账额度由充值倍率统一计算。', 'Each product renders one card on the recharge tab. Credited balance is calculated from the global recharge multiplier.') }}
          </p>
        </div>
        <div class="flex flex-wrap items-center gap-2">
          <button type="button" class="btn btn-secondary" :disabled="loading" @click="loadProducts">
            <Icon name="refresh" size="sm" :class="loading ? 'animate-spin' : ''" />
            {{ localText('刷新', 'Refresh') }}
          </button>
          <button type="button" class="btn btn-primary" @click="addProduct">
            <Icon name="plus" size="sm" />
            {{ localText('添加充值商品', 'Add product') }}
          </button>
        </div>
      </div>

      <div class="grid gap-3 p-5 sm:grid-cols-3">
        <div class="rounded-2xl bg-gray-50 p-4 dark:bg-dark-800/70">
          <p class="text-xs text-gray-500 dark:text-gray-400">{{ localText('商品数量', 'Products') }}</p>
          <p class="mt-1 text-2xl font-black text-gray-950 dark:text-white">{{ products.length }}</p>
        </div>
        <div class="rounded-2xl bg-gray-50 p-4 dark:bg-dark-800/70">
          <p class="text-xs text-gray-500 dark:text-gray-400">{{ localText('推荐卡片', 'Featured') }}</p>
          <p class="mt-1 text-2xl font-black text-gray-950 dark:text-white">{{ featuredCount }}</p>
        </div>
        <div class="rounded-2xl bg-gray-50 p-4 dark:bg-dark-800/70">
          <p class="text-xs text-gray-500 dark:text-gray-400">{{ localText('最高面额', 'Highest amount') }}</p>
          <p class="mt-1 text-2xl font-black text-gray-950 dark:text-white">¥{{ highestAmount.toFixed(2) }}</p>
        </div>
      </div>
    </div>

    <div v-if="loading" class="rounded-3xl border border-gray-200 bg-white p-8 text-center dark:border-dark-700 dark:bg-dark-900">
      <div class="mx-auto h-8 w-8 animate-spin rounded-full border-4 border-primary-500 border-t-transparent"></div>
      <p class="mt-3 text-sm text-gray-500 dark:text-gray-400">{{ localText('正在加载充值商品...', 'Loading recharge products...') }}</p>
    </div>

    <div v-else-if="products.length === 0" class="rounded-3xl border border-dashed border-gray-300 bg-white p-10 text-center dark:border-dark-700 dark:bg-dark-900">
      <Icon name="gift" size="xl" class="mx-auto mb-3 text-gray-300 dark:text-dark-600" />
      <p class="text-lg font-semibold text-gray-900 dark:text-white">{{ localText('还没有充值商品', 'No recharge products') }}</p>
      <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">{{ localText('添加后用户端充值页会立即按排序展示。', 'Add products and they will appear on the recharge tab in sort order.') }}</p>
      <button type="button" class="btn btn-primary mt-5" @click="addProduct">{{ localText('创建第一个商品', 'Create first product') }}</button>
    </div>

    <div v-else class="space-y-4">
      <article
        v-for="(product, index) in products"
        :key="product.id"
        class="overflow-hidden rounded-3xl border border-gray-200 bg-white shadow-sm dark:border-dark-700 dark:bg-dark-900"
      >
        <div class="flex flex-col gap-3 border-b border-gray-100 p-5 dark:border-dark-700 md:flex-row md:items-center md:justify-between">
          <div class="min-w-0">
            <div class="flex flex-wrap items-center gap-2">
              <h3 class="truncate text-base font-bold text-gray-950 dark:text-white">
                {{ product.name || localText('未命名商品', 'Untitled product') }}
              </h3>
              <span v-if="product.recommended" class="rounded-full bg-amber-100 px-2 py-0.5 text-xs font-semibold text-amber-700 dark:bg-amber-500/15 dark:text-amber-200">
                {{ localText('推荐', 'Featured') }}
              </span>
            </div>
            <p class="mt-1 truncate text-xs text-gray-400">{{ product.id }}</p>
          </div>
          <button type="button" class="btn btn-secondary btn-sm text-red-600 hover:text-red-700 dark:text-red-400" @click="removeProduct(index)">
            <Icon name="trash" size="sm" />
            {{ localText('删除', 'Delete') }}
          </button>
        </div>

        <div class="grid gap-5 p-5 lg:grid-cols-[1fr_260px]">
          <div class="space-y-4">
            <div class="grid gap-3 md:grid-cols-2">
              <div>
                <label class="input-label">{{ localText('商品名称', 'Product name') }}</label>
                <input v-model="product.name" type="text" class="input" :placeholder="localText('例如：体验', 'e.g. Starter')" />
              </div>
              <div>
                <label class="input-label">{{ localText('副标题', 'Subtitle') }}</label>
                <input v-model="product.description" type="text" class="input" :placeholder="localText('例如：适合初次体验', 'e.g. For first-time users')" />
              </div>
            </div>

            <div class="grid gap-3 md:grid-cols-4">
              <div>
                <label class="input-label">{{ localText('金额（CNY）', 'Amount (CNY)') }}</label>
                <input v-model.number="product.amount" type="number" min="0" step="0.01" class="input" />
              </div>
              <div>
                <label class="input-label">{{ localText('Creem Product ID', 'Creem Product ID') }}</label>
                <input v-model="product.creem_product_id" type="text" class="input" :placeholder="localText('可选，仅 Creem 使用', 'Optional, only for Creem')" />
              </div>
              <div>
                <label class="input-label">{{ localText('角标', 'Badge') }}</label>
                <input v-model="product.badge" type="text" class="input" :placeholder="localText('推荐', 'Recommended')" />
              </div>
              <div>
                <label class="input-label">{{ localText('排序', 'Sort order') }}</label>
                <input v-model.number="product.sort_order" type="number" min="0" step="1" class="input" />
              </div>
              <div>
                <label class="input-label">{{ localText('推荐卡', 'Featured') }}</label>
                <button
                  type="button"
                  :class="[
                    'mt-1 inline-flex h-10 w-full items-center justify-center rounded-xl border text-sm font-semibold transition-colors',
                    product.recommended
                      ? 'border-amber-300 bg-amber-100 text-amber-700 dark:border-amber-400/40 dark:bg-amber-500/15 dark:text-amber-200'
                      : 'border-gray-200 bg-white text-gray-500 hover:bg-gray-50 dark:border-dark-600 dark:bg-dark-800 dark:text-gray-300'
                  ]"
                  @click="product.recommended = !product.recommended"
                >
                  {{ product.recommended ? localText('已推荐', 'Featured') : localText('普通', 'Standard') }}
                </button>
              </div>
            </div>

            <div>
              <label class="input-label">{{ localText('卖点列表', 'Feature list') }}</label>
              <textarea
                :value="product.features.join('\n')"
                rows="3"
                class="input"
                :placeholder="localText('每行一个卖点', 'One feature per line')"
                @input="updateProductFeatures(index, ($event.target as HTMLTextAreaElement).value)"
              ></textarea>
            </div>
          </div>

          <div class="rounded-3xl border border-gray-100 bg-gray-50 p-4 dark:border-dark-700 dark:bg-dark-800/60">
            <p class="text-xs font-semibold uppercase tracking-[0.2em] text-gray-400">{{ localText('用户端预览', 'User preview') }}</p>
            <div class="mt-4 rounded-3xl border border-gray-200 bg-white p-4 shadow-sm dark:border-dark-700 dark:bg-dark-900">
              <div v-if="product.recommended || product.badge" class="-mx-4 -mt-4 mb-4 rounded-t-3xl bg-amber-500 px-4 py-2 text-center text-xs font-bold tracking-[0.18em] text-white">
                {{ product.badge || localText('推荐', 'Recommended') }}
              </div>
              <h4 class="text-lg font-black text-gray-950 dark:text-white">{{ product.name || localText('商品名称', 'Product name') }}</h4>
              <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">{{ product.description || localText('副标题说明', 'Subtitle') }}</p>
              <div class="mt-4 flex items-end gap-1 text-gray-950 dark:text-white">
                <span class="text-sm text-gray-400">¥</span>
                <span class="text-3xl font-black">{{ Number(product.amount || 0).toFixed(2) }}</span>
              </div>
              <p class="mt-2 text-xs text-gray-500 dark:text-gray-400">
                {{ localText('到账额度由倍率自动计算', 'Credited balance is calculated automatically') }}
              </p>
              <div class="mt-4 space-y-1.5">
                <div v-for="feature in product.features.slice(0, 3)" :key="feature" class="flex gap-2 text-xs text-gray-600 dark:text-gray-300">
                  <span class="text-amber-500">✓</span>
                  <span>{{ feature }}</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </article>
    </div>

    <div class="sticky bottom-4 z-10 flex flex-col gap-3 rounded-3xl border border-gray-200 bg-white/90 p-4 shadow-xl backdrop-blur dark:border-dark-700 dark:bg-dark-900/90 sm:flex-row sm:items-center sm:justify-between">
      <div>
        <p class="text-sm font-semibold text-gray-950 dark:text-white">
          {{ dirty ? localText('有未保存的商品修改', 'Unsaved product changes') : localText('商品配置已同步', 'Product catalog is in sync') }}
        </p>
        <p class="text-xs text-gray-500 dark:text-gray-400">{{ localText('保存后会覆盖当前充值商品列表。', 'Saving overwrites the current recharge product list.') }}</p>
      </div>
      <button type="button" class="btn btn-primary" :disabled="saving || !dirty" @click="saveProducts">
        <span v-if="saving" class="h-4 w-4 animate-spin rounded-full border-2 border-white border-t-transparent"></span>
        {{ saving ? localText('保存中...', 'Saving...') : localText('保存充值商品', 'Save products') }}
      </button>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import settingsAPI from '@/api/admin/settings'
import { useAppStore } from '@/stores/app'
import { extractApiErrorMessage } from '@/utils/apiError'
import type { RechargeProduct } from '@/types/payment'
import Icon from '@/components/icons/Icon.vue'

const { locale } = useI18n()
const appStore = useAppStore()

const products = ref<RechargeProduct[]>([])
const originalSnapshot = ref('[]')
const loading = ref(false)
const saving = ref(false)

const localText = (zh: string, en: string) => locale.value.startsWith('zh') ? zh : en

const featuredCount = computed(() => products.value.filter((product) => product.recommended).length)
const highestAmount = computed(() => products.value.reduce((max, product) => Math.max(max, Number(product.amount) || 0), 0))
const dirty = computed(() => snapshot(products.value) !== originalSnapshot.value)

function snapshot(input: RechargeProduct[]) {
  return JSON.stringify(normalizeProducts(input))
}

function normalizeProducts(input: RechargeProduct[] | null | undefined): RechargeProduct[] {
  if (!Array.isArray(input)) return []
  return input
    .map((product, index) => ({
      id: product.id || `recharge-${index + 1}`,
      name: product.name || '',
      description: product.description || '',
      amount: Number(product.amount) || 0,
      credited_amount: Number(product.credited_amount) || 0,
      creem_product_id: product.creem_product_id || '',
      badge: product.badge || '',
      recommended: Boolean(product.recommended),
      features: Array.isArray(product.features) ? product.features.filter(Boolean) : [],
      sort_order: Number(product.sort_order) || (index + 1) * 10,
    }))
    .sort((a, b) => {
      if (a.sort_order === b.sort_order) return a.name.localeCompare(b.name)
      return a.sort_order - b.sort_order
    })
}

function createProductDraft(): RechargeProduct {
  const nextSortOrder = products.value.length > 0
    ? Math.max(...products.value.map((item) => item.sort_order || 0)) + 10
    : 10
  return {
    id: `recharge-${Date.now()}-${Math.random().toString(36).slice(2, 8)}`,
    name: '',
    description: '',
    amount: 0,
    credited_amount: 0,
    creem_product_id: '',
    badge: '',
    recommended: false,
    features: [],
    sort_order: nextSortOrder,
  }
}

function addProduct() {
  products.value.push(createProductDraft())
}

function removeProduct(index: number) {
  products.value.splice(index, 1)
}

function updateProductFeatures(index: number, raw: string) {
  products.value[index].features = raw
    .split('\n')
    .map((line) => line.trim())
    .filter(Boolean)
}

async function loadProducts() {
  loading.value = true
  try {
    const settings = await settingsAPI.getSettings()
    products.value = normalizeProducts(settings.payment_recharge_products)
    originalSnapshot.value = snapshot(products.value)
  } catch (error: unknown) {
    appStore.showError(extractApiErrorMessage(error, localText('加载充值商品失败', 'Failed to load recharge products')))
  } finally {
    loading.value = false
  }
}

async function saveProducts() {
  saving.value = true
  try {
    const normalized = normalizeProducts(products.value)
    const settings = await settingsAPI.updateSettings({ payment_recharge_products: normalized })
    products.value = normalizeProducts(settings.payment_recharge_products)
    originalSnapshot.value = snapshot(products.value)
    appStore.showSuccess(localText('充值商品已保存', 'Recharge products saved'))
  } catch (error: unknown) {
    appStore.showError(extractApiErrorMessage(error, localText('保存充值商品失败', 'Failed to save recharge products')))
  } finally {
    saving.value = false
  }
}

onMounted(loadProducts)
</script>
