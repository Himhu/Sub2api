import { describe, expect, it } from 'vitest'

import en from '../locales/en'
import zh from '../locales/zh'

describe('OpenAI Fast/Flex policy locale keys', () => {
  it('describes target and other-model actions without whitelist terminology', () => {
    expect(zh.admin.settings.openaiFastPolicy).toMatchObject({
      tierAll: '全部 tier 值',
      tierUltrafast: 'ultrafast',
      tierMissing: '省略 tier',
      modelWhitelist: '目标模型',
      fallbackAction: '其他模型处理方式',
      summaryTargetModels: '目标模型',
      summaryOtherModels: '其他模型'
    })
    expect(zh.admin.settings.openaiFastPolicy.modelWhitelistHint).toContain(
      '留空时“处理方式”应用于全部模型'
    )
    expect(zh.admin.settings.openaiFastPolicy.modelWhitelistHint).not.toContain('白名单')

    expect(en.admin.settings.openaiFastPolicy).toMatchObject({
      tierAll: 'All tier values',
      tierUltrafast: 'ultrafast',
      tierMissing: 'Omitted tier',
      modelWhitelist: 'Target models',
      fallbackAction: 'Other models action',
      summaryTargetModels: 'Target models',
      summaryOtherModels: 'Other models'
    })
    expect(en.admin.settings.openaiFastPolicy.modelWhitelistHint).toContain(
      'Leave empty to apply Action to all models'
    )
    expect(en.admin.settings.openaiFastPolicy.modelWhitelistHint).not.toContain('whitelist')
  })
})
