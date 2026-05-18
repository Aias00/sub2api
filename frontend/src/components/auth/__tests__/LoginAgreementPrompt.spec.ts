import { describe, expect, it } from "vitest";
import { mount } from "@vue/test-utils";

import LoginAgreementPrompt from "../LoginAgreementPrompt.vue";

describe("LoginAgreementPrompt", () => {
  const documents = [
    { id: "terms", title: "商业服务条款", content_md: "content" },
    { id: "usage-policy", title: "使用政策", content_md: "content" },
  ];

  it("keeps the modal-mode entry visible after acceptance", () => {
    const wrapper = mount(LoginAgreementPrompt, {
      props: {
        accepted: true,
        documents,
        mode: "modal",
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

    expect(wrapper.text()).toContain("登录条款入口");
    expect(wrapper.text()).toContain("您已同意当前版本条款，可随时重新查看相关文档。");
    expect(wrapper.text()).toContain("查看条款");
  });

  it("shows the stronger gate copy before acceptance in modal mode", () => {
    const wrapper = mount(LoginAgreementPrompt, {
      props: {
        accepted: false,
        documents,
        mode: "modal",
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

    expect(wrapper.text()).toContain("继续登录前需要先同意最新条款。");
    expect(wrapper.text()).toContain("查看并同意");
  });
});
