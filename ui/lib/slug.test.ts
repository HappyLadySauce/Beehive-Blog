import { describe, expect, it } from "vitest";

import { slugifyContentTitle, slugifyFromName, slugifyTaxonomyName } from "./slug";

describe("slugifyFromName", () => {
  it("converts English to kebab-case", () => {
    expect(slugifyFromName("Hello World")).toBe("hello-world");
  });

  it("converts Chinese to pinyin segments", () => {
    const slug = slugifyFromName("AI 协作写作");
    expect(slug).toMatch(/^[a-z0-9]+(-[a-z0-9]+)*$/);
    expect(slug).toContain("ai");
    expect(slug.length).toBeGreaterThan(0);
  });

  it("handles mixed Chinese and English", () => {
    const slug = slugifyFromName("Go 语言入门");
    expect(slug).toMatch(/^[a-z0-9]+(-[a-z0-9]+)*$/);
    expect(slug).toContain("go");
  });

  it("returns draft fallback for empty or punctuation-only names", () => {
    expect(slugifyFromName("")).toMatch(/^draft-[a-z0-9]+$/);
    expect(slugifyFromName("   ")).toMatch(/^draft-[a-z0-9]+$/);
    expect(slugifyFromName("!!!")).toMatch(/^draft-[a-z0-9]+$/);
  });

  it("truncates to maxLength", () => {
    const longName = "中".repeat(40);
    const slug = slugifyFromName(longName, { maxLength: 64 });
    expect(slug.length).toBeLessThanOrEqual(64);
    expect(slug).toMatch(/^[a-z0-9]+(-[a-z0-9]+)*$/);
  });

  it("defaults taxonomy helper to 64 characters", () => {
    const longName = "测".repeat(50);
    expect(slugifyTaxonomyName(longName).length).toBeLessThanOrEqual(64);
  });

  it("allows longer slugs for content titles", () => {
    const longEnglish = "word ".repeat(80).trim();
    const slug = slugifyContentTitle(longEnglish);
    expect(slug.length).toBeLessThanOrEqual(200);
    expect(slug.startsWith("word")).toBe(true);
  });
});
