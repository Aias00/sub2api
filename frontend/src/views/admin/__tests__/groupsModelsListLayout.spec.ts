import { readFileSync } from "node:fs";
import { fileURLToPath } from "node:url";
import { dirname, resolve } from "node:path";

import { describe, expect, it } from "vitest";

const currentDir = dirname(fileURLToPath(import.meta.url));
const groupsViewSource = readFileSync(
  resolve(currentDir, "../GroupsView.vue"),
  "utf8",
);

describe("groups models list layout", () => {
  it("keeps the toolbar outside of the scrolling list content", () => {
    expect(groupsViewSource).toContain("overflow-hidden rounded-lg border");
    expect(groupsViewSource).toContain("max-h-64 space-y-2 overflow-y-auto p-2");
    expect(groupsViewSource).not.toContain("sticky top-0");
  });

  it("does not synthesize missing subscription type as standard when editing groups", () => {
    expect(groupsViewSource).not.toContain('group.subscription_type || "standard"');
    expect(groupsViewSource).not.toContain("group.subscription_type || 'standard'");
    expect(groupsViewSource).toContain('editForm.subscription_type = group.subscription_type || ""');
    expect(groupsViewSource).toContain("subscription_type: editForm.subscription_type || undefined");
  });
});
