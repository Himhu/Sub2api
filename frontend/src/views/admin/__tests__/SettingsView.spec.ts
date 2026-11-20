import { beforeEach, describe, expect, it, vi } from "vitest";
import { defineComponent, h } from "vue";
import { flushPromises, mount } from "@vue/test-utils";

import SettingsView from "../SettingsView.vue";

const {
  getSettings,
  updateSettings,
  getWebSearchEmulationConfig,
  updateWebSearchEmulationConfig,
  getAdminApiKey,
  getOverloadCooldownSettings,
  getRateLimit429CooldownSettings,
  updateRateLimit429CooldownSettings,
  getPanelRateLimitSettings,
  updatePanelRateLimitSettings,
  getStreamTimeoutSettings,
  getRectifierSettings,
  getBetaPolicySettings,
  getUpstreamBillingProbeSettings,
  updateUpstreamBillingProbeSettings,
  getOllamaCloudUsageSettings,
  updateOllamaCloudUsageSettings,
  listProxies,
  fetchPublicSettings,
  adminSettingsFetch,
  showError,
  showSuccess,
} = vi.hoisted(() => ({
  getSettings: vi.fn(),
  updateSettings: vi.fn(),
  getWebSearchEmulationConfig: vi.fn(),
  updateWebSearchEmulationConfig: vi.fn(),
  getAdminApiKey: vi.fn(),
  getOverloadCooldownSettings: vi.fn(),
  getRateLimit429CooldownSettings: vi.fn(),
  updateRateLimit429CooldownSettings: vi.fn(),
  getPanelRateLimitSettings: vi.fn().mockResolvedValue({
    enabled: true,
    user_rpm: 240,
    heavy_rpm: 60,
    exempt_admin: true,
    public_ip_rpm: 300,
  }),
  updatePanelRateLimitSettings: vi.fn().mockImplementation(async (payload) => payload),
  getStreamTimeoutSettings: vi.fn(),
  getRectifierSettings: vi.fn(),
  getBetaPolicySettings: vi.fn(),
  getUpstreamBillingProbeSettings: vi.fn().mockResolvedValue({
    enabled: true,
    interval_minutes: 30,
  }),
  updateUpstreamBillingProbeSettings: vi.fn().mockImplementation(async (payload) => payload),
  getOllamaCloudUsageSettings: vi.fn().mockResolvedValue({
    enabled: false,
    interval_minutes: 60,
    debounce_minutes: 1,
  }),
  updateOllamaCloudUsageSettings: vi.fn().mockImplementation(async (payload) => payload),
  listProxies: vi.fn(),
  fetchPublicSettings: vi.fn(),
  adminSettingsFetch: vi.fn(),
  showError: vi.fn(),
  showSuccess: vi.fn(),
}));

const localeRef = vi.hoisted(() => ({ value: "zh-CN" }));

vi.mock("@/api", () => ({
  adminAPI: {
    settings: {
      getSettings,
      updateSettings,
      getWebSearchEmulationConfig,
      updateWebSearchEmulationConfig,
      getAdminApiKey,
      getOverloadCooldownSettings,
      getRateLimit429CooldownSettings,
      updateRateLimit429CooldownSettings,
      getPanelRateLimitSettings,
      updatePanelRateLimitSettings,
      getStreamTimeoutSettings,
      getRectifierSettings,
      getBetaPolicySettings,
    },
    accounts: {
      getUpstreamBillingProbeSettings,
      updateUpstreamBillingProbeSettings,
      getOllamaCloudUsageSettings,
      updateOllamaCloudUsageSettings,
    },
    proxies: {
      list: listProxies,
    },
  },
}));

vi.mock("@/stores", () => ({
  useAppStore: () => ({
    showError,
    showSuccess,
    showWarning: vi.fn(),
    showInfo: vi.fn(),
    fetchPublicSettings,
  }),
}));

vi.mock("@/stores/adminSettings", () => ({
  useAdminSettingsStore: () => ({
    fetch: adminSettingsFetch,
  }),
}));

vi.mock("@/utils/apiError", () => ({
  extractApiErrorMessage: () => "error",
}));

