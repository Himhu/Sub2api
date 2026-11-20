/**
 * 用户侧模型定价类型（模型广场 / 支持模型徽章共用）。
 */

import type { BillingMode } from '@/constants/channel'

export interface UserPricingInterval {
  min_tokens: number
  max_tokens: number | null
  tier_label?: string
  input_price: number | null
  output_price: number | null
  cache_write_price: number | null
  cache_write_1h_price?: number | null
  cache_write_12h_price?: number | null
  cache_read_price: number | null
  input_multiplier?: number | null
  output_multiplier?: number | null
  cache_write_multiplier?: number | null
  cache_read_multiplier?: number | null
  per_request_price: number | null
}

export interface UserSupportedModelPricing {
  billing_mode: BillingMode
  input_price: number | null
  output_price: number | null
  cache_write_price: number | null
  cache_write_1h_price?: number | null
  cache_write_12h_price?: number | null
  cache_read_price: number | null
  reasoning_effort_multipliers?: Record<string, number> | null
  image_input_price: number | null
  image_output_price: number | null
  per_request_price: number | null
  intervals: UserPricingInterval[]
}

export interface UserSupportedModel {
  name: string
  platform: string
  pricing: UserSupportedModelPricing | null
}

/**
 * 渠道下单个平台的子视图：用户可访问的分组 + 该平台支持的模型。
 * 后端把一个渠道按平台聚合成 sections，前端可以把渠道名作为 row-group
 * 一次渲染，后面按 sections 顺序用 rowspan 铺开。
 */
