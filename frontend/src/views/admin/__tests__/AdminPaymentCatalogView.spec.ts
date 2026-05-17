import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'

import { flushPromises, mount } from '@vue/test-utils'
import { defineComponent } from 'vue'
import { describe, expect, it, vi } from 'vitest'

import AdminPaymentPlansView from '../orders/AdminPaymentPlansView.vue'

const { getPlans, updatePlan, deletePlan, getGroups, showError, showSuccess } = vi.hoisted(() => ({
  getPlans: vi.fn(),
  updatePlan: vi.fn(),
  deletePlan: vi.fn(),
  getGroups: vi.fn(),
  showError: vi.fn(),
  showSuccess: vi.fn(),
}))

const localeRef = vi.hoisted(() => ({ value: 'zh-CN' }))

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => key,
      locale: localeRef,
    }),
  }
})

vi.mock('@/api/admin/payment', () => ({
  adminPaymentAPI: {
    getPlans,
    updatePlan,
    deletePlan,
  },
}))

vi.mock('@/api/admin', () => ({
  default: {
    groups: {
      getAll: getGroups,
    },
  },
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({
    showError,
    showSuccess,
  }),
}))

vi.mock('@/utils/apiError', () => ({
  extractI18nErrorMessage: () => 'error',
}))

const plansViewSource = readFileSync(resolve(process.cwd(), 'src/views/admin/orders/AdminPaymentPlansView.vue'), 'utf8')
const rechargeManagerSource = readFileSync(resolve(process.cwd(), 'src/views/admin/orders/RechargeProductsManager.vue'), 'utf8')
const zhLocaleSource = readFileSync(resolve(process.cwd(), 'src/i18n/locales/zh.ts'), 'utf8')
const enLocaleSource = readFileSync(resolve(process.cwd(), 'src/i18n/locales/en.ts'), 'utf8')

const AppLayoutStub = defineComponent({ template: '<main><slot /></main>' })
const IconStub = defineComponent({
  props: ['name'],
  template: '<span class="icon-stub">{{ name }}</span>',
})
const DataTableStub = defineComponent({
  props: ['columns', 'data', 'loading'],
  template: '<div data-testid="plans-table"><slot /></div>',
})
const RechargeProductsManagerStub = defineComponent({
  template: '<section data-testid="recharge-products-manager">充值商品管理</section>',
})

describe('Admin products and plans catalog page', () => {
  it('combines recharge product and subscription plan management behind one route', () => {
    expect(plansViewSource).toContain("import RechargeProductsManager from './RechargeProductsManager.vue'")
    expect(plansViewSource).toContain("type CatalogTabKey = 'recharge' | 'plans'")
    expect(plansViewSource).toContain("const activeTab = ref<CatalogTabKey>('plans')")
    expect(plansViewSource).toContain("activeTab === 'recharge'")
    expect(plansViewSource).toContain("activeTab === 'plans'")
    expect(plansViewSource).toContain("key: 'recharge'")
    expect(plansViewSource).toContain("key: 'plans'")
  })

  it('keeps recharge products settings-backed without changing subscription plan APIs', () => {
    expect(rechargeManagerSource).toContain('settingsAPI.getSettings()')
    expect(rechargeManagerSource).toContain('settingsAPI.updateSettings({ payment_recharge_products: normalized })')
    expect(rechargeManagerSource).toContain('credited_amount: Number(product.credited_amount) || 0')
    expect(plansViewSource).toContain('adminPaymentAPI.getPlans()')
    expect(plansViewSource).toContain('adminPaymentAPI.updatePlan')
    expect(plansViewSource).toContain('adminPaymentAPI.deletePlan')
  })

  it('renames the admin navigation label to products and plans', () => {
    expect(zhLocaleSource).toContain("paymentPlans: '商品/套餐'")
    expect(zhLocaleSource).toContain("plansPageTitle: '商品/套餐管理'")
    expect(enLocaleSource).toContain("paymentPlans: 'Products & Plans'")
    expect(enLocaleSource).toContain("plansPageTitle: 'Products & Plans'")
  })

  it('renders the combined catalog shell and switches to recharge product management', async () => {
    getPlans.mockResolvedValue({
      data: [
        {
          id: 1,
          name: 'Pro',
          group_id: 10,
          price: 100,
          original_price: 0,
          validity_days: 30,
          validity_unit: 'days',
          for_sale: true,
          sort_order: 10,
          features: '高速使用',
        },
      ],
    })
    getGroups.mockResolvedValue([])

    const wrapper = mount(AdminPaymentPlansView, {
      global: {
        stubs: {
          AppLayout: AppLayoutStub,
          Icon: IconStub,
          DataTable: DataTableStub,
          GroupBadge: true,
          PlanEditDialog: true,
          ConfirmDialog: true,
          RechargeProductsManager: RechargeProductsManagerStub,
        },
      },
    })

    await flushPromises()

    expect(wrapper.text()).toContain('商品/套餐管理')
    expect(wrapper.text()).toContain('订阅套餐')
    expect(wrapper.find('[data-testid="plans-table"]').exists()).toBe(true)

    await wrapper.get('button:nth-of-type(1)').trigger('click')

    expect(wrapper.find('[data-testid="recharge-products-manager"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="plans-table"]').exists()).toBe(false)
  })
})