vi.mock("vue-i18n", async () => {
  const actual = await vi.importActual<typeof import("vue-i18n")>("vue-i18n");
  const translations: Record<string, string> = {
    "admin.settings.openaiExperimentalScheduler.title": "OpenAI 实验调度策略",
    "admin.settings.openaiExperimentalScheduler.description": "默认关闭。开启后仅影响本网关在 OpenAI 账号间的实验性调度选择逻辑，不代表上游 OpenAI 官方能力。",
    "admin.settings.openaiExperimentalScheduler.lowRatePriorityTitle": "低倍率优先",
    "admin.settings.openaiExperimentalScheduler.lowRatePriorityDescription": "开启后优先选择计费倍率较低的账号；倍率相同时，再比较账号优先级和当前负载等。启用实验调度策略后，此开关不生效。",
    "admin.settings.openaiExperimentalScheduler.oauthRateTitle": "OAuth 调度参考倍率",
    "admin.settings.openaiExperimentalScheduler.oauthRatePriorityDescription": "OAuth 账号按此参考倍率参与低倍率优先排序；留空时使用各自的账号倍率。API Key 账号优先使用有效探测倍率，无有效探测时使用账号倍率。",
    "admin.settings.openaiExperimentalScheduler.oauthRateWeightedDescription": "计算“计费倍率”得分时，OAuth 账号使用此参考倍率；留空时使用各自的账号倍率。API Key 账号优先使用有效探测倍率，无有效探测时使用账号倍率。",
    "admin.settings.openaiExperimentalScheduler.oauthRateInvalid": "OAuth 调度参考倍率必须是非负数字，或留空以使用账号倍率。",
    "admin.settings.openaiExperimentalScheduler.stickyWeightedTitle": "粘性加权",
    "admin.settings.openaiExperimentalScheduler.stickyWeightedDescription": "开启后 previous_response_id 和 session_hash 粘性进入高级调度打分；关闭时仍按旧逻辑硬命中粘性账号。",
    "admin.settings.openaiExperimentalScheduler.subscriptionPriorityTitle": "订阅优先",
    "admin.settings.openaiExperimentalScheduler.subscriptionPriorityDescription": "开启后先在 ChatGPT 订阅账号池中按权值选取；订阅池拿不到席位时再回退到非订阅账号池。",
    "admin.settings.openaiExperimentalScheduler.weightsTitle": "调度权值覆盖",
    "admin.settings.openaiExperimentalScheduler.weightsDescription": "留空时使用配置/环境变量值；配置未设置时使用内置默认值。页面非空设置优先。",
    "admin.settings.openaiExperimentalScheduler.defaultPlaceholder": "配置/默认：{value}",
    "admin.settings.openaiExperimentalScheduler.topKLabel": "TopK",
    "admin.settings.openaiExperimentalScheduler.priorityWeight": "优先级",
    "admin.settings.openaiExperimentalScheduler.loadWeight": "负载",
    "admin.settings.openaiExperimentalScheduler.queueWeight": "排队",
    "admin.settings.openaiExperimentalScheduler.errorRateWeight": "错误率",
    "admin.settings.openaiExperimentalScheduler.ttftWeight": "首包延迟",
    "admin.settings.openaiExperimentalScheduler.resetWeight": "重置窗口",
    "admin.settings.openaiExperimentalScheduler.quotaHeadroomWeight": "额度余量",
    "admin.settings.openaiExperimentalScheduler.upstreamCostWeight": "计费倍率",
    "admin.settings.openaiExperimentalScheduler.previousResponseWeight": "previous_response 粘性",
    "admin.settings.openaiExperimentalScheduler.sessionStickyWeight": "session_hash 粘性",
    "admin.settings.upstreamBillingProbe.title": "上游倍率自动探测",
    "admin.settings.upstreamBillingProbe.description": "定期获取 OpenAI API Key 所连接上游 Sub2API 站点声明的计费倍率。",
    "admin.settings.upstreamBillingProbe.enabled": "启用全局自动探测",
    "admin.settings.upstreamBillingProbe.enabledHint": "开启后，仅对账号自身已启用自动检测的账号执行定时探测。",
    "admin.settings.upstreamBillingProbe.intervalMinutes": "探测周期（分钟）",
    "admin.settings.upstreamBillingProbe.intervalHint": "范围 5–1440 分钟。",
    "admin.settings.upstreamBillingProbe.saved": "上游倍率自动探测设置已保存",
    "admin.settings.upstreamBillingProbe.saveFailed": "保存上游倍率自动探测设置失败",
    "admin.settings.openaiFastPolicy.summaryTargetModels": "目标模型",
    "admin.settings.openaiFastPolicy.summaryAllModels": "全部模型",
    "admin.settings.openaiFastPolicy.summaryOtherModels": "其他模型",
    "admin.settings.openaiFastPolicy.summaryAction.filter": "过滤",
    "admin.settings.openaiFastPolicy.summaryAction.pass": "透传",
    "admin.settings.security.passkeyDeploymentHint":
      "请由服务器运维在部署配置中将 webauthn.enabled 设为 true，填写 webauthn.rp_id（仅域名）与 webauthn.rp_origins（完整 HTTPS 来源），然后重启服务。",
    "admin.settings.site.uploadImage": "上传图片",
    "admin.settings.site.remove": "移除",
  };
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string, params?: Record<string, string>) =>
        (translations[key] ?? key).replace(/\{(\w+)\}/g, (_, token) => params?.[token] ?? `{${token}}`),
      locale: localeRef,
    }),
  };
});

const AppLayoutStub = { template: "<div><slot /></div>" };
const ToggleStub = defineComponent({
  props: {
    modelValue: {
      type: Boolean,
      default: false,
    },
  },
  emits: ["update:modelValue"],
  inheritAttrs: false,
  setup(props, { attrs, emit }) {
    return () =>
      h("input", {
        ...attrs,
        class: "toggle-stub",
        type: "checkbox",
        checked: props.modelValue,
        onChange: (event: Event) => {
          emit("update:modelValue", (event.target as HTMLInputElement).checked);
        },
      });
  },
});

const SelectStub = defineComponent({
  props: {
    modelValue: {
      type: [String, Number, Boolean, null],
      default: "",
    },
    options: {
      type: Array,
      default: () => [],
    },
    placeholder: {
      type: String,
      default: "",
    },
  },
  emits: ["update:modelValue", "change"],
  setup(props, { emit }) {
    const onChange = (event: Event) => {
      const target = event.target as HTMLSelectElement;
      emit("update:modelValue", target.value);
      const option =
        (props.options as Array<Record<string, unknown>>).find(
          (item) => String(item.value ?? "") === target.value,
        ) ?? null;
      emit("change", target.value, option);
    };

    return () =>
      h(
        "select",
        {
          class: "select-stub",
          value: props.modelValue ?? "",
          "data-placeholder": props.placeholder,
          onChange,
        },
        (props.options as Array<Record<string, unknown>>).map((option) =>
          h(
            "option",
            {
              key: `${String(option.value ?? "")}:${String(option.label ?? "")}`,
              value: option.value as string,
            },
            String(option.label ?? ""),
          ),
        ),
      );
  },
});

const ImageUploadStub = defineComponent({
  props: {
    modelValue: {
      type: String,
      default: "",
    },
    uploadLabel: {
      type: String,
      default: "",
    },
    removeLabel: {
      type: String,
      default: "",
    },
    placeholder: {
      type: String,
      default: "",
    },
  },
  setup(props) {
    return () =>
      h("div", {
        class: "image-upload-stub",
        "data-model-value": props.modelValue,
        "data-upload-label": props.uploadLabel,
        "data-remove-label": props.removeLabel,
        "data-placeholder": props.placeholder,
      });
  },
});

