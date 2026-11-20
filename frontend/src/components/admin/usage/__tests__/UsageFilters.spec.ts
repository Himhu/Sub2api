import { describe, expect, it, vi, beforeEach, afterEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'

import UsageFilters from '../UsageFilters.vue'

// --- i18n messages (only what UsageFilters needs) ---
const messages: Record<string, string> = {
  'usage.apiKeyFilter': 'API Key',
  'admin.usage.searchApiKeyPlaceholder': 'Search API key...',
  'usage.model': 'Model',
  'admin.usage.allModels': 'All Models',
  'admin.usage.account': 'Account',
  'admin.usage.searchAccountPlaceholder': 'Search account...',
  'usage.type': 'Type',
  'admin.usage.allTypes': 'All Types',
  'usage.ws': 'WS',
  'usage.stream': 'Stream',
  'usage.sync': 'Sync',
  'usage.compactionFilter': 'Request Kind',
  'usage.allCompactionTypes': 'All Requests',
  'usage.compactionOnly': 'Compaction Only',
  'admin.usage.billingType': 'Billing Type',
  'admin.usage.allBillingTypes': 'All Billing Types',
  'admin.usage.billingTypeBalance': 'Balance',
  'admin.usage.billingTypeSubscription': 'Subscription',
  'admin.usage.billingMode': 'Billing Mode',
  'admin.usage.allBillingModes': 'All Billing Modes',
  'admin.usage.billingModeToken': 'Token',
  'admin.usage.billingModePerRequest': 'Per Request',
  'admin.usage.billingModeImage': 'Image',
	'admin.usage.upstreamModelAudit': 'Upstream model audit',
	'admin.usage.allUpstreamModelAudit': 'All response model states',
	'admin.usage.upstreamModelMismatchOnly': 'Mismatched only',
	'admin.usage.upstreamModelMatchedOnly': 'Matched only',
  'admin.usage.group': 'Group',
  'admin.usage.allGroups': 'All Groups',
  'common.refresh': 'Refresh',
  'common.reset': 'Reset',
  'admin.usage.cleanup.button': 'Cleanup',
  'usage.exportExcel': 'Export',
}

// Mock vue-i18n
vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => messages[key] ?? key,
    }),
  }
})

// Mock the admin API module — we control searchApiKeys return value per test
const mockSearchApiKeys = vi.fn().mockResolvedValue([])
const mockGroupsList = vi.fn().mockResolvedValue({ items: [] })
const mockGetModelStats = vi.fn().mockResolvedValue({ models: [] })
const mockAccountsList = vi.fn().mockResolvedValue({ items: [] })

vi.mock('@/api/admin', () => ({
  adminAPI: {
    usage: {
      searchApiKeys: (...args: any[]) => mockSearchApiKeys(...args),
    },
    groups: { list: (...args: any[]) => mockGroupsList(...args) },
    dashboard: { getModelStats: (...args: any[]) => mockGetModelStats(...args) },
    accounts: { list: (...args: any[]) => mockAccountsList(...args) },
  },
}))

// Default props helper
const defaultFilters = () => ({
  api_key_id: undefined as number | undefined,
  account_id: undefined,
  model: null,
  request_type: null,
  native_compaction_v2: null,
  billing_type: null,
  billing_mode: null,
	upstream_model_mismatch: null,
  group_id: null,
  start_date: '',
  end_date: '',
})

function mountFilters(filters = defaultFilters()) {
  return mount(UsageFilters, {
    props: {
      modelValue: filters,
      exporting: false,
      startDate: '2026-05-01',
      endDate: '2026-05-28',
      showActions: false,
      modelOptions: [],
    },
    global: {
      stubs: {
        Select: true,
        Teleport: true,
      },
    },
  })
}

