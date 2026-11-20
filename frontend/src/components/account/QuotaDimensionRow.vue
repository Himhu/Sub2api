<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import type { QuotaResetMode } from '@/constants/account'

const { t } = useI18n()

const props = defineProps<{
  dim: 'daily' | 'weekly' | 'total'
  label: string
  limit: number | null
  // Reset mode (only for daily/weekly, null for total)
  resetMode: QuotaResetMode | null
  resetHour: number | null
  resetDay: number | null  // weekly only
  resetTimezone: string | null
  hintRolling: string
  hintFixed: string
  // Shared options passed from parent
  hourOptions: number[]
  dayOptions: { value: number; key: string }[]
  timezoneOptions?: string[]
}>()

const emit = defineEmits<{
  'update:limit': [value: number | null]
  'update:resetMode': [value: QuotaResetMode | null]
  'update:resetHour': [value: number | null]
  'update:resetDay': [value: number | null]
  'update:resetTimezone': [value: string | null]
}>()

const hasResetMode = props.dim !== 'total'

const onLimitInput = (e: Event) => {
  const raw = (e.target as HTMLInputElement).valueAsNumber
  emit('update:limit', Number.isNaN(raw) ? null : raw)
}

const onModeChange = (e: Event) => {
  const val = (e.target as HTMLSelectElement).value as QuotaResetMode
  emit('update:resetMode', val)
  if (val === 'fixed') {
    if (props.resetHour == null) emit('update:resetHour', 0)
    if (props.dim === 'weekly' && props.resetDay == null) emit('update:resetDay', 1)
    if (!props.resetTimezone) emit('update:resetTimezone', 'UTC')
  }
}

function getTimezoneOffsetLabel(tz: string): string {
  try {
    const dtf = new Intl.DateTimeFormat('en-US', { timeZone: tz, timeZoneName: 'shortOffset' })
    const parts = dtf.formatToParts(new Date())
    const tzPart = parts.find(p => p.type === 'timeZoneName')
    return tzPart ? (tzPart.value === 'GMT' ? 'GMT+0' : tzPart.value) : ''
  } catch {
    return ''
  }
}
</script>

<template>
  <div>
    <label class="text-xs font-medium text-gray-700 dark:text-gray-300 mb-1 block">{{ label }}</label>

    <!-- Input row -->
    <div class="flex items-center gap-2">
      <div class="relative flex-1">
        <span class="absolute left-2.5 top-1/2 -translate-y-1/2 text-gray-500 dark:text-gray-400 text-sm">$</span>
        <input :value="limit" @input="onLimitInput" type="number" min="0" step="0.01" class="input pl-6 py-1.5 text-sm" :placeholder="t('admin.accounts.quotaLimitPlaceholder')" />
      </div>
    </div>

    <!-- Reset mode row (daily/weekly only) -->
    <div v-if="hasResetMode" class="mt-1 flex items-center gap-2 flex-wrap">
      <label class="text-xs text-gray-500 dark:text-gray-400 whitespace-nowrap">{{ t('admin.accounts.quotaResetMode') }}</label>
      <select :value="resetMode || 'rolling'" @change="onModeChange" class="input py-1 text-xs w-auto">
        <option value="rolling">{{ t('admin.accounts.quotaResetModeRolling') }}</option>
        <option value="fixed">{{ t('admin.accounts.quotaResetModeFixed') }}</option>
      </select>
      <template v-if="resetMode === 'fixed'">
        <!-- Weekly: day of week selector -->
        <template v-if="dim === 'weekly'">
          <label class="text-xs text-gray-500 dark:text-gray-400 whitespace-nowrap">{{ t('admin.accounts.quotaWeeklyResetDay') }}</label>
          <select :value="resetDay ?? 1" @change="emit('update:resetDay', Number(($event.target as HTMLSelectElement).value))" class="input py-1 text-xs w-28">
            <option v-for="d in dayOptions" :key="d.value" :value="d.value">{{ t('admin.accounts.dayOfWeek.' + d.key) }}</option>
          </select>
        </template>
        <label class="text-xs text-gray-500 dark:text-gray-400 whitespace-nowrap">{{ t('admin.accounts.quotaResetHour') }}</label>
        <select :value="resetHour ?? 0" @change="emit('update:resetHour', Number(($event.target as HTMLSelectElement).value))" class="input py-1 text-xs w-24">
          <option v-for="h in hourOptions" :key="h" :value="h">{{ String(h).padStart(2, '0') }}:00</option>
        </select>
        <template v-if="timezoneOptions && timezoneOptions.length > 0">
          <select :value="resetTimezone || 'UTC'" @change="emit('update:resetTimezone', ($event.target as HTMLSelectElement).value)" class="input py-1 text-xs w-auto">
            <option v-for="tz in timezoneOptions" :key="tz" :value="tz">{{ tz }} ({{ getTimezoneOffsetLabel(tz) }})</option>
          </select>
        </template>
      </template>
      <span class="text-[11px] text-gray-500 dark:text-gray-400">
        <template v-if="resetMode === 'fixed'">{{ hintFixed }}</template>
        <template v-else>{{ hintRolling }}</template>
      </span>
    </div>

    <!-- Total dimension hint (no reset mode) -->
    <p v-if="!hasResetMode" class="input-hint mb-0 text-[11px]">{{ hintRolling }}</p>
  </div>
</template>
