"use client";

import { FormEvent, useEffect, useMemo, useState } from "react";
import { ChevronDown, Github, Globe2, Loader2, Mail, Save, Send, Settings, UserRound } from "lucide-react";

import { humanizeApiError } from "@/lib/api/client";
import {
  getGithubOAuth2Settings,
  getProfileSettings,
  getSiteSettings,
  getSettings,
  patchGithubOAuth2Settings,
  patchProfileSettings,
  patchSiteSettings,
  patchSettings,
  testEmailSettings
} from "@/lib/api/settings";
import { ToastMessage } from "@/components/toast/ToastProvider";
import type {
  EmailSettingsPublic,
  GithubOAuth2SettingsPublic,
  ProfileSettingsPublic,
  SettingsResponse,
  SiteSettingsPublic
} from "@/lib/api/types";
import styles from "./Studio.module.css";
import { StudioPanel } from "./StudioPanel";
import { StudioSelect } from "./StudioSelect";
import { StudioTopbar } from "./StudioTopbar";

type PasswordMode = "keep" | "set" | "clear";
type SettingsSection = "email" | "github" | "profile" | "site";

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

const defaultProfile: ProfileSettingsPublic = {
  display_name: "Beehive",
  avatar_url: "",
  headline: "个人博客与知识中台",
  bio: "",
  location: "",
  website: ""
};

const defaultSite: SiteSettingsPublic = {
  name: "Beehive",
  url: "",
  subtitle: "Beehive Blog",
  description: "个人博客、AI 协作创作与面向智能体的个人知识中台。",
  keywords: "",
  logo_url: "",
  favicon_url: "",
  icp_beian: "",
  police_beian: "",
  footer_text: ""
};

const tlsOptions = [
  { value: "none", label: "None" },
  { value: "starttls", label: "STARTTLS" },
  { value: "tls", label: "TLS" }
];

let settingsInflight: Promise<SettingsResponse> | null = null;
let githubSettingsInflight: Promise<SettingsResponse> | null = null;
let profileSettingsInflight: Promise<SettingsResponse> | null = null;
let siteSettingsInflight: Promise<SettingsResponse> | null = null;

function requestSettings() {
  return getSettings();
}

// loadSettings dedupes in-flight settings requests on initial mount.
// loadSettings 在首屏挂载时对进行中的设置请求去重。
function loadSettings() {
  if (settingsInflight) {
    return settingsInflight;
  }
  const promise = requestSettings().finally(() => {
    settingsInflight = null;
  });
  settingsInflight = promise;
  return promise;
}

function requestGithubOAuth2Settings() {
  return getGithubOAuth2Settings();
}

// loadGithubOAuth2Settings dedupes in-flight GitHub settings requests when the tab opens.
// loadGithubOAuth2Settings 在打开 GitHub 标签时对进行中的请求去重。
function loadGithubOAuth2Settings() {
  if (githubSettingsInflight) {
    return githubSettingsInflight;
  }
  const promise = requestGithubOAuth2Settings().finally(() => {
    githubSettingsInflight = null;
  });
  githubSettingsInflight = promise;
  return promise;
}

function requestProfileSettings() {
  return getProfileSettings();
}

function loadProfileSettings() {
  if (profileSettingsInflight) {
    return profileSettingsInflight;
  }
  const promise = requestProfileSettings().finally(() => {
    profileSettingsInflight = null;
  });
  profileSettingsInflight = promise;
  return promise;
}

function requestSiteSettings() {
  return getSiteSettings();
}

function loadSiteSettings() {
  if (siteSettingsInflight) {
    return siteSettingsInflight;
  }
  const promise = requestSiteSettings().finally(() => {
    siteSettingsInflight = null;
  });
  siteSettingsInflight = promise;
  return promise;
}

// resetSettingsPageModuleStateForTests clears module caches between unit tests.
// resetSettingsPageModuleStateForTests 在单元测试之间清空模块级缓存。
export function resetSettingsPageModuleStateForTests() {
  settingsInflight = null;
  githubSettingsInflight = null;
  profileSettingsInflight = null;
  siteSettingsInflight = null;
}

