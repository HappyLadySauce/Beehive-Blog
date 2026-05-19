"use client";

import Link from "next/link";
import { Loader2 } from "lucide-react";
import { useEffect, useState } from "react";
import { useRouter, useSearchParams } from "next/navigation";

import { login } from "@/lib/api/auth";
import { humanizeApiError } from "@/lib/api/client";
import { useToast } from "@/components/toast/ToastProvider";
import { useAuth } from "./AuthProvider";

export function GithubCallback() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const { setAuth } = useAuth();
  const toast = useToast();
  const code = searchParams.get("code");
  const state = searchParams.get("state");
  const hasCallbackParams = Boolean(code && state);
  const [message, setMessage] = useState("正在完成 GitHub 登录...");

  useEffect(() => {
    if (!code || !state) {
      toast.error("GitHub 回调缺少 code 或 state。");
      return;
    }

    login({ grant_type: "github_oauth2", code, state })
      .then((payload) => {
        setAuth(payload);
        router.replace("/studio");
      })
      .catch((error) => {
        const text = humanizeApiError(error);
        setMessage(text);
        toast.error(text);
      });
  }, [code, router, setAuth, state, toast]);

  return (
    <section className="auth-panel">
      <div className="status-line">
        <Loader2 aria-hidden className="spin" size={20} />
        <p>{hasCallbackParams ? message : "GitHub 回调缺少 code 或 state。"}</p>
      </div>
      <Link className="text-link" href="/login">返回登录页</Link>
    </section>
  );
}
