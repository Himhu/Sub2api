<template>
  <div>
    <!-- Loading state -->
    <div v-if="props.loading && !props.stats" class="space-y-0.5">
      <div class="h-3 w-12 animate-pulse rounded bg-gray-200 dark:bg-gray-700"></div>
      <div class="h-3 w-16 animate-pulse rounded bg-gray-200 dark:bg-gray-700"></div>
      <div class="h-3 w-10 animate-pulse rounded bg-gray-200 dark:bg-gray-700"></div>
    </div>

    <!-- Error state -->
    <div v-else-if="props.error && !props.stats" class="text-xs text-red-500">
      {{ props.error }}
    </div>

    <!-- Stats data -->
    <div v-else-if="props.stats" class="space-y-0.5 text-xs">
      <!-- Requests -->
      <div class="flex items-center gap-1">
        <span class="text-gray-500 dark:text-gray-400"
          >{{ t('admin.accounts.stats.requests') }}:</span
        >
        <span class="font-medium text-gray-700 dark:text-gray-300">{{
          formatNumber(props.stats.requests)
        }}</span>
      </div>
      <!-- Tokens -->
      <div class="flex items-center gap-1">
        <span class="text-gray-500 dark:text-gray-400"
          >{{ t('admin.accounts.stats.tokens') }}:</span
        >
        <span class="font-medium text-gray-700 dark:text-gray-300">{{
          formatTokens(props.stats.tokens)
        }}</span>
      </div>
      <!-- Cost (Account) -->
      <div class="flex items-center gap-1">
        <span class="text-gray-500 dark:text-gray-400">{{ t('usage.accountBilled') }}:</span>
        <span class="font-medium text-emerald-600 dark:text-emerald-400">{{
          formatCurrency(props.stats.cost)
        }}</span>
        <span
          v-if="upstreamRateLabel"
          class="text-gray-400 dark:text-gray-500"
          data-testid="today-stats-upstream-rate"
          >{{ upstreamRateLabel }}</span
        >
      </div>
      <!-- Cost (User/API Key) -->
      <div v-if="props.stats.user_cost != null" class="flex items-center gap-1">
        <span class="text-gray-500 dark:text-gray-400">{{ t('usage.userBilled') }}:</span>
        <span class="font-medium text-gray-700 dark:text-gray-300">{{
          formatCurrency(props.stats.user_cost)
        }}</span>
        <span
          v-if="props.showRates"
          class="text-gray-400 dark:text-gray-500"
          data-testid="today-stats-site-rate"
          :title="siteRateTitle || undefined"
          >{{ siteRateLabel }}</span
        >
      </div>
    </div>

    <!-- No data -->
    <div v-else class="text-xs text-gray-400">-</div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { WindowStats } from '@/types'
import { formatNumber, formatCurrency } from '@/utils/format'

const props = withDefaults(
  defineProps<{
    stats?: WindowStats | null
    loading?: boolean
    error?: string | null
    // 传入线路的倍率后，在金额旁标出对应倍率，便于对照上游成本与本站扣费。
    showRates?: boolean
    rateMultiplier?: number | null
    siteRateMultiplier?: number | null
    // 线路所属分组的倍率；本站倍率未填写时按这些分组倍率扣费
    groupRateMultipliers?: number[]
  }>(),
  {
    stats: null,
    loading: false,
    error: null,
    showRates: false,
    rateMultiplier: null,
    siteRateMultiplier: null,
    groupRateMultipliers: () => []
  }
)

const { t } = useI18n()

const formatRate = (value: number): string => `×${Number(value.toFixed(4))}`

// 账号计费倍率（上游成本口径）
const upstreamRateLabel = computed(() => {
  if (!props.showRates || props.rateMultiplier == null) return ''
  return formatRate(props.rateMultiplier)
})

const hasSiteRate = computed(
  () => props.siteRateMultiplier != null && props.siteRateMultiplier >= 0
)

// 本站倍率：未填写时扣费沿用分组倍率，直接列出所属分组的实际倍率
// （多个分组倍率不同则全部列出；未归属分组时后端按 1 计费）。
const siteRateLabel = computed(() => {
  if (hasSiteRate.value) return formatRate(props.siteRateMultiplier as number)
  const rates = [...new Set(props.groupRateMultipliers.filter(rate => Number.isFinite(rate)))]
  if (rates.length === 0) return formatRate(1)
  return rates
    .sort((a, b) => a - b)
    .map(formatRate)
    .join(' / ')
})

// 鼠标悬停时说明该倍率来自分组而非线路自身
const siteRateTitle = computed(() =>
  hasSiteRate.value ? '' : t('admin.accounts.siteRateMultiplierFollowGroup')
)

// Format large token numbers (e.g., 1234567 -> 1.23M)
const formatTokens = (tokens: number): string => {
  if (tokens >= 1000000) {
    return `${(tokens / 1000000).toFixed(2)}M`
  } else if (tokens >= 1000) {
    return `${(tokens / 1000).toFixed(1)}K`
  }
  return tokens.toString()
}
</script>
