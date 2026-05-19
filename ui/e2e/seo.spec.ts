import { expect, test } from "@playwright/test";

test.describe("SEO routes", () => {
  test("robots.txt disallows studio and auth paths", async ({ request }) => {
    const response = await request.get("/robots.txt");
    expect(response.ok()).toBeTruthy();

    const body = await response.text();
    expect(body).toContain("Disallow: /studio");
    expect(body).toContain("Disallow: /login");
    expect(body).toContain("Sitemap:");
  });

  test("sitemap.xml returns XML", async ({ request }) => {
    const response = await request.get("/sitemap.xml");
    expect(response.ok()).toBeTruthy();

    const contentType = response.headers()["content-type"] ?? "";
    expect(contentType).toMatch(/xml/i);

    const body = await response.text();
    expect(body).toContain("<urlset");
  });
});
