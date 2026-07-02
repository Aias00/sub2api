import { readFileSync } from "node:fs";
import { resolve } from "node:path";
import { describe, expect, it } from "vitest";

import {
  buildCommercialLoginAgreementDocuments,
  buildPrivacyPolicyLoginAgreementDocument,
  mergePrivacyPolicyIntoLoginAgreementDocuments,
  renderLoginAgreementDocumentContent,
} from "../loginAgreementTemplates";

const templateSource = readFileSync(
  resolve(process.cwd(), "src/utils/loginAgreementTemplates.ts"),
  "utf8",
);

describe("loginAgreementTemplates", () => {
  it("builds a commercial legal document bundle with site context", () => {
    const docs = buildCommercialLoginAgreementDocuments({
      siteName: "cloudbase",
      frontendUrl: "https://cloudbase.eu.org/",
      contactInfo: "support@cloudbase.eu.org",
      effectiveDate: "2026-05-18",
    });

    expect(docs).toHaveLength(5);
    expect(docs[0]).toMatchObject({
      id: "terms",
      title: "商业服务条款",
    });
    expect(docs[1]).toMatchObject({
      id: "privacy-policy",
      title: "隐私条款",
    });
    expect(docs[0].content_md).toContain("cloudbase");
    expect(docs[0].content_md).toContain("{{site_url}}");
    expect(docs[0].content_md).toContain("{{contact_info}}");
    expect(docs[0].content_md).toContain("{{effective_date}}");
    expect(docs[0].content_md).toContain("{{updated_date}}");
    expect(docs[2].title).toBe("使用政策");
    expect(docs[3].id).toBe("supported-regions");
    expect(docs[4].id).toBe("service-specific-terms");
    expect(docs[3].content_md).toContain("## 2. 当前一般支持的国家和地区");
    expect(docs[3].content_md).toContain("北美");
    expect(docs[3].content_md).toContain("{{contact_info}}");
  });

  it("does not inject a local default site name into commercial templates", () => {
    const privacy = buildPrivacyPolicyLoginAgreementDocument();

    expect(privacy.content_md).not.toContain("Cloudbase");
    expect(privacy.content_md).toContain("本隐私条款说明  在您访问网站");
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
    expect(rendered).toContain("当前站点地址：");
    expect(rendered).not.toContain("以您当前访问本服务时所使用的域名为准");
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

    const privacyPolicy = renderLoginAgreementDocumentContent(
      "- 如您对本隐私条款、数据处理方式、信息访问、更正、删除、导出或申诉有疑问，请联系：${contactInfo}",
      {
        documentId: "privacy-policy",
        updatedAt: "2026-05-18",
        contactInfo: "support@cloudbase.eu.org",
      },
    );
    expect(privacyPolicy).toContain("support@cloudbase.eu.org");
  });

  it("does not synthesize local defaults for missing dynamic document settings", () => {
    const rendered = renderLoginAgreementDocumentContent(
      [
        "**生效日期：** {{effective_date}}",
        "**最后更新：** {{updated_date}}",
        "当前站点地址：{{site_url}}",
        "- 联系方式：{{contact_info}}",
      ].join("\n"),
      { documentId: "terms" },
    );

    expect(rendered).toContain("**生效日期：** ");
    expect(rendered).toContain("**最后更新：** ");
    expect(rendered).toContain("当前站点地址：");
    expect(rendered).toContain("- 联系方式：");
    expect(rendered).not.toContain("2026-03-31");
    expect(rendered).not.toContain("请通过站点设置中的客服联系方式与运营方联系。");
    expect(rendered).not.toContain("以您当前访问本服务时所使用的域名为准");
    expect(templateSource).not.toContain('"2026-03-31"');
    expect(templateSource).not.toContain("请通过站点设置中的客服联系方式与运营方联系。");
    expect(templateSource).not.toContain("以您当前访问本服务时所使用的域名为准");
  });

  it("appends privacy policy without changing existing configured documents", () => {
    const existing = [
      { id: "terms", title: "商业服务条款", content_md: "terms-body" },
      { id: "usage-policy", title: "使用政策", content_md: "usage-body" },
      { id: "supported-regions", title: "支持的国家和地区", content_md: "regions-body" },
      { id: "service-specific-terms", title: "服务特定条款", content_md: "service-body" },
    ];

    const merged = mergePrivacyPolicyIntoLoginAgreementDocuments(existing, {
      siteName: "cloudbase",
      contactInfo: "support@cloudbase.eu.org",
      effectiveDate: "2026-05-18",
    });

    expect(merged).toHaveLength(5);
    expect(merged[0]).toEqual(existing[0]);
    expect(merged[1]).toMatchObject({
      id: "privacy-policy",
      title: "隐私条款",
    });
    expect(merged[2]).toEqual(existing[1]);
    expect(merged[3]).toEqual(existing[2]);
    expect(merged[4]).toEqual(existing[3]);
  });

  it("does not duplicate an existing privacy policy document", () => {
    const privacy = buildPrivacyPolicyLoginAgreementDocument({
      siteName: "cloudbase",
      contactInfo: "support@cloudbase.eu.org",
      effectiveDate: "2026-05-18",
    });
    const existing = [
      { id: "terms", title: "商业服务条款", content_md: "terms-body" },
      privacy,
      { id: "usage-policy", title: "使用政策", content_md: "usage-body" },
    ];

    const merged = mergePrivacyPolicyIntoLoginAgreementDocuments(existing, {
      siteName: "cloudbase",
      contactInfo: "support@cloudbase.eu.org",
      effectiveDate: "2026-05-18",
    });

    expect(merged).toEqual(existing);
  });
});
