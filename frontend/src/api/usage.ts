/**
 * Usage tracking API endpoints
 * Handles per-key usage statistics for the API key page
 */

import { apiClient } from './client'

export interface BatchApiKeyUsageStats {
  api_key_id: number
  today_actual_cost: number
  total_actual_cost: number
}

export interface BatchApiKeysUsageResponse {
  stats: Record<string, BatchApiKeyUsageStats>
}

/**
 * Get batch usage stats for user's own API keys
 * @param apiKeyIds - Array of API key IDs
 * @param options - Optional request options
 * @returns Usage stats map keyed by API key ID
 */
export async function getDashboardApiKeysUsage(
  apiKeyIds: number[],
  options?: {
    signal?: AbortSignal
  }
): Promise<BatchApiKeysUsageResponse> {
  const { data } = await apiClient.post<BatchApiKeysUsageResponse>(
    '/usage/dashboard/api-keys-usage',
    {
      api_key_ids: apiKeyIds
    },
    {
      signal: options?.signal
    }
  )
  return data
}

export const usageAPI = {
  getDashboardApiKeysUsage
}

export default usageAPI
