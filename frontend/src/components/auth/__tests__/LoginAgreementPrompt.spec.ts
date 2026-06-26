import { describe, expect, it } from "vitest";
import { mount } from "@vue/test-utils";

import LoginAgreementPrompt from "../LoginAgreementPrompt.vue";
import type { AuthShellLabels } from "@/utils/authShell";

describe("LoginAgreementPrompt", () => {
  const documents = [
    { id: "terms", title: "商业服务条款", content_md: "content" },
    { id: "usage-policy", title: "使用政策", content_md: "content" },
  ];
  const shellLabels: AuthShellLabels = {
    agreementAcceptedDescription: "配置已同意说明",
    agreementAcceptedTitle: "配置登录条款入口",
    agreementReviewDescription: "配置待确认说明",
    agreementReviewTitle: "配置待确认标题",
    agreementViewAndAccept: "配置查看并同意",
    agreementViewTerms: "配置查看条款",
  };

  it("keeps the modal-mode entry visible after acceptance", () => {
    const wrapper = mount(LoginAgreementPrompt, {
      props: {
        accepted: true,
        documents,
        mode: "modal",
        shellLabels,
        updatedAt: "2026-05-18",
        visible: false,
      },
      global: {
        stubs: {
          RouterLink: {
            props: ["to"],
            template: "<a :href=\"typeof to === 'string' ? to : '/legal/' + to.params.documentId\"><slot /></a>",
          },
          Icon: true,
          Teleport: true,
          Transition: false,
        },
      },
    });

    expect(wrapper.text()).toContain("配置登录条款入口");
    expect(wrapper.text()).toContain("配置已同意说明");
    expect(wrapper.text()).toContain("配置查看条款");
  });

  it("shows the stronger gate copy before acceptance in modal mode", () => {
    const wrapper = mount(LoginAgreementPrompt, {
      props: {
        accepted: false,
        documents,
        mode: "modal",
        shellLabels,
        updatedAt: "2026-05-18",
        visible: false,
      },
      global: {
        stubs: {
          RouterLink: {
            props: ["to"],
            template: "<a :href=\"typeof to === 'string' ? to : '/legal/' + to.params.documentId\"><slot /></a>",
          },
          Icon: true,
          Teleport: true,
          Transition: false,
        },
      },
    });

    expect(wrapper.text()).toContain("配置待确认标题");
    expect(wrapper.text()).toContain("配置待确认说明");
    expect(wrapper.text()).toContain("配置查看并同意");
  });
});
