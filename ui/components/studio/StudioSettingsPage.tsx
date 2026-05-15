"use client";

import { FormEvent, useEffect, useMemo, useState } from "react";
import { ChevronDown, Github, Loader2, Mail, Save, Send, Settings } from "lucide-react";

import { humanizeApiError } from "@/lib/api/client";
import {
  getGithubOAuth2Settings,
  getSettings,
  patchGithubOAuth2Settings,
  patchSettings,
  testEmailSettings
} from "@/lib/api/settings";
import type { EmailSettingsPublic, GithubOAuth2SettingsPublic, SettingsResponse } from "@/lib/api/types";
import styles from "./Studio.module.css";
import { StudioPanel } from "./StudioPanel";
import { StudioSelect } from "./StudioSelect";
import { StudioTopbar } from "./StudioTopbar";

type PasswordMode = "keep" | "set" | "clear";
type SettingsSection = "email" | "github";

const defaultEmail: EmailSettingsPublic = {
  enabled: false,
  host: "",
  port: 587,
  username: "",
  password_set: false,
  from: "",
  from_name: "",
  tls: "starttls"
};

const defaultGithubOAuth2: GithubOAuth2SettingsPublic = {
  enabled: false,
  client_id: "",
  client_secret_set: false,
  redirect_url: "",
  auth_url: "https://github.com/login/oauth/authorize",
  token_url: "https://github.com/login/oauth/access_token",
  user_info_url: "https://api.github.com/user",
  allow_non_github_endpoints: false
};

const tlsOptions = [
  { value: "none", label: "None" },
  { value: "starttls", label: "STARTTLS" },
  { value: "tls", label: "TLS" }
];

let settingsLoadRequest: Promise<SettingsResponse> | null = null;
let githubSettingsLoadRequest: Promise<SettingsResponse> | null = null;

function loadSettings() {
  if (!settingsLoadRequest) {
    settingsLoadRequest = getSettings().finally(() => {
      settingsLoadRequest = null;
    });
  }
  return settingsLoadRequest;
}

function loadGithubOAuth2Settings() {
  if (!githubSettingsLoadRequest) {
    githubSettingsLoadRequest = getGithubOAuth2Settings().finally(() => {
      githubSettingsLoadRequest = null;
    });
  }
  return githubSettingsLoadRequest;
}

