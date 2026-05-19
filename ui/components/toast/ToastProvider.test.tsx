import { act, fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { ToastProvider, useToast } from "./ToastProvider";

function ToastHarness() {
  const toast = useToast();
  return (
    <>
      <button type="button" onClick={() => toast.success("保存成功")}>
        成功
      </button>
      <button type="button" onClick={() => toast.error("保存失败")}>
        错误
      </button>
    </>
  );
}

describe("ToastProvider", () => {
  it("auto dismisses success messages", () => {
    vi.useFakeTimers();
    render(
      <ToastProvider>
        <ToastHarness />
      </ToastProvider>
    );

    fireEvent.click(screen.getByRole("button", { name: "成功" }));
    expect(screen.getByRole("status")).toHaveTextContent("保存成功");

    act(() => {
      vi.advanceTimersByTime(3000);
    });
    expect(screen.queryByText("保存成功")).not.toBeInTheDocument();
    vi.useRealTimers();
  });

  it("keeps error messages until manually dismissed", () => {
    vi.useFakeTimers();
    render(
      <ToastProvider>
        <ToastHarness />
      </ToastProvider>
    );

    fireEvent.click(screen.getByRole("button", { name: "错误" }));
    expect(screen.getByRole("alert")).toHaveTextContent("保存失败");

    act(() => {
      vi.advanceTimersByTime(3000);
    });
    expect(screen.getByText("保存失败")).toBeInTheDocument();

    fireEvent.click(screen.getByRole("button", { name: "关闭消息" }));
    expect(screen.queryByText("保存失败")).not.toBeInTheDocument();
    vi.useRealTimers();
  });
});