export function StudioSettingsPage() {
  const [settings, setSettings] = useState<SettingsResponse | null>(null);
  const [activeSection, setActiveSection] = useState<SettingsSection>("email");
  const [email, setEmail] = useState<EmailSettingsPublic>(defaultEmail);
  const [githubOAuth2, setGithubOAuth2] = useState<GithubOAuth2SettingsPublic>(defaultGithubOAuth2);
  const [profile, setProfile] = useState<ProfileSettingsPublic>(defaultProfile);
  const [site, setSite] = useState<SiteSettingsPublic>(defaultSite);
  const [password, setPassword] = useState("");
  const [passwordMode, setPasswordMode] = useState<PasswordMode>("keep");
  const [githubSecret, setGithubSecret] = useState("");
  const [githubSecretMode, setGithubSecretMode] = useState<PasswordMode>("keep");
  const [githubAdvancedOpen, setGithubAdvancedOpen] = useState(false);
  const [testRecipient, setTestRecipient] = useState("");
  const [loading, setLoading] = useState(true);
  const [githubLoaded, setGithubLoaded] = useState(false);
  const [profileLoaded, setProfileLoaded] = useState(false);
  const [siteLoaded, setSiteLoaded] = useState(false);
  const [loadingGithub, setLoadingGithub] = useState(false);
  const [loadingProfile, setLoadingProfile] = useState(false);
  const [loadingSite, setLoadingSite] = useState(false);
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
        setGithubOAuth2(payload.github_oauth2 ?? defaultGithubOAuth2);
        setProfile(payload.profile ?? defaultProfile);
        setSite(payload.site ?? defaultSite);
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

  useEffect(() => {
    if (activeSection !== "github" || githubLoaded || !settings) {
      return;
    }

    let active = true;
    const currentSettings = settings;
    void loadGithubSettings();

    async function loadGithubSettings() {
      setLoadingGithub(true);
      try {
        const githubPayload = await loadGithubOAuth2Settings();
        if (!active) return;
        const githubOAuth2Settings = githubPayload.github_oauth2 ?? currentSettings.github_oauth2 ?? defaultGithubOAuth2;
        setSettings((current) =>
          current
            ? {
                ...current,
                revision: Math.max(current.revision, githubPayload.revision),
                github_oauth2: githubOAuth2Settings
              }
            : current
        );
        setGithubOAuth2(githubOAuth2Settings);
        setGithubLoaded(true);
      } catch (error) {
        if (active) setMessage({ tone: "error", text: humanizeApiError(error) });
      } finally {
        if (active) setLoadingGithub(false);
      }
    }

    return () => {
      active = false;
    };
  }, [activeSection, githubLoaded, settings]);

  useEffect(() => {
    if (activeSection !== "profile" || profileLoaded || !settings) {
      return;
    }

    let active = true;
    void loadProfile();

    async function loadProfile() {
      setLoadingProfile(true);
      try {
        const profilePayload = await loadProfileSettings();
        if (!active) return;
        const profileSettings = profilePayload.profile ?? defaultProfile;
        setSettings((current) =>
          current
            ? {
                ...current,
                revision: Math.max(current.revision, profilePayload.revision),
                profile: profileSettings
              }
            : current
        );
        setProfile(profileSettings);
        setProfileLoaded(true);
      } catch (error) {
        if (active) setMessage({ tone: "error", text: humanizeApiError(error) });
      } finally {
        if (active) setLoadingProfile(false);
      }
    }

    return () => {
      active = false;
    };
  }, [activeSection, profileLoaded, settings]);

  useEffect(() => {
    if (activeSection !== "site" || siteLoaded || !settings) {
      return;
    }

    let active = true;
    void loadSite();

    async function loadSite() {
      setLoadingSite(true);
      try {
        const sitePayload = await loadSiteSettings();
        if (!active) return;
        const siteSettings = sitePayload.site ?? defaultSite;
        setSettings((current) =>
          current
            ? {
                ...current,
                revision: Math.max(current.revision, sitePayload.revision),
                site: siteSettings
              }
            : current
        );
        setSite(siteSettings);
        setSiteLoaded(true);
      } catch (error) {
        if (active) setMessage({ tone: "error", text: humanizeApiError(error) });
      } finally {
        if (active) setLoadingSite(false);
      }
    }

    return () => {
      active = false;
    };
  }, [activeSection, siteLoaded, settings]);

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

  function updateProfile<K extends keyof ProfileSettingsPublic>(key: K, value: ProfileSettingsPublic[K]) {
    setProfile((current) => ({ ...current, [key]: value }));
  }

  function updateSite<K extends keyof SiteSettingsPublic>(key: K, value: SiteSettingsPublic[K]) {
    setSite((current) => ({ ...current, [key]: value }));
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

  async function onSubmitProfile(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const validation = validateProfile(profile);
    if (validation) {
      setMessage({ tone: "error", text: validation });
      return;
    }

    setSaving(true);
    setMessage(null);
    try {
      const next = await patchProfileSettings({
        display_name: profile.display_name,
        avatar_url: profile.avatar_url,
        headline: profile.headline,
        bio: profile.bio,
        location: profile.location,
        website: profile.website
      });
      setSettings(next);
      setProfile(next.profile ?? defaultProfile);
      setProfileLoaded(true);
      setMessage({ tone: "success", text: "个人设置已保存。" });
    } catch (error) {
      setMessage({ tone: "error", text: humanizeApiError(error) });
    } finally {
      setSaving(false);
    }
  }

  async function onSubmitSite(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const validation = validateSite(site);
    if (validation) {
      setMessage({ tone: "error", text: validation });
      return;
    }

    setSaving(true);
    setMessage(null);
    try {
      const next = await patchSiteSettings({
        name: site.name,
        url: site.url,
        subtitle: site.subtitle,
        description: site.description,
        keywords: site.keywords,
        logo_url: site.logo_url,
        favicon_url: site.favicon_url,
        icp_beian: site.icp_beian,
        police_beian: site.police_beian,
        footer_text: site.footer_text
      });
      setSettings(next);
      setSite(next.site ?? defaultSite);
      setSiteLoaded(true);
      setMessage({ tone: "success", text: "站点设置已保存。" });
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
        <button
          aria-pressed={activeSection === "profile"}
          className={activeSection === "profile" ? styles.segmentedTabActive : styles.segmentedTab}
          type="button"
          onClick={() => {
            setActiveSection("profile");
            setMessage(null);
          }}
        >
          <UserRound aria-hidden size={18} />
          个人设置
        </button>
        <button
          aria-pressed={activeSection === "site"}
          className={activeSection === "site" ? styles.segmentedTabActive : styles.segmentedTab}
          type="button"
          onClick={() => {
            setActiveSection("site");
            setMessage(null);
          }}
        >
          <Globe2 aria-hidden size={18} />
          站点设置
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

            <ToastMessage message={message} />
          </form>
          ) : activeSection === "github" ? (
          loadingGithub ? (
          <div className={styles.emptyState} role="status">
            <Loader2 aria-hidden className="spin" size={24} />
            <strong>正在加载 GitHub OAuth2 设置...</strong>
          </div>
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

            <ToastMessage message={message} />
          </form>
          )
          ) : activeSection === "profile" ? (
          loadingProfile ? (
          <div className={styles.emptyState} role="status">
            <Loader2 aria-hidden className="spin" size={24} />
            <strong>正在加载个人设置...</strong>
          </div>
          ) : (
          <form className={styles.formGrid} id="studio-settings-form" onSubmit={onSubmitProfile}>
            <label className={styles.field}>
              <span>显示名称</span>
              <input value={profile.display_name} onChange={(event) => updateProfile("display_name", event.target.value)} />
            </label>

            <label className={styles.field}>
              <span>头像 URL</span>
              <input placeholder="https://example.com/avatar.png" value={profile.avatar_url} onChange={(event) => updateProfile("avatar_url", event.target.value)} />
            </label>

            <label className={styles.fieldFull}>
              <span>一句话简介</span>
              <input value={profile.headline} onChange={(event) => updateProfile("headline", event.target.value)} />
            </label>

            <label className={styles.fieldFull}>
              <span>个人介绍</span>
              <textarea
                className={styles.textarea}
                value={profile.bio}
                onChange={(event) => updateProfile("bio", event.target.value)}
              />
            </label>

            <label className={styles.field}>
              <span>所在地</span>
              <input value={profile.location} onChange={(event) => updateProfile("location", event.target.value)} />
            </label>

            <label className={styles.field}>
              <span>个人网站</span>
              <input placeholder="https://example.com" value={profile.website} onChange={(event) => updateProfile("website", event.target.value)} />
            </label>

            <div className={`${styles.metaRow} ${styles.fieldFull}`}>
              <span className={styles.muted}>这些信息会用于公开首页个人卡片。</span>
              {settings ? <span className={styles.muted}>Revision {settings.revision}</span> : null}
            </div>

            <ToastMessage message={message} />
          </form>
          )
          ) : loadingSite ? (
          <div className={styles.emptyState} role="status">
            <Loader2 aria-hidden className="spin" size={24} />
            <strong>正在加载站点设置...</strong>
          </div>
          ) : (
          <form className={styles.formGrid} id="studio-settings-form" onSubmit={onSubmitSite}>
            <label className={styles.field}>
              <span>站点名称</span>
              <input value={site.name} onChange={(event) => updateSite("name", event.target.value)} />
            </label>

            <label className={styles.field}>
              <span>站点 URL</span>
              <input placeholder="https://example.com" value={site.url} onChange={(event) => updateSite("url", event.target.value)} />
            </label>

            <label className={styles.fieldFull}>
              <span>副标题</span>
              <input value={site.subtitle} onChange={(event) => updateSite("subtitle", event.target.value)} />
            </label>

            <label className={styles.fieldFull}>
              <span>站点描述</span>
              <textarea
                className={styles.textarea}
                value={site.description}
                onChange={(event) => updateSite("description", event.target.value)}
              />
            </label>

            <label className={styles.fieldFull}>
              <span>关键词</span>
              <input placeholder="blog, AI, knowledge" value={site.keywords} onChange={(event) => updateSite("keywords", event.target.value)} />
            </label>

            <label className={styles.field}>
              <span>Logo URL</span>
              <input placeholder="https://example.com/logo.png" value={site.logo_url} onChange={(event) => updateSite("logo_url", event.target.value)} />
            </label>

            <label className={styles.field}>
              <span>Favicon URL</span>
              <input placeholder="https://example.com/favicon.ico" value={site.favicon_url} onChange={(event) => updateSite("favicon_url", event.target.value)} />
            </label>

            <label className={styles.field}>
              <span>ICP备案号</span>
              <input value={site.icp_beian} onChange={(event) => updateSite("icp_beian", event.target.value)} />
            </label>

            <label className={styles.field}>
              <span>公安备案号</span>
              <input value={site.police_beian} onChange={(event) => updateSite("police_beian", event.target.value)} />
            </label>

            <label className={styles.fieldFull}>
              <span>页脚文案</span>
              <input value={site.footer_text} onChange={(event) => updateSite("footer_text", event.target.value)} />
            </label>

            <div className={`${styles.metaRow} ${styles.fieldFull}`}>
              <span className={styles.muted}>这些信息会进入公开站点概览，可用于首页、SEO 和页脚展示。</span>
              {settings ? <span className={styles.muted}>Revision {settings.revision}</span> : null}
            </div>

            <ToastMessage message={message} />
          </form>
          )
        )}
      </StudioPanel>
    </>
  );
}

function settingsPanelTitle(section: SettingsSection) {
  if (section === "github") return "GitHub OAuth2";
  if (section === "profile") return "个人设置";
  if (section === "site") return "站点设置";
  return "邮件 SMTP";
}

function settingsPanelIcon(section: SettingsSection) {
  if (section === "github") return <Github aria-hidden size={20} />;
  if (section === "profile") return <UserRound aria-hidden size={20} />;
  if (section === "site") return <Globe2 aria-hidden size={20} />;
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

function validateProfile(profile: ProfileSettingsPublic) {
  if (profile.display_name.trim() === "") {
    return "显示名称不能为空。";
  }
  for (const [label, value] of [
    ["头像 URL", profile.avatar_url],
    ["个人网站", profile.website]
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

function validateSite(site: SiteSettingsPublic) {
  if (site.name.trim() === "") {
    return "站点名称不能为空。";
  }
  for (const [label, value] of [
    ["站点 URL", site.url],
    ["Logo URL", site.logo_url],
    ["Favicon URL", site.favicon_url]
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
