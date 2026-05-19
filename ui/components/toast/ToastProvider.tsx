"use client";

import { X } from "lucide-react";
import { createContext, ReactNode, useCallback, useContext, useEffect, useMemo, useRef, useState } from "react";

export type ToastTone = "success" | "error";

export type ToastInput = {
  tone: ToastTone;
  text: string;
};

type ToastItem = ToastInput & {
  id: string;
};

type ToastContextValue = {
  clear: () => void;
  dismiss: (id: string) => void;
  error: (text: string) => string;
  showToast: (input: ToastInput) => string;
  success: (text: string) => string;
};

const ToastContext = createContext<ToastContextValue | null>(null);
const successAutoDismissMs = 3000;

const noopToastContext: ToastContextValue = {
  clear: () => undefined,
  dismiss: () => undefined,
  error: () => "",
  showToast: () => "",
  success: () => ""
};

export function ToastProvider({ children }: { children: ReactNode }) {
  const [items, setItems] = useState<ToastItem[]>([]);

  const dismiss = useCallback((id: string) => {
    setItems((current) => current.filter((item) => item.id !== id));
  }, []);

  const clear = useCallback(() => {
    setItems([]);
  }, []);

  const showToast = useCallback(
    (input: ToastInput) => {
      const id = `${Date.now()}-${Math.random().toString(36).slice(2)}`;
      setItems((current) => [...current, { ...input, id }]);
      if (input.tone === "success") {
        window.setTimeout(() => dismiss(id), successAutoDismissMs);
      }
      return id;
    },
    [dismiss]
  );

  const value = useMemo<ToastContextValue>(
    () => ({
      clear,
      dismiss,
      error: (text) => showToast({ tone: "error", text }),
      showToast,
      success: (text) => showToast({ tone: "success", text })
    }),
    [clear, dismiss, showToast]
  );

  return (
    <ToastContext.Provider value={value}>
      {children}
      <ToastViewport items={items} onDismiss={dismiss} />
    </ToastContext.Provider>
  );
}

export function useToast() {
  return useContext(ToastContext) ?? noopToastContext;
}

export function ToastMessage({ message }: { message: ToastInput | null }) {
  const toast = useToast();
  const lastMessageKey = useRef("");

  useEffect(() => {
    if (!message) {
      lastMessageKey.current = "";
      return;
    }
    const key = `${message.tone}:${message.text}`;
    if (lastMessageKey.current === key) return;
    lastMessageKey.current = key;
    toast.showToast(message);
  }, [message, toast]);

  return null;
}

function ToastViewport({ items, onDismiss }: { items: ToastItem[]; onDismiss: (id: string) => void }) {
  if (items.length === 0) return null;

  return (
    <div className="toastViewport" aria-label="全站消息提示">
      {items.map((item) => (
        <div className={`toastItem ${item.tone === "success" ? "toastSuccess" : "toastError"}`} key={item.id} role={item.tone === "error" ? "alert" : "status"}>
          <span>{item.text}</span>
          <button className="toastClose" type="button" aria-label="关闭消息" onClick={() => onDismiss(item.id)}>
            <X aria-hidden size={16} />
          </button>
        </div>
      ))}
    </div>
  );
}