export function StudioSettingsPage() {
  const [settings, setSettings] = useState<SettingsResponse | null>(null);
  const [activeSection, setActiveSection] = useState<SettingsSection>("email");
  const [email, setEmail] = useState<EmailSettingsPublic>(defaultEmail);
  const [githubOAuth2, setGithubOAuth2] = useState<GithubOAuth2SettingsPublic>(defaultGithubOAuth2);
  const [password, setPassword] = useState("");
  const [passwordMode, setPasswordMode] = useState<PasswordMode>("keep");
  const [githubSecret, setGithubSecret] = useState("");
  const [githubSecretMode, setGithubSecretMode] = useState<PasswordMode>("keep");
  const [githubAdvancedOpen, setGithubAdvancedOpen] = useState(false);
  const [testRecipient, setTestRecipient] = useState("");
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [testing, setTesting] = useState(false);
  const [message, setMessage] = useState<{ tone: "success" | "error"; text: string } | null>(null);

  useEffect(() => {
    let active = true;

    Promise.all([loadSettings(), loadGithubOAuth2Settings()])
      .then(([emailPayload, githubPayload]) => {
        if (!active) return;
        const payload = {
          ...emailPayload,
          revision: Math.max(emailPayload.revision, githubPayload.revision),
          github_oauth2: githubPayload.github_oauth2 ?? emailPayload.github_oauth2 ?? defaultGithubOAuth2
        };
        setSettings(payload);
        setEmail(payload.email);
        setGithubOAuth2(payload.github_oauth2 ?? defaultGithubOAuth2);
        setTestRecipient(payload.email.from || payload.email.username);
      })
      .catch((error) => {
        if (!active) return;
        setMessage({ tone: "error", text: humanizeApiError(error) });
      })
      .finally(() => {
        if (active) {
          setLoading(false);
        }
      });

    return () => {
      active = false;
    };
  }, []);

  const passwordHint = useMemo(() => {
    if (passwordMode === "clear") return "保存后会清空当前 SMTP 密码。";
    if (password.trim() !== "") return "保存后会更新 SMTP 密码。";
    return email.password_set ? "当前已设置密码；留空保存不会修改密码。" : "当前未设置密码。";
  }, [email.password_set, password, passwordMode]);

  const githubSecretHint = useMemo(() => {
    if (githubSecretMode === "clear") return "保存后会清空当前 GitHub Client Secret。";
    if (githubSecret.trim() !== "") return "保存后会更新 GitHub Client Secret。";
    return githubOAuth2.client_secret_set ? "当前已设置 Client Secret；留空保存不会修改。" : "当前未设置 Client Secret。";
  }, [githubOAuth2.client_secret_set, githubSecret, githubSecretMode]);

  function updateEmail<K extends keyof EmailSettingsPublic>(key: K, value: EmailSettingsPublic[K]) {
    setEmail((current) => ({ ...current, [key]: value }));
  }

  function updateGithubOAuth2<K extends keyof GithubOAuth2SettingsPublic>(key: K, value: GithubOAuth2SettingsPublic[K]) {
    setGithubOAuth2((current) => ({ ...current, [key]: value }));
  }

  async function onSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const validation = validateEmail(email);
    if (validation) {
      setMessage({ tone: "error", text: validation });
      return;
    }

    setSaving(true);
    setMessage(null);
    try {
      const next = await patchSettings({
        email: {
          enabled: email.enabled,
          host: email.host,
          port: email.port,
          username: email.username,
          from: email.from,
          from_name: email.from_name,
          tls: email.tls,
          ...(passwordMode === "clear" ? { password: "" } : password !== "" ? { password } : {})
        }
      });
      setSettings(next);
      setEmail(next.email);
      setTestRecipient((current) => current || next.email.from || next.email.username);
      setPassword("");
      setPasswordMode("keep");
      setMessage({ tone: "success", text: "设置已保存。" });
    } catch (error) {
      setMessage({ tone: "error", text: humanizeApiError(error) });
    } finally {
      setSaving(false);
    }
  }

  async function onSubmitGithubOAuth2(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const validation = validateGithubOAuth2(githubOAuth2, githubSecret, githubSecretMode);
    if (validation) {
      setMessage({ tone: "error", text: validation });
      return;
    }

    setSaving(true);
    setMessage(null);
    try {
      const next = await patchGithubOAuth2Settings({
        enabled: githubOAuth2.enabled,
        client_id: githubOAuth2.client_id,
        redirect_url: githubOAuth2.redirect_url,
        auth_url: githubOAuth2.auth_url,
        token_url: githubOAuth2.token_url,
        user_info_url: githubOAuth2.user_info_url,
        allow_non_github_endpoints: githubOAuth2.allow_non_github_endpoints,
        ...(githubSecretMode === "clear" ? { client_secret: "" } : githubSecret !== "" ? { client_secret: githubSecret } : {})
      });
      setSettings(next);
      setGithubOAuth2(next.github_oauth2 ?? defaultGithubOAuth2);
      setGithubSecret("");
      setGithubSecretMode("keep");
      setMessage({ tone: "success", text: "GitHub OAuth2 设置已保存。" });
    } catch (error) {
      setMessage({ tone: "error", text: humanizeApiError(error) });
    } finally {
      setSaving(false);
    }
  }

  async function onSendTestEmail() {
    const recipient = testRecipient.trim();
    const validation = validateTestRecipient(email, recipient);
    if (validation) {
      setMessage({ tone: "error", text: validation });
      return;
    }

    setTesting(true);
    setMessage(null);
    try {
      const result = await testEmailSettings({ recipient });
      setMessage({ tone: "success", text: `测试邮件已发送至 ${result.recipient}。` });
    } catch (error) {
      setMessage({ tone: "error", text: humanizeApiError(error) });
    } finally {
      setTesting(false);
    }
  }

  return (
    <>
      <StudioTopbar
        actions={
          <button className="primary-button" disabled={loading || saving || !settings} form="studio-settings-form" type="submit">
            {saving ? <Loader2 aria-hidden className="spin" size={18} /> : <Save aria-hidden size={18} />}
            保存设置
          </button>
        }
        description="配置应用级邮件与登录集成；敏感字段只通过 BFF Cookie 会话写入。"
        eyebrow="Application settings"
        title="设置"
      />

      <div className={styles.segmentedTabs} aria-label="设置分类">
        <button
          aria-pressed={activeSection === "email"}
          className={activeSection === "email" ? styles.segmentedTabActive : styles.segmentedTab}
          type="button"
          onClick={() => {
            setActiveSection("email");
            setMessage(null);
          }}
        >
          <Mail aria-hidden size={18} />
          Email
        </button>
        <button
          aria-pressed={activeSection === "github"}
          className={activeSection === "github" ? styles.segmentedTabActive : styles.segmentedTab}
          type="button"
          onClick={() => {
            setActiveSection("github");
            setMessage(null);
          }}
        >
          <Github aria-hidden size={18} />
          GitHub OAuth2
        </button>
      </div>

      <StudioPanel action={settingsPanelIcon(activeSection)} title={settingsPanelTitle(activeSection)}>
        {loading ? (
          <div className={styles.emptyState} role="status">
            <Loader2 aria-hidden className="spin" size={24} />
            <strong>正在加载设置...</strong>
          </div>
        ) : !settings ? (
          <div className={styles.emptyState} role="alert">
            <Settings aria-hidden size={28} />
            <strong>设置加载失败</strong>
            <span>{message?.text ?? "无法读取应用设置，请稍后再试。"}</span>
          </div>
        ) : (
          activeSection === "email" ? (
          <form className={styles.formGrid} id="studio-settings-form" onSubmit={onSubmit}>
            <label className={styles.checkboxField}>
              <input
                checked={email.enabled}
                type="checkbox"
                onChange={(event) => updateEmail("enabled", event.target.checked)}
              />
              <span>启用 SMTP 邮件发送</span>
            </label>

            <label className={styles.field}>
              <span>SMTP Host</span>
              <input value={email.host} onChange={(event) => updateEmail("host", event.target.value)} />
            </label>

            <label className={styles.field}>
              <span>SMTP Port</span>
              <input
                max={65535}
                min={1}
                type="number"
                value={email.port}
                onChange={(event) => updateEmail("port", Number(event.target.value))}
              />
            </label>

            <label className={styles.field}>
              <span>Username</span>
              <input autoComplete="username" value={email.username} onChange={(event) => updateEmail("username", event.target.value)} />
            </label>

            <label className={styles.field}>
              <span>TLS 模式</span>
              <StudioSelect ariaLabel="TLS 模式" options={tlsOptions} value={email.tls} onChange={(value) => updateEmail("tls", value)} />
            </label>

            <label className={styles.field}>
              <span>发件人邮箱</span>
              <input type="email" value={email.from} onChange={(event) => updateEmail("from", event.target.value)} />
            </label>

            <label className={styles.field}>
              <span>发件人名称</span>
              <input value={email.from_name} onChange={(event) => updateEmail("from_name", event.target.value)} />
            </label>

            <label className={styles.fieldFull}>
              <span>SMTP 密码</span>
              <div className={styles.passwordRow}>
                <input
                  autoComplete="new-password"
                  placeholder={email.password_set ? "已设置；留空保持不变" : "未设置"}
                  type="password"
                  value={password}
                  onChange={(event) => {
                    setPassword(event.target.value);
                    setPasswordMode(event.target.value === "" ? "keep" : "set");
                  }}
                />
                <button
                  className="secondary-button"
                  disabled={!email.password_set && password === ""}
                  type="button"
                  onClick={() => {
                    setPassword("");
                    setPasswordMode("clear");
                  }}
                >
                  清空密码
                </button>
              </div>
            </label>

            <div className={`${styles.metaRow} ${styles.fieldFull}`}>
              <span className={`${styles.statusPill} ${email.password_set ? styles.statusReady : styles.statusPending}`}>
                {email.password_set ? "Password set" : "No password"}
              </span>
              <span className={styles.muted}>{passwordHint}</span>
              {settings ? <span className={styles.muted}>Revision {settings.revision}</span> : null}
            </div>

            <label className={styles.fieldFull}>
              <span>测试收件人</span>
              <div className={styles.passwordRow}>
                <input
                  aria-label="测试收件人"
                  autoComplete="email"
                  placeholder="recipient@example.com"
                  type="email"
                  value={testRecipient}
                  onChange={(event) => setTestRecipient(event.target.value)}
                />
                <button
                  className="secondary-button"
                  disabled={saving || testing}
                  type="button"
                  onClick={onSendTestEmail}
                >
                  {testing ? <Loader2 aria-hidden className="spin" size={18} /> : <Send aria-hidden size={18} />}
                  发送测试邮件
                </button>
              </div>
              <span className={styles.muted}>测试使用已保存的 SMTP 配置；未保存修改不会参与发送。</span>
            </label>

            {message ? (
              <p
                className={`${styles.message} ${message.tone === "success" ? styles.messageSuccess : styles.messageError}`}
                role={message.tone === "error" ? "alert" : "status"}
              >
                {message.text}
              </p>
            ) : null}
          </form>
          ) : (
          <form className={styles.formGrid} id="studio-settings-form" onSubmit={onSubmitGithubOAuth2}>
            <label className={styles.checkboxField}>
              <input
                aria-label="启用 GitHub OAuth2 登录"
                checked={githubOAuth2.enabled}
                type="checkbox"
                onChange={(event) => updateGithubOAuth2("enabled", event.target.checked)}
              />
              <span>启用 GitHub OAuth2 登录</span>
            </label>

            <label className={styles.field}>
              <span>Client ID</span>
              <input value={githubOAuth2.client_id} onChange={(event) => updateGithubOAuth2("client_id", event.target.value)} />
            </label>

            <label className={styles.field}>
              <span>Redirect URL</span>
              <input value={githubOAuth2.redirect_url} onChange={(event) => updateGithubOAuth2("redirect_url", event.target.value)} />
            </label>

            <label className={styles.fieldFull}>
              <span>Client Secret</span>
              <div className={styles.passwordRow}>
                <input
                  autoComplete="new-password"
                  placeholder={githubOAuth2.client_secret_set ? "已设置；留空保持不变" : "未设置"}
                  type="password"
                  value={githubSecret}
                  onChange={(event) => {
                    setGithubSecret(event.target.value);
                    setGithubSecretMode(event.target.value === "" ? "keep" : "set");
                  }}
                />
                <button
                  className="secondary-button"
                  disabled={!githubOAuth2.client_secret_set && githubSecret === ""}
                  type="button"
                  onClick={() => {
                    setGithubSecret("");
                    setGithubSecretMode("clear");
                  }}
                >
                  清空 Client Secret
                </button>
              </div>
            </label>

            <div className={`${styles.metaRow} ${styles.fieldFull}`}>
              <span className={`${styles.statusPill} ${githubOAuth2.client_secret_set ? styles.statusReady : styles.statusPending}`}>
                {githubOAuth2.client_secret_set ? "Client secret set" : "Client secret not set"}
              </span>
              <span className={styles.muted}>{githubSecretHint}</span>
              {settings ? <span className={styles.muted}>Revision {settings.revision}</span> : null}
            </div>

            <button
              aria-expanded={githubAdvancedOpen}
              className={`${styles.disclosureButton} ${styles.fieldFull}`}
              type="button"
              onClick={() => setGithubAdvancedOpen((open) => !open)}
            >
              <ChevronDown aria-hidden className={githubAdvancedOpen ? styles.disclosureIconOpen : styles.disclosureIcon} size={18} />
              高级设置
            </button>

            {githubAdvancedOpen ? (
              <>
                <label className={styles.field}>
                  <span>Auth URL</span>
                  <input value={githubOAuth2.auth_url} onChange={(event) => updateGithubOAuth2("auth_url", event.target.value)} />
                </label>
                <label className={styles.field}>
                  <span>Token URL</span>
                  <input value={githubOAuth2.token_url} onChange={(event) => updateGithubOAuth2("token_url", event.target.value)} />
                </label>
                <label className={styles.fieldFull}>
                  <span>User Info URL</span>
                  <input value={githubOAuth2.user_info_url} onChange={(event) => updateGithubOAuth2("user_info_url", event.target.value)} />
                </label>
                <label className={styles.checkboxField}>
                  <input
                    aria-label="允许非 GitHub 端点"
                    checked={githubOAuth2.allow_non_github_endpoints}
                    type="checkbox"
                    onChange={(event) => updateGithubOAuth2("allow_non_github_endpoints", event.target.checked)}
                  />
                  <span>允许非 GitHub 端点</span>
                </label>
              </>
            ) : null}

            {message ? (
              <p
                className={`${styles.message} ${message.tone === "success" ? styles.messageSuccess : styles.messageError}`}
                role={message.tone === "error" ? "alert" : "status"}
              >
                {message.text}
              </p>
            ) : null}
          </form>
          )
        )}
      </StudioPanel>
    </>
  );
}

