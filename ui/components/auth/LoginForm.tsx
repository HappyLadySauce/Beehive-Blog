"use client";

import Link from "next/link";
import { Github, Loader2, LogIn } from "lucide-react";
import { FormEvent, useState } from "react";
import { useRouter, useSearchParams } from "next/navigation";

import { beginGithubOAuth, login } from "@/lib/api/auth";
import { humanizeApiError } from "@/lib/api/client";
import { useAuth } from "./AuthProvider";

export function LoginForm() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const { setAuth } = useAuth();
  const [account, setAccount] = useState("");
  const [password, setPassword] = useState("");
  const [message, setMessage] = useState<string | null>(null);
  const [pending, setPending] = useState(false);
  const [githubPending, setGithubPending] = useState(false);

  const next = searchParams.get("next") ?? "/studio";

  async function onSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setPending(true);
    setMessage(null);
    try {
      const payload = await login({ grant_type: "local", account, password });
      setAuth(payload);
      router.replace(next);
    } catch (error) {
      setMessage(humanizeApiError(error));
    } finally {
      setPending(false);
    }
  }

  async function onGithubLogin() {
    setGithubPending(true);
    setMessage(null);
    try {
      const payload = await beginGithubOAuth();
      window.location.assign(payload.auth_url);
    } catch (error) {
      setMessage(humanizeApiError(error));
      setGithubPending(false);
    }
  }

  return (
    <form className="auth-form" onSubmit={onSubmit}>
      <label>
        <span>用户名或邮箱</span>
        <input
          autoComplete="username"
          maxLength={320}
          required
          value={account}
          onChange={(event) => setAccount(event.target.value)}
        />
      </label>
      <label>
        <span>密码</span>
        <input
          autoComplete="current-password"
          maxLength={72}
          minLength={8}
          required
          type="password"
          value={password}
          onChange={(event) => setPassword(event.target.value)}
        />
      </label>
      {message ? <p className="form-message" role="alert">{message}</p> : null}
      <button className="primary-button" disabled={pending || githubPending} type="submit">
        {pending ? <Loader2 aria-hidden className="spin" size={18} /> : <LogIn aria-hidden size={18} />}
        登录
      </button>
      <button className="secondary-button" disabled={pending || githubPending} type="button" onClick={onGithubLogin}>
        {githubPending ? <Loader2 aria-hidden className="spin" size={18} /> : <Github aria-hidden size={18} />}
        GitHub 登录
      </button>
      <p className="auth-switch">
        还没有账号？<Link href="/register">注册</Link>
      </p>
    </form>
  );
}
