import { describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import AccountTodayStatsCell from '../AccountTodayStatsCell.vue'
import type { WindowStats } from '@/types'

// format.ts 会加载真实的 i18n 实例，所以只替换 useI18n，保留 createI18n 等其余导出。
vi.mock('vue-i18n', async (importOriginal) => ({
  ...(await importOriginal<typeof import('vue-i18n')>()),
  useI18n: () => ({ t: (key: string) => key }),
}))

const stats = { requests: 10, tokens: 254, cost: 0.002446, user_cost: 0.005869 } as WindowStats

describe('AccountTodayStatsCell', () => {
  it('shows the upstream and site multipliers next to the two costs', () => {
    const wrapper = mount(AccountTodayStatsCell, {
      props: { stats, showRates: true, rateMultiplier: 0.5, siteRateMultiplier: 1.2 },
    })
    expect(wrapper.get('[data-testid="today-stats-upstream-rate"]').text()).toBe('×0.5')
    expect(wrapper.get('[data-testid="today-stats-site-rate"]').text()).toBe('×1.2')
  })

  it('labels the site rate as following the group when the account has none', () => {
    const wrapper = mount(AccountTodayStatsCell, {
      props: { stats, showRates: true, rateMultiplier: 1, siteRateMultiplier: null },
    })
    expect(wrapper.get('[data-testid="today-stats-upstream-rate"]').text()).toBe('×1')
    expect(wrapper.get('[data-testid="today-stats-site-rate"]').text()).toBe(
      'admin.accounts.siteRateMultiplierFollowGroup',
    )
  })

  it('keeps a zero site multiplier (free line) distinct from "follow the group"', () => {
    const wrapper = mount(AccountTodayStatsCell, {
      props: { stats, showRates: true, rateMultiplier: 1, siteRateMultiplier: 0 },
    })
    expect(wrapper.get('[data-testid="today-stats-site-rate"]').text()).toBe('×0')
  })

  it('renders no multipliers unless asked to', () => {
    const wrapper = mount(AccountTodayStatsCell, { props: { stats } })
    expect(wrapper.find('[data-testid="today-stats-upstream-rate"]').exists()).toBe(false)
    expect(wrapper.find('[data-testid="today-stats-site-rate"]').exists()).toBe(false)
  })
})