function settingsPanelTitle(section: SettingsSection) {
  if (section === "github") return "GitHub OAuth2";
  return "邮件 SMTP";
}

function settingsPanelIcon(section: SettingsSection) {
  if (section === "github") return <Github aria-hidden size={20} />;
  return <Settings aria-hidden size={20} />;
}

function validateEmail(email: EmailSettingsPublic) {
  if (email.port < 1 || email.port > 65535 || !Number.isInteger(email.port)) {
    return "SMTP port 必须在 1 到 65535 之间。";
  }
  if (!["none", "starttls", "tls"].includes(email.tls)) {
    return "TLS 模式必须是 none、starttls 或 tls。";
  }
  if (!email.enabled) {
    return null;
  }
  if (email.host.trim() === "") {
    return "启用 SMTP 时必须填写 host。";
  }
  if (email.from.trim() === "") {
    return "启用 SMTP 时必须填写发件人邮箱。";
  }
  if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email.from.trim())) {
    return "发件人邮箱格式不正确。";
  }
  return null;
}

function validateTestRecipient(email: EmailSettingsPublic, recipient: string) {
  if (!email.enabled) {
    return "发送测试邮件前必须先启用 SMTP。";
  }
  if (recipient === "") {
    return "请填写测试收件人邮箱。";
  }
  if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(recipient)) {
    return "测试收件人邮箱格式不正确。";
  }
  return null;
}

