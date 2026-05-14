import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { StudioSettingsPage } from "./StudioSettingsPage";

const getSettings = vi.hoisted(() => vi.fn());
const getGithubOAuth2Settings = vi.hoisted(() => vi.fn());
const getAttachmentSettings = vi.hoisted(() => vi.fn());
const patchSettings = vi.hoisted(() => vi.fn());
const patchGithubOAuth2Settings = vi.hoisted(() => vi.fn());
const patchAttachmentSettings = vi.hoisted(() => vi.fn());
const testEmailSettings = vi.hoisted(() => vi.fn());

vi.mock("@/lib/api/settings", () => ({
  getAttachmentSettings,
  getGithubOAuth2Settings,
  getSettings,
  patchAttachmentSettings,
  patchGithubOAuth2Settings,
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
  },
  github_oauth2: {
    enabled: false,
    client_id: "",
    client_secret_set: false,
    redirect_url: "",
    auth_url: "https://github.com/login/oauth/authorize",
    token_url: "https://github.com/login/oauth/access_token",
    user_info_url: "https://api.github.com/user",
    allow_non_github_endpoints: false
  },
  attachment: {
    default_storage: "local",
    local_root: "data/attachments",
    max_bytes: 10485760,
    allowed_mime_prefixes: ["image/", "application/pdf"],
    presign_ttl_seconds: 900,
    s3: { bucket: "", upload_base_url: "", download_base_url: "" },
    oss: { bucket: "", upload_base_url: "", download_base_url: "" }
  }
};

describe("StudioSettingsPage", () => {
  beforeEach(() => {
    getSettings.mockReset();
    getGithubOAuth2Settings.mockReset();
    getAttachmentSettings.mockReset();
    patchSettings.mockReset();
    patchGithubOAuth2Settings.mockReset();
    patchAttachmentSettings.mockReset();
    testEmailSettings.mockReset();
    getSettings.mockResolvedValue(baseSettings);
    getGithubOAuth2Settings.mockResolvedValue(baseSettings);
    getAttachmentSettings.mockResolvedValue(baseSettings);
    patchSettings.mockResolvedValue({ ...baseSettings, revision: 6 });
    patchGithubOAuth2Settings.mockResolvedValue({ ...baseSettings, revision: 6 });
    patchAttachmentSettings.mockResolvedValue({ ...baseSettings, revision: 6 });
    testEmailSettings.mockResolvedValue({ recipient: "reader@example.com" });
  });

  it("loads and renders SMTP settings", async () => {
    render(<StudioSettingsPage />);

    expect(screen.getByText("正在加载设置...")).toBeInTheDocument();
    await waitFor(() => expect(screen.getByLabelText("SMTP Host")).toHaveValue("smtp.example.com"));
    expect(screen.getByText("Revision 5")).toBeInTheDocument();
    expect(screen.getByText("Password set")).toBeInTheDocument();
  });

  it("loads disabled SMTP settings with empty host and from fields", async () => {
    getSettings.mockResolvedValue({
      revision: 6,
      email: {
        enabled: false,
        host: "",
        port: 587,
        username: "",
        password_set: false,
        from: "",
        from_name: "",
        tls: "starttls"
      }
    });

    render(<StudioSettingsPage />);

    await waitFor(() => expect(screen.getByLabelText("SMTP Host")).toHaveValue(""));
    expect(screen.getByLabelText("发件人邮箱")).toHaveValue("");
    expect(screen.getByLabelText("启用 SMTP 邮件发送")).not.toBeChecked();
    expect(screen.getByText("Revision 6")).toBeInTheDocument();
    expect(screen.getByText("No password")).toBeInTheDocument();
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

  it("switches to GitHub OAuth2 settings and renders disabled empty configuration", async () => {
    render(<StudioSettingsPage />);
    await waitFor(() => expect(screen.getByRole("button", { name: "GitHub OAuth2" })).toBeInTheDocument());

    fireEvent.click(screen.getByRole("button", { name: "GitHub OAuth2" }));

    expect(screen.getByLabelText("启用 GitHub OAuth2 登录")).not.toBeChecked();
    expect(screen.getByLabelText("Client ID")).toHaveValue("");
    expect(screen.getByLabelText("Redirect URL")).toHaveValue("");
    expect(screen.getByText("Client secret not set")).toBeInTheDocument();
  });

  it("validates enabled GitHub OAuth2 before saving", async () => {
    render(<StudioSettingsPage />);
    await waitFor(() => expect(screen.getByRole("button", { name: "GitHub OAuth2" })).toBeInTheDocument());
    fireEvent.click(screen.getByRole("button", { name: "GitHub OAuth2" }));
    fireEvent.click(screen.getByLabelText("启用 GitHub OAuth2 登录"));
    fireEvent.click(screen.getByRole("button", { name: "保存设置" }));

    expect(await screen.findByRole("alert")).toHaveTextContent("启用 GitHub OAuth2 时必须填写 Client ID。");
    expect(patchGithubOAuth2Settings).not.toHaveBeenCalled();
  });

  it("saves GitHub OAuth2 without sending client secret by default", async () => {
    getGithubOAuth2Settings.mockResolvedValue({
      ...baseSettings,
      github_oauth2: {
        ...baseSettings.github_oauth2,
        enabled: true,
        client_id: "existing-client",
        client_secret_set: true,
        redirect_url: "http://localhost:3000/auth/github/callback"
      }
    });
    render(<StudioSettingsPage />);
    await waitFor(() => expect(screen.getByRole("button", { name: "GitHub OAuth2" })).toBeInTheDocument());
    fireEvent.click(screen.getByRole("button", { name: "GitHub OAuth2" }));

    fireEvent.change(screen.getByLabelText("Client ID"), { target: { value: "next-client" } });
    fireEvent.click(screen.getByRole("button", { name: "保存设置" }));

    await waitFor(() => expect(patchGithubOAuth2Settings).toHaveBeenCalled());
    expect(patchGithubOAuth2Settings.mock.calls[0][0]).toMatchObject({ client_id: "next-client" });
    expect(patchGithubOAuth2Settings.mock.calls[0][0]).not.toHaveProperty("client_secret");
  });

  it("sends an empty client secret when clearing the stored GitHub secret", async () => {
    getGithubOAuth2Settings.mockResolvedValue({
      ...baseSettings,
      github_oauth2: {
        ...baseSettings.github_oauth2,
        enabled: true,
        client_id: "existing-client",
        client_secret_set: true,
        redirect_url: "http://localhost:3000/auth/github/callback"
      }
    });
    render(<StudioSettingsPage />);
    await waitFor(() => expect(screen.getByRole("button", { name: "GitHub OAuth2" })).toBeInTheDocument());
    fireEvent.click(screen.getByRole("button", { name: "GitHub OAuth2" }));

    fireEvent.click(screen.getByRole("button", { name: "清空 Client Secret" }));
    fireEvent.click(screen.getByRole("button", { name: "保存设置" }));

    await waitFor(() => expect(patchGithubOAuth2Settings).toHaveBeenCalled());
    expect(patchGithubOAuth2Settings.mock.calls[0][0].client_secret).toBe("");
  });

  it("keeps GitHub advanced settings collapsed until expanded", async () => {
    render(<StudioSettingsPage />);
    await waitFor(() => expect(screen.getByRole("button", { name: "GitHub OAuth2" })).toBeInTheDocument());
    fireEvent.click(screen.getByRole("button", { name: "GitHub OAuth2" }));

    expect(screen.queryByLabelText("Auth URL")).not.toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "高级设置" }));
    expect(screen.getByLabelText("Auth URL")).toHaveValue("https://github.com/login/oauth/authorize");
    expect(screen.getByLabelText("允许非 GitHub 端点")).not.toBeChecked();
  });

  it("switches to attachment settings and renders storage defaults", async () => {
    render(<StudioSettingsPage />);
    await waitFor(() => expect(screen.getByRole("button", { name: "Attachments" })).toBeInTheDocument());

    fireEvent.click(screen.getByRole("button", { name: "Attachments" }));

    expect(screen.getByLabelText("默认存储")).toHaveValue("local");
    expect(screen.getByLabelText("本地存储目录")).toHaveValue("data/attachments");
    expect(screen.getByLabelText("最大上传大小 (MB)")).toHaveValue(10);
    expect(screen.getByLabelText("预签名有效期 (秒)")).toHaveValue(900);
    expect(screen.getByLabelText("允许的 MIME 前缀")).toHaveValue("image/\napplication/pdf");
  });

  it("saves attachment settings with parsed MIME prefixes", async () => {
    render(<StudioSettingsPage />);
    await waitFor(() => expect(screen.getByRole("button", { name: "Attachments" })).toBeInTheDocument());
    fireEvent.click(screen.getByRole("button", { name: "Attachments" }));

    fireEvent.change(screen.getByLabelText("默认存储"), { target: { value: "s3" } });
    fireEvent.change(screen.getByLabelText("最大上传大小 (MB)"), { target: { value: "20" } });
    fireEvent.change(screen.getByLabelText("允许的 MIME 前缀"), { target: { value: "image/\ntext/" } });
    fireEvent.click(screen.getByRole("button", { name: "保存设置" }));

    await waitFor(() => expect(patchAttachmentSettings).toHaveBeenCalled());
    expect(patchAttachmentSettings.mock.calls[0][0]).toMatchObject({
      default_storage: "s3",
      max_bytes: 20971520,
      allowed_mime_prefixes: ["image/", "text/"]
    });
    expect(await screen.findByText("附件设置已保存。")).toBeInTheDocument();
  });

  it("keeps attachment remote settings collapsed until expanded", async () => {
    render(<StudioSettingsPage />);
    await waitFor(() => expect(screen.getByRole("button", { name: "Attachments" })).toBeInTheDocument());
    fireEvent.click(screen.getByRole("button", { name: "Attachments" }));

    expect(screen.queryByLabelText("S3 Bucket")).not.toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "远端存储设置" }));
    expect(screen.getByLabelText("S3 Bucket")).toHaveValue("");
    expect(screen.getByLabelText("OSS Bucket")).toHaveValue("");
  });

  it("validates incomplete attachment remote settings before saving", async () => {
    render(<StudioSettingsPage />);
    await waitFor(() => expect(screen.getByRole("button", { name: "Attachments" })).toBeInTheDocument());
    fireEvent.click(screen.getByRole("button", { name: "Attachments" }));
    fireEvent.click(screen.getByRole("button", { name: "远端存储设置" }));
    fireEvent.change(screen.getByLabelText("S3 Upload Base URL"), { target: { value: "https://s3.example.com/upload" } });
    fireEvent.click(screen.getByRole("button", { name: "保存设置" }));

    expect(await screen.findByRole("alert")).toHaveTextContent("S3 Bucket 不能为空。");
    expect(patchAttachmentSettings).not.toHaveBeenCalled();
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
