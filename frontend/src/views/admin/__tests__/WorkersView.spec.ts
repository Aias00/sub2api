import { describe, expect, it, vi, beforeEach } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { defineComponent, h } from 'vue'

import WorkersView from '../WorkersView.vue'

const {
  getRuntimeWorkers,
  manageRuntimeWorker,
  showError,
  showSuccess,
} = vi.hoisted(() => ({
  getRuntimeWorkers: vi.fn(),
  manageRuntimeWorker: vi.fn(),
  showError: vi.fn(),
  showSuccess: vi.fn(),
}))

vi.mock('@/api/admin', () => ({
  adminAPI: {
    settings: {
      getRuntimeWorkers,
      manageRuntimeWorker,
    },
  },
}))

vi.mock('@/stores', () => ({
  useAppStore: () => ({
    showError,
    showSuccess,
  }),
}))

vi.mock('@/utils/apiError', () => ({
  extractApiErrorMessage: (_error: unknown, fallback: string) => fallback,
}))

const IconStub = defineComponent({
  setup() {
    return () => h('span')
  },
})

function mountView() {
  return mount(WorkersView, {
    props: {
      embedded: true,
    },
    global: {
      stubs: {
        AppLayout: { template: '<div><slot /></div>' },
        Icon: IconStub,
      },
    },
  })
}

describe('WorkersView', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    getRuntimeWorkers.mockResolvedValue({
      management: {
        enabled: true,
        deploy_enabled: true,
      },
      workers: [
        {
          id: 'hot-worker',
          name: 'Hot Worker',
          node_id: 'hot-worker',
          container_name: 'sub2api-hot-worker',
          container_state: 'running',
          image: 'registry.example.com/app:old',
          health: 'idle',
          deployable: true,
          manageable: true,
          queue: 0,
          running: 0,
          failed: 0,
          stale: 0,
          total: 0,
          succeeded: 0,
          last_updated_at: null,
          attention_reasons: [],
        },
      ],
    })
    manageRuntimeWorker.mockResolvedValue({})
  })

  it('submits worker image deployment through the modal form', async () => {
    const wrapper = mountView()
    await flushPromises()

    const deployButton = wrapper
      .findAll('button')
      .find((button) => button.text().includes('更新镜像'))
    expect(deployButton).toBeDefined()
    await deployButton!.trigger('click')
    await flushPromises()

    await wrapper.get('#worker-image-input').setValue('registry.example.com/app:new')
    await wrapper.get('form').trigger('submit')
    await flushPromises()

    expect(manageRuntimeWorker).toHaveBeenCalledWith('hot-worker', 'deploy', {
      image: 'registry.example.com/app:new',
      pull: true,
      restart: true,
    })
    expect(showSuccess).toHaveBeenCalledWith('Worker 镜像更新已执行')
  })
})