function validateGithubOAuth2(githubOAuth2: GithubOAuth2SettingsPublic, githubSecret: string, githubSecretMode: PasswordMode) {
  if (!githubOAuth2.enabled) {
    return null;
  }
  if (githubOAuth2.client_id.trim() === "") {
    return "启用 GitHub OAuth2 时必须填写 Client ID。";
  }
  if (!githubOAuth2.client_secret_set && githubSecret.trim() === "" && githubSecretMode !== "clear") {
    return "启用 GitHub OAuth2 时必须填写 Client Secret。";
  }
  if (githubOAuth2.redirect_url.trim() === "") {
    return "启用 GitHub OAuth2 时必须填写 Redirect URL。";
  }
  for (const [label, value] of [
    ["Redirect URL", githubOAuth2.redirect_url],
    ["Auth URL", githubOAuth2.auth_url],
    ["Token URL", githubOAuth2.token_url],
    ["User Info URL", githubOAuth2.user_info_url]
  ] as const) {
    if (value.trim() === "") continue;
    try {
      const parsed = new URL(value);
      if (parsed.protocol !== "http:" && parsed.protocol !== "https:") {
        return `${label} 必须使用 http 或 https。`;
      }
    } catch {
      return `${label} 格式不正确。`;
    }
  }
  return null;
}