const baseSettingsResponse = {
  totp_enabled: false,
  totp_encryption_key_configured: false,
  passkey_enabled: true,
  passkey_configured: true,
  passkey_rp_id: "sub3.nebula-spaces.com",
  passkey_rp_origins: ["https://sub3.nebula-spaces.com"],
  site_name: "Sub2API",
  site_logo: "",
  site_subtitle: "",
  api_base_url: "",
  contact_info: "",
  doc_url: "",
  home_content: "",
  compact_home_enabled: false,
  hide_ccs_import_button: false,
  table_default_page_size: 20,
  table_page_size_options: [10, 20, 50, 100],
  backend_mode_enabled: false,
  custom_menu_items: [],
  custom_endpoints: [],
  turnstile_enabled: false,
  turnstile_site_key: "",
  turnstile_secret_key_configured: false,
  tencent_captcha_enabled: false,
  tencent_captcha_app_id: "",
  tencent_captcha_app_secret_key_configured: false,
  tencent_captcha_cloud_secret_id_configured: false,
  tencent_captcha_cloud_secret_key_configured: false,
  api_key_acl_trust_forwarded_ip: true,
  forwarded_client_ip_headers: [],
  enable_model_fallback: false,
  fallback_model_anthropic: "",
  fallback_model_openai: "",
  fallback_model_gemini: "",
  fallback_model_antigravity: "",
  grok_default_text_model: "grok-4.5",
  grok_cross_client_model_map_enabled: false,
  enable_identity_patch: false,
  identity_patch_prompt: "",
  ops_monitoring_enabled: false,
  ops_realtime_monitoring_enabled: false,
  ops_query_mode_default: "auto",
  ops_metrics_interval_seconds: 60,
  min_claude_code_version: "",
  max_claude_code_version: "",
  allow_ungrouped_key_scheduling: false,
  openai_ttft_mode: "semantic",
  enable_fingerprint_unification: true,
  enable_metadata_passthrough: false,
  enable_cch_signing: false,
  enable_claude_oauth_system_prompt_injection: true,
  claude_oauth_system_prompt: "",
  claude_oauth_system_prompt_blocks: "",
  enable_anthropic_cache_ttl_1h_injection: false,
  rewrite_message_cache_control: false,
  enable_client_dateline_normalization: true,
  antigravity_user_agent_version: "",
  openai_codex_user_agent: "",
  openai_low_upstream_rate_priority_enabled: false,
  openai_oauth_scheduling_rate_multiplier: 1,
  openai_advanced_scheduler_enabled: false,
  openai_advanced_scheduler_sticky_weighted_enabled: false,
  openai_advanced_scheduler_subscription_priority_enabled: false,
  openai_advanced_scheduler_lb_top_k: "",
  openai_advanced_scheduler_weight_priority: "",
  openai_advanced_scheduler_weight_load: "",
  openai_advanced_scheduler_weight_queue: "",
  openai_advanced_scheduler_weight_error_rate: "",
  openai_advanced_scheduler_weight_ttft: "",
  openai_advanced_scheduler_weight_reset: "",
  openai_advanced_scheduler_weight_quota_headroom: "",
  openai_advanced_scheduler_weight_upstream_cost: "",
  openai_advanced_scheduler_weight_previous_response: "",
  openai_advanced_scheduler_weight_session_sticky: "",
  openai_advanced_scheduler_effective_lb_top_k: "7",
  openai_advanced_scheduler_effective_weight_priority: "1",
  openai_advanced_scheduler_effective_weight_load: "1",
  openai_advanced_scheduler_effective_weight_queue: "0.7",
  openai_advanced_scheduler_effective_weight_error_rate: "0.8",
  openai_advanced_scheduler_effective_weight_ttft: "0.5",
  openai_advanced_scheduler_effective_weight_reset: "0",
  openai_advanced_scheduler_effective_weight_quota_headroom: "0",
  openai_advanced_scheduler_effective_weight_upstream_cost: "0",
  openai_advanced_scheduler_effective_weight_previous_response: "5",
  openai_advanced_scheduler_effective_weight_session_sticky: "3",
  // 平台限额嵌套字段（新后端契约）
};

function mountView() {
  return mount(SettingsView, {
    global: {
      stubs: {
        AppLayout: AppLayoutStub,
        Select: SelectStub,
        Toggle: ToggleStub,
        Icon: true,
        ProxySelector: true,
        ImageUpload: ImageUploadStub,
        BackupSettings: true,
      },
    },
  });
}

async function openSecurityTab(wrapper: ReturnType<typeof mountView>) {
  const securityTabButton = wrapper
    .findAll("button")
    .find((node) => node.text().includes("admin.settings.tabs.security"));

  expect(securityTabButton).toBeDefined();
  await securityTabButton?.trigger("click");
  await flushPromises();
}

async function openGatewayTab(wrapper: ReturnType<typeof mountView>) {
  const gatewayTabButton = wrapper
    .findAll("button")
    .find((node) => node.text().includes("admin.settings.tabs.gateway"));

  expect(gatewayTabButton).toBeDefined();
  await gatewayTabButton?.trigger("click");
  await flushPromises();
}

