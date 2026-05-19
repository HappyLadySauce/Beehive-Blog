"use client";

import Link from "next/link";
import { Loader2, UserPlus } from "lucide-react";
import { FormEvent, useState } from "react";
import { useRouter } from "next/navigation";

import { register } from "@/lib/api/auth";
import { humanizeApiError } from "@/lib/api/client";
import { useToast } from "@/components/toast/ToastProvider";
import { useAuth } from "./AuthProvider";

export function RegisterForm() {
  const router = useRouter();
  const { setAuth } = useAuth();
  const toast = useToast();
  const [username, setUsername] = useState("");
  const [email, setEmail] = useState("");
  const [nickname, setNickname] = useState("");
  const [password, setPassword] = useState("");
  const [pending, setPending] = useState(false);

  async function onSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setPending(true);
    try {
      const payload = await register({
        username,
        password,
        email: email || undefined,
        nickname: nickname || undefined
      });
      setAuth(payload);
      router.replace("/");
    } catch (error) {
      toast.error(humanizeApiError(error));
    } finally {
      setPending(false);
    }
  }

  return (
    <form className="auth-form" onSubmit={onSubmit}>
      <label>
        <span>用户名</span>
        <input autoComplete="username" maxLength={64} required value={username} onChange={(event) => setUsername(event.target.value)} />
      </label>
      <label>
        <span>邮箱</span>
        <input autoComplete="email" maxLength={320} type="email" value={email} onChange={(event) => setEmail(event.target.value)} />
      </label>
      <label>
        <span>昵称</span>
        <input autoComplete="nickname" maxLength={128} value={nickname} onChange={(event) => setNickname(event.target.value)} />
      </label>
      <label>
        <span>密码</span>
        <input
          autoComplete="new-password"
          maxLength={72}
          minLength={8}
          required
          type="password"
          value={password}
          onChange={(event) => setPassword(event.target.value)}
        />
      </label>
      <button className="primary-button" disabled={pending} type="submit">
        {pending ? <Loader2 aria-hidden className="spin" size={18} /> : <UserPlus aria-hidden size={18} />}
        注册并登录
      </button>
      <p className="auth-switch">
        已有账号？<Link href="/login">登录</Link>
      </p>
    </form>
  );
}
