"use client";

import { FormEvent, useEffect, useMemo, useState } from "react";
import { Loader2, Save, Send, Settings } from "lucide-react";

import { humanizeApiError } from "@/lib/api/client";
import { getSettings, patchSettings, testEmailSettings } from "@/lib/api/settings";
import type { EmailSettingsPublic, SettingsResponse } from "@/lib/api/types";
import styles from "./Studio.module.css";
import { StudioPanel } from "./StudioPanel";
import { StudioTopbar } from "./StudioTopbar";

type PasswordMode = "keep" | "set" | "clear";

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

let settingsLoadRequest: Promise<SettingsResponse> | null = null;

function loadSettings() {
  if (!settingsLoadRequest) {
    settingsLoadRequest = getSettings().finally(() => {
      settingsLoadRequest = null;
    });
  }
  return settingsLoadRequest;
}

export function StudioSettingsPage() {
  const [settings, setSettings] = useState<SettingsResponse | null>(null);
  const [email, setEmail] = useState<EmailSettingsPublic>(defaultEmail);
  const [password, setPassword] = useState("");
  const [passwordMode, setPasswordMode] = useState<PasswordMode>("keep");
  const [testRecipient, setTestRecipient] = useState("");
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [testing, setTesting] = useState(false);
  const [message, setMessage] = useState<{ tone: "success" | "error"; text: string } | null>(null);

  useEffect(() => {
    let active = true;

    loadSettings()
      .then((payload) => {
        if (!active) return;
        setSettings(payload);
        setEmail(payload.email);
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

  function updateEmail<K extends keyof EmailSettingsPublic>(key: K, value: EmailSettingsPublic[K]) {
    setEmail((current) => ({ ...current, [key]: value }));
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
        description="配置应用级 SMTP 邮件发送能力；敏感字段只通过 BFF Cookie 会话写入。"
        eyebrow="Application settings"
        title="设置"
      />

      <StudioPanel action={<Settings aria-hidden size={20} />} title="邮件 SMTP">
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
              <select value={email.tls} onChange={(event) => updateEmail("tls", event.target.value)}>
                <option value="none">None</option>
                <option value="starttls">STARTTLS</option>
                <option value="tls">TLS</option>
              </select>
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
        )}
      </StudioPanel>
    </>
  );
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
