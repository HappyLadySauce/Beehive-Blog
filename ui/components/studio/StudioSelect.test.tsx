import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { StudioSelect } from "./StudioSelect";

describe("StudioSelect", () => {
  it("renders a custom combobox and selects options", () => {
    const onChange = vi.fn();

    const { container } = render(
      <StudioSelect
        ariaLabel="默认存储"
        options={[
          { value: "local", label: "Local" },
          { value: "s3", label: "S3" },
          { value: "oss", label: "OSS" }
        ]}
        value="local"
        onChange={onChange}
      />
    );

    expect(container.querySelector("select")).not.toBeInTheDocument();

    const trigger = screen.getByRole("combobox", { name: "默认存储" });
    expect(trigger).toHaveTextContent("Local");

    fireEvent.click(trigger);
    fireEvent.click(screen.getByRole("option", { name: "S3" }));

    expect(onChange).toHaveBeenCalledWith("s3");
  });

  it("accepts layout class names from parent surfaces", () => {
    render(
      <StudioSelect
        ariaLabel="按状态筛选"
        className="compactSelect"
        options={[
          { value: "", label: "全部状态" },
          { value: "active", label: "活跃" }
        ]}
        value=""
        onChange={() => undefined}
      />
    );

    expect(screen.getByRole("combobox", { name: "按状态筛选" }).parentElement).toHaveClass("compactSelect");
  });
});
