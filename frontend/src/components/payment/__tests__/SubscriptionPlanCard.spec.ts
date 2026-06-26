import { readFileSync } from "node:fs";
import { mount } from "@vue/test-utils";
import { describe, expect, it } from "vitest";
import { createI18n } from "vue-i18n";
import SubscriptionPlanCard from "../SubscriptionPlanCard.vue";

const i18n = createI18n({
  legacy: false,
  locale: "en",
  fallbackWarn: false,
  missingWarn: false,
  messages: {
    en: {
      payment: {
        days: "days",
        models: "Models",
        planCard: {
          quota: "Quota",
          rate: "Rate",
          unlimited: "Unlimited",
        },
        subscribeNow: "Subscribe now",
      },
    },
  },
});

const subscriptionPlanCardSource = readFileSync("src/components/payment/SubscriptionPlanCard.vue", "utf8");

const mountPlanCard = (groupPlatform: string, groupDisplayLabel?: string) =>
  mount(SubscriptionPlanCard, {
    props: {
      plan: {
        id: 1,
        group_id: 10,
        group_platform: groupPlatform,
        group_display_label: groupDisplayLabel,
        name: "Pro",
        price: 10,
        amount: 1000,
        features: [],
        rate_multiplier: 1,
        validity_days: 30,
        validity_unit: "day",
        supported_model_scopes: ["claude", "gemini_text", "gemini_image"],
        is_active: true,
      },
    },
    global: { plugins: [i18n] },
  });

const mountPlanCardWithLabels = () =>
  mount(SubscriptionPlanCard, {
    props: {
      plan: {
        id: 1,
        group_id: 10,
        group_platform: "openai",
        name: "Pro",
        price: 10,
        amount: 1000,
        features: [],
        rate_multiplier: 1,
        validity_days: 30,
        validity_unit: "day",
        is_active: true,
      },
      labels: {
        rate: "Configured rate",
        quota: "Configured quota",
        unlimited: "Configured unlimited",
        subscribeNow: "Configured subscribe",
        days: " configured days",
      },
      currency: "USD",
      locale: "en-US",
    },
    global: { plugins: [i18n] },
  });

describe("SubscriptionPlanCard", () => {
  it("does not show Antigravity model scopes for OpenAI plans", () => {
    const text = mountPlanCard("openai").text();

    expect(text).not.toContain("Claude");
    expect(text).not.toContain("Gemini");
    expect(text).not.toContain("Imagen");
  });

  it("shows model scopes for Antigravity plans", () => {
    const text = mountPlanCard("antigravity").text();

    expect(text).toContain("Claude");
    expect(text).toContain("Gemini");
    expect(text).toContain("Imagen");
  });

  it("uses the backend-provided group display label when present", () => {
    const text = mountPlanCard("openai", "Configured GPT").text();

    expect(text).toContain("Configured GPT");
    expect(text).not.toContain("OpenAI");
  });

  it("uses configured shell labels when provided by the parent payment page", () => {
    const text = mountPlanCardWithLabels().text();

    expect(text).toContain("Configured rate");
    expect(text).toContain("Configured quota");
    expect(text).toContain("Configured unlimited");
    expect(text).toContain("Configured subscribe");
    expect(text).toContain("30 configured days");
  });

  it("does not synthesize a day validity suffix when unit data is missing", () => {
    const wrapper = mount(SubscriptionPlanCard, {
      props: {
        plan: {
          id: 1,
          group_id: 10,
          group_platform: "openai",
          name: "Pro",
          price: 10,
          amount: 1000,
          features: [],
          rate_multiplier: 1,
          validity_days: 30,
          validity_unit: "",
          is_active: true,
        },
        labels: {
          days: " configured days",
          subscribeNow: "Configured subscribe",
        },
      },
      global: { plugins: [i18n] },
    });

    expect(wrapper.text()).not.toContain("30 configured days");
    expect(wrapper.text()).not.toContain("/ 30");
  });

  it("does not synthesize an API platform label when platform data is missing", () => {
    const wrapper = mount(SubscriptionPlanCard, {
      props: {
        plan: {
          id: 1,
          group_id: 10,
          group_platform: "",
          name: "Pro",
          price: 10,
          amount: 1000,
          features: [],
          rate_multiplier: 1,
          validity_days: 30,
          validity_unit: "day",
          is_active: true,
        },
        labels: {
          days: " configured days",
          subscribeNow: "Configured subscribe",
        },
      },
      global: { plugins: [i18n] },
    });

    expect(wrapper.text()).not.toContain("API");
  });

  it("formats plan prices from the parent payment currency instead of hard-coded dollars", () => {
    const wrapper = mount(SubscriptionPlanCard, {
      props: {
        plan: {
          id: 1,
          group_id: 10,
          group_platform: "openai",
          name: "Pro",
          price: 128,
          original_price: 168,
          features: [],
          rate_multiplier: 1,
          daily_limit_usd: 20,
          validity_days: 30,
          validity_unit: "day",
          for_sale: true,
          sort_order: 1,
          description: "",
        },
        currency: "EUR",
        locale: "en-US",
      },
      global: { plugins: [i18n] },
    });

    expect(wrapper.text()).toContain("€128.00");
    expect(wrapper.text()).toContain("€168.00");
    expect(wrapper.text()).toContain("$20.00");
  });

  it("does not render payment price or quota limits with template-level dollar prefixes", () => {
    expect(subscriptionPlanCardSource).not.toContain("> $</span>");
    expect(subscriptionPlanCardSource).not.toContain("${{ plan.original_price }}");
    expect(subscriptionPlanCardSource).not.toContain("${{ plan.daily_limit_usd }}");
    expect(subscriptionPlanCardSource).not.toContain("${{ plan.weekly_limit_usd }}");
    expect(subscriptionPlanCardSource).not.toContain("${{ plan.monthly_limit_usd }}");
  });

  it("does not carry local subscription plan i18n fallbacks in the component", () => {
    expect(subscriptionPlanCardSource).not.toContain("payment.planCard.rate");
    expect(subscriptionPlanCardSource).not.toContain("payment.subscribeNow");
    expect(subscriptionPlanCardSource).not.toContain("payment.days");
    expect(subscriptionPlanCardSource).not.toContain("useI18n");
    expect(subscriptionPlanCardSource).not.toContain("labels?.rate || 'rate'");
    expect(subscriptionPlanCardSource).not.toContain("labels?.subscribeNow || 'subscribeNow'");
    expect(subscriptionPlanCardSource).not.toContain("props.labels?.days || 'days'");
    expect(subscriptionPlanCardSource).not.toContain("group_platform || ''");
    expect(subscriptionPlanCardSource).not.toContain("validity_unit || 'day'");
  });
});