describe('UsageFilters — API key search dropdown', () => {
  beforeEach(() => {
    vi.useFakeTimers()
    mockSearchApiKeys.mockReset().mockResolvedValue([])
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('has no user search filter', () => {
    const wrapper = mountFilters()
    expect(wrapper.find('[aria-label="Clear user filter"]').exists()).toBe(false)
    // API key + account search inputs only
    expect(wrapper.findAll('.usage-filter-dropdown')).toHaveLength(2)
  })

  it('loads keys on focus without a user, searches by keyword only, and selection sets api_key_id', async () => {
    mockSearchApiKeys.mockResolvedValue([
      { id: 11, name: 'key-a', user_id: 1 },
      { id: 12, name: 'key-b', user_id: 1 },
    ])

    const wrapper = mountFilters()
    const input = wrapper.find('input[type="text"]')

    await input.trigger('focus')
    vi.advanceTimersByTime(300)
    await flushPromises()
    expect(mockSearchApiKeys).toHaveBeenCalledTimes(1)
    expect(mockSearchApiKeys).toHaveBeenLastCalledWith('')

    await input.setValue('key-b')
    vi.advanceTimersByTime(300)
    await flushPromises()
    expect(mockSearchApiKeys).toHaveBeenLastCalledWith('key-b')

    const buttons = wrapper.findAll('.usage-filter-dropdown button[type="button"]')
    const target = buttons.find((b) => b.text().includes('key-b'))
    expect(target).toBeDefined()
    await target!.trigger('click')
    await flushPromises()

    expect(wrapper.props('modelValue').api_key_id).toBe(12)
    expect(wrapper.emitted('change')).toBeTruthy()
  })
})

describe('UsageFilters — model options come from prop (no dup request)', () => {
  beforeEach(() => {
    vi.useFakeTimers()
    mockGetModelStats.mockClear()
    mockGroupsList.mockClear()
  })
  afterEach(() => { vi.useRealTimers() })

  it('does not call dashboard.getModelStats on mount and renders model options from prop', async () => {
    const wrapper = mount(UsageFilters, {
      props: {
        modelValue: defaultFilters(),
        exporting: false,
        startDate: '2026-05-01',
        endDate: '2026-05-28',
        showActions: false,
        modelOptions: ['claude-3', 'gpt-4o'],
      },
      global: { stubs: { Select: true, Teleport: true } },
    })
    await flushPromises()

    expect(mockGetModelStats).not.toHaveBeenCalled()

    const opts = (wrapper.vm as any).modelOptions as Array<{ value: string | null; label: string }>
    expect(opts.map((o) => o.value)).toEqual([null, 'claude-3', 'gpt-4o'])
  })
})

describe('UsageFilters — native compaction filter', () => {
  it('offers only All/Compaction and emits the independent boolean filter', async () => {
    const SelectStub = {
      name: 'Select',
      props: ['modelValue', 'options'],
      emits: ['update:modelValue', 'change'],
      template: '<div />',
    }
    const filters = defaultFilters()
    const wrapper = mount(UsageFilters, {
      props: {
        modelValue: filters,
        exporting: false,
        startDate: '2026-05-01',
        endDate: '2026-05-28',
        showActions: false,
        modelOptions: [],
      },
      global: { stubs: { Select: SelectStub, Teleport: true } },
    })

    const compactionSelect = wrapper.findAllComponents(SelectStub).find((select: any) =>
      (select.props('options') as Array<{ value: unknown }>).some((option) => option.value === true)
    )
    expect(compactionSelect).toBeDefined()
    expect(compactionSelect!.props('options')).toEqual([
      { value: null, label: 'All Requests' },
      { value: true, label: 'Compaction Only' },
    ])
    expect(compactionSelect!.props('options')).not.toEqual(
      expect.arrayContaining([expect.objectContaining({ value: false })])
    )

    compactionSelect!.vm.$emit('update:modelValue', true)
    compactionSelect!.vm.$emit('change')
    await wrapper.vm.$nextTick()

    expect(filters.native_compaction_v2).toBe(true)
    expect(wrapper.emitted('change')).toBeTruthy()
  })
})
