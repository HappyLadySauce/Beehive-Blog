import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { StudioSettingsPage } from "./StudioSettingsPage";

const getSettings = vi.hoisted(() => vi.fn());
const patchSettings = vi.hoisted(() => vi.fn());
const testEmailSettings = vi.hoisted(() => vi.fn());

vi.mock("@/lib/api/settings", () => ({
  getSettings,
  patchSettings,
  testEmailSettings
}));

const baseSettings = {
  revision: 5,
  email: {
    enabled: true,
    host: "smtp.example.com",
    port: 587,
    username: "robot",
    password_set: true,
    from: "robot@example.com",
    from_name: "Beehive",
    tls: "starttls"
  }
};

describe("StudioSettingsPage", () => {
  beforeEach(() => {
    getSettings.mockReset();
    patchSettings.mockReset();
    testEmailSettings.mockReset();
    getSettings.mockResolvedValue(baseSettings);
    patchSettings.mockResolvedValue({ ...baseSettings, revision: 6 });
    testEmailSettings.mockResolvedValue({ recipient: "reader@example.com" });
  });

  it("loads and renders SMTP settings", async () => {
    render(<StudioSettingsPage />);

    expect(screen.getByText("正在加载设置...")).toBeInTheDocument();
    await waitFor(() => expect(screen.getByLabelText("SMTP Host")).toHaveValue("smtp.example.com"));
    expect(screen.getByText("Revision 5")).toBeInTheDocument();
    expect(screen.getByText("Password set")).toBeInTheDocument();
  });

  it("saves visible fields without sending password by default", async () => {
    render(<StudioSettingsPage />);
    await waitFor(() => expect(screen.getByLabelText("SMTP Host")).toHaveValue("smtp.example.com"));

    fireEvent.change(screen.getByLabelText("发件人名称"), { target: { value: "Beehive Mailer" } });
    fireEvent.click(screen.getByRole("button", { name: "保存设置" }));

    await waitFor(() => expect(patchSettings).toHaveBeenCalled());
    expect(patchSettings.mock.calls[0][0].email).toMatchObject({ from_name: "Beehive Mailer" });
    expect(patchSettings.mock.calls[0][0].email).not.toHaveProperty("password");
    expect(await screen.findByText("设置已保存。")).toBeInTheDocument();
  });

  it("validates enabled SMTP before saving", async () => {
    render(<StudioSettingsPage />);
    await waitFor(() => expect(screen.getByLabelText("SMTP Host")).toHaveValue("smtp.example.com"));

    fireEvent.change(screen.getByLabelText("SMTP Host"), { target: { value: "" } });
    fireEvent.click(screen.getByRole("button", { name: "保存设置" }));

    expect(await screen.findByRole("alert")).toHaveTextContent("启用 SMTP 时必须填写 host。");
    expect(patchSettings).not.toHaveBeenCalled();
  });

  it("sends an empty password when clearing the stored password", async () => {
    render(<StudioSettingsPage />);
    await waitFor(() => expect(screen.getByLabelText("SMTP Host")).toHaveValue("smtp.example.com"));

    fireEvent.click(screen.getByRole("button", { name: "清空密码" }));
    fireEvent.click(screen.getByRole("button", { name: "保存设置" }));

    await waitFor(() => expect(patchSettings).toHaveBeenCalled());
    expect(patchSettings.mock.calls[0][0].email.password).toBe("");
  });

  it("sends a SMTP test email to the requested recipient", async () => {
    render(<StudioSettingsPage />);
    await waitFor(() => expect(screen.getByLabelText("测试收件人")).toHaveValue("robot@example.com"));

    fireEvent.change(screen.getByLabelText("测试收件人"), { target: { value: "reader@example.com" } });
    fireEvent.click(screen.getByRole("button", { name: "发送测试邮件" }));

    await waitFor(() => expect(testEmailSettings).toHaveBeenCalledWith({ recipient: "reader@example.com" }));
    expect(await screen.findByText("测试邮件已发送至 reader@example.com。")).toBeInTheDocument();
  });

  it("validates test recipient before sending SMTP test email", async () => {
    render(<StudioSettingsPage />);
    await waitFor(() => expect(screen.getByLabelText("测试收件人")).toHaveValue("robot@example.com"));

    fireEvent.change(screen.getByLabelText("测试收件人"), { target: { value: "invalid-email" } });
    fireEvent.click(screen.getByRole("button", { name: "发送测试邮件" }));

    expect(await screen.findByRole("alert")).toHaveTextContent("测试收件人邮箱格式不正确。");
    expect(testEmailSettings).not.toHaveBeenCalled();
  });

  it("shows a safe error state when settings cannot be loaded", async () => {
    getSettings.mockRejectedValue(new Error("network failed"));

    render(<StudioSettingsPage />);

    expect(await screen.findByRole("alert")).toHaveTextContent("设置加载失败");
    expect(screen.getByRole("button", { name: "保存设置" })).toBeDisabled();
    expect(screen.queryByLabelText("SMTP Host")).not.toBeInTheDocument();
  });
});
