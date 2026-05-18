import { describe, expect, it } from "vitest";

import {
  buildCommercialLoginAgreementDocuments,
  isLegacyBlankLoginAgreementDocuments,
  renderLoginAgreementDocumentContent,
} from "../loginAgreementTemplates";

describe("loginAgreementTemplates", () => {
  it("builds a commercial legal document bundle with site context", () => {
    const docs = buildCommercialLoginAgreementDocuments({
      siteName: "cloudbase",
      frontendUrl: "https://cloudbase.eu.org/",
      contactInfo: "support@cloudbase.eu.org",
      effectiveDate: "2026-05-18",
    });

    expect(docs).toHaveLength(4);
    expect(docs[0]).toMatchObject({
      id: "terms",
      title: "商业服务条款",
    });
    expect(docs[0].content_md).toContain("cloudbase");
    expect(docs[0].content_md).toContain("{{site_url}}");
    expect(docs[0].content_md).toContain("{{contact_info}}");
    expect(docs[0].content_md).toContain("{{effective_date}}");
    expect(docs[0].content_md).toContain("{{updated_date}}");
    expect(docs[1].title).toBe("使用政策");
    expect(docs[2].id).toBe("supported-regions");
    expect(docs[3].id).toBe("service-specific-terms");
    expect(docs[2].content_md).toContain("## 2. 当前一般支持的国家和地区");
    expect(docs[2].content_md).toContain("北美");
    expect(docs[2].content_md).toContain("{{contact_info}}");
  });

  it("renders placeholders and normalizes legacy dynamic lines", () => {
    const rendered = renderLoginAgreementDocumentContent(
      [
        "**生效日期：** 2026-03-31",
        "**最后更新：** 2026-03-31",
        "当前站点地址：http://127.0.0.1:18082",
        "- 联系方式：旧联系方式",
        "如果您需要确认某个国家、地区或业务类型是否可以接入，请联系：旧联系方式",
        "- 如您需要企业协议、定制账期、发票、白名单、专属支持或其他商业安排，请联系：${contactInfo}",
      ].join("\n"),
      {
        documentId: "terms",
        updatedAt: "2026-05-18",
        frontendUrl: "",
        contactInfo: "support@cloudbase.eu.org",
      },
    );

    expect(rendered).toContain("**生效日期：** 2026-05-18");
    expect(rendered).toContain("**最后更新：** 2026-05-18");
    expect(rendered).toContain("当前站点地址：以您当前访问本服务时所使用的域名为准");
    expect(rendered).toContain("- 联系方式：support@cloudbase.eu.org");

    const serviceTerms = renderLoginAgreementDocumentContent(
      "- 如您需要企业协议、定制账期、发票、白名单、专属支持或其他商业安排，请联系：${contactInfo}",
      {
        documentId: "service-specific-terms",
        updatedAt: "2026-05-18",
        contactInfo: "support@cloudbase.eu.org",
      },
    );
    expect(serviceTerms).toContain("support@cloudbase.eu.org");
    expect(serviceTerms).not.toContain("${contactInfo}");
  });

  it("treats legacy blank documents as replaceable", () => {
    expect(
      isLegacyBlankLoginAgreementDocuments([
        { id: "terms", title: "服务条款", content_md: "" },
        { id: "usage-policy", title: "使用政策", content_md: "" },
        {
          id: "supported-regions",
          title: "支持的国家和地区",
          content_md: "",
        },
        {
          id: "service-specific-terms",
          title: "服务特定条款",
          content_md: "",
        },
      ]),
    ).toBe(true);

    expect(
      isLegacyBlankLoginAgreementDocuments([
        { id: "terms", title: "商业服务条款", content_md: "已有内容" },
      ]),
    ).toBe(false);
  });
});