describe("admin SettingsView", () => {
  beforeEach(() => {
    getSettings.mockReset();
    updateSettings.mockReset();
    getWebSearchEmulationConfig.mockReset();
    updateWebSearchEmulationConfig.mockReset();
    getAdminApiKey.mockReset();
    getOverloadCooldownSettings.mockReset();
    getRateLimit429CooldownSettings.mockReset();
    updateRateLimit429CooldownSettings.mockReset();
    getStreamTimeoutSettings.mockReset();
    getRectifierSettings.mockReset();
    getBetaPolicySettings.mockReset();
    getUpstreamBillingProbeSettings.mockReset();
    updateUpstreamBillingProbeSettings.mockReset();
    getOllamaCloudUsageSettings.mockReset();
    updateOllamaCloudUsageSettings.mockReset();
    listProxies.mockReset();
    fetchPublicSettings.mockReset();
    adminSettingsFetch.mockReset();
    showError.mockReset();
    showSuccess.mockReset();
    localeRef.value = "zh-CN";

    getSettings.mockResolvedValue({ ...baseSettingsResponse });
    updateSettings.mockImplementation(async (payload) => ({
      ...baseSettingsResponse,
      ...payload,
    }));
    getWebSearchEmulationConfig.mockResolvedValue({
      enabled: false,
      providers: [],
    });
    updateWebSearchEmulationConfig.mockResolvedValue({
      enabled: false,
      providers: [],
    });
    getAdminApiKey.mockResolvedValue({
      exists: false,
      masked_key: "",
    });
    getOverloadCooldownSettings.mockResolvedValue({
      enabled: true,
      cooldown_minutes: 10,
    });
    getRateLimit429CooldownSettings.mockResolvedValue({
      enabled: true,
      cooldown_seconds: 5,
    });
    updateRateLimit429CooldownSettings.mockImplementation(async (payload) => payload);
    getStreamTimeoutSettings.mockResolvedValue({
      enabled: true,
      action: "temp_unsched",
      temp_unsched_minutes: 5,
      threshold_count: 3,
      threshold_window_minutes: 10,
    });
    getRectifierSettings.mockResolvedValue({
      enabled: true,
      thinking_signature_enabled: true,
      thinking_budget_enabled: true,
      apikey_signature_enabled: false,
      apikey_signature_patterns: [],
    });
    getBetaPolicySettings.mockResolvedValue({
      rules: [],
    });
    getUpstreamBillingProbeSettings.mockResolvedValue({
      enabled: true,
      interval_minutes: 30,
    });
    updateUpstreamBillingProbeSettings.mockImplementation(async (payload) => payload);
    getOllamaCloudUsageSettings.mockResolvedValue({
      enabled: false,
      interval_minutes: 60,
      debounce_minutes: 1,
    });
    updateOllamaCloudUsageSettings.mockImplementation(async (payload) => payload);
    listProxies.mockResolvedValue({
      items: [],
    });
    fetchPublicSettings.mockResolvedValue(undefined);
    adminSettingsFetch.mockResolvedValue(undefined);
  });

  it("loads and saves the open button visibility for each custom menu", async () => {
    const menuItems = [
      { id: "docs", label: "Docs", url: "https://example.com/docs", icon_svg: "", visibility: "user", sort_order: 0 },
      { id: "help", label: "Help", url: "https://example.com/help", icon_svg: "", visibility: "user", sort_order: 1, hide_open_button: true },
    ];
    getSettings.mockResolvedValue({ ...baseSettingsResponse, custom_menu_items: menuItems });
    const wrapper = mountView();
    await flushPromises();

    const toggles = wrapper.findAll<HTMLInputElement>('[data-testid="custom-menu-hide-open-button"]');
    expect(toggles.map(toggle => toggle.element.checked)).toEqual([false, true]);
    await toggles[0].setValue(true);
    await toggles[1].setValue(false);
    await wrapper.find("form").trigger("submit.prevent");
    await flushPromises();

    expect(updateSettings).toHaveBeenCalledWith(expect.objectContaining({
      custom_menu_items: [
        { ...menuItems[0], hide_open_button: true },
        { ...menuItems[1], hide_open_button: false },
      ],
    }));
    wrapper.unmount();
  });

  it("submits the compact home page toggle", async () => {
    const wrapper = mountView();
    await flushPromises();

    const toggle = wrapper.get('[data-testid="compact-home-toggle"]');
    expect((toggle.element as HTMLInputElement).checked).toBe(false);

    await toggle.setValue(true);
    await wrapper.find("form").trigger("submit.prevent");
    await flushPromises();

    expect(updateSettings).toHaveBeenCalledWith(
      expect.objectContaining({ compact_home_enabled: true }),
    );
  });

  it("renders panel rate limit card and saves settings", async () => {
    getPanelRateLimitSettings.mockClear();
    updatePanelRateLimitSettings.mockClear();
    getPanelRateLimitSettings.mockResolvedValue({
      enabled: true,
      user_rpm: 240,
      heavy_rpm: 60,
      exempt_admin: true,
      public_ip_rpm: 300,
    });
    updatePanelRateLimitSettings.mockImplementation(async (payload) => payload);

    const wrapper = mountView();
    await flushPromises();

    expect(getPanelRateLimitSettings).toHaveBeenCalled();
    expect(wrapper.text()).toContain("admin.settings.panelRateLimit.title");
    expect(wrapper.text()).toContain("admin.settings.panelRateLimit.proxySafeNote");

    const userRpmInput = wrapper.find('[data-testid="panel-rate-limit-user-rpm"]');
    expect(userRpmInput.exists()).toBe(true);
    await userRpmInput.setValue("120");

    const saveButton = wrapper.find('[data-testid="panel-rate-limit-save"]');
    expect(saveButton.exists()).toBe(true);
    await saveButton.trigger("click");
    await flushPromises();

    expect(updatePanelRateLimitSettings).toHaveBeenCalledWith({
      enabled: true,
      user_rpm: 120,
      heavy_rpm: 60,
      exempt_admin: true,
      public_ip_rpm: 300,
    });
    expect(showSuccess).toHaveBeenCalled();
  });

  it("shows valid passkey RP configuration and persists the sign-in toggle", async () => {
    const wrapper = mountView();

    await flushPromises();
    await openSecurityTab(wrapper);

    const settings = wrapper.get('[data-testid="passkey-settings"]');
    const toggle = settings.get('[data-testid="passkey-toggle"]');
    expect(toggle.attributes("disabled")).toBeUndefined();
    expect(settings.text()).toContain("sub3.nebula-spaces.com");
    expect(settings.text()).toContain("https://sub3.nebula-spaces.com");
    expect(settings.text()).not.toContain("webauthn.enabled");

    await toggle.setValue(false);
    await wrapper.find("form").trigger("submit.prevent");
    await flushPromises();

    expect(updateSettings).toHaveBeenCalledWith(
      expect.objectContaining({ passkey_enabled: false }),
    );
  });

  it("人机验证切换到腾讯天御并保存四项配置", async () => {
    const wrapper = mountView();
    await flushPromises();
    await openSecurityTab(wrapper);

    const masterToggle = wrapper.get('[data-testid="captcha-enabled-toggle"]');
    await masterToggle.setValue(true);
    // 默认选中 Turnstile
    expect(wrapper.text()).toContain("admin.settings.turnstile.siteKey");

    await wrapper.get('[data-testid="captcha-provider-tencent"]').trigger("click");
    await flushPromises();

    const card = wrapper
      .findAll(".card")
      .find((node) => node.text().includes("admin.settings.captcha.title"));
    expect(card).toBeDefined();
    expect(card!.text()).not.toContain("admin.settings.turnstile.siteKey");
    expect(card!.get('a[href="https://console.cloud.tencent.com/captcha"]').exists()).toBe(true);
    expect(card!.get('a[href="https://console.cloud.tencent.com/cam/capi"]').exists()).toBe(true);
    expect(
      card!.get('a[href="https://cloud.tencent.com/document/product/1110/36841"]').exists(),
    ).toBe(true);
    const inputs = card!.findAll("input").filter((input) => input.attributes("type") !== "checkbox");
    await inputs[0]!.setValue("123456789");
    await inputs[1]!.setValue("app-secret-value");
    await inputs[2]!.setValue("cloud-secret-id-value");
    await inputs[3]!.setValue("cloud-secret-key-value");

    await wrapper.find("form").trigger("submit.prevent");
    await flushPromises();

    expect(updateSettings).toHaveBeenCalledWith(
      expect.objectContaining({
        turnstile_enabled: false,
        tencent_captcha_enabled: true,
        aliyun_captcha_enabled: false,
        tencent_captcha_app_id: "123456789",
        tencent_captcha_app_secret_key: "app-secret-value",
        tencent_captcha_cloud_secret_id: "cloud-secret-id-value",
        tencent_captcha_cloud_secret_key: "cloud-secret-key-value",
        tencent_captcha_region: "cn",
      }),
    );
  });

  it("腾讯天御切换到国际站后保存站点并更新控制台入口", async () => {
    const wrapper = mountView();
    await flushPromises();
    await openSecurityTab(wrapper);

    await wrapper.get('[data-testid="captcha-enabled-toggle"]').setValue(true);
    await wrapper.get('[data-testid="captcha-provider-tencent"]').trigger("click");
    await wrapper.get('[data-testid="tencent-captcha-region-intl"]').trigger("click");

    const card = wrapper
      .findAll(".card")
      .find((node) => node.text().includes("admin.settings.captcha.title"));
    expect(card).toBeDefined();
    expect(card!.get('a[href="https://console.tencentcloud.com/captcha/graphical"]').exists()).toBe(
      true,
    );
    expect(card!.get('a[href="https://console.tencentcloud.com/cam/capi"]').exists()).toBe(true);

    await wrapper.find("form").trigger("submit.prevent");
    await flushPromises();

    expect(updateSettings).toHaveBeenCalledWith(
      expect.objectContaining({
        tencent_captcha_enabled: true,
        tencent_captcha_region: "intl",
      }),
    );
  });

  it("人机验证切换到阿里云并保存配置", async () => {
    const wrapper = mountView();
    await flushPromises();
    await openSecurityTab(wrapper);

    const masterToggle = wrapper.get('[data-testid="captcha-enabled-toggle"]');
    await masterToggle.setValue(true);

    await wrapper.get('[data-testid="captcha-provider-aliyun"]').trigger("click");
    await flushPromises();

    const card = wrapper
      .findAll(".card")
      .find((node) => node.text().includes("admin.settings.captcha.title"));
    expect(card).toBeDefined();
    expect(card!.text()).toContain("admin.settings.aliyunCaptcha.region");
    expect(card!.text()).not.toContain("admin.settings.turnstile.siteKey");
    const inputs = card!.findAll("input").filter((input) => input.attributes("type") !== "checkbox");
    await inputs[0]!.setValue("prefix-1");
    await inputs[1]!.setValue("scene-1");
    await inputs[2]!.setValue("ak-id");
    await inputs[3]!.setValue("ak-secret-value");

    await wrapper.find("form").trigger("submit.prevent");
    await flushPromises();

    expect(updateSettings).toHaveBeenCalledWith(
      expect.objectContaining({
        turnstile_enabled: false,
        tencent_captcha_enabled: false,
        aliyun_captcha_enabled: true,
        aliyun_captcha_prefix: "prefix-1",
        aliyun_captcha_scene_id: "scene-1",
        aliyun_captcha_access_key_id: "ak-id",
        aliyun_captcha_access_key_secret: "ak-secret-value",
        aliyun_captcha_region: "cn",
      }),
    );
  });

  it("关闭人机验证总开关会同时关闭所有服务商", async () => {
    getSettings.mockResolvedValueOnce({
      ...baseSettingsResponse,
      tencent_captcha_enabled: true,
      tencent_captcha_app_id: "123456789",
      tencent_captcha_app_secret_key_configured: true,
      tencent_captcha_cloud_secret_id_configured: true,
      tencent_captcha_cloud_secret_key_configured: true,
    });
    const wrapper = mountView();
    await flushPromises();
    await openSecurityTab(wrapper);

    const masterToggle = wrapper.get('[data-testid="captcha-enabled-toggle"]');
    expect((masterToggle.element as HTMLInputElement).checked).toBe(true);
    // 加载后选中项跟随已启用的服务商
    expect(wrapper.text()).toContain("admin.settings.tencentCaptcha.appId");

    await masterToggle.setValue(false);
    await wrapper.find("form").trigger("submit.prevent");
    await flushPromises();

    expect(updateSettings).toHaveBeenCalledWith(
      expect.objectContaining({
        turnstile_enabled: false,
        tencent_captcha_enabled: false,
        aliyun_captcha_enabled: false,
      }),
    );
  });

  it("disables passkey sign-in when the RP configuration is unavailable", async () => {
    getSettings.mockResolvedValueOnce({
      ...baseSettingsResponse,
      passkey_enabled: false,
      passkey_configured: false,
      passkey_rp_id: "",
      passkey_rp_origins: [],
    });
    const wrapper = mountView();

    await flushPromises();
    await openSecurityTab(wrapper);

    const settings = wrapper.get('[data-testid="passkey-settings"]');
    expect(settings.get('[data-testid="passkey-toggle"]').attributes("disabled")).toBeDefined();
    const status = settings.get('[data-testid="passkey-config-status"]');
    expect(status.text()).toContain(
      "admin.settings.security.passkeyNotConfigured",
    );
    expect(status.text()).toContain("webauthn.enabled");
    expect(status.text()).toContain("webauthn.rp_id");
    expect(status.text()).toContain("webauthn.rp_origins");
    expect(status.text()).toContain("然后重启服务");
  });

  it("loads, edits, validates, and saves forwarded client-IP headers", async () => {
    getSettings.mockResolvedValueOnce({
      ...baseSettingsResponse,
      api_key_acl_trust_forwarded_ip: false,
      forwarded_client_ip_headers: ["cf-connecting-ip", "X-Real-IP"],
    });
    const wrapper = mountView();

    await flushPromises();
    await openSecurityTab(wrapper);

    const card = wrapper
      .findAll(".card")
      .find((node) => node.text().includes("admin.settings.apiKeyAcl.title"));
    expect(card).toBeDefined();
    const toggle = card!.get('input[type="checkbox"]');
    expect((toggle.element as HTMLInputElement).checked).toBe(false);
    expect(card!.find('[data-testid="forwarded-client-ip-headers-input"]').exists()).toBe(false);

    await toggle.setValue(true);
    expect(card!.findAll('[data-testid="forwarded-client-ip-header-tag"]')).toHaveLength(2);
    expect(card!.text()).toContain("Cf-Connecting-Ip");
    expect(card!.text()).toContain("X-Real-Ip");
    showError.mockClear();

    const input = card!.get('[data-testid="forwarded-client-ip-headers-input"]');
    await input.setValue("x-client-ip");
    await input.trigger("keydown", { key: "Enter" });
    await input.setValue("X-CLIENT-IP");
    await input.trigger("keydown", { key: "Enter" });
    await input.setValue("invalid header");
    await input.trigger("keydown", { key: "Enter" });
    expect(showError).toHaveBeenCalledTimes(1);
    expect(card!.findAll('[data-testid="forwarded-client-ip-header-tag"]')).toHaveLength(3);

    const realIpTag = card!
      .findAll('[data-testid="forwarded-client-ip-header-tag"]')
      .find((tag) => tag.text().includes("X-Real-Ip"));
    expect(realIpTag).toBeDefined();
    await realIpTag!.get("button").trigger("click");
    expect(card!.text()).not.toContain("X-Real-Ip");

    await toggle.setValue(false);
    expect(card!.find('[data-testid="forwarded-client-ip-headers-input"]').exists()).toBe(false);
    await toggle.setValue(true);
    expect(card!.text()).toContain("X-Client-Ip");

    await wrapper.find("form").trigger("submit.prevent");
    await flushPromises();

    expect(updateSettings).toHaveBeenCalledWith(
      expect.objectContaining({
        api_key_acl_trust_forwarded_ip: true,
        forwarded_client_ip_headers: ["Cf-Connecting-Ip", "X-Client-Ip"],
      }),
    );
  });

  it("submits Anthropic cache TTL injection gateway setting", async () => {
    getSettings.mockResolvedValueOnce({
      ...baseSettingsResponse,
      enable_anthropic_cache_ttl_1h_injection: true,
    });

    const wrapper = mountView();

    await flushPromises();
    await wrapper.find("form").trigger("submit.prevent");
    await flushPromises();

    expect(updateSettings).toHaveBeenCalledTimes(1);
    expect(updateSettings).toHaveBeenCalledWith(
      expect.objectContaining({
        enable_anthropic_cache_ttl_1h_injection: true,
      }),
    );
  });

  it("submits message cache_control rewrite gateway setting", async () => {
    getSettings.mockResolvedValueOnce({
      ...baseSettingsResponse,
      rewrite_message_cache_control: true,
    });

    const wrapper = mountView();

    await flushPromises();
    await wrapper.find("form").trigger("submit.prevent");
    await flushPromises();

    expect(updateSettings).toHaveBeenCalledTimes(1);
    expect(updateSettings).toHaveBeenCalledWith(
      expect.objectContaining({
        rewrite_message_cache_control: true,
      }),
    );
  });

  it("submits Claude OAuth system prompt injection gateway settings", async () => {
    const blocks = `[{"type":"text","text":"custom block","cache_control":true}]`;
    getSettings.mockResolvedValueOnce({
      ...baseSettingsResponse,
      enable_claude_oauth_system_prompt_injection: false,
      claude_oauth_system_prompt_blocks: blocks,
    });

    const wrapper = mountView();

    await flushPromises();
    await wrapper.find("form").trigger("submit.prevent");
    await flushPromises();

    expect(updateSettings).toHaveBeenCalledTimes(1);
    expect(updateSettings).toHaveBeenCalledWith(
      expect.objectContaining({
        enable_claude_oauth_system_prompt_injection: false,
      }),
    );
    const payload = updateSettings.mock.calls[0][0] as {
      claude_oauth_system_prompt_blocks: string;
    };
    expect(JSON.parse(payload.claude_oauth_system_prompt_blocks)).toEqual([
      {
        enabled: true,
        type: "text",
        text: "custom block",
        cache_control: {
          type: "ephemeral",
          ttl: "5m",
        },
      },
    ]);
  });

  it("submits Antigravity user agent version gateway setting", async () => {
    getSettings.mockResolvedValueOnce({
      ...baseSettingsResponse,
      antigravity_user_agent_version: "1.23.2",
    });

    const wrapper = mountView();

    await flushPromises();
    await wrapper.find("form").trigger("submit.prevent");
    await flushPromises();

    expect(updateSettings).toHaveBeenCalledTimes(1);
    expect(updateSettings).toHaveBeenCalledWith(
      expect.objectContaining({
        antigravity_user_agent_version: "1.23.2",
      }),
    );
  });

  it("renders advanced scheduler copy as local experimental gateway policy", async () => {
    const wrapper = mountView();

    await flushPromises();

    expect(wrapper.text()).toContain("OpenAI 实验调度策略");
    expect(wrapper.text()).toContain(
      "默认关闭。开启后仅影响本网关在 OpenAI 账号间的实验性调度选择逻辑",
    );
    expect(wrapper.text()).not.toContain("OpenAI 高级调度器");
  });

  it("summarizes target and other-model actions, then switches to all models", async () => {
    getSettings.mockResolvedValueOnce({
      ...baseSettingsResponse,
      openai_fast_policy_settings: {
        rules: [
          {
            service_tier: "all",
            action: "filter",
            scope: "all",
            model_whitelist: ["gpt-5.6-sol"],
            fallback_action: "pass",
          },
        ],
      },
    });
    const wrapper = mountView();

    await flushPromises();
    await openGatewayTab(wrapper);

    const summary = wrapper.get('[data-testid="openai-fast-policy-summary-0"]');
    expect(summary.text()).toContain("目标模型");
    expect(summary.text()).toContain("过滤");
    expect(summary.text()).toContain("其他模型");
    expect(summary.text()).toContain("透传");

    await wrapper
      .get(
        '[role="group"][aria-labelledby="openai-fast-policy-models-label-0"] input[type="text"]',
      )
      .setValue("");
    expect(summary.text()).toContain("全部模型");
    expect(summary.text()).toContain("过滤");
    expect(summary.text()).not.toContain("其他模型");
    expect(summary.text()).not.toContain("透传");
  });

  it("loads and saves upstream billing probe settings from the gateway tab", async () => {
    getUpstreamBillingProbeSettings.mockResolvedValueOnce({
      enabled: false,
      interval_minutes: 45,
    });

    const wrapper = mountView();

    await flushPromises();
    await openGatewayTab(wrapper);

    const card = wrapper.get('[data-testid="upstream-billing-probe-settings"]');
    expect(card.isVisible()).toBe(true);
    expect(card.text()).toContain("上游倍率自动探测");
    expect(
      (card.get('[data-testid="upstream-billing-probe-enabled"]').element as HTMLInputElement)
        .checked,
    ).toBe(false);
    expect(card.find('[data-testid="upstream-billing-probe-interval"]').exists()).toBe(false);

    await card.get('[data-testid="upstream-billing-probe-enabled"]').setValue(true);
    await card.get('[data-testid="upstream-billing-probe-interval"]').setValue(60);
    await card.get('[data-testid="upstream-billing-probe-save"]').trigger("click");
    await flushPromises();

    expect(updateUpstreamBillingProbeSettings).toHaveBeenCalledWith({
      enabled: true,
      interval_minutes: 60,
    });
    expect(showSuccess).toHaveBeenCalledWith("上游倍率自动探测设置已保存");
  });

  it("loads and saves configurable Grok cross-client model mapping", async () => {
    getSettings.mockResolvedValueOnce({
      ...baseSettingsResponse,
      grok_default_text_model: "grok-4.1-fast",
      grok_cross_client_model_map_enabled: true,
    });
    const wrapper = mountView();

    await flushPromises();
    await openGatewayTab(wrapper);

    const modelInput = wrapper.get('[data-testid="grok-default-text-model"]');
    const mappingToggle = wrapper.get(
      '[data-testid="grok-cross-client-model-map-toggle"]',
    );
    expect((modelInput.element as HTMLInputElement).value).toBe("grok-4.1-fast");
    expect((mappingToggle.element as HTMLInputElement).checked).toBe(true);

    await modelInput.setValue("grok-custom-text");
    await mappingToggle.setValue(false);
    await wrapper.find("form").trigger("submit.prevent");
    await flushPromises();

    const payload = updateSettings.mock.calls.at(-1)?.[0] as Record<string, unknown>;
    expect(payload.grok_default_text_model).toBe("grok-custom-text");
    expect(payload.grok_cross_client_model_map_enabled).toBe(false);
  });

  it("loads and saves the OpenAI Responses first-token metric mode", async () => {
    getSettings.mockResolvedValueOnce({
      ...baseSettingsResponse,
      openai_ttft_mode: "visible",
    });
    const wrapper = mountView();

    await flushPromises();
    await openGatewayTab(wrapper);

    const modeSelect = wrapper.get('[data-testid="openai-ttft-mode"]');
    expect((modeSelect.element as HTMLSelectElement).value).toBe("visible");

    await modeSelect.setValue("semantic");
    await wrapper.find("form").trigger("submit.prevent");
    await flushPromises();

    const payload = updateSettings.mock.calls.at(-1)?.[0] as Record<string, unknown>;
    expect(payload.openai_ttft_mode).toBe("semantic");
  });

  it("loads fail-safe-off Ollama Cloud usage refresh settings and saves an explicit opt-in", async () => {
    const wrapper = mountView();

    await flushPromises();
    await openGatewayTab(wrapper);

    const card = wrapper.get('[data-testid="ollama-cloud-usage-global-settings"]');
    expect(card.isVisible()).toBe(true);
    expect(
      (card.get('[data-testid="ollama-cloud-usage-global-enabled"]').element as HTMLInputElement)
        .checked,
    ).toBe(false);
    expect(card.find('[data-testid="ollama-cloud-usage-global-interval"]').exists()).toBe(false);

    await card.get('[data-testid="ollama-cloud-usage-global-enabled"]').setValue(true);
    await card.get('[data-testid="ollama-cloud-usage-global-debounce"]').setValue(3);
    await card.get('[data-testid="ollama-cloud-usage-global-interval"]').setValue(90);
    await card.get('[data-testid="ollama-cloud-usage-global-save"]').trigger("click");
    await flushPromises();

    expect(updateOllamaCloudUsageSettings).toHaveBeenCalledWith({
      enabled: true,
      interval_minutes: 90,
      debounce_minutes: 3,
    });
  });

  it.each([false, true])("clears the OAuth rate without losing zero (weighted=%s)", async (weighted) => {
    getSettings.mockResolvedValueOnce({
      ...baseSettingsResponse,
      openai_low_upstream_rate_priority_enabled: !weighted,
      openai_advanced_scheduler_enabled: weighted,
      openai_oauth_scheduling_rate_multiplier: 0.7,
    });
    const wrapper = mountView();
    await flushPromises();
    const input = wrapper.get('[data-testid="openai-oauth-scheduling-rate-multiplier"]');
    expect(input.attributes("required")).toBeUndefined();
    await input.setValue("");
    await wrapper.find("form").trigger("submit.prevent");
    await flushPromises();
    expect(updateSettings).toHaveBeenLastCalledWith(expect.objectContaining({
      openai_oauth_scheduling_rate_multiplier: null,
    }));
    await input.setValue("0");
    await wrapper.find("form").trigger("submit.prevent");
    await flushPromises();
    expect(updateSettings).toHaveBeenLastCalledWith(expect.objectContaining({
      openai_oauth_scheduling_rate_multiplier: 0,
    }));
    updateSettings.mockClear();
    await input.setValue("-1");
    await wrapper.find("form").trigger("submit.prevent");
    await flushPromises();
    expect(updateSettings).not.toHaveBeenCalled();
    expect(showError).toHaveBeenCalledWith("OAuth 调度参考倍率必须是非负数字，或留空以使用账号倍率。");
  });

  it("loads and preserves an explicitly cleared OAuth rate", async () => {
    getSettings.mockResolvedValueOnce({
      ...baseSettingsResponse,
      openai_advanced_scheduler_enabled: true,
      openai_oauth_scheduling_rate_multiplier: null,
    });
    const wrapper = mountView();
    await flushPromises();
    const input = wrapper.get<HTMLInputElement>('[data-testid="openai-oauth-scheduling-rate-multiplier"]');
    expect(input.element.value).toBe("");
    await wrapper.find("form").trigger("submit.prevent");
    await flushPromises();
    expect(updateSettings).toHaveBeenCalledWith(expect.objectContaining({
      openai_oauth_scheduling_rate_multiplier: null,
    }));
  });

  it("places and explains rate controls for both scheduling modes", async () => {
    const wrapper = mountView();

    await flushPromises();
    expect(
      wrapper.find('[data-testid="openai-oauth-scheduling-rate-multiplier"]').exists(),
    ).toBe(false);

    const lowRateToggle = wrapper.get('[data-testid="openai-low-rate-priority-toggle"]');
    await lowRateToggle.setValue(true);
    const priorityModeText = wrapper.text();
    expect(priorityModeText).toContain(
      "OAuth 账号按此参考倍率参与低倍率优先排序；留空时使用各自的账号倍率。",
    );
    expect(priorityModeText.indexOf("低倍率优先")).toBeLessThan(
      priorityModeText.indexOf("OAuth 调度参考倍率"),
    );
    expect(priorityModeText.indexOf("OAuth 调度参考倍率")).toBeLessThan(
      priorityModeText.indexOf("OpenAI 实验调度策略"),
    );

    const oauthRateInput = wrapper.get(
      '[data-testid="openai-oauth-scheduling-rate-multiplier"]',
    );
    await oauthRateInput.setValue("0.05");
    await wrapper.find("form").trigger("submit.prevent");
    await flushPromises();

    expect(updateSettings).toHaveBeenCalledWith(
      expect.objectContaining({
        openai_low_upstream_rate_priority_enabled: true,
        openai_oauth_scheduling_rate_multiplier: 0.05,
      }),
    );

    await wrapper
      .get('[data-testid="openai-advanced-scheduler-toggle"]')
      .setValue(true);
    expect(
      wrapper.find('[data-testid="openai-low-rate-priority-toggle"]').exists(),
    ).toBe(false);
    expect(
      wrapper.find('[data-testid="openai-oauth-scheduling-rate-multiplier"]').exists(),
    ).toBe(true);
    const weightedModeText = wrapper.text();
    expect(weightedModeText).toContain(
      "计算“计费倍率”得分时，OAuth 账号使用此参考倍率；留空时使用各自的账号倍率。",
    );
    expect(weightedModeText).not.toContain(
      "OAuth 账号按此参考倍率参与低倍率优先排序；",
    );
    expect(weightedModeText.indexOf("订阅优先")).toBeLessThan(
      weightedModeText.indexOf("OAuth 调度参考倍率"),
    );
    expect(weightedModeText.indexOf("OAuth 调度参考倍率")).toBeLessThan(
      weightedModeText.indexOf("调度权值覆盖"),
    );
    expect(weightedModeText).toContain("计费倍率");
  });
});
