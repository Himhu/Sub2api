/**
 * Admin API barrel export
 * Centralized exports for all admin API modules
 */

import dashboardAPI from './dashboard'
import groupsAPI from './groups'
import accountsAPI from './accounts'
import proxiesAPI from './proxies'
import settingsAPI from './settings'
import usageAPI from './usage'
import geminiAPI from './gemini'
import antigravityAPI from './antigravity'
import grokAPI from './grok'
import cnProvidersAPI from './cnProviders'
import opsAPI from './ops'
import errorPassthroughAPI from './errorPassthrough'
import dataManagementAPI from './dataManagement'
import apiKeysAPI from './apiKeys'
import scheduledTestsAPI from './scheduledTests'
import backupAPI from './backup'
import tlsFingerprintProfileAPI from './tlsFingerprintProfile'
import channelsAPI from './channels'
import channelMonitorAPI from './channelMonitor'
import channelMonitorTemplateAPI from './channelMonitorTemplate'
import riskControlAPI from './riskControl'
import adminComplianceAPI from './compliance'
import auditAPI from './audit'
import pluginsAPI from './plugins'

/**
 * Unified admin API object for convenient access
 */
export const adminAPI = {
  dashboard: dashboardAPI,
  groups: groupsAPI,
  accounts: accountsAPI,
  proxies: proxiesAPI,
  settings: settingsAPI,
  usage: usageAPI,
  gemini: geminiAPI,
  antigravity: antigravityAPI,
  grok: grokAPI,
  cnProviders: cnProvidersAPI,
  ops: opsAPI,
  errorPassthrough: errorPassthroughAPI,
  dataManagement: dataManagementAPI,
  apiKeys: apiKeysAPI,
  scheduledTests: scheduledTestsAPI,
  backup: backupAPI,
  tlsFingerprintProfiles: tlsFingerprintProfileAPI,
  channels: channelsAPI,
  channelMonitor: channelMonitorAPI,
  channelMonitorTemplate: channelMonitorTemplateAPI,
  riskControl: riskControlAPI,
  compliance: adminComplianceAPI,
  audit: auditAPI,
  plugins: pluginsAPI
}

export {
  dashboardAPI,
  groupsAPI,
  accountsAPI,
  proxiesAPI,
  settingsAPI,
  usageAPI,
  geminiAPI,
  antigravityAPI,
  grokAPI,
  cnProvidersAPI,
  opsAPI,
  errorPassthroughAPI,
  dataManagementAPI,
  apiKeysAPI,
  scheduledTestsAPI,
  backupAPI,
  tlsFingerprintProfileAPI,
  channelsAPI,
  channelMonitorAPI,
  channelMonitorTemplateAPI,
  riskControlAPI,
  adminComplianceAPI,
  auditAPI,
  pluginsAPI
}

export default adminAPI

// Re-export types used by components
export type { AuditLog, AuditLogQuery, AuditLogListResponse } from './audit'
export type { ErrorPassthroughRule, CreateRuleRequest, UpdateRuleRequest } from './errorPassthrough'
export type { BackupAgentHealth, DataManagementConfig } from './dataManagement'
export type { TLSFingerprintProfile, CreateProfileRequest, UpdateProfileRequest } from './tlsFingerprintProfile'
export type { ContentModerationConfig, ContentModerationLog, ModerationMode } from './riskControl'
export type {
  PluginInstallation,
  PluginCompatibility,
  PluginUISession,
  PluginTestResult
} from './plugins'
