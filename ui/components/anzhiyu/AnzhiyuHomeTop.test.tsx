import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import { AnzhiyuHomeTop } from "./AnzhiyuHomeTop";

describe("AnzhiyuHomeTop", () => {
  it("uses the current Beehive brand copy when no featured content is available", () => {
    render(<AnzhiyuHomeTop featured={[]} />);

    expect(screen.getByText("Beehive")).toBeInTheDocument();
    expect(screen.getByText("个人博客与知识中台")).toBeInTheDocument();
    expect(screen.getByText("Beehive Blog")).toBeInTheDocument();
    expect(screen.queryByText("生活明朗，万物可爱。")).not.toBeInTheDocument();
  });
});
