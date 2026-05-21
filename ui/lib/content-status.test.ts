import { describe, expect, it } from "vitest";

import { planStatusTransition } from "./content-status";

describe("planStatusTransition", () => {
  it("returns empty steps when status is unchanged", () => {
    expect(planStatusTransition("draft", "draft")).toEqual([]);
  });

  it("plans draft to published via review", () => {
    expect(planStatusTransition("draft", "published")).toEqual(["review", "published"]);
  });

  it("plans published back to draft via archived", () => {
    expect(planStatusTransition("published", "draft")).toEqual(["archived", "draft"]);
  });

  it("allows single-step transitions", () => {
    expect(planStatusTransition("draft", "review")).toEqual(["review"]);
    expect(planStatusTransition("review", "published")).toEqual(["published"]);
  });
});
